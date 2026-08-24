package sjskills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const classificationDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func classificationHash(suffix byte) TreeHash {
	digest := []byte(classificationDigest)
	digest[len(digest)-1] = suffix
	return TreeHash{Algorithm: TreeHashAlgorithmSHA256V2, Digest: string(digest)}
}

func classificationInventory(t *testing.T, root string, agents, claude []InventoryEntry, trusted bool, records []ProvenanceRecord) ProjectInventory {
	t.Helper()
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	return ProjectInventory{
		Root:      root,
		StatePath: layout.ReconcilerStatePath,
		State: ProvenanceState{
			Version: ProvenanceStateVersion,
			Records: records,
		},
		StateTrusted: trusted,
		Roots: []ProjectSkillsRootInventory{
			{Target: TargetAgents, Path: layout.AgentsSkillsPath, Exists: true, Safe: true, Entries: agents},
			{Target: TargetClaude, Path: layout.ClaudeSkillsPath, Exists: true, Safe: true, Entries: claude},
		},
	}
}

func classificationEntry(root string, target Target, name string, hash *TreeHash) InventoryEntry {
	entry := InventoryEntry{Target: target, Name: name, Path: filepath.Join(root, string(target), ManagedSkillsDirectoryName, name), Kind: InventoryEntryDirectory, Hash: hash}
	return entry
}

func classificationRecord(skill string, target Target, source string, hash TreeHash) ProvenanceRecord {
	return ProvenanceRecord{
		Scope:             ScopeProject,
		Skill:             skill,
		Target:            target,
		SourceIdentity:    source,
		TreeHashAlgorithm: hash.Algorithm,
		TreeHash:          hash.Digest,
	}
}

func classificationSkill(name string, targets ...Target) DesiredSkill {
	return DesiredSkill{
		Name: name, Source: "Owner/Repo", Scope: ScopeProject, Manager: ManagerSkillsCLI,
		Mode: ModeCopy, Targets: targets,
	}
}

func findProjectState(t *testing.T, classification ProjectClassification, target Target, skill string) ProjectState {
	t.Helper()
	for _, state := range classification.States {
		if state.Target == target && state.Skill == skill {
			return state
		}
	}
	t.Fatalf("missing state %s/%s in %#v", target, skill, classification.States)
	return ProjectState{}
}

func TestClassifyProjectCoversOwnershipMatrix(t *testing.T) {
	root := t.TempDir()
	oldHash := classificationHash('1')
	expectedHash := classificationHash('2')
	changedHash := classificationHash('3')
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{
		classificationSkill("exact", TargetAgents),
		classificationSkill("outdated", TargetAgents),
		classificationSkill("modified", TargetAgents),
		classificationSkill("missing", TargetAgents),
		classificationSkill("unmanaged", TargetAgents),
		classificationSkill("malformed", TargetAgents),
		classificationSkill("symlink", TargetAgents),
		classificationSkill("source-mismatch", TargetAgents),
		classificationSkill("other-target", TargetClaude),
	}}
	expected := map[string]TreeHash{}
	for _, skill := range desired.Skills {
		expected[skill.Name] = expectedHash
	}
	malformed := classificationEntry(root, TargetAgents, "malformed", nil)
	malformed.Problem = InventoryProblemEntryUnverifiable
	symlink := classificationEntry(root, TargetAgents, "symlink", nil)
	symlink.Kind = InventoryEntrySymlink
	symlink.Problem = InventoryProblemSymlinkNotHashed
	otherTarget := classificationEntry(root, TargetAgents, "other-target", &expectedHash)
	unknown := classificationEntry(root, TargetClaude, "unknown", &expectedHash)
	inventory := classificationInventory(t, root,
		[]InventoryEntry{
			classificationEntry(root, TargetAgents, "exact", &expectedHash),
			classificationEntry(root, TargetAgents, "outdated", &oldHash),
			classificationEntry(root, TargetAgents, "modified", &changedHash),
			classificationEntry(root, TargetAgents, "unmanaged", &expectedHash),
			malformed, symlink,
			classificationEntry(root, TargetAgents, "source-mismatch", &expectedHash),
			otherTarget,
		},
		[]InventoryEntry{unknown}, true, []ProvenanceRecord{
			classificationRecord("exact", TargetAgents, "github:owner/repo", expectedHash),
			classificationRecord("outdated", TargetAgents, "github:owner/repo", oldHash),
			classificationRecord("modified", TargetAgents, "github:owner/repo", oldHash),
			classificationRecord("source-mismatch", TargetAgents, "github:other/repo", expectedHash),
		})
	classification, err := ClassifyProject(desired, expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		target Target
		name   string
		kind   ProjectStateKind
		action PlanAction
		reason ProjectStateReason
	}{
		{TargetAgents, "exact", ProjectStateExact, PlanActionUnchanged, ProjectStateReasonVerifiedExact},
		{TargetAgents, "outdated", ProjectStateOutdated, PlanActionUpdate, ProjectStateReasonVerifiedUpdate},
		{TargetAgents, "modified", ProjectStateModified, PlanActionBlocked, ProjectStateReasonLocalModification},
		{TargetAgents, "missing", ProjectStateMissing, PlanActionInstall, ProjectStateReasonExpectedEntryAbsent},
		{TargetAgents, "unmanaged", ProjectStateUnmanaged, PlanActionBlocked, ProjectStateReasonDesiredPathUnmanaged},
		{TargetAgents, "malformed", ProjectStateMalformed, PlanActionBlocked, ProjectStateReasonCurrentEntryUnverifiable},
		{TargetAgents, "symlink", ProjectStateMisplaced, PlanActionBlocked, ProjectStateReasonExpectedCopyIsSymlink},
		{TargetAgents, "source-mismatch", ProjectStateProtected, PlanActionBlocked, ProjectStateReasonProvenanceSourceMismatch},
	}
	for _, test := range cases {
		state := findProjectState(t, classification, test.target, test.name)
		if state.Kind != test.kind || state.Action != test.action || state.Reason != test.reason {
			t.Errorf("%s = %#v, want %s/%s/%s", test.name, state, test.kind, test.action, test.reason)
		}
	}
	misplaced := findProjectState(t, classification, TargetAgents, "other-target")
	if misplaced.Kind != ProjectStateMisplaced || misplaced.Action != PlanActionUnchanged || misplaced.Reason != ProjectStateReasonUnmanagedMisplacedPreserved {
		t.Fatalf("misplaced collision = %#v", misplaced)
	}
	unmanaged := findProjectState(t, classification, TargetClaude, "unknown")
	if unmanaged.Kind != ProjectStateUnmanaged || unmanaged.Action != PlanActionUnchanged || unmanaged.Reason != ProjectStateReasonUnmanagedEntryPreserved {
		t.Fatalf("unmanaged extra = %#v", unmanaged)
	}
	if len(classification.States) != len(cases)+3 {
		t.Fatalf("states = %d, want %d", len(classification.States), len(cases)+3)
	}
}

func TestClassifyProjectRemovedManagedAndUntrustedState(t *testing.T) {
	root := t.TempDir()
	hash := classificationHash('1')
	changed := classificationHash('2')
	clean := classificationEntry(root, TargetAgents, "removed-clean", &hash)
	dirty := classificationEntry(root, TargetAgents, "removed-dirty", &changed)
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{classificationSkill("wanted", TargetAgents)}}
	expected := map[string]TreeHash{"wanted": hash}
	inventory := classificationInventory(t, root, []InventoryEntry{clean, dirty}, nil, true, []ProvenanceRecord{
		classificationRecord("removed-clean", TargetAgents, "github:owner/repo", hash),
		classificationRecord("removed-dirty", TargetAgents, "github:owner/repo", hash),
	})
	classification, err := ClassifyProject(desired, expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	cleanState := findProjectState(t, classification, TargetAgents, "removed-clean")
	if cleanState.Kind != ProjectStateOutdated || cleanState.Action != PlanActionQuarantine || cleanState.Reason != ProjectStateReasonPreviouslyManagedNotDesired {
		t.Fatalf("removed clean = %#v", cleanState)
	}
	dirtyState := findProjectState(t, classification, TargetAgents, "removed-dirty")
	if dirtyState.Kind != ProjectStateModified || dirtyState.Action != PlanActionBlocked || dirtyState.Reason != ProjectStateReasonPreviouslyManagedModified {
		t.Fatalf("removed dirty = %#v", dirtyState)
	}

	untrusted := classificationInventory(t, root, []InventoryEntry{classificationEntry(root, TargetAgents, "wanted", &hash)}, nil, false, nil)
	classification, err = ClassifyProject(desired, expected, untrusted)
	if err != nil {
		t.Fatal(err)
	}
	state := findProjectState(t, classification, "", "")
	if state.Kind != ProjectStateProtected || state.Action != PlanActionBlocked || state.Path != untrusted.StatePath || state.Reason != ProjectStateReasonProvenanceUntrusted {
		t.Fatalf("untrusted state marker = %#v", state)
	}
	wanted := findProjectState(t, classification, TargetAgents, "wanted")
	if wanted.Kind != ProjectStateUnmanaged || wanted.Action != PlanActionBlocked {
		t.Fatalf("untrusted bytes claimed ownership = %#v", wanted)
	}
}

func TestClassifyProjectManagerBoundaries(t *testing.T) {
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{
		{Name: "manual", Scope: ScopeProject, Manager: ManagerManual, Targets: []Target{TargetAgents}},
		{Name: "workflow", Scope: ScopeProject, Manager: ManagerWorkflow, Targets: []Target{TargetClaude}, Workflow: "external"},
	}}
	classification, err := ClassifyProject(desired, nil, classificationInventory(t, t.TempDir(), nil, nil, true, nil))
	if err != nil {
		t.Fatal(err)
	}
	manual := findProjectState(t, classification, TargetAgents, "manual")
	if manual.Kind != ProjectStateProtected || manual.Action != PlanActionManual || manual.Expected != nil {
		t.Fatalf("manual = %#v", manual)
	}
	workflow := findProjectState(t, classification, TargetClaude, "workflow")
	if workflow.Kind != ProjectStateProtected || workflow.Action != PlanActionWorkflow || workflow.Expected != nil {
		t.Fatalf("workflow = %#v", workflow)
	}
}

func TestCanonicalProjectSourceIdentityNormalizesSafeSources(t *testing.T) {
	cases := map[string]string{
		"Owner/Repo":        "github:owner/repo",
		"Owner/Repo/Skills": "github:owner/repo/skills",
		"https://github.com/OWNER/Repo/tree/main/skills":     "github:owner/repo",
		"https://github.com:443/OWNER/Repo/tree/main/skills": "github:owner/repo",
		"HTTPS://EXAMPLE.com/":                               "https://example.com",
		"https://example.com/skills.git/":                    "https://example.com/skills",
		"https://GIT.example.com:8443/skills.git":            "https://git.example.com:8443/skills",
		"https://github.com:8443/OWNER/Repo":                 "https://github.com:8443/OWNER/Repo",
	}
	for source, want := range cases {
		got, ok := canonicalProjectSourceIdentity(source)
		if !ok || got != want {
			t.Errorf("canonicalProjectSourceIdentity(%q) = %q, %v; want %q, true", source, got, ok, want)
		}
	}
	for _, source := range []string{
		"https://token@github.com/owner/repo",
		"https://github.com/owner/repo?token=secret",
		"http://github.com/owner/repo",
		"https://example.com:0/repo",
		"owner/../repo",
	} {
		if _, ok := canonicalProjectSourceIdentity(source); ok {
			t.Errorf("unsafe source %q was accepted", source)
		}
	}
}

func TestClassifyProjectValidatesInputsAndIsDeterministicDetached(t *testing.T) {
	root := t.TempDir()
	hash := classificationHash('1')
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{classificationSkill("alpha", TargetClaude, TargetAgents)}}
	inventory := classificationInventory(t, root, nil, nil, true, nil)
	if _, err := ClassifyProject(desired, map[string]TreeHash{"alpha": TreeHash{Algorithm: "bad", Digest: "UPPER"}}, inventory); err == nil {
		t.Fatal("invalid expected hash accepted")
	}
	badScope := desired
	badScope.Skills = []DesiredSkill{classificationSkill("alpha", TargetAgents)}
	badScope.Skills[0].Scope = Scope("")
	if _, err := ClassifyProject(badScope, map[string]TreeHash{"alpha": hash}, inventory); err == nil || !issueCode(err, IssueMalformedInput) {
		t.Fatalf("missing desired skill scope was accepted: %v", err)
	}
	manualExpected := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{{Name: "manual", Scope: ScopeProject, Manager: ManagerManual, Targets: []Target{TargetAgents}}}}
	if _, err := ClassifyProject(manualExpected, map[string]TreeHash{"manual": hash}, inventory); err == nil || !issueCode(err, IssueMissingReference) {
		t.Fatalf("manual expected hash was accepted: %v", err)
	}
	duplicateTarget := desired
	duplicateTarget.Skills = []DesiredSkill{classificationSkill("alpha", TargetAgents, TargetAgents)}
	if _, err := ClassifyProject(duplicateTarget, map[string]TreeHash{"alpha": hash}, inventory); err == nil {
		t.Fatal("duplicate desired target accepted")
	}
	duplicateRecord := classificationRecord("alpha", TargetAgents, "github:owner/repo", hash)
	duplicateInventory := classificationInventory(t, root, nil, nil, true, []ProvenanceRecord{duplicateRecord, duplicateRecord})
	if _, err := ClassifyProject(DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{classificationSkill("alpha", TargetAgents)}}, map[string]TreeHash{"alpha": hash}, duplicateInventory); err == nil {
		t.Fatal("duplicate provenance accepted")
	}
	outsidePathEntry := classificationEntry(root, TargetAgents, "alpha", &hash)
	outsidePathEntry.Path = filepath.Join(root, "outside", "alpha")
	outsidePathInventory := classificationInventory(t, root, []InventoryEntry{outsidePathEntry}, nil, true, []ProvenanceRecord{duplicateRecord})
	if _, err := ClassifyProject(DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{classificationSkill("alpha", TargetAgents)}}, map[string]TreeHash{"alpha": hash}, outsidePathInventory); err == nil || !issueCode(err, IssuePathEscape) {
		t.Fatalf("outside observed path was accepted: %v", err)
	}

	validDesired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{classificationSkill("alpha", TargetAgents), classificationSkill("beta", TargetClaude)}}
	validExpected := map[string]TreeHash{"alpha": hash, "beta": hash}
	validInventory := classificationInventory(t, root, []InventoryEntry{classificationEntry(root, TargetAgents, "alpha", &hash)}, []InventoryEntry{}, true, []ProvenanceRecord{classificationRecord("alpha", TargetAgents, "github:owner/repo", hash)})
	first, err := ClassifyProject(validDesired, validExpected, validInventory)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ClassifyProject(validDesired, validExpected, validInventory)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("classification JSON differs:\n%s\n%s", firstJSON, secondJSON)
	}
	if got := []Target{first.States[0].Target, first.States[1].Target}; !reflect.DeepEqual(got, []Target{TargetAgents, TargetClaude}) {
		t.Fatalf("target order = %#v", got)
	}
	firstState := findProjectState(t, first, TargetAgents, "alpha")
	firstState.Current.Digest = strings.Repeat("f", 64)
	if findProjectState(t, second, TargetAgents, "alpha").Current.Digest == strings.Repeat("f", 64) {
		t.Fatal("classification hash pointers are not detached")
	}
	if _, statErr := os.Stat(filepath.Join(root, string(TargetAgents), ManagedSkillsDirectoryName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("classifier changed fixture filesystem: %v", statErr)
	}
}
