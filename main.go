package main

import (
	"io"
	"os"

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

func main() {
	cmd.OnAbort = coverage.Flush
	cmd.AbortOnCtrlC()

	exitCode := run()
	coverage.Flush()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func run() int {
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
	// directly instead of holding the daemon's global mutex.
	if !isDaemonCommand(actions) && !daemon.SkipForwarding(noDaemon) {
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

		exitCode := 0
		if err := executor.MainExecute(env, actions, &params); err != nil {
			if aux4Err, ok := err.(core.Aux4Error); ok {
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
