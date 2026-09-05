package sjskills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStrictRemovalRecoversAbruptApplyAndRestore(t *testing.T) {
	const rootEnv = "SJSKILLS_STRICT_CRASH_ROOT"
	const phaseEnv = "SJSKILLS_STRICT_CRASH_PHASE"
	if root := os.Getenv(rootEnv); root != "" {
		layout, err := LayoutForProject(root)
		if err != nil {
			t.Fatal(err)
		}
		deps := updateTestDeps()
		deps.afterStateCommit = func() { os.Exit(94) }
		if os.Getenv(phaseEnv) == "restore" {
			deps.newQuarantineID = nil
			_, err = RestoreProjectQuarantine(context.Background(), layout, testQuarantineID, deps)
		} else {
			session := &ProjectApplySession{Layout: layout, Desired: DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{}}, Expected: map[string]TreeHash{}}
			makeSessionPlanCurrent(t, session)
			_, err = ApplyProjectChanges(context.Background(), session, deps)
		}
		t.Fatalf("failed to reach crash point: %v", err)
	}
	for _, ownership := range []string{"unowned", "modified", "owned"} {
		for _, phase := range []string{"apply", "restore"} {
			t.Run(ownership+"/"+phase, func(t *testing.T) {
				session, _, skill, _ := removedApplyFixture(t, []Target{TargetAgents})
				destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
				if ownership == "unowned" {
					data, err := marshalApplyState(emptyTrustedProvenanceState())
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(session.Layout.ReconcilerStatePath, data, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if ownership == "modified" {
					if err := os.WriteFile(filepath.Join(destination, "local"), []byte("local edits\n"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				observed := hashSkillAt(t, destination)
				makeSessionPlanCurrent(t, session)
				if phase == "restore" {
					if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
						t.Fatal(err)
					}
				}
				beforeState, err := os.ReadFile(session.Layout.ReconcilerStatePath)
				if err != nil {
					t.Fatal(err)
				}
				child := exec.Command(os.Args[0], "-test.run=^TestStrictRemovalRecoversAbruptApplyAndRestore$")
				child.Env = append(os.Environ(), rootEnv+"="+session.Layout.Root, phaseEnv+"="+phase)
				output, err := child.CombinedOutput()
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 94 {
					t.Fatalf("crash: %v %s", err, output)
				}
				if phase == "restore" {
					_, err = RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
				} else {
					makeSessionPlanCurrent(t, session)
					_, err = ApplyProjectChanges(context.Background(), session, updateTestDeps())
				}
				if err == nil {
					t.Fatal("expected recovered transaction to request rerun")
				}
				afterState, readErr := os.ReadFile(session.Layout.ReconcilerStatePath)
				if readErr != nil || !bytes.Equal(beforeState, afterState) {
					t.Fatalf("recovery lost pre-state: %v\n%s\n%s", readErr, beforeState, afterState)
				}
				if phase == "restore" {
					if _, err := os.Lstat(destination); !os.IsNotExist(err) {
						t.Fatalf("restore rollback left active copy: %v", err)
					}
					if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{}); err != nil {
						t.Fatal(err)
					}
				}
				if got := hashSkillAt(t, destination); got != observed {
					t.Fatalf("recovery lost original bytes: %#v", got)
				}
				if phase == "restore" {
					state := readApplyState(t, session)
					wantRecords := 0
					if ownership == "owned" {
						wantRecords = 1
					}
					if len(state.Records) != wantRecords {
						t.Fatalf("restoration invented ownership: %#v", state)
					}
				}
			})
		}
	}
}

func TestRemovalJournalCannotDiscardMatchingProvenance(t *testing.T) {
	session, _, _, _ := removedApplyFixture(t, []Target{TargetAgents})
	preimage, err := captureApplyStatePreimage(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := buildApplyState(preimage.state, session, fixedApplyTime())
	if err != nil {
		t.Fatal(err)
	}
	data, err := marshalApplyState(candidate)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := newApplyJournal(preimage, session, data, testQuarantineID, testQuarantineID)
	if err != nil {
		t.Fatal(err)
	}
	journal.Entries[0].OldSourceIdentity = ""
	if validProjectTransactionJournal(journal) {
		t.Fatal("journal discarded matching ownership")
	}
}

func TestGlobalStrictRemovalRestoresModifiedCopyWithoutOwnership(t *testing.T) {
	session := newGlobalApplyFixture(t)
	if _, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{}); err != nil {
		t.Fatal(err)
	}
	old := writeGlobalSkill(t, session.Layout.AgentsSkillsPath, "project-status", "old\n")
	inventory, err := InspectGlobal(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	inventory.State.Records = append(inventory.State.Records, ProvenanceRecord{Scope: ScopeGlobal, Skill: "project-status", Target: TargetAgents, SourceIdentity: "github:example/skills", TreeHashAlgorithm: old.Algorithm, TreeHash: old.Digest, RecordedAt: fixedApplyTime()})
	data, err := json.Marshal(inventory.State)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(session.Layout.ProvenanceStatePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	observed := writeGlobalSkill(t, session.Layout.AgentsSkillsPath, "project-status", "local edits\n")
	inventory, err = InspectGlobal(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(session.Registry, session.Desired, session.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	session.Plan, err = TranslateGlobalClassification(Plan{Desired: session.Desired}, classification)
	if err != nil {
		t.Fatal(err)
	}
	bindGlobalApplyFixture(t, session)
	result, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Quarantine == nil || len(result.Quarantined) != 1 {
		t.Fatalf("result: %#v", result)
	}
	if _, err := RestoreGlobalQuarantine(context.Background(), session.Layout, result.Quarantine.ID, ApplyDeps{}); err != nil {
		t.Fatal(err)
	}
	if got := hashSkillAt(t, filepath.Join(session.Layout.AgentsSkillsPath, "project-status")); got != observed {
		t.Fatalf("local edits lost: %#v", got)
	}
	inventory, err = InspectGlobal(session.Layout)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range inventory.State.Records {
		if record.Skill == "project-status" {
			t.Fatal("modified copy acquired ownership")
		}
	}
}
