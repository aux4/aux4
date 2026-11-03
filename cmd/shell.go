package cmd

import (
	"os"
	"sync"
)

// Shell and Bash are resolved once per process and cached — cheap for the normal CLI invocation,
// which only ever runs once. The daemon breaks that assumption: it serves every client request
// in the same process, setting that client's environment (including $SHELL) right before running
// the command (see daemon.Server.handleConnection), so the resolution must be retaken per request
// rather than reused from whichever request happened to run first. ResolveShell/ResolveBash give
// the daemon that hook, mirroring output.ResolveColor().
var (
	shellMutex    sync.RWMutex
	shellResolved bool
	shellPath     string
	shellErr      error

	bashMutex    sync.RWMutex
	bashResolved bool
	bashPath     string
	bashErr      error
)

// Shell returns the shell aux4 runs command lines with ($SHELL -c "<command line>").
//
// On macOS and Linux it is $SHELL, or /bin/sh when $SHELL is not set. On Windows there is no
// /bin/sh: aux4 command lines are POSIX shell, so it uses bash — $SHELL when it points to a bash
// that can be run, otherwise Git for Windows' bash — and returns a clear error when there is none.
// The decision is resolved the first time it is needed and cached for the rest of the process;
// the daemon calls ResolveShell() per request to replace that cache with the current client's
// environment.
func Shell() (string, error) {
	shellMutex.RLock()
	if shellResolved {
		defer shellMutex.RUnlock()
		return shellPath, shellErr
	}
	shellMutex.RUnlock()

	return ResolveShell()
}

// ResolveShell recomputes the shell from the current $SHELL, replacing the cached value. The
// daemon uses it: each request runs with the client's environment, so the decision has to be
// taken again per request instead of reusing the one cached when the daemon (or its first
// request) started.
func ResolveShell() (string, error) {
	path, err := resolveShell(os.Getenv("SHELL"))

	shellMutex.Lock()
	shellPath, shellErr = path, err
	shellResolved = true
	shellMutex.Unlock()

	return path, err
}

// Bash returns the bash used by the alias: executor ("bash" on macOS and Linux, found on PATH;
// Git for Windows' bash on Windows, or $SHELL on Windows when it already points to a runnable
// bash). Cached the same way as Shell, with ResolveBash as the daemon's per-request hook.
func Bash() (string, error) {
	bashMutex.RLock()
	if bashResolved {
		defer bashMutex.RUnlock()
		return bashPath, bashErr
	}
	bashMutex.RUnlock()

	return ResolveBash()
}

// ResolveBash recomputes the bash path from the current environment, replacing the cached value.
// Same per-request rationale as ResolveShell.
func ResolveBash() (string, error) {
	path, err := resolveBash()

	bashMutex.Lock()
	bashPath, bashErr = path, err
	bashResolved = true
	bashMutex.Unlock()

	return path, err
}
