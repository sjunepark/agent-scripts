package sjskills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestReviewedAdvisoriesAreStrictButOutsideSemanticApproval(t *testing.T) {
	now := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	envelope := reviewedPlanFixture()
	advisory := newAdvisory(ScopeGlobal)
	advisory.Freshness = AdvisoryFresh
	advisory.ObservedAt = &now
	advisory.Findings = []AdvisoryFinding{{Category: AdvisoryMissing, Skill: "fixture", Target: TargetAgents, Reason: "expected-entry-absent"}}
	envelope.Advisories = []Advisory{advisory}
	load := func(t *testing.T, data []byte) (ReviewedPlan, error) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "plan.json")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		return LoadReviewedPlan(path, hex.EncodeToString(hash[:]))
	}
	data, _ := json.Marshal(envelope)
	reviewed, err := load(t, data)
	if err != nil {
		t.Fatal(err)
	}
	current := envelope
	current.Advisories = nil
	if err := VerifyReviewedPlanRecheck(reviewed, current); err != nil {
		t.Fatalf("omitted advisory affects approval: %v", err)
	}
	changed := advisory
	later := now.Add(time.Hour)
	changed.ObservedAt = &later
	changed.Cached = true
	changed.Findings = []AdvisoryFinding{{Category: AdvisoryConflict, Skill: "different", Target: TargetClaude, Reason: "local-modification"}}
	current.Advisories = []Advisory{changed}
	if err := VerifyReviewedPlanRecheck(reviewed, current); err != nil {
		t.Fatalf("advisory change affects approval: %v", err)
	}
	for _, mutate := range []func(*Envelope){
		func(e *Envelope) { e.Warnings = append(e.Warnings, Warning{Code: "stable", Message: "changed"}) },
		func(e *Envelope) { e.Plan = cloneReviewedPlan(e.Plan); e.Plan.Operations[0].Current.Detail = "changed" },
		func(e *Envelope) {
			e.Plan = cloneReviewedPlan(e.Plan)
			e.Plan.Operations[0].Reason = "different-ownership"
		},
		func(e *Envelope) {
			e.Plan = cloneReviewedPlan(e.Plan)
			e.Plan.Operations[0].Expected.Detail = "changed"
		},
	} {
		candidate := current
		mutate(&candidate)
		if VerifyReviewedPlanRecheck(reviewed, candidate) == nil {
			t.Fatal("stable evidence change ignored")
		}
	}
	// Exact artifact bytes, including notices, still require a matching digest.
	path := filepath.Join(t.TempDir(), "artifact.json")
	originalHash := sha256.Sum256(data)
	changedData, _ := json.Marshal(current)
	_ = os.WriteFile(path, changedData, 0600)
	if _, err := LoadReviewedPlan(path, hex.EncodeToString(originalHash[:])); reviewedIssueCode(err) != IssueReconciliationConflict {
		t.Fatalf("notice byte digest not bound: %v", err)
	}
	for _, mutation := range []func(*Advisory){
		func(a *Advisory) { a.Freshness = "clean" }, func(a *Advisory) { a.Scope = "other" }, func(a *Advisory) { a.ObservedAt = nil }, func(a *Advisory) { a.ReviewCommand = "apply" },
		func(a *Advisory) {
			a.Findings = []AdvisoryFinding{{Category: "install", Reason: "expected-entry-absent"}}
		},
		func(a *Advisory) {
			a.Findings = []AdvisoryFinding{{Category: AdvisoryMissing, Target: ".pi", Reason: "expected-entry-absent"}}
		},
	} {
		bad := advisory
		mutation(&bad)
		candidate := envelope
		candidate.Advisories = []Advisory{bad}
		bytes, _ := json.Marshal(candidate)
		if _, err := load(t, bytes); reviewedIssueCode(err) != IssueMalformedInput {
			t.Fatalf("invalid advisory accepted: %+v %v", bad, err)
		}
	}
	unknown := strings.Replace(string(data), `"freshness":"fresh"`, `"unexpected":true,"freshness":"fresh"`, 1)
	if _, err := load(t, []byte(unknown)); reviewedIssueCode(err) != IssueMalformedInput {
		t.Fatalf("unknown nested advisory field accepted: %v", err)
	}
}
