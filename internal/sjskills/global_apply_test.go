package sjskills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGlobalMutationErrorsNameTheGlobalBoundary(t *testing.T) {
	if _, err := ApplyGlobalChanges(context.Background(), nil, ApplyDeps{}); err == nil ||
		err.Error() != "global apply unavailable: materialization session is unavailable" {
		t.Fatalf("apply error = %v", err)
	}
	layout, layoutErr := LayoutForGlobal(canonicalTempHome(t))
	if layoutErr != nil {
		t.Fatal(layoutErr)
	}
	if _, err := RestoreGlobalQuarantine(context.Background(), layout, "invalid", ApplyDeps{}); err == nil ||
		err.Error() != "global restore conflict: quarantine identifier is invalid" {
		t.Fatalf("restore error = %v", err)
	}
}

func TestApplyGlobalChangesRollsBackPartialInstall(t *testing.T) {
	session := newGlobalApplyFixture(t)
	calls := 0
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{
		Now: func() time.Time { return time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC) },
		BeforePublish: func(AppliedPlacement) error {
			calls++
			if calls == 2 {
				return errors.New("injected publish failure")
			}
			return nil
		},
	})
	if err == nil || len(result.Installed) != 0 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, _ := session.Layout.ManagedSkillsPath(target)
		if _, statErr := os.Lstat(filepath.Join(root, "base")); !os.IsNotExist(statErr) {
			t.Fatalf("%s placement survived rollback: %v", target, statErr)
		}
	}
	if _, statErr := os.Lstat(session.Layout.ProvenanceStatePath); !os.IsNotExist(statErr) {
		t.Fatalf("global provenance survived rollback: %v", statErr)
	}
}

func TestApplyGlobalChangesInstallsFixedBaselineAndWritesCurrentState(t *testing.T) {
	session := newGlobalApplyFixture(t)
	recordedAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt }})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Installed) != 2 || len(result.Updated) != 0 || result.Quarantine != nil {
		t.Fatalf("result = %#v", result)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, _ := session.Layout.ManagedSkillsPath(target)
		got, hashErr := HashSkillTree(filepath.Join(root, "base"))
		if hashErr != nil || got != session.Expected["base"] {
			t.Fatalf("%s placement hash=%#v err=%v", target, got, hashErr)
		}
	}
	data, err := os.ReadFile(session.Layout.ProvenanceStatePath)
	if err != nil {
		t.Fatal(err)
	}
	var state GlobalProvenanceState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	if state.Version != GlobalProvenanceStateVersion || len(state.Records) != 2 {
		t.Fatalf("state = %#v", state)
	}
	for _, record := range state.Records {
		if record.Scope != ScopeGlobal || record.RecordedAt != recordedAt {
			t.Fatalf("record = %#v", record)
		}
	}
	if _, err := os.Lstat(filepath.Join(session.Layout.DerivedStatePath, applyLockName)); !os.IsNotExist(err) {
		t.Fatalf("global lock survived apply: %v", err)
	}
}

func TestApplyGlobalChangesUpdatesThroughDurableQuarantine(t *testing.T) {
	session := newGlobalApplyFixture(t)
	firstTime := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	if _, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return firstTime }}); err != nil {
		t.Fatal(err)
	}
	oldHash := session.Expected["base"]
	snapshot, _ := session.Materialized.SnapshotFor("base")
	if err := os.WriteFile(filepath.Join(snapshot.Path, "SKILL.md"), []byte("base-v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newHash, err := HashSkillTree(snapshot.Path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Hash = newHash
	session.Expected["base"] = newHash
	inventory, err := InspectGlobal(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(session.Registry, session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	session.Plan, err = TranslateGlobalClassification(Plan{Desired: session.Desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return firstTime.Add(time.Hour) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Updated) != 2 || result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineCommitted {
		t.Fatalf("result = %#v", result)
	}
	manifestPath := filepath.Join(session.Layout.QuarantinePath, result.Quarantine.ID, applyManifestName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, valid := DecodeProjectQuarantineManifest(data)
	if !valid || len(manifest.Entries) != 2 {
		t.Fatalf("manifest valid=%v value=%#v", valid, manifest)
	}
	for _, entry := range manifest.Entries {
		if entry.OldTreeHash != oldHash.Digest || entry.NewTreeHash != newHash.Digest {
			t.Fatalf("entry = %#v", entry)
		}
		quarantined := filepath.Join(session.Layout.QuarantinePath, result.Quarantine.ID, filepath.FromSlash(entry.QuarantinedPlacement))
		got, hashErr := HashSkillTree(quarantined)
		if hashErr != nil || got != oldHash {
			t.Fatalf("quarantined hash=%#v err=%v", got, hashErr)
		}
	}
}

func TestRestoreGlobalQuarantineRestoresOldGlobalPlacements(t *testing.T) {
	session := newGlobalApplyFixture(t)
	recordedAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	if _, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt }}); err != nil {
		t.Fatal(err)
	}
	oldHash := session.Expected["base"]
	snapshot, _ := session.Materialized.SnapshotFor("base")
	if err := os.WriteFile(filepath.Join(snapshot.Path, "SKILL.md"), []byte("base-v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newHash, _ := HashSkillTree(snapshot.Path)
	snapshot.Hash = newHash
	session.Expected["base"] = newHash
	inventory, _ := InspectGlobal(session.Layout)
	classification, _ := ClassifyGlobal(session.Registry, session.Desired, session.Expected, inventory)
	session.Plan, _ = TranslateGlobalClassification(Plan{Desired: session.Desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	updated, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt.Add(time.Hour) }})
	if err != nil || updated.Quarantine == nil {
		t.Fatalf("update result=%#v err=%v", updated, err)
	}
	parked := t.TempDir()
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, _ := session.Layout.ManagedSkillsPath(target)
		if err := os.Rename(filepath.Join(root, "base"), filepath.Join(parked, string(target)+"-base")); err != nil {
			t.Fatal(err)
		}
	}
	restored, err := RestoreGlobalQuarantine(context.Background(), session.Layout, updated.Quarantine.ID, ApplyDeps{Now: func() time.Time { return recordedAt.Add(2 * time.Hour) }})
	if err != nil {
		t.Fatal(err)
	}
	if restored.Status != ProjectQuarantineRestored || len(restored.Restored) != 2 {
		t.Fatalf("restored = %#v", restored)
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, _ := session.Layout.ManagedSkillsPath(target)
		got, hashErr := HashSkillTree(filepath.Join(root, "base"))
		if hashErr != nil || got != oldHash {
			t.Fatalf("%s restored hash=%#v err=%v", target, got, hashErr)
		}
	}
	inventory, err = InspectGlobal(session.Layout)
	if err != nil || inventory.ProvenanceFormat != GlobalProvenanceCurrent || len(inventory.State.Records) != 2 {
		t.Fatalf("inventory=%#v err=%v", inventory, err)
	}
	for _, record := range inventory.State.Records {
		if record.Scope != ScopeGlobal || record.TreeHash != oldHash.Digest {
			t.Fatalf("record = %#v", record)
		}
	}
}

func TestApplyGlobalChangesMigratesTrustedLegacyStateWithoutReplacingPlacements(t *testing.T) {
	session := newGlobalApplyFixture(t)
	hash := session.Expected["base"]
	for _, target := range []Target{TargetAgents, TargetClaude} {
		root, _ := session.Layout.ManagedSkillsPath(target)
		got := writeGlobalSkill(t, root, "base", "base\n")
		if got != hash {
			t.Fatalf("fixture hash=%#v want=%#v", got, hash)
		}
	}
	formerHash := writeGlobalSkill(t, session.Layout.AgentsSkillsPath, "former", "former\n")
	recordedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	writeLegacyGlobalState(t, session.Layout.ProvenanceStatePath, []legacyGlobalRecordFixture{
		{Root: "shared", Skill: "base", Source: "example/skills", HashAlgorithm: hash.Algorithm, Hash: hash.Digest, RecordedAt: recordedAt},
		{Root: "claude", Skill: "base", Source: "example/skills", HashAlgorithm: hash.Algorithm, Hash: hash.Digest, RecordedAt: recordedAt},
		{Root: "shared", Skill: "former", Source: "example/skills", HashAlgorithm: formerHash.Algorithm, Hash: formerHash.Digest, RecordedAt: recordedAt},
	})
	inventory, err := InspectGlobal(session.Layout)
	if err != nil || inventory.ProvenanceFormat != GlobalProvenanceLegacyV1 {
		t.Fatalf("legacy inventory=%#v err=%v", inventory, err)
	}
	classification, err := ClassifyGlobal(session.Registry, session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	session.Plan, err = TranslateGlobalClassification(Plan{Desired: session.Desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt.Add(time.Hour) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Installed)+len(result.Updated)+len(result.Quarantined) != 0 || result.Quarantine != nil || !result.Migrated {
		t.Fatalf("result = %#v", result)
	}
	inventory, err = InspectGlobal(session.Layout)
	if err != nil || inventory.ProvenanceFormat != GlobalProvenanceCurrent || inventory.MigrationRequired {
		t.Fatalf("current inventory=%#v err=%v", inventory, err)
	}
	if len(inventory.State.Records) != 3 {
		t.Fatalf("migrated records = %#v", inventory.State.Records)
	}
	foundFormer := false
	for _, record := range inventory.State.Records {
		if record.Target == TargetAgents && record.Skill == "former" {
			foundFormer = record.Scope == ScopeGlobal && record.TreeHash == formerHash.Digest
		}
	}
	if !foundFormer {
		t.Fatalf("trusted former-profile provenance was not preserved: %#v", inventory.State.Records)
	}
}

func TestApplyGlobalChangesRejectsMigrationStatusRace(t *testing.T) {
	session := newGlobalApplyFixture(t)
	recordedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if _, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt }}); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectGlobal(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(session.Registry, session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	session.Plan, err = TranslateGlobalClassification(Plan{Desired: session.Desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil || planHasEvidence(session.Plan, "provenance-migration") {
		t.Fatalf("current plan=%#v err=%v", session.Plan, err)
	}
	hash := session.Expected["base"]
	writeLegacyGlobalState(t, session.Layout.ProvenanceStatePath, []legacyGlobalRecordFixture{
		{Root: "shared", Skill: "base", Source: "example/skills", HashAlgorithm: hash.Algorithm, Hash: hash.Digest, RecordedAt: recordedAt},
		{Root: "claude", Skill: "base", Source: "example/skills", HashAlgorithm: hash.Algorithm, Hash: hash.Digest, RecordedAt: recordedAt},
	})
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: func() time.Time { return recordedAt.Add(time.Hour) }})
	if err == nil || err.Error() != "global apply conflict: global provenance migration changed after planning" || result.Migrated {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	inventory, inspectErr := InspectGlobal(session.Layout)
	if inspectErr != nil || inventory.ProvenanceFormat != GlobalProvenanceLegacyV1 {
		t.Fatalf("raced provenance was mutated: inventory=%#v err=%v", inventory, inspectErr)
	}
}

func newGlobalApplyFixture(t *testing.T) *GlobalApplySession {
	t.Helper()
	home := canonicalTempHome(t)
	layout, err := LayoutForGlobal(home)
	if err != nil {
		t.Fatal(err)
	}
	registry := minimalGlobalRegistry(t)
	desired, err := ResolveGlobal(registry)
	if err != nil {
		t.Fatal(err)
	}
	stage := t.TempDir()
	snapshotPath := filepath.Join(stage, ".agents", "skills", "base")
	if err := os.MkdirAll(snapshotPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshotPath, "SKILL.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	materialized := &MaterializationPlan{root: stage, snapshots: map[string]*SkillSnapshot{}}
	snapshot := &SkillSnapshot{Skill: desired.Skills[0], Path: snapshotPath, Hash: hash, plan: materialized, stageRoot: stage}
	materialized.snapshots["base"] = snapshot
	base := Plan{Desired: desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}
	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(registry, desired, map[string]TreeHash{"base": hash}, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := TranslateGlobalClassification(base, classification)
	if err != nil {
		t.Fatal(err)
	}
	return &GlobalApplySession{
		Layout: layout, Registry: registry, Desired: desired, Plan: plan,
		Expected: map[string]TreeHash{"base": hash}, Materialized: materialized,
	}
}
