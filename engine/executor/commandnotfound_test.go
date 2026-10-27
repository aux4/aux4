package executor

import (
	"strings"
	"testing"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
)

// buildTestEnv loads a single raw .aux4 package definition and initializes a
// VirtualEnvironment from it, exactly like the real CLI does at startup.
func buildTestEnv(t *testing.T, aux4JSON string) *engine.VirtualEnvironment {
	t.Helper()

	lib := engine.LocalLibrary()
	if err := lib.Load("/x/.aux4", "test-pkg", []byte(aux4JSON)); err != nil {
		t.Fatalf("load: %v", err)
	}

	env, err := engine.InitializeVirtualEnvironment(lib, &engine.VirtualExecutorRegisty{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	env.CurrentProfile = "main"
	return env
}

func run(env *engine.VirtualEnvironment, args ...string) error {
	_, actions, params := param.ParseArgs(args)
	env.OriginalActions = actions
	env.OriginalParams = &params
	return MainExecute(env, actions, &params)
}

// Reproduces the dojo trace: an agent called "aux4 ai agent browser find", but the
// real command is "aux4 agent browser find" -- "browser" exists only one level
// down from "agent" directly, not under "ai agent".
func TestMainExecute_CommandNotFound_WrongParentLevel(t *testing.T) {
	env := buildTestEnv(t, `{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": [
			{"name": "main", "commands": [
				{"name": "ai", "execute": ["profile:ai"]},
				{"name": "agent", "execute": ["profile:agent"]}
			]},
			{"name": "ai", "commands": [
				{"name": "agent", "execute": ["profile:ai:agent"]}
			]},
			{"name": "ai:agent", "commands": [
				{"name": "ask", "execute": ["echo ask"]},
				{"name": "chat", "execute": ["echo chat"]}
			]},
			{"name": "agent", "commands": [
				{"name": "browser", "execute": ["profile:agent:browser"]}
			]},
			{"name": "agent:browser", "commands": [
				{"name": "find", "execute": ["echo find"]},
				{"name": "run", "execute": ["echo run"]}
			]}
		]
	}`)

	err := run(env, "ai", "agent", "browser", "find")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	msg := err.Error()
	if !strings.HasPrefix(msg, "Command not found:") {
		t.Fatalf("message must start with %q, got %q", "Command not found:", msg)
	}
	if !strings.Contains(msg, "Command not found: `browser` under `aux4 ai agent`.") {
		t.Errorf("message missing the failure location, got %q", msg)
	}
	if !strings.Contains(msg, "Available commands here: ask, chat") {
		t.Errorf("message missing the sibling command list, got %q", msg)
	}
	if !strings.Contains(msg, "Did you mean: aux4 agent browser find") {
		t.Errorf("message missing the suggestion, got %q", msg)
	}
}

// A sibling command exists at the same level under a different spelling.
func TestMainExecute_CommandNotFound_Misspelling(t *testing.T) {
	env := buildTestEnv(t, `{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": [
			{"name": "main", "commands": [
				{"name": "browser", "execute": ["echo browser"]},
				{"name": "config", "execute": ["echo config"]}
			]}
		]
	}`)

	err := run(env, "borwser")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	msg := err.Error()
	if !strings.HasPrefix(msg, "Command not found:") {
		t.Fatalf("message must start with %q, got %q", "Command not found:", msg)
	}
	if !strings.Contains(msg, "Command not found: `borwser`.") {
		t.Errorf("message missing the failure location, got %q", msg)
	}
	if !strings.Contains(msg, "Did you mean: aux4 browser") {
		t.Errorf("message missing the close-spelling suggestion, got %q", msg)
	}
}

// Nothing resembles the typo anywhere -- no suggestion should be fabricated.
func TestMainExecute_CommandNotFound_NothingSimilar(t *testing.T) {
	env := buildTestEnv(t, `{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": [
			{"name": "main", "commands": [
				{"name": "config", "execute": ["echo config"]},
				{"name": "jobs", "execute": ["echo jobs"]}
			]}
		]
	}`)

	err := run(env, "zzzzzzzz")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	msg := err.Error()
	if !strings.HasPrefix(msg, "Command not found:") {
		t.Fatalf("message must start with %q, got %q", "Command not found:", msg)
	}
	if strings.Contains(msg, "Did you mean:") {
		t.Errorf("message should not fabricate a suggestion, got %q", msg)
	}
	if !strings.Contains(msg, "Available commands here: config, jobs") {
		t.Errorf("message missing the sibling command list, got %q", msg)
	}
}

// The exit code for an unresolved command must stay 127, unchanged by the richer
// message.
func TestMainExecute_CommandNotFound_ExitCodeUnchanged(t *testing.T) {
	env := buildTestEnv(t, `{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": [
			{"name": "main", "commands": [
				{"name": "config", "execute": ["echo config"]}
			]}
		]
	}`)

	err := run(env, "nope")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	aux4Err, ok := err.(core.Aux4Error)
	if !ok {
		t.Fatalf("expected core.Aux4Error, got %T", err)
	}
	if aux4Err.ExitCode != 127 {
		t.Errorf("ExitCode = %d, want 127", aux4Err.ExitCode)
	}
}
