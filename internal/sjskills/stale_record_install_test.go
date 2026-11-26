package sjskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeStaleInstallRecord commits provenance for the fixture placement while
// leaving the placement absent, as a fresh checkout of a project that tracks
// .sjskills/state.json but git-ignores its target root does. The record names
// an older tree so the install must replace it rather than merely confirm it.
func writeStaleInstallRecord(t *testing.T, session *ProjectApplySession) []byte {
	t.Helper()
	stale, err := buildApplyState(emptyTrustedProvenanceState(), session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	for index := range stale.Records {
		stale.Records[index].TreeHash = strings.Repeat("a", 64)
	}
	data, err := marshalApplyState(stale)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(session.Layout.DerivedDirectoryPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return data
}

func readInstalledRecord(t *testing.T, session *ProjectApplySession) ProvenanceRecord {
	t.Helper()
	data, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	state, valid := decodeTransactionProvenanceState(data, ScopeProject, false)
	if !valid || len(state.Records) != 1 {
		t.Fatalf("installed state valid=%v state=%#v", valid, state)
	}
	return state.Records[0]
}

func TestApplyProjectChangesInstallsOverStaleRecordForAbsentPlacement(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents})
	writeStaleInstallRecord(t, session)

	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{Now: fixedApplyTime})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Installed) != 1 || result.Installed[0] != (AppliedPlacement{Skill: skill.Name, Target: TargetAgents}) {
		t.Fatalf("installed = %#v", result.Installed)
	}
	if got := hashSkillAt(t, filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)); got != hash {
		t.Fatalf("placed hash=%#v, want %#v", got, hash)
	}
	if record := readInstalledRecord(t, session); record.TreeHash != hash.Digest || record.Skill != skill.Name || record.Target != TargetAgents {
		t.Fatalf("stale record was not replaced: %#v", record)
	}
}

func TestApplyProjectChangesRecoversInterruptedInstallOverStaleRecord(t *testing.T) {
	session, _, skill, hash := newApplyFixture(t, []Target{TargetAgents})
	staleData := writeStaleInstallRecord(t, session)
	info, err := os.Stat(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	staleState, valid := decodeTransactionProvenanceState(staleData, ScopeProject, false)
	if !valid {
		t.Fatal("stale state does not decode")
	}
	preimage := &applyStatePreimage{exists: true, data: staleData, state: staleState, info: info, mode: info.Mode()}
	candidate, err := buildApplyState(staleState, session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, candidateData, testQuarantineID, "")
	if err != nil {
		t.Fatalf("journal over stale record: %v", err)
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
	// Apply commits candidate state privately; the restored pre-state keeps
	// the committed checkout's broader mode.
	if err := os.Remove(session.Layout.ReconcilerStatePath); err != nil {
		t.Fatal(err)
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
	restored, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil || !bytes.Equal(restored, staleData) {
		t.Fatalf("recovery did not restore the committed stale record: err=%v state=%s", err, restored)
	}
	if restoredInfo, err := os.Stat(session.Layout.ReconcilerStatePath); err != nil || (runtime.GOOS != "windows" && restoredInfo.Mode().Perm() != info.Mode().Perm()) {
		t.Fatalf("recovery did not restore the committed state mode: info=%v err=%v", restoredInfo, err)
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("transaction journal survived recovery: %v", err)
	}

	result, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{Now: fixedApplyTime})
	if err != nil || len(result.Installed) != 1 || result.Installed[0].Skill != skill.Name {
		t.Fatalf("post-recovery rerun result=%#v err=%v", result, err)
	}
	if record := readInstalledRecord(t, session); record.TreeHash != hash.Digest {
		t.Fatalf("rerun did not replace stale record: %#v", record)
	}
}
