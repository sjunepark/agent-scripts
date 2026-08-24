package sjskills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const ManifestFileName = "sjskills.toml"

// ProjectRoot identifies the nearest project manifest found from a file or
// directory. Both paths are absolute; Root is canonicalized through existing
// symlinks so callers can safely use it as the project identity.
type ProjectRoot struct {
	Root         string
	ManifestPath string
}

// DiscoverProjectRoot walks from start toward the filesystem root without
// changing anything. A regular file starts the walk at its parent directory;
// a directory starts the walk at itself. The nearest manifest wins.
func DiscoverProjectRoot(start string) (ProjectRoot, error) {
	absolute, err := filepath.Abs(start)
	if err != nil {
		return ProjectRoot{}, manifestMissingError(start)
	}
	canonicalStart, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return ProjectRoot{}, manifestMissingError(absolute)
	}
	info, err := os.Stat(canonicalStart)
	if err != nil {
		return ProjectRoot{}, manifestMissingError(absolute)
	}
	current := canonicalStart
	if !info.IsDir() {
		current = filepath.Dir(current)
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return ProjectRoot{}, manifestMissingError(current)
	}

	for {
		manifestPath := filepath.Join(current, ManifestFileName)
		manifestInfo, statErr := os.Lstat(manifestPath)
		switch {
		case statErr == nil:
			if !manifestInfo.Mode().IsRegular() {
				return ProjectRoot{}, &ValidationErrors{Issues: []Issue{{
					Code: IssueMalformedInput, Path: manifestPath, Message: "sjskills.toml must be a regular file",
				}}}
			}
			return ProjectRoot{Root: current, ManifestPath: manifestPath}, nil
		case !errors.Is(statErr, os.ErrNotExist):
			return ProjectRoot{}, fmt.Errorf("inspect manifest %s: %w", manifestPath, statErr)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return ProjectRoot{}, manifestMissingError(start)
		}
		current = parent
	}
}

func manifestMissingError(start string) error {
	return &ValidationErrors{Issues: []Issue{{
		Code:    IssueManifestMissing,
		Path:    start,
		Message: fmt.Sprintf("no %s found from %s upward", ManifestFileName, start),
	}}}
}
