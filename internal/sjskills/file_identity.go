//go:build !windows

package sjskills

import "os"

// lstatIdentity captures an identity that remains valid after a rename or replacement.
func lstatIdentity(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}
