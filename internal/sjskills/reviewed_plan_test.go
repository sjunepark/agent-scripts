package sjskills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReviewedPlanBindsDigestAndContract(t *testing.T) {
	envelope := reviewedPlanFixture()
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	digest := sha256.Sum256(data)
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	reviewed, err := LoadReviewedPlan(path, strings.ToUpper(hex.EncodeToString(digest[:])))
	if err != nil || reviewed.SHA256() != hex.EncodeToString(digest[:]) || reviewed.envelope.Plan == nil {
		t.Fatalf("reviewed=%#v err=%v", reviewed, err)
	}

	tampered := append([]byte(nil), data...)
	tampered[len(tampered)-2] = ' '
	if err := os.WriteFile(path, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadReviewedPlan(path, hex.EncodeToString(digest[:])); reviewedIssueCode(err) != IssueReconciliationConflict {
		t.Fatalf("tampered digest error=%v", err)
	}
}

func TestLoadReviewedPlanRejectsMalformedApprovalInputs(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "plan.json")
	for _, test := range []struct {
		name   string
		data   string
		digest string
	}{
		{name: "invalid digest", data: "{}\n", digest: "not-a-digest"},
		{name: "not plan", data: `{"operation":"apply","result":"success","error":null,"warnings":[],"evidence":[]}` + "\n"},
		{name: "trailing document", data: string(mustReviewedPlanJSON(t)) + "{}\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := []byte(test.data)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			digest := test.digest
			if digest == "" {
				hash := sha256.Sum256(data)
				digest = hex.EncodeToString(hash[:])
			}
			if _, err := LoadReviewedPlan(path, digest); reviewedIssueCode(err) != IssueMalformedInput {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestVerifyReviewedPlanRecheckRequiresExactStableEvidence(t *testing.T) {
	reviewed := ReviewedPlan{envelope: reviewedPlanFixture(), sha256: strings.Repeat("a", 64), verified: true}
	current := reviewed.envelope
	current.Operation = CommandOperationApply
	current.Evidence = append(append([]Evidence(nil), current.Evidence...), Evidence{Kind: "materialization", Detail: "current cleanup is pending"})
	if err := VerifyReviewedPlanRecheck(reviewed, current); err != nil {
		t.Fatalf("matching recheck: %v", err)
	}

	changed := current
	changed.Plan = cloneReviewedPlan(current.Plan)
	changed.Plan.Operations[0].Expected.Detail = "sha256-v2:changed"
	if err := VerifyReviewedPlanRecheck(reviewed, changed); reviewedIssueCode(err) != IssueReconciliationConflict {
		t.Fatalf("changed expected content error=%v", err)
	}
}

func reviewedPlanFixture() Envelope {
	plan := Plan{
		Desired:    DesiredState{Scope: ScopeGlobal, Skills: []DesiredSkill{}},
		Operations: []PlanOperation{{Action: PlanActionInstall, Skill: "fixture", Target: TargetAgents, Expected: PlanEvidence{Kind: "expected", Detail: "sha256-v2:abc"}}},
	}
	return Envelope{
		Operation: CommandOperationPlan,
		Result:    ResultSuccess,
		Warnings:  []Warning{},
		Evidence:  []Evidence{{Kind: "registry", Detail: "embedded version 4"}, {Kind: "expected-content", Detail: "fixture sha256-v2:abc"}, {Kind: "materialization", Detail: "plan cleanup complete"}},
		Plan:      &plan,
	}
}

func mustReviewedPlanJSON(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(reviewedPlanFixture())
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func cloneReviewedPlan(plan *Plan) *Plan {
	clone := *plan
	clone.Operations = append([]PlanOperation(nil), plan.Operations...)
	return &clone
}

func reviewedIssueCode(err error) IssueCode {
	issue, ok := err.(*Issue)
	if !ok {
		return ""
	}
	return issue.Code
}
