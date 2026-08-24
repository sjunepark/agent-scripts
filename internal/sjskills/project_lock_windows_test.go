//go:build windows

package sjskills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsApplyLockCanBeRemovedWhileHeld(t *testing.T) {
	path := filepath.Join(t.TempDir(), "apply.lock")
	file, err := createApplyLockFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := lockApplyFile(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("delete-sharing lock could not be removed while held: %v", err)
	}
	if err := unlockApplyFile(file); err != nil {
		t.Fatal(err)
	}
}
