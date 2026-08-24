package sjskills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestApplyProjectInstallsRejectsBlockedPlanBeforeLock(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{}); err == nil {
		t.Fatal("ApplyProjectInstalls unexpectedly accepted blocked plan")
	} else {
		var applyErr *ApplyError
		if !errors.As(err, &applyErr) || !applyErr.Conflict() {
			t.Fatalf("ApplyProjectInstalls() error = %v, want conflict", err)
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

func TestApplyProjectInstallsCopiesOnePlacementAndWritesSortedProvenance(t *testing.T) {
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
	result, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt }})
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
	if err != nil || placedInfo.Mode().Perm()&0o111 == 0 {
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

func TestApplyProjectInstallsReportsCommittedPlacementsOnFinalizationFailure(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents})
	result, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{
		Now: func() time.Time { return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC) },
		beforeUnlock: func() error {
			return errors.New("injected finalization failure")
		},
	})
	if err == nil {
		t.Fatal("ApplyProjectInstalls unexpectedly succeeded after finalization failure")
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

func TestApplyProjectInstallsRollsBackEarlierPlacementsOnPublishFailure(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectInstalls unexpectedly succeeded after injected publication failure")
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

func TestApplyProjectInstallsPreservesNoReplaceCollision(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectInstalls unexpectedly replaced a raced destination")
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "external")
	content, err := os.ReadFile(destination)
	if err != nil || string(content) != "keep\n" {
		t.Fatalf("raced destination content = %q, err=%v", content, err)
	}
}

func TestApplyProjectInstallsRejectsSnapshotTamperBeforePublication(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("ApplyProjectInstalls unexpectedly published a tampered snapshot")
	}
	if _, err := os.Stat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !os.IsNotExist(err) {
		t.Fatalf("tampered snapshot was published: err=%v", err)
	}
}

func TestApplyProjectInstallsNoOpRevalidatesAfterReviewedPlan(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
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

func TestApplyProjectInstallsMixedPlanRevalidatesUnchangedPlacementBeforeCommit(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
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

func TestApplyProjectInstallsRollbackPreservesRacedReplacement(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly succeeded after rollback race")
	}
	if !replaced {
		t.Fatal("rollback race hook did not run")
	}
	recovered, err := filepath.Glob(filepath.Join(session.Layout.DerivedDirectoryPath, ".sjskills-recovery-*", "*", "external"))
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 1 {
		t.Fatalf("recovered raced files = %v, want one", recovered)
	}
	data, err := os.ReadFile(recovered[0])
	if err != nil || string(data) != "preserve\n" {
		t.Fatalf("recovered raced content = %q, err=%v", data, err)
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); !os.IsNotExist(err) {
		t.Fatalf("raced source path survived atomic preservation move: err=%v", err)
	}
}

func TestApplyProjectInstallsLockAcquisitionCleansOnlyCreatedDirectory(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	deps := ApplyDeps{beforeLock: func() error { return errors.New("injected preflight failure") }}
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly succeeded after lock preflight failure")
	}
	if _, err := os.Lstat(session.Layout.DerivedDirectoryPath); !os.IsNotExist(err) {
		t.Fatalf("proved-created derived directory survived acquisition failure: err=%v", err)
	}

	competitor := []byte("competitor\n")
	deps.beforeLock = func() error {
		return os.WriteFile(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName), competitor, 0o600)
	}
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil {
		t.Fatal("apply unexpectedly acquired a raced lock")
	}
	data, err := os.ReadFile(filepath.Join(session.Layout.DerivedDirectoryPath, applyLockName))
	if err != nil || string(data) != string(competitor) {
		t.Fatalf("competitor lock was not preserved: data=%q err=%v", data, err)
	}
}

func TestApplyProjectInstallsPreservesExistingUnknownWithoutDuplicatingWarning(t *testing.T) {
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
	result, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{})
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

func TestApplyProjectInstallsReportsUnknownAppearingBeforeLockedRefresh(t *testing.T) {
	session, _, _, _ := newApplyFixture(t, []Target{TargetAgents})
	unknown := filepath.Join(session.Layout.AgentsSkillsPath, "raced-unknown")
	deps := ApplyDeps{beforeLock: func() error {
		if err := os.MkdirAll(unknown, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(unknown, "SKILL.md"), []byte("preserve\n"), 0o644)
	}}
	result, err := ApplyProjectInstalls(context.Background(), session, deps)
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

func TestApplyProjectInstallsReportsLostOrReplacedLock(t *testing.T) {
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
			if _, err := ApplyProjectInstalls(context.Background(), session, deps); err == nil || !strings.Contains(err.Error(), "lock changed") {
				t.Fatalf("ApplyProjectInstalls() error = %v, want lock changed", err)
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

func TestApplyProjectInstallsSyncsStagedHierarchyAndPublishedParent(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, deps); err != nil {
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

func TestApplyProjectInstallsRejectsMaterializedSymlinkBeforeWrite(t *testing.T) {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{}); err == nil {
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
	if _, err := ApplyProjectInstalls(context.Background(), session, ApplyDeps{Now: func() time.Time {
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
