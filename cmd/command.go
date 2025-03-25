package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/output"
)

var commandsAvailable = map[string]bool{}

func AbortOnCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		output.Out(output.StdErr).Println("Process aborted")
		os.Exit(130)
	}()
}

func ExecuteCommandLine(instruction string) (string, string, error) {
	return executeCommand(instruction, true, false)
}

func ExecuteCommandLineNoOutput(instruction string) (string, string, error) {
	return executeCommand(instruction, false, false)
}

func ExecuteCommandLineWithStdIn(instruction string) (string, string, error) {
	return executeCommand(instruction, true, true)
}

func executeCommand(instruction string, withStdOut bool, withStdIn bool) (string, string, error) {
	var cmd *exec.Cmd

	cmd = exec.Command("bash", "-c", instruction)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if withStdOut {
		cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
		cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if withStdIn {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return stdout.String(), stderr.String(), core.Aux4Error{
				ExitCode: exitError.ExitCode(),
				Cause:    err,
			}
		}
		return stdout.String(), stderr.String(), core.InternalError("Error waiting the command execute", err)
	}

	return stdout.String(), stderr.String(), nil
}

