package main

import (
  "io"
  "os/exec"
)

func ExecuteCommandLine(instruction string) (string, string, error) {
	cmd := exec.Command("bash", "-c", instruction)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", InternalError("Error getting stdout from command", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", InternalError("Error getting stderr from command", err)
	}

	if err := cmd.Start(); err != nil {
		return "", "", InternalError("Error executing command", err)
	}

	errorData, err := io.ReadAll(stderr)
	if err != nil {
		return "", "", InternalError("Error reading the command error output", err)
	}

	data, err := io.ReadAll(stdout)
	if err != nil {
		return "", "", InternalError("Error reading the command output", err)
	}
	if err := cmd.Wait(); err != nil {
    exitError, ok := err.(*exec.ExitError)
    if ok {
      return string(data), string(errorData), Aux4Error{
        Message: exitError.Error(), 
        ExitCode: exitError.ExitCode(),
        Cause: err,
      }
    }
		return string(data), string(errorData), InternalError("Error waiting the command execute", err)
	}
  return string(data), string(errorData), nil 
}
