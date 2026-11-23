package sjskills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// rewriteStateAsLegacy turns the session's state into what a release before
// tree-sha256-v3 would have written for the same placements.
func rewriteStateAsLegacy(t *testing.T, session *ProjectApplySession) ProvenanceState {
	t.Helper()
	state := readApplyState(t, session)
	for index, record := range state.Records {
		root, err := session.Layout.ManagedSkillsPath(record.Target)
		if err != nil {
			t.Fatal(err)
		}
		legacy := legacyTreeHash(t, filepath.Join(root, record.Skill))
		state.Records[index].TreeHashAlgorithm = legacy.Algorithm
		state.Records[index].TreeHash = legacy.Digest
	}
	writeApplyStateFixture(t, session, state)
	return state
}

func writeApplyStateFixture(t *testing.T, session *ProjectApplySession, state ProvenanceState) {
	t.Helper()
	data, err := marshalApplyState(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyTreeHashProvenanceIsVerifiedAndPersistedOnApply(t *testing.T) {
	targets := []Target{TargetAgents, TargetClaude}
	session, _, _, hash := newApplyFixture(t, targets)
	applyFixture(t, session)
	legacyState := rewriteStateAsLegacy(t, session)

	inventory, err := InspectProject(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.TreeHashMigration != len(targets) {
		t.Fatalf("tree hash migration = %d, want %d", inventory.TreeHashMigration, len(targets))
	}
	for _, record := range inventory.State.Records {
		if record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V3 || record.TreeHash != hash.Digest {
			t.Fatalf("legacy record was not read as the placement's v3 digest: %#v", record)
		}
	}
	makeSessionPlanCurrent(t, session)
	if !planHasEvidence(session.Plan, treeHashMigrationEvidenceKind) {
		t.Fatalf("plan does not report the provenance rewrite: %#v", session.Plan.Evidence)
	}
	for _, operation := range session.Plan.Operations {
		if operation.Action != PlanActionUnchanged || operation.Reason != string(ProjectStateReasonVerifiedExact) {
			t.Fatalf("legacy provenance did not verify the placement: %#v", operation)
		}
	}

	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !result.TreeHashesUpgraded || len(result.Installed)+len(result.Updated)+len(result.Quarantined) != 0 {
		t.Fatalf("result = %#v", result)
	}
	persisted := readApplyState(t, session)
	if len(persisted.Records) != len(legacyState.Records) {
		t.Fatalf("persisted records = %#v", persisted.Records)
	}
	for index, record := range persisted.Records {
		want := legacyState.Records[index]
		want.TreeHashAlgorithm, want.TreeHash = hash.Algorithm, hash.Digest
		if record != want {
			t.Fatalf("persisted record = %#v, want %#v", record, want)
		}
	}

	makeSessionPlanCurrent(t, session)
	if planHasEvidence(session.Plan, treeHashMigrationEvidenceKind) {
		t.Fatalf("migrated plan still reports a rewrite: %#v", session.Plan.Evidence)
	}
	again, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil || again.TreeHashesUpgraded {
		t.Fatalf("second apply result = %#v err = %v", again, err)
	}
}

func TestLegacyTreeHashProvenanceForAChangedTreeStaysUnverified(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents, TargetClaude})
	applyFixture(t, session)
	rewriteStateAsLegacy(t, session)
	if err := os.WriteFile(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "SKILL.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	inventory, err := InspectProject(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.TreeHashMigration != 1 {
		t.Fatalf("tree hash migration = %d, want 1", inventory.TreeHashMigration)
	}
	for _, record := range inventory.State.Records {
		switch record.Target {
		case TargetAgents:
			if record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 {
				t.Fatalf("changed tree upgraded its legacy record: %#v", record)
			}
		case TargetClaude:
			if record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V3 || record.TreeHash != hash.Digest {
				t.Fatalf("unchanged tree did not upgrade: %#v", record)
			}
		}
	}
	classification, err := ClassifyProject(session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range classification.States {
		want := ProjectStateReasonVerifiedExact
		if state.Target == TargetAgents {
			want = ProjectStateReasonLocalModification
		}
		if state.Reason != want {
			t.Fatalf("%s state = %#v, want %s", state.Target, state, want)
		}
	}
}

// TestLegacyTreeHashProvenanceUpdatesThroughTheTransaction updates skills
// whose records are still tree-sha256-v2. The rewrite lands before the
// transaction, so the quarantine keeps the old copy's verified source.
func TestLegacyTreeHashProvenanceUpdatesThroughTheTransaction(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	legacyState := rewriteStateAsLegacy(t, session)
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Operations) != 1 || session.Plan.Operations[0].Reason != string(ProjectStateReasonVerifiedUpdate) {
		t.Fatalf("legacy provenance plan = %#v", session.Plan.Operations)
	}
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !result.TreeHashesUpgraded || len(result.Updated) != 1 {
		t.Fatalf("result = %#v", result)
	}
	manifest := readUpdateManifest(t, session)
	entry := manifest.Entries[0]
	if entry.TreeHashAlgorithm != TreeHashAlgorithmSHA256V3 || entry.OldTreeHash != oldHash.Digest ||
		entry.OldSourceIdentity != legacyState.Records[0].SourceIdentity {
		t.Fatalf("quarantine entry = %#v", entry)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHashAlgorithm != TreeHashAlgorithmSHA256V3 ||
		state.Records[0].TreeHash != session.Expected[skill.Name].Digest {
		t.Fatalf("updated provenance = %#v", state.Records)
	}
}

// TestRestoreAcceptsLegacyTreeHashQuarantine restores a quarantine whose
// manifest and provenance an earlier release wrote with tree-sha256-v2.
func TestRestoreAcceptsLegacyTreeHashQuarantine(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName)
	manifest := readUpdateManifest(t, session)
	entry := &manifest.Entries[0]
	quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(entry.QuarantinedPlacement))
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	oldLegacy, newLegacy := legacyTreeHash(t, quarantined), legacyTreeHash(t, destination)
	entry.TreeHashAlgorithm, entry.OldTreeHash, entry.NewTreeHash = TreeHashAlgorithmSHA256V2, oldLegacy.Digest, newLegacy.Digest
	data, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	rewriteStateAsLegacy(t, session)
	removeRestoreDestinations(t, session, skill, []Target{TargetAgents})

	if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps()); err != nil {
		t.Fatal(err)
	}
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("restored hash = %#v, want %#v", got, oldHash)
	}
	state := readApplyState(t, session)
	if len(state.Records) != 1 || state.Records[0].TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 || state.Records[0].TreeHash != oldLegacy.Digest {
		t.Fatalf("restored provenance = %#v", state.Records)
	}
	inventory, err := InspectProject(session.Layout)
	if err != nil || inventory.TreeHashMigration != 1 || inventory.State.Records[0].TreeHash != oldHash.Digest {
		t.Fatalf("restored legacy record did not verify: %#v err = %v", inventory.State.Records, err)
	}
}
