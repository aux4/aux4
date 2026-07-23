package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/output"
)

var commandsAvailable = map[string]bool{}

// redirectTokenPattern matches the aux4 output-redirect shortcuts as standalone
// shell tokens. It requires a separator (start, whitespace, ';', '&', '(') before
// the '>' and a separator (end, whitespace, ';', '&', '|', ')') after the token,
// so it never touches a real redirect like `2>ignore` or `>ignore.log`.
var redirectTokenPattern = regexp.MustCompile(`(^|[\s;&(])>ignore(Error|Output)?($|[\s;&|)])`)

// expandRedirects rewrites the aux4 redirect shortcuts into shell redirections
// just before the command reaches the shell:
//
//	>ignore        -> >/dev/null 2>&1   (discard stdout and stderr)
//	>ignoreError   -> 2>/dev/null       (discard stderr)
//	>ignoreOutput  -> >/dev/null        (discard stdout)
func expandRedirects(instruction string) string {
	return redirectTokenPattern.ReplaceAllStringFunc(instruction, func(match string) string {
		groups := redirectTokenPattern.FindStringSubmatch(match)
		lead, kind, trail := groups[1], groups[2], groups[3]

		var replacement string
		switch kind {
		case "Error":
			replacement = "2>/dev/null"
		case "Output":
			replacement = ">/dev/null"
		default:
			replacement = ">/dev/null 2>&1"
		}
		return lead + replacement + trail
	})
}

var OnAbort func()

func AbortOnCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if OnAbort != nil {
			OnAbort()
		}
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

func ExecuteCommandLineNoOutputWithStdIn(instruction string) (string, string, error) {
	return executeCommand(instruction, false, true)
}

func ExecuteCommandWithPipedStdin(instruction string, stdinData string) error {
	var shell = os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	instruction = expandRedirects(instruction)
	cmd := exec.Command(shell, "-c", instruction)
	cmd.Stdin = bytes.NewBufferString(stdinData)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return core.Aux4Error{
				ExitCode: exitError.ExitCode(),
				Cause:    err,
			}
		}
		return core.InternalError("Error executing render command", err)
	}

	return nil
}

func executeCommand(instruction string, withStdOut bool, withStdIn bool) (string, string, error) {
	var cmd *exec.Cmd

	var shell = os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	instruction = expandRedirects(instruction)
	cmd = exec.Command(shell, "-c", instruction)
	cmd.Env = append(os.Environ(), "CLICOLOR_FORCE=1")

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

