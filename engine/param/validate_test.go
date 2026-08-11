package param

import (
	"testing"

	"aux4.dev/aux4/core"
)

// commandWith builds a command declaring the given variable names, for the
// unknown-parameter check.
func commandWith(names ...string) core.Command {
	vars := make([]*core.CommandHelpVariable, 0, len(names))
	for _, n := range names {
		vars = append(vars, &core.CommandHelpVariable{Name: n})
	}
	return core.Command{Help: &core.CommandHelp{Variables: vars}}
}

func paramsWith(names ...string) *Parameters {
	p := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	for _, n := range names {
		p.Update(n, "x")
	}
	return p
}

// A flag that differs from a declared variable only by case/separators is a typo
// and must be caught, whichever form the user reached for.
func TestUnknownParameterIsCaughtForEachSeparatorForm(t *testing.T) {
	cmd := commandWith("customerId")
	for _, given := range []string{"customer-id", "customer_id", "CUSTOMER_ID", "customerid", "CustomerId"} {
		err := ValidateParameterNames(cmd, paramsWith(given))
		if err == nil {
			t.Fatalf("expected --%s to be rejected as a misspelling of customerId", given)
		}
		aux4Err, ok := err.(core.Aux4Error)
		if !ok || aux4Err.Message == "" {
			t.Fatalf("expected an Aux4Error naming the suggestion, got %v", err)
		}
	}
}

// The error names the variable the user meant.
func TestUnknownParameterErrorSuggestsTheDeclaredName(t *testing.T) {
	err := ValidateParameterNames(commandWith("eventId"), paramsWith("event-id"))
	if err == nil {
		t.Fatal("expected --event-id to be rejected")
	}
	want := "Unknown parameter --event-id, did you mean --eventId?"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

// The exact declared name is always accepted.
func TestExactDeclaredNameIsAccepted(t *testing.T) {
	if err := ValidateParameterNames(commandWith("eventId", "calendarId"), paramsWith("eventId", "calendarId")); err != nil {
		t.Fatalf("declared names must be accepted, got %v", err)
	}
}

// A genuinely different, undeclared parameter is left alone — passing undeclared
// variables is a supported feature, so it must not error.
func TestUndeclaredParameterThatMatchesNothingIsAllowed(t *testing.T) {
	if err := ValidateParameterNames(commandWith("eventId"), paramsWith("somethingElse")); err != nil {
		t.Fatalf("undeclared parameter that matches nothing must pass, got %v", err)
	}
}

// A command that declares no variables accepts anything (unchanged behaviour).
func TestCommandWithNoVariablesAcceptsAnything(t *testing.T) {
	cmd := core.Command{Help: &core.CommandHelp{}}
	if err := ValidateParameterNames(cmd, paramsWith("anything", "at-all")); err != nil {
		t.Fatalf("a command with no declared variables must accept anything, got %v", err)
	}
	if err := ValidateParameterNames(core.Command{}, paramsWith("anything")); err != nil {
		t.Fatalf("a command with no help must accept anything, got %v", err)
	}
}

// Core-injected parameters (set by aux4 itself) and reserved parameters must
// never be flagged, even against a command that declares a colliding name.
func TestInjectedAndReservedParametersAreSkipped(t *testing.T) {
	cmd := commandWith("packageDir", "prettify") // even if a command declared these
	p := paramsWith("aux4HomeDir", "packageDir", "configDir", "prettify", "__command")
	if err := ValidateParameterNames(cmd, p); err != nil {
		t.Fatalf("injected/reserved parameters must be skipped, got %v", err)
	}
}

// A nested object field (--obj.field, stored under the base name) is accepted
// when the base variable is declared.
func TestNestedObjectFieldOnDeclaredBaseIsAccepted(t *testing.T) {
	p := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	p.UpdateField("body.subject", "hi") // stored under base "body"
	if err := ValidateParameterNames(commandWith("body"), p); err != nil {
		t.Fatalf("nested field on a declared base must be accepted, got %v", err)
	}
}

// The real case from the evals: --event-id where the calendar delete declares eventId.
func TestCalendarDeleteEventIdTypoIsCaught(t *testing.T) {
	cmd := commandWith("calendarId", "eventId", "tokenFile", "apiUrl")
	if err := ValidateParameterNames(cmd, paramsWith("event-id")); err == nil {
		t.Fatal("expected --event-id to be rejected in favour of --eventId")
	}
	// and the correct form still passes
	if err := ValidateParameterNames(cmd, paramsWith("eventId")); err != nil {
		t.Fatalf("--eventId must be accepted, got %v", err)
	}
}

func TestNormalizeParameterName(t *testing.T) {
	cases := map[string]string{
		"customerId":  "customerid",
		"customer-id": "customerid",
		"customer_id": "customerid",
		"CUSTOMER-ID": "customerid",
		"event.id":    "eventid",
		"threadId":    "threadid",
	}
	for in, want := range cases {
		if got := normalizeParameterName(in); got != want {
			t.Fatalf("normalizeParameterName(%q) = %q, want %q", in, got, want)
		}
	}
}
