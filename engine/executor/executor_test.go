package executor

import (
	"testing"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
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

type callerCaptureExecutor struct {
	callerPackage any
}

func (capture *callerCaptureExecutor) Execute(_ *engine.VirtualEnvironment, _ core.Command, _ []string, params *param.Parameters) error {
	capture.callerPackage = params.JustGet("__callerPackage")
	return nil
}

func TestAux4CommandExecutorInjectsImmediateCallerPackage(t *testing.T) {
	library := engine.LocalLibrary()
	if err := library.Load("/target/.aux4", "target", []byte(`{
		"scope":"aux4",
		"name":"target",
		"version":"1.0.0",
		"profiles":[{"name":"main","commands":[{"name":"capture","help":{"text":"capture"}}]}]
	}`)); err != nil {
		t.Fatal(err)
	}

	registry := engine.CreateVirtualExecutorRegistry()
	capture := &callerCaptureExecutor{}
	registry.RegisterExecutor("main.capture", capture)
	env, err := engine.InitializeVirtualEnvironment(library, registry)
	if err != nil {
		t.Fatal(err)
	}

	caller := core.Command{Ref: core.CommandRef{Package: "community/aquarium@2.4.1"}}
	executor := &Aux4CommandExecutor{Command: "aux4 capture --__callerPackage spoofed/package"}
	if err := executor.Execute(env, caller, nil, newParams()); err != nil {
		t.Fatal(err)
	}

	if capture.callerPackage != "community/aquarium" {
		t.Fatalf("__callerPackage = %v, want community/aquarium", capture.callerPackage)
	}
}

func TestPackageIdentity(t *testing.T) {
	tests := map[string]string{
		"community/aquarium@2.4.1": "community/aquarium",
		"community/aquarium":       "community/aquarium",
		".aux4":                    ".aux4",
	}

	for input, want := range tests {
		if got := packageIdentity(input); got != want {
			t.Fatalf("packageIdentity(%q) = %q, want %q", input, got, want)
		}
	}
}
