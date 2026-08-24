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

func TestApplyProjectChangesQuarantinesRemovedPlacementWithoutMaterialization(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	unknown := filepath.Join(session.Layout.AgentsSkillsPath, "unknown")
	if err := os.MkdirAll(unknown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unknown, "keep"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeSessionPlanCurrent(t, session)
	if session.Materialized != nil || len(session.Plan.Operations) != 1 || session.Plan.Operations[0].Action != PlanActionQuarantine {
		t.Fatalf("removal session = %#v", session)
	}

	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Quarantined, []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}}) ||
		len(result.Installed) != 0 || len(result.Updated) != 0 {
		t.Fatalf("removal result = %#v", result)
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineCommitted {
		t.Fatalf("removal quarantine result = %#v", result.Quarantine)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !os.IsNotExist(err) {
		t.Fatalf("removed placement survived: %v", err)
	}
	if got := hashSkillAt(t, filepath.Join(session.Layout.QuarantinePath, testQuarantineID, "entries", string(TargetAgents), skill.Name)); got != oldHash {
		t.Fatalf("quarantined hash = %#v, want %#v", got, oldHash)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineCommitted || len(manifest.Entries) != 1 {
		t.Fatalf("removal manifest = %#v", manifest)
	}
	entry := manifest.Entries[0]
	if entry.Action != ProjectQuarantineEntryActionRemove || entry.Status != ProjectQuarantineEntryQuarantined ||
		entry.OldSourceIdentity != "github:example/demo" || entry.OldTreeHash != oldHash.Digest ||
		entry.NewSourceIdentity != "" || entry.NewTreeHash != "" {
		t.Fatalf("removal manifest entry = %#v", entry)
	}
	encoded, err := os.ReadFile(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "newSourceIdentity") || strings.Contains(string(encoded), "newTreeHash") {
		t.Fatalf("remove manifest serialized replacement fields: %s", encoded)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 0 {
		t.Fatalf("removed provenance survived: %#v", state)
	}
	if data, err := os.ReadFile(filepath.Join(unknown, "keep")); err != nil || string(data) != "keep\n" {
		t.Fatalf("unknown content = %q err=%v", data, err)
	}
}

func TestApplyProjectChangesRemovedPlacementIsIdempotent(t *testing.T) {
	session, _, skill, _ := removedApplyFixture(t, []Target{TargetAgents})
	first, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil || len(first.Quarantined) != 1 {
		t.Fatalf("first removal result = %#v err=%v", first, err)
	}
	entriesBefore, err := os.ReadDir(session.Layout.QuarantinePath)
	if err != nil {
		t.Fatal(err)
	}
	makeSessionPlanCurrent(t, session)
	second, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Quarantined) != 0 || second.Quarantine != nil || len(second.Plan.Operations) != 0 {
		t.Fatalf("idempotent removal result = %#v", second)
	}
	entriesAfter, err := os.ReadDir(session.Layout.QuarantinePath)
	if err != nil || len(entriesAfter) != len(entriesBefore) || skill.Name == "" {
		t.Fatalf("quarantine runs changed: before=%v after=%v err=%v", entriesBefore, entriesAfter, err)
	}
}

func TestApplyProjectChangesMixedInstallUpdateRemoveUsesOneDeterministicManifest(t *testing.T) {
	session, _, prior, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	removed := addApplyFixtureSkill(t, session, "removed", []Target{TargetAgents})
	session.Plan.Desired = session.Desired
	makeSessionPlanCurrent(t, session)
	applyFixture(t, session)
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[prior.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newHash := rehashApplySnapshot(t, session, prior.Name)
	added := addApplyFixtureSkill(t, session, "added", []Target{TargetAgents})
	// The updated desired set intentionally drops only the trusted removed
	// placement while retaining one install and one update in the same plan.
	session.Desired.Skills = []DesiredSkill{prior, added}
	session.Plan.Desired = session.Desired
	session.Expected = map[string]TreeHash{prior.Name: newHash, added.Name: session.Expected[added.Name]}
	makeSessionPlanCurrent(t, session)
	if got := []PlanAction{session.Plan.Operations[0].Action, session.Plan.Operations[1].Action, session.Plan.Operations[2].Action}; !reflect.DeepEqual(got, []PlanAction{PlanActionInstall, PlanActionUpdate, PlanActionQuarantine}) {
		t.Fatalf("mixed operations = %#v", session.Plan.Operations)
	}

	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Installed, []AppliedPlacement{{Skill: added.Name, Target: TargetAgents}}) ||
		!reflect.DeepEqual(result.Updated, []AppliedPlacement{{Skill: prior.Name, Target: TargetAgents}}) ||
		!reflect.DeepEqual(result.Quarantined, []AppliedPlacement{{Skill: removed.Name, Target: TargetAgents}}) {
		t.Fatalf("mixed result = %#v", result)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineCommitted || len(manifest.Entries) != 2 ||
		manifest.Entries[0].Skill != prior.Name || manifest.Entries[0].Action != ProjectQuarantineEntryActionUpdate || manifest.Entries[0].Status != ProjectQuarantineEntryReplaced ||
		manifest.Entries[1].Skill != removed.Name || manifest.Entries[1].Action != ProjectQuarantineEntryActionRemove || manifest.Entries[1].Status != ProjectQuarantineEntryQuarantined {
		t.Fatalf("mixed manifest = %#v", manifest)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 2 || state.Records[0].Skill != added.Name || state.Records[1].Skill != prior.Name {
		t.Fatalf("mixed provenance = %#v", state)
	}
}

func TestApplyProjectChangesRefusesModifiedRemovedPlacementBeforeWrites(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.WriteFile(filepath.Join(destination, "local"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Operations) != 1 || session.Plan.Operations[0].Action != PlanActionBlocked {
		t.Fatalf("modified removal plan = %#v", session.Plan.Operations)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("modified removal unexpectedly applied")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got == oldHash {
		t.Fatal("modified removal unexpectedly retained the old exact hash")
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("modified removal created quarantine: %v", err)
	}
}

func TestApplyProjectChangesRejectsTamperedRemovedExpectedBeforeWrites(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	session.Plan.Operations[0].Expected = PlanEvidence{Kind: projectEvidenceUnavailable, Detail: projectEvidenceUnavailable}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("tampered removal expected evidence unexpectedly applied")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("tampered removal changed original = %#v, want %#v", got, oldHash)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("tampered removal created quarantine: %v", err)
	}
}

func TestApplyProjectChangesRejectsRemovedProvenanceSourceDriftBeforeWrites(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	state := readApplyState(t, session)
	state.Records[0].SourceIdentity = "github:other/repo"
	data, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("source-drifted removal unexpectedly applied")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("source-drifted removal changed original = %#v, want %#v", got, oldHash)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("source-drifted removal created quarantine: %v", err)
	}
}

func TestApplyProjectChangesRejectsRemovedProvenanceHashDriftBeforeWrites(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	state := readApplyState(t, session)
	state.Records[0].TreeHash = strings.Repeat("f", 64)
	data, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("hash-drifted removal unexpectedly applied")
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("hash-drifted removal changed original = %#v, want %#v", got, oldHash)
	}
	if _, err := os.Lstat(session.Layout.QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("hash-drifted removal created quarantine: %v", err)
	}
}

func TestApplyProjectChangesRefusesRemovedQuarantineCollision(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	collision := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, "entries", string(TargetAgents), skill.Name)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now: fixedApplyTime, newQuarantineID: fixedQuarantineID,
		beforeQuarantine: func(AppliedPlacement) error {
			if err := os.MkdirAll(collision, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(collision, "external"), []byte("keep\n"), 0o644)
		},
	})
	if err == nil {
		t.Fatal("removal unexpectedly overwrote quarantine collision")
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("collision result = %#v", result)
	}
	if data, readErr := os.ReadFile(filepath.Join(collision, "external")); readErr != nil || string(data) != "keep\n" {
		t.Fatalf("collision content = %q err=%v", data, readErr)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("collision changed original = %#v, want %#v", got, oldHash)
	}
}

func TestApplyProjectChangesRefusesRacedRemovedQuarantineAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Windows privileges")
	}
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	outside := t.TempDir()
	parent := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, "entries", string(TargetAgents))
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now: fixedApplyTime, newQuarantineID: fixedQuarantineID,
		beforeQuarantine: func(AppliedPlacement) error {
			if err := os.Remove(parent); err != nil {
				return err
			}
			return os.Symlink(outside, parent)
		},
	})
	if err == nil {
		t.Fatal("removal unexpectedly followed raced quarantine ancestor")
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
		t.Fatalf("ancestor race result = %#v", result)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("ancestor race changed original = %#v, want %#v", got, oldHash)
	}
	if entries, readErr := os.ReadDir(outside); readErr != nil || len(entries) != 0 {
		t.Fatalf("outside ancestor entries = %v err=%v", entries, readErr)
	}
}

func TestApplyProjectChangesRollsBackRemovedPlacementsAndProvenanceOnCommitFailure(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents, TargetClaude})
	priorState := readApplyState(t, session)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now: fixedApplyTime, newQuarantineID: fixedQuarantineID,
		beforeCommit: func() error { return errors.New("injected state commit failure") },
	})
	if err == nil {
		t.Fatal("removal unexpectedly committed after state failure")
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack || len(result.Quarantined) != 0 {
		t.Fatalf("commit failure result = %#v", result)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		if got := hashPlacedSkill(t, session, skill.Name, target); got != oldHash {
			t.Fatalf("restored %s hash = %#v, want %#v", target, got, oldHash)
		}
	}
	if got := readApplyState(t, session); !reflect.DeepEqual(got, priorState) {
		t.Fatalf("provenance after rollback = %#v, want %#v", got, priorState)
	}
}

func TestApplyProjectChangesPreservesRacedRemovedReplacementAndQuarantine(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now: fixedApplyTime, newQuarantineID: fixedQuarantineID,
		beforeCommit: func() error {
			if err := os.Mkdir(destination, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644); err != nil {
				return err
			}
			return errors.New("injected commit failure after external replacement")
		},
	})
	if err == nil {
		t.Fatal("raced removal unexpectedly committed")
	}
	if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("raced removal result = %#v", result)
	}
	if data, readErr := os.ReadFile(filepath.Join(destination, "external")); readErr != nil || string(data) != "keep\n" {
		t.Fatalf("raced external content = %q err=%v", data, readErr)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired || manifest.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("raced removal manifest = %#v", manifest)
	}
	oldPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, oldPath); got != oldHash {
		t.Fatalf("raced quarantined hash = %#v, want %#v", got, oldHash)
	}
}

func TestProjectQuarantineManifestDecodingIsStrictAndBounded(t *testing.T) {
	manifest := ProjectQuarantineManifest{
		Version: ProjectQuarantineManifestVersion, ID: testQuarantineID,
		CreatedAt: fixedApplyTime(), Status: ProjectQuarantineCommitted,
		Entries: []ProjectQuarantineManifestEntry{{
			Action: ProjectQuarantineEntryActionUpdate, Skill: "demo", Target: TargetAgents,
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
	differentSource := manifest
	differentSource.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	differentSource.Entries[0] = manifest.Entries[0]
	differentSource.Entries[0].NewSourceIdentity = "github:other/repo"
	differentSourceData, err := json.Marshal(differentSource)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := DecodeProjectQuarantineManifest(differentSourceData); ok {
		t.Fatal("update manifest with differing canonical source identities was accepted")
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

	remove := manifest
	remove.Status = ProjectQuarantineCommitted
	remove.Entries = []ProjectQuarantineManifestEntry{{
		Action: ProjectQuarantineEntryActionRemove, Skill: "demo", Target: TargetAgents,
		OriginalPlacement: projectOriginalPlacement(TargetAgents, "demo"), QuarantinedPlacement: projectQuarantinedPlacement(TargetAgents, "demo"),
		OldSourceIdentity: "github:example/demo", TreeHashAlgorithm: TreeHashAlgorithmSHA256V2,
		OldTreeHash: strings.Repeat("a", 64), Status: ProjectQuarantineEntryQuarantined,
	}}
	removeData, err := json.Marshal(remove)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(removeData), "newSourceIdentity") || strings.Contains(string(removeData), "newTreeHash") {
		t.Fatalf("remove manifest emitted update-only fields: %s", removeData)
	}
	if _, ok := DecodeProjectQuarantineManifest(removeData); !ok {
		t.Fatal("valid remove quarantine manifest was rejected")
	}
	var removeObject map[string]any
	if err := json.Unmarshal(removeData, &removeObject); err != nil {
		t.Fatal(err)
	}
	removeEntries := removeObject["entries"].([]any)
	removeEntries[0].(map[string]any)["newTreeHash"] = ""
	explicitEmpty, err := json.Marshal(removeObject)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := DecodeProjectQuarantineManifest(explicitEmpty); ok {
		t.Fatal("remove manifest with explicit update-only field was accepted")
	}
	activeUpdate := manifest
	activeUpdate.Status = ProjectQuarantineActive
	activeUpdate.Entries[0].Status = ProjectQuarantineEntryRestored
	activeUpdateData, err := json.Marshal(activeUpdate)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := DecodeProjectQuarantineManifest(activeUpdateData); ok {
		t.Fatal("active update manifest with removal status was accepted")
	}
	activeRemove := remove
	activeRemove.Status = ProjectQuarantineActive
	activeRemove.Entries[0].Status = ProjectQuarantineEntryReplaced
	activeRemoveData, err := json.Marshal(activeRemove)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := DecodeProjectQuarantineManifest(activeRemoveData); ok {
		t.Fatal("active removal manifest with update status was accepted")
	}
	for name, mutate := range map[string]func(*ProjectQuarantineManifestEntry){
		"remove-new-source": func(entry *ProjectQuarantineManifestEntry) { entry.NewSourceIdentity = "github:other/repo" },
		"remove-new-hash":   func(entry *ProjectQuarantineManifestEntry) { entry.NewTreeHash = strings.Repeat("b", 64) },
		"update-missing-source": func(entry *ProjectQuarantineManifestEntry) {
			entry.Action = ProjectQuarantineEntryActionUpdate
			entry.NewTreeHash = strings.Repeat("b", 64)
		},
		"update-equal-hash": func(entry *ProjectQuarantineManifestEntry) {
			entry.Action = ProjectQuarantineEntryActionUpdate
			entry.NewSourceIdentity = entry.OldSourceIdentity
			entry.NewTreeHash = entry.OldTreeHash
			entry.Status = ProjectQuarantineEntryReplaced
		},
		"unknown-action": func(entry *ProjectQuarantineManifestEntry) { entry.Action = "unknown" },
	} {
		invalidRemove := remove
		invalidRemove.Entries = append([]ProjectQuarantineManifestEntry(nil), remove.Entries...)
		invalidRemove.Entries[0] = remove.Entries[0]
		mutate(&invalidRemove.Entries[0])
		data, marshalErr := json.Marshal(invalidRemove)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		if _, ok := DecodeProjectQuarantineManifest(data); ok {
			t.Fatalf("invalid %s manifest was accepted", name)
		}
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

func removedApplyFixture(t *testing.T, targets []Target) (*ProjectApplySession, string, DesiredSkill, TreeHash) {
	t.Helper()
	session, project, skill, _ := newApplyFixture(t, targets)
	applyFixture(t, session)
	oldHash := hashPlacedSkill(t, session, skill.Name, targets[0])
	session.Desired = DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{}}
	session.Expected = map[string]TreeHash{}
	session.Materialized = nil
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
