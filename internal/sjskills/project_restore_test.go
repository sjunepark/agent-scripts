package sjskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRestoreProjectQuarantineRestoresCommittedUpdate(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	newHash := session.Expected[skill.Name]
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != testQuarantineID || result.Status != ProjectQuarantineRestored ||
		!reflect.DeepEqual(result.Restored, []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}}) {
		t.Fatalf("restore result = %#v", result)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restored hash = %#v, want %#v", got, oldHash)
	}
	if got := readApplyState(t, session); len(got.Records) != 1 || got.Records[0].TreeHash != oldHash.Digest || got.Records[0].TreeHash == newHash.Digest {
		t.Fatalf("restored provenance = %#v", got)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRestored || manifest.Entries[0].Status != ProjectQuarantineEntryRestored {
		t.Fatalf("restored manifest = %#v", manifest)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))); !os.IsNotExist(err) {
		t.Fatalf("quarantine source survived restore: %v", err)
	}
}

func TestRestoreProjectQuarantineRestoresCommittedRemoval(t *testing.T) {
	session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Restored, []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}}) || result.Status != ProjectQuarantineRestored {
		t.Fatalf("restore result = %#v", result)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restored hash = %#v, want %#v", got, oldHash)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].SourceIdentity != "github:example/demo" || state.Records[0].TreeHash != oldHash.Digest {
		t.Fatalf("restored removal provenance = %#v", state)
	}
}

func TestRestoreProjectQuarantineIsIdempotent(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	first, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err != nil || len(first.Restored) != 1 {
		t.Fatalf("first restore = %#v err=%v", first, err)
	}
	second, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != ProjectQuarantineRestored || len(second.Restored) != 0 || second.ID != testQuarantineID || skill.Name == "" {
		t.Fatalf("idempotent restore = %#v", second)
	}
}

func TestRestoreProjectQuarantineRestoresMixedUpdateAndRemoval(t *testing.T) {
	session, updated, removed, updatedOldHash, removedOldHash := mixedRestoreApplyFixture(t)
	removeRestoreDestinations(t, session, updated, []Target{TargetAgents})
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	want := []AppliedPlacement{
		{Skill: updated.Name, Target: TargetAgents},
		{Skill: removed.Name, Target: TargetAgents},
	}
	if result.ID != testQuarantineID || result.Status != ProjectQuarantineRestored || !reflect.DeepEqual(result.Restored, want) {
		t.Fatalf("mixed restore result = %#v, want %#v", result, want)
	}
	if got := hashPlacedSkill(t, session, updated.Name, TargetAgents); got != updatedOldHash {
		t.Fatalf("restored update hash = %#v, want %#v", got, updatedOldHash)
	}
	if got := hashPlacedSkill(t, session, removed.Name, TargetAgents); got != removedOldHash {
		t.Fatalf("restored removal hash = %#v, want %#v", got, removedOldHash)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRestored || len(manifest.Entries) != 2 {
		t.Fatalf("mixed restored manifest = %#v", manifest)
	}
	for _, entry := range manifest.Entries {
		if entry.Status != ProjectQuarantineEntryRestored {
			t.Fatalf("mixed restored entry = %#v", entry)
		}
	}
	state := readApplyState(t, session)
	if len(state.Records) != 2 || state.Records[0].TreeHash != updatedOldHash.Digest || state.Records[1].TreeHash != removedOldHash.Digest {
		t.Fatalf("mixed restored provenance = %#v", state)
	}
}

func TestRestoreProjectQuarantineRejectsInvalidIDsWithoutPathAccess(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	ids := []string{"", strings.ToUpper(testQuarantineID), strings.Repeat("a", 31), strings.Repeat("a", 33), "../" + testQuarantineID}
	for _, id := range ids {
		result, err := RestoreProjectQuarantine(context.Background(), session.Layout, id, updateTestDeps())
		if err == nil {
			t.Fatalf("restore unexpectedly accepted invalid id %q", id)
		}
		if result.ID != "" || result.Status != "" || len(result.Restored) != 0 {
			t.Fatalf("invalid id %q returned result %#v", id, result)
		}
		if strings.Contains(err.Error(), session.Layout.Root) {
			t.Fatalf("invalid id %q leaked a path: %v", id, err)
		}
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("invalid id changed manifest = %#v", got)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); err != nil {
		t.Fatalf("invalid id changed destination: %v", err)
	}
}

func TestRestoreProjectQuarantinePreflightsAllEntriesBeforeMoving(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents, TargetClaude})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	collision := filepath.Join(session.Layout.ClaudeSkillsPath, skill.Name)
	if err := os.MkdirAll(collision, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(collision, "keep"), []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err == nil {
		t.Fatal("restore unexpectedly succeeded with one destination collision")
	}
	var restoreErr *RestoreError
	if !errors.As(err, &restoreErr) || !restoreErr.Conflict() {
		t.Fatalf("restore error = %v", err)
	}
	if result.Status != ProjectQuarantineCommitted || len(result.Restored) != 0 {
		t.Fatalf("collision result = %#v", result)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("manifest changed during preflight = %#v", got)
	}
	got, readErr := os.ReadFile(filepath.Join(collision, "keep"))
	if readErr != nil || string(got) != "external\n" {
		t.Fatalf("collision content changed = %q err=%v", got, readErr)
	}
}

func TestRestoreProjectQuarantineRejectsCommittedQuarantineTamper(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	manifest := readUpdateManifest(t, session)
	quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if err := os.WriteFile(filepath.Join(quarantine, "tampered"), []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err == nil {
		t.Fatal("tampered quarantine unexpectedly restored")
	}
	if result.Status != ProjectQuarantineCommitted || len(result.Restored) != 0 || strings.Contains(err.Error(), session.Layout.Root) {
		t.Fatalf("tampered quarantine result=%#v err=%v", result, err)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("tampered quarantine changed manifest = %#v", got)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); err != nil {
		t.Fatalf("tampered quarantine changed destination: %v", err)
	}
}

func TestRestoreProjectQuarantineRejectsProvenanceDrift(t *testing.T) {
	t.Run("update-source", func(t *testing.T) {
		session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
		if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
			t.Fatal(err)
		}
		removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
		state := readApplyState(t, session)
		state.Records[0].SourceIdentity = "github:other/repo"
		writeRestoreState(t, session, state)
		result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
		assertRestorePreflightRefused(t, session, result, err)
	})

	t.Run("update-hash", func(t *testing.T) {
		session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
		if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
			t.Fatal(err)
		}
		removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
		state := readApplyState(t, session)
		state.Records[0].TreeHash = strings.Repeat("f", 64)
		writeRestoreState(t, session, state)
		result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
		assertRestorePreflightRefused(t, session, result, err)
	})

	t.Run("removal-unexpected-record", func(t *testing.T) {
		session, _, skill, oldHash := removedApplyFixture(t, []Target{TargetAgents})
		if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
			t.Fatal(err)
		}
		state := ProvenanceState{Version: ProvenanceStateVersion, Records: []ProvenanceRecord{{
			Scope: ScopeProject, Skill: skill.Name, Target: TargetAgents,
			SourceIdentity: "github:example/demo", TreeHashAlgorithm: oldHash.Algorithm,
			TreeHash: oldHash.Digest, RecordedAt: fixedApplyTime(),
		}}}
		writeRestoreState(t, session, state)
		result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
		assertRestorePreflightRefused(t, session, result, err)
	})
}

func TestRestoreProjectQuarantineRejectsDestinationAppearanceRace(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	managed, err := session.Layout.ManagedSkillsPath(TargetAgents)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(managed, skill.Name)
	deps := updateTestDeps()
	deps.beforeRestoreMove = func(AppliedPlacement) error {
		if err := os.Mkdir(destination, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644)
	}
	result, restoreErr := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if restoreErr == nil || result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("destination race result=%#v err=%v", result, restoreErr)
	}
	if content, readErr := os.ReadFile(filepath.Join(destination, "external")); readErr != nil || string(content) != "keep\n" {
		t.Fatalf("destination race content=%q err=%v", content, readErr)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired || manifest.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("destination race manifest=%#v", manifest)
	}
	quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, quarantine); got != oldHash {
		t.Fatalf("destination race quarantine hash=%#v want=%#v", got, oldHash)
	}
}

func TestRestoreProjectQuarantineRollsBackAfterMoveFailure(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents, TargetClaude})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents, TargetClaude})
	moves := 0
	deps := updateTestDeps()
	deps.PublishNoReplace = func(source, destination string) error {
		moves++
		if moves == 2 {
			return errors.New("injected restore move failure")
		}
		return publishNoReplace(source, destination)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after injected move failure")
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("rollback result = %#v", result)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, rootErr := session.Layout.ManagedSkillsPath(target)
		if rootErr != nil {
			t.Fatal(rootErr)
		}
		if _, statErr := os.Lstat(filepath.Join(root, skill.Name)); !os.IsNotExist(statErr) {
			t.Fatalf("destination %s after rollback = %v, want absent result=%#v restoreErr=%v", target, statErr, result, err)
		}
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("manifest after rollback = %#v", got)
	}
}

func TestRestoreProjectQuarantineRollsBackAfterManifestProgressFailure(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	manifestWrites := 0
	deps := updateTestDeps()
	deps.beforeRestoreManifest = func(status ProjectQuarantineStatus) error {
		if status == ProjectQuarantineRestoring {
			manifestWrites++
			if manifestWrites == 2 {
				return errors.New("injected restore progress failure")
			}
		}
		return nil
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after manifest progress failure")
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("progress rollback result = %#v", result)
	}
	assertRestoreRollbackFixture(t, session, skill, oldHash)
}

func TestRestoreProjectQuarantineRollsBackAfterSyncFailure(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	managedRoot, err := session.Layout.ManagedSkillsPath(TargetAgents)
	if err != nil {
		t.Fatal(err)
	}
	failed := false
	deps := updateTestDeps()
	deps.SyncDir = func(path string) error {
		if !failed && path == managedRoot {
			failed = true
			return errors.New("injected restore sync failure")
		}
		return syncApplyDirectory(path)
	}
	result, restoreErr := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if restoreErr == nil || !failed {
		t.Fatalf("restore unexpectedly succeeded after sync failure result=%#v err=%v", result, restoreErr)
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("sync rollback result = %#v", result)
	}
	assertRestoreRollbackFixture(t, session, skill, oldHash)
}

func TestRestoreProjectQuarantineRollsBackAfterProvenanceCommitFailure(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	deps := updateTestDeps()
	deps.beforeRestoreCommit = func() error { return errors.New("injected restore provenance failure") }
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after provenance failure")
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("provenance rollback result = %#v", result)
	}
	assertRestoreRollbackFixture(t, session, skill, oldHash)
}

func TestRestoreProjectQuarantineRollsBackCreatedManagedAncestors(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	if err := os.RemoveAll(filepath.Join(session.Layout.Root, string(TargetAgents))); err != nil {
		t.Fatal(err)
	}
	deps := updateTestDeps()
	deps.beforeRestoreCommit = func() error { return errors.New("injected restore provenance failure") }
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after provenance failure")
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("created ancestor rollback result = %#v", result)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.Root, string(TargetAgents))); !os.IsNotExist(err) {
		t.Fatalf("created managed ancestor survived rollback: %v", err)
	}
	quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(readUpdateManifest(t, session).Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, quarantine); got != oldHash {
		t.Fatalf("quarantined hash after ancestor rollback = %#v, want %#v", got, oldHash)
	}
}

func TestRestoreProjectQuarantineCleansPartiallyCreatedManagedAncestors(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	managedAncestor := filepath.Join(session.Layout.Root, string(TargetAgents))
	if err := os.RemoveAll(managedAncestor); err != nil {
		t.Fatal(err)
	}
	failed := false
	deps := updateTestDeps()
	deps.SyncDir = func(path string) error {
		if path == session.Layout.Root && !failed {
			failed = true
			return errors.New("injected partial ancestor sync failure")
		}
		return syncApplyDirectory(path)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil || !failed || result.Status != ProjectQuarantineCommitted {
		t.Fatalf("partial ancestor result=%#v err=%v failed=%v", result, err, failed)
	}
	if _, statErr := os.Lstat(managedAncestor); !os.IsNotExist(statErr) {
		t.Fatalf("partial ancestor survived rollback: %v", statErr)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("partial ancestor manifest=%#v", got)
	}
}

func TestRestoreProjectQuarantineRollsBackAfterFinalManifestFailure(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	deps := updateTestDeps()
	deps.beforeRestoreManifest = func(status ProjectQuarantineStatus) error {
		if status == ProjectQuarantineRestored {
			return errors.New("injected final manifest failure")
		}
		return nil
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after final manifest failure")
	}
	if result.Status != ProjectQuarantineCommitted {
		t.Fatalf("final manifest rollback result = %#v", result)
	}
	assertRestoreRollbackFixture(t, session, skill, oldHash)
}

func TestRestoreProjectQuarantinePreservesExternalRollbackReplacement(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents, TargetClaude})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents, TargetClaude})
	moves := 0
	deps := updateTestDeps()
	deps.PublishNoReplace = func(source, destination string) error {
		moves++
		if moves == 2 {
			return errors.New("injected restore move failure")
		}
		return publishNoReplace(source, destination)
	}
	deps.beforeRestoreRollback = func(placement AppliedPlacement) error {
		if placement.Target != TargetAgents {
			return nil
		}
		managed, managedErr := session.Layout.ManagedSkillsPath(TargetAgents)
		if managedErr != nil {
			return managedErr
		}
		manifest := readUpdateManifest(t, session)
		quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
		destination := filepath.Join(managed, skill.Name)
		if err := os.Rename(destination, quarantine); err != nil {
			return err
		}
		if err := os.Mkdir(destination, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil {
		t.Fatal("restore unexpectedly succeeded after external rollback replacement")
	}
	if result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("external replacement result = %#v err=%v", result, err)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired || manifest.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("external replacement manifest = %#v", manifest)
	}
	managed, err := session.Layout.ManagedSkillsPath(TargetAgents)
	if err != nil {
		t.Fatal(err)
	}
	if content, readErr := os.ReadFile(filepath.Join(managed, skill.Name, "external")); readErr != nil || string(content) != "keep\n" {
		t.Fatalf("external replacement content = %q err=%v", content, readErr)
	}
	quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, quarantine); got != oldHash {
		t.Fatalf("preserved quarantine hash = %#v, want %#v", got, oldHash)
	}
}

func TestRestoreProjectQuarantineRejectsUntouchedQuarantineSwapDuringRollback(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	manifest := readUpdateManifest(t, session)
	quarantine := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	backup := quarantine + ".original"
	deps := updateTestDeps()
	deps.beforeRestoreMove = func(AppliedPlacement) error {
		if err := os.Rename(quarantine, backup); err != nil {
			return err
		}
		if err := os.Mkdir(quarantine, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(quarantine, "external"), []byte("keep\n"), 0o644)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil || result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("untouched swap result=%#v err=%v", result, err)
	}
	if strings.Contains(err.Error(), session.Layout.Root) || result.ID != testQuarantineID {
		t.Fatalf("untouched swap leaked path or handle: result=%#v err=%v", result, err)
	}
	recovered := readUpdateManifest(t, session)
	if recovered.Status != ProjectQuarantineRecoveryRequired || recovered.Entries[0].Status != ProjectQuarantineEntryRecoveryRequired {
		t.Fatalf("untouched swap manifest=%#v", recovered)
	}
	if content, readErr := os.ReadFile(filepath.Join(quarantine, "external")); readErr != nil || string(content) != "keep\n" {
		t.Fatalf("untouched external content=%q err=%v", content, readErr)
	}
	if got := hashSkillAt(t, backup); got != oldHash {
		t.Fatalf("untouched original hash=%#v want=%#v", got, oldHash)
	}
}

func TestRestoreProjectQuarantineRestoredIdempotenceFailsClosedOnTamper(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	managed, err := session.Layout.ManagedSkillsPath(TargetAgents)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managed, skill.Name, "tampered"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, restoreErr := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if restoreErr == nil || result.Status != ProjectQuarantineRestored || len(result.Restored) != 0 {
		t.Fatalf("restored tamper result=%#v err=%v", result, restoreErr)
	}
	if strings.Contains(restoreErr.Error(), session.Layout.Root) {
		t.Fatalf("restored tamper leaked path: %v", restoreErr)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineRestored {
		t.Fatalf("restored tamper changed manifest=%#v", got)
	}
}

func TestRestoreProjectQuarantineRestoredIdempotenceFailsClosedOnProvenanceMismatch(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	state := readApplyState(t, session)
	state.Records[0].TreeHash = strings.Repeat("f", 64)
	writeRestoreState(t, session, state)
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err == nil || result.Status != ProjectQuarantineRestored || len(result.Restored) != 0 {
		t.Fatalf("restored provenance mismatch result=%#v err=%v", result, err)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineRestored {
		t.Fatalf("restored provenance mismatch changed manifest=%#v", got)
	}
}

func TestRestoreProjectQuarantineRejectsManifestSwapRace(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	manifestPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName)
	backup := manifestPath + ".original"
	deps := updateTestDeps()
	deps.beforeRestoreManifest = func(status ProjectQuarantineStatus) error {
		if status != ProjectQuarantineRestoring {
			return nil
		}
		if err := os.Rename(manifestPath, backup); err != nil {
			return err
		}
		return os.WriteFile(manifestPath, []byte("external manifest\n"), 0o600)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil || result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("manifest swap result=%#v err=%v", result, err)
	}
	if strings.Contains(err.Error(), session.Layout.Root) || result.ID != testQuarantineID {
		t.Fatalf("manifest swap leaked path or handle: result=%#v err=%v", result, err)
	}
	if data, readErr := os.ReadFile(manifestPath); readErr != nil || string(data) != "external manifest\n" {
		t.Fatalf("manifest swap external content=%q err=%v", data, readErr)
	}
}

func TestRestoreProjectQuarantineRejectsBoundarySwapRace(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	manifest := readUpdateManifest(t, session)
	parent := filepath.Dir(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement)))
	backup := parent + ".original"
	deps := updateTestDeps()
	deps.beforeRestoreMove = func(AppliedPlacement) error {
		if err := os.Rename(parent, backup); err != nil {
			return err
		}
		return os.Mkdir(parent, 0o755)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil || result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("boundary swap result=%#v err=%v", result, err)
	}
	if strings.Contains(err.Error(), session.Layout.Root) || result.ID != testQuarantineID {
		t.Fatalf("boundary swap leaked path or handle: result=%#v err=%v", result, err)
	}
	if _, statErr := os.Lstat(backup); statErr != nil {
		t.Fatalf("boundary swap lost original parent: %v", statErr)
	}
}

func TestRestoreProjectQuarantineHonorsCancellation(t *testing.T) {
	session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := RestoreProjectQuarantine(ctx, session.Layout, testQuarantineID, updateTestDeps())
	if err == nil || result.ID != "" || result.Status != "" || len(result.Restored) != 0 || strings.Contains(err.Error(), session.Layout.Root) {
		t.Fatalf("cancelled restore result=%#v err=%v", result, err)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("cancelled restore changed manifest=%#v", got)
	}
}

func TestRestoreProjectQuarantinePreservesExecutableBits(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	executable := filepath.Join(session.Materialized.snapshots[skill.Name].Path, "run.sh")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldHash := rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
	applyFixture(t, session)
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "run.sh"))
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("restored executable mode = %v err=%v", info, err)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restored executable tree hash=%#v want=%#v", got, oldHash)
	}
}

func TestRestoreProjectQuarantineRejectsRestoringManifest(t *testing.T) {
	session, _, _, _ := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	manifest := readUpdateManifest(t, session)
	manifest.Status = ProjectQuarantineRestoring
	data, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
	if err == nil {
		t.Fatal("restoring manifest unexpectedly accepted")
	}
	if !strings.Contains(err.Error(), "different recovery workflow") || strings.Contains(err.Error(), session.Layout.Root) {
		t.Fatalf("restoring error = %v", err)
	}
}

func TestRestoreProjectQuarantineRejectsOtherDurableStatuses(t *testing.T) {
	for name, status := range map[string]ProjectQuarantineStatus{
		"prepared":          ProjectQuarantinePrepared,
		"active":            ProjectQuarantineActive,
		"rolled-back":       ProjectQuarantineRolledBack,
		"recovery-required": ProjectQuarantineRecoveryRequired,
	} {
		t.Run(name, func(t *testing.T) {
			session, _, skill, _ := updateApplyFixture(t, []Target{TargetAgents})
			if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
				t.Fatal(err)
			}
			manifest := readUpdateManifest(t, session)
			manifest.Status = status
			switch status {
			case ProjectQuarantinePrepared, ProjectQuarantineRolledBack:
				manifest.Entries[0].Status = ProjectQuarantineEntryPending
			case ProjectQuarantineActive:
				manifest.Entries[0].Status = ProjectQuarantineEntryQuarantined
			case ProjectQuarantineRecoveryRequired:
				manifest.Entries[0].Status = ProjectQuarantineEntryRecoveryRequired
			}
			data, err := marshalProjectQuarantineManifest(manifest)
			if err != nil {
				t.Fatal(err)
			}
			manifestPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName)
			if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			result, restoreErr := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps())
			if restoreErr == nil {
				t.Fatal("restore unexpectedly accepted an in-progress manifest")
			}
			if result.ID != "" || result.Status != "" || len(result.Restored) != 0 {
				t.Fatalf("rejected status result = %#v", result)
			}
			after, err := os.ReadFile(manifestPath)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("manifest changed after rejected status: before=%q after=%q err=%v", before, after, err)
			}
			root, err := session.Layout.ManagedSkillsPath(TargetAgents)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Join(root, skill.Name)); err != nil {
				t.Fatalf("destination changed after rejected status: %v", err)
			}
		})
	}
}

func mixedRestoreApplyFixture(t *testing.T) (*ProjectApplySession, DesiredSkill, DesiredSkill, TreeHash, TreeHash) {
	t.Helper()
	session, _, updated, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	removed := addApplyFixtureSkill(t, session, "removed", []Target{TargetAgents})
	makeSessionPlanCurrent(t, session)
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removedOldHash := hashPlacedSkill(t, session, removed.Name, TargetAgents)
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[updated.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updatedOldHash := hashPlacedSkill(t, session, updated.Name, TargetAgents)
	newHash := rehashApplySnapshot(t, session, updated.Name)
	session.Desired = DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{updated}}
	session.Plan.Desired = session.Desired
	session.Expected = map[string]TreeHash{updated.Name: newHash}
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Operations) != 2 || session.Plan.Operations[0].Action != PlanActionUpdate || session.Plan.Operations[1].Action != PlanActionQuarantine {
		t.Fatalf("mixed restore operations = %#v", session.Plan.Operations)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	return session, updated, removed, updatedOldHash, removedOldHash
}

func writeRestoreState(t *testing.T, session *ProjectApplySession, state ProvenanceState) {
	t.Helper()
	data, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertRestorePreflightRefused(t *testing.T, session *ProjectApplySession, result RestoreResult, err error) {
	t.Helper()
	if err == nil || result.Status != ProjectQuarantineCommitted || len(result.Restored) != 0 {
		t.Fatalf("restore unexpectedly crossed preflight: result=%#v err=%v", result, err)
	}
	if strings.Contains(err.Error(), session.Layout.Root) {
		t.Fatalf("restore error leaked path: %v", err)
	}
	if got := readUpdateManifest(t, session); got.Status != ProjectQuarantineCommitted {
		t.Fatalf("preflight refusal changed manifest=%#v", got)
	}
}

func removeRestoreDestinations(t *testing.T, session *ProjectApplySession, skill DesiredSkill, targets []Target) {
	t.Helper()
	for _, target := range targets {
		root, err := session.Layout.ManagedSkillsPath(target)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(filepath.Join(root, skill.Name)); err != nil {
			t.Fatal(err)
		}
	}
}

func assertRestoreRollbackFixture(t *testing.T, session *ProjectApplySession, skill DesiredSkill, oldHash TreeHash) {
	t.Helper()
	root, err := session.Layout.ManagedSkillsPath(TargetAgents)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Lstat(filepath.Join(root, skill.Name)); !os.IsNotExist(statErr) {
		t.Fatalf("destination after rollback = %v, want absent", statErr)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineCommitted || manifest.Entries[0].Status != ProjectQuarantineEntryReplaced {
		t.Fatalf("manifest after rollback = %#v", manifest)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(manifest.Entries[0].QuarantinedPlacement))
	if got := hashSkillAt(t, quarantined); got != oldHash {
		t.Fatalf("quarantined hash after rollback = %#v, want %#v", got, oldHash)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHash == oldHash.Digest {
		t.Fatalf("state after rollback = %#v", state)
	}
}
