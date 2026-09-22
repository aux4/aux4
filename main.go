package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"aux4.dev/aux4/aux4"
	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/config"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/coverage"
	"aux4.dev/aux4/daemon"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/executor"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

// crashExitCode is returned when a panic escapes to the top-level recover()
// backstop. It is a generic, non-zero "something went wrong internally" code,
// distinct from the specific Aux4Error exit codes (1 for a normal internal
// error, 127 for command not found, 130 for user-aborted, etc.).
const crashExitCode = 1

func main() {
	cmd.OnAbort = coverage.Flush
	cmd.AbortOnCtrlC()

	exitCode := run()
	coverage.Flush()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

// isDebugEnabled reports whether AUX4_DEBUG is set, matching the convention
// already used by output.DebugOutput.
func isDebugEnabled() bool {
	return os.Getenv("AUX4_DEBUG") == "true"
}

// formatCrashReport turns a recovered panic value plus its stack trace into
// the message shown to the user. It is a backstop for bugs that slipped past
// every other guard — it must never look like the command's own output, so
// the user reports it instead of assuming their command failed. The full Go
// stack trace is only ever printed when AUX4_DEBUG=true, so diagnosability
// isn't lost, but the default path stays clean for every other user.
func formatCrashReport(recovered any, stack []byte, debugEnabled bool) string {
	message := fmt.Sprintf(
		"aux4 hit an internal error and could not continue: %v\n"+
			"This is a bug in aux4 itself, not in your command — please report it at https://github.com/aux4/aux4/issues with what you ran.",
		recovered,
	)

	if debugEnabled {
		message += "\n\n" + string(stack)
	} else {
		message += "\nRe-run with AUX4_DEBUG=true for the full stack trace."
	}

	return message
}

// reportCrash prints a formatted crash report to stderr and returns the exit
// code the process should use. Kept separate from the recover() call site so
// it is unit-testable without needing to trigger a real panic.
func reportCrash(recovered any, stack []byte) int {
	output.Out(output.StdErr).Println(output.Red(formatCrashReport(recovered, stack, isDebugEnabled())))
	return crashExitCode
}

func run() (exitCode int) {
	// Top-level backstop: aux4 core must never crash out to a raw Go stack
	// trace. Every known panic source is fixed at its root cause (see
	// CORE-036/CORE-037), but this catches whatever is still missed, so the
	// user always gets a clean, actionable message and a non-zero exit code
	// instead of a stack trace. It is a backstop, not a substitute for fixing
	// root causes — a caught panic here is still a bug worth reporting.
	defer func() {
		if r := recover(); r != nil {
			exitCode = reportCrash(r, debug.Stack())
		}
	}()

	return runCommandFn()
}

// runCommandFn is a package-level indirection over runCommand purely so tests
// can substitute a panicking stand-in to exercise the recover() wiring in
// run() without needing a real, hard-to-trigger internal panic.
var runCommandFn = runCommand

func runCommand() int {
	// Handle daemon server mode (launched by `aux4 aux4 daemon start`)
	if len(os.Args) >= 3 && os.Args[1] == "-daemon-server" {
		socketPath := os.Args[2]
		startDaemonServer(socketPath)
		return 0
	}

	// The --noDaemon flag is an aux4-level flag consumed here. It is stripped
	// from the raw argv before command parsing so it never leaks into actions
	// or the command's parameters, and it never consumes the following argument
	// (so `aux4 --noDaemon mcp` keeps `mcp` as the command). It applies only to
	// this invocation and is not propagated to any subprocess.
	noDaemon, args := daemon.ExtractNoDaemonFlag(os.Args[1:])

	aux4Params, actions, params := param.ParseArgs(args)

	output.SetPrettify(params.IsEnabled(output.PrettifyParameter))

	// Check if daemon is running and forward the command, unless the user opted
	// out (--noDaemon or AUX4_NO_DAEMON=1) so a long-running server runs
	// directly instead of holding the daemon's global mutex. Also run directly
	// when AUX4_SECURITY is set: the daemon started with its own environment and
	// cannot see this invocation's policy, so forwarding would bypass it.
	if !isDaemonCommand(actions) && !daemon.SkipForwarding(noDaemon) && os.Getenv("AUX4_SECURITY") == "" {
		socketPath := daemon.FindSocketPath(".")
		if conn := daemon.Connect(socketPath); conn != nil {
			return daemon.Forward(conn, args)
		}
	}

	library := engine.LocalLibrary()

	if err := library.Load("", "aux4", []byte(aux4.DefaultAux4())); err != nil {
		output.Out(output.StdErr).Println(err)
		return err.(core.Aux4Error).ExitCode
	}

	var aux4Files = config.ListAux4Files(".", aux4Params)

	for _, aux4File := range aux4Files {
		if err := library.LoadFile(aux4File); err != nil {
			output.Out(output.StdErr).Println(output.Red("Error loading file"), output.Red(aux4File), output.Red(err))
		}
	}

	registry := engine.CreateVirtualExecutorRegistry()
	registry.RegisterExecutor("aux4.version", &executor.Aux4VersionExecutor{})
	registry.RegisterExecutor("aux4.shell", &executor.Aux4ShellExecutor{})
	registry.RegisterExecutor("aux4.autoinstall", &executor.Aux4AutoInstallExecutor{})
	registry.RegisterExecutor("aux4.completion", &executor.Aux4CompletionExecutor{})
	registry.RegisterExecutor("aux4.autocomplete", &executor.Aux4AutocompleteExecutor{})
	registry.RegisterExecutor("aux4.hooks", &executor.Aux4HooksExecutor{})
	registry.RegisterExecutor("aux4:daemon.start", &executor.Aux4DaemonStartExecutor{})
	registry.RegisterExecutor("aux4:daemon.stop", &executor.Aux4DaemonStopExecutor{})
	registry.RegisterExecutor("aux4:daemon.status", &executor.Aux4DaemonStatusExecutor{})

	env, err := engine.InitializeVirtualEnvironment(library, registry)
	if err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return err.(core.Aux4Error).ExitCode
	}

	// Keep the invocation as typed, for hooks. params still holds argv only at this
	// point — MainExecute injects packageDir/aux4HomeDir/configDir later, and config.yaml
	// values are resolved lazily on Get — so this captures the command line and nothing else.
	env.OriginalActions = actions
	env.OriginalParams = params.Clone()

	// Resolve the command-exposure policy once, before anything runs, and freeze
	// it on the shared env so profile routing and nested in-process calls inherit
	// it. Then block the top-level invocation if the policy forbids it.
	policy, err := resolveSecurityPolicy(&params)
	if err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return 1
	}
	env.Security = policy

	if err := enforceSecurity(policy, actions); err != nil {
		if aux4Err, ok := err.(core.Aux4Error); ok {
			if aux4Err.Message != "" {
				output.Out(output.StdErr).Println(output.Red(aux4Err.Message))
			}
			return aux4Err.ExitCode
		}
		return 1
	}

	if err := executor.MainExecute(env, actions, &params); err != nil {
		if aux4Err, ok := err.(core.Aux4Error); ok {
			if aux4Err.Message != "" {
				output.Out(output.StdErr).Println(output.Red(aux4Err.Message))
			}
			return aux4Err.ExitCode
		}
		return 1
	}

	return 0
}

// isDaemonCommand returns true if the command is managing the daemon itself
// (we don't want to forward daemon start/stop/status to the daemon)
func isDaemonCommand(actions []string) bool {
	if len(actions) >= 2 && actions[0] == "aux4" && actions[1] == "daemon" {
		return true
	}
	return false
}

// runMainExecuteRecovered runs executor.MainExecute with its own recover, so a
// panic inside a daemon-served command surfaces as a panicValue instead of
// unwinding into the daemon's serving loop. Kept separate from the call site
// so the surrounding pipe/stdio cleanup in executeFn always runs, whether the
// command errored, panicked, or completed normally.
func runMainExecuteRecovered(env *engine.VirtualEnvironment, actions []string, params *param.Parameters) (err error, panicValue any, stack []byte) {
	defer func() {
		if r := recover(); r != nil {
			panicValue = r
			stack = debug.Stack()
		}
	}()

	err = executor.MainExecute(env, actions, params)
	return
}

// startDaemonServer builds the environment and starts the daemon server process
func startDaemonServer(socketPath string) {
	library, registry := buildDaemonLibrary()
	if library == nil {
		os.Exit(1)
	}

	executeFn := func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
		// Parse and execute. Strip --noDaemon defensively so it never reaches
		// the command even if a client forwarded it into the daemon.
		_, cleanArgs := daemon.ExtractNoDaemonFlag(args)
		_, actions, params := param.ParseArgs(cleanArgs)

		// Build a FRESH environment per request from the shared, already-parsed
		// library. The expensive work (parsing global.aux4 + package .aux4 files)
		// happened once in buildDaemonLibrary; this only rebuilds the in-memory
		// profile maps, so it stays warm. A fresh env gives each request its own
		// mutable state (CurrentProfile etc.) — essential for re-entrancy: a
		// nested command must not corrupt the profile pointer of the parent that
		// is parked waiting for it (see daemon.Server.executeCommand).
		env, err := engine.InitializeVirtualEnvironment(library, registry)
		if err != nil {
			msg := "failed to initialize aux4 environment"
			if aux4Err, ok := err.(core.Aux4Error); ok && aux4Err.Message != "" {
				msg = aux4Err.Message
			}
			io.WriteString(stderr, msg+"\n")
			return 1
		}

		// Redirect os.Stdout, os.Stderr, and os.Stdin. This is process-global,
		// but the server only ever runs one execution ACTIVELY at a time (a
		// nested call runs while its parent is parked in the shell-out that
		// spawned it), and the save/restore below is stack-correct, so the
		// redirection nests safely.
		origStdout := os.Stdout
		origStderr := os.Stderr
		origStdin := os.Stdin

		stdoutR, stdoutW, _ := os.Pipe()
		stderrR, stderrW, _ := os.Pipe()
		stdinR, stdinW, _ := os.Pipe()

		os.Stdout = stdoutW
		os.Stderr = stderrW
		os.Stdin = stdinR

		// The daemon serves each request with the client's environment, so the
		// color decision has to be taken again per request instead of using the
		// one cached when the daemon started.
		output.ResolveColor()
		output.SetPrettify(params.IsEnabled(output.PrettifyParameter))

		// Stream pipe output to the writers
		done := make(chan struct{}, 3)
		go func() {
			io.Copy(stdout, stdoutR)
			done <- struct{}{}
		}()
		go func() {
			io.Copy(stderr, stderrR)
			done <- struct{}{}
		}()
		// Pipe client stdin into the command's stdin
		go func() {
			io.Copy(stdinW, stdin)
			stdinW.Close()
			done <- struct{}{}
		}()

		// Resolve and enforce the exposure policy per request. AUX4_SECURITY-based
		// policies never reach here (the client runs those directly instead of
		// forwarding), so only config/param policies are resolved in the daemon.
		policy, perr := resolveSecurityPolicy(&params)
		if perr == nil {
			env.Security = policy
			if serr := enforceSecurity(policy, actions); serr != nil {
				if aux4Err, ok := serr.(core.Aux4Error); ok {
					if aux4Err.Message != "" {
						stderrW.WriteString(aux4Err.Message + "\n")
					}
					stdoutW.Close()
					stderrW.Close()
					<-done
					<-done
					os.Stdout = origStdout
					os.Stderr = origStderr
					os.Stdin = origStdin
					return aux4Err.ExitCode
				}
			}
		}

		// A panic here must never bring down the daemon process — it is shared
		// by every connected client, not just the one that triggered it. This
		// mirrors the top-level recover() in run(), but reports back over the
		// client's own stderr pipe instead of the daemon process's stderr.
		exitCode := 0
		mainErr, panicValue, stack := runMainExecuteRecovered(env, actions, &params)
		if panicValue != nil {
			stderrW.WriteString(formatCrashReport(panicValue, stack, isDebugEnabled()) + "\n")
			exitCode = crashExitCode
		} else if mainErr != nil {
			if aux4Err, ok := mainErr.(core.Aux4Error); ok {
				if aux4Err.Message != "" {
					stderrW.WriteString(aux4Err.Message + "\n")
				}
				exitCode = aux4Err.ExitCode
			} else {
				exitCode = 1
			}
		}

		// Close write ends and wait for readers to finish
		stdoutW.Close()
		stderrW.Close()
		<-done
		<-done

		// Restore
		os.Stdout = origStdout
		os.Stderr = origStderr
		os.Stdin = origStdin

		return exitCode
	}

	if err := daemon.StartServer(socketPath, executeFn); err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		os.Exit(1)
	}
}

// buildDaemonLibrary loads global.aux4 + the local .aux4 files ONCE (the
// expensive parse) and returns the library + executor registry. The daemon then
// builds a fresh VirtualEnvironment from these per request, so the parse cost is
// paid once at startup while each request gets isolated mutable state.
func buildDaemonLibrary() (*engine.Library, *engine.VirtualExecutorRegisty) {
	library := engine.LocalLibrary()

	if err := library.Load("", "aux4", []byte(aux4.DefaultAux4())); err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return nil, nil
	}

	aux4Params := param.Aux4Parameters{}
	aux4Files := config.ListAux4Files(".", aux4Params)

	for _, aux4File := range aux4Files {
		if err := library.LoadFile(aux4File); err != nil {
			output.Out(output.StdErr).Println(output.Red("Error loading file"), output.Red(aux4File), output.Red(err))
		}
	}

	registry := engine.CreateVirtualExecutorRegistry()
	registry.RegisterExecutor("aux4.version", &executor.Aux4VersionExecutor{})
	registry.RegisterExecutor("aux4.shell", &executor.Aux4ShellExecutor{})
	registry.RegisterExecutor("aux4.autoinstall", &executor.Aux4AutoInstallExecutor{})
	registry.RegisterExecutor("aux4.completion", &executor.Aux4CompletionExecutor{})
	registry.RegisterExecutor("aux4.autocomplete", &executor.Aux4AutocompleteExecutor{})
	registry.RegisterExecutor("aux4.hooks", &executor.Aux4HooksExecutor{})
	registry.RegisterExecutor("aux4:daemon.start", &executor.Aux4DaemonStartExecutor{})
	registry.RegisterExecutor("aux4:daemon.stop", &executor.Aux4DaemonStopExecutor{})
	registry.RegisterExecutor("aux4:daemon.status", &executor.Aux4DaemonStatusExecutor{})

	// Validate the environment builds cleanly at startup (fail fast) rather than
	// only discovering a broken package on the first request.
	if _, err := engine.InitializeVirtualEnvironment(library, registry); err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return nil, nil
	}

	return library, registry
}
