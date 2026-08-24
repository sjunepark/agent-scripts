//go:build windows

package sjskills

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishNoReplaceWindowsPreservesCollision(t *testing.T) {
	parent := t.TempDir()
	source := filepath.Join(parent, "source")
	destination := filepath.Join(parent, "destination")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "sentinel"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := publishNoReplace(source, destination); !errors.Is(err, os.ErrExist) {
		t.Fatalf("publishNoReplace collision error = %v, want os.ErrExist", err)
	}
	if _, err := os.Stat(filepath.Join(destination, "sentinel")); err != nil {
		t.Fatalf("collision destination changed: %v", err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("collision source disappeared: %v", err)
	}
}
