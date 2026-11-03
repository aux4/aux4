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

// TestShellResolvesPerRequestNotJustOncePerProcess reproduces the daemon scenario: the daemon
// serves every client request in the same process and sets $SHELL fresh before each one (see
// daemon.Server.handleConnection + main.go's executeFn), so Shell() must reflect whichever
// request is running, not whichever one happened to run first. Before the fix, the shell was
// cached with sync.Once for the lifetime of the process, so the first request's $SHELL stuck for
// every later one.
func TestShellResolvesPerRequestNotJustOncePerProcess(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	first, err := ResolveShell()
	if err != nil || first != "/bin/bash" {
		t.Fatalf("request 1: ResolveShell() = %q, %v; want /bin/bash", first, err)
	}
	// Plain Shell() within the same "request" returns the value just resolved.
	if got, err := Shell(); err != nil || got != "/bin/bash" {
		t.Fatalf("request 1: Shell() = %q, %v; want /bin/bash", got, err)
	}

	// A second daemon request comes in with a different client environment.
	t.Setenv("SHELL", "/bin/zsh")
	second, err := ResolveShell()
	if err != nil || second != "/bin/zsh" {
		t.Fatalf("request 2: ResolveShell() = %q, %v; want /bin/zsh (got the request 1 value cached instead)", second, err)
	}
	if got, err := Shell(); err != nil || got != "/bin/zsh" {
		t.Fatalf("request 2: Shell() = %q, %v; want /bin/zsh", got, err)
	}
}
