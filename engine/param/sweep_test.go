package param

import (
	"testing"

	"aux4.dev/aux4/core"
)

// This file covers the CORE-036/CORE-037 sweep: other spots in engine/param
// with the same "unguarded nil/length assumption" shape — a slice expression
// whose bounds are computed with arithmetic that can exceed the string's
// length, or reflect.TypeOf(x).Kind() with no nil check on x. Wrong usage
// must produce a clear error, never a panic.

// TestExpr_MalformedBracketReference reproduces a panic in Parameters.Expr:
// "${items[3}" (missing closing bracket) computed
// originalName[open+1:close] with close == -1 (strings.Index returning "not
// found"), producing an invalid slice range "[:-1]".
func TestExpr_MalformedBracketReference(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	params.Update("items", []any{"a", "b", "c"})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Expr panicked on a malformed bracket reference: %v", r)
		}
	}()

	// No assertion on the resolved value — the point is that a malformed
	// reference must not crash the process. Falling back to the unindexed
	// base name (or erroring) are both acceptable; panicking is not.
	if _, err := params.Expr(core.Command{}, []string{}, "items[3"); err != nil {
		t.Logf("Expr returned an error (acceptable): %v", err)
	}
}

// TestNvl_BareSingleQuoteCandidate reproduces a panic in nvl(): a candidate
// that is exactly one quote character (e.g. from "nvl(',fallback)") matches
// both HasPrefix and HasSuffix against itself, so the "strip the quotes"
// branch computed candidate[1:len(candidate)-1] == candidate[1:0], an
// invalid (low > high) slice range.
func TestNvl_BareSingleQuoteCandidate(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nvl() panicked on a bare single-quote candidate: %v", r)
		}
	}()

	if _, err := InjectParameters(core.Command{}, "nvl(',default)", []string{}, params); err != nil {
		t.Logf("nvl() returned an error (acceptable): %v", err)
	}
}

// TestPath_BareSingleQuoteSegment is the same shape as
// TestNvl_BareSingleQuoteCandidate, in path()'s segment parsing.
func TestPath_BareSingleQuoteSegment(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	params.Update("dir", "data")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("path() panicked on a bare single-quote segment: %v", r)
		}
	}()

	if _, err := InjectParameters(core.Command{}, "path(dir/')", []string{}, params); err != nil {
		t.Logf("path() returned an error (acceptable): %v", err)
	}
}

// TestParametersSet_NilValue is a defensive regression for Parameters.Set:
// reflect.TypeOf(nil) returns a nil reflect.Type, and .Kind() on it
// dereferences a nil pointer — the same shape as CORE-037's each: bug. No
// current caller passes nil today (both callers guard with "if result !=
// nil" first), but Set is exported API within the package and should not
// crash on nil regardless.
func TestParametersSet_NilValue(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Parameters.Set panicked on a nil value: %v", r)
		}
	}()

	params.Set("x", nil)

	got := params.JustGet("x")
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
