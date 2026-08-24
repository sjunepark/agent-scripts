//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package sjskills

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

var errApplyLockHeld = errors.New("apply lock is held")

func lockApplyFile(file *os.File) error {
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return errApplyLockHeld
		}
		return err
	}
	return nil
}

func unlockApplyFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}

func createApplyLockFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
}

func openExistingApplyLockFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, 0)
}

func validApplyPrivateFileMode(mode os.FileMode) bool {
	return mode.Perm() == 0o600
}
