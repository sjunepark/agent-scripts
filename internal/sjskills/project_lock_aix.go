//go:build aix

package sjskills

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

var errApplyLockHeld = errors.New("apply lock is held")

func lockApplyFile(file *os.File) error {
	lock := unix.Flock_t{Type: unix.F_WRLCK, Whence: 0, Start: 0, Len: 1}
	if err := unix.FcntlFlock(file.Fd(), unix.F_SETLK, &lock); err != nil {
		if errors.Is(err, unix.EACCES) || errors.Is(err, unix.EAGAIN) {
			return errApplyLockHeld
		}
		return err
	}
	return nil
}

func unlockApplyFile(file *os.File) error {
	lock := unix.Flock_t{Type: unix.F_UNLCK, Whence: 0, Start: 0, Len: 1}
	return unix.FcntlFlock(file.Fd(), unix.F_SETLK, &lock)
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
