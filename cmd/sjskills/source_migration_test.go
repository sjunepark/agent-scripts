package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

func sourceMigrationPlan() sjskills.Plan {
	migrated := string(sjskills.ProjectStateReasonFormerSourceIdentity)
	return sjskills.Plan{Operations: []sjskills.PlanOperation{
		{Action: sjskills.PlanActionUnchanged, Skill: "cagr-research", Target: sjskills.TargetAgents, Reason: migrated},
		{Action: sjskills.PlanActionUnchanged, Skill: "cagr-research", Target: sjskills.TargetClaude, Reason: migrated},
		{Action: sjskills.PlanActionUnchanged, Skill: "kasb", Target: sjskills.TargetAgents, Reason: string(sjskills.ProjectStateReasonVerifiedExact)},
		{Action: sjskills.PlanActionUpdate, Skill: "excel", Target: sjskills.TargetAgents, Reason: string(sjskills.ProjectStateReasonLocalModification)},
	}}
}

func TestPlanRendersAndCountsSourceIdentityMigrations(t *testing.T) {
	plan := sourceMigrationPlan()
	if got := planSourceMigrationCount(plan); got != 2 {
		t.Fatalf("source migration count = %d, want 2", got)
	}
	var output bytes.Buffer
	renderPlanOperations(&output, plan.Operations)
	want := "unchanged: .agents/skills/cagr-research (former-source-identity)\n" +
		"unchanged: .claude/skills/cagr-research (former-source-identity)\n" +
		"update: .agents/skills/excel (local-modification)\n"
	if output.String() != want {
		t.Fatalf("rendered plan = %q, want %q", output.String(), want)
	}
}

func TestConfirmApplyIncludesSourceIdentityMigrations(t *testing.T) {
	for _, test := range []struct {
		global     bool
		updates    int
		migrations int
		provenance bool
		want       string
	}{
		{false, 0, 1, false, "Apply 1 project change (1 source identity migration)? [y/N] "},
		{false, 1, 16, false, "Apply 17 project changes (1 update, 16 source identity migrations)? [y/N] "},
		{true, 0, 2, true, "Apply 3 global changes (2 source identity migrations, provenance migration)? [y/N] "},
	} {
		var output bytes.Buffer
		confirmed, err := confirmApply(strings.NewReader("n\n"), &output, test.global, 0, test.updates, 0, test.migrations, test.provenance)
		if err != nil || confirmed || output.String() != test.want {
			t.Errorf("confirmed=%v err=%v output=%q want=%q", confirmed, err, output.String(), test.want)
		}
	}
}

func TestApplyExecutionEvidenceReportsSourceIdentityMigrations(t *testing.T) {
	result := sjskills.ApplyResult{SourceMigrated: []sjskills.AppliedPlacement{{Skill: "cagr-research", Target: sjskills.TargetAgents}}}
	evidence := applyExecutionEvidence(result, nil)
	if last := evidence[len(evidence)-1]; last.Kind != "migration" || last.Detail != "re-recorded 1 unchanged project placements under their source's current identity" {
		t.Fatalf("evidence = %#v", evidence)
	}
}
