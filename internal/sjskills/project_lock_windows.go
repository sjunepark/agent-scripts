//go:build windows

package sjskills

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

var errApplyLockHeld error = windows.ERROR_LOCK_VIOLATION

func lockApplyFile(file *os.File) error {
	var overlapped windows.Overlapped
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return errApplyLockHeld
	}
	return err
}

func unlockApplyFile(file *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &overlapped)
}

func createApplyLockFile(path string) (*os.File, error) {
	return openApplyLockFileWindows(path, windows.CREATE_NEW)
}

func openExistingApplyLockFile(path string) (*os.File, error) {
	return openApplyLockFileWindows(path, windows.OPEN_EXISTING)
}

func openApplyLockFileWindows(path string, disposition uint32) (*os.File, error) {
	pathPointer, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(
		pathPointer,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		disposition,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		if disposition == windows.CREATE_NEW && (errors.Is(err, windows.ERROR_FILE_EXISTS) || errors.Is(err, windows.ERROR_ALREADY_EXISTS)) {
			return nil, os.ErrExist
		}
		return nil, err
	}
	return os.NewFile(uintptr(handle), path), nil
}

func validApplyPrivateFileMode(_ os.FileMode) bool {
	// Windows reports synthesized Unix permission bits. Access control is
	// provided by the containing user-owned project directory and native ACLs.
	return true
}
