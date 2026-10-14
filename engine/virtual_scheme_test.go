package engine

import (
	"encoding/json"
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
