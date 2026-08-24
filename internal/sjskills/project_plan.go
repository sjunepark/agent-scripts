package sjskills

import (
	"fmt"
	"sort"
	"strconv"
)

const (
	projectEvidenceTreeHash      = "tree-hash"
	projectEvidenceAbsent        = "absent"
	projectEvidenceUnavailable   = "unavailable"
	projectEvidenceNotApplicable = "not-applicable"
)

// TranslateProjectClassification copies the deterministic project states into
// the existing Plan envelope.  It deliberately does not introduce a second
// operation model: desired placements become PlanOperation values, while
// unowned observations and state-level markers remain warnings because an
// operation would falsely imply ownership.
//
// Classification state order is already deterministic (target, UTF-16 skill,
// then path), so operations and warnings retain that order.  The function is
// pure: neither argument, and no filesystem or process state, is modified.
func TranslateProjectClassification(plan Plan, classification ProjectClassification) (Plan, error) {
	if plan.Desired.Scope != ScopeProject {
		return Plan{}, &ValidationErrors{Issues: []Issue{{
			Code: IssueMalformedInput, Path: "plan.desired.scope", Message: "project translation requires project scope",
		}}}
	}
	return translateManagedClassification(plan, classification, projectStateWarning, projectPreservedWarning), nil
}

type classificationWarningFunc func(ProjectState) Warning

func translateManagedClassification(plan Plan, classification ProjectClassification, stateWarning, preservedWarning classificationWarningFunc) Plan {
	result := plan
	result.Desired = cloneDesiredState(plan.Desired)
	result.Operations = make([]PlanOperation, 0, len(classification.States))
	result.Warnings = append([]Warning{}, plan.Warnings...)
	result.Evidence = append([]Evidence{}, plan.Evidence...)

	desiredByPlacement := make(map[string]DesiredSkill)
	for _, skill := range result.Desired.Skills {
		for _, target := range skill.Targets {
			desiredByPlacement[projectPlacementKey(target, skill.Name)] = skill
		}
	}

	states := append([]ProjectState(nil), classification.States...)
	sort.SliceStable(states, func(i, j int) bool {
		left, right := states[i], states[j]
		if leftRank, rightRank := projectTargetRank(left.Target), projectTargetRank(right.Target); leftRank != rightRank {
			return leftRank < rightRank
		}
		if nameOrder := compareUTF16(left.Skill, right.Skill); nameOrder != 0 {
			return nameOrder < 0
		}
		return compareUTF16(left.Path, right.Path) < 0
	})
	for _, state := range states {
		desired, desiredPlacement := desiredByPlacement[projectPlacementKey(state.Target, state.Skill)]
		if state.Skill == "" {
			result.Warnings = append(result.Warnings, stateWarning(state))
			continue
		}

		// An observed entry that is not one of the desired placements cannot
		// safely become a plan operation.  In particular, an unknown byte
		// sequence must not be claimed merely because it shares a name with a
		// desired skill at another target.  Provenance-backed removed entries
		// retain their quarantine/blocked operation below.
		if !desiredPlacement && state.Action == PlanActionUnchanged {
			result.Warnings = append(result.Warnings, preservedWarning(state))
			continue
		}
		if !desiredPlacement && state.Manager != ManagerSkillsCLI && state.Action == PlanActionBlocked {
			result.Warnings = append(result.Warnings, preservedWarning(state))
			continue
		}

		operation := PlanOperation{
			Action:   state.Action,
			Skill:    state.Skill,
			Target:   state.Target,
			Manager:  state.Manager,
			Reason:   projectStableReason(state.Reason),
			Current:  projectCurrentEvidence(state),
			Expected: projectExpectedEvidence(state.Expected, desired),
		}
		if desiredPlacement {
			operation.Manager = desired.Manager
			operation.SourceID = desired.SourceID
			operation.Source = desired.Source
		} else {
			// This is an existing placement with trusted provenance but no
			// current desired identity (for example, a removed skill).  Keep
			// ownership evidence bounded and explicit without inventing a
			// registry source ID.
			operation.Source = state.SourceIdentity
		}
		result.Operations = append(result.Operations, operation)
	}

	return result
}

func cloneDesiredState(state DesiredState) DesiredState {
	state.Skills = append([]DesiredSkill(nil), state.Skills...)
	for index := range state.Skills {
		state.Skills[index].Targets = copyTargets(state.Skills[index].Targets)
	}
	if state.Skills == nil {
		state.Skills = []DesiredSkill{}
	}
	return state
}

func projectCurrentEvidence(state ProjectState) PlanEvidence {
	if state.Current != nil {
		return projectTreeHashEvidence(state.Current, projectEvidenceUnavailable)
	}
	if state.Kind == ProjectStateMissing {
		return PlanEvidence{Kind: projectEvidenceAbsent, Detail: projectEvidenceAbsent}
	}
	return PlanEvidence{Kind: projectEvidenceUnavailable, Detail: projectEvidenceUnavailable}
}

func projectExpectedEvidence(hash *TreeHash, desired DesiredSkill) PlanEvidence {
	if hash != nil {
		return projectTreeHashEvidence(hash, projectEvidenceUnavailable)
	}
	if desired.Manager == ManagerManual || desired.Manager == ManagerWorkflow {
		return PlanEvidence{Kind: projectEvidenceNotApplicable, Detail: projectEvidenceNotApplicable}
	}
	return PlanEvidence{Kind: projectEvidenceUnavailable, Detail: projectEvidenceUnavailable}
}

func projectTreeHashEvidence(hash *TreeHash, emptyKind string) PlanEvidence {
	if hash == nil {
		return PlanEvidence{Kind: emptyKind, Detail: emptyKind}
	}
	if hash.Algorithm != TreeHashAlgorithmSHA256V2 || !lowercaseDigestPattern.MatchString(hash.Digest) {
		return PlanEvidence{Kind: projectEvidenceUnavailable, Detail: projectEvidenceUnavailable}
	}
	return PlanEvidence{Kind: projectEvidenceTreeHash, Detail: hash.Algorithm + ":" + hash.Digest}
}

func projectStateWarning(state ProjectState) Warning {
	reason := projectStableReason(state.Reason)
	return Warning{
		Code:    "project-state",
		Message: fmt.Sprintf("project state marker preserved (%s)", reason),
	}
}

func projectPreservedWarning(state ProjectState) Warning {
	reason := projectStableReason(state.Reason)
	target := string(state.Target)
	if target == "" {
		target = "project"
	}
	return Warning{
		Code:    "unmanaged-preserved",
		Message: fmt.Sprintf("preserved unmanaged project entry %s/%s (%s)", target, projectStableSkillName(state.Skill), reason),
	}
}

func projectStableSkillName(name string) string {
	if isPortableName(name) {
		return name
	}
	return strconv.QuoteToASCII(name)
}

func projectStableReason(reason ProjectStateReason) string {
	switch reason {
	case ProjectStateReasonRootUnavailable,
		ProjectStateReasonExpectedEntryAbsent,
		ProjectStateReasonCurrentEntryUnverifiable,
		ProjectStateReasonExpectedCopyIsSymlink,
		ProjectStateReasonDesiredPathUnmanaged,
		ProjectStateReasonProvenanceSourceMismatch,
		ProjectStateReasonVerifiedExact,
		ProjectStateReasonVerifiedUpdate,
		ProjectStateReasonLocalModification,
		ProjectStateReasonMalformedEntryPreserved,
		ProjectStateReasonPreviouslyManagedNotDesired,
		ProjectStateReasonPreviouslyManagedModified,
		ProjectStateReasonUnmanagedMisplacedPreserved,
		ProjectStateReasonUnmanagedEntryPreserved,
		ProjectStateReasonProvenanceUntrusted,
		ProjectStateReasonManualManager,
		ProjectStateReasonWorkflowManager,
		ProjectStateReasonFormerGlobalSkillPreserved,
		ProjectStateReasonTrustedPlacementAbsent:
		return string(reason)
	default:
		return "unspecified"
	}
}
