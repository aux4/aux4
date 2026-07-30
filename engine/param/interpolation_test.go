package param

import (
	"strings"
	"testing"

	"aux4.dev/aux4/core"
)

// inject runs InjectParameters over an instruction with the given variables,
// mirroring how an execute line is resolved at runtime.
func inject(t *testing.T, instruction string, vars map[string]string) string {
	t.Helper()
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	for name, value := range vars {
		params.Update(name, value)
	}
	result, err := InjectParameters(core.Command{}, instruction, []string{}, params)
	if err != nil {
		t.Fatalf("InjectParameters(%q) error: %v", instruction, err)
	}
	return result
}

// A variable's VALUE is data, not code. Once a value has been substituted into
// the instruction it must never be scanned again by a later resolver, otherwise
// ordinary user text that happens to look like a function call is silently
// rewritten. Authored calls in the instruction itself must still resolve.

func TestValueContainingFunctionSyntaxIsNotResolved(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"nvl", "Use nvl(name, fallback) to pick the first non-empty value."},
		{"object", "Review the object(s) returned by the API."},
		{"objectWithAlias", "To rename a key use object(name:target) in your .aux4."},
		{"param", "To forward a flag, use param(name:alias) in your .aux4 file."},
		{"params", "Pass several with params(host, port, user)."},
		{"arg", "The first positional is arg(0) and the rest are args(1,2)."},
		{"path", "Absolute paths come from path(dir) at runtime."},
		{"value", "Quote a single one with value(database)."},
		{"values", "Emit positional args with values(a, b, c)."},
		{"nested", "Combine nvl(a, b) with object(x:y) and param(p:q) together."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := inject(t, "store ${description}", map[string]string{
				"description": test.value,
				"name":        "notes",
				"dir":         "/tmp",
				"database":    "app.db",
			})
			want := "store " + test.value
			if got != want {
				t.Errorf("value was re-interpolated\n  typed: %q\n  saved: %q", want, got)
			}
		})
	}
}

// The same must hold when the value reaches the instruction through values()
// rather than ${...} — that is how aux4/todo forwards a --description.
func TestValueSubstitutedByValuesIsNotResolved(t *testing.T) {
	text := "Review the object(s) and use nvl(name, other) here."
	got := inject(t, "run values(description)", map[string]string{
		"description": text,
		"name":        "notes",
	})
	if !strings.Contains(got, text) {
		t.Errorf("value pasted by values() was re-interpolated\n  typed: %q\n  got:   %q", text, got)
	}
}

// param() emits --flag 'value'; a value containing function syntax must survive.
func TestValueSubstitutedByParamIsNotResolved(t *testing.T) {
	text := "use nvl(a, b) here"
	got := inject(t, "run param(description)", map[string]string{
		"description": text,
		"a":           "AAA",
	})
	if !strings.Contains(got, text) {
		t.Errorf("value pasted by param() was re-interpolated\n  typed: %q\n  got:   %q", text, got)
	}
}

// Authored calls must keep working exactly as before — the fix must not
// disable resolution, only stop it from re-reading substituted values.
func TestAuthoredFunctionsStillResolve(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		vars        map[string]string
		want        string
	}{
		{"nvl picks first non-empty", "echo nvl(missing, fallback)", map[string]string{"fallback": "second"}, "echo second"},
		{"param emits flag", "run param(host)", map[string]string{"host": "db1"}, "run --host 'db1'"},
		{"param alias", "run param(host:server)", map[string]string{"host": "db1"}, "run --server 'db1'"},
		{"params emits several", "run params(host, port)", map[string]string{"host": "db1", "port": "3306"}, "run --host 'db1' --port '3306'"},
		{"value quotes", "run value(host)", map[string]string{"host": "db1"}, "run 'db1'"},
		{"values quotes several", "run values(host, port)", map[string]string{"host": "db1", "port": "3306"}, "run 'db1' '3306'"},
		{"object builds json", "echo object(host:server)", map[string]string{"host": "db1"}, `echo {"server":"db1"}`},
		{"braced variable", "echo ${host}", map[string]string{"host": "db1"}, "echo db1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := inject(t, test.instruction, test.vars); got != test.want {
				t.Errorf("InjectParameters(%q) = %q, want %q", test.instruction, got, test.want)
			}
		})
	}
}

// A value that itself looks like a variable reference must not be expanded
// either — the same data-is-not-code rule.
func TestValueContainingVariableReferenceIsNotExpanded(t *testing.T) {
	got := inject(t, "store ${description}", map[string]string{
		"description": "the template is ${item.file} verbatim",
		"item.file":   "SHOULD-NOT-APPEAR",
	})
	want := "store the template is ${item.file} verbatim"
	if got != want {
		t.Errorf("value containing a variable reference was expanded\n  got:  %q\n  want: %q", got, want)
	}
}

// Deliberate indirection and nested defaults must keep working: only the
// parentheses of a value are protected, never the value's use as an argument.
func TestDollarArgIndirectionStillWorks(t *testing.T) {
	p := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	p.Update("key", "id")
	p.Update("id", "123")
	got, _ := InjectParameters(core.Command{}, "set:x=object($key)", []string{}, p)
	if want := `set:x={"id":"123"}`; got != want {
		t.Errorf("object($key) = %q, want %q", got, want)
	}
}

func TestNestedVariableInDefaultStillExpands(t *testing.T) {
	p := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	p.Update("aux4HomeDir", "/home/u/.aux4.config")
	p.Update("db", "${aux4HomeDir}/backup/catalog.db")
	got, _ := InjectParameters(core.Command{}, "use ${db}", []string{}, p)
	if want := "use /home/u/.aux4.config/backup/catalog.db"; got != want {
		t.Errorf("nested default = %q, want %q", got, want)
	}
}

func TestValueWithParensStillQuotesForShell(t *testing.T) {
	p := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	p.Update("msg", "hello (world)")
	got, _ := InjectParameters(core.Command{}, "echo value(msg)", []string{}, p)
	if want := "echo 'hello (world)'"; got != want {
		t.Errorf("value with parens = %q, want %q", got, want)
	}
}

// Whitespace inside a value is data too. The trailing \s+ collapse exists so an
// unquoted value cannot split an instruction into several shell lines, but the
// quoting resolvers already emit shell-safe output ('...' with quotes escaped,
// or JSON), so their values must survive byte for byte.

const multiline = "first paragraph\n\nsecond paragraph\n  indented line"

func TestQuotedValueKeepsItsWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
	}{
		{"value", "run value(description)"},
		{"values", "run values(description)"},
		{"param", "run param(description)"},
		{"params", "run params(description)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := inject(t, test.instruction, map[string]string{"description": multiline})
			if !strings.Contains(got, multiline) {
				t.Errorf("quoted value lost its whitespace\n  typed: %q\n  got:   %q", multiline, got)
			}
		})
	}
}

// object() marshals to JSON, which escapes newlines itself — the value must
// still round-trip through the resolver unchanged.
func TestObjectValueKeepsItsWhitespace(t *testing.T) {
	got := inject(t, "echo object(description:text)", map[string]string{"description": multiline})
	want := `echo {"text":"first paragraph\n\nsecond paragraph\n  indented line"}`
	if got != want {
		t.Errorf("object() = %q, want %q", got, want)
	}
}

// A raw ${var} is NOT quoted by anything, so a newline there could split the
// instruction into two shell commands. That case keeps collapsing, as before.
func TestRawVariableStillCollapsesWhitespace(t *testing.T) {
	got := inject(t, "echo ${description}", map[string]string{"description": "line one\nline two"})
	if want := "echo line one line two"; got != want {
		t.Errorf("raw ${var} = %q, want %q", got, want)
	}
}

// Authored formatting in the execute string itself is still tidied up.
func TestAuthoredWhitespaceStillCollapses(t *testing.T) {
	got := inject(t, "echo   one\n\techo   two", map[string]string{})
	if want := "echo one echo two"; got != want {
		t.Errorf("authored whitespace = %q, want %q", got, want)
	}
}

// An empty result must not leave a placeholder behind. A placeholder is not
// whitespace, so it would hold apart the spaces around it and survive the
// collapse, leaving a gap where the value used to be.
func TestEmptyProtectedValueLeavesNoGap(t *testing.T) {
	got := inject(t, "cmd param(name) param(age) param(undefined)", map[string]string{
		"name": "Joe",
		"age":  "20",
	})

	want := "cmd --name 'Joe' --age '20'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
