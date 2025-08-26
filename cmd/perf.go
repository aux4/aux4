package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"aux4.dev/aux4/core"
)

// extractArgs parses a command line string into arguments, handling quoted strings
func extractArgs(instruction string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false
	for _, r := range instruction {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
		} else if r == '\\' {
			escapeNext = true
		} else if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if unicode.IsSpace(r) && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// Reuse buffers to reduce GC pressure
var bufferPool = struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}{
	stdout: &bytes.Buffer{},
	stderr: &bytes.Buffer{},
}

func ExecuteCommandLinePerf(instruction string) (string, string, time.Duration, error) {
	// Check if command needs shell features BEFORE timing starts
	needsShell := false
	for i := 0; i < len(instruction); i++ {
		c := instruction[i]
		if c == '|' || c == '>' || c == '<' || c == '&' || c == ';' || c == '$' || c == '*' || c == '?' || c == '[' || c == ']' || c == '{' || c == '}' || c == '~' {
			needsShell = true
			break
		}
	}

	// Pre-parse args for direct execution if needed
	var args []string
	if !needsShell {
		args = extractArgs(instruction)
		if len(args) == 0 {
			return "", "", 0, core.Aux4Error{ExitCode: 1, Message: "empty command"}
		}
	}

	// Reuse pre-allocated buffers to minimize GC
	stdout := bufferPool.stdout
	stderr := bufferPool.stderr
	stdout.Reset()
	stderr.Reset()

	// Start high-precision timing
	startTimeNano := time.Now().UnixNano()

	var cmd *exec.Cmd
	if needsShell {
		// Use shell for complex commands
		cmd = exec.Command("sh", "-c", instruction)
	} else {
		// Execute directly for simple commands
		cmd = exec.Command(args[0], args[1:]...)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	endTimeNano := time.Now().UnixNano()
	executionTime := time.Duration(endTimeNano - startTimeNano)

	// Copy strings to avoid holding buffer references
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return stdoutStr, stderrStr, executionTime, core.Aux4Error{
				ExitCode: exitError.ExitCode(),
				Cause:    err,
			}
		}
		return stdoutStr, stderrStr, executionTime, core.InternalError("Error executing command", err)
	}

	return stdoutStr, stderrStr, executionTime, nil
}