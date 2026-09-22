package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine/param"
)

func newParams() *param.Parameters {
	_, _, p := param.ParseArgs([]string{})
	return &p
}

func TestSetCommandExecutor_ShellValueWithSemicolons(t *testing.T) {
	// Reproduces the panic: set:name=!cmd1; cmd2 was split on ";"
	// producing a fragment with no "=", causing index out of range.
	executor := &SetCommandExecutor{
		Command: "set:result=!echo hello; echo world",
	}

	params := newParams()

	err := executor.Execute(nil, core.Command{}, []string{}, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	value, _ := params.Expr(core.Command{}, []string{}, "${result}")
	// The shell command "echo hello; echo world" should produce "hello\nworld"
	expected := "hello\nworld"
	if value != expected {
		t.Errorf("got %q, want %q", value, expected)
	}
}

func TestSetCommandExecutor_MultiAssignmentStillWorks(t *testing.T) {
	// Multi-assignment set:a=1;b=2 must continue to work.
	executor := &SetCommandExecutor{
		Command: "set:a=hello;b=world",
	}

	params := newParams()

	err := executor.Execute(nil, core.Command{}, []string{}, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a, _ := params.Expr(core.Command{}, []string{}, "${a}")
	b, _ := params.Expr(core.Command{}, []string{}, "${b}")

	if a != "hello" {
		t.Errorf("a: got %q, want %q", a, "hello")
	}
	if b != "world" {
		t.Errorf("b: got %q, want %q", b, "world")
	}
}

// TestEachCommandExecutor_NoResponseSet reproduces CORE-037: each: always
// iterates the "response" variable. When response was never set,
// reflect.TypeOf(nil) returns a nil reflect.Type, and calling .Kind() on it
// dereferenced a nil pointer. It must instead return a clear error.
func TestEachCommandExecutor_NoResponseSet(t *testing.T) {
	executor := &EachCommandExecutor{
		Command: "each:echo ${item}",
	}

	params := newParams()

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("each: panicked instead of returning an error: %v", r)
			}
		}()
		err = executor.Execute(nil, core.Command{}, []string{}, params)
	}()

	if err == nil {
		t.Fatal("expected an error when response is unset, got nil")
	}

	aux4Err, ok := err.(core.Aux4Error)
	if !ok {
		t.Fatalf("expected core.Aux4Error, got %T: %v", err, err)
	}

	if !strings.Contains(aux4Err.Message, "response") {
		t.Errorf("expected error message to mention 'response', got %q", aux4Err.Message)
	}
}

// TestEachCommandExecutor_ResponseIsNonIterableScalar covers response set to
// a scalar (e.g. a number) that is neither a slice/array nor a string.
func TestEachCommandExecutor_ResponseIsNonIterableScalar(t *testing.T) {
	executor := &EachCommandExecutor{
		Command: "each:echo ${item}",
	}

	params := newParams()
	params.Update("response", 42)

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("each: panicked instead of returning an error: %v", r)
			}
		}()
		err = executor.Execute(nil, core.Command{}, []string{}, params)
	}()

	if err == nil {
		t.Fatal("expected an error when response is a non-iterable scalar, got nil")
	}

	if _, ok := err.(core.Aux4Error); !ok {
		t.Fatalf("expected core.Aux4Error, got %T: %v", err, err)
	}
}

// TestEachCommandExecutor_MultipleVariablePassedZeroTimes reproduces a more
// realistic trigger for the same nil-response class: a `multiple: true`
// variable that the user simply did not pass. Passing zero values for an
// optional repeatable flag is normal, correct usage — not a mistake — and
// must not crash. value(*) silently omits a `multiple: true` key that was
// never passed (rather than yielding an empty list), so a command that
// spreads value(*) and never explicitly populates ${response} lands on the
// exact same nil response as CORE-037's original repro.
func TestEachCommandExecutor_MultipleVariablePassedZeroTimes(t *testing.T) {
	defaultValue := ""
	command := core.Command{
		Name: "body",
		Help: &core.CommandHelp{
			Variables: []*core.CommandHelpVariable{
				{Name: "options", Multiple: true, Default: &defaultValue},
			},
		},
	}

	// --options was never passed on the command line.
	_, _, params := param.ParseArgs([]string{})

	setExecutor := &SetCommandExecutor{Command: "set:allParams=value(*)"}
	if err := setExecutor.Execute(nil, command, []string{}, &params); err != nil {
		t.Fatalf("unexpected error building allParams: %v", err)
	}

	eachExecutor := &EachCommandExecutor{Command: "each:echo ${item}"}

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("each: panicked on an unset multiple:true variable passed zero times: %v", r)
			}
		}()
		err = eachExecutor.Execute(nil, command, []string{}, &params)
	}()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if _, ok := err.(core.Aux4Error); !ok {
		t.Fatalf("expected core.Aux4Error, got %T: %v", err, err)
	}
}

// TestEachCommandExecutor_NormalListStillWorks ensures the fix does not
// regress the working case: response set to a []any list iterates normally.
func TestEachCommandExecutor_NormalListStillWorks(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	executor := &EachCommandExecutor{
		Command: "each:echo ${item} >> " + outFile,
	}

	params := newParams()
	params.Update("response", []any{"a", "b", "c"})

	err := executor.Execute(nil, core.Command{}, []string{}, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("failed to read output file: %v", readErr)
	}

	expected := "a\nb\nc\n"
	if string(content) != expected {
		t.Errorf("got %q, want %q", string(content), expected)
	}
}

// TestEachCommandExecutor_NormalStringResponseStillWorks covers the string
// branch: response set to a newline-delimited string iterates per line.
func TestEachCommandExecutor_NormalStringResponseStillWorks(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	executor := &EachCommandExecutor{
		Command: "each:echo ${item} >> " + outFile,
	}

	params := newParams()
	params.Update("response", "x\ny")

	err := executor.Execute(nil, core.Command{}, []string{}, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("failed to read output file: %v", readErr)
	}

	expected := "x\ny\n"
	if string(content) != expected {
		t.Errorf("got %q, want %q", string(content), expected)
	}
}
