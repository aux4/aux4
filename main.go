package main

import (
	"io"
	"os"

	"aux4.dev/aux4/aux4"
	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/config"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/daemon"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/executor"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

func main() {
	cmd.AbortOnCtrlC()

	// Handle daemon server mode (launched by `aux4 aux4 daemon start`)
	if len(os.Args) >= 3 && os.Args[1] == "-daemon-server" {
		socketPath := os.Args[2]
		startDaemonServer(socketPath)
		return
	}

	aux4Params, actions, params := param.ParseArgs(os.Args[1:])

	// Check if daemon is running and forward the command
	if !isDaemonCommand(actions) {
		socketPath := daemon.FindSocketPath(".")
		if conn := daemon.Connect(socketPath); conn != nil {
			exitCode := daemon.Forward(conn, os.Args[1:])
			os.Exit(exitCode)
		}
	}

	library := engine.LocalLibrary()

	if err := library.Load("", "aux4", []byte(aux4.DefaultAux4())); err != nil {
		output.Out(output.StdErr).Println(err)
		os.Exit(err.(core.Aux4Error).ExitCode)
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
		os.Exit(err.(core.Aux4Error).ExitCode)
	}

	if err := executor.MainExecute(env, actions, &params); err != nil {
		if aux4Err, ok := err.(core.Aux4Error); ok {
			if aux4Err.Message != "" {
				output.Out(output.StdErr).Println(output.Red(aux4Err.Message))
			}
			os.Exit(aux4Err.ExitCode)
		} else {
			os.Exit(1)
		}
	}
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
	env := buildDaemonEnvironment()
	if env == nil {
		os.Exit(1)
	}

	executeFn := func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
		// Redirect os.Stdout, os.Stderr, and os.Stdin
		origStdout := os.Stdout
		origStderr := os.Stderr
		origStdin := os.Stdin

		stdoutR, stdoutW, _ := os.Pipe()
		stderrR, stderrW, _ := os.Pipe()
		stdinR, stdinW, _ := os.Pipe()

		os.Stdout = stdoutW
		os.Stderr = stderrW
		os.Stdin = stdinR

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

		// Parse and execute
		_, actions, params := param.ParseArgs(args)
		env.SetProfile("main")

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

func buildDaemonEnvironment() *engine.VirtualEnvironment {
	library := engine.LocalLibrary()

	if err := library.Load("", "aux4", []byte(aux4.DefaultAux4())); err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return nil
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

	env, err := engine.InitializeVirtualEnvironment(library, registry)
	if err != nil {
		output.Out(output.StdErr).Println(output.Red(err))
		return nil
	}

	return env
}
