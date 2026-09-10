package sjskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestModifiedUpdateRestoresActualPreimageOnPublicationFailure(t *testing.T) {
	for _, equal := range []bool{false, true} {
		t.Run(modifiedRecoveryCaseName(equal), func(t *testing.T) {
			session, skill, oldHash := modifiedRecoveryFixture(t, equal)
			preState := readModifiedRecoveryState(t, session)
			destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
			failedPublication := false
			deps := updateTestDeps()
			deps.PublishNoReplace = func(source, target string) error {
				if target == destination && !failedPublication {
					failedPublication = true
					return errors.New("injected modified replacement publication failure")
				}
				return publishNoReplace(source, target)
			}
			result, err := ApplyProjectChanges(context.Background(), session, deps)
			if err == nil || !failedPublication {
				t.Fatalf("did not reach failed publication: result=%#v err=%v", result, err)
			}
			if result.Quarantine == nil || result.Quarantine.Status != ProjectQuarantineRolledBack {
				t.Fatalf("rollback quarantine = %#v", result.Quarantine)
			}
			assertModifiedRecoveryPreimage(t, session, skill, oldHash, preState)
			manifest := readUpdateManifest(t, session)
			if manifest.Entries[0].OldSourceIdentity != "" || manifest.Entries[0].OldTreeHash != oldHash.Digest {
				t.Fatalf("modified backup acquired ownership or lost its hash: %#v", manifest)
			}
		})
	}
}

func TestModifiedUpdateRejectsEditsAfterReview(t *testing.T) {
	for _, phase := range []string{"before-apply", "before-quarantine"} {
		t.Run(phase, func(t *testing.T) {
			session, skill, _ := modifiedRecoveryFixture(t, false)
			preState := readModifiedRecoveryState(t, session)
			destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
			raced := false
			edit := func(AppliedPlacement) error {
				raced = true
				return os.WriteFile(filepath.Join(destination, "later-edit"), []byte("preserve this newer edit\n"), 0o644)
			}
			deps := updateTestDeps()
			if phase == "before-apply" {
				if err := edit(AppliedPlacement{}); err != nil {
					t.Fatal(err)
				}
			} else {
				deps.beforeQuarantine = edit
			}
			if _, err := ApplyProjectChanges(context.Background(), session, deps); err == nil || !raced {
				t.Fatalf("reviewed content race was not exercised and rejected: raced=%v err=%v", raced, err)
			}
			if data, err := os.ReadFile(filepath.Join(destination, "later-edit")); err != nil || string(data) != "preserve this newer edit\n" {
				t.Fatalf("raced bytes changed: %q err=%v", data, err)
			}
			if got := readModifiedRecoveryState(t, session); !bytes.Equal(got, preState) {
				t.Fatal("reviewed content race changed provenance")
			}
		})
	}
}

func TestModifiedUpdateRecoversAfterAbruptProcessExit(t *testing.T) {
	const childEnv = "SJSKILLS_MODIFIED_APPLY_CRASH"
	const rootEnv = "SJSKILLS_MODIFIED_APPLY_ROOT"
	const stageEnv = "SJSKILLS_MODIFIED_APPLY_STAGE"
	if phase := os.Getenv(childEnv); phase != "" {
		session, _, skill, _ := newApplyFixtureAt(t, os.Getenv(rootEnv), os.Getenv(stageEnv), []Target{TargetAgents})
		setModifiedRecoverySnapshot(t, session, skill)
		makeSessionPlanCurrent(t, session)
		deps := updateTestDeps()
		switch phase {
		case "before-quarantine":
			deps.afterQuarantineRunSync = func() { os.Exit(91) }
		case "after-quarantine":
			deps.BeforePublish = func(AppliedPlacement) error { os.Exit(91); return nil }
		case "after-publication":
			deps.afterStateCommit = func() { os.Exit(91) }
		default:
			t.Fatalf("unknown child phase %q", phase)
		}
		_, _ = ApplyProjectChanges(context.Background(), session, deps)
		os.Exit(92)
	}
	for _, equal := range []bool{false, true} {
		for _, phase := range []string{"before-quarantine", "after-quarantine", "after-publication"} {
			t.Run(modifiedRecoveryCaseName(equal)+"/"+phase, func(t *testing.T) {
				session, skill, oldHash := modifiedRecoveryFixture(t, equal)
				preState := readModifiedRecoveryState(t, session)
				runModifiedRecoveryChild(t, "TestModifiedUpdateRecoversAfterAbruptProcessExit", childEnv+"="+phase,
					rootEnv+"="+session.Layout.Root, stageEnv+"="+session.Materialized.root)
				if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
					t.Fatalf("abrupt modified update recovery error=%v", err)
				}
				assertModifiedRecoveryPreimage(t, session, skill, oldHash, preState)
				if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("recovered journal survived: %v", err)
				}
				makeSessionPlanCurrent(t, session)
				deps := updateTestDeps()
				deps.newQuarantineID = func() (string, error) { return "abcdef0123456789abcdef0123456789", nil }
				result, err := ApplyProjectChanges(context.Background(), session, deps)
				if err != nil || len(result.Updated) != 1 {
					t.Fatalf("rerun modified update result=%#v err=%v", result, err)
				}
				if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != session.Expected[skill.Name] {
					t.Fatalf("rerun did not install expected content: %#v", got)
				}
			})
		}
	}
}

func TestModifiedRestoreRecoversAfterAbruptProcessExit(t *testing.T) {
	const childEnv = "SJSKILLS_MODIFIED_RESTORE_CRASH"
	const rootEnv = "SJSKILLS_MODIFIED_RESTORE_ROOT"
	if os.Getenv(childEnv) == "1" {
		layout, err := LayoutForProject(os.Getenv(rootEnv))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = RestoreProjectQuarantine(context.Background(), layout, testQuarantineID, ApplyDeps{afterStateCommit: func() { os.Exit(91) }})
		os.Exit(92)
	}
	for _, equal := range []bool{false, true} {
		t.Run(modifiedRecoveryCaseName(equal), func(t *testing.T) {
			session, skill, oldHash := modifiedRecoveryFixture(t, equal)
			if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err != nil {
				t.Fatal(err)
			}
			preState := readModifiedRecoveryState(t, session)
			removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
			runModifiedRecoveryChild(t, "TestModifiedRestoreRecoversAfterAbruptProcessExit", childEnv+"=1", rootEnv+"="+session.Layout.Root)
			if got := readApplyState(t, session); len(got.Records) != 0 {
				t.Fatalf("interrupted restore adopted modified bytes: %#v", got)
			}
			if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{}); err == nil || err.Error() != "project restore unavailable: interrupted project transaction was recovered; rerun restore" {
				t.Fatalf("abrupt modified restore recovery error=%v", err)
			}
			destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
			if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("restore recovery did not remove its publication: %v", err)
			}
			if got := readModifiedRecoveryState(t, session); !bytes.Equal(got, preState) {
				t.Fatal("restore recovery failed to restore exact replacement provenance")
			}
			quarantined := filepath.Join(session.Layout.QuarantinePath, testQuarantineID, filepath.FromSlash(projectQuarantinedPlacement(TargetAgents, skill.Name)))
			if got := hashSkillAt(t, quarantined); got != oldHash {
				t.Fatalf("restore recovery lost local bytes: %#v want %#v", got, oldHash)
			}
			for range 2 {
				result, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, ApplyDeps{})
				if err != nil || result.Status != ProjectQuarantineRestored {
					t.Fatalf("post-recovery restore result=%#v err=%v", result, err)
				}
				if got := hashSkillAt(t, destination); got != oldHash {
					t.Fatalf("restored local bytes=%#v want %#v", got, oldHash)
				}
				if got := readApplyState(t, session); len(got.Records) != 0 {
					t.Fatalf("restored modified copy acquired ownership: %#v", got)
				}
			}
		})
	}
}

func TestModifiedUpdateJournalRejectsInventedOwnership(t *testing.T) {
	for _, equal := range []bool{false, true} {
		t.Run(modifiedRecoveryCaseName(equal), func(t *testing.T) {
			session, _, _ := modifiedRecoveryFixture(t, equal)
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
			if journal.Entries[0].OldSourceIdentity != "" {
				t.Fatalf("modified journal claims old ownership: %#v", journal.Entries[0])
			}
			for _, mutation := range []string{"invent-old-owner", "missing-pre-record", "different-pre-source"} {
				t.Run(mutation, func(t *testing.T) {
					changed := journal
					changed.Entries = append([]projectJournalEntry(nil), journal.Entries...)
					state := preimage.state
					state.Records = append([]ProvenanceRecord(nil), preimage.state.Records...)
					switch mutation {
					case "invent-old-owner":
						changed.Entries[0].OldSourceIdentity = changed.Entries[0].NewSourceIdentity
					case "missing-pre-record":
						state.Records = nil
					case "different-pre-source":
						state.Records[0].SourceIdentity = "github:other/repository"
					}
					if mutation != "invent-old-owner" {
						changed.PreState.Data, err = marshalApplyState(state)
						if err != nil {
							t.Fatal(err)
						}
					}
					if validProjectTransactionJournal(changed) {
						t.Fatal("journal accepted fabricated ownership evidence")
					}
				})
			}
		})
	}
}

func modifiedRecoveryFixture(t *testing.T, equal bool) (*ProjectApplySession, DesiredSkill, TreeHash) {
	t.Helper()
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
	applyFixture(t, session)
	setModifiedRecoverySnapshot(t, session, skill)
	content := "locally edited\n"
	if equal {
		content = "published\n"
	}
	destination := filepath.Join(session.Layout.AgentsSkillsPath, skill.Name)
	if err := os.WriteFile(filepath.Join(destination, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	oldHash := hashSkillAt(t, destination)
	makeSessionPlanCurrent(t, session)
	return session, skill, oldHash
}

func setModifiedRecoverySnapshot(t *testing.T, session *ProjectApplySession, skill DesiredSkill) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "SKILL.md"), []byte("published\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rehashApplySnapshot(t, session, skill.Name)
}

func modifiedRecoveryCaseName(equal bool) string {
	if equal {
		return "equal-hashes"
	}
	return "different-hashes"
}

func readModifiedRecoveryState(t *testing.T, session *ProjectApplySession) []byte {
	t.Helper()
	data, err := os.ReadFile(session.Layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertModifiedRecoveryPreimage(t *testing.T, session *ProjectApplySession, skill DesiredSkill, oldHash TreeHash, preState []byte) {
	t.Helper()
	if got := hashPlacedSkill(t, session, skill.Name, TargetAgents); got != oldHash {
		t.Fatalf("recovery did not preserve actual local bytes: %#v want %#v", got, oldHash)
	}
	if got := readModifiedRecoveryState(t, session); !bytes.Equal(got, preState) {
		t.Fatal("recovery did not preserve exact historical provenance")
	}
}

func runModifiedRecoveryChild(t *testing.T, testName string, environment ...string) {
	t.Helper()
	child := exec.Command(os.Args[0], "-test.run=^"+testName+"$")
	child.Env = append(os.Environ(), environment...)
	output, err := child.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 91 {
		t.Fatalf("abrupt recovery child err=%v output=%q", err, output)
	}
}
