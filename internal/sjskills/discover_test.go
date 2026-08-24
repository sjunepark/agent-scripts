package sjskills

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeDiscoveryManifest(t *testing.T, directory string) string {
	t.Helper()
	path := filepath.Join(directory, ManifestFileName)
	if err := os.WriteFile(path, []byte("version = 1\nprofiles = [\"dev\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func TestDiscoverProjectRootFromNestedDirectoryAndFile(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := writeDiscoveryManifest(t, root)
	discovered, err := DiscoverProjectRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if discovered.Root != canonicalRoot || discovered.ManifestPath != manifestPath {
		t.Fatalf("discovered = %#v, want root %q manifest %q", discovered, canonicalRoot, manifestPath)
	}

	filePath := filepath.Join(nested, "input.txt")
	if err := os.WriteFile(filePath, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	fromFile, err := DiscoverProjectRoot(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if fromFile != discovered {
		t.Fatalf("file discovery = %#v, directory discovery = %#v", fromFile, discovered)
	}
}

func TestDiscoverProjectRootChoosesNearestManifest(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	deep := filepath.Join(nested, "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDiscoveryManifest(t, root)
	nestedManifest := writeDiscoveryManifest(t, nested)
	discovered, err := DiscoverProjectRoot(deep)
	if err != nil {
		t.Fatal(err)
	}
	canonicalNested, err := filepath.EvalSymlinks(nested)
	if err != nil {
		t.Fatal(err)
	}
	if discovered.Root != canonicalNested || discovered.ManifestPath != nestedManifest {
		t.Fatalf("nearest discovery = %#v, want root %q manifest %q", discovered, canonicalNested, nestedManifest)
	}
}

func TestDiscoverProjectRootMissingIsTyped(t *testing.T) {
	_, err := DiscoverProjectRoot(t.TempDir())
	if err == nil {
		t.Fatal("missing manifest unexpectedly succeeded")
	}
	var validation *ValidationErrors
	if !errors.As(err, &validation) || !issueCode(err, IssueManifestMissing) {
		t.Fatalf("missing manifest error = %v", err)
	}
}

func TestDiscoverProjectRootRejectsSymlinkedManifest(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), ManifestFileName)
	if err := os.WriteFile(target, []byte("version = 1\nprofiles = [\"dev\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, ManifestFileName)); err != nil {
		t.Skipf("create manifest symlink: %v", err)
	}
	_, err := DiscoverProjectRoot(root)
	if err == nil || !issueCode(err, IssueMalformedInput) {
		t.Fatalf("symlinked manifest error = %v, want malformed input", err)
	}
}
