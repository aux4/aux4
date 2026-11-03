package cmd

import (
	"os"
	"sync"
)

var (
	shellOnce sync.Once
	shellPath string
	shellErr  error

	bashOnce sync.Once
	bashPath string
	bashErr  error
)

// Shell returns the shell aux4 runs command lines with ($SHELL -c "<command line>").
//
// On macOS and Linux it is $SHELL, or /bin/sh when $SHELL is not set. On Windows there is no
// /bin/sh: aux4 command lines are POSIX shell, so it uses bash — $SHELL when it points to a bash
// that can be run, otherwise Git for Windows' bash — and returns a clear error when there is none.
func Shell() (string, error) {
	shellOnce.Do(func() {
		shellPath, shellErr = resolveShell(os.Getenv("SHELL"))
	})
	return shellPath, shellErr
}

// Bash returns the bash used by the alias: executor ("bash" on macOS and Linux, found on PATH;
// Git for Windows' bash on Windows).
func Bash() (string, error) {
	bashOnce.Do(func() {
		bashPath, bashErr = resolveBash()
	})
	return bashPath, bashErr
}
