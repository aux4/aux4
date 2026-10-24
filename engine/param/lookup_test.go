package param

import (
	"testing"

	"aux4.dev/aux4/core"
)

// When a required variable has no value and the prompt cannot be completed
// (no TTY: CI, pipes, AI agents), the message must tell the caller exactly
// which flag to pass, not describe the failed prompt attempt.
func TestMissingParameterMessageIncludesFlagAndHelpText(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "session", Text: "Session identifier"}

	message := missingParameterMessage(variable)

	expected := "Missing required parameter --session: Session identifier"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

func TestMissingParameterMessageOmitsColonWhenNoHelpText(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "name", Text: ""}

	message := missingParameterMessage(variable)

	expected := "Missing required parameter --name"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}
