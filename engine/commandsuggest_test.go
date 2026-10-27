package engine

import (
	"strings"
	"testing"
)

// buildEnv is a small helper shared by these tests: it loads one package made of
// the given profiles into a fresh library and initializes a VirtualEnvironment from
// it, the same way the real CLI does at startup.
func buildEnv(t *testing.T, profiles ...map[string]any) *VirtualEnvironment {
	t.Helper()
	lib := LocalLibrary()
	p := pkg("", "aux4", "test-pkg", "1.0.0", profiles...)
	if err := lib.Load("/x/.aux4", "test-pkg", pkgJSON(t, p)); err != nil {
		t.Fatalf("load: %v", err)
	}
	env, err := InitializeVirtualEnvironment(lib, &VirtualExecutorRegisty{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	return env
}

// Mirrors the dojo trace: the agent typed "ai agent browser find", but "browser"
// only exists one level down from "agent" directly, not under "ai agent".
func TestFindCommandPath_WrongParentLevel(t *testing.T) {
	env := buildEnv(t,
		profile("main",
			cmd("ai", "profile:ai"),
			cmd("agent", "profile:agent")),
		profile("ai",
			cmd("agent", "profile:ai:agent")),
		profile("ai:agent",
			cmd("ask", "echo ask")),
		profile("agent",
			cmd("browser", "profile:agent:browser")),
		profile("agent:browser",
			cmd("find", "echo find"),
			cmd("run", "echo run")),
	)

	got := env.FindCommandPath("browser", []string{"find"})
	want := "agent browser find"
	if strings.Join(got, " ") != want {
		t.Fatalf("FindCommandPath(browser, [find]) = %v, want %q", got, want)
	}
}

func TestFindCommandPath_NothingSimilar(t *testing.T) {
	env := buildEnv(t,
		profile("main",
			cmd("agent", "profile:agent")),
		profile("agent",
			cmd("ask", "echo ask")),
	)

	got := env.FindCommandPath("browser", []string{"find"})
	if got != nil {
		t.Fatalf("FindCommandPath(browser, [find]) = %v, want nil", got)
	}
}

func TestFindCommandPath_IgnoresPrivateCommands(t *testing.T) {
	lib := LocalLibrary()
	p := map[string]any{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": []map[string]any{
			{"name": "main", "commands": []map[string]any{
				{"name": "agent", "execute": []string{"profile:agent"}},
			}},
			{"name": "agent", "commands": []map[string]any{
				{"name": "browser", "execute": []string{"profile:agent:browser"}, "private": true},
			}},
			{"name": "agent:browser", "commands": []map[string]any{
				{"name": "find", "execute": []string{"echo find"}},
			}},
		},
	}
	if err := lib.Load("/x/.aux4", "test-pkg", pkgJSON(t, p)); err != nil {
		t.Fatalf("load: %v", err)
	}
	env, err := InitializeVirtualEnvironment(lib, &VirtualExecutorRegisty{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	got := env.FindCommandPath("browser", []string{"find"})
	if got != nil {
		t.Fatalf("FindCommandPath should not match a private command, got %v", got)
	}
}

func TestVisibleCommandNames_ExcludesPrivate(t *testing.T) {
	lib := LocalLibrary()
	p := map[string]any{
		"scope": "aux4", "name": "test-pkg", "version": "1.0.0",
		"profiles": []map[string]any{
			{"name": "main", "commands": []map[string]any{
				{"name": "visible", "execute": []string{"echo visible"}},
				{"name": "hidden", "execute": []string{"echo hidden"}, "private": true},
			}},
		},
	}
	if err := lib.Load("/x/.aux4", "test-pkg", pkgJSON(t, p)); err != nil {
		t.Fatalf("load: %v", err)
	}
	env, err := InitializeVirtualEnvironment(lib, &VirtualExecutorRegisty{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	names := env.GetProfile("main").VisibleCommandNames()
	if strings.Join(names, ",") != "visible" {
		t.Fatalf("VisibleCommandNames() = %v, want [visible]", names)
	}
}
