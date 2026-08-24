package sjskills

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
)

type GlobalClassification struct {
	Managed           ProjectClassification     `json:"managed"`
	LegacyRoot        GlobalLegacyRootInventory `json:"legacyRoot"`
	Protected         []GlobalProtectedLocation `json:"protected"`
	ProvenanceFormat  GlobalProvenanceFormat    `json:"provenanceFormat"`
	MigrationRequired bool                      `json:"migrationRequired"`
}

// ClassifyGlobal applies the managed-exact ownership rules to the fixed global
// roots. Legacy and protected locations remain observations only. In
// particular, byte equality and Skills CLI vendor locks never confer ownership.
func ClassifyGlobal(registry Registry, desired DesiredState, expected map[string]TreeHash, inventory GlobalInventory) (GlobalClassification, error) {
	desiredByName, expectedByName, recordsByKey, rootsByTarget, issues := validateGlobalClassificationInputs(registry, desired, expected, inventory)
	if err := newValidationErrors(issues); err != nil {
		return GlobalClassification{}, err
	}
	managed := classifyManagedState(inventory.Home, inventory.StatePath, inventory.StateTrusted, desiredByName, expectedByName, recordsByKey, rootsByTarget)

	known := make(map[string]struct{}, len(registry.Skills))
	for _, declaration := range registry.Skills {
		known[declaration.Name] = struct{}{}
	}
	desiredPlacements := make(map[string]struct{})
	for _, skill := range desiredByName {
		for _, target := range skill.Targets {
			desiredPlacements[projectPlacementKey(target, skill.Name)] = struct{}{}
		}
	}
	stateKeys := make(map[string]struct{}, len(managed.States))
	for index := range managed.States {
		state := &managed.States[index]
		key := projectPlacementKey(state.Target, state.Skill)
		if state.Skill != "" {
			stateKeys[key] = struct{}{}
		}
		_, desiredPlacement := desiredPlacements[key]
		if !desiredPlacement && state.SourceIdentity != "" {
			// Legacy provenance identifies prior global ownership but deliberately
			// grants no removal authority. Former-profile and other non-baseline
			// placements remain report-only through the v1 global apply boundary.
			state.Action = PlanActionUnchanged
			if state.Reason == ProjectStateReasonPreviouslyManagedNotDesired {
				state.Reason = ProjectStateReasonUnmanagedEntryPreserved
			}
			if _, isKnown := known[state.Skill]; isKnown && state.Reason == ProjectStateReasonUnmanagedEntryPreserved {
				state.Reason = ProjectStateReasonFormerGlobalSkillPreserved
			}
		} else if state.Reason == ProjectStateReasonUnmanagedEntryPreserved {
			if _, isKnown := known[state.Skill]; isKnown {
				state.Reason = ProjectStateReasonFormerGlobalSkillPreserved
			}
		}
	}
	if inventory.StateTrusted {
		for key, record := range recordsByKey {
			if _, desiredPlacement := desiredPlacements[key]; desiredPlacement {
				continue
			}
			if _, observed := stateKeys[key]; observed {
				continue
			}
			if !rootsByTarget[record.Target].Safe {
				continue
			}
			managed.States = append(managed.States, ProjectState{
				Kind:           ProjectStateUnmanaged,
				Action:         PlanActionUnchanged,
				Skill:          record.Skill,
				Target:         record.Target,
				Path:           filepath.Join(rootsByTarget[record.Target].Path, record.Skill),
				Manager:        ManagerSkillsCLI,
				SourceIdentity: record.SourceIdentity,
				Reason:         ProjectStateReasonTrustedPlacementAbsent,
				Expected:       treeHashPointerFromRecord(record),
			})
		}
	}
	sort.SliceStable(managed.States, func(i, j int) bool {
		left, right := managed.States[i], managed.States[j]
		if leftRank, rightRank := projectTargetRank(left.Target), projectTargetRank(right.Target); leftRank != rightRank {
			return leftRank < rightRank
		}
		if order := compareUTF16(left.Skill, right.Skill); order != 0 {
			return order < 0
		}
		return compareUTF16(left.Path, right.Path) < 0
	})

	return GlobalClassification{
		Managed:           detachProjectClassification(managed),
		LegacyRoot:        inventory.LegacyRoot,
		Protected:         append([]GlobalProtectedLocation(nil), inventory.Protected...),
		ProvenanceFormat:  inventory.ProvenanceFormat,
		MigrationRequired: inventory.MigrationRequired,
	}, nil
}

func validateGlobalClassificationInputs(registry Registry, desired DesiredState, expected map[string]TreeHash, inventory GlobalInventory) (map[string]DesiredSkill, map[string]TreeHash, map[string]ProvenanceRecord, map[Target]ProjectSkillsRootInventory, []Issue) {
	var issues []Issue
	if err := ValidateRegistry(registry); err != nil {
		issues = append(issues, Issue{Code: IssueMalformedInput, Path: "registry", Message: "global classification requires a valid version 4 registry"})
	}
	resolved, resolveErr := ResolveGlobal(registry)
	if resolveErr != nil || !reflect.DeepEqual(resolved, desired) {
		issues = append(issues, Issue{Code: IssueMalformedInput, Path: "desired", Message: "global desired state must equal the registry fixed baseline"})
	}
	desiredByName := make(map[string]DesiredSkill, len(desired.Skills))
	if desired.Scope != ScopeGlobal {
		issues = append(issues, Issue{Code: IssueMalformedInput, Path: "desired.scope", Message: "must be global scope"})
	}
	for index, skill := range desired.Skills {
		path := "desired.skills[" + strconv.Itoa(index) + "]"
		if !isPortableName(skill.Name) {
			issues = append(issues, Issue{Code: IssueInvalidName, Path: path + ".name", Message: "must be a portable skill name"})
		}
		if _, duplicate := desiredByName[skill.Name]; duplicate {
			issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".name", Message: "duplicate desired skill name"})
		} else {
			desiredByName[skill.Name] = skill
		}
		if skill.Scope != ScopeGlobal {
			issues = append(issues, Issue{Code: IssueMalformedInput, Path: path + ".scope", Message: "must be global scope"})
		}
		seenTargets := make(map[Target]struct{}, len(skill.Targets))
		if len(skill.Targets) == 0 {
			issues = append(issues, Issue{Code: IssueEmptySelection, Path: path + ".targets", Message: "must contain at least one target"})
		}
		for _, target := range skill.Targets {
			if _, supported := supportedTargets[target]; !supported {
				issues = append(issues, Issue{Code: IssueInvalidTarget, Path: path + ".targets", Message: "unsupported target"})
			}
			if _, duplicate := seenTargets[target]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".targets", Message: "duplicate target"})
			}
			seenTargets[target] = struct{}{}
		}
		if skill.Manager == ManagerSkillsCLI {
			if skill.Mode != ModeCopy {
				issues = append(issues, Issue{Code: IssueInvalidMode, Path: path + ".mode", Message: "skills-cli installation must use copy mode"})
			}
			if _, ok := canonicalProjectSourceIdentity(skill.Source); !ok {
				issues = append(issues, Issue{Code: IssueInvalidSource, Path: path + ".source", Message: "must be a credential-free remote source"})
			}
		} else if skill.Manager != ManagerManual && skill.Manager != ManagerWorkflow {
			issues = append(issues, Issue{Code: IssueInvalidManager, Path: path + ".manager", Message: "unsupported manager"})
		}
	}

	expectedByName := make(map[string]TreeHash, len(expected))
	for name, hash := range expected {
		skill, desiredName := desiredByName[name]
		if !desiredName || skill.Manager != ManagerSkillsCLI {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "expected." + name, Message: "expected hash requires a desired skills-cli skill"})
			continue
		}
		if hash.Algorithm != TreeHashAlgorithmSHA256V2 || !lowercaseDigestPattern.MatchString(hash.Digest) {
			issues = append(issues, Issue{Code: IssueMalformedInput, Path: "expected." + name, Message: "expected hash must use tree-sha256-v2 and lowercase 64-hex digest"})
		}
		expectedByName[name] = hash
	}
	for name, skill := range desiredByName {
		if skill.Manager == ManagerSkillsCLI {
			if _, ok := expected[name]; !ok {
				issues = append(issues, Issue{Code: IssueMissingReference, Path: "expected." + name, Message: "skills-cli desired skill requires one expected hash"})
			}
		}
	}

	layout, layoutErr := LayoutForGlobal(inventory.Home)
	if layoutErr != nil || layout.Home != inventory.Home {
		issues = append(issues, Issue{Code: IssueInvalidRoot, Path: "inventory.home", Message: "inventory home must be the exact absolute global home"})
	}
	rootsByTarget := make(map[Target]ProjectSkillsRootInventory, len(inventory.Roots))
	if layoutErr == nil {
		if inventory.StatePath != layout.ProvenanceStatePath {
			issues = append(issues, Issue{Code: IssuePathEscape, Path: "inventory.statePath", Message: "inventory state path does not match the global layout"})
		}
		for index, root := range inventory.Roots {
			path := "inventory.roots[" + strconv.Itoa(index) + "]"
			if _, duplicate := rootsByTarget[root.Target]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".target", Message: "duplicate inventory target"})
			} else {
				rootsByTarget[root.Target] = root
			}
			expectedPath, pathErr := layout.ManagedSkillsPath(root.Target)
			if pathErr != nil || root.Path != expectedPath {
				issues = append(issues, Issue{Code: IssuePathEscape, Path: path + ".path", Message: "inventory root path does not match the global layout"})
			}
			seenNames := make(map[string]struct{}, len(root.Entries))
			for entryIndex, entry := range root.Entries {
				entryPath := path + ".entries[" + strconv.Itoa(entryIndex) + "]"
				if entry.Target != root.Target || entry.Path != filepath.Join(root.Path, entry.Name) {
					issues = append(issues, Issue{Code: IssuePathEscape, Path: entryPath, Message: "inventory entry does not match its global root"})
				}
				if _, duplicate := seenNames[entry.Name]; duplicate {
					issues = append(issues, Issue{Code: IssueDuplicate, Path: entryPath + ".name", Message: "duplicate inventory entry name"})
				}
				seenNames[entry.Name] = struct{}{}
			}
		}
		if inventory.LegacyRoot.ID != "pi" || inventory.LegacyRoot.Kind != GlobalLocationLegacy || inventory.LegacyRoot.Path != layout.LegacyPiSkillsPath {
			issues = append(issues, Issue{Code: IssuePathEscape, Path: "inventory.legacyRoot", Message: "legacy root does not match the global layout"})
		}
		seenLegacyNames := make(map[string]struct{}, len(inventory.LegacyRoot.Entries))
		for index, entry := range inventory.LegacyRoot.Entries {
			path := fmt.Sprintf("inventory.legacyRoot.entries[%d]", index)
			if entry.Target != "" || entry.Path != filepath.Join(inventory.LegacyRoot.Path, entry.Name) {
				issues = append(issues, Issue{Code: IssuePathEscape, Path: path, Message: "legacy entry does not match the modeled root"})
			}
			if _, duplicate := seenLegacyNames[entry.Name]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".name", Message: "duplicate legacy entry name"})
			}
			seenLegacyNames[entry.Name] = struct{}{}
		}
		expectedProtected := map[string]struct {
			kind GlobalLocationKind
			path string
		}{
			"agents-vendor-lock": {kind: GlobalLocationVendorMetadata, path: layout.AgentsVendorLockPath},
			"claude-vendor-lock": {kind: GlobalLocationVendorMetadata, path: layout.ClaudeVendorLockPath},
			"legacy-quarantine":  {kind: GlobalLocationBackup, path: layout.LegacyQuarantinePath},
			"codex-plugin-cache": {kind: GlobalLocationCache, path: layout.CodexPluginCachePath},
			"codex-system":       {kind: GlobalLocationRuntime, path: layout.CodexSystemSkillsPath},
		}
		seenProtected := make(map[string]struct{}, len(inventory.Protected))
		for index, location := range inventory.Protected {
			path := fmt.Sprintf("inventory.protected[%d]", index)
			expected, ok := expectedProtected[location.ID]
			if !ok || location.Kind != expected.kind || location.Path != expected.path {
				issues = append(issues, Issue{Code: IssuePathEscape, Path: path, Message: "protected location does not match the global layout"})
			}
			if _, duplicate := seenProtected[location.ID]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".id", Message: "duplicate protected location"})
			}
			seenProtected[location.ID] = struct{}{}
		}
		if len(seenProtected) != len(expectedProtected) {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "inventory.protected", Message: "inventory must contain every fixed protected location"})
		}
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		if _, ok := rootsByTarget[target]; !ok {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "inventory.roots", Message: "inventory must contain exactly one root for each target"})
		}
	}

	recordsByKey := make(map[string]ProvenanceRecord, len(inventory.State.Records))
	if inventory.StateTrusted {
		if inventory.ProvenanceFormat == "" || inventory.State.Version != GlobalProvenanceStateVersion || !validGlobalProvenanceRecords(inventory.State.Records) {
			issues = append(issues, Issue{Code: IssueMalformedInput, Path: "inventory.state", Message: "trusted global provenance is invalid"})
		}
		if inventory.MigrationRequired != (inventory.ProvenanceFormat == GlobalProvenanceLegacyV1) {
			issues = append(issues, Issue{Code: IssueMalformedInput, Path: "inventory.migrationRequired", Message: "migration evidence is inconsistent"})
		}
		for index, record := range inventory.State.Records {
			key := projectPlacementKey(record.Target, record.Skill)
			if _, duplicate := recordsByKey[key]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: fmt.Sprintf("inventory.state.records[%d]", index), Message: "duplicate trusted provenance placement"})
			} else {
				recordsByKey[key] = record
			}
		}
	} else if len(inventory.State.Records) != 0 || inventory.MigrationRequired {
		issues = append(issues, Issue{Code: IssueMalformedInput, Path: "inventory.state", Message: "untrusted global provenance cannot carry records or migration authority"})
	}
	return desiredByName, expectedByName, recordsByKey, rootsByTarget, issues
}

// TranslateGlobalClassification keeps global planning on the shared Plan
// contract while adding only path-free migration/protection warnings.
func TranslateGlobalClassification(plan Plan, classification GlobalClassification) (Plan, error) {
	if plan.Desired.Scope != ScopeGlobal {
		return Plan{}, &ValidationErrors{Issues: []Issue{{
			Code: IssueMalformedInput, Path: "plan.desired.scope", Message: "global translation requires global scope",
		}}}
	}
	result := translateManagedClassification(plan, classification.Managed, globalStateWarning, globalPreservedWarning)
	result.Evidence = append(result.Evidence, Evidence{Kind: "global-inventory", Detail: "fixed managed and migration roots inspected read-only"})
	if classification.MigrationRequired {
		result.Evidence = append(result.Evidence, Evidence{Kind: "provenance-migration", Detail: "trusted legacy version 1 provenance requires version 2 migration"})
	}
	for _, entry := range classification.LegacyRoot.Entries {
		result.Warnings = append(result.Warnings, Warning{
			Code:    "legacy-preserved",
			Message: fmt.Sprintf("preserved legacy global entry pi/%s", projectStableSkillName(entry.Name)),
		})
	}
	for _, problem := range classification.LegacyRoot.Problems {
		result.Warnings = append(result.Warnings, Warning{Code: "legacy-protected", Message: fmt.Sprintf("legacy global root preserved (%s)", problem.Reason)})
	}
	for _, location := range classification.Protected {
		if !location.Exists && location.Problem == "" {
			continue
		}
		reason := "protected"
		if location.Problem != "" {
			reason = string(location.Problem)
		}
		result.Warnings = append(result.Warnings, Warning{
			Code:    "global-protected",
			Message: fmt.Sprintf("preserved %s global location %s (%s)", location.Kind, location.ID, reason),
		})
	}
	return result, nil
}

func globalStateWarning(state ProjectState) Warning {
	return Warning{Code: "global-state", Message: fmt.Sprintf("global state marker preserved (%s)", projectStableReason(state.Reason))}
}

func globalPreservedWarning(state ProjectState) Warning {
	if state.Reason == ProjectStateReasonTrustedPlacementAbsent {
		return Warning{
			Code:    "global-migration",
			Message: fmt.Sprintf("trusted former global placement %s/%s is absent (%s)", state.Target, projectStableSkillName(state.Skill), projectStableReason(state.Reason)),
		}
	}
	if state.Reason == ProjectStateReasonFormerGlobalSkillPreserved {
		return Warning{
			Code:    "global-migration",
			Message: fmt.Sprintf("former global entry %s/%s preserved (%s)", state.Target, projectStableSkillName(state.Skill), projectStableReason(state.Reason)),
		}
	}
	if state.SourceIdentity != "" {
		return Warning{
			Code:    "global-migration",
			Message: fmt.Sprintf("former managed global entry %s/%s preserved (%s)", state.Target, projectStableSkillName(state.Skill), projectStableReason(state.Reason)),
		}
	}
	return Warning{
		Code:    "unmanaged-preserved",
		Message: fmt.Sprintf("preserved unmanaged global entry %s/%s (%s)", state.Target, projectStableSkillName(state.Skill), projectStableReason(state.Reason)),
	}
}
