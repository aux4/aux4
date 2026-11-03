//go:build !windows

package cmd

import "testing"

func TestResolveShellUnix(t *testing.T) {
	if got, err := resolveShell(""); err != nil || got != "/bin/sh" {
		t.Fatalf("resolveShell(\"\") = %q, %v; want /bin/sh", got, err)
	}
	if got, err := resolveShell("/bin/zsh"); err != nil || got != "/bin/zsh" {
		t.Fatalf("resolveShell(/bin/zsh) = %q, %v; want /bin/zsh", got, err)
	}
	if got, err := resolveBash(); err != nil || got != "bash" {
		t.Fatalf("resolveBash() = %q, %v; want bash", got, err)
	}
}
