package sjskills

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	GlobalProvenanceStateVersion = 2
	globalStateRelativePath      = ".agents/.global-skill-state.json"
	globalDerivedRelativePath    = ".agents/.sjskills-global"
)

// GlobalLayout is the fixed, path-only user-home boundary used by global
// reconciliation. AgentsSkillsPath and ClaudeSkillsPath are managed placement
// roots; provenance, derived state, and quarantine are private reconciler-owned
// state. Every other named location is read-only evidence or an explicitly
// protected vendor, backup, legacy, cache, or runtime boundary.
type GlobalLayout struct {
	Home                  string
	AgentsSkillsPath      string
	ClaudeSkillsPath      string
	ProvenanceStatePath   string
	DerivedStatePath      string
	QuarantinePath        string
	AgentsVendorLockPath  string
	ClaudeVendorLockPath  string
	LegacyPiSkillsPath    string
	LegacyQuarantinePath  string
	CodexSystemSkillsPath string
	CodexPluginCachePath  string
}

// LayoutForGlobal derives the complete fixed global boundary without touching
// the filesystem. The caller supplies the home explicitly so tests never need
// to inspect the process user's real home.
func LayoutForGlobal(home string) (GlobalLayout, error) {
	if home == "" || strings.ContainsRune(home, 0) || !filepath.IsAbs(home) {
		return GlobalLayout{}, &ValidationErrors{Issues: []Issue{{
			Code: IssueInvalidRoot, Path: "home", Message: "global home must be a non-empty absolute path",
		}}}
	}
	home = filepath.Clean(home)
	layout := GlobalLayout{
		Home:                  home,
		AgentsSkillsPath:      filepath.Join(home, string(TargetAgents), ManagedSkillsDirectoryName),
		ClaudeSkillsPath:      filepath.Join(home, string(TargetClaude), ManagedSkillsDirectoryName),
		ProvenanceStatePath:   filepath.Join(home, filepath.FromSlash(globalStateRelativePath)),
		DerivedStatePath:      filepath.Join(home, filepath.FromSlash(globalDerivedRelativePath)),
		QuarantinePath:        filepath.Join(home, filepath.FromSlash(globalDerivedRelativePath), QuarantineDirectoryName),
		AgentsVendorLockPath:  filepath.Join(home, string(TargetAgents), ".skill-lock.json"),
		ClaudeVendorLockPath:  filepath.Join(home, string(TargetClaude), ".skill-lock.json"),
		LegacyPiSkillsPath:    filepath.Join(home, ".pi", "agent", ManagedSkillsDirectoryName),
		LegacyQuarantinePath:  filepath.Join(home, ".skill-quarantine"),
		CodexSystemSkillsPath: filepath.Join(home, ".codex", ManagedSkillsDirectoryName, ".system"),
		CodexPluginCachePath:  filepath.Join(home, ".codex", "plugins", "cache"),
	}
	for label, candidate := range map[string]string{
		"agents skills":        layout.AgentsSkillsPath,
		"claude skills":        layout.ClaudeSkillsPath,
		"global provenance":    layout.ProvenanceStatePath,
		"global derived state": layout.DerivedStatePath,
		"global quarantine":    layout.QuarantinePath,
		"agents vendor lock":   layout.AgentsVendorLockPath,
		"claude vendor lock":   layout.ClaudeVendorLockPath,
		"legacy pi skills":     layout.LegacyPiSkillsPath,
		"legacy quarantine":    layout.LegacyQuarantinePath,
		"codex system skills":  layout.CodexSystemSkillsPath,
		"codex plugin cache":   layout.CodexPluginCachePath,
	} {
		if !pathWithin(home, candidate) || sameInspectionPath(home, candidate) {
			return GlobalLayout{}, &ValidationErrors{Issues: []Issue{{
				Code: IssuePathEscape, Path: label, Message: fmt.Sprintf("path %q escapes the global home", candidate),
			}}}
		}
	}
	return layout, nil
}

// mutationLayout maps the fixed global boundary onto the shared transaction
// engine. The project manifest path is intentionally empty: global intent is
// the embedded registry, while all mutable recovery data stays in the private
// derived directory under .agents.
func (layout GlobalLayout) mutationLayout() DerivedLayout {
	return DerivedLayout{
		Root:                 layout.Home,
		AgentsSkillsPath:     layout.AgentsSkillsPath,
		ClaudeSkillsPath:     layout.ClaudeSkillsPath,
		DerivedDirectoryPath: layout.DerivedStatePath,
		ReconcilerStatePath:  layout.ProvenanceStatePath,
		QuarantinePath:       layout.QuarantinePath,
	}
}

// ManagedSkillsPath returns one of the only two global mutation roots.
func (layout GlobalLayout) ManagedSkillsPath(target Target) (string, error) {
	var candidate string
	switch target {
	case TargetAgents:
		candidate = layout.AgentsSkillsPath
	case TargetClaude:
		candidate = layout.ClaudeSkillsPath
	default:
		return "", &ValidationErrors{Issues: []Issue{{
			Code: IssueInvalidTarget, Path: "target", Message: fmt.Sprintf("unsupported target %q", target),
		}}}
	}
	if !pathWithin(layout.Home, candidate) || sameInspectionPath(layout.Home, candidate) {
		return "", &ValidationErrors{Issues: []Issue{{
			Code: IssuePathEscape, Path: "target placement", Message: "global placement escapes the selected home",
		}}}
	}
	return candidate, nil
}
