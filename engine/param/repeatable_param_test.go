package param

import (
	"os/exec"
	"strings"
	"testing"

	"aux4.dev/aux4/core"
)

// CORE-044: param(name**) — the repeatable-flag helper used for multiple:true
// variables (param(state**), param(options**), ...) — built each "--flag
// 'value'" with a bare fmt.Sprintf and no escaping, unlike every other
// parameter function (value(), param(name) without **, values(), params()),
// which all go through escapeValue. A value containing a single quote broke
// out of the '...' wrapper and ran arbitrary shell commands.
//
// injectMultiple mirrors real CLI parsing (ParseArgs) for a repeatable flag:
// --name v1 --name v2 ...
func injectMultiple(t *testing.T, instruction string, name string, values ...string) string {
	t.Helper()
	args := []string{}
	for _, v := range values {
		args = append(args, "--"+name, v)
	}
	_, actions, params := ParseArgs(args)
	result, err := InjectParameters(core.Command{}, instruction, actions, &params)
	if err != nil {
		t.Fatalf("InjectParameters(%q) error: %v", instruction, err)
	}
	return result
}

// runShell executes the resolved instruction exactly like aux4 does at
// runtime (sh -c) and reports whether it actually ran more than the intended
// command — i.e. whether the payload escaped the quotes.
func runShell(t *testing.T, instruction string) (string, error) {
	t.Helper()
	cmd := exec.Command("sh", "-c", instruction)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestRepeatableParamEscapesSingleQuoteInjection(t *testing.T) {
	marker := t.TempDir() + "/pwned"
	payload := "'; touch " + marker + "; '"

	got := injectMultiple(t, "echo start param(state**) echo end", "state", payload)

	runShell(t, got)

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via param(state**): marker file was created\n  instruction: %s\n  resolved:    %s", "echo start param(state**) echo end", got)
	}
}

func TestRepeatableParamEscapesVariousPayloads(t *testing.T) {
	marker := t.TempDir() + "/pwned2"

	payloads := []string{
		"'",
		`"`,
		"$(touch " + marker + ")",
		"`touch " + marker + "`",
		"; touch " + marker + " ;",
		"*.txt",
		"line1\nline2",
		"", // empty value
		"héllo wörld 日本語 🎉",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			got := injectMultiple(t, "echo noop param(state**)", "state", payload)
			runShell(t, got)
		})
	}

	if _, err := exec.Command("test", "-e", marker).CombinedOutput(); err == nil {
		t.Fatalf("command injection via param(state**): marker file was created from one of the payloads")
	}
}

func TestRepeatableParamEmitsOneFlagPerValue(t *testing.T) {
	got := injectMultiple(t, "run param(state**)", "state", "a", "b")
	want := "run --state 'a' --state 'b'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Backward compatibility: ordinary values (spaces, '=', JSON-looking text)
// must still come through byte-identical to what the target program
// received before the fix — only an embedded single quote now takes a
// different (and correct) path.
func TestRepeatableParamNormalValuesUnchanged(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"plain", "hello"},
		{"spaces", "hello world"},
		{"equals", "key=value"},
		{"json-ish", `{"a":1}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := injectMultiple(t, "run param(state**)", "state", test.value)
			want := "run --state '" + test.value + "'"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}

			out, err := runShell(t, "printf '%s' "+mustSingleFlag(t, got))
			if err != nil {
				t.Fatalf("resolved instruction failed to run: %v (%s)", err, out)
			}
			if out != test.value {
				t.Errorf("value did not survive the shell byte-identical: got %q want %q", out, test.value)
			}
		})
	}
}

// mustSingleFlag extracts the quoted value portion from "run --state '...'"
// so the test can feed exactly that token to printf and confirm the target
// program receives the value byte-identical.
func mustSingleFlag(t *testing.T, resolved string) string {
	t.Helper()
	idx := strings.Index(resolved, "'")
	if idx == -1 {
		t.Fatalf("no quoted value found in %q", resolved)
	}
	return resolved[idx:]
}
