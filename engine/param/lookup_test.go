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

	message := missingParameterMessage(variable, nil)

	expected := "Missing required parameter --session: Session identifier"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

func TestMissingParameterMessageOmitsColonWhenNoHelpText(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "name", Text: ""}

	message := missingParameterMessage(variable, nil)

	expected := "Missing required parameter --name"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

// When an unexpected flag was passed that this command doesn't declare, it is named
// and suggested as the likely wrong name for the missing parameter.
func TestMissingParameterMessageSuggestsASingleUnexpectedFlag(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "messageId", Text: "The message ID to retrieve"}

	message := missingParameterMessage(variable, []string{"id"})

	expected := "Missing required parameter --messageId: The message ID to retrieve (you passed --id, which this command does not have — did you mean --messageId?)"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

// With no unexpected flags, the message is unchanged from before this feature.
func TestMissingParameterMessageUnchangedWhenNoUnexpectedFlags(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "messageId", Text: "The message ID to retrieve"}

	message := missingParameterMessage(variable, []string{})

	expected := "Missing required parameter --messageId: The message ID to retrieve"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

// With two or more unexpected flags, the one that contains (or is contained by) the
// missing name wins over a pure edit-distance pick.
func TestMissingParameterMessagePrefersSubstringMatchAmongUnexpectedFlags(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "messageId", Text: "The message ID to retrieve"}

	message := missingParameterMessage(variable, []string{"format", "id"})

	expected := "Missing required parameter --messageId: The message ID to retrieve (you passed --id, which this command does not have — did you mean --messageId?)"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

// With multiple unexpected flags and no substring relationship, edit distance picks
// the closest one.
func TestMissingParameterMessageFallsBackToEditDistanceAmongUnexpectedFlags(t *testing.T) {
	variable := core.CommandHelpVariable{Name: "calendarId", Text: "Calendar ID"}

	message := missingParameterMessage(variable, []string{"somethingUnrelatedXYZ", "calendraId"})

	expected := "Missing required parameter --calendarId: Calendar ID (you passed --calendraId, which this command does not have — did you mean --calendarId?)"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}
