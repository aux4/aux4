package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pkgJSON(t *testing.T, v map[string]any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func cmd(name string, execute ...string) map[string]any {
	return map[string]any{"name": name, "execute": execute}
}

func profile(name string, commands ...map[string]any) map[string]any {
	return map[string]any{"name": name, "commands": commands}
}

func pkg(repository, scope, name, version string, profiles ...map[string]any) map[string]any {
	p := map[string]any{"scope": scope, "name": name, "version": version, "profiles": profiles}
	if repository != "" {
		p["repository"] = repository
	}
	return p
}

func TestPackageKey(t *testing.T) {
	cases := []struct {
		repository, scope, name, want string
	}{
		{"", "aux4", "browser", "public:aux4/browser"},
		{"public", "aux4", "browser", "public:aux4/browser"},
		{"system", "aux4", "agent-vm", "system:aux4/agent-vm"},
		{"", "agent", "browser", "public:agent/browser"},
		{"", "", "/home/u/.aux4.config/global.aux4", "/home/u/.aux4.config/global.aux4"},
		{"system", "", "builtin", "builtin"},
	}
	for _, c := range cases {
		lib := LocalLibrary()
		data := pkgJSON(t, pkg(c.repository, c.scope, c.name, "1.0.0"))
		if err := lib.Load("/x/.aux4", c.name, data); err != nil {
			t.Fatalf("load %+v: %v", c, err)
		}
		if len(lib.orderedPackages) != 1 || lib.orderedPackages[0] != c.want {
			t.Errorf("key for %+v = %v, want %s", c, lib.orderedPackages, c.want)
		}
	}
}

func TestLibraryCollisions(t *testing.T) {
	load := func(lib *Library, p map[string]any) error {
		return lib.Load("/x/.aux4", "", pkgJSON(t, p))
	}

	t.Run("same name different scope loads both", func(t *testing.T) {
		lib := LocalLibrary()
		if err := load(lib, pkg("", "aux4", "browser", "1.0.0")); err != nil {
			t.Fatal(err)
		}
		if err := load(lib, pkg("", "agent", "browser", "1.0.0")); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("same package different repository loads both", func(t *testing.T) {
		lib := LocalLibrary()
		if err := load(lib, pkg("public", "aux4", "tool", "1.0.0")); err != nil {
			t.Fatal(err)
		}
		if err := load(lib, pkg("system", "aux4", "tool", "1.0.0")); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("same package same repository collides even with another version", func(t *testing.T) {
		lib := LocalLibrary()
		if err := load(lib, pkg("", "aux4", "tool", "1.0.0")); err != nil {
			t.Fatal(err)
		}
		err := load(lib, pkg("public", "aux4", "tool", "2.0.0"))
		if err == nil || !strings.Contains(err.Error(), "Package public:aux4/tool already exists") {
			t.Fatalf("expected collision, got %v", err)
		}
	})
}

// The merge is first-loaded-wins per command across packages, so the result must
// depend only on LOAD order, never on the Library key. The packages below are
// loaded in an order that is neither alphabetical by key nor by name, and two of
// them contribute the same agent-manager routing command.
func TestMergeFollowsLoadOrderNotKey(t *testing.T) {
	lib := LocalLibrary()
	packages := []map[string]any{
		pkg("", "zeta", "agent-host", "1.0.0",
			profile("main", cmd("agent-manager", "profile:agent-manager")),
			profile("agent-manager", cmd("agents", "echo host-agents"))),
		pkg("system", "aux4", "kb-agent", "1.0.0",
			profile("agent-manager", cmd("kb", "profile:agent-manager:kb")),
			profile("agent-manager:kb", cmd("run", "echo kb-run"))),
		pkg("", "alpha", "kb-agent", "1.0.0",
			profile("agent-manager", cmd("kb", "profile:agent-manager:other")),
			profile("agent-manager:kb", cmd("run", "echo shadowed"))),
		pkg("", "aux4", "browser", "1.0.0", profile("main", cmd("browser", "echo aux4-browser"))),
		pkg("", "agent", "browser", "1.0.0", profile("main", cmd("browser", "echo agent-browser"))),
	}
	for _, p := range packages {
		if err := lib.Load("/x/.aux4", "", pkgJSON(t, p)); err != nil {
			t.Fatal(err)
		}
	}

	env, err := InitializeVirtualEnvironment(lib, &VirtualExecutorRegisty{})
	if err != nil {
		t.Fatal(err)
	}

	check := func(profileName, command, want string) {
		t.Helper()
		p := env.GetProfile(profileName)
		if p == nil {
			t.Fatalf("profile %s missing", profileName)
		}
		c, ok := p.Commands[command]
		if !ok {
			t.Fatalf("%s/%s missing", profileName, command)
		}
		if c.Execute[0] != want {
			t.Errorf("%s/%s = %q, want %q", profileName, command, c.Execute[0], want)
		}
	}

	check("main", "agent-manager", "profile:agent-manager")
	check("agent-manager", "kb", "profile:agent-manager:kb")
	check("agent-manager:kb", "run", "echo kb-run")
	check("main", "browser", "echo aux4-browser")

	if got := strings.Join(env.GetProfile("agent-manager").CommandsOrdered, ","); got != "agents,kb" {
		t.Errorf("agent-manager command order = %s", got)
	}
}

// Save must write profiles in first-seen order, identically on every run.
func TestSaveIsDeterministic(t *testing.T) {
	var first string
	for i := 0; i < 20; i++ {
		lib := LocalLibrary()
		p := pkg("", "aux4", "many", "1.0.0",
			profile("main", cmd("a", "profile:p1")),
			profile("p1", cmd("x", "echo")), profile("p2", cmd("x", "echo")),
			profile("p3", cmd("x", "echo")), profile("p4", cmd("x", "echo")),
			profile("p5", cmd("x", "echo")), profile("p6", cmd("x", "echo")))
		if err := lib.Load("/x/.aux4", "", pkgJSON(t, p)); err != nil {
			t.Fatal(err)
		}
		env, err := InitializeVirtualEnvironment(lib, &VirtualExecutorRegisty{})
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(t.TempDir(), "global.aux4")
		if err := env.Save(out); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(out)
		var saved struct {
			Profiles []struct{ Name string } `json:"profiles"`
		}
		json.Unmarshal(data, &saved)
		names := []string{}
		for _, sp := range saved.Profiles {
			names = append(names, sp.Name)
		}
		got := strings.Join(names, ",")
		if got != "main,p1,p2,p3,p4,p5,p6" {
			t.Fatalf("profile order = %s", got)
		}
		if first == "" {
			first = string(data)
		} else if first != string(data) {
			t.Fatal("Save output differs between runs")
		}
	}
}
