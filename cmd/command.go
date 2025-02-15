package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

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
	// isExpression := strings.Contains(instruction, "|") || strings.Contains(instruction, "&") || strings.Contains(instruction, ">") || strings.Contains(instruction, "<") || strings.Contains(instruction, "=")

	var cmd *exec.Cmd
	// if isExpression {
	cmd = exec.Command("bash", "-c", instruction)
	// } else {
	// 	args := splitCommandLineIntoArgs(instruction)
	//    cmd = exec.Command(args[0], args[1:]...)
	// }

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
	if _, exists := commandsAvailable[command]; !exists {
		_, err := exec.LookPath(command)
		commandsAvailable[command] = err == nil
	}
	return commandsAvailable[command]
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

	return args
}
