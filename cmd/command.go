package cmd

import (
	"aux4/core"
	"aux4/output"
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
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
	isExpression := strings.Contains(instruction, "|") || strings.Contains(instruction, "&") || strings.Contains(instruction, ">") || strings.Contains(instruction, "<")

	var cmd *exec.Cmd

	if isExpression {
		cmd = exec.Command("bash", "-c", instruction)
	} else {
    args := splitCommandLineIntoArgs(instruction)
		cmd = exec.Command(args[0], args[1:]...)
	}

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

func splitCommandLineIntoArgs(instruction string) []string {
	argsRegex := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	instructionArgs := argsRegex.FindAllString(instruction, -1)

	args := []string{}

	for _, arg := range instructionArgs {
		unquoted := arg
		if strings.HasPrefix(arg, "\"") || strings.HasPrefix(arg, "'") {
			unquoted = arg[1 : len(arg)-1]
		}

		args = append(args, unquoted)
	}

	return strings.Fields(instruction)
}
