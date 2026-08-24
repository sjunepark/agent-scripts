package sjskills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLayoutMapsDerivedProjectPaths(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project with spaces")
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root != filepath.Clean(root) {
		t.Fatalf("layout root = %q, want %q", layout.Root, filepath.Clean(root))
	}
	if layout.ManifestPath != filepath.Join(layout.Root, ManifestFileName) {
		t.Fatalf("manifest path = %q", layout.ManifestPath)
	}
	if layout.AgentsSkillsPath != filepath.Join(layout.Root, string(TargetAgents), ManagedSkillsDirectoryName) {
		t.Fatalf("agents path = %q", layout.AgentsSkillsPath)
	}
	if layout.ClaudeSkillsPath != filepath.Join(layout.Root, string(TargetClaude), ManagedSkillsDirectoryName) {
		t.Fatalf("claude path = %q", layout.ClaudeSkillsPath)
	}
	if layout.ReconcilerStatePath != filepath.Join(layout.Root, DerivedDirectoryName, ProvenanceStateFileName) {
		t.Fatalf("state path = %q", layout.ReconcilerStatePath)
	}
	if layout.QuarantinePath != filepath.Join(layout.Root, DerivedDirectoryName, QuarantineDirectoryName) {
		t.Fatalf("quarantine path = %q", layout.QuarantinePath)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		path, err := layout.ManagedSkillsPath(target)
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(path) {
			t.Fatalf("%s path is not absolute: %q", target, path)
		}
		relative, err := filepath.Rel(layout.Root, path)
		if err != nil || relative == ".." || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator) {
			t.Fatalf("%s path escaped root: %q", target, path)
		}
	}
}

func TestLayoutRejectsInvalidRootTargetAndEscapedPlacement(t *testing.T) {
	if _, err := LayoutForProject("relative/project"); err == nil || !issueCode(err, IssueInvalidRoot) {
		t.Fatalf("relative root error = %v", err)
	}
	root := t.TempDir()
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := layout.ManagedSkillsPath(Target("../escape")); err == nil || !issueCode(err, IssueInvalidTarget) {
		t.Fatalf("invalid target error = %v", err)
	}
	layout.AgentsSkillsPath = filepath.Join(filepath.Dir(root), "escape", ManagedSkillsDirectoryName)
	if _, err := layout.ManagedSkillsPath(TargetAgents); err == nil || !issueCode(err, IssuePathEscape) {
		t.Fatalf("escaped placement error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, DerivedDirectoryName)); !os.IsNotExist(err) {
		t.Fatalf("layout unexpectedly touched filesystem: stat error = %v", err)
	}
}

func TestProvenanceStateShapeIsMinimalAndVersioned(t *testing.T) {
	recordedAt := time.Date(2026, time.August, 24, 9, 30, 0, 0, time.UTC)
	state := ProvenanceState{
		Version: ProvenanceStateVersion,
		Records: []ProvenanceRecord{{
			Scope: ScopeProject, Skill: "example", Target: TargetAgents,
			SourceIdentity: "example/catalog", TreeHashAlgorithm: TreeHashAlgorithmSHA256V2,
			TreeHash: "0123456789abcdef", RecordedAt: recordedAt,
		}},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var stateShape map[string]json.RawMessage
	if err := json.Unmarshal(data, &stateShape); err != nil {
		t.Fatal(err)
	}
	if len(stateShape) != 2 {
		t.Fatalf("state has extra fields: %s", data)
	}
	var recordShape map[string]json.RawMessage
	var records []json.RawMessage
	if err := json.Unmarshal(stateShape["records"], &records); err != nil || len(records) != 1 {
		t.Fatalf("records = %s, err = %v", stateShape["records"], err)
	}
	if err := json.Unmarshal(records[0], &recordShape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"scope", "skill", "target", "sourceIdentity", "treeHashAlgorithm", "treeHash", "recordedAt"} {
		if _, ok := recordShape[field]; !ok {
			t.Fatalf("record missing %q: %s", field, records[0])
		}
	}
	for _, field := range []string{"rollout", "lock", "lockPath", "status"} {
		if _, ok := recordShape[field]; ok {
			t.Fatalf("record invented prohibited field %q: %s", field, records[0])
		}
	}
	if string(stateShape["version"]) != "1" {
		t.Fatalf("state version = %s", stateShape["version"])
	}
}
