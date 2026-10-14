package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aux4.dev/aux4/core"
)

func TestSchemeRegistry_RegisterAndGet(t *testing.T) {
	registry := NewSchemeRegistry()

	registry.Register("cloud", core.SchemeProvider{Get: "cloud files get", Put: "cloud files upload"})

	provider, ok := registry.Get("cloud")
	if !ok {
		t.Fatal("expected cloud scheme to be registered")
	}
	if provider.Get != "cloud files get" || provider.Put != "cloud files upload" {
		t.Errorf("unexpected provider: %+v", provider)
	}

	if _, ok := registry.Get("missing"); ok {
		t.Error("expected unregistered scheme to be absent")
	}
}

func TestSchemeRegistry_FirstWins(t *testing.T) {
	registry := NewSchemeRegistry()

	registry.Register("cloud", core.SchemeProvider{Get: "first get", Put: "first put"})
	registry.Register("cloud", core.SchemeProvider{Get: "second get", Put: "second put"})

	provider, ok := registry.Get("cloud")
	if !ok {
		t.Fatal("expected cloud scheme to be registered")
	}
	if provider.Get != "first get" {
		t.Errorf("expected first registration to win, got %q", provider.Get)
	}
}

func TestSchemeRegistry_NilSafe(t *testing.T) {
	var registry *SchemeRegistry
	if _, ok := registry.Get("cloud"); ok {
		t.Error("expected nil registry to report no providers")
	}
}

func TestInitializeVirtualEnvironment_CollectsSchemes(t *testing.T) {
	library := LocalLibrary()

	data := pkgJSON(t, pkg("", "fake", "cloud-provider", "0.1.0",
		profile("main"),
	))

	var pack core.Package
	if err := json.Unmarshal(data, &pack); err != nil {
		t.Fatal(err)
	}
	pack.Schemes = map[string]core.SchemeProvider{
		"cloud": {Get: "cloud files get", Put: "cloud files upload"},
	}

	if err := library.load(pack); err != nil {
		t.Fatal(err)
	}

	registry := CreateVirtualExecutorRegistry()
	env, err := InitializeVirtualEnvironment(library, registry)
	if err != nil {
		t.Fatal(err)
	}

	provider, ok := env.Schemes.Get("cloud")
	if !ok {
		t.Fatal("expected cloud scheme to be collected from the installed package")
	}
	if provider.Get != "cloud files get" {
		t.Errorf("unexpected provider.Get: %q", provider.Get)
	}
}

// A scheme collected from an installed package's own .aux4 must survive a
// round trip through VirtualEnvironment.Save (the file pkger writes as
// global.aux4 on every install). Without this, a <scheme>:// provider only
// resolves for the process that loaded the declaring package's .aux4
// directly — never for a command run against the aggregated global
// environment pkger produces, which is the normal runtime path.
func TestVirtualEnvironmentSave_PersistsSchemes(t *testing.T) {
	library := LocalLibrary()

	data := pkgJSON(t, pkg("", "aux4", "cloud-provider", "0.1.0",
		profile("main"),
	))

	var pack core.Package
	if err := json.Unmarshal(data, &pack); err != nil {
		t.Fatal(err)
	}
	pack.Schemes = map[string]core.SchemeProvider{
		"cloud": {Get: "aux4 cloud files get", Put: "aux4 cloud files upload"},
	}

	if err := library.load(pack); err != nil {
		t.Fatal(err)
	}

	registry := CreateVirtualExecutorRegistry()
	env, err := InitializeVirtualEnvironment(library, registry)
	if err != nil {
		t.Fatal(err)
	}

	savePath := filepath.Join(t.TempDir(), "global.aux4")
	if err := env.Save(savePath); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatal(err)
	}

	var reloaded core.Package
	if err := json.Unmarshal(saved, &reloaded); err != nil {
		t.Fatal(err)
	}

	provider, ok := reloaded.Schemes["cloud"]
	if !ok {
		t.Fatalf("saved global.aux4 is missing the cloud scheme: %s", saved)
	}
	if provider.Get != "aux4 cloud files get" || provider.Put != "aux4 cloud files upload" {
		t.Errorf("unexpected provider after save/reload: %+v", provider)
	}

	// Reloading the saved file into a fresh library/environment must also
	// resolve the scheme — this is the exact path a second `aux4` invocation
	// takes against the on-disk global.aux4.
	reloadLibrary := LocalLibrary()
	if err := reloadLibrary.LoadFile(savePath); err != nil {
		t.Fatal(err)
	}
	reloadRegistry := CreateVirtualExecutorRegistry()
	reloadEnv, err := InitializeVirtualEnvironment(reloadLibrary, reloadRegistry)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reloadEnv.Schemes.Get("cloud"); !ok {
		t.Fatal("expected cloud scheme to resolve after reloading the saved global.aux4")
	}
}
