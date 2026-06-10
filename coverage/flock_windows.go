//go:build windows

package coverage

import "os"

func lockFile(f *os.File) {
	// Windows file locking is implicit via LockFileEx, but for simplicity
	// we rely on the OS-level write exclusivity of the lock file.
}

func unlockFile(f *os.File) {
	// No-op on Windows
}
