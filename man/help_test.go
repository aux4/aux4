package man

import (
	"strings"
	"testing"
)

// TestBreakLines_ExactlyAtBoundary reproduces CORE-036: when a word-wrapped
// segment's remaining length is exactly equal to maxLength (maxLineLength -
// len(spacing)), breakLines panicked with "slice bounds out of range" because
// the guard used strict "<" instead of "<=", leaving end = maxLength+1 on a
// string of length maxLength.
//
// spacing = "  " (len 2), maxLineLength = 20 => maxLength = 18.
// The text is split by an explicit "\n" so the second segment ("remaining")
// is exactly 18 characters long with no spaces (so word-break can't shift it).
func TestBreakLines_ExactlyAtBoundary(t *testing.T) {
	spacing := "  "
	maxLineLength := 20
	maxLength := maxLineLength - len(spacing) // 18

	second := strings.Repeat("a", maxLength)
	text := "x\n" + second

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("breakLines panicked on exact-boundary input: %v", r)
		}
	}()

	result := breakLines(text, maxLineLength, spacing)

	expected := spacing + "x\n" + spacing + second
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

// TestBreakLines_OneUnderBoundary is the same shape one character shorter
// than the boundary — must keep working exactly as before.
func TestBreakLines_OneUnderBoundary(t *testing.T) {
	spacing := "  "
	maxLineLength := 20
	maxLength := maxLineLength - len(spacing) // 18

	second := strings.Repeat("a", maxLength-1) // 17 chars
	text := "x\n" + second

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("breakLines panicked on one-under-boundary input: %v", r)
		}
	}()

	result := breakLines(text, maxLineLength, spacing)

	expected := spacing + "x\n" + spacing + second
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

// TestBreakLines_OneOverBoundary is one character longer than the boundary —
// must still hard-wrap without panicking, since there is no space to break on.
func TestBreakLines_OneOverBoundary(t *testing.T) {
	spacing := "  "
	maxLineLength := 20
	maxLength := maxLineLength - len(spacing) // 18

	second := strings.Repeat("a", maxLength+1) // 19 chars, no spaces

	text := "x\n" + second

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("breakLines panicked on one-over-boundary input: %v", r)
		}
	}()

	result := breakLines(text, maxLineLength, spacing)

	if strings.Contains(result, "\x00") {
		t.Errorf("unexpected null byte in result: %q", result)
	}

	// Every wrapped line (after removing the spacing prefix) must fit within maxLength.
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimPrefix(line, spacing)
		if len(trimmed) > maxLength {
			t.Errorf("line exceeds maxLength %d: %q (len %d)", maxLength, trimmed, len(trimmed))
		}
	}

	rejoined := strings.ReplaceAll(strings.ReplaceAll(result, "\n", ""), spacing, "")
	if rejoined != "x"+second {
		t.Errorf("wrapped content lost data: got %q, want %q", rejoined, "x"+second)
	}
}

// TestBreakLines_EmptyString ensures the empty/short-string path (handled by
// the early-return shortcut, before the loop is ever entered) still works.
func TestBreakLines_EmptyString(t *testing.T) {
	spacing := "  "
	maxLineLength := 100

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("breakLines panicked on empty string: %v", r)
		}
	}()

	result := breakLines("", maxLineLength, spacing)

	if result != spacing {
		t.Errorf("got %q, want %q", result, spacing)
	}
}

// TestBreakLines_ShortString covers a short string well under maxLineLength,
// which should be returned as a single spaced line without entering the loop.
func TestBreakLines_ShortString(t *testing.T) {
	spacing := "  "
	maxLineLength := 100

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("breakLines panicked on short string: %v", r)
		}
	}()

	result := breakLines("hello world", maxLineLength, spacing)

	expected := spacing + "hello world"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}
