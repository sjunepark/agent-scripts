//go:build windows

package sjskills

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

const (
	moveFileWriteThrough    = 0x8
	moveFileReplaceExisting = 0x1
)

func publishNoReplace(source, destination string) error {
	from, err := windows.UTF16PtrFromString(filepath.Clean(source))
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(filepath.Clean(destination))
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, moveFileWriteThrough)
}

func replaceFileAtomic(source, destination string) error {
	from, err := windows.UTF16PtrFromString(filepath.Clean(source))
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(filepath.Clean(destination))
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, moveFileReplaceExisting|moveFileWriteThrough)
}

func syncApplyDirectory(path string) error {
	// Windows does not provide a portable directory fsync through this small
	// adapter. File contents are flushed before publication; the state replace
	// remains atomic and failure is reported by the file operation itself.
	return nil
}
