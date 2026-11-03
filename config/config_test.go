package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestAux4HomeDirectoryUsesHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if got, want := GetAux4HomeDirectory(), filepath.Join(home, ".aux4.config"); got != want {
		t.Fatalf("GetAux4HomeDirectory() = %q, want %q", got, want)
	}
}

func TestAux4HomeDirectoryWithoutHome(t *testing.T) {
	t.Setenv("HOME", "")
	if runtime.GOOS == "windows" {
		profile := t.TempDir()
		t.Setenv("USERPROFILE", profile)
		if got, want := GetAux4HomeDirectory(), filepath.Join(profile, ".aux4.config"); got != want {
			t.Fatalf("GetAux4HomeDirectory() = %q, want %q", got, want)
		}
		return
	}
	if got := GetAux4HomeDirectory(); got != ".aux4.config" {
		t.Fatalf("GetAux4HomeDirectory() = %q, want .aux4.config (unchanged on %s)", got, runtime.GOOS)
	}
}
