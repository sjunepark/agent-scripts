package sjskills

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// GlobalLocationKind explains why a modeled global path is not a managed
// skills root. These paths are observable only as bounded migration evidence;
// the global planner never enumerates protected locations.
type GlobalLocationKind string

const (
	GlobalLocationVendorMetadata GlobalLocationKind = "vendor-metadata"
	GlobalLocationBackup         GlobalLocationKind = "backup"
	GlobalLocationLegacy         GlobalLocationKind = "legacy"
	GlobalLocationCache          GlobalLocationKind = "cache"
	GlobalLocationRuntime        GlobalLocationKind = "runtime"
)

type GlobalProtectedLocation struct {
	ID      string                 `json:"id"`
	Kind    GlobalLocationKind     `json:"kind"`
	Path    string                 `json:"path"`
	Exists  bool                   `json:"exists"`
	Safe    bool                   `json:"safe"`
	Problem InventoryProblemReason `json:"problem,omitempty"`
}

type GlobalLegacyRootInventory struct {
	ID       string             `json:"id"`
	Kind     GlobalLocationKind `json:"kind"`
	Path     string             `json:"path"`
	Exists   bool               `json:"exists"`
	Safe     bool               `json:"safe"`
	Entries  []InventoryEntry   `json:"entries"`
	Problems []InventoryProblem `json:"problems,omitempty"`
}

type GlobalProvenanceState struct {
	Version int                `json:"version"`
	Records []ProvenanceRecord `json:"records"`
}

type GlobalProvenanceFormat string

const (
	GlobalProvenanceMissing  GlobalProvenanceFormat = "missing"
	GlobalProvenanceLegacyV1 GlobalProvenanceFormat = "legacy-v1"
	GlobalProvenanceCurrent  GlobalProvenanceFormat = "current-v2"
)

// GlobalInventory is the complete read-only fixed-home boundary. Roots always
// contain exactly the two managed targets. LegacyRoot is enumerated only for
// migration reporting, while Protected is never enumerated.
type GlobalInventory struct {
	Home              string                       `json:"home"`
	Roots             []ProjectSkillsRootInventory `json:"roots"`
	StatePath         string                       `json:"statePath"`
	State             GlobalProvenanceState        `json:"state"`
	StateTrusted      bool                         `json:"stateTrusted"`
	ProvenanceFormat  GlobalProvenanceFormat       `json:"provenanceFormat"`
	MigrationRequired bool                         `json:"migrationRequired"`
	LegacyRoot        GlobalLegacyRootInventory    `json:"legacyRoot"`
	Protected         []GlobalProtectedLocation    `json:"protected"`
	Problems          []InventoryProblem           `json:"problems,omitempty"`
}

// InspectGlobal inventories only paths named by GlobalLayout. It never scans
// the selected home, follows a root symlink, or enumerates a protected path.
func InspectGlobal(layout GlobalLayout) (GlobalInventory, error) {
	home, err := establishGlobalInspectionBoundary(layout)
	if err != nil {
		return GlobalInventory{}, err
	}
	// The OS home may be reached through a stable platform alias such as
	// macOS /var -> /private/var. Inventory uses the canonical boundary and a
	// freshly derived layout so every descendant proof shares one identity.
	layout, err = LayoutForGlobal(home)
	if err != nil {
		return GlobalInventory{}, err
	}
	inventory := GlobalInventory{
		Home:             home,
		Roots:            make([]ProjectSkillsRootInventory, 0, 2),
		StatePath:        layout.ProvenanceStatePath,
		State:            emptyGlobalProvenanceState(),
		StateTrusted:     true,
		ProvenanceFormat: GlobalProvenanceMissing,
		Problems:         make([]InventoryProblem, 0),
		Protected:        make([]GlobalProtectedLocation, 0, 5),
	}
	for _, modeled := range []struct {
		target Target
		path   string
	}{
		{target: TargetAgents, path: layout.AgentsSkillsPath},
		{target: TargetClaude, path: layout.ClaudeSkillsPath},
	} {
		root := inspectProjectSkillsRoot(home, modeled.target, modeled.path)
		inventory.Roots = append(inventory.Roots, root)
		inventory.Problems = append(inventory.Problems, root.Problems...)
	}

	state, trusted, format, stateProblems := inspectGlobalProvenanceState(home, layout.ProvenanceStatePath)
	inventory.State = state
	inventory.StateTrusted = trusted
	inventory.ProvenanceFormat = format
	inventory.MigrationRequired = trusted && format == GlobalProvenanceLegacyV1
	inventory.Problems = append(inventory.Problems, stateProblems...)

	inventory.LegacyRoot = inspectGlobalLegacyRoot(home, layout.LegacyPiSkillsPath)
	inventory.Problems = append(inventory.Problems, inventory.LegacyRoot.Problems...)
	for _, protected := range []struct {
		id        string
		kind      GlobalLocationKind
		path      string
		directory bool
	}{
		{id: "agents-vendor-lock", kind: GlobalLocationVendorMetadata, path: layout.AgentsVendorLockPath},
		{id: "claude-vendor-lock", kind: GlobalLocationVendorMetadata, path: layout.ClaudeVendorLockPath},
		{id: "legacy-quarantine", kind: GlobalLocationBackup, path: layout.LegacyQuarantinePath, directory: true},
		{id: "codex-plugin-cache", kind: GlobalLocationCache, path: layout.CodexPluginCachePath, directory: true},
		{id: "codex-system", kind: GlobalLocationRuntime, path: layout.CodexSystemSkillsPath, directory: true},
	} {
		location := inspectGlobalProtectedLocation(home, protected.id, protected.kind, protected.path, protected.directory)
		inventory.Protected = append(inventory.Protected, location)
		if location.Problem != "" {
			inventory.Problems = append(inventory.Problems, InventoryProblem{Path: location.Path, Reason: location.Problem})
		}
	}
	return detachGlobalInventory(inventory), nil
}

func establishGlobalInspectionBoundary(layout GlobalLayout) (string, error) {
	expected, err := LayoutForGlobal(layout.Home)
	if err != nil {
		return "", err
	}
	if !sameInspectionPath(layout.Home, expected.Home) ||
		!sameInspectionPath(layout.AgentsSkillsPath, expected.AgentsSkillsPath) ||
		!sameInspectionPath(layout.ClaudeSkillsPath, expected.ClaudeSkillsPath) ||
		!sameInspectionPath(layout.ProvenanceStatePath, expected.ProvenanceStatePath) ||
		!sameInspectionPath(layout.DerivedStatePath, expected.DerivedStatePath) ||
		!sameInspectionPath(layout.QuarantinePath, expected.QuarantinePath) ||
		!sameInspectionPath(layout.AgentsVendorLockPath, expected.AgentsVendorLockPath) ||
		!sameInspectionPath(layout.ClaudeVendorLockPath, expected.ClaudeVendorLockPath) ||
		!sameInspectionPath(layout.LegacyPiSkillsPath, expected.LegacyPiSkillsPath) ||
		!sameInspectionPath(layout.LegacyQuarantinePath, expected.LegacyQuarantinePath) ||
		!sameInspectionPath(layout.CodexSystemSkillsPath, expected.CodexSystemSkillsPath) ||
		!sameInspectionPath(layout.CodexPluginCachePath, expected.CodexPluginCachePath) {
		return "", inspectionBoundaryError(IssuePathEscape, "global layout does not match the selected home")
	}
	info, err := os.Lstat(layout.Home)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", inspectionBoundaryError(IssueInvalidRoot, "global home must be an available real directory")
	}
	canonical, err := filepath.EvalSymlinks(layout.Home)
	if err != nil {
		return "", inspectionBoundaryError(IssueInvalidRoot, "global home canonical identity is unavailable")
	}
	return filepath.Clean(canonical), nil
}

func inspectGlobalLegacyRoot(home, path string) GlobalLegacyRootInventory {
	result := GlobalLegacyRootInventory{ID: "pi", Kind: GlobalLocationLegacy, Path: path, Entries: []InventoryEntry{}, Problems: []InventoryProblem{}}
	probe := proveInspectionPath(home, path)
	if !probe.exists && !probe.unsafe {
		result.Safe = true
		return result
	}
	result.Exists = probe.exists
	if probe.unsafe || probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.IsDir() {
		result.addProblem(probeReasonOr(probe.reason, InventoryProblemRootNotDirectory))
		return result
	}
	result.Safe = true
	children, err := os.ReadDir(path)
	if err != nil {
		result.Safe = false
		result.addProblem(InventoryProblemRootUnreadable)
		return result
	}
	sort.SliceStable(children, func(i, j int) bool { return compareUTF16(children[i].Name(), children[j].Name()) < 0 })
	for _, child := range children {
		entry := inspectProjectEntry("", filepath.Join(path, child.Name()), child.Name())
		result.Entries = append(result.Entries, entry)
		if entry.Problem != "" {
			result.Problems = append(result.Problems, InventoryProblem{Name: entry.Name, Path: entry.Path, Reason: entry.Problem})
		}
	}
	return result
}

func (root *GlobalLegacyRootInventory) addProblem(reason InventoryProblemReason) {
	root.Safe = false
	root.Problems = append(root.Problems, InventoryProblem{Path: root.Path, Reason: reason})
}

func probeReasonOr(reason, fallback InventoryProblemReason) InventoryProblemReason {
	if reason != "" {
		return reason
	}
	return fallback
}

func inspectGlobalProtectedLocation(home, id string, kind GlobalLocationKind, path string, directory bool) GlobalProtectedLocation {
	result := GlobalProtectedLocation{ID: id, Kind: kind, Path: path}
	probe := proveInspectionPath(home, path)
	if !probe.exists && !probe.unsafe {
		result.Safe = true
		return result
	}
	result.Exists = probe.exists
	if probe.unsafe || probe.info == nil {
		result.Problem = probeReasonOr(probe.reason, InventoryProblemRootUnreadable)
		return result
	}
	if probe.info.Mode()&os.ModeSymlink != 0 || (directory && !probe.info.IsDir()) || (!directory && !probe.info.Mode().IsRegular()) {
		result.Problem = InventoryProblemRootNotDirectory
		return result
	}
	result.Safe = true
	return result
}

func inspectGlobalProvenanceState(home, path string) (GlobalProvenanceState, bool, GlobalProvenanceFormat, []InventoryProblem) {
	missing := emptyGlobalProvenanceState()
	probe := proveInspectionPath(home, path)
	if !probe.exists && !probe.unsafe {
		return missing, true, GlobalProvenanceMissing, nil
	}
	problem := InventoryProblem{Path: path}
	if probe.unsafe {
		problem.Reason = InventoryProblemStateUnsafeAncestor
		return GlobalProvenanceState{}, false, "", []InventoryProblem{problem}
	}
	if probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		problem.Reason = InventoryProblemStateNotRegular
		return GlobalProvenanceState{}, false, "", []InventoryProblem{problem}
	}
	data, err := readBoundedInspectionFile(path, maxProvenanceStateBytes)
	if err != nil {
		problem.Reason = InventoryProblemStateUnreadable
		return GlobalProvenanceState{}, false, "", []InventoryProblem{problem}
	}
	state, format, valid := decodeGlobalProvenanceState(data)
	if !valid {
		problem.Reason = InventoryProblemStateInvalid
		return GlobalProvenanceState{}, false, "", []InventoryProblem{problem}
	}
	return state, true, format, nil
}

func emptyGlobalProvenanceState() GlobalProvenanceState {
	return GlobalProvenanceState{Version: GlobalProvenanceStateVersion, Records: []ProvenanceRecord{}}
}

func decodeGlobalProvenanceState(data []byte) (GlobalProvenanceState, GlobalProvenanceFormat, bool) {
	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return GlobalProvenanceState{}, "", false
	}
	switch header.Version {
	case GlobalProvenanceStateVersion:
		var state GlobalProvenanceState
		if !decodeStrictJSON(data, &state) || !validGlobalProvenanceRecords(state.Records) {
			return GlobalProvenanceState{}, "", false
		}
		state.Records = sortedProvenanceRecords(state.Records)
		return state, GlobalProvenanceCurrent, true
	case 1:
		var legacy struct {
			Version int `json:"version"`
			Records []struct {
				Root          string    `json:"root"`
				Skill         string    `json:"skill"`
				Source        string    `json:"source"`
				HashAlgorithm string    `json:"hashAlgorithm"`
				Hash          string    `json:"hash"`
				RecordedAt    time.Time `json:"recordedAt"`
			} `json:"records"`
		}
		if !decodeStrictJSON(data, &legacy) || legacy.Records == nil {
			return GlobalProvenanceState{}, "", false
		}
		state := emptyGlobalProvenanceState()
		seen := make(map[string]struct{}, len(legacy.Records))
		for _, record := range legacy.Records {
			var target Target
			switch record.Root {
			case "shared":
				target = TargetAgents
			case "claude":
				target = TargetClaude
			default:
				return GlobalProvenanceState{}, "", false
			}
			identity, ok := canonicalProjectSourceIdentity(record.Source)
			key := projectPlacementKey(target, record.Skill)
			if !ok || !isPortableName(record.Skill) || record.HashAlgorithm != TreeHashAlgorithmSHA256V2 || !lowercaseDigestPattern.MatchString(record.Hash) || record.RecordedAt.IsZero() {
				return GlobalProvenanceState{}, "", false
			}
			if _, duplicate := seen[key]; duplicate {
				return GlobalProvenanceState{}, "", false
			}
			seen[key] = struct{}{}
			state.Records = append(state.Records, ProvenanceRecord{
				Scope: ScopeGlobal, Skill: record.Skill, Target: target, SourceIdentity: identity,
				TreeHashAlgorithm: record.HashAlgorithm, TreeHash: record.Hash, RecordedAt: record.RecordedAt,
			})
		}
		state.Records = sortedProvenanceRecords(state.Records)
		return state, GlobalProvenanceLegacyV1, true
	default:
		return GlobalProvenanceState{}, "", false
	}
}

func decodeStrictJSON(data []byte, value any) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return false
	}
	var trailing any
	return decoder.Decode(&trailing) == io.EOF
}

func validGlobalProvenanceRecords(records []ProvenanceRecord) bool {
	if records == nil {
		return false
	}
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		key := projectPlacementKey(record.Target, record.Skill)
		if record.Scope != ScopeGlobal || !isPortableName(record.Skill) ||
			(record.Target != TargetAgents && record.Target != TargetClaude) ||
			!isCanonicalProjectSourceIdentity(record.SourceIdentity) ||
			record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 ||
			!lowercaseDigestPattern.MatchString(record.TreeHash) || record.RecordedAt.IsZero() {
			return false
		}
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}

func sortedProvenanceRecords(records []ProvenanceRecord) []ProvenanceRecord {
	result := cloneProvenanceRecords(records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Target != result[j].Target {
			return projectTargetRank(result[i].Target) < projectTargetRank(result[j].Target)
		}
		return compareUTF16(result[i].Skill, result[j].Skill) < 0
	})
	return result
}

func detachGlobalInventory(inventory GlobalInventory) GlobalInventory {
	inventory.Roots = cloneInventoryRoots(inventory.Roots)
	for index := range inventory.Roots {
		inventory.Roots[index].Entries = cloneInventoryEntries(inventory.Roots[index].Entries)
		inventory.Roots[index].Problems = cloneInventoryProblems(inventory.Roots[index].Problems)
	}
	inventory.State.Records = cloneProvenanceRecords(inventory.State.Records)
	inventory.LegacyRoot.Entries = cloneInventoryEntries(inventory.LegacyRoot.Entries)
	inventory.LegacyRoot.Problems = cloneInventoryProblems(inventory.LegacyRoot.Problems)
	inventory.Protected = append([]GlobalProtectedLocation(nil), inventory.Protected...)
	inventory.Problems = cloneInventoryProblems(inventory.Problems)
	return inventory
}
