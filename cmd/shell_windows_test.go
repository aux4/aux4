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

func TestIsWslLauncher(t *testing.T) {
	t.Setenv("SystemRoot", `C:\Windows`)
	if !isWslLauncher(`C:\Windows\System32\bash.exe`) {
		t.Fatal("System32 bash is the WSL launcher")
	}
	if isWslLauncher(`C:\Program Files\Git\bin\bash.exe`) {
		t.Fatal("Git bash is not the WSL launcher")
	}
}
