package sjskills

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	DerivedDirectoryName       = ".sjskills"
	ManagedSkillsDirectoryName = "skills"
	ProvenanceStateFileName    = "state.json"
	QuarantineDirectoryName    = "quarantine"
	ProvenanceStateVersion     = 1
	TreeHashAlgorithmSHA256V2  = "tree-sha256-v2"
)

// DerivedLayout maps one already-discovered canonical project root to every
// path owned by reconciliation. The manifest is the only
// committed intent; placements and .sjskills state are generated, local
// derived state and are never another configuration source.
type DerivedLayout struct {
	Root                 string
	ManifestPath         string
	AgentsSkillsPath     string
	ClaudeSkillsPath     string
	DerivedDirectoryPath string
	ReconcilerStatePath  string
	QuarantinePath       string
}

// LayoutForProject builds a path-only layout from a canonical absolute root.
// It performs no filesystem access and uses filepath operations so separators,
// volumes, and relative calculations remain portable across macOS and Windows.
func LayoutForProject(root string) (DerivedLayout, error) {
	cleanRoot, err := validateLayoutRoot(root)
	if err != nil {
		return DerivedLayout{}, err
	}
	layout := DerivedLayout{
		Root:                 cleanRoot,
		ManifestPath:         filepath.Join(cleanRoot, ManifestFileName),
		AgentsSkillsPath:     filepath.Join(cleanRoot, string(TargetAgents), ManagedSkillsDirectoryName),
		ClaudeSkillsPath:     filepath.Join(cleanRoot, string(TargetClaude), ManagedSkillsDirectoryName),
		DerivedDirectoryPath: filepath.Join(cleanRoot, DerivedDirectoryName),
		ReconcilerStatePath:  filepath.Join(cleanRoot, DerivedDirectoryName, ProvenanceStateFileName),
		QuarantinePath:       filepath.Join(cleanRoot, DerivedDirectoryName, QuarantineDirectoryName),
	}
	if err := validateLayoutPaths(layout); err != nil {
		return DerivedLayout{}, err
	}
	return layout, nil
}

// ManagedSkillsPath returns the generated project placement for one supported
// target. Target is validated independently so callers cannot turn a
// target value into an escaped path.
func (layout DerivedLayout) ManagedSkillsPath(target Target) (string, error) {
	if err := validateLayoutRootOnly(layout.Root); err != nil {
		return "", err
	}
	var path string
	switch target {
	case TargetAgents:
		path = layout.AgentsSkillsPath
	case TargetClaude:
		path = layout.ClaudeSkillsPath
	default:
		return "", &ValidationErrors{Issues: []Issue{{
			Code: IssueInvalidTarget, Path: "target", Message: fmt.Sprintf("unsupported target %q", target),
		}}}
	}
	if err := ensureContained(layout.Root, path, "target placement"); err != nil {
		return "", err
	}
	return path, nil
}

func validateLayoutRoot(root string) (string, error) {
	if err := validateLayoutRootOnly(root); err != nil {
		return "", err
	}
	return filepath.Clean(root), nil
}

func validateLayoutRootOnly(root string) error {
	if root == "" || strings.ContainsRune(root, 0) || !filepath.IsAbs(root) {
		return &ValidationErrors{Issues: []Issue{{
			Code: IssueInvalidRoot, Path: "root", Message: "project root must be a non-empty absolute path",
		}}}
	}
	return nil
}

func validateLayoutPaths(layout DerivedLayout) error {
	paths := []struct {
		name string
		path string
	}{
		{name: "manifest", path: layout.ManifestPath},
		{name: "agents skills", path: layout.AgentsSkillsPath},
		{name: "claude skills", path: layout.ClaudeSkillsPath},
		{name: "derived directory", path: layout.DerivedDirectoryPath},
		{name: "reconciler state", path: layout.ReconcilerStatePath},
		{name: "quarantine", path: layout.QuarantinePath},
	}
	for _, candidate := range paths {
		if err := ensureContained(layout.Root, candidate.path, candidate.name); err != nil {
			return err
		}
	}
	return nil
}

func ensureContained(root, candidate, label string) error {
	if !filepath.IsAbs(candidate) {
		return pathEscapeError(label, candidate)
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return pathEscapeError(label, candidate)
	}
	return nil
}

func pathEscapeError(label, path string) error {
	return &ValidationErrors{Issues: []Issue{{
		Code: IssuePathEscape, Path: label, Message: fmt.Sprintf("path %q escapes the project root", path),
	}}}
}
