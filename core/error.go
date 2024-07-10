package core

import (
  "fmt"
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
  return fmt.Sprintf("%s %v", err.Message, err.Cause)
}

func InternalError(message string, cause error) Aux4Error {
  return Aux4Error{
    Message: message,
    ExitCode: 1,
    Cause: cause,
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

func PathNotFoundError() Aux4Error {
  return Aux4Error{
    Message: "Path not found",
    ExitCode: 1,
  }
}
