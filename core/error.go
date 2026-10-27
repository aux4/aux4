package core

import (
  "fmt"
  "strings"
)

type Aux4Error struct {
  Message string
  ExitCode int
  Cause error
}

func (err Aux4Error) Error() string {
  if err.Cause == nil {
    return err.Message
  }
  return fmt.Sprintf("%s: %v", err.Message, err.Cause)
}

func InternalError(message string, cause error) Aux4Error {
  return Aux4Error{
    Message: message,
    ExitCode: 1,
    Cause: cause,
  }
}

func VariableNotFoundError(variableName string) Aux4Error {
	return Aux4Error{
		Message: fmt.Sprintf("Variable not found: %s", variableName),
		ExitCode: 1,
	}
}

// UnknownParameterError is returned when a passed parameter is the same as a declared
// variable once case and separators are ignored (e.g. --customer-id for a declared
// customerId). It names the parameter that was meant, so a mistyped flag is corrected
// rather than silently ignored.
func UnknownParameterError(given string, suggestion string) Aux4Error {
	return Aux4Error{
		Message:  fmt.Sprintf("Unknown parameter --%s, did you mean --%s?", given, suggestion),
		ExitCode: 1,
	}
}

func UserAbortedError() Aux4Error {
  return Aux4Error{
    Message: "User aborted",
    ExitCode: 130,
  }
}

func AccessDeniedError(cause error) Aux4Error {
  return Aux4Error{
    Message: "Access denied",
    ExitCode: 2,
    Cause: cause,
  }
} 

func CommandNotFoundError(message string) Aux4Error {
  return Aux4Error{
    Message: fmt.Sprintf("Command not found: %s", message),
    ExitCode: 127,
  }
}

// CommandNotFoundDetailedError builds an actionable "command not found" message for
// a profile-tree lookup: it names the command that wasn't found and the path typed
// to reach the level that failed, lists the sibling commands that level actually
// has, and -- when the caller supplies one -- suggests the command they probably
// meant. The first line always starts with "Command not found:" so any existing
// caller or test matching on that prefix keeps working.
func CommandNotFoundDetailedError(name string, prefix string, available []string, suggestion string) Aux4Error {
	var b strings.Builder

	if prefix != "" {
		fmt.Fprintf(&b, "Command not found: `%s` under `aux4 %s`.", name, prefix)
	} else {
		fmt.Fprintf(&b, "Command not found: `%s`.", name)
	}

	if len(available) > 0 {
		fmt.Fprintf(&b, "\nAvailable commands here: %s", strings.Join(available, ", "))
	}

	if suggestion != "" {
		fmt.Fprintf(&b, "\nDid you mean: %s", suggestion)
	}

	return Aux4Error{
		Message:  b.String(),
		ExitCode: 127,
	}
}

func PathNotFoundError() Aux4Error {
  return Aux4Error{
    Message: "Path not found",
    ExitCode: 1,
  }
}
