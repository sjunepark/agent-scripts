//go:build windows

package sjskills

import (
	"os"

	"golang.org/x/sys/windows"
)

// lstatIdentity captures an identity that remains valid after a rename or replacement.
// Windows os.Lstat defers reading the file ID until os.SameFile, which may then
// inspect a missing or reused path. File.Stat reads the ID from a handle eagerly.
func lstatIdentity(path string) (os.FileInfo, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "lstatIdentity", Path: path, Err: err}
	}
	handle, err := windows.CreateFile(name, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, &os.PathError{Op: "lstatIdentity", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(handle), path)
	defer file.Close()
	return file.Stat()
}
