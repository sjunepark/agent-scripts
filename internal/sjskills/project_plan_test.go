package sjskills

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateProjectClassificationStableOperationsAndEvidence(t *testing.T) {
	expectedAlpha := classificationHash('a')
	expectedBeta := classificationHash('b')
	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{
		{
			Name: "alpha", SourceID: "catalog", Source: "https://github.com/example/catalog/tree/main/skills",
			Scope: ScopeProject, Origin: "profile:dev", Manager: ManagerSkillsCLI, Mode: ModeCopy,
			Targets: []Target{TargetAgents},
		},
		{
			Name: "beta", Source: "example/beta", Scope: ScopeProject, Origin: "direct",
			Manager: ManagerSkillsCLI, Mode: ModeCopy, Targets: []Target{TargetClaude},
		},
	}}
	plan := Plan{
		Desired:    desired,
		Operations: []PlanOperation{},
		Warnings:   []Warning{{Code: "manual-action", Message: "keep"}},
		Evidence:   []Evidence{{Kind: "resolution", Detail: "resolved 2 desired skills"}},
	}
	classification := ProjectClassification{Root: "/private/project", States: []ProjectState{
		{
			Kind: ProjectStateUnmanaged, Action: PlanActionUnchanged, Skill: "unknown", Target: TargetAgents,
			Path: "/private/project/.agents/skills/unknown", Reason: ProjectStateReasonUnmanagedEntryPreserved,
		},
		{
			Kind: ProjectStateMissing, Action: PlanActionInstall, Skill: "beta", Target: TargetClaude,
			Path: "/private/project/.claude/skills/beta", Manager: ManagerSkillsCLI,
			Reason: ProjectStateReasonExpectedEntryAbsent, Expected: &expectedBeta,
		},
		{
			Kind: ProjectStateProtected, Action: PlanActionBlocked, Path: "/private/project/.sjskills/state.json",
			Reason: ProjectStateReasonProvenanceUntrusted,
		},
		{
			Kind: ProjectStateExact, Action: PlanActionUnchanged, Skill: "alpha", Target: TargetAgents,
			Path: "/private/project/.agents/skills/alpha", Manager: ManagerSkillsCLI,
			Reason: ProjectStateReasonVerifiedExact, Current: &expectedAlpha, Expected: &expectedAlpha,
		},
		{
			Kind: ProjectStateExact, Action: PlanActionQuarantine, Skill: "gamma", Target: TargetAgents,
			Path: "/private/project/.agents/skills/gamma", Manager: ManagerSkillsCLI,
			SourceIdentity: "github:example/gamma", Reason: ProjectStateReasonPreviouslyManagedNotDesired,
			Current: &expectedAlpha, Expected: &expectedAlpha,
		},
	}}

	translated, err := TranslateProjectClassification(plan, classification)
	if err != nil {
		t.Fatal(err)
	}
	if len(translated.Operations) != 3 {
		t.Fatalf("operations = %#v, want desired placements plus managed removal", translated.Operations)
	}
	alpha, gamma, beta := translated.Operations[0], translated.Operations[1], translated.Operations[2]
	if alpha.Skill != "alpha" || alpha.Target != TargetAgents || alpha.Action != PlanActionUnchanged || alpha.SourceID != "catalog" || alpha.Source != desired.Skills[0].Source {
		t.Fatalf("alpha operation = %#v", alpha)
	}
	if alpha.Current != (PlanEvidence{Kind: projectEvidenceTreeHash, Detail: expectedAlpha.Algorithm + ":" + expectedAlpha.Digest}) || alpha.Expected != alpha.Current {
		t.Fatalf("alpha evidence = %#v", alpha)
	}
	if gamma.Skill != "gamma" || gamma.Target != TargetAgents || gamma.Action != PlanActionQuarantine || gamma.Manager != ManagerSkillsCLI || gamma.SourceID != "" || gamma.Source != "github:example/gamma" {
		t.Fatalf("gamma operation = %#v", gamma)
	}
	if beta.Skill != "beta" || beta.Target != TargetClaude || beta.Action != PlanActionInstall || beta.SourceID != "" || beta.Source != "example/beta" {
		t.Fatalf("beta operation = %#v", beta)
	}
	if beta.Current != (PlanEvidence{Kind: projectEvidenceAbsent, Detail: projectEvidenceAbsent}) || beta.Expected.Detail != expectedBeta.Algorithm+":"+expectedBeta.Digest {
		t.Fatalf("beta evidence = %#v", beta)
	}
	if len(translated.Warnings) != 3 || translated.Warnings[1].Code != "unmanaged-preserved" || translated.Warnings[2].Code != "project-state" {
		t.Fatalf("warnings = %#v", translated.Warnings)
	}

	first, err := json.Marshal(translated)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(translated)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("translation JSON is not stable:\n%s\n%s", first, second)
	}
	if strings.Contains(string(first), "sjskills-materialize-") || strings.Contains(string(first), ".sjskills/state.json") || strings.Contains(string(first), "/private/project") {
		t.Fatalf("translation leaked path evidence: %s", first)
	}
}

func TestTranslateProjectClassificationRejectsGlobalPlan(t *testing.T) {
	_, err := TranslateProjectClassification(Plan{Desired: DesiredState{Scope: ScopeGlobal}}, ProjectClassification{})
	if err == nil || !issueCode(err, IssueMalformedInput) {
		t.Fatalf("global translation error = %v, want malformed input", err)
	}
}

func TestTranslateProjectClassificationBoundsUnverifiableEvidence(t *testing.T) {
	plan := Plan{Desired: DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{{
		Name: "symlinked", Scope: ScopeProject, Manager: ManagerSkillsCLI, Mode: ModeCopy,
		Source: "example/symlinked", Targets: []Target{TargetAgents},
	}}}}
	classification := ProjectClassification{States: []ProjectState{{
		Kind: ProjectStateMisplaced, Action: PlanActionBlocked, Skill: "symlinked", Target: TargetAgents,
		Reason:   ProjectStateReasonExpectedCopyIsSymlink,
		Current:  &TreeHash{Algorithm: TreeHashAlgorithmSHA256V2, Digest: "/private/staging/path"},
		Expected: &TreeHash{Algorithm: "raw-error", Digest: "/private/staging/path"},
	}}}
	translated, err := TranslateProjectClassification(plan, classification)
	if err != nil {
		t.Fatal(err)
	}
	if len(translated.Operations) != 1 {
		t.Fatalf("operations = %#v, want one blocked operation", translated.Operations)
	}
	operation := translated.Operations[0]
	if operation.Current != (PlanEvidence{Kind: projectEvidenceUnavailable, Detail: projectEvidenceUnavailable}) || operation.Expected != operation.Current {
		t.Fatalf("bounded evidence = %#v, want unavailable markers", operation)
	}
	encoded, err := json.Marshal(translated)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "/private/staging/path") || strings.Contains(string(encoded), "raw-error") {
		t.Fatalf("unverifiable evidence leaked raw data: %s", encoded)
	}
}

func TestTranslateProjectClassificationEscapesInvalidObservedName(t *testing.T) {
	plan := Plan{Desired: DesiredState{Scope: ScopeProject}}
	classification := ProjectClassification{States: []ProjectState{{
		Kind: ProjectStateMalformed, Action: PlanActionBlocked, Skill: "bad\nname", Target: TargetAgents,
		Reason: ProjectStateReasonMalformedEntryPreserved,
	}}}
	translated, err := TranslateProjectClassification(plan, classification)
	if err != nil {
		t.Fatal(err)
	}
	if len(translated.Warnings) != 1 || strings.Contains(translated.Warnings[0].Message, "\n") || !strings.Contains(translated.Warnings[0].Message, `"bad\nname"`) {
		t.Fatalf("warning = %#v, want one escaped line", translated.Warnings)
	}
}
