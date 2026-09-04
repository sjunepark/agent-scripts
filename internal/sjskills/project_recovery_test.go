package sjskills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestApplyProjectChangesRejectsSemanticallyUnrelatedJournalWithoutTouchingUnknownTree(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	preimage := &applyStatePreimage{state: emptyTrustedProvenanceState()}
	candidate, err := buildApplyState(preimage.state, session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, "")
	if err != nil {
		t.Fatal(err)
	}
	candidate.Records[0].Skill = "other"
	journal.CandidateState, err = marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	crafted, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	crafted = append(crafted, '\n')
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	unknownPath := filepath.Join(destination, "external")
	if err := os.WriteFile(unknownPath, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTransactionJournalPath(session.Layout), crafted, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err == nil || err.Error() != "project apply conflict: transaction recovery evidence is malformed" {
		t.Fatalf("crafted journal error=%v", err)
	}
	if data, readErr := os.ReadFile(unknownPath); readErr != nil || string(data) != "keep\n" {
		t.Fatalf("unknown tree changed: data=%q err=%v", data, readErr)
	}
}

func TestApplyProjectChangesRecoversJournalCommittedBeforeQuarantinePublication(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	preimage, err := captureApplyStatePreimage(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := buildApplyState(preimage.state, session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, testQuarantineID)
	if err != nil {
		t.Fatal(err)
	}
	journal.Entries[0].ManagedRootExisted = true
	journal.Entries[0].ManagedParentExisted = true
	journalData, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTransactionJournalPath(session.Layout), journalData, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("pre-quarantine recovery error=%v", err)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("pre-quarantine recovery changed placement=%#v want=%#v", got, oldHash)
	}
	if _, statErr := os.Lstat(filepath.Join(session.Layout.QuarantinePath, testQuarantineID)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("pre-quarantine recovery created a run: %v", statErr)
	}
}

func TestApplyRecoversAfterAbruptExitWithEmptyDurableQuarantineRun(t *testing.T) {
	const childEnv = "SJSKILLS_ABRUPT_QUARANTINE_CHILD"
	const rootEnv = "SJSKILLS_ABRUPT_QUARANTINE_ROOT"
	const stageEnv = "SJSKILLS_ABRUPT_QUARANTINE_STAGE"
	if mode := os.Getenv(childEnv); mode != "" {
		session, _, skill, _ := newApplyFixtureAt(t, os.Getenv(rootEnv), os.Getenv(stageEnv), []Target{TargetAgents})
		if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		rehashApplySnapshot(t, session, skill.Name)
		makeSessionPlanCurrent(t, session)
		deps := updateTestDeps()
		if mode == "run" {
			deps.afterQuarantineRunSync = func() { os.Exit(95) }
		} else {
			deps.afterInitialManifestSync = func() { os.Exit(97) }
		}
		_, _ = ApplyProjectChanges(context.Background(), session, deps)
		os.Exit(96)
	}
	for _, test := range []struct {
		mode string
		exit int
	}{{mode: "run", exit: 95}, {mode: "manifest", exit: 97}} {
		t.Run(test.mode, func(t *testing.T) {
			project := t.TempDir()
			stage := t.TempDir()
			session, _, skill, _ := newApplyFixtureAt(t, project, stage, []Target{TargetAgents})
			applyFixture(t, session)
			oldHash := hashPlacedSkill(t, session, skill.Name, TargetAgents)
			child := exec.Command(os.Args[0], "-test.run=^TestApplyRecoversAfterAbruptExitWithEmptyDurableQuarantineRun$")
			child.Env = append(os.Environ(), childEnv+"="+test.mode, rootEnv+"="+project, stageEnv+"="+stage)
			output, err := child.CombinedOutput()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != test.exit {
				t.Fatalf("abrupt quarantine child err=%v output=%q", err, output)
			}
			runPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID)
			if entries, readErr := os.ReadDir(runPath); readErr != nil || len(entries) != 0 {
				t.Fatalf("abrupt quarantine run entries=%v err=%v", entries, readErr)
			}
			rehashApplySnapshot(t, session, skill.Name)
			makeSessionPlanCurrent(t, session)
			if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
				t.Fatalf("abrupt quarantine recovery error=%v", err)
			}
			if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
				t.Fatalf("abrupt quarantine recovery changed placement=%#v want=%#v", got, oldHash)
			}
			if _, statErr := os.Lstat(runPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("empty interrupted quarantine run survived: %v", statErr)
			}
			stagedPath := filepath.Join(session.Layout.DerivedDirectoryPath, ".quarantine-manifest-"+testQuarantineID+"-staged")
			if _, statErr := os.Lstat(stagedPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("active interrupted manifest staging survived: %v", statErr)
			}
			if test.mode == "manifest" {
				recoveryPath := filepath.Join(projectRecoveryRunPath(session.Layout, testQuarantineID), "prepared-manifest.partial")
				if info, statErr := os.Lstat(recoveryPath); statErr != nil || !info.Mode().IsRegular() {
					t.Fatalf("interrupted manifest staging was not preserved: info=%v err=%v", info, statErr)
				}
			}
		})
	}
}

func TestApplyAbruptTemporaryPlacementRemainsPrivate(t *testing.T) {
	const childEnv = "SJSKILLS_ABRUPT_PRIVATE_STAGE_CHILD"
	const rootEnv = "SJSKILLS_ABRUPT_PRIVATE_STAGE_ROOT"
	const stageEnv = "SJSKILLS_ABRUPT_PRIVATE_STAGE_SOURCE"
	if os.Getenv(childEnv) == "1" {
		session, _, _, _ := newApplyFixtureAt(t, os.Getenv(rootEnv), os.Getenv(stageEnv), []Target{TargetAgents})
		deps := ApplyDeps{
			newQuarantineID: func() (string, error) { return testQuarantineID, nil },
			BeforePublish:   func(AppliedPlacement) error { os.Exit(99); return nil },
		}
		_, _ = ApplyProjectChanges(context.Background(), session, deps)
		os.Exit(98)
	}
	project := t.TempDir()
	stage := t.TempDir()
	child := exec.Command(os.Args[0], "-test.run=^TestApplyAbruptTemporaryPlacementRemainsPrivate$")
	child.Env = append(os.Environ(), childEnv+"=1", rootEnv+"="+project, stageEnv+"="+stage)
	output, err := child.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 99 {
		t.Fatalf("abrupt private-stage child err=%v output=%q", err, output)
	}
	session, _, skill, expected := newApplyFixtureAt(t, project, stage, []Target{TargetAgents})
	managedEntries, err := os.ReadDir(session.Layout.AgentsSkillsPath)
	if err != nil || len(managedEntries) != 0 {
		t.Fatalf("managed root contains interrupted staging: entries=%v err=%v", managedEntries, err)
	}
	privatePattern := filepath.Join(projectRecoveryRunPath(session.Layout, testQuarantineID), "staging", string(TargetAgents), applyInstallPattern+"*")
	privateStages, err := filepath.Glob(privatePattern)
	if err != nil || len(privateStages) != 1 {
		t.Fatalf("private interrupted staging=%v err=%v", privateStages, err)
	}
	if got := hashSkillAt(t, privateStages[0]); got != expected {
		t.Fatalf("private interrupted staging hash=%#v want=%#v", got, expected)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("private-stage recovery error=%v", err)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("private stage became a managed placement: %v", err)
	}
}

func TestApplyProjectChangesDoesNotRollbackAfterJournalUnlinkCommit(t *testing.T) {
	session, _, skill, expected := newApplyFixture(t, []Target{TargetAgents})
	armed := false
	deps := ApplyDeps{
		afterStateCommit: func() { armed = true },
		SyncDir: func(path string) error {
			if armed && filepath.Clean(path) == filepath.Clean(session.Layout.DerivedDirectoryPath) {
				if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); errors.Is(err, os.ErrNotExist) {
					return errors.New("injected journal directory sync failure")
				}
			}
			return syncApplyDirectory(path)
		},
	}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply unavailable: transaction recovery directory could not be synced" {
		t.Fatalf("journal unlink commit error=%v", err)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != expected {
		t.Fatalf("journal unlink failure rolled placement back=%#v want=%#v", got, expected)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("journal unlink commit left journal: %v", statErr)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHash != expected.Digest {
		t.Fatalf("journal unlink failure rolled provenance back=%#v", state)
	}
}

func TestApplyRollbackPreservesProvenanceReplacementRacedDuringRemoval(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	replaced := false
	deps := ApplyDeps{
		afterStateCommit: func() {
			file, err := os.OpenFile(projectTransactionJournalPath(session.Layout), os.O_APPEND|os.O_WRONLY, 0)
			if err != nil {
				panic(err)
			}
			if _, err := file.WriteString("\n"); err != nil {
				panic(err)
			}
			if err := file.Close(); err != nil {
				panic(err)
			}
		},
		beforeStateRecoveryMove: func() {
			replaced = true
			if err := os.Remove(session.Layout.ReconcilerStatePath); err != nil {
				panic(err)
			}
			if err := os.WriteFile(session.Layout.ReconcilerStatePath, []byte("external\n"), 0o600); err != nil {
				panic(err)
			}
		},
	}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply unavailable: provenance restoration could not be verified" {
		t.Fatalf("provenance removal race error=%v", err)
	}
	if !replaced {
		t.Fatal("provenance removal race hook did not run")
	}
	if data, readErr := os.ReadFile(session.Layout.ReconcilerStatePath); readErr != nil || string(data) != "external\n" {
		t.Fatalf("raced provenance was not preserved: data=%q err=%v", data, readErr)
	}
}

func TestApplyJournalFinalizationPreservesByteIdenticalReplacement(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	swapped := false
	deps := ApplyDeps{beforeJournalClear: func() {
		if swapped {
			return
		}
		swapped = true
		path := projectTransactionJournalPath(session.Layout)
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		if err := os.Remove(path); err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			panic(err)
		}
	}}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply conflict: transaction recovery evidence changed" {
		t.Fatalf("journal replacement error=%v", err)
	}
	if !swapped {
		t.Fatal("journal replacement hook did not run")
	}
	if info, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil || !info.Mode().IsRegular() {
		t.Fatalf("external journal replacement was not preserved: info=%v err=%v", info, statErr)
	}
	if _, statErr := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("journal replacement left managed placement: %v", statErr)
	}
}

func TestApplyJournalFinalizationPreservesInPlaceRewriteAfterMove(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	rewritten := false
	deps := ApplyDeps{
		newQuarantineID: func() (string, error) { return testQuarantineID, nil },
		afterJournalMove: func() {
			if rewritten {
				return
			}
			rewritten = true
			matches, err := filepath.Glob(filepath.Join(projectRecoveryRunPath(session.Layout, testQuarantineID), "transaction-apply-*.cleared"))
			if err != nil || len(matches) != 1 {
				panic("journal tombstone was not found")
			}
			file, err := os.OpenFile(matches[0], os.O_WRONLY|os.O_TRUNC, 0)
			if err != nil {
				panic(err)
			}
			if _, err := file.WriteString("external\n"); err != nil {
				panic(err)
			}
			if err := file.Close(); err != nil {
				panic(err)
			}
		},
	}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply conflict: transaction recovery evidence changed during finalization" {
		t.Fatalf("in-place journal rewrite error=%v", err)
	}
	if !rewritten {
		t.Fatal("in-place journal rewrite hook did not run")
	}
	if data, readErr := os.ReadFile(projectTransactionJournalPath(session.Layout)); readErr != nil || string(data) != "external\n" {
		t.Fatalf("in-place journal rewrite was not returned: data=%q err=%v", data, readErr)
	}
	if _, statErr := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("in-place journal rewrite left managed placement: %v", statErr)
	}
}

func TestInterruptedInstallRecoveryReturnsTreeSwappedBeforeMove(t *testing.T) {
	session, skill := simulateInterruptedInstallForRecovery(t, false)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	swapped := false
	deps := ApplyDeps{beforeRecoveryTreeMove: func() {
		if swapped {
			return
		}
		swapped = true
		if err := os.RemoveAll(destination); err != nil {
			panic(err)
		}
		if err := os.Mkdir(destination, 0o755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644); err != nil {
			panic(err)
		}
	}}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply conflict: interrupted recovery tree changed during move" {
		t.Fatalf("recovery tree swap error=%v", err)
	}
	if data, readErr := os.ReadFile(filepath.Join(destination, "external")); readErr != nil || string(data) != "keep\n" {
		t.Fatalf("recovery tree swap was not returned: data=%q err=%v", data, readErr)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("recovery tree swap journal disappeared: %v", statErr)
	}
}

func TestInterruptedInstallRecoveryRejectsByteIdenticalTreeSwap(t *testing.T) {
	session, skill := simulateInterruptedInstallForRecovery(t, false)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	swapped := false
	deps := ApplyDeps{beforeRecoveryTreeMove: func() {
		if swapped {
			return
		}
		swapped = true
		replacement := filepath.Join(session.Layout.Root, "identical-external")
		if err := os.Mkdir(replacement, 0o755); err != nil {
			panic(err)
		}
		if err := copyApplyTree(destination, replacement, func(*os.File) error { return nil }); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(destination); err != nil {
			panic(err)
		}
		if err := os.Rename(replacement, destination); err != nil {
			panic(err)
		}
	}}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply conflict: interrupted recovery tree changed during move" {
		t.Fatalf("byte-identical recovery swap error=%v", err)
	}
	if !swapped {
		t.Fatal("byte-identical recovery swap hook did not run")
	}
	if got := hashSkillAt(t, destination); got != session.Expected[skill.Name] {
		t.Fatalf("byte-identical external tree was not returned: got=%#v", got)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("byte-identical swap journal disappeared: %v", statErr)
	}
}

func TestInterruptedInstallRecoveryRejectsSwapAfterHashBeforeIdentityProof(t *testing.T) {
	session, skill := simulateInterruptedInstallForRecovery(t, false)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	swapped := false
	deps := ApplyDeps{afterRecoveryTreeHash: func() {
		if swapped {
			return
		}
		swapped = true
		replacement := filepath.Join(session.Layout.Root, "post-hash-identical-external")
		if err := os.Mkdir(replacement, 0o755); err != nil {
			panic(err)
		}
		if err := copyApplyTree(destination, replacement, func(*os.File) error { return nil }); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(destination); err != nil {
			panic(err)
		}
		if err := os.Rename(replacement, destination); err != nil {
			panic(err)
		}
	}}
	_, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil || err.Error() != "project apply conflict: interrupted recovery source changed before move" {
		t.Fatalf("post-hash recovery swap error=%v", err)
	}
	if !swapped {
		t.Fatal("post-hash recovery swap hook did not run")
	}
	if got := hashSkillAt(t, destination); got != session.Expected[skill.Name] {
		t.Fatalf("post-hash external tree changed: got=%#v", got)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("post-hash swap journal disappeared: %v", statErr)
	}
}

func TestInterruptedInstallRecoveryPreservesPreexistingManagedParent(t *testing.T) {
	session, skill := simulateInterruptedInstallForRecovery(t, true)
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("pre-existing parent recovery error=%v", err)
	}
	parent := filepath.Dir(session.Layout.AgentsSkillsPath)
	if info, err := os.Lstat(parent); err != nil || !info.IsDir() {
		t.Fatalf("pre-existing managed parent was removed: info=%v err=%v", info, err)
	}
	if _, err := os.Lstat(session.Layout.AgentsSkillsPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("transaction-created managed root survived: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted placement survived: %v", err)
	}
}

func TestRestoreDoesNotRollbackAfterJournalUnlinkCommit(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
	armed := false
	deps := ApplyDeps{
		afterStateCommit: func() { armed = true },
		SyncDir: func(path string) error {
			if armed && filepath.Clean(path) == filepath.Clean(session.Layout.DerivedDirectoryPath) {
				if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); errors.Is(err, os.ErrNotExist) {
					return errors.New("injected restore journal directory sync failure")
				}
			}
			return syncApplyDirectory(path)
		},
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, deps)
	if err == nil || err.Error() != "project restore unavailable: restore recovery evidence could not be finalized" {
		t.Fatalf("restore journal unlink result=%#v err=%v", result, err)
	}
	if result.Status != ProjectQuarantineRestored {
		t.Fatalf("restore journal unlink status=%q", result.Status)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restore journal unlink rolled placement back=%#v want=%#v", got, oldHash)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("restore journal unlink left journal: %v", statErr)
	}
}

func TestApplyProjectChangesRecoversInterruptedUpdateBeforeRerun(t *testing.T) {
	session, skill, oldHash, newHash := simulateInterruptedUpdate(t)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("interrupted update recovery error = %v", err)
	}
	if got := hashSkillAt(t, destination); got != oldHash {
		t.Fatalf("restored update hash=%#v, want %#v", got, oldHash)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRolledBack || len(manifest.Entries) != 1 || manifest.Entries[0].Status != ProjectQuarantineEntryRestored {
		t.Fatalf("recovered update manifest=%#v", manifest)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("transaction journal survived update recovery: %v", err)
	}
	makeSessionPlanCurrent(t, session)
	deps := updateTestDeps()
	deps.newQuarantineID = func() (string, error) { return "abcdef0123456789abcdef0123456789", nil }
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err != nil || len(result.Updated) != 1 || result.Updated[0].Skill != skill.Name {
		t.Fatalf("post-recovery update result=%#v err=%v", result, err)
	}
	if got := hashSkillAt(t, destination); got != newHash {
		t.Fatalf("post-recovery update hash=%#v, want %#v", got, newHash)
	}
}

func TestProjectTransactionJournalDecodingIsStrictAndBounded(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	preimage := &applyStatePreimage{state: emptyTrustedProvenanceState()}
	candidate, err := buildApplyState(preimage.state, session, time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if decoded, valid := decodeProjectTransactionJournal(data); !valid || decoded.ID != testQuarantineID {
		t.Fatalf("valid journal decoded=%#v valid=%v", decoded, valid)
	}
	secondRecord := candidate.Records[0]
	secondRecord.Skill = "other"
	candidate.Records = append(candidate.Records, secondRecord)
	twoSkillState, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	twoSkillJournal := journal
	twoSkillJournal.CandidateState = twoSkillState
	twoSkillJournal.Entries = append([]projectJournalEntry(nil), journal.Entries...)
	twoSkillJournal.Entries[0].ManagedParentExisted = true
	twoSkillJournal.Entries[0].ManagedRootExisted = true
	secondEntry := twoSkillJournal.Entries[0]
	secondEntry.Skill = "other"
	twoSkillJournal.Entries = append(twoSkillJournal.Entries, secondEntry)
	if !validProjectTransactionJournal(twoSkillJournal) {
		t.Fatal("consistent two-skill ancestor evidence was rejected")
	}
	twoSkillJournal.Entries[1].ManagedParentExisted = false
	twoSkillJournal.Entries[1].ManagedRootExisted = false
	if validProjectTransactionJournal(twoSkillJournal) {
		t.Fatal("inconsistent per-target ancestor evidence was accepted")
	}
	unknown := append([]byte(nil), data[:len(data)-2]...)
	unknown = append(unknown, []byte(",\n  \"unknown\": true\n}\n")...)
	badHash := bytes.Replace(data, []byte(session.Expected["demo"].Digest), []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), 1)
	for name, candidate := range map[string][]byte{
		"unknown-field": unknown,
		"trailing-json": append(append([]byte(nil), data...), []byte("{}\n")...),
		"bad-hash":      badHash,
		"oversized":     bytes.Repeat([]byte("x"), maxProjectTransactionJournalSize+1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, valid := decodeProjectTransactionJournal(candidate); valid {
				t.Fatal("invalid transaction journal was accepted")
			}
		})
	}
}

func TestApplyProjectChangesPreservesAmbiguousInterruptedUpdate(t *testing.T) {
	session, skill, oldHash, _ := simulateInterruptedUpdate(t)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.RemoveAll(destination); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) || !applyErr.Conflict() || err.Error() != "project apply conflict: interrupted managed placement is ambiguous" {
		t.Fatalf("ambiguous interrupted update error=%v", err)
	}
	if result.Quarantine == nil || result.Quarantine.ID != testQuarantineID || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("ambiguous recovery handle=%#v", result.Quarantine)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "external")); err != nil || string(data) != "keep\n" {
		t.Fatalf("external replacement changed: data=%q err=%v", data, err)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
	if got := hashSkillAt(t, quarantined); got != oldHash {
		t.Fatalf("ambiguous quarantined hash=%#v, want %#v", got, oldHash)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); err != nil {
		t.Fatalf("ambiguous recovery journal disappeared: %v", err)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("ambiguous manifest status=%q", manifest.Status)
	}
}

func TestApplyProvenanceAmbiguityReportsQuarantineHandleBeforeMutation(t *testing.T) {
	session, skill, oldHash, newHash := simulateInterruptedUpdate(t)
	state := readApplyState(t, session)
	state.Records[0].RecordedAt = state.Records[0].RecordedAt.Add(time.Second)
	thirdState, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, thirdState, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err == nil || err.Error() != "project apply conflict: interrupted provenance state is ambiguous" {
		t.Fatalf("provenance ambiguity error=%v", err)
	}
	if result.Quarantine == nil || result.Quarantine.ID != testQuarantineID || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("provenance ambiguity handle=%#v", result.Quarantine)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != newHash {
		t.Fatalf("provenance ambiguity mutated placement=%#v want=%#v", got, newHash)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
	if got := hashSkillAt(t, quarantined); got != oldHash {
		t.Fatalf("provenance ambiguity mutated quarantine=%#v want=%#v", got, oldHash)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("provenance ambiguity removed journal: %v", statErr)
	}
}

func TestApplyRecoveryRejectsPreStatePermissionDrift(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows access control uses ACLs, not Unix permission bits")
	}
	session, skill, _, newHash := simulateInterruptedUpdate(t)
	journalData, err := os.ReadFile(projectTransactionJournalPath(session.Layout))
	if err != nil {
		t.Fatal(err)
	}
	journal, valid := decodeProjectTransactionJournal(journalData)
	if !valid || !journal.PreState.Exists {
		t.Fatal("interrupted journal pre-state is invalid")
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, journal.PreState.Data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(session.Layout.ReconcilerStatePath, 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err == nil || err.Error() != "project apply conflict: interrupted provenance state is ambiguous" {
		t.Fatalf("pre-state permission drift error=%v", err)
	}
	if result.Quarantine == nil || result.Quarantine.ID != testQuarantineID || result.Quarantine.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("pre-state permission drift handle=%#v", result.Quarantine)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != newHash {
		t.Fatalf("pre-state permission drift mutated placement=%#v want=%#v", got, newHash)
	}
	if info, statErr := os.Lstat(session.Layout.ReconcilerStatePath); statErr != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("pre-state permission drift was silently restored: info=%v err=%v", info, statErr)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("pre-state permission drift removed journal: %v", statErr)
	}
}

func TestRestoreProjectQuarantineRecoversInterruptedRestoreBeforeRerun(t *testing.T) {
	session, skill, oldHash := simulateInterruptedRestore(t)
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
	if err == nil || err.Error() != "project restore unavailable: interrupted project transaction was recovered; rerun restore" {
		t.Fatalf("interrupted restore result=%#v err=%v", result, err)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted restored destination survived recovery: %v", err)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
	if got := hashSkillAt(t, quarantined); got != oldHash {
		t.Fatalf("recovered restore quarantine hash=%#v, want %#v", got, oldHash)
	}
	manifest := readUpdateManifest(t, session)
	if manifest.Status != ProjectQuarantineCommitted || manifest.Entries[0].Status != ProjectQuarantineEntryReplaced {
		t.Fatalf("recovered restore manifest=%#v", manifest)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("restore recovery journal survived: %v", err)
	}
	result, err = RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
	if err != nil || result.Status != ProjectQuarantineRestored || len(result.Restored) != 1 {
		t.Fatalf("post-recovery restore result=%#v err=%v", result, err)
	}
	if got := hashSkillAt(t, destination); got != oldHash {
		t.Fatalf("post-recovery restored hash=%#v, want %#v", got, oldHash)
	}
}

func TestRestoreProvenanceAmbiguityReportsQuarantineHandleBeforeMutation(t *testing.T) {
	session, skill, oldHash := simulateInterruptedRestore(t)
	state := readApplyState(t, session)
	state.Records[0].RecordedAt = state.Records[0].RecordedAt.Add(time.Second)
	thirdState, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, thirdState, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
	if err == nil || err.Error() != "project restore conflict: interrupted project transaction is ambiguous" {
		t.Fatalf("restore provenance ambiguity result=%#v err=%v", result, err)
	}
	if result.ID != testQuarantineID || result.Status != ProjectQuarantineRecoveryRequired {
		t.Fatalf("restore provenance ambiguity handle=%#v", result)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if got := hashSkillAt(t, destination); got != oldHash {
		t.Fatalf("restore provenance ambiguity mutated placement=%#v want=%#v", got, oldHash)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("restore provenance ambiguity removed journal: %v", statErr)
	}
}

func TestRestoreRecoveryRejectsMissingCommittedRunBeforeMovingPlacement(t *testing.T) {
	session, skill, oldHash := simulateInterruptedRestore(t)
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.RemoveAll(filepath.Join(session.Layout.QuarantinePath, testQuarantineID)); err != nil {
		t.Fatal(err)
	}
	_, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
	if err == nil || err.Error() != "project restore conflict: interrupted project transaction is ambiguous" {
		t.Fatalf("missing restore run error=%v", err)
	}
	if got := hashSkillAt(t, destination); got != oldHash {
		t.Fatalf("missing restore run moved placement=%#v want=%#v", got, oldHash)
	}
	if _, statErr := os.Lstat(projectTransactionJournalPath(session.Layout)); statErr != nil {
		t.Fatalf("missing restore run journal disappeared: %v", statErr)
	}
}

func TestRestoreProjectQuarantineRecoversAfterAbruptProcessExit(t *testing.T) {
	const childEnv = "SJSKILLS_ABRUPT_RESTORE_CHILD"
	const rootEnv = "SJSKILLS_ABRUPT_RESTORE_ROOT"
	if os.Getenv(childEnv) == "1" {
		layout, err := LayoutForProject(os.Getenv(rootEnv))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = RestoreProjectQuarantine(context.Background(), layout, testQuarantineID, ApplyDeps{afterStateCommit: func() { os.Exit(93) }})
		os.Exit(94)
	}
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.RemoveAll(destination); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestRestoreProjectQuarantineRecoversAfterAbruptProcessExit$")
	child.Env = append(os.Environ(), childEnv+"=1", rootEnv+"="+session.Layout.Root)
	output, err := child.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 93 {
		t.Fatalf("abrupt restore child err=%v output=%q", err, output)
	}
	if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{}); err == nil || err.Error() != "project restore unavailable: interrupted project transaction was recovered; rerun restore" {
		t.Fatalf("abrupt restore recovery error=%v", err)
	}
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("abrupt restored destination survived recovery: %v", err)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
	if got := hashSkillAt(t, quarantined); got != oldHash {
		t.Fatalf("abrupt recovered quarantine hash=%#v, want %#v", got, oldHash)
	}
	result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
	if err != nil || result.Status != ProjectQuarantineRestored {
		t.Fatalf("post-abrupt restore result=%#v err=%v", result, err)
	}
}

func simulateInterruptedUpdate(t *testing.T) (*ProjectApplySession, DesiredSkill, TreeHash, TreeHash) {
	t.Helper()
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	newHash := session.Expected[skill.Name]
	preimage, err := captureApplyStatePreimage(session.Layout)
	if err != nil || !preimage.exists {
		t.Fatalf("capture preimage=%#v err=%v", preimage, err)
	}
	recordedAt := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)
	candidate, err := buildApplyState(preimage.state, session, recordedAt)
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, testQuarantineID)
	if err != nil {
		t.Fatal(err)
	}
	journal.Entries[0].ManagedParentExisted = true
	journal.Entries[0].ManagedRootExisted = true
	journalData, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
	if err := os.MkdirAll(filepath.Dir(quarantined), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(destination, quarantined); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyApplyTree(session.Materialized.snapshots[skill.Name].Path, destination, func(file *os.File) error { return file.Sync() }); err != nil {
		t.Fatal(err)
	}
	record, ok := candidateRecordByKey(preimage.state, projectPlacementKey(TargetAgents, skill.Name))
	if !ok {
		t.Fatal("old provenance record missing")
	}
	manifest := ProjectQuarantineManifest{
		Version: ProjectQuarantineManifestVersion, ID: testQuarantineID, CreatedAt: recordedAt,
		Status: ProjectQuarantineActive,
		Entries: []ProjectQuarantineManifestEntry{{
			Action: ProjectQuarantineEntryActionUpdate, Skill: skill.Name, Target: TargetAgents,
			OriginalPlacement:    projectOriginalPlacement(TargetAgents, skill.Name),
			QuarantinedPlacement: projectQuarantinedPlacement(TargetAgents, skill.Name),
			OldSourceIdentity:    record.SourceIdentity, NewSourceIdentity: record.SourceIdentity,
			TreeHashAlgorithm: oldHash.Algorithm, OldTreeHash: oldHash.Digest, NewTreeHash: newHash.Digest,
			Status: ProjectQuarantineEntryReplaced,
		}},
	}
	manifestData, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName), manifestData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTransactionJournalPath(session.Layout), journalData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, candidateData, 0o600); err != nil {
		t.Fatal(err)
	}
	return session, skill, oldHash, newHash
}

func simulateInterruptedInstallForRecovery(t *testing.T, parentExisted bool) (*ProjectApplySession, DesiredSkill) {
	t.Helper()
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	if parentExisted {
		if err := os.MkdirAll(filepath.Dir(session.Layout.AgentsSkillsPath), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	preimage := &applyStatePreimage{state: emptyTrustedProvenanceState()}
	candidate, err := buildApplyState(preimage.state, session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, "")
	if err != nil {
		t.Fatal(err)
	}
	journal.Entries[0].ManagedParentExisted = parentExisted
	journalData, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTransactionJournalPath(session.Layout), journalData, 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyApplyTree(session.Materialized.snapshots[skill.Name].Path, destination, func(file *os.File) error { return file.Sync() }); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, candidateData, 0o600); err != nil {
		t.Fatal(err)
	}
	return session, skill
}

func simulateInterruptedRestore(t *testing.T) (*ProjectApplySession, DesiredSkill, TreeHash) {
	t.Helper()
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName)
	preManifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, valid := DecodeProjectQuarantineManifest(preManifestData)
	if !valid {
		t.Fatal("committed restore manifest is invalid")
	}
	preimage, err := captureApplyStatePreimage(session.Layout)
	if err != nil || !preimage.exists {
		t.Fatalf("capture restore preimage=%#v err=%v", preimage, err)
	}
	entry := manifest.Entries[0]
	restoreEntries := []restoreEntry{{
		entry: entry, oldHash: TreeHash{Algorithm: entry.TreeHashAlgorithm, Digest: entry.OldTreeHash},
	}}
	candidate, err := buildRestoreProvenanceState(preimage.state, restoreEntries, time.Date(2026, time.March, 4, 5, 6, 7, 0, time.UTC), ScopeProject)
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal := projectTransactionJournal{
		Version: projectTransactionJournalVersion, ID: testQuarantineID, Kind: projectTransactionKindRestore,
		QuarantineID: testQuarantineID, CandidateState: candidateData, PreManifest: preManifestData,
		PreState: projectJournalState{Exists: true, Mode: uint32(preimage.mode.Perm()), Data: preimage.data},
		Entries: []projectJournalEntry{{
			Action: PlanActionUpdate, Skill: skill.Name, Target: TargetAgents,
			ManagedParentExisted: true,
			ManagedRootExisted:   true,
			OldSourceIdentity:    entry.OldSourceIdentity, NewSourceIdentity: entry.NewSourceIdentity,
			TreeHashAlgorithm: entry.TreeHashAlgorithm, OldTreeHash: entry.OldTreeHash, NewTreeHash: entry.NewTreeHash,
		}},
	}
	journalData, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.RemoveAll(destination); err != nil {
		t.Fatal(err)
	}
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(entry.QuarantinedPlacement))
	if err := os.Rename(quarantined, destination); err != nil {
		t.Fatal(err)
	}
	interruptedManifest := manifest
	interruptedManifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	interruptedManifest.Status = ProjectQuarantineRestoring
	interruptedManifest.Entries[0].Status = ProjectQuarantineEntryRestored
	interruptedManifestData, err := marshalProjectQuarantineManifest(interruptedManifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, interruptedManifestData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTransactionJournalPath(session.Layout), journalData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, candidateData, 0o600); err != nil {
		t.Fatal(err)
	}
	return session, skill, oldHash
}
