package sjskills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

const testQuarantineID = "0123456789abcdef0123456789abcdef"

func TestApplyProjectChangesUpdatesThroughCommittedQuarantine(t *testing.T) {
	session, _, skill, oldHash := newApplyFixture(t, []Target{TargetAgents})
	oldExecutable := filepath.Join(session.Materialized.snapshots[skill.Name].Path, "run.sh")
	if err := os.WriteFile(oldExecutable, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
	applyFixture(t, session)
	oldHash = hashPlacedSkill(t, session, skill.Name, TargetAgents)

	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "run.sh"), []byte("#!/bin/sh\necho new\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	newHash := rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Operations) != 1 || session.Plan.Operations[0].Action != PlanActionUpdate {
		t.Fatalf("plan operations = %#v", session.Plan.Operations)
	}

	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Installed) != 0 || !reflect.DeepEqual(result.Updated, []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}}) {
		t.Fatalf("apply result = %#v", result)
	}
	if result.Quarantine == nil || result.Quarantine.ID != testQuarantineID || result.Quarantine.Status != ProjectQuarantineCommitted {
		t.Fatalf("quarantine result = %#v", result.Quarantine)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != newHash {
		t.Fatalf("new placement hash = %#v, want %#v", got, newHash)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineCommitted || len(manifest.Entries) != 1 || manifest.Entries[0].Status != ProjectQuarantineEntryReplaced || manifest.Entries[0].OldTreeHash != oldHash.Digest || manifest.Entries[0].NewTreeHash != newHash.Digest {
		t.Fatalf("manifest = %#v", manifest)
	}
	runPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID)
	if info, err := os.Stat(runPath); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("quarantine run mode = %v err=%v, want 0700", info, err)
	}
	if info, err := os.Stat(filepath.Join(runPath, applyManifestName)); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("quarantine manifest mode = %v err=%v, want 0600", info, err)
	}
	entry := manifest.Entries[0]
	if filepath.IsAbs(entry.OriginalPlacement) || filepath.IsAbs(entry.QuarantinedPlacement) ||
		entry.OriginalPlacement != ".agents/skills/"+skill.Name || entry.QuarantinedPlacement != "entries/.agents/"+skill.Name ||
		strings.Contains(entry.OriginalPlacement, session.Layout.Root) || strings.Contains(entry.QuarantinedPlacement, session.Layout.Root) {
		t.Fatalf("manifest placements are not path-free relative identities: %#v", entry)
	}
	oldPath := filepath.Join(runPath, filepath.FromSlash(entry.QuarantinedPlacement))
	if got := hashSkillAt(t, oldPath); got != oldHash {
		t.Fatalf("quarantined hash = %#v, want %#v", got, oldHash)
	}
	info, err := os.Stat(filepath.Join(oldPath, "run.sh"))
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("quarantined executable mode lost: info=%v err=%v", info, err)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHash != newHash.Digest {
		t.Fatalf("provenance = %#v", state)
	}

	makeSessionPlanCurrent(t, session)
	second, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err != nil {
		t.Fatal(err)
	}
	if second.Quarantine != nil || len(second.Installed) != 0 || len(second.Updated) != 0 {
		t.Fatalf("idempotent result = %#v", second)
	}
	entries, err := os.ReadDir(session.Layout.QuarantinePath)
	if err != nil || len(entries) != 1 {
		t.Fatalf("quarantine runs = %v err=%v", entries, err)
	}
}

func TestApplyProjectChangesRecordsManifestWhenQuarantineRootSyncFails(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	deps := updateTestDeps()
	injected := false
	rootSyncs := 0
	deps.SyncDir = func(path string) error {
		if path == session.Layout.QuarantinePath {
			rootSyncs++
			if !injected {
				injected = true
				return errors.New("injected quarantine root sync failure")
			}
		}
		return syncApplyDirectory(path)
	}

	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || !injected {
		t.Fatalf("root sync failure result=%#v err=%v injected=%v", result, err, injected)
	}
	if rootSyncs < 2 {
		t.Fatalf("quarantine root sync attempts = %d, want initial attempt plus durability retry", rootSyncs)
	}
	assertRolledBackPreparationEvidence(t, session, result, skill, oldHash)
}

func TestApplyProjectChangesRecordsManifestWhenQuarantineEntryDirectorySyncFails(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	runPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID)
	deps := updateTestDeps()
	injected := false
	deps.SyncDir = func(path string) error {
		if !injected && path == runPath {
			injected = true
			return errors.New("injected quarantine entry directory sync failure")
		}
		return syncApplyDirectory(path)
	}

	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || !injected {
		t.Fatalf("entry sync failure result=%#v err=%v injected=%v", result, err, injected)
	}
	assertRolledBackPreparationEvidence(t, session, result, skill, oldHash)
}

func TestApplyProjectChangesReportsCommittedUpdateOnFinalizationFailure(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	newHash := session.Expected[skill.Name]
	deps := updateTestDeps()
	deps.beforeUnlock = func() error { return errors.New("injected finalization failure") }
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("update unexpectedly reported clean finalization")
	}
	if !reflect.DeepEqual(result.Updated, []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}}) ||
		result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineCommitted {
		t.Fatalf("committed result = %#v", result)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != newHash {
		t.Fatalf("committed placement = %#v, want %#v", got, newHash)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHash != newHash.Digest {
		t.Fatalf("committed state = %#v", state)
	}
}

func TestApplyProjectChangesMixedInstallAndUpdatePreservesUnknown(t *testing.T) {
	session, _, prior, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	unknown := filepath.Join(session.Layout.AgentsSkillsPath, "unknown")
	if err := os.Mkdir(unknown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unknown, "keep"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[prior.Name].Path, "SKILL.md"), []byte("demo-new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rehashApplySnapshot(t, session, prior.Name)
	added := addApplyFixtureSkill(t, session, "added", []Target{TargetClaude})
	makeSessionPlanCurrent(t, session)

	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Installed, []AppliedPlacement{{Skill: added.Name, Target: TargetClaude}}) || !reflect.DeepEqual(result.Updated, []AppliedPlacement{{Skill: prior.Name, Target: TargetAgents}}) {
		t.Fatalf("mixed result = %#v", result)
	}
	if content, err := os.ReadFile(filepath.Join(unknown, "keep")); err != nil || string(content) != "keep\n" {
		t.Fatalf("unknown content = %q err=%v", content, err)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 2 {
		t.Fatalf("mixed provenance = %#v", state)
	}
}

func TestApplyProjectChangesRejectsUpdateRaceBeforeMove(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now:             fixedApplyTime,
		newQuarantineID: fixedQuarantineID,
		beforeQuarantine: func(AppliedPlacement) error {
			return os.WriteFile(filepath.Join(destination, "local"), []byte("changed\n"), 0o644)
		},
	})
	if err == nil {
		t.Fatal("update unexpectedly accepted raced content")
	}
	if _, err := os.Stat(filepath.Join(destination, "local")); err != nil {
		t.Fatalf("raced content was not preserved: %v", err)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("recovery result = %#v", result.Quarantine)
	}
}

func TestApplyProjectChangesRejectsUpdateSnapshotSourceDriftBeforeWrite(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	session.Materialized.snapshots[skill.Name].Skill.Source = "other/source"
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil {
		t.Fatal("update unexpectedly accepted snapshot source drift")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("placement changed = %#v, want %#v", got, oldHash)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("source drift created quarantine: err=%v", err)
	}
}

func TestApplyProjectChangesRefusesQuarantineDestinationCollision(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	collision := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, "entries", string(TargetAgents), skill.Name)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now:             fixedApplyTime,
		newQuarantineID: fixedQuarantineID,
		beforeQuarantine: func(AppliedPlacement) error {
			if err := os.Mkdir(collision, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(collision, "external"), []byte("keep\n"), 0o644)
		},
	})
	if err == nil {
		t.Fatal("update unexpectedly overwrote quarantine collision")
	}
	if content, err := os.ReadFile(filepath.Join(collision, "external")); err != nil || string(content) != "keep\n" {
		t.Fatalf("collision content = %q err=%v", content, err)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("original changed = %#v, want %#v", got, oldHash)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("recovery result = %#v", result.Quarantine)
	}
}

func TestApplyProjectChangesPreservesSwappedUpdatePublicationAndOldQuarantine(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	expected := session.Expected[skill.Name]
	var externalInfo os.FileInfo
	deps := updateTestDeps()
	deps.PublishNoReplace = func(source, target string) error {
		if target != destination {
			return publishNoReplace(source, target)
		}
		info, err := publishThenSwapWithIdenticalExternal(source, target)
		externalInfo = info
		return err
	}

	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("update unexpectedly owned a swapped publication")
	}
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) || applyErr.Conflict() || err.Error() != "project apply unavailable: rollback could not be verified" {
		t.Fatalf("swapped update error = %v, want stable recovery unavailable", err)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("swapped update recovery result = %#v", result.Quarantine)
	}
	currentInfo, statErr := os.Lstat(destination)
	if statErr != nil || externalInfo == nil || !os.SameFile(externalInfo, currentInfo) {
		t.Fatalf("external update destination was not preserved: info=%v err=%v", currentInfo, statErr)
	}
	if got := hashSkillAt(t, destination); got != expected {
		t.Fatalf("external update hash = %#v, want %#v", got, expected)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired || manifest.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("swapped update manifest = %#v", manifest)
	}
	oldPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, oldPath); got != oldHash {
		t.Fatalf("recoverable old hash = %#v, want %#v", got, oldHash)
	}
}

func TestApplyProjectChangesRestoresOriginalAfterUpdatePublishFailure(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	publishes := 0
	deps := updateTestDeps()
	deps.PublishNoReplace = func(source, destination string) error {
		publishes++
		if publishes == 3 {
			return errors.New("injected replacement publication failure")
		}
		return publishNoReplace(source, destination)
	}
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("update unexpectedly succeeded after publish failure")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restored hash = %#v, want %#v", got, oldHash)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("rollback result = %#v", result.Quarantine)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Entries[0].Status != ProjectQuarantineEntryRestored {
		t.Fatalf("rollback manifest = %#v", manifest)
	}
}

func TestApplyProjectChangesRestoresMultipleUpdatesAfterPartialFailure(t *testing.T) {
	session, _, _, oldHash := updateApplyFixture(t, []Target{TargetAgents, TargetClaude})
	publishes := 0
	deps := updateTestDeps()
	deps.PublishNoReplace = func(source, destination string) error {
		publishes++
		if publishes == 5 {
			return errors.New("injected second update publication failure")
		}
		return publishNoReplace(source, destination)
	}
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("multi-update unexpectedly succeeded")
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		if got := hashPlacedSkill(t, session, "demo", target); got != oldHash {
			t.Fatalf("restored %s hash = %#v, want %#v", target, got, oldHash)
		}
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("multi-update rollback = %#v", result.Quarantine)
	}
}

func TestApplyProjectChangesPreservesAmbiguousReplacementAndRecoveryEvidence(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	deps := updateTestDeps()
	deps.BeforePublish = func(AppliedPlacement) error {
		if err := os.Mkdir(destination, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644); err != nil {
			return err
		}
		return errors.New("injected ambiguous replacement")
	}
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("ambiguous update unexpectedly succeeded")
	}
	if content, err := os.ReadFile(filepath.Join(destination, "external")); err != nil || string(content) != "keep\n" {
		t.Fatalf("ambiguous content = %q err=%v", content, err)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("recovery result = %#v", result.Quarantine)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired || manifest.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("recovery manifest = %#v", manifest)
	}
	oldPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, oldPath); got != oldHash {
		t.Fatalf("recoverable old hash = %#v, want %#v", got, oldHash)
	}
}

func TestApplyProjectChangesRejectsUnsafeQuarantineAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Windows privileges")
	}
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	outside := t.TempDir()
	if err := os.Symlink(outside, session.Layout.QuarantinePath); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil {
		t.Fatal("update unexpectedly followed quarantine symlink")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("placement changed = %#v, want %#v", got, oldHash)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside quarantine entries = %v err=%v", entries, err)
	}
}

func TestApplyProjectChangesDoesNotCreateQuarantineForInstallOrNoOp(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	if result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err != nil || result.Quarantine != nil {
		t.Fatalf("install result=%#v err=%v", result, err)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("install created quarantine: err=%v", err)
	}
	makeSessionPlanCurrent(t, session)
	if result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err != nil || result.Quarantine != nil {
		t.Fatalf("no-op result=%#v err=%v", result, err)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("no-op created quarantine: err=%v", err)
	}
}

func TestProjectQuarantineManifestDecodingIsStrictAndBounded(t *testing.T) {
	manifest := ProjectQuarantineManifest{
		Version: ProjectQuarantineManifestVersion, ID: testQuarantineID,
		CreatedAt: fixedApplyTime(), Status: ProjectQuarantineCommitted,
		Entries: []ProjectQuarantineManifestEntry{{
			Skill: "demo", Target: TargetAgents,
			OriginalPlacement: projectOriginalPlacement(TargetAgents, "demo"), QuarantinedPlacement: projectQuarantinedPlacement(TargetAgents, "demo"),
			OldSourceIdentity: "github:example/demo", NewSourceIdentity: "github:example/demo",
			TreeHashAlgorithm: TreeHashAlgorithmSHA256V2,
			OldTreeHash:       strings.Repeat("a", 64), NewTreeHash: strings.Repeat("b", 64),
			Status: ProjectQuarantineEntryReplaced,
		}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := DecodeProjectQuarantineManifest(data); !ok {
		t.Fatal("valid quarantine manifest was rejected")
	}
	unknown := append(data[:len(data)-1], []byte(`,"unknown":true}`)...)
	if _, ok := DecodeProjectQuarantineManifest(unknown); ok {
		t.Fatal("unknown manifest field was accepted")
	}
	if _, ok := DecodeProjectQuarantineManifest(append(data, []byte(" true")...)); ok {
		t.Fatal("trailing manifest value was accepted")
	}
	if _, ok := DecodeProjectQuarantineManifest(bytes.Repeat([]byte("x"), maxProjectQuarantineManifestSize+1)); ok {
		t.Fatal("oversized manifest was accepted")
	}
	manifest.Entries[0].Status = "invented"
	invalid, _ := json.Marshal(manifest)
	if _, ok := DecodeProjectQuarantineManifest(invalid); ok {
		t.Fatal("invalid entry status was accepted")
	}
}

func updateApplyFixture(t *testing.T, targets []Target) (*ProjectApplySession, string, DesiredSkill, TreeHash) {
	t.Helper()
	session, project, skill, _ := newApplyFixture(t, targets)
	applyFixture(t, session)
	oldHash := hashPlacedSkill(t, session, skill.Name, targets[0])
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
	return session, project, skill, oldHash
}

func rehashApplySnapshot(t *testing.T, session *ProjectApplySession, skill string) TreeHash {
	t.Helper()
	snapshot := session.Materialized.snapshots[skill]
	hash, err := HashSkillTree(snapshot.Path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Hash = hash
	session.Expected[skill] = hash
	return hash
}

func updateTestDeps() ApplyDeps {
	return ApplyDeps{Now: fixedApplyTime, newQuarantineID: fixedQuarantineID}
}

func fixedApplyTime() time.Time {
	return time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)
}

func fixedQuarantineID() (string, error) { return testQuarantineID, nil }

func hashPlacedSkill(t *testing.T, session *ProjectApplySession, skill string, target Target) TreeHash {
	t.Helper()
	root, err := session.Layout.ManagedSkillsPath(target)
	if err != nil {
		t.Fatal(err)
	}
	return hashSkillAt(t, filepath.Join(root, skill))
}

func hashSkillAt(t *testing.T, path string) TreeHash {
	t.Helper()
	hash, err := HashSkillTree(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func readUpdateManifest(t *testing.T, session *ProjectApplySession) ProjectQuarantineManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName))
	if err != nil {
		t.Fatal(err)
	}
	manifest, ok := DecodeProjectQuarantineManifest(data)
	if !ok {
		t.Fatalf("manifest is invalid: %q", data)
	}
	return manifest
}

func assertRolledBackPreparationEvidence(t *testing.T, session *ProjectApplySession, result ApplyResult, skill DesiredSkill, oldHash TreeHash) {
	t.Helper()
	if result.Quarantine == nil || result.Quarantine.ID != testQuarantineID || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("preparation recovery result = %#v", result.Quarantine)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.ID != result.Quarantine.ID || manifest.Status != ProjectQuarantineRolledBack || len(manifest.Entries) != 1 ||
		manifest.Entries[0].Status != ProjectQuarantineEntryPending {
		t.Fatalf("preparation recovery manifest = %#v", manifest)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("preparation failure changed original = %#v, want %#v", got, oldHash)
	}
}

func readApplyState(t *testing.T, session *ProjectApplySession) ProvenanceState {
	t.Helper()
	data, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	state, ok := decodeProvenanceState(data)
	if !ok {
		t.Fatalf("state is invalid: %q", data)
	}
	return state
}
