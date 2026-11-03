//go:build windows

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aux4.dev/aux4/core"
)

const noBashMessage = "aux4 runs commands with bash, and no bash was found on this Windows machine. " +
	"Install Git for Windows (https://git-scm.com/download/win), which includes Git Bash, " +
	"or set the SHELL environment variable to the full path of bash.exe"

// resolveShell picks the shell on Windows. $SHELL wins when it can be run as given (a full path,
// or a name on PATH such as "bash"). Git Bash exports SHELL=/usr/bin/bash, a path that only exists
// inside its own environment, so a POSIX-style $SHELL naming bash or sh falls back to Git's bash.
func resolveShell(shell string) (string, error) {
	if shell != "" {
		if path, ok := runnableShell(shell); ok {
			return path, nil
		}
		base := strings.TrimSuffix(strings.ToLower(filepath.Base(filepath.FromSlash(shell))), ".exe")
		if base != "bash" && base != "sh" {
			return "", core.InternalError("SHELL is set to '"+shell+"', which can't be found. "+noBashMessage, nil)
		}
	}
	return findBash()
}

func resolveBash() (string, error) {
	if shell := os.Getenv("SHELL"); shell != "" {
		base := strings.TrimSuffix(strings.ToLower(filepath.Base(filepath.FromSlash(shell))), ".exe")
		if base == "bash" {
			if path, ok := runnableShell(shell); ok {
				return path, nil
			}
		}
	}
	return findBash()
}

// runnableShell resolves a shell given as a full Windows path or as a command name on PATH. The
// bash.exe in System32 (and the WindowsApps alias) is the WSL launcher, which would run the command
// inside a Linux VM with different paths, so it is never picked by name.
func runnableShell(shell string) (string, bool) {
	if strings.HasPrefix(shell, "/") {
		return "", false
	}
	if strings.ContainsAny(shell, `\/:`) {
		if info, err := os.Stat(shell); err == nil && !info.IsDir() {
			return shell, true
		}
		if info, err := os.Stat(shell + ".exe"); err == nil && !info.IsDir() {
			return shell + ".exe", true
		}
		return "", false
	}
	path, err := exec.LookPath(shell)
	if err != nil || isWslLauncher(path) {
		return "", false
	}
	return path, true
}

func isWslLauncher(path string) bool {
	lower := strings.ToLower(path)
	if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" && strings.HasPrefix(lower, strings.ToLower(filepath.Join(systemRoot, "System32"))+`\`) {
		return true
	}
	return strings.Contains(lower, `\windowsapps\`)
}

// findBash looks for Git for Windows' bash: next to git.exe on PATH, then in the default install
// locations, then any other bash.exe on PATH (MSYS2, Cygwin) that isn't the WSL launcher.
func findBash() (string, error) {
	var candidates []string
	if git, err := exec.LookPath("git"); err == nil {
		// git.exe lives in <root>\cmd, <root>\bin or <root>\mingw64\bin
		dir := filepath.Dir(git)
		for _, root := range []string{filepath.Dir(dir), filepath.Dir(filepath.Dir(dir))} {
			candidates = append(candidates, filepath.Join(root, "bin", "bash.exe"), filepath.Join(root, "usr", "bin", "bash.exe"))
		}
	}
	for _, env := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
		if dir := os.Getenv(env); dir != "" {
			candidates = append(candidates, filepath.Join(dir, "Git", "bin", "bash.exe"))
		}
	}
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		candidates = append(candidates, filepath.Join(dir, "Programs", "Git", "bin", "bash.exe"))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	if path, ok := runnableShell("bash"); ok {
		return path, nil
	}
	return "", core.InternalError(noBashMessage, nil)
}
