//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeExe(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("MZ"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// isolate points every place findBash looks at an empty temp directory.
func isolate(t *testing.T) string {
	dir := t.TempDir()
	for _, env := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"} {
		t.Setenv(env, filepath.Join(dir, "none"))
	}
	t.Setenv("PATH", filepath.Join(dir, "empty"))
	return dir
}

func TestResolveShellFindsGitBashNextToGit(t *testing.T) {
	dir := isolate(t)
	fakeExe(t, filepath.Join(dir, "Git", "cmd", "git.exe"))
	bash := filepath.Join(dir, "Git", "bin", "bash.exe")
	fakeExe(t, bash)
	t.Setenv("PATH", filepath.Join(dir, "Git", "cmd"))

	for _, shell := range []string{"", "/usr/bin/bash", "/bin/sh"} {
		got, err := resolveShell(shell)
		if err != nil || !strings.EqualFold(got, bash) {
			t.Fatalf("resolveShell(%q) = %q, %v; want %q", shell, got, err, bash)
		}
	}
}

func TestResolveShellDefaultInstallLocation(t *testing.T) {
	dir := isolate(t)
	bash := filepath.Join(dir, "Program Files", "Git", "bin", "bash.exe")
	fakeExe(t, bash)
	t.Setenv("ProgramFiles", filepath.Join(dir, "Program Files"))

	got, err := resolveShell("")
	if err != nil || !strings.EqualFold(got, bash) {
		t.Fatalf("resolveShell(\"\") = %q, %v; want %q", got, err, bash)
	}
}

func TestResolveShellUsesShellFullPath(t *testing.T) {
	dir := isolate(t)
	shell := filepath.Join(dir, "msys64", "usr", "bin", "bash.exe")
	fakeExe(t, shell)

	got, err := resolveShell(shell)
	if err != nil || got != shell {
		t.Fatalf("resolveShell(%q) = %q, %v", shell, got, err)
	}
}

func TestResolveShellNoBash(t *testing.T) {
	isolate(t)
	_, err := resolveShell("")
	if err == nil || !strings.Contains(err.Error(), "Git for Windows") {
		t.Fatalf("resolveShell(\"\") error = %v; want a message pointing at Git for Windows", err)
	}
	_, err = resolveShell("zsh")
	if err == nil || !strings.Contains(err.Error(), "SHELL is set to 'zsh'") {
		t.Fatalf("resolveShell(zsh) error = %v", err)
	}
}

// TestShellResolvesPerRequestNotJustOncePerProcess reproduces the daemon scenario: the daemon
// serves every client request in the same process and sets $SHELL fresh before each one (see
// daemon.Server.handleConnection + main.go's executeFn), so Shell() must reflect whichever
// request is running, not whichever one happened to run first. Before the fix, the shell was
// cached with sync.Once for the lifetime of the process, so the first request's $SHELL stuck for
// every later one.
func TestShellResolvesPerRequestNotJustOncePerProcess(t *testing.T) {
	dir := isolate(t)
	bashShell := filepath.Join(dir, "msys64", "usr", "bin", "bash.exe")
	zshShell := filepath.Join(dir, "msys64", "usr", "bin", "zsh.exe")
	fakeExe(t, bashShell)
	fakeExe(t, zshShell)

	t.Setenv("SHELL", bashShell)
	first, err := ResolveShell()
	if err != nil || first != bashShell {
		t.Fatalf("request 1: ResolveShell() = %q, %v; want %q", first, err, bashShell)
	}
	// Plain Shell() within the same "request" returns the value just resolved.
	if got, err := Shell(); err != nil || got != bashShell {
		t.Fatalf("request 1: Shell() = %q, %v; want %q", got, err, bashShell)
	}

	// A second daemon request comes in with a different client environment. zsh.exe isn't named
	// bash/sh, but it's given as a full path, so resolveShell accepts it as given (same as a real
	// bash.exe at a non-default location would be).
	t.Setenv("SHELL", zshShell)
	second, err := ResolveShell()
	if err != nil || second != zshShell {
		t.Fatalf("request 2: ResolveShell() = %q, %v; want %q (got the request 1 value cached instead)", second, err, zshShell)
	}
	if got, err := Shell(); err != nil || got != zshShell {
		t.Fatalf("request 2: Shell() = %q, %v; want %q", got, err, zshShell)
	}
}

// TestBashResolvesPerRequest is the Bash() equivalent of the above. resolveBash on Windows reads
// $SHELL directly (unlike macOS/Linux, where it always returns "bash"), so it is subject to the
// exact same daemon staleness bug.
func TestBashResolvesPerRequest(t *testing.T) {
	dir := isolate(t)
	bash1 := filepath.Join(dir, "first", "bash.exe")
	bash2 := filepath.Join(dir, "second", "bash.exe")
	fakeExe(t, bash1)
	fakeExe(t, bash2)

	t.Setenv("SHELL", bash1)
	if got, err := ResolveBash(); err != nil || got != bash1 {
		t.Fatalf("request 1: ResolveBash() = %q, %v; want %q", got, err, bash1)
	}

	t.Setenv("SHELL", bash2)
	if got, err := ResolveBash(); err != nil || got != bash2 {
		t.Fatalf("request 2: ResolveBash() = %q, %v; want %q (got the request 1 value cached instead)", got, err, bash2)
	}
	if got, err := Bash(); err != nil || got != bash2 {
		t.Fatalf("request 2: Bash() = %q, %v; want %q", got, err, bash2)
	}
}

func TestIsWslLauncher(t *testing.T) {
	t.Setenv("SystemRoot", `C:\Windows`)
	if !isWslLauncher(`C:\Windows\System32\bash.exe`) {
		t.Fatal("System32 bash is the WSL launcher")
	}
	if isWslLauncher(`C:\Program Files\Git\bin\bash.exe`) {
		t.Fatal("Git bash is not the WSL launcher")
	}
}
