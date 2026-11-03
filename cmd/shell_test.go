package cmd

import "testing"

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

// TestBashResolvesPerRequest is the Bash() equivalent of the above. On Windows resolveBash reads
// $SHELL too, so it is subject to the same daemon staleness; on macOS/Linux resolveBash always
// returns "bash" regardless of $SHELL, so this only asserts ResolveBash() keeps working per call.
func TestBashResolvesPerRequest(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if _, err := ResolveBash(); err != nil {
		t.Fatalf("request 1: ResolveBash() error = %v", err)
	}
	t.Setenv("SHELL", "/bin/zsh")
	if _, err := ResolveBash(); err != nil {
		t.Fatalf("request 2: ResolveBash() error = %v", err)
	}
	if _, err := Bash(); err != nil {
		t.Fatalf("Bash() error = %v", err)
	}
}
