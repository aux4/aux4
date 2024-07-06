package cmd

import (
	"aux4/core"
	"aux4/output"
	"bytes"
	"os"
	"os/exec"
	"os/signal"
)

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
	return executeCommand(instruction, false)
}

func ExecuteCommandLineWithStdIn(instruction string) (string, string, error) {
	return executeCommand(instruction, true)
}

func executeCommand(instruction string, withStdIn bool) (string, string, error) {
	cmd := exec.Command("bash", "-c", instruction)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if withStdIn {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return stdout.String(), stderr.String(), core.Aux4Error{
				Message:  exitError.Error(),
				ExitCode: exitError.ExitCode(),
				Cause:    err,
			}
		}
		return stdout.String(), stderr.String(), core.InternalError("Error waiting the command execute", err)
	}
	return stdout.String(), stderr.String(), nil
}

func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
