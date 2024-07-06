package main

import (
	"bytes"
	"os"
	"os/exec"
)

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
			return stdout.String(), stderr.String(), Aux4Error{
				Message:  exitError.Error(),
				ExitCode: exitError.ExitCode(),
				Cause:    err,
			}
		}
		return stdout.String(), stderr.String(), InternalError("Error waiting the command execute", err)
	}
	return stdout.String(), stderr.String(), nil
}

func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
