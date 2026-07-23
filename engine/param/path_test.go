package param

import (
	"path/filepath"
	"testing"

	"aux4.dev/aux4/core"
)

func injectPath(t *testing.T, instruction string, vars map[string]string) string {
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

// abs mirrors path()'s own resolution so the expectation is CWD-independent.
func abs(t *testing.T, parts ...string) string {
	t.Helper()
	a, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return a
}

func TestPathResolvesRelativeToAbsolute(t *testing.T) {
	got := injectPath(t, "path(db)", map[string]string{"db": "local.db"})
	if want := abs(t, "local.db"); got != want {
		t.Errorf("path(db) = %q, want %q", got, want)
	}
}

func TestPathKeepsAbsoluteUnchanged(t *testing.T) {
	got := injectPath(t, "path(db)", map[string]string{"db": "/var/data/x.db"})
	if got != "/var/data/x.db" {
		t.Errorf("path(db) = %q, want /var/data/x.db", got)
	}
}

func TestPathJoinsSegmentsWithQuotedLiteral(t *testing.T) {
	got := injectPath(t, "path(dir/'backups'/name)", map[string]string{"dir": "data", "name": "shop"})
	if want := abs(t, "data", "backups", "shop"); got != want {
		t.Errorf("path(dir/'backups'/name) = %q, want %q", got, want)
	}
}

func TestPathCollapsesParentSegment(t *testing.T) {
	got := injectPath(t, "path(dir/../'archive'/name)", map[string]string{"dir": "data", "name": "shop"})
	if want := abs(t, "archive", "shop"); got != want {
		t.Errorf("path(dir/../'archive'/name) = %q, want %q", got, want)
	}
}

func TestPathDotIsLiteralSegment(t *testing.T) {
	got := injectPath(t, "path('.'/name)", map[string]string{"name": "shop"})
	if want := abs(t, "shop"); got != want {
		t.Errorf("path('.'/name) = %q, want %q", got, want)
	}
}

func TestPathMissingVariableStaysEmpty(t *testing.T) {
	got := injectPath(t, "path(missing)", map[string]string{})
	if got != "" {
		t.Errorf("path(missing) = %q, want empty", got)
	}
}

func TestPathEscapedIsLiteral(t *testing.T) {
	got := injectPath(t, "\\path(x)", map[string]string{"x": "y"})
	if got != "path(x)" {
		t.Errorf("\\path(x) = %q, want path(x)", got)
	}
}
