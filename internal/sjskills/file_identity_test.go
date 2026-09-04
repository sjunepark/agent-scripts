package sjskills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLstatIdentitySurvivesMoveAndRejectsReplacement(t *testing.T) {
	for _, directory := range []bool{false, true} {
		name := "file"
		if directory {
			name = "directory"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			destination := filepath.Join(root, "destination")
			create := func(path string) {
				t.Helper()
				var err error
				if directory {
					err = os.Mkdir(path, 0o755)
				} else {
					err = os.WriteFile(path, []byte("same content"), 0o644)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			create(source)
			before, err := lstatIdentity(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := publishNoReplace(source, destination); err != nil {
				t.Fatal(err)
			}
			// Reusing the old path must not change the captured identity.
			create(source)
			after, err := os.Lstat(destination)
			if err != nil || !os.SameFile(before, after) {
				t.Fatalf("identity lost after move: %v", err)
			}
			replacement, err := os.Lstat(source)
			if err != nil || os.SameFile(before, replacement) {
				t.Fatalf("replacement accepted as original: %v", err)
			}
		})
	}
}

func TestLstatIdentityDoesNotFollowSymlink(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	info, err := lstatIdentity(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink followed: info=%v err=%v", info, err)
	}
}
