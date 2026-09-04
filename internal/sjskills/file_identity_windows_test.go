//go:build windows

package sjskills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsIdentityPath(t *testing.T) {
	longSuffix := strings.Repeat(`nested\`, 45) + "demo"
	for _, test := range []struct{ name, path, want string }{
		{"short", `C:\skills\demo`, `C:\skills\demo`},
		{"drive", `C:\skills\` + longSuffix, `\\?\C:\skills\` + longSuffix},
		{"normalized", `C:/skills/old/../` + longSuffix, `\\?\C:\skills\` + longSuffix},
		{"long input short result", `C:\skills\` + strings.Repeat(`old\..\`, 45) + "demo", `\\?\C:\skills\demo`},
		{"unc", `\\server\share\` + longSuffix, `\\?\UNC\server\share\` + longSuffix},
		{"extended drive", `\\?\C:\skills\demo`, `\\?\C:\skills\demo`},
		{"extended unc", `\\?\UNC\server\share\demo`, `\\?\UNC\server\share\demo`},
		{"device", `\\.\C:\skills\demo`, `\\.\C:\skills\demo`},
		{"nt", `\??\C:\skills\demo`, `\??\C:\skills\demo`},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := windowsIdentityPath(test.path)
			if err != nil || got != test.want {
				t.Fatalf("windowsIdentityPath(%q) = %q, %v; want %q", test.path, got, err, test.want)
			}
		})
	}
	if _, err := windowsIdentityPath(""); err == nil {
		t.Fatal("empty path unexpectedly accepted")
	}
	relative := filepath.Join("relative", longSuffix)
	abs, err := filepath.Abs(relative)
	if err != nil {
		t.Fatal(err)
	}
	want, err := windowsIdentityPath(abs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := windowsIdentityPath(relative)
	if err != nil || got != want {
		t.Fatalf("relative path = %q, %v; want %q", got, err, want)
	}
}

func TestLstatIdentityWindowsLongPath(t *testing.T) {
	root := t.TempDir()
	longParent := root
	for len(longParent) < 320 {
		longParent = filepath.Join(longParent, strings.Repeat("nested", 6))
	}
	if err := os.MkdirAll(longParent, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []bool{false, true} {
		name := "file"
		if directory {
			name = "directory"
		}
		t.Run(name, func(t *testing.T) {
			source := filepath.Join(longParent, name)
			var err error
			if directory {
				err = os.Mkdir(source, 0o755)
			} else {
				err = os.WriteFile(source, []byte("content"), 0o644)
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := lstatIdentity(source)
			if err != nil {
				t.Fatal(err)
			}
			destination := source + "-moved"
			if err := os.Rename(source, destination); err != nil {
				t.Fatal(err)
			}
			after, err := os.Lstat(destination)
			if err != nil || !os.SameFile(before, after) {
				t.Fatalf("long-path identity lost after move: %v", err)
			}
		})
	}
}
