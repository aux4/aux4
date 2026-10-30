package param

import (
	"os/exec"
	"strings"
	"testing"
)

// CORE-046 sweep: Parameters.String() — used by the `alias:` executor to
// forward every CLI parameter into another shell command, and to populate
// the `${__params}` hook variable — rebuilt each "--flag 'value'" via
// valueToString(value, escape=false) and then manually wrapped the raw,
// unescaped result in single quotes. A value containing a single quote
// broke out of that wrapper, exactly like the pre-CORE-044 param(name**)
// bug. Fixed by passing escape=true, matching every other param helper.

// paramsFromArgs builds a *Parameters the same way the engine does from
// real CLI args: --name1 v1 --name2 v2 ...
func paramsFromArgs(args ...string) *Parameters {
	_, _, params := ParseArgs(args)
	return &params
}

func TestParametersStringEscapesSingleQuoteInjection(t *testing.T) {
	marker := t.TempDir() + "/pwned"
	payload := "'; touch " + marker + "; '"

	params := paramsFromArgs("--value", payload)
	got := "echo start " + params.String() + " echo end"

	runShell(t, got)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via Parameters.String(): marker file was created\n  resolved: %s", got)
	}
}

func TestParametersStringEscapesVariousPayloads(t *testing.T) {
	marker := t.TempDir() + "/pwned2"

	payloads := []string{
		"'",
		`"`,
		"$(touch " + marker + ")",
		"`touch " + marker + "`",
		"; touch " + marker + " ;",
		"*.txt",
		"héllo wörld 日本語 🎉",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			params := paramsFromArgs("--value", payload)
			runShell(t, "echo noop "+params.String())
		})
	}

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via Parameters.String(): marker file was created from one of the payloads")
	}
}

// Backward compatibility: ordinary values must still round-trip
// byte-identical to the pre-fix output (`--name 'value'`) — this is what
// the existing hooks.test.md "${__params}" fixture asserts
// (`params=--name 'alice'`), so the fix must not change that case.
func TestParametersStringNormalValuesUnchanged(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"plain", "alice"},
		{"spaces", "hello world"},
		{"equals", "key=value"},
		{"json-ish", `{"a":1}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := paramsFromArgs("--name", test.value)
			got := params.String()
			want := "--name '" + test.value + "'"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// TestParametersStringOmitsReservedParams confirms the fix did not disturb
// the existing reserved-parameter filtering — e.g. "security" (the
// command-exposure policy, consumed by core itself) must never be
// forwarded/re-broadcast into the built string.
func TestParametersStringOmitsReservedParams(t *testing.T) {
	params := paramsFromArgs("--name", "alice", "--security", "deny:*")
	got := params.String()
	if strings.Contains(got, "security") {
		t.Errorf("expected reserved param 'security' to be omitted, got %q", got)
	}
	if got != "--name 'alice'" {
		t.Errorf("got %q, want %q", got, "--name 'alice'")
	}
}
