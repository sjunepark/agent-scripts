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
	session, _, skill, _ := newApplyFixture(t, targets)
	// An executable script is the case tree-sha256-v2 hashed differently by
	// host; on macOS and Linux its legacy record carries "executable".
	script := filepath.Join(session.Materialized.snapshots[skill.Name].Path, "run.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho demo\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	hash := rehashApplySnapshot(t, session, skill.Name)
	makeSessionPlanCurrent(t, session)
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

// TestLegacyTreeHashProvenanceRemovesWithVerifiedSource quarantines a skill
// that is no longer desired while its record is still tree-sha256-v2.
func TestLegacyTreeHashProvenanceRemovesWithVerifiedSource(t *testing.T) {
	session, _, _, oldHash := removedApplyFixture(t, []Target{TargetAgents})
	legacyState := rewriteStateAsLegacy(t, session)
	makeSessionPlanCurrent(t, session)
	if len(session.Plan.Operations) != 1 || session.Plan.Operations[0].Action != PlanActionQuarantine ||
		session.Plan.Operations[0].Source != legacyState.Records[0].SourceIdentity {
		t.Fatalf("legacy removal plan = %#v", session.Plan.Operations)
	}
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	if !result.TreeHashesUpgraded || len(result.Quarantined) != 1 {
		t.Fatalf("result = %#v", result)
	}
	entry := readUpdateManifest(t, session).Entries[0]
	if entry.TreeHashAlgorithm != TreeHashAlgorithmSHA256V3 || entry.OldTreeHash != oldHash.Digest ||
		entry.OldSourceIdentity != legacyState.Records[0].SourceIdentity {
		t.Fatalf("quarantine entry = %#v", entry)
	}
	if state := readApplyState(t, session); len(state.Records) != 0 {
		t.Fatalf("removed provenance survived: %#v", state.Records)
	}
}

// TestRecoveryAcceptsLegacyTreeHashJournal recovers an update that a release
// before tree-sha256-v3 interrupted after replacing the placement: its state,
// journal, and quarantine manifest all carry tree-sha256-v2 digests.
func TestRecoveryAcceptsLegacyTreeHashJournal(t *testing.T) {
	session, _, skill, oldHash := updateApplyFixture(t, []Target{TargetAgents})
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	snapshot := session.Materialized.snapshots[skill.Name].Path
	oldLegacy, newLegacy := legacyTreeHash(t, destination), legacyTreeHash(t, snapshot)
	rewriteStateAsLegacy(t, session)
	preimage, err := captureApplyStatePreimage(session.Layout)
	if err != nil || !preimage.exists {
		t.Fatalf("capture preimage=%#v err=%v", preimage, err)
	}
	record := preimage.state.Records[0]
	recordedAt := fixedApplyTime()
	candidate := preimage.state
	candidate.Records = []ProvenanceRecord{record}
	candidate.Records[0].TreeHash, candidate.Records[0].RecordedAt = newLegacy.Digest, recordedAt
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal := projectTransactionJournal{
		Version: projectTransactionJournalVersion, ID: testQuarantineID, Kind: projectTransactionKindApply,
		QuarantineID: testQuarantineID, CandidateState: candidateData,
		PreState: projectJournalState{Exists: true, Mode: uint32(preimage.mode.Perm()), Data: preimage.data},
		Entries: []projectJournalEntry{{
			Action: PlanActionUpdate, Skill: skill.Name, Target: TargetAgents,
			ManagedParentExisted: true, ManagedRootExisted: true,
			OldSourceIdentity: record.SourceIdentity, NewSourceIdentity: record.SourceIdentity,
			TreeHashAlgorithm: TreeHashAlgorithmSHA256V2, OldTreeHash: oldLegacy.Digest, NewTreeHash: newLegacy.Digest,
		}},
	}
	if !validProjectTransactionJournal(journal) {
		t.Fatalf("legacy journal fixture is invalid: %#v", journal)
	}
	journalData, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := copyApplyTree(snapshot, destination, func(file *os.File) error { return file.Sync() }); err != nil {
		t.Fatal(err)
	}
	manifest := ProjectQuarantineManifest{
		Version: ProjectQuarantineManifestVersion, ID: testQuarantineID, CreatedAt: recordedAt,
		Status: ProjectQuarantineActive,
		Entries: []ProjectQuarantineManifestEntry{{
			Action: ProjectQuarantineEntryActionUpdate, Skill: skill.Name, Target: TargetAgents,
			OriginalPlacement:    projectOriginalPlacement(TargetAgents, skill.Name),
			QuarantinedPlacement: projectQuarantinedPlacement(TargetAgents, skill.Name),
			OldSourceIdentity:    record.SourceIdentity, NewSourceIdentity: record.SourceIdentity,
			TreeHashAlgorithm: TreeHashAlgorithmSHA256V2, OldTreeHash: oldLegacy.Digest, NewTreeHash: newLegacy.Digest,
			Status: ProjectQuarantineEntryReplaced,
		}},
	}
	manifestData, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string][]byte{
		filepath.Join(session.Layout.QuarantinePath, testQuarantineID, applyManifestName): manifestData,
		projectTransactionJournalPath(session.Layout):                                     journalData,
		session.Layout.ReconcilerStatePath:                                                candidateData,
	} {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
		t.Fatalf("legacy recovery error = %v", err)
	}
	if got := hashSkillAt(t, destination); got != oldHash {
		t.Fatalf("recovered hash = %#v, want %#v", got, oldHash)
	}
	if state := readApplyState(t, session); len(state.Records) != 1 || state.Records[0] != record {
		t.Fatalf("recovered provenance = %#v, want %#v", state.Records, record)
	}
	makeSessionPlanCurrent(t, session)
	if !planHasEvidence(session.Plan, treeHashMigrationEvidenceKind) || session.Plan.Operations[0].Reason != string(ProjectStateReasonVerifiedUpdate) {
		t.Fatalf("post-recovery plan = %#v", session.Plan)
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
