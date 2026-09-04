//go:build windows

package sjskills

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// lstatIdentity captures an identity that remains valid after a rename or replacement.
// Windows os.Lstat defers reading the file ID until os.SameFile, which may then
// inspect a missing or reused path. File.Stat reads the ID from a handle eagerly.
func lstatIdentity(path string) (os.FileInfo, error) {
	openPath, err := windowsIdentityPath(path)
	if err != nil {
		return nil, &os.PathError{Op: "lstatIdentity", Path: path, Err: err}
	}
	name, err := windows.UTF16PtrFromString(openPath)
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

func windowsIdentityPath(path string) (string, error) {
	if path == "" {
		return "", os.ErrInvalid
	}
	normalized := filepath.FromSlash(path)
	if strings.HasPrefix(normalized, `\\?\`) || strings.HasPrefix(normalized, `\\.\`) || strings.HasPrefix(normalized, `\??\`) {
		return path, nil
	}
	abs, err := filepath.Abs(normalized)
	if err != nil {
		return "", err
	}
	// Like os.Lstat, retain ordinary Win32 semantics for short paths. Longer
	// paths need an absolute extended form even without the host's long-path opt-in.
	if len(path) < 248 && len(abs) < 248 {
		return path, nil
	}
	if strings.HasPrefix(abs, `\\`) {
		return `\\?\UNC\` + abs[2:], nil
	}
	return `\\?\` + abs, nil
}
