package sjskills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestApplyProjectChangesRejectsBlockedPlanBeforeLock(t *testing.T) {
	root := t.TempDir()
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	desired := DesiredState{
		Scope: ScopeProject,
		Skills: []DesiredSkill{{
			Name: "blocked", Source: "example/blocked", SourceID: "blocked",
			Scope: ScopeProject, Origin: "test", Manager: ManagerSkillsCLI,
			Mode: ModeCopy, Targets: []Target{TargetAgents},
		}},
	}
	session := &ProjectApplySession{
		Layout:       layout,
		Desired:      desired,
		Plan:         Plan{Desired: desired, Operations: []PlanOperation{{Action: PlanActionBlocked, Skill: "blocked", Target: TargetAgents, Manager: ManagerSkillsCLI}}},
		Expected:     map[string]TreeHash{"blocked": {Algorithm: TreeHashAlgorithmSHA256V2, Digest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
		Materialized: &MaterializationPlan{snapshots: map[string]*SkillSnapshot{}},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("ApplyProjectChanges unexpectedly accepted blocked plan")
	} else {
		var applyErr *ApplyError
		if !errors.As(err, &applyErr) || !applyErr.Conflict() {
			t.Fatalf("ApplyProjectChanges() error = %v, want conflict", err)
		}
	}
	if _, err := os.Lstat(layout.DerivedDirectoryPath); !os.IsNotExist(err) {
		t.Fatalf("static rejection created derived state: err=%v", err)
	}
	if _, err := os.Lstat(filepath.Join(layout.AgentsSkillsPath, "blocked")); !os.IsNotExist(err) {
		t.Fatalf("static rejection created placement: err=%v", err)
	}
}

func TestApplyErrorConflictClassification(t *testing.T) {
	err := applyConflict("stable reason")
	if !err.(*ApplyError).Conflict() {
		t.Fatal("apply conflict did not report conflict")
	}
	if got := err.Error(); got != "project apply conflict: stable reason" {
		t.Fatalf("error = %q", got)
	}
}

func TestApplyProjectChangesCopiesOnePlacementAndWritesSortedProvenance(t *testing.T) {
	project := t.TempDir()
	stage := t.TempDir()
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutForProject(canonicalProject)
	if err != nil {
		t.Fatal(err)
	}
	skill := DesiredSkill{
		Name: "demo", Source: "example/demo", SourceID: "demo-source",
		Scope: ScopeProject, Origin: "test", Manager: ManagerSkillsCLI,
		Mode: ModeCopy, Targets: []Target{TargetAgents},
	}
	snapshotPath := filepath.Join(stage, ".agents", "skills", skill.Name)
	if err := os.MkdirAll(snapshotPath, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(snapshotPath, "run.sh")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshotPath, "SKILL.md"), []byte("demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	materialized := &MaterializationPlan{
		root:      stage,
		snapshots: map[string]*SkillSnapshot{},
	}
	snapshot := &SkillSnapshot{Skill: skill, Path: snapshotPath, Hash: hash, plan: materialized, stageRoot: stage}
	materialized.snapshots[skill.Name] = snapshot
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{skill}}
	session := &ProjectApplySession{
		Layout: layout, Desired: desired,
		Plan: Plan{Desired: desired, Operations: []PlanOperation{{
			Action: PlanActionInstall, Skill: skill.Name, Target: TargetAgents,
			Manager: ManagerSkillsCLI, SourceID: skill.SourceID, Source: skill.Source,
			Reason:   string(ProjectStateReasonExpectedEntryAbsent),
			Current:  PlanEvidence{Kind: projectEvidenceAbsent, Detail: projectEvidenceAbsent},
			Expected: PlanEvidence{Kind: projectEvidenceTreeHash, Detail: hash.Algorithm + ":" + hash.Digest},
		}}},
		Expected: map[string]TreeHash{skill.Name: hash}, Materialized: materialized,
	}
	recordedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt }})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Installed) != 1 || result.Installed[0] != (AppliedPlacement{Skill: skill.Name, Target: TargetAgents}) {
		t.Fatalf("installed = %#v", result.Installed)
	}
	placed := filepath.Join(layout.AgentsSkillsPath, skill.Name)
	placedHash, err := HashSkillTree(placed)
	if err != nil || placedHash != hash {
		t.Fatalf("placed hash = %#v, err=%v; want %#v", placedHash, err, hash)
	}
	placedInfo, err := os.Stat(filepath.Join(placed, "run.sh"))
	if err != nil || (runtime.GOOS != "windows" && placedInfo.Mode().Perm()&0o111 == 0) {
		t.Fatalf("executable mode not preserved: info=%v err=%v", placedInfo, err)
	}
	stateData, err := os.ReadFile(layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	var state ProvenanceState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Records) != 1 || state.Records[0].RecordedAt != recordedAt || state.Records[0].TreeHash != hash.Digest {
		t.Fatalf("state = %#v", state)
	}
	if string(stateData[len(stateData)-1]) != "\n" {
		t.Fatalf("state is not newline terminated: %q", stateData)
	}
}

func TestApplyProjectChangesReportsCommittedPlacementsOnFinalizationFailure(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents})
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{
		Now: func() time.Time { return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC) },
		beforeUnlock: func() error {
			return errors.New("injected finalization failure")
		},
	})
	if err == nil {
		t.Fatal("ApplyProjectChanges unexpectedly succeeded after finalization failure")
	}
	if len(result.Installed) != 1 || result.Installed[0] != (AppliedPlacement{Skill: skill.Name, Target: TargetAgents}) {
		t.Fatalf("installed = %#v, want retained committed placement", result.Installed)
	}
	placedHash, hashErr := HashSkillTree(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name))
	if hashErr != nil || placedHash != hash {
		t.Fatalf("retained placement hash=%#v err=%v, want %#v", placedHash, hashErr, hash)
	}
	inventory, inspectErr := InspectProject(session.Layout)
	if inspectErr != nil || !inventory.StateTrusted || len(inventory.State.Records) != 1 {
		t.Fatalf("committed provenance inventory=%#v err=%v", inventory, inspectErr)
	}
}

func TestApplyProjectChangesRollsBackEarlierPlacementsOnPublishFailure(t *testing.T) {
	session, project, skill, hash := newApplyFixture(t, []Target{TargetAgents, TargetClaude})
	publishCalls := 0
	deps := ApplyDeps{
		Now: func() time.Time { return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC) },
		PublishNoReplace: func(source, destination string) error {
			publishCalls++
			if publishCalls == 2 {
				return errors.New("injected publication failure")
			}
			return publishNoReplace(source, destination)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectChanges unexpectedly succeeded after injected publication failure")
	}
	if publishCalls != 3 {
		t.Fatalf("publish calls = %d, want two publications plus rollback move", publishCalls)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, err := session.Layout.ManagedSkillsPath(target)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatalf("target root %s survived rollback: err=%v", target, err)
		}
	}
	if _, err := os.Stat(session.Layout.ReconcilerStatePath); !os.IsNotExist(err) {
		t.Fatalf("provenance state survived failed transaction: err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName)); !os.IsNotExist(err) {
		t.Fatalf("apply lock survived failed transaction: err=%v", err)
	}
	if hash == (TreeHash{}) || skill.Name == "" || project == "" {
		t.Fatal("fixture unexpectedly empty")
	}
}

func TestApplyProjectChangesPreservesNoReplaceCollision(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	deps := ApplyDeps{
		BeforePublish: func(placement AppliedPlacement) error {
			root, err := session.Layout.ManagedSkillsPath(placement.Target)
			if err != nil {
				return err
			}
			destination := filepath.Join(root, placement.Skill)
			if err := os.Mkdir(destination, 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(destination, "external"), []byte("keep\n"), 0o644)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectChanges unexpectedly replaced a raced destination")
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "external")
	content, err := os.ReadFile(destination)
	if err != nil || string(content) != "keep\n" {
		t.Fatalf("raced destination content = %q, err=%v", content, err)
	}
}

func TestApplyProjectChangesReclaimsRecognizableStaleLock(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName)
	if err := os.WriteFile(lockPath, []byte(applyLockMarker), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err != nil || len(result.Installed) != 1 || result.Installed[0].Skill != skill.Name {
		t.Fatalf("stale-lock apply result=%#v err=%v", result, err)
	}
	if _, err := os.Lstat(lockPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale lock survived successful apply: %v", err)
	}
}

func TestApplyProjectChangesRejectsRecognizableHeldLock(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName)
	handle, _, err := openApplyLockFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	if err := lockApplyFile(handle); err != nil {
		t.Fatal(err)
	}
	defer unlockApplyFile(handle)
	_, err = ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) || !applyErr.Conflict() || err.Error() != "project apply conflict: another project apply is active" {
		t.Fatalf("held-lock error = %v", err)
	}
}

func TestApplyProjectChangesRecoversInterruptedInstallBeforeRerun(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents})
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	preimage := &applyStatePreimage{state: emptyTrustedProvenanceState()}
	recordedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	candidate, err := buildApplyState(preimage.state, session, recordedAt)
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
	if got := hashSkillAt(t, destination); got != hash {
		t.Fatalf("simulated landed install hash=%#v, want %#v", got, hash)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, candidateData, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("interrupted recovery error = %v", err)
	}
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted placement survived recovery: %v", err)
	}
	if _, err := os.Lstat(session.Layout.ReconcilerStatePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted provenance survived recovery: %v", err)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("transaction journal survived recovery: %v", err)
	}
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err != nil || len(result.Installed) != 1 || result.Installed[0].Skill != skill.Name {
		t.Fatalf("post-recovery rerun result=%#v err=%v", result, err)
	}
}

func TestApplyProjectChangesRollsBackLandedStateReplacementError(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	landed := false
	deps := ApplyDeps{publishStateNoReplace: func(source, destination string) error {
		if err := publishNoReplace(source, destination); err != nil {
			return err
		}
		landed = true
		return errors.New("injected error after atomic replacement")
	}}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil || !landed {
		t.Fatalf("landed replacement result err=%v landed=%v", err, landed)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("placement survived landed state rollback: %v", err)
	}
	if _, err := os.Lstat(session.Layout.ReconcilerStatePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("landed state survived rollback: %v", err)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("journal survived landed state rollback: %v", err)
	}
}

func TestApplyProjectChangesRehashesTemporaryPlacementImmediatelyBeforePublish(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	var staged string
	deps := ApplyDeps{
		MakeTempDir: func(parent, pattern string) (string, error) {
			path, err := os.MkdirTemp(parent, pattern)
			staged = path
			return path, err
		},
		BeforePublish: func(AppliedPlacement) error {
			return os.WriteFile(filepath.Join(staged, "raced"), []byte("external\n"), 0o644)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil || err.Error() != "project apply conflict: temporary placement changed before publication" {
		t.Fatalf("staged publication race error=%v", err)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("raced temporary placement was published: %v", err)
	}
}

func TestApplyProjectChangesRecoversAfterAbruptProcessExit(t *testing.T) {
	const childEnv = "SJSKILLS_ABRUPT_APPLY_CHILD"
	const rootEnv = "SJSKILLS_ABRUPT_APPLY_ROOT"
	const stageEnv = "SJSKILLS_ABRUPT_APPLY_STAGE"
	if os.Getenv(childEnv) == "1" {
		session, _, _, _ := newApplyFixtureAt(t, os.Getenv(rootEnv), os.Getenv(stageEnv), []Target{TargetAgents})
		_, _ = ApplyProjectChanges(context.Background(), session, ApplyDeps{afterStateCommit: func() { os.Exit(91) }})
		os.Exit(92)
	}
	project := t.TempDir()
	stage := t.TempDir()
	child := exec.Command(os.Args[0], "-test.run=^TestApplyProjectChangesRecoversAfterAbruptProcessExit$")
	child.Env = append(os.Environ(), childEnv+"=1", rootEnv+"="+project, stageEnv+"="+stage)
	output, err := child.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 91 {
		t.Fatalf("abrupt child err=%v output=%q", err, output)
	}
	session, _, skill, _ := newApplyFixtureAt(t, project, stage, []Target{TargetAgents})
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); err != nil {
		t.Fatalf("abrupt child did not leave journal: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName)); err != nil || string(data) != applyLockMarker {
		t.Fatalf("abrupt child lock data=%q err=%v", data, err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("abrupt recovery error=%v", err)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("abrupt placement survived recovery: %v", err)
	}
	if _, err := os.Lstat(session.Layout.ReconcilerStatePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("abrupt provenance survived recovery: %v", err)
	}
}

func TestApplyProjectChangesDoesNotOwnSwappedInstallPublication(t *testing.T) {
	session, _, skill, expected := newApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	var externalInfo os.FileInfo
	deps := ApplyDeps{PublishNoReplace: func(source, target string) error {
		info, err := publishThenSwapWithIdenticalExternal(source, target)
		externalInfo = info
		return err
	}}

	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err == nil {
		t.Fatal("apply unexpectedly owned a swapped install publication")
	}
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) || !applyErr.Conflict() {
		t.Fatalf("swapped install error = %v, want conflict", err)
	}
	if len(result.Installed) != 0 || result.Quarantine != nil {
		t.Fatalf("swapped install result = %#v", result)
	}
	currentInfo, statErr := os.Lstat(destination)
	if statErr != nil || externalInfo == nil || !os.SameFile(externalInfo, currentInfo) {
		t.Fatalf("external install destination was not preserved: info=%v err=%v", currentInfo, statErr)
	}
	if got := hashSkillAt(t, destination); got != expected {
		t.Fatalf("external install hash = %#v, want %#v", got, expected)
	}
	if _, statErr := os.Lstat(session.Layout.ReconcilerStatePath); !os.IsNotExist(statErr) {
		t.Fatalf("swapped install wrote provenance: err=%v", statErr)
	}
}

func TestApplyProjectChangesRejectsSnapshotTamperBeforePublication(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	deps := ApplyDeps{
		BeforePublish: func(AppliedPlacement) error {
			snapshot, ok := session.Materialized.SnapshotFor(skill.Name)
			if !ok {
				return errors.New("snapshot missing")
			}
			return os.WriteFile(filepath.Join(snapshot.Path, "tampered"), []byte("changed\n"), 0o644)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectChanges unexpectedly published a tampered snapshot")
	}
	if _, err := os.Stat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !os.IsNotExist(err) {
		t.Fatalf("tampered snapshot was published: err=%v", err)
	}
}

func TestApplyProjectChangesNoOpRevalidatesAfterReviewedPlan(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	makeSessionPlanCurrent(t, session)
	stateBefore, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	deps := ApplyDeps{beforeCommit: func() error {
		return os.WriteFile(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "drift"), []byte("changed\n"), 0o644)
	}}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("no-op apply unexpectedly reported success after placement drift")
	}
	stateAfter, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stateAfter) != string(stateBefore) {
		t.Fatal("no-op drift changed provenance")
	}
}

func TestApplyProjectChangesMixedPlanRevalidatesUnchangedPlacementBeforeCommit(t *testing.T) {
	session, _, prior, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	priorState, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	added := addApplyFixtureSkill(t, session, "added", []Target{TargetAgents})
	makeSessionPlanCurrent(t, session)
	actions := map[string]PlanAction{}
	for _, operation := range session.Plan.Operations {
		actions[operation.Skill] = operation.Action
	}
	if actions[prior.Name] != PlanActionUnchanged || actions[added.Name] != PlanActionInstall {
		t.Fatalf("mixed plan actions = %#v", actions)
	}
	deps := ApplyDeps{beforeCommit: func() error {
		return os.WriteFile(filepath.Join(session.Layout.AgentsSkillsPath, prior.Name, "drift"), []byte("changed\n"), 0o644)
	}}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("mixed apply unexpectedly committed after unchanged placement drift")
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, added.Name)); !os.IsNotExist(err) {
		t.Fatalf("new placement survived failed candidate proof: err=%v", err)
	}
	stateAfter, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stateAfter) != string(priorState) {
		t.Fatal("failed mixed apply changed provenance")
	}
	if _, err := os.Stat(filepath.Join(session.Layout.AgentsSkillsPath, prior.Name, "drift")); err != nil {
		t.Fatalf("raced unchanged content was not preserved: %v", err)
	}
}

func TestApplyProjectChangesRollbackPreservesRacedReplacement(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents, TargetClaude})
	publishCalls := 0
	replaced := false
	deps := ApplyDeps{
		PublishNoReplace: func(source, destination string) error {
			publishCalls++
			if publishCalls == 2 {
				return errors.New("injected publication failure")
			}
			return publishNoReplace(source, destination)
		},
		beforeRollback: func(placement AppliedPlacement) error {
			if replaced {
				return nil
			}
			replaced = true
			destination := filepath.Join(session.Layout.AgentsSkillsPath, placement.Skill)
			if err := os.RemoveAll(destination); err != nil {
				return err
			}
			if err := os.Mkdir(destination, 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(destination, "external"), []byte("preserve\n"), 0o644)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly succeeded after rollback race")
	}
	if !replaced {
		t.Fatal("rollback race hook did not run")
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	data, err := os.ReadFile(filepath.Join(destination, "external"))
	if err != nil || string(data) != "preserve\n" {
		t.Fatalf("raced destination content = %q, err=%v", data, err)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); err != nil {
		t.Fatalf("recovery journal disappeared after ambiguous rollback: %v", err)
	}
}

func TestApplyProjectChangesLockAcquisitionCleansOnlyCreatedDirectory(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	deps := ApplyDeps{beforeLock: func() error { return errors.New("injected preflight failure") }}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly succeeded after lock preflight failure")
	}
	if _, err := os.Lstat(session.Layout.DerivedDirectoryPath); !os.IsNotExist(err) {
		t.Fatalf("proved-created derived directory survived acquisition failure: err=%v", err)
	}

	competitor := []byte("competitor\n")
	deps.beforeLock = func() error {
		return os.WriteFile(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName), competitor, 0o600)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly acquired a raced lock")
	}
	data, err := os.ReadFile(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName))
	if err != nil || string(data) != string(competitor) {
		t.Fatalf("competitor lock was not preserved: data=%q err=%v", data, err)
	}
}

func TestApplyProjectChangesPreservesExistingUnknownWithoutDuplicatingWarning(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	unknown := filepath.Join(session.Layout.AgentsSkillsPath, "unknown")
	if err := os.MkdirAll(unknown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unknown, "SKILL.md"), []byte("unknown\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Warnings) != 1 {
		t.Fatalf("reviewed warnings = %#v, want one", session.Plan.Warnings)
	}
	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Plan.Warnings) != 1 || result.Plan.Warnings[0].Code != "unmanaged-preserved" {
		t.Fatalf("applied warnings = %#v, want one preserved warning", result.Plan.Warnings)
	}
	if data, err := os.ReadFile(filepath.Join(unknown, "SKILL.md")); err != nil || string(data) != "unknown\n" {
		t.Fatalf("unknown entry changed: data=%q err=%v", data, err)
	}
}

func TestApplyProjectChangesReportsUnknownAppearingBeforeLockedRefresh(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	unknown := filepath.Join(session.Layout.AgentsSkillsPath, "raced-unknown")
	deps := ApplyDeps{beforeLock: func() error {
		if err := os.MkdirAll(unknown, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(unknown, "SKILL.md"), []byte("preserve\n"), 0o644)
	}}
	result, err := ApplyProjectChanges(context.Background(), session, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Plan.Warnings) != 1 || !strings.Contains(result.Plan.Warnings[0].Message, "raced-unknown") {
		t.Fatalf("applied warnings = %#v, want raced unknown", result.Plan.Warnings)
	}
	if data, err := os.ReadFile(filepath.Join(unknown, "SKILL.md")); err != nil || string(data) != "preserve\n" {
		t.Fatalf("raced unknown changed: data=%q err=%v", data, err)
	}
}

func TestApplyProjectChangesReportsLostOrReplacedLock(t *testing.T) {
	for _, test := range []struct {
		name    string
		replace bool
	}{
		{name: "lost"},
		{name: "replaced", replace: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
			lockPath := filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName)
			deps := ApplyDeps{beforeCommit: func() error {
				if err := os.Remove(lockPath); err != nil {
					return err
				}
				if test.replace {
					return os.WriteFile(lockPath, []byte("replacement\n"), 0o600)
				}
				return nil
			}}
			if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil || !strings.Contains(err.Error(), "lock changed") {
				t.Fatalf("ApplyProjectChanges() error = %v, want lock changed", err)
			}
			if test.replace {
				data, err := os.ReadFile(lockPath)
				if err != nil || string(data) != "replacement\n" {
					t.Fatalf("replacement lock changed: data=%q err=%v", data, err)
				}
			}
		})
	}
}

func TestApplyProjectChangesSyncsStagedHierarchyAndPublishedParent(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	snapshot, _ := session.Materialized.SnapshotFor(skill.Name)
	nested := filepath.Join(snapshot.Path, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "file"), []byte("nested\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(snapshot.Path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Hash = hash
	session.Expected[skill.Name] = hash
	session.Plan.Operations[0].Expected.Detail = hash.Algorithm + ":" + hash.Digest
	var syncedDirs []string
	var syncedFiles []string
	deps := ApplyDeps{
		SyncFile: func(file *os.File) error {
			syncedFiles = append(syncedFiles, filepath.Base(file.Name()))
			return file.Sync()
		},
		SyncDir: func(path string) error {
			syncedDirs = append(syncedDirs, path)
			return syncApplyDirectory(path)
		},
	}
	if _, err := ApplyProjectChanges(context.Background(), session, deps); err != nil {
		t.Fatal(err)
	}
	if len(syncedFiles) < 3 { // two copied files plus the provenance temporary file
		t.Fatalf("synced files = %v", syncedFiles)
	}
	rootIndexes := []int{}
	agentsIndex, nestedIndex, stageIndex, parentIndex := -1, -1, -1, -1
	for index, path := range syncedDirs {
		switch {
		case path == session.Layout.Root:
			rootIndexes = append(rootIndexes, index)
		case path == filepath.Dir(session.Layout.AgentsSkillsPath):
			agentsIndex = index
		case filepath.Base(path) == "nested":
			nestedIndex = index
		case strings.HasPrefix(filepath.Base(path), applyInstallPattern):
			stageIndex = index
		case path == session.Layout.AgentsSkillsPath:
			parentIndex = index
		}
	}
	if len(rootIndexes) < 2 || agentsIndex <= rootIndexes[1] || nestedIndex <= agentsIndex || stageIndex <= nestedIndex || parentIndex <= stageIndex {
		t.Fatalf("directory sync order = %v", syncedDirs)
	}
}

func TestApplyProjectChangesRejectsMaterializedSymlinkBeforeWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Windows privileges")
	}
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	snapshot, _ := session.Materialized.SnapshotFor(skill.Name)
	if err := os.WriteFile(filepath.Join(snapshot.Path, "target"), []byte("target\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(snapshot.Path, "safe-link")); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(snapshot.Path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Hash = hash
	session.Expected[skill.Name] = hash
	session.Plan.Operations[0].Expected.Detail = hash.Algorithm + ":" + hash.Digest
	if err := snapshot.Verify(); err != nil {
		t.Fatalf("safe in-tree symlink was not accepted by materializer verification: %v", err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("apply unexpectedly accepted a symlink copy tree")
	}
	if _, err := os.Lstat(session.Layout.DerivedDirectoryPath); !os.IsNotExist(err) {
		t.Fatalf("symlink rejection wrote derived state: err=%v", err)
	}
}

func newApplyFixture(t *testing.T, targets []Target) (*ProjectApplySession, string, DesiredSkill, TreeHash) {
	t.Helper()
	project := t.TempDir()
	stage := t.TempDir()
	return newApplyFixtureAt(t, project, stage, targets)
}

func newApplyFixtureAt(t *testing.T, project, stage string, targets []Target) (*ProjectApplySession, string, DesiredSkill, TreeHash) {
	t.Helper()
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutForProject(canonicalProject)
	if err != nil {
		t.Fatal(err)
	}
	skill := DesiredSkill{Name: "demo", Source: "example/demo", SourceID: "demo-source", Scope: ScopeProject, Origin: "test", Manager: ManagerSkillsCLI, Mode: ModeCopy, Targets: targets}
	snapshotPath := filepath.Join(stage, ".agents", "skills", skill.Name)
	if err := os.MkdirAll(snapshotPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshotPath, "SKILL.md"), []byte("demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	materialized := &MaterializationPlan{root: stage, snapshots: map[string]*SkillSnapshot{}}
	materialized.snapshots[skill.Name] = &SkillSnapshot{Skill: skill, Path: snapshotPath, Hash: hash, plan: materialized, stageRoot: stage}
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{skill}}
	operations := make([]PlanOperation, 0, len(targets))
	for _, target := range targets {
		operations = append(operations, PlanOperation{Action: PlanActionInstall, Skill: skill.Name, Target: target, Manager: ManagerSkillsCLI, SourceID: skill.SourceID, Source: skill.Source, Reason: string(ProjectStateReasonExpectedEntryAbsent), Current: PlanEvidence{Kind: projectEvidenceAbsent, Detail: projectEvidenceAbsent}, Expected: PlanEvidence{Kind: projectEvidenceTreeHash, Detail: hash.Algorithm + ":" + hash.Digest}})
	}
	return &ProjectApplySession{Layout: layout, Desired: desired, Plan: Plan{Desired: desired, Operations: operations}, Expected: map[string]TreeHash{skill.Name: hash}, Materialized: materialized}, project, skill, hash
}

func applyFixture(t *testing.T, session *ProjectApplySession) {
	t.Helper()
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{Now: func() time.Time {
		return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	}}); err != nil {
		t.Fatal(err)
	}
}

func makeSessionPlanCurrent(t *testing.T, session *ProjectApplySession) {
	t.Helper()
	inventory, err := InspectProject(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyProject(session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := TranslateProjectClassification(Plan{Desired: session.Desired}, classification)
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
}

func addApplyFixtureSkill(t *testing.T, session *ProjectApplySession, name string, targets []Target) DesiredSkill {
	t.Helper()
	skill := DesiredSkill{Name: name, Source: "example/" + name, SourceID: name + "-source", Scope: ScopeProject, Origin: "test", Manager: ManagerSkillsCLI, Mode: ModeCopy, Targets: targets}
	path := filepath.Join(session.Materialized.root, ".agents", "skills", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(path)
	if err != nil {
		t.Fatal(err)
	}
	session.Materialized.snapshots[name] = &SkillSnapshot{Skill: skill, Path: path, Hash: hash, plan: session.Materialized, stageRoot: session.Materialized.root}
	session.Desired.Skills = append(session.Desired.Skills, skill)
	sort.SliceStable(session.Desired.Skills, func(i, j int) bool {
		return compareUTF16(session.Desired.Skills[i].Name, session.Desired.Skills[j].Name) < 0
	})
	session.Expected[name] = hash
	return skill
}

func publishThenSwapWithIdenticalExternal(source, destination string) (os.FileInfo, error) {
	if err := publishNoReplace(source, destination); err != nil {
		return nil, err
	}
	external, err := os.MkdirTemp(filepath.Dir(destination), ".external-swap-")
	if err != nil {
		return nil, err
	}
	if err := copyApplyTree(destination, external, func(*os.File) error { return nil }); err != nil {
		return nil, err
	}
	displaced := destination + ".published-swap"
	if err := publishNoReplace(destination, displaced); err != nil {
		return nil, err
	}
	if err := publishNoReplace(external, destination); err != nil {
		return nil, err
	}
	info, err := os.Lstat(destination)
	if err != nil {
		return nil, err
	}
	if err := os.RemoveAll(displaced); err != nil {
		return nil, err
	}
	return info, nil
}
