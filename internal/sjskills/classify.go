package sjskills

import (
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ProjectStateKind is the complete state vocabulary emitted by the project
// classifier.  Unknown bytes never acquire a managed state by matching the
// expected tree; trusted provenance is the ownership proof.
type ProjectStateKind string

const (
	ProjectStateMissing   ProjectStateKind = "missing"
	ProjectStateExact     ProjectStateKind = "exact"
	ProjectStateOutdated  ProjectStateKind = "outdated"
	ProjectStateModified  ProjectStateKind = "modified"
	ProjectStateUnmanaged ProjectStateKind = "unmanaged"
	ProjectStateMalformed ProjectStateKind = "malformed"
	ProjectStateMisplaced ProjectStateKind = "misplaced"
	ProjectStateProtected ProjectStateKind = "protected"
)

// ProjectStateReason is a bounded explanation for one state.  Reasons are
// deliberately independent of operating-system errors and filesystem text.
type ProjectStateReason string

const (
	ProjectStateReasonRootUnavailable             ProjectStateReason = "root-unavailable"
	ProjectStateReasonExpectedEntryAbsent         ProjectStateReason = "expected-entry-absent"
	ProjectStateReasonCurrentEntryUnverifiable    ProjectStateReason = "current-entry-unverifiable"
	ProjectStateReasonExpectedCopyIsSymlink       ProjectStateReason = "expected-copy-is-symlink"
	ProjectStateReasonDesiredPathUnmanaged        ProjectStateReason = "desired-path-unmanaged"
	ProjectStateReasonProvenanceSourceMismatch    ProjectStateReason = "provenance-source-mismatch"
	ProjectStateReasonVerifiedExact               ProjectStateReason = "verified-exact"
	ProjectStateReasonVerifiedUpdate              ProjectStateReason = "verified-update"
	ProjectStateReasonLocalModification           ProjectStateReason = "local-modification"
	ProjectStateReasonMalformedEntryPreserved     ProjectStateReason = "malformed-entry-preserved"
	ProjectStateReasonPreviouslyManagedNotDesired ProjectStateReason = "previously-managed-not-desired"
	ProjectStateReasonPreviouslyManagedModified   ProjectStateReason = "previously-managed-modified"
	ProjectStateReasonUnmanagedMisplacedPreserved ProjectStateReason = "unmanaged-misplaced-preserved"
	ProjectStateReasonUnmanagedEntryPreserved     ProjectStateReason = "unmanaged-entry-preserved"
	ProjectStateReasonProvenanceUntrusted         ProjectStateReason = "provenance-untrusted"
	ProjectStateReasonManualManager               ProjectStateReason = "manual-manager"
	ProjectStateReasonWorkflowManager             ProjectStateReason = "workflow-manager"
)

// ProjectState is one desired placement or one observed entry.  Current and
// Expected are copied values so callers cannot mutate classifier-owned output.
type ProjectState struct {
	Kind           ProjectStateKind   `json:"kind"`
	Action         PlanAction         `json:"action"`
	Skill          string             `json:"skill,omitempty"`
	Target         Target             `json:"target,omitempty"`
	Path           string             `json:"path"`
	Manager        Manager            `json:"manager,omitempty"`
	SourceIdentity string             `json:"sourceIdentity,omitempty"`
	Reason         ProjectStateReason `json:"reason"`
	Current        *TreeHash          `json:"current,omitempty"`
	Expected       *TreeHash          `json:"expected,omitempty"`
}

// ProjectClassification is a detached, deterministic read-only result.
type ProjectClassification struct {
	Root   string         `json:"root"`
	States []ProjectState `json:"states"`
}

var projectSourceSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var projectCanonicalIdentityPattern = regexp.MustCompile(`^[a-z0-9_.-]+/[a-z0-9_.-]+(?:/[a-z0-9_.-]+)*$`)

// ClassifyProject validates the pure desired/hash/inventory boundary and
// classifies every desired placement plus every observed non-canonical entry.
// It only reads its arguments; no filesystem, process, clock, or network call
// is made here.
func ClassifyProject(desired DesiredState, expected map[string]TreeHash, inventory ProjectInventory) (ProjectClassification, error) {
	desiredByName, expectedByName, recordsByKey, rootsByTarget, issues := validateProjectClassificationInputs(desired, expected, inventory)
	if err := newValidationErrors(issues); err != nil {
		return ProjectClassification{}, err
	}

	result := ProjectClassification{
		Root:   inventory.Root,
		States: make([]ProjectState, 0),
	}
	if !inventory.StateTrusted {
		result.States = append(result.States, ProjectState{
			Kind:   ProjectStateProtected,
			Action: PlanActionBlocked,
			Path:   inventory.StatePath,
			Reason: ProjectStateReasonProvenanceUntrusted,
		})
	}

	placements := desiredPlacements(desiredByName)
	desiredPlacementKeys := make(map[string]struct{}, len(placements))
	for _, placement := range placements {
		desiredPlacementKeys[projectPlacementKey(placement.target, placement.skill.Name)] = struct{}{}
	}
	for _, placement := range placements {
		root := rootsByTarget[placement.target]
		path := filepath.Join(root.Path, placement.skill.Name)
		expectedHash := expectedByName[placement.skill.Name]
		state := ProjectState{
			Skill:   placement.skill.Name,
			Target:  placement.target,
			Path:    path,
			Manager: placement.skill.Manager,
		}

		switch placement.skill.Manager {
		case ManagerManual:
			state.Kind = ProjectStateProtected
			state.Action = PlanActionManual
			state.Reason = ProjectStateReasonManualManager
		case ManagerWorkflow:
			state.Kind = ProjectStateProtected
			state.Action = PlanActionWorkflow
			state.Reason = ProjectStateReasonWorkflowManager
		case ManagerSkillsCLI:
			state.Expected = cloneTreeHashPointer(&expectedHash)
			state.SourceIdentity, _ = canonicalProjectSourceIdentity(placement.skill.Source)
			classifyCopyPlacement(&state, placement.skill, root, recordsByKey, expectedHash)
		}
		result.States = append(result.States, state)
	}

	for _, target := range []Target{TargetAgents, TargetClaude} {
		root := rootsByTarget[target]
		if !root.Safe {
			continue
		}
		entries := append([]InventoryEntry(nil), root.Entries...)
		sort.SliceStable(entries, func(i, j int) bool {
			return compareUTF16(entries[i].Name, entries[j].Name) < 0
		})
		for _, entry := range entries {
			if _, desiredAtTarget := desiredPlacementKeys[projectPlacementKey(target, entry.Name)]; desiredAtTarget {
				continue
			}
			key := projectPlacementKey(target, entry.Name)
			if record, managed := recordsByKey[key]; managed {
				result.States = append(result.States, classifyRemovedManagedEntry(entry, record))
				continue
			}
			result.States = append(result.States, classifyUnownedObservedEntry(entry, desiredByName))
		}
	}

	sort.SliceStable(result.States, func(i, j int) bool {
		left, right := result.States[i], result.States[j]
		leftRank, rightRank := projectTargetRank(left.Target), projectTargetRank(right.Target)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if nameOrder := compareUTF16(left.Skill, right.Skill); nameOrder != 0 {
			return nameOrder < 0
		}
		return compareUTF16(left.Path, right.Path) < 0
	})
	return detachProjectClassification(result), nil
}

type projectDesiredPlacement struct {
	skill  DesiredSkill
	target Target
}

func desiredPlacements(desiredByName map[string]DesiredSkill) []projectDesiredPlacement {
	placements := make([]projectDesiredPlacement, 0)
	for _, skill := range desiredByName {
		for _, target := range skill.Targets {
			placements = append(placements, projectDesiredPlacement{skill: skill, target: target})
		}
	}
	sort.SliceStable(placements, func(i, j int) bool {
		left, right := placements[i], placements[j]
		if projectTargetRank(left.target) != projectTargetRank(right.target) {
			return projectTargetRank(left.target) < projectTargetRank(right.target)
		}
		return compareUTF16(left.skill.Name, right.skill.Name) < 0
	})
	return placements
}

func classifyCopyPlacement(state *ProjectState, skill DesiredSkill, root ProjectSkillsRootInventory, records map[string]ProvenanceRecord, expected TreeHash) {
	if !root.Safe {
		state.Kind = ProjectStateProtected
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonRootUnavailable
		return
	}
	var entry *InventoryEntry
	for index := range root.Entries {
		if root.Entries[index].Name == skill.Name {
			entry = &root.Entries[index]
			break
		}
	}
	if entry == nil {
		state.Kind = ProjectStateMissing
		state.Action = PlanActionInstall
		state.Reason = ProjectStateReasonExpectedEntryAbsent
		return
	}
	if entry.Kind == InventoryEntrySymlink {
		state.Kind = ProjectStateMisplaced
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonExpectedCopyIsSymlink
		return
	}
	if entry.Problem != "" || entry.Kind != InventoryEntryDirectory || entry.Hash == nil {
		state.Kind = ProjectStateMalformed
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonCurrentEntryUnverifiable
		state.Current = cloneTreeHashPointer(entry.Hash)
		return
	}
	state.Current = cloneTreeHashPointer(entry.Hash)
	record, managed := records[projectPlacementKey(root.Target, skill.Name)]
	if !managed {
		state.Kind = ProjectStateUnmanaged
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonDesiredPathUnmanaged
		return
	}
	if record.SourceIdentity != state.SourceIdentity {
		state.Kind = ProjectStateProtected
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonProvenanceSourceMismatch
		return
	}
	if treeHashMatchesRecord(*entry.Hash, record) {
		if entry.Hash.Algorithm == expected.Algorithm && entry.Hash.Digest == expected.Digest {
			state.Kind = ProjectStateExact
			state.Action = PlanActionUnchanged
			state.Reason = ProjectStateReasonVerifiedExact
			return
		}
		state.Kind = ProjectStateOutdated
		state.Action = PlanActionUpdate
		state.Reason = ProjectStateReasonVerifiedUpdate
		return
	}
	state.Kind = ProjectStateModified
	state.Action = PlanActionBlocked
	state.Reason = ProjectStateReasonLocalModification
}

func classifyRemovedManagedEntry(entry InventoryEntry, record ProvenanceRecord) ProjectState {
	state := ProjectState{
		Target:         entry.Target,
		Skill:          entry.Name,
		Path:           entry.Path,
		Manager:        ManagerSkillsCLI,
		SourceIdentity: record.SourceIdentity,
		Expected:       treeHashPointerFromRecord(record),
	}
	if entry.Problem == "" && entry.Kind == InventoryEntryDirectory && entry.Hash != nil {
		state.Current = cloneTreeHashPointer(entry.Hash)
		if treeHashMatchesRecord(*entry.Hash, record) {
			state.Kind = ProjectStateOutdated
			state.Action = PlanActionQuarantine
			state.Reason = ProjectStateReasonPreviouslyManagedNotDesired
			return state
		}
	}
	state.Kind = ProjectStateModified
	state.Action = PlanActionBlocked
	state.Reason = ProjectStateReasonPreviouslyManagedModified
	return state
}

func classifyUnownedObservedEntry(entry InventoryEntry, desiredByName map[string]DesiredSkill) ProjectState {
	state := ProjectState{
		Target: entry.Target,
		Skill:  entry.Name,
		Path:   entry.Path,
	}
	if entry.Problem != "" || entry.Kind != InventoryEntryDirectory || entry.Hash == nil {
		state.Kind = ProjectStateMalformed
		state.Action = PlanActionBlocked
		state.Reason = ProjectStateReasonMalformedEntryPreserved
		state.Current = cloneTreeHashPointer(entry.Hash)
		return state
	}
	if desired, known := desiredByName[entry.Name]; known {
		state.Kind = ProjectStateMisplaced
		state.Action = PlanActionUnchanged
		state.Manager = desired.Manager
		if desired.Manager == ManagerSkillsCLI {
			state.SourceIdentity, _ = canonicalProjectSourceIdentity(desired.Source)
		}
		state.Reason = ProjectStateReasonUnmanagedMisplacedPreserved
		state.Current = cloneTreeHashPointer(entry.Hash)
		return state
	}
	state.Kind = ProjectStateUnmanaged
	state.Action = PlanActionUnchanged
	state.Reason = ProjectStateReasonUnmanagedEntryPreserved
	state.Current = cloneTreeHashPointer(entry.Hash)
	return state
}

func validateProjectClassificationInputs(desired DesiredState, expected map[string]TreeHash, inventory ProjectInventory) (map[string]DesiredSkill, map[string]TreeHash, map[string]ProvenanceRecord, map[Target]ProjectSkillsRootInventory, []Issue) {
	var issues []Issue
	desiredByName := make(map[string]DesiredSkill, len(desired.Skills))
	if desired.Scope != ScopeProject {
		issues = append(issues, Issue{Code: IssueMalformedInput, Path: "desired.scope", Message: "must be project scope"})
	}
	for index, skill := range desired.Skills {
		path := "desired.skills[" + strconv.Itoa(index) + "]"
		if skill.Name == "" || !isPortableName(skill.Name) {
			issues = append(issues, Issue{Code: IssueInvalidName, Path: path + ".name", Message: "must be a portable skill name"})
		}
		if _, exists := desiredByName[skill.Name]; exists {
			issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".name", Message: "duplicate desired skill name"})
		} else {
			desiredByName[skill.Name] = skill
		}
		if skill.Scope != ScopeProject {
			issues = append(issues, Issue{Code: IssueMalformedInput, Path: path + ".scope", Message: "must be project scope"})
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
		switch skill.Manager {
		case ManagerSkillsCLI:
			if skill.Mode != ModeCopy {
				issues = append(issues, Issue{Code: IssueInvalidMode, Path: path + ".mode", Message: "skills-cli installation must use copy mode"})
			}
			if _, ok := canonicalProjectSourceIdentity(skill.Source); !ok {
				issues = append(issues, Issue{Code: IssueInvalidSource, Path: path + ".source", Message: "must be a credential-free remote source"})
			}
		case ManagerManual, ManagerWorkflow:
		case ManagerNone:
			issues = append(issues, Issue{Code: IssueInvalidManager, Path: path + ".manager", Message: "manager none is invalid"})
		default:
			issues = append(issues, Issue{Code: IssueInvalidManager, Path: path + ".manager", Message: "unsupported manager"})
		}
	}

	expectedByName := make(map[string]TreeHash, len(expected))
	for name, hash := range expected {
		skill, desiredName := desiredByName[name]
		if !desiredName {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "expected." + name, Message: "expected hash has no desired skill"})
			continue
		}
		if skill.Manager != ManagerSkillsCLI {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "expected." + name, Message: "expected hash requires a skills-cli desired skill"})
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

	rootsByTarget := make(map[Target]ProjectSkillsRootInventory, len(inventory.Roots))
	layout, layoutErr := LayoutForProject(inventory.Root)
	if layoutErr != nil || layout.Root != inventory.Root {
		issues = append(issues, Issue{Code: IssueInvalidRoot, Path: "inventory.root", Message: "inventory root must be the exact absolute project root"})
	} else {
		if inventory.StatePath != layout.ReconcilerStatePath {
			issues = append(issues, Issue{Code: IssuePathEscape, Path: "inventory.statePath", Message: "inventory state path does not match the project layout"})
		}
		for index, root := range inventory.Roots {
			path := "inventory.roots[" + strconv.Itoa(index) + "]"
			if _, supported := supportedTargets[root.Target]; !supported {
				issues = append(issues, Issue{Code: IssueInvalidTarget, Path: path + ".target", Message: "unsupported inventory target"})
			}
			if _, duplicate := rootsByTarget[root.Target]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path + ".target", Message: "duplicate inventory target"})
			} else {
				rootsByTarget[root.Target] = root
			}
			expectedPath, _ := layout.ManagedSkillsPath(root.Target)
			if expectedPath == "" || root.Path != expectedPath {
				issues = append(issues, Issue{Code: IssuePathEscape, Path: path + ".path", Message: "inventory root path does not match the project layout"})
			}
			seenNames := make(map[string]struct{}, len(root.Entries))
			for entryIndex, entry := range root.Entries {
				entryPath := path + ".entries[" + strconv.Itoa(entryIndex) + "]"
				if entry.Target != root.Target {
					issues = append(issues, Issue{Code: IssueInvalidTarget, Path: entryPath + ".target", Message: "entry target does not match its root"})
				}
				expectedEntryPath := filepath.Join(root.Path, entry.Name)
				if entry.Path != expectedEntryPath {
					issues = append(issues, Issue{Code: IssuePathEscape, Path: entryPath + ".path", Message: "inventory entry path does not match its root and name"})
				}
				if _, duplicate := seenNames[entry.Name]; duplicate {
					issues = append(issues, Issue{Code: IssueDuplicate, Path: entryPath + ".name", Message: "duplicate inventory entry name"})
				}
				seenNames[entry.Name] = struct{}{}
			}
		}
	}
	for _, target := range []Target{TargetAgents, TargetClaude} {
		if _, ok := rootsByTarget[target]; !ok {
			issues = append(issues, Issue{Code: IssueMissingReference, Path: "inventory.roots", Message: "inventory must contain exactly one root for each target"})
		}
	}

	recordsByKey := make(map[string]ProvenanceRecord, len(inventory.State.Records))
	if inventory.StateTrusted {
		for index, record := range inventory.State.Records {
			path := "inventory.state.records[" + strconv.Itoa(index) + "]"
			key := projectPlacementKey(record.Target, record.Skill)
			if _, duplicate := recordsByKey[key]; duplicate {
				issues = append(issues, Issue{Code: IssueDuplicate, Path: path, Message: "duplicate trusted provenance placement"})
			} else {
				recordsByKey[key] = record
			}
			if !isCanonicalProjectSourceIdentity(record.SourceIdentity) {
				issues = append(issues, Issue{Code: IssueInvalidSource, Path: path + ".sourceIdentity", Message: "trusted provenance source identity is unsafe"})
			}
		}
	}
	return desiredByName, expectedByName, recordsByKey, rootsByTarget, issues
}

func projectPlacementKey(target Target, skill string) string { return string(target) + "\x00" + skill }

func projectTargetRank(target Target) int {
	switch target {
	case TargetAgents:
		return 0
	case TargetClaude:
		return 1
	default:
		return 2
	}
}

func treeHashMatchesRecord(current TreeHash, record ProvenanceRecord) bool {
	return current.Algorithm == record.TreeHashAlgorithm && current.Digest == record.TreeHash
}

func treeHashPointerFromRecord(record ProvenanceRecord) *TreeHash {
	hash := TreeHash{Algorithm: record.TreeHashAlgorithm, Digest: record.TreeHash}
	return &hash
}

func cloneTreeHashPointer(hash *TreeHash) *TreeHash {
	if hash == nil {
		return nil
	}
	copyHash := *hash
	return &copyHash
}

func detachProjectClassification(classification ProjectClassification) ProjectClassification {
	classification.States = append([]ProjectState(nil), classification.States...)
	for index := range classification.States {
		classification.States[index].Current = cloneTreeHashPointer(classification.States[index].Current)
		classification.States[index].Expected = cloneTreeHashPointer(classification.States[index].Expected)
	}
	if classification.States == nil {
		classification.States = []ProjectState{}
	}
	return classification
}

// canonicalProjectSourceIdentity accepts only source forms that are safe to
// compare as ownership evidence.  GitHub URLs intentionally use repository
// identity like the legacy reconciler; shorthand catalog suffixes remain part
// of the identity so two shorthand paths cannot silently claim one another.
func canonicalProjectSourceIdentity(source string) (string, bool) {
	if source == "" || strings.TrimSpace(source) != source || strings.ContainsRune(source, 0) {
		return "", false
	}
	if !strings.Contains(source, "://") {
		parts := strings.Split(source, "/")
		if len(parts) < 2 {
			return "", false
		}
		for _, part := range parts {
			if part == "" || part == "." || part == ".." || !projectSourceSegmentPattern.MatchString(part) {
				return "", false
			}
		}
		owner := strings.ToLower(parts[0])
		repository := strings.ToLower(strings.TrimSuffix(parts[1], ".git"))
		if repository == "" || repository == "." || repository == ".." {
			return "", false
		}
		identity := []string{owner, repository}
		for _, part := range parts[2:] {
			identity = append(identity, strings.ToLower(part))
		}
		return "github:" + strings.Join(identity, "/"), true
	}

	parsed, err := url.Parse(source)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", false
	}
	port := parsed.Port()
	if port == "" && strings.HasSuffix(parsed.Host, ":") {
		return "", false
	}
	if port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return "", false
		}
		port = strconv.Itoa(portNumber)
	}
	hostPort := host
	if port != "" {
		if strings.Contains(host, ":") {
			hostPort = "[" + host + "]"
		}
		hostPort += ":" + port
	}
	path, err := url.PathUnescape(parsed.Path)
	if err != nil || strings.ContainsAny(path, "\x00\r\n\t ") {
		return "", false
	}
	if path == "" || strings.Trim(path, "/") == "" {
		if host == "github.com" && (port == "" || port == "443") {
			return "", false
		}
		return "https://" + hostPort, true
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	if host == "github.com" && (port == "" || port == "443") {
		if len(parts) < 2 {
			return "", false
		}
		if !projectSourceSegmentPattern.MatchString(parts[0]) || !projectSourceSegmentPattern.MatchString(parts[1]) {
			return "", false
		}
		repository := strings.TrimSuffix(strings.ToLower(parts[1]), ".git")
		if repository == "" || repository == "." || repository == ".." {
			return "", false
		}
		return "github:" + strings.ToLower(parts[0]) + "/" + repository, true
	}
	cleanPath := strings.TrimRight(path, "/")
	cleanPath = strings.TrimSuffix(cleanPath, ".git")
	cleanPath = strings.TrimRight(cleanPath, "/")
	if cleanPath == "" {
		return "", false
	}
	return "https://" + hostPort + cleanPath, true
}

func isCanonicalProjectSourceIdentity(identity string) bool {
	if strings.HasPrefix(identity, "github:") {
		rest := strings.TrimPrefix(identity, "github:")
		if !projectCanonicalIdentityPattern.MatchString(rest) || strings.ToLower(rest) != rest {
			return false
		}
		for _, part := range strings.Split(rest, "/") {
			if part == "." || part == ".." {
				return false
			}
		}
		return true
	}
	canonical, ok := canonicalProjectSourceIdentity(identity)
	return ok && canonical == identity
}
