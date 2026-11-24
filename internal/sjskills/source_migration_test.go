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
	"time"
)

const (
	formerFixtureIdentity  = "github:example/demo"
	renamedFixtureSource   = "renamed/demo"
	renamedFixtureIdentity = "github:renamed/demo"
)

// renameApplyFixtureSource moves the applied fixture skill to a new
// repository location, optionally declaring its installed identity former.
func renameApplyFixtureSource(t *testing.T, session *ProjectApplySession, skill DesiredSkill, declare bool) DesiredSkill {
	t.Helper()
	renamed := skill
	renamed.Source = renamedFixtureSource
	renamed.Targets = copyTargets(skill.Targets)
	if declare {
		renamed.FormerSourceIdentities = []string{formerFixtureIdentity}
	}
	session.Desired = DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{renamed}}
	session.Materialized.snapshots[skill.Name].Skill = renamed
	makeSessionPlanCurrent(t, session)
	return renamed
}

func assertOperationReasons(t *testing.T, plan Plan, action PlanAction, reason ProjectStateReason, count int) {
	t.Helper()
	if len(plan.Operations) != count {
		t.Fatalf("operations = %#v, want %d", plan.Operations, count)
	}
	for _, operation := range plan.Operations {
		if operation.Action != action || operation.Reason != string(reason) {
			t.Fatalf("operation = %#v, want %s (%s)", operation, action, reason)
		}
	}
}

func TestClassifyProjectAcceptsOnlyDeclaredFormerSourceIdentities(t *testing.T) {
	root := t.TempDir()
	oldHash := classificationHash('1')
	expectedHash := classificationHash('2')
	changedHash := classificationHash('3')
	renamed := func(name string) DesiredSkill {
		skill := classificationSkill(name, TargetAgents)
		skill.FormerSourceIdentities = []string{"github:old/repo"}
		return skill
	}
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{
		renamed("exact"), renamed("outdated"), renamed("modified"), renamed("unrelated"),
		classificationSkill("undeclared", TargetAgents),
	}}
	expected := map[string]TreeHash{}
	for _, skill := range desired.Skills {
		expected[skill.Name] = expectedHash
	}
	inventory := classificationInventory(t, root, []InventoryEntry{
		classificationEntry(root, TargetAgents, "exact", &expectedHash),
		classificationEntry(root, TargetAgents, "outdated", &oldHash),
		classificationEntry(root, TargetAgents, "modified", &changedHash),
		classificationEntry(root, TargetAgents, "unrelated", &expectedHash),
		classificationEntry(root, TargetAgents, "undeclared", &expectedHash),
	}, nil, true, []ProvenanceRecord{
		classificationRecord("exact", TargetAgents, "github:old/repo", expectedHash),
		classificationRecord("outdated", TargetAgents, "github:old/repo", oldHash),
		classificationRecord("modified", TargetAgents, "github:old/repo", oldHash),
		classificationRecord("unrelated", TargetAgents, "github:other/repo", expectedHash),
		classificationRecord("undeclared", TargetAgents, "github:old/repo", expectedHash),
	})
	classification, err := ClassifyProject(desired, expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		kind   ProjectStateKind
		action PlanAction
		reason ProjectStateReason
	}{
		{"exact", ProjectStateExact, PlanActionUnchanged, ProjectStateReasonFormerSourceIdentity},
		{"outdated", ProjectStateOutdated, PlanActionUpdate, ProjectStateReasonVerifiedUpdate},
		{"modified", ProjectStateModified, PlanActionUpdate, ProjectStateReasonLocalModification},
		{"unrelated", ProjectStateProtected, PlanActionBlocked, ProjectStateReasonProvenanceSourceMismatch},
		{"undeclared", ProjectStateProtected, PlanActionBlocked, ProjectStateReasonProvenanceSourceMismatch},
	} {
		state := findProjectState(t, classification, TargetAgents, test.name)
		if state.Kind != test.kind || state.Action != test.action || state.Reason != test.reason || state.SourceIdentity != "github:owner/repo" {
			t.Errorf("%s = %#v, want %s/%s/%s", test.name, state, test.kind, test.action, test.reason)
		}
	}
	plan, err := TranslateProjectClassification(Plan{Desired: desired}, classification)
	if err != nil {
		t.Fatal(err)
	}
	if got := operationsWithReason(plan.Operations, PlanActionUnchanged, ProjectStateReasonFormerSourceIdentity); len(got) != 1 || got[0].Skill != "exact" {
		t.Fatalf("migration operations = %#v", got)
	}

	for name, mutate := range map[string]func(*DesiredSkill){
		"current": func(skill *DesiredSkill) { skill.FormerSourceIdentities = []string{"github:owner/repo"} },
		"local":   func(skill *DesiredSkill) { skill.FormerSourceIdentities = []string{"local:./skills/exact"} },
		"duplicate": func(skill *DesiredSkill) {
			skill.FormerSourceIdentities = []string{"github:old/repo", "github:old/repo"}
		},
		"direct": func(skill *DesiredSkill) { skill.Origin = "direct" },
		"manual": func(skill *DesiredSkill) { skill.Manager = ManagerManual },
	} {
		invalid := renamed("exact")
		mutate(&invalid)
		state := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{invalid}}
		hashes := map[string]TreeHash{}
		if invalid.Manager == ManagerSkillsCLI {
			hashes[invalid.Name] = expectedHash
		}
		if _, err := ClassifyProject(state, hashes, classificationInventory(t, root, nil, nil, true, nil)); err == nil || !strings.Contains(err.Error(), "formerSourceIdentities") {
			t.Errorf("%s former identities err = %v", name, err)
		}
	}
}

func TestApplyProjectChangesMigratesExactPlacementsFromFormerSourceIdentity(t *testing.T) {
	session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents, TargetClaude})
	applyFixture(t, session)
	before := readApplyState(t, session)

	renameApplyFixtureSource(t, session, skill, false)
	assertOperationReasons(t, session.Plan, PlanActionBlocked, ProjectStateReasonProvenanceSourceMismatch, 2)

	renameApplyFixtureSource(t, session, skill, true)
	assertOperationReasons(t, session.Plan, PlanActionUnchanged, ProjectStateReasonFormerSourceIdentity, 2)
	result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil {
		t.Fatal(err)
	}
	want := []AppliedPlacement{{Skill: skill.Name, Target: TargetAgents}, {Skill: skill.Name, Target: TargetClaude}}
	if !reflect.DeepEqual(result.SourceMigrated, want) || len(result.Installed)+len(result.Updated)+len(result.Quarantined) != 0 || result.Quarantine != nil || result.Migrated {
		t.Fatalf("apply result = %#v", result)
	}
	after := readApplyState(t, session)
	if len(after.Records) != len(before.Records) {
		t.Fatalf("records = %#v, before %#v", after.Records, before.Records)
	}
	for index, record := range after.Records {
		previous := before.Records[index]
		if previous.SourceIdentity != formerFixtureIdentity {
			t.Fatalf("fixture record = %#v", previous)
		}
		previous.SourceIdentity = renamedFixtureIdentity
		if record != previous {
			t.Fatalf("migrated record = %#v, want %#v", record, previous)
		}
	}
	if _, err := os.Lstat(projectTransactionJournalPath(session.Layout)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("migration left a journal: %v", err)
	}

	makeSessionPlanCurrent(t, session)
	assertOperationReasons(t, session.Plan, PlanActionUnchanged, ProjectStateReasonVerifiedExact, 2)
	second, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
	if err != nil || second.SourceMigrated != nil {
		t.Fatalf("idempotent result = %#v err=%v", second, err)
	}
}

func TestApplyProjectChangesUpdatesFromFormerSourceIdentity(t *testing.T) {
	t.Run("verified-update", func(t *testing.T) {
		session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
		applyFixture(t, session)
		oldHash := hashPlacedSkill(t, session, skill.Name, TargetAgents)
		if err := os.WriteFile(filepath.Join(session.Materialized.snapshots[skill.Name].Path, "SKILL.md"), []byte("updated\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		newHash := rehashApplySnapshot(t, session, skill.Name)
		renameApplyFixtureSource(t, session, skill, true)
		assertOperationReasons(t, session.Plan, PlanActionUpdate, ProjectStateReasonVerifiedUpdate, 1)

		result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
		if err != nil || len(result.Updated) != 1 || result.SourceMigrated != nil {
			t.Fatalf("apply result = %#v err=%v", result, err)
		}
		// The verified preimage is quarantined under the migrated identity, so
		// the manifest keeps its same-source update invariant.
		entry := readUpdateManifest(t, session).Entries[0]
		if entry.OldSourceIdentity != renamedFixtureIdentity || entry.NewSourceIdentity != renamedFixtureIdentity || entry.OldTreeHash != oldHash.Digest || entry.NewTreeHash != newHash.Digest {
			t.Fatalf("manifest entry = %#v", entry)
		}
		state := readApplyState(t, session)
		if len(state.Records) != 1 || state.Records[0].SourceIdentity != renamedFixtureIdentity || state.Records[0].TreeHash != newHash.Digest {
			t.Fatalf("state = %#v", state)
		}

		removeRestoreDestinations(t, session, skill, []Target{TargetAgents})
		if _, err := RestoreProjectQuarantine(context.Background(), session.Layout, testQuarantineID, updateTestDeps()); err != nil {
			t.Fatal(err)
		}
		restored := readApplyState(t, session)
		if len(restored.Records) != 1 || restored.Records[0].SourceIdentity != renamedFixtureIdentity || restored.Records[0].TreeHash != oldHash.Digest {
			t.Fatalf("restored state = %#v", restored)
		}
		makeSessionPlanCurrent(t, session)
		assertOperationReasons(t, session.Plan, PlanActionUpdate, ProjectStateReasonVerifiedUpdate, 1)
	})

	t.Run("local-modification", func(t *testing.T) {
		session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
		applyFixture(t, session)
		if err := os.WriteFile(filepath.Join(session.Layout.AgentsSkillsPath, skill.Name, "local-note"), []byte("keep\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		modifiedHash := hashPlacedSkill(t, session, skill.Name, TargetAgents)
		renameApplyFixtureSource(t, session, skill, true)
		assertOperationReasons(t, session.Plan, PlanActionUpdate, ProjectStateReasonLocalModification, 1)

		result, err := ApplyProjectChanges(context.Background(), session, updateTestDeps())
		if err != nil || len(result.Updated) != 1 {
			t.Fatalf("apply result = %#v err=%v", result, err)
		}
		entry := readUpdateManifest(t, session).Entries[0]
		if entry.OldSourceIdentity != "" || entry.OldTreeHash != modifiedHash.Digest || entry.NewSourceIdentity != renamedFixtureIdentity {
			t.Fatalf("modified copy acquired ownership in quarantine: %#v", entry)
		}
		state := readApplyState(t, session)
		if len(state.Records) != 1 || state.Records[0].SourceIdentity != renamedFixtureIdentity || state.Records[0].TreeHash != session.Expected[skill.Name].Digest {
			t.Fatalf("state = %#v", state)
		}
	})
}

func TestSourceMigrationRecoversAfterAbruptProcessExit(t *testing.T) {
	const childEnv = "SJSKILLS_SOURCE_MIGRATION_CRASH"
	const rootEnv = "SJSKILLS_SOURCE_MIGRATION_ROOT"
	const stageEnv = "SJSKILLS_SOURCE_MIGRATION_STAGE"
	if phase := os.Getenv(childEnv); phase != "" {
		session, _, skill, _ := newApplyFixtureAt(t, os.Getenv(rootEnv), os.Getenv(stageEnv), []Target{TargetAgents})
		renameApplyFixtureSource(t, session, skill, true)
		deps := updateTestDeps()
		switch phase {
		case "before-commit":
			deps.beforeCommit = func() error { os.Exit(91); return nil }
		case "after-commit":
			deps.afterStateCommit = func() { os.Exit(91) }
		default:
			t.Fatalf("unknown child phase %q", phase)
		}
		_, _ = ApplyProjectChanges(context.Background(), session, deps)
		os.Exit(92)
	}
	for _, phase := range []string{"before-commit", "after-commit"} {
		t.Run(phase, func(t *testing.T) {
			session, _, skill, _ := newApplyFixture(t, []Target{TargetAgents})
			applyFixture(t, session)
			preState, err := os.ReadFile(session.Layout.ReconcilerStatePath)
			if err != nil {
				t.Fatal(err)
			}
			runModifiedRecoveryChild(t, "TestSourceMigrationRecoversAfterAbruptProcessExit", childEnv+"="+phase,
				rootEnv+"="+session.Layout.Root, stageEnv+"="+session.Materialized.root)
			data, err := os.ReadFile(projectTransactionJournalPath(session.Layout))
			if err != nil {
				t.Fatalf("interrupted migration left no journal: %v", err)
			}
			journal, ok := decodeProjectTransactionJournal(data)
			if !ok || len(journal.Entries) != 0 || len(journal.SourceMigrations) != 1 || journal.SourceMigrations[0].OldSourceIdentity != formerFixtureIdentity {
				t.Fatalf("interrupted journal = %#v valid=%v", journal, ok)
			}

			renameApplyFixtureSource(t, session, skill, true)
			if _, err := ApplyProjectChanges(context.Background(), session, updateTestDeps()); err == nil || err.Error() != "project apply unavailable: interrupted project transaction was recovered; rerun apply" {
				t.Fatalf("recovery error = %v", err)
			}
			if got, err := os.ReadFile(session.Layout.ReconcilerStatePath); err != nil || !bytes.Equal(got, preState) {
				t.Fatalf("recovery did not restore the recorded provenance: %q err=%v", got, err)
			}
			makeSessionPlanCurrent(t, session)
			deps := updateTestDeps()
			deps.newQuarantineID = func() (string, error) { return "abcdef0123456789abcdef0123456789", nil }
			result, err := ApplyProjectChanges(context.Background(), session, deps)
			if err != nil || len(result.SourceMigrated) != 1 {
				t.Fatalf("rerun result = %#v err=%v", result, err)
			}
			if state := readApplyState(t, session); state.Records[0].SourceIdentity != renamedFixtureIdentity {
				t.Fatalf("rerun state = %#v", state)
			}
		})
	}
}

func TestJournalSourceMigrationsMustDescribeTheStateTransition(t *testing.T) {
	hash := classificationHash('1')
	record := classificationRecord("demo", TargetAgents, formerFixtureIdentity, hash)
	record.RecordedAt = fixedApplyTime()
	encode := func(records ...ProvenanceRecord) []byte {
		data, err := marshalApplyState(ProvenanceState{Version: ProvenanceStateVersion, Records: records})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	migrated := record
	migrated.SourceIdentity = renamedFixtureIdentity
	other := classificationRecord("other", TargetClaude, formerFixtureIdentity, hash)
	other.RecordedAt = fixedApplyTime()
	otherMigrated := other
	otherMigrated.SourceIdentity = renamedFixtureIdentity
	valid := func() projectTransactionJournal {
		return projectTransactionJournal{
			Version: projectTransactionJournalVersion, ID: testQuarantineID, Kind: projectTransactionKindApply,
			PreState:       projectJournalState{Exists: true, Mode: 0o600, Data: encode(record, other)},
			CandidateState: encode(migrated, otherMigrated),
			Entries:        []projectJournalEntry{},
			SourceMigrations: []sourceIdentityMigration{
				{Skill: "demo", Target: TargetAgents, OldSourceIdentity: formerFixtureIdentity, NewSourceIdentity: renamedFixtureIdentity},
				{Skill: "other", Target: TargetClaude, OldSourceIdentity: formerFixtureIdentity, NewSourceIdentity: renamedFixtureIdentity},
			},
		}
	}
	if !validProjectTransactionJournal(valid()) {
		t.Fatal("valid migration journal rejected")
	}
	for name, mutate := range map[string]func(*projectTransactionJournal){
		"undeclared": func(j *projectTransactionJournal) { j.SourceMigrations = j.SourceMigrations[:1] },
		"none":       func(j *projectTransactionJournal) { j.SourceMigrations = nil },
		"empty":      func(j *projectTransactionJournal) { j.SourceMigrations = []sourceIdentityMigration{} },
		"wrong-old":  func(j *projectTransactionJournal) { j.SourceMigrations[0].OldSourceIdentity = "github:other/demo" },
		"unchanged":  func(j *projectTransactionJournal) { j.SourceMigrations[0].NewSourceIdentity = formerFixtureIdentity },
		"local":      func(j *projectTransactionJournal) { j.SourceMigrations[0].NewSourceIdentity = "local:./demo" },
		"unsorted": func(j *projectTransactionJournal) {
			j.SourceMigrations[0], j.SourceMigrations[1] = j.SourceMigrations[1], j.SourceMigrations[0]
		},
		"duplicate": func(j *projectTransactionJournal) { j.SourceMigrations[1] = j.SourceMigrations[0] },
		"absent":    func(j *projectTransactionJournal) { j.SourceMigrations[1].Skill = "absent" },
		"wrong-new": func(j *projectTransactionJournal) { j.CandidateState = encode(migrated, other) },
		"hash-changed": func(j *projectTransactionJournal) {
			changed := migrated
			changed.TreeHash = classificationHash('2').Digest
			j.CandidateState = encode(changed, otherMigrated)
		},
	} {
		journal := valid()
		mutate(&journal)
		if validProjectTransactionJournal(journal) {
			t.Errorf("%s: invalid migration journal accepted", name)
		}
	}
	restore := valid()
	restore.Kind = projectTransactionKindRestore
	if validJournalSourceMigrations(restore) {
		t.Error("restore journal accepted source migrations")
	}
}

func TestRegistryFormerIdentitiesAreExplicitAndUnambiguous(t *testing.T) {
	withKicpa := func(former ...string) Registry {
		registry := fixtureRegistry(t)
		source := registry.Sources["kicpa"]
		source.Location = "https://github.com/cpaikr/kicpa/tree/main/skills"
		source.FormerIdentities = former
		registry.Sources["kicpa"] = source
		return registry
	}
	registry := withKicpa("github:sjunepark/kicpa")
	if err := ValidateRegistry(registry); err != nil {
		t.Fatal(err)
	}
	desired, err := ResolveProject(registry, Manifest{Version: ManifestVersion, Profiles: []string{"kicpa", "kicpa-private"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sawKicpa := false
	for _, skill := range desired.Skills {
		if skill.SourceID == "kicpa" {
			sawKicpa = true
			if !reflect.DeepEqual(skill.FormerSourceIdentities, []string{"github:sjunepark/kicpa"}) {
				t.Fatalf("%s former identities = %#v", skill.Name, skill.FormerSourceIdentities)
			}
		} else if skill.FormerSourceIdentities != nil {
			t.Fatalf("%s gained former identities %#v", skill.Name, skill.FormerSourceIdentities)
		}
	}
	if !sawKicpa {
		t.Fatal("kicpa profile resolved no kicpa-source skill")
	}

	for name, test := range map[string]struct {
		registry Registry
		code     IssueCode
	}{
		"empty":            {withKicpa([]string{}...), IssueEmptySelection},
		"current":          {withKicpa("github:cpaikr/kicpa"), IssueInvalidSource},
		"not-canonical":    {withKicpa("sjunepark/kicpa"), IssueInvalidSource},
		"uppercase":        {withKicpa("github:SJunePark/kicpa"), IssueInvalidSource},
		"local":            {withKicpa("local:./kicpa"), IssueInvalidSource},
		"duplicate":        {withKicpa("github:sjunepark/kicpa", "github:sjunepark/kicpa"), IssueDuplicate},
		"unsorted":         {withKicpa("github:sjunepark/kicpa", "github:a/kicpa"), IssueMalformedInput},
		"live-identity":    {withKicpa("github:cpaikr/darty"), IssueCollision},
		"locationless":     {withFormer(withKicpa(), "context7-mcp-manual", "github:old/context7"), IssueInvalidSource},
		"claimed-by-other": {withFormer(withKicpa("github:sjunepark/kicpa"), "darty", "github:sjunepark/kicpa"), IssueCollision},
	} {
		if err := ValidateRegistry(test.registry); !issueCode(err, test.code) {
			t.Errorf("%s: err = %v, want %s", name, err, test.code)
		}
	}
}

func withFormer(registry Registry, sourceID string, former ...string) Registry {
	source := registry.Sources[sourceID]
	source.FormerIdentities = former
	registry.Sources[sourceID] = source
	return registry
}

func TestStatusReportsPendingSourceIdentityMigrationAsUpdate(t *testing.T) {
	findings := summarizeStatus(Plan{Operations: []PlanOperation{
		{Action: PlanActionUnchanged, Skill: "migrated", Target: TargetAgents, Reason: string(ProjectStateReasonFormerSourceIdentity)},
		{Action: PlanActionUnchanged, Skill: "exact", Target: TargetAgents, Reason: string(ProjectStateReasonVerifiedExact)},
	}})
	want := []AdvisoryFinding{{AdvisoryUpdate, "migrated", TargetAgents, string(ProjectStateReasonFormerSourceIdentity)}}
	if !reflect.DeepEqual(findings, want) {
		t.Fatalf("findings = %#v, want %#v", findings, want)
	}
}

func TestApplyGlobalChangesMigratesFormerSourceIdentity(t *testing.T) {
	session := newGlobalApplyFixture(t)
	if _, err := ApplyGlobalChanges(context.Background(), session, ApplyDeps{Now: fixedApplyTime}); err != nil {
		t.Fatal(err)
	}
	registry := minimalGlobalRegistry(t)
	registry.Sources["fixture"] = Source{Kind: SourceExternal, Location: "renamed/skills", FormerIdentities: []string{"github:example/skills"}}
	replanned := replanGlobalApplyFixture(t, session, registry)
	assertOperationReasons(t, replanned.Plan, PlanActionUnchanged, ProjectStateReasonFormerSourceIdentity, 2)
	later := func() time.Time { return fixedApplyTime().Add(time.Hour) }
	result, err := ApplyGlobalChanges(context.Background(), replanned, ApplyDeps{Now: later})
	if err != nil || len(result.SourceMigrated) != 2 || len(result.Installed)+len(result.Updated) != 0 || result.Migrated {
		t.Fatalf("result = %#v err=%v", result, err)
	}
	inventory, err := InspectGlobal(replanned.Layout)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range inventory.State.Records {
		if record.SourceIdentity != "github:renamed/skills" || record.RecordedAt != fixedApplyTime() {
			t.Fatalf("global record = %#v", record)
		}
	}
	assertOperationReasons(t, replanGlobalApplyFixture(t, replanned, registry).Plan, PlanActionUnchanged, ProjectStateReasonVerifiedExact, 2)
}

func replanGlobalApplyFixture(t *testing.T, previous *GlobalApplySession, registry Registry) *GlobalApplySession {
	t.Helper()
	desired, err := ResolveGlobal(registry)
	if err != nil {
		t.Fatal(err)
	}
	previous.Materialized.snapshots["base"].Skill = desired.Skills[0]
	inventory, err := InspectGlobal(previous.Layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(registry, desired, previous.Expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := TranslateGlobalClassification(Plan{Desired: desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	session := &GlobalApplySession{Layout: previous.Layout, Registry: registry, Desired: desired, Plan: plan,
		Expected: previous.Expected, Materialized: previous.Materialized}
	bindGlobalApplyFixture(t, session)
	return session
}
