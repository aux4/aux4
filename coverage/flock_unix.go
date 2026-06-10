//go:build !windows

package coverage

import (
	"os"
	"syscall"
)

func lockFile(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

func unlockFile(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
