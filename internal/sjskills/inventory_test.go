package sjskills

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func inventoryTestLayout(t *testing.T) (string, DerivedLayout) {
	t.Helper()
	root := t.TempDir()
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = canonical
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	return root, layout
}

func inventoryRoot(t *testing.T, inventory ProjectInventory, target Target) ProjectSkillsRootInventory {
	t.Helper()
	root, ok := inventory.RootFor(target)
	if !ok {
		t.Fatalf("missing root %s in %#v", target, inventory.Roots)
	}
	return root
}

func inventoryProblem(root ProjectSkillsRootInventory, reason InventoryProblemReason) bool {
	for _, problem := range root.Problems {
		if problem.Reason == reason {
			return true
		}
	}
	return false
}

func inventoryStateProblem(inventory ProjectInventory, reason InventoryProblemReason) bool {
	for _, problem := range inventory.Problems {
		if problem.Reason == reason && problem.Path == inventory.StatePath {
			return true
		}
	}
	return false
}

func findInventoryEntry(t *testing.T, root ProjectSkillsRootInventory, name string) InventoryEntry {
	t.Helper()
	for _, entry := range root.Entries {
		if entry.Name == name {
			return entry
		}
	}
	t.Fatalf("missing entry %q in %#v", name, root.Entries)
	return InventoryEntry{}
}

func TestInspectProjectMissingModeledStateIsSafeAndExact(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	before := snapshotInventoryTestTree(t, root)
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.Root != root || inventory.StatePath != layout.ReconcilerStatePath {
		t.Fatalf("boundary = %#v", inventory)
	}
	if len(inventory.Roots) != 2 {
		t.Fatalf("roots = %#v, want exactly two", inventory.Roots)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		rootInventory := inventoryRoot(t, inventory, target)
		if rootInventory.Exists || !rootInventory.Safe || len(rootInventory.Entries) != 0 || len(rootInventory.Problems) != 0 {
			t.Fatalf("missing %s root = %#v", target, rootInventory)
		}
	}
	if !inventory.StateTrusted || inventory.State.Version != ProvenanceStateVersion || inventory.State.Records == nil || len(inventory.State.Records) != 0 {
		t.Fatalf("missing state = %#v trusted=%v", inventory.State, inventory.StateTrusted)
	}
	if len(inventory.Problems) != 0 {
		t.Fatalf("missing-state problems = %#v", inventory.Problems)
	}
	assertInventoryTestTreeUnchanged(t, before)
}

func TestInspectProjectEnumeratesOnlyTwoRootsInDeterministicOrder(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	for _, path := range []string{layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"zeta", "alpha", "middle"} {
		path := filepath.Join(layout.AgentsSkillsPath, name)
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(layout.ClaudeSkillsPath, "claude-only"), 0o755); err != nil {
		t.Fatal(err)
	}

	first, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	second, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("non-deterministic inventory:\n%s\n%s", firstJSON, secondJSON)
	}
	if got := []Target{first.Roots[0].Target, first.Roots[1].Target}; !reflect.DeepEqual(got, []Target{TargetAgents, TargetClaude}) {
		t.Fatalf("root order = %#v", got)
	}
	agents := inventoryRoot(t, first, TargetAgents)
	gotNames := make([]string, 0, len(agents.Entries))
	for _, entry := range agents.Entries {
		gotNames = append(gotNames, entry.Name)
		if entry.Kind != InventoryEntryDirectory || entry.Hash == nil || entry.Hash.Digest == "" {
			t.Fatalf("entry = %#v", entry)
		}
	}
	wantNames := []string{"alpha", "middle", "zeta"}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("entry order = %#v, want %#v", gotNames, wantNames)
	}
	if len(first.Roots) != 2 || first.Roots[0].Path != layout.AgentsSkillsPath || first.Roots[1].Path != layout.ClaudeSkillsPath {
		t.Fatalf("inventory modeled paths = %#v", first.Roots)
	}
	if filepath.Clean(first.Roots[0].Path) == filepath.Clean(root) {
		t.Fatal("inventory unexpectedly modeled project root")
	}
}

func TestInspectProjectRejectsLayoutMismatchAndNonCanonicalRoot(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	layout.AgentsSkillsPath = filepath.Join(root, "elsewhere", ManagedSkillsDirectoryName)
	if _, err := InspectProject(layout); err == nil || !issueCode(err, IssuePathEscape) {
		t.Fatalf("layout mismatch error = %v", err)
	}

	parent := t.TempDir()
	canonicalRoot := filepath.Join(parent, "canonical")
	if err := os.Mkdir(canonicalRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	linkedRoot := filepath.Join(parent, "linked")
	if err := os.Symlink(canonicalRoot, linkedRoot); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	linkedLayout, err := LayoutForProject(linkedRoot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InspectProject(linkedLayout); err == nil || !issueCode(err, IssueInvalidRoot) {
		t.Fatalf("non-canonical root error = %v", err)
	}
}

func TestInspectProjectRefusesUnsafeRootAncestorsWithoutReadingOutside(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "sentinel")
	if err := os.WriteFile(sentinel, []byte("outside-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, string(TargetAgents))); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layout.ClaudeSkillsPath, "safe"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.ClaudeSkillsPath, "safe", "file"), []byte("safe"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := snapshotInventoryTestTree(t, root)
	outsideBefore := snapshotInventoryTestTree(t, outside)
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	agents := inventoryRoot(t, inventory, TargetAgents)
	if agents.Safe || len(agents.Entries) != 0 || !inventoryProblem(agents, InventoryProblemRootUnsafeAncestor) {
		t.Fatalf("unsafe agents root = %#v", agents)
	}
	claude := inventoryRoot(t, inventory, TargetClaude)
	if !claude.Safe || len(claude.Entries) != 1 {
		t.Fatalf("safe claude root = %#v", claude)
	}
	assertInventoryTestTreeUnchanged(t, before)
	assertInventoryTestTreeUnchanged(t, outsideBefore)

	if err := os.Remove(filepath.Join(root, string(TargetAgents))); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(root, "inside-target")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inside, "sentinel"), []byte("inside-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(inside, filepath.Join(root, string(TargetAgents))); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	inventory, err = InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	agents = inventoryRoot(t, inventory, TargetAgents)
	if agents.Safe || len(agents.Entries) != 0 || !inventoryProblem(agents, InventoryProblemRootUnsafeAncestor) {
		t.Fatalf("inside unsafe agents root = %#v", agents)
	}
}

func TestInspectProjectRejectsRootFileAndSpecialEntry(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(filepath.Dir(layout.AgentsSkillsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.AgentsSkillsPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	agents := inventoryRoot(t, inventory, TargetAgents)
	if !agents.Exists || agents.Safe || len(agents.Entries) != 0 || !inventoryProblem(agents, InventoryProblemRootNotDirectory) {
		t.Fatalf("file root = %#v", agents)
	}

	if err := os.Remove(layout.AgentsSkillsPath); err != nil {
		t.Fatal(err)
	}
	if !createSpecialEntry(t, layout.AgentsSkillsPath) {
		return
	}
	inventory, err = InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	agents = inventoryRoot(t, inventory, TargetAgents)
	if !agents.Exists || agents.Safe || len(agents.Entries) != 0 || !inventoryProblem(agents, InventoryProblemRootNotDirectory) {
		t.Fatalf("special root = %#v", agents)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatal(err)
	}
}

func TestInspectProjectHashesDirectoriesAndExecutableBits(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(layout.AgentsSkillsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(layout.AgentsSkillsPath, "demo")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(directory, "run.sh")
	if err := os.WriteFile(file, []byte("#!/bin/sh\necho demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	entry := findInventoryEntry(t, inventoryRoot(t, inventory, TargetAgents), "demo")
	if entry.Hash == nil || entry.Hash.Algorithm != TreeHashAlgorithmSHA256V2 || entry.Hash.Digest == "" || entry.Problem != "" {
		t.Fatalf("non-executable entry = %#v", entry)
	}
	want, err := HashSkillTree(directory)
	if err != nil || entry.Hash == nil || *entry.Hash != want {
		t.Fatalf("inventory hash = %#v, direct = %#v, err = %v", entry.Hash, want, err)
	}
	if err := os.Chmod(file, 0o755); err != nil {
		t.Fatal(err)
	}
	changed, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	changedEntry := findInventoryEntry(t, inventoryRoot(t, changed, TargetAgents), "demo")
	if changedEntry.Hash == nil || entry.Hash == nil || *changedEntry.Hash == *entry.Hash {
		t.Fatalf("executable-bit change did not change hash: %#v", changedEntry.Hash)
	}
}

func TestInventoryEntryHashesAreOptionalInJSONAndDetached(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(layout.AgentsSkillsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(layout.AgentsSkillsPath, "demo")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.AgentsSkillsPath, "plain"), []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(layout.AgentsSkillsPath, "link")
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}

	var encoded struct {
		Roots []struct {
			Entries []map[string]json.RawMessage `json:"entries"`
		} `json:"roots"`
	}
	data, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &encoded); err != nil {
		t.Fatal(err)
	}
	for _, entry := range inventoryRoot(t, inventory, TargetAgents).Entries {
		if entry.Name == "demo" {
			continue
		}
		var wire map[string]json.RawMessage
		for _, root := range encoded.Roots {
			for _, candidate := range root.Entries {
				var name string
				if err := json.Unmarshal(candidate["name"], &name); err == nil && name == entry.Name {
					wire = candidate
				}
			}
		}
		if _, present := wire["hash"]; present {
			t.Fatalf("unverified %q unexpectedly serialized a hash: %s", entry.Name, data)
		}
	}

	original := inventoryRoot(t, inventory, TargetAgents)
	copyRoot, ok := inventory.RootFor(TargetAgents)
	if !ok {
		t.Fatal("RootFor did not return agents root")
	}
	originalEntry := findInventoryEntry(t, original, "demo")
	copyEntry := findInventoryEntry(t, copyRoot, "demo")
	if originalEntry.Hash == nil || copyEntry.Hash == nil || originalEntry.Hash == copyEntry.Hash {
		t.Fatalf("hash pointer was not detached: original=%p copy=%p", originalEntry.Hash, copyEntry.Hash)
	}
	for index := range copyRoot.Entries {
		if copyRoot.Entries[index].Name == "demo" {
			copyRoot.Entries[index].Hash.Digest = strings.Repeat("f", 64)
		}
	}
	if findInventoryEntry(t, inventoryRoot(t, inventory, TargetAgents), "demo").Hash.Digest == strings.Repeat("f", 64) {
		t.Fatal("RootFor hash mutation changed original inventory")
	}

	detached := detachProjectInventory(inventory)
	for index := range detached.Roots[0].Entries {
		if detached.Roots[0].Entries[index].Name == "demo" {
			detached.Roots[0].Entries[index].Hash.Digest = strings.Repeat("e", 64)
		}
	}
	if findInventoryEntry(t, inventoryRoot(t, inventory, TargetAgents), "demo").Hash.Digest == strings.Repeat("e", 64) {
		t.Fatal("detached inventory hash mutation changed original inventory")
	}
}

func TestInspectProjectDoesNotFollowChildSymlink(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(layout.AgentsSkillsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "sentinel")
	if err := os.WriteFile(sentinel, []byte("must not be opened"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(layout.AgentsSkillsPath, "external-link")
	if err := os.Symlink(sentinel, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	entry := findInventoryEntry(t, inventoryRoot(t, inventory, TargetAgents), "external-link")
	if entry.Kind != InventoryEntrySymlink || entry.LinkTarget != sentinel || entry.Hash != nil || entry.Problem != InventoryProblemSymlinkNotHashed {
		t.Fatalf("symlink entry = %#v", entry)
	}
}

func TestInspectProjectIncludesMalformedChildName(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(layout.AgentsSkillsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "Not_portable"
	if err := os.Mkdir(filepath.Join(layout.AgentsSkillsPath, name), 0o755); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	entry := findInventoryEntry(t, inventoryRoot(t, inventory, TargetAgents), name)
	if entry.Problem != InventoryProblemEntryInvalidName || entry.Kind != InventoryEntryDirectory || entry.Hash == nil || entry.Hash.Digest == "" {
		t.Fatalf("malformed entry = %#v", entry)
	}
	if !inventoryProblem(inventoryRoot(t, inventory, TargetAgents), InventoryProblemEntryInvalidName) {
		t.Fatalf("malformed child problem missing: %#v", inventory.Problems)
	}
}

func TestInspectProjectLoadsAndSortsStrictProvenance(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(filepath.Dir(layout.ReconcilerStatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	state := ProvenanceState{
		Version: ProvenanceStateVersion,
		Records: []ProvenanceRecord{
			validInventoryProvenanceRecord("zeta", TargetClaude),
			validInventoryProvenanceRecord("alpha", TargetAgents),
			validInventoryProvenanceRecord("alpha", TargetClaude),
		},
	}
	writeInventoryState(t, layout.ReconcilerStatePath, state)
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if !inventory.StateTrusted || len(inventory.State.Records) != len(state.Records) || len(inventory.Problems) != 0 {
		t.Fatalf("trusted state = %#v trusted=%v problems=%#v", inventory.State, inventory.StateTrusted, inventory.Problems)
	}
	gotKeys := make([]string, 0, len(inventory.State.Records))
	for _, record := range inventory.State.Records {
		gotKeys = append(gotKeys, string(record.Target)+"/"+record.Skill)
	}
	wantKeys := []string{".agents/alpha", ".claude/alpha", ".claude/zeta"}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("records = %#v, want %#v", gotKeys, wantKeys)
	}
	copyRecords := inventory.Records()
	copyRecords[0].Skill = "changed"
	if inventory.State.Records[0].Skill == "changed" {
		t.Fatal("Records returned an attached slice")
	}
}

func TestInspectProjectFailsClosedForMalformedProvenanceClasses(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(filepath.Dir(layout.ReconcilerStatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	validRecordJSON := `{"scope":"project","skill":"demo","target":".agents","sourceIdentity":"github:example/demo","treeHashAlgorithm":"tree-sha256-v2","treeHash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","recordedAt":"2026-08-24T09:30:00Z"}`
	unknownRecordFieldJSON := strings.TrimSuffix(validRecordJSON, "}") + `,"extra":true}`
	cases := []struct {
		name string
		data string
	}{
		{name: "unknown-field", data: `{"version":1,"records":[],"extra":true}`},
		{name: "unknown-record-field", data: `{"version":1,"records":[` + unknownRecordFieldJSON + `]}`},
		{name: "version", data: `{"version":2,"records":[]}`},
		{name: "nil-records", data: `{"version":1,"records":null}`},
		{name: "bad-scope", data: `{"version":1,"records":[{"scope":"global","skill":"demo","target":".agents","sourceIdentity":"x","treeHashAlgorithm":"tree-sha256-v2","treeHash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","recordedAt":"2026-08-24T09:30:00Z"}]}`},
		{name: "bad-target", data: strings.Replace(validRecordJSON, `"target":".agents"`, `"target":"other"`, 1)},
		{name: "bad-name", data: strings.Replace(validRecordJSON, `"skill":"demo"`, `"skill":"Bad_Name"`, 1)},
		{name: "bad-source", data: strings.Replace(validRecordJSON, `"sourceIdentity":"github:example/demo"`, `"sourceIdentity":""`, 1)},
		{name: "bad-hash", data: strings.Replace(validRecordJSON, `"treeHash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`, `"treeHash":"ABC"`, 1)},
		{name: "bad-algorithm", data: strings.Replace(validRecordJSON, `"treeHashAlgorithm":"tree-sha256-v2"`, `"treeHashAlgorithm":"sha256"`, 1)},
		{name: "bad-time", data: strings.Replace(validRecordJSON, `"recordedAt":"2026-08-24T09:30:00Z"`, `"recordedAt":"0001-01-01T00:00:00Z"`, 1)},
		{name: "trailing-document", data: `{"version":1,"records":[]} {}`},
		{name: "duplicate", data: `{"version":1,"records":[` + validRecordJSON + "," + validRecordJSON + `]}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := os.WriteFile(layout.ReconcilerStatePath, []byte(testCase.data), 0o644); err != nil {
				t.Fatal(err)
			}
			inventory, err := InspectProject(layout)
			if err != nil {
				t.Fatal(err)
			}
			if inventory.StateTrusted || len(inventory.State.Records) != 0 || !inventoryStateProblem(inventory, InventoryProblemStateInvalid) {
				t.Fatalf("malformed %s inventory = %#v trusted=%v", testCase.name, inventory, inventory.StateTrusted)
			}
		})
	}
}

func TestInspectProjectRejectsOversizedProvenanceState(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(filepath.Dir(layout.ReconcilerStatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	data := bytes.Repeat([]byte("x"), int(maxProvenanceStateBytes)+1)
	if err := os.WriteFile(layout.ReconcilerStatePath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.StateTrusted || len(inventory.State.Records) != 0 || !inventoryStateProblem(inventory, InventoryProblemStateUnreadable) {
		t.Fatalf("oversized state inventory = %#v trusted=%v", inventory, inventory.StateTrusted)
	}
}

func TestInspectProjectRefusesUnsafeOrSpecialProvenanceState(t *testing.T) {
	root, layout := inventoryTestLayout(t)
	outside := t.TempDir()
	outsideState := filepath.Join(outside, "state.json")
	if err := os.WriteFile(outsideState, []byte(`{"version":1,"records":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideState, layout.ReconcilerStatePath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	outsideBefore := snapshotInventoryTestTree(t, outside)
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.StateTrusted || len(inventory.State.Records) != 0 || !inventoryStateProblem(inventory, InventoryProblemStateUnsafeAncestor) {
		t.Fatalf("state symlink inventory = %#v", inventory)
	}
	assertInventoryTestTreeUnchanged(t, outsideBefore)

	if err := os.Remove(layout.ReconcilerStatePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Dir(layout.ReconcilerStatePath)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Dir(layout.ReconcilerStatePath)); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	inventory, err = InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.StateTrusted || len(inventory.State.Records) != 0 || !inventoryStateProblem(inventory, InventoryProblemStateUnsafeAncestor) {
		t.Fatalf("state ancestor symlink inventory = %#v", inventory)
	}

	if err := os.Remove(filepath.Join(root, DerivedDirectoryName)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, DerivedDirectoryName), 0o755); err != nil {
		t.Fatal(err)
	}
	if !createSpecialEntry(t, layout.ReconcilerStatePath) {
		return
	}
	inventory, err = InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.StateTrusted || !inventoryStateProblem(inventory, InventoryProblemStateNotRegular) {
		t.Fatalf("special state inventory = %#v", inventory)
	}
}

func validInventoryProvenanceRecord(skill string, target Target) ProvenanceRecord {
	return ProvenanceRecord{
		Scope:             ScopeProject,
		Skill:             skill,
		Target:            target,
		SourceIdentity:    "github:example/" + skill,
		TreeHashAlgorithm: TreeHashAlgorithmSHA256V2,
		TreeHash:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		RecordedAt:        time.Date(2026, time.August, 24, 9, 30, 0, 0, time.UTC),
	}
}

func writeInventoryState(t *testing.T, path string, state ProvenanceState) {
	t.Helper()
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

type inventoryTestSnapshot struct {
	path string
	data []byte
	mode os.FileMode
}

func snapshotInventoryTestTree(t *testing.T, root string) []inventoryTestSnapshot {
	t.Helper()
	var snapshots []inventoryTestSnapshot
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		snapshots = append(snapshots, inventoryTestSnapshot{path: path, data: data, mode: info.Mode()})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].path < snapshots[j].path })
	return snapshots
}

func assertInventoryTestTreeUnchanged(t *testing.T, snapshots []inventoryTestSnapshot) {
	t.Helper()
	for _, snapshot := range snapshots {
		info, err := os.Lstat(snapshot.path)
		if err != nil {
			t.Fatalf("snapshot path %q changed: %v", snapshot.path, err)
		}
		if info.Mode() != snapshot.mode {
			t.Fatalf("snapshot mode %q changed: got %v want %v", snapshot.path, info.Mode(), snapshot.mode)
		}
		data, err := os.ReadFile(snapshot.path)
		if err != nil {
			t.Fatalf("snapshot read %q: %v", snapshot.path, err)
		}
		if !reflect.DeepEqual(data, snapshot.data) {
			t.Fatalf("snapshot bytes %q changed", snapshot.path)
		}
	}
}

func TestInspectProjectDoesNotReturnRawFilesystemErrors(t *testing.T) {
	_, layout := inventoryTestLayout(t)
	if err := os.MkdirAll(layout.AgentsSkillsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "unreadable"
	if err := os.Mkdir(filepath.Join(layout.AgentsSkillsPath, name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.AgentsSkillsPath, name, "fifo"), []byte("not special"), 0o644); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range inventory.Problems {
		if strings.Contains(string(problem.Reason), string(filepath.Separator)) || strings.Contains(string(problem.Reason), "permission") || strings.Contains(string(problem.Reason), "no such") {
			t.Fatalf("unbounded/raw problem = %#v", problem)
		}
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal("unexpected filesystem error returned")
	}
}
