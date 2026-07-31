package executor

import (
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
