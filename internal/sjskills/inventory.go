package sjskills

// This file contains the read-only boundary between a canonical project and
// the later reconciliation planner.  It deliberately knows only about the
// two project skill roots and the reconciler provenance file.  In particular,
// it does not inspect a manifest, enumerate the project, or mutate anything.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const maxProvenanceStateBytes int64 = 1 << 20

// InventoryEntryKind is the only type classification exposed by project
// inventory.  A regular file and every other special entry intentionally
// share the "other" classification: neither is a managed skill directory.
type InventoryEntryKind string

const (
	InventoryEntryDirectory InventoryEntryKind = "directory"
	InventoryEntrySymlink   InventoryEntryKind = "symlink"
	InventoryEntryOther     InventoryEntryKind = "other"
)

// InventoryProblemReason is a bounded, stable reason consumed by the project
// classifier. It never contains an operating-system error or a filesystem
// path.  The path and (when applicable) observed name are separate typed
// fields on InventoryProblem.
type InventoryProblemReason string

const (
	InventoryProblemRootUnsafeAncestor       InventoryProblemReason = "root-unsafe-ancestor"
	InventoryProblemRootUnreadable           InventoryProblemReason = "root-unreadable"
	InventoryProblemRootOutsideProject       InventoryProblemReason = "root-outside-project"
	InventoryProblemRootNotDirectory         InventoryProblemReason = "root-not-directory"
	InventoryProblemRootCanonicalUnavailable InventoryProblemReason = "root-canonical-unavailable"
	InventoryProblemEntryUnreadable          InventoryProblemReason = "entry-unreadable"
	InventoryProblemEntryInvalidName         InventoryProblemReason = "entry-invalid-name"
	InventoryProblemEntryUnverifiable        InventoryProblemReason = "entry-unverifiable"
	InventoryProblemEntryUnsupportedType     InventoryProblemReason = "entry-unsupported-type"
	InventoryProblemSymlinkTargetUnreadable  InventoryProblemReason = "symlink-target-unreadable"
	InventoryProblemSymlinkNotHashed         InventoryProblemReason = "symlink-not-hashed"
	InventoryProblemStateUnsafeAncestor      InventoryProblemReason = "state-unsafe-ancestor"
	InventoryProblemStateUnreadable          InventoryProblemReason = "state-unreadable"
	InventoryProblemStateOutsideProject      InventoryProblemReason = "state-outside-project"
	InventoryProblemStateNotRegular          InventoryProblemReason = "state-not-regular"
	InventoryProblemStateInvalid             InventoryProblemReason = "state-invalid"
)

// InventoryProblem identifies one bounded inspection problem.  Path is the
// modeled path or immediate child path; Reason is always one of the stable
// constants above and never includes an underlying error string.
type InventoryProblem struct {
	Target Target                 `json:"target,omitempty"`
	Name   string                 `json:"name,omitempty"`
	Path   string                 `json:"path,omitempty"`
	Reason InventoryProblemReason `json:"reason"`
}

// InventoryEntry is one immediate child of a safe modeled skills root.
// Hash is non-nil only when a directory was hashed successfully with
// HashSkillTree.  Symlinks record only their textual target and are never
// followed or hashed by inventory.
type InventoryEntry struct {
	Target     Target                 `json:"target"`
	Name       string                 `json:"name"`
	Path       string                 `json:"path"`
	Kind       InventoryEntryKind     `json:"kind"`
	Hash       *TreeHash              `json:"hash,omitempty"`
	LinkTarget string                 `json:"linkTarget,omitempty"`
	Problem    InventoryProblemReason `json:"problem,omitempty"`
}

// ProjectSkillsRootInventory is the exact-state view of one modeled target.
// Safe is true for a missing root (absence is safe to classify later) and for
// an existing root whose complete ancestor chain and canonical identity were
// proven safe.  Unsafe roots have no entries because they are never read.
type ProjectSkillsRootInventory struct {
	Target   Target             `json:"target"`
	Path     string             `json:"path"`
	Exists   bool               `json:"exists"`
	Safe     bool               `json:"safe"`
	Entries  []InventoryEntry   `json:"entries"`
	Problems []InventoryProblem `json:"problems,omitempty"`
}

// ProjectInventory is the complete read-only project boundary.  Roots always
// contain exactly .agents/skills and .claude/skills, in target order.  State
// is trusted provenance only; malformed or unsafe state contributes a
// problem and leaves State.Records empty.  No other project path is modeled.
type ProjectInventory struct {
	Root         string                       `json:"root"`
	Roots        []ProjectSkillsRootInventory `json:"roots"`
	StatePath    string                       `json:"statePath"`
	State        ProvenanceState              `json:"state"`
	StateTrusted bool                         `json:"stateTrusted"`
	Problems     []InventoryProblem           `json:"problems,omitempty"`
}

// InspectProject establishes a canonical project boundary, inventories only
// the two modeled skills roots, and loads only the modeled provenance state.
// The supplied layout must be the exact path-only layout for its canonical
// root.  An invalid, unavailable, or non-canonical project root returns an
// error; all modeled-root/state failures are represented in the inventory so
// the safe portions remain available to later classification.
func InspectProject(layout DerivedLayout) (ProjectInventory, error) {
	root, err := establishInspectionBoundary(layout)
	if err != nil {
		return ProjectInventory{}, err
	}

	inventory := ProjectInventory{
		Root:         root,
		Roots:        make([]ProjectSkillsRootInventory, 0, 2),
		StatePath:    layout.ReconcilerStatePath,
		State:        emptyTrustedProvenanceState(),
		StateTrusted: true,
		Problems:     make([]InventoryProblem, 0),
	}

	for _, modeled := range []struct {
		target Target
		path   string
	}{
		{target: TargetAgents, path: layout.AgentsSkillsPath},
		{target: TargetClaude, path: layout.ClaudeSkillsPath},
	} {
		rootInventory := inspectProjectSkillsRoot(root, modeled.target, modeled.path)
		inventory.Roots = append(inventory.Roots, rootInventory)
		inventory.Problems = append(inventory.Problems, rootInventory.Problems...)
	}

	state, trusted, stateProblems := inspectProvenanceState(root, layout.ReconcilerStatePath)
	inventory.State = state
	inventory.StateTrusted = trusted
	inventory.Problems = append(inventory.Problems, stateProblems...)
	return detachProjectInventory(inventory), nil
}

// RootFor returns a detached copy of one target inventory.
func (inventory ProjectInventory) RootFor(target Target) (ProjectSkillsRootInventory, bool) {
	for _, root := range inventory.Roots {
		if root.Target == target {
			root.Entries = cloneInventoryEntries(root.Entries)
			root.Problems = cloneInventoryProblems(root.Problems)
			return root, true
		}
	}
	return ProjectSkillsRootInventory{}, false
}

// Records returns a detached copy of the trusted provenance records.
func (inventory ProjectInventory) Records() []ProvenanceRecord {
	return cloneProvenanceRecords(inventory.State.Records)
}

func establishInspectionBoundary(layout DerivedLayout) (string, error) {
	if err := validateLayoutRootOnly(layout.Root); err != nil {
		return "", err
	}
	cleanRoot := filepath.Clean(layout.Root)
	expected, err := LayoutForProject(cleanRoot)
	if err != nil {
		return "", err
	}
	if !sameInspectionPath(cleanRoot, expected.Root) ||
		!sameInspectionPath(layout.Root, expected.Root) ||
		!sameInspectionPath(layout.ManifestPath, expected.ManifestPath) ||
		!sameInspectionPath(layout.AgentsSkillsPath, expected.AgentsSkillsPath) ||
		!sameInspectionPath(layout.ClaudeSkillsPath, expected.ClaudeSkillsPath) ||
		!sameInspectionPath(layout.DerivedDirectoryPath, expected.DerivedDirectoryPath) ||
		!sameInspectionPath(layout.ReconcilerStatePath, expected.ReconcilerStatePath) ||
		!sameInspectionPath(layout.QuarantinePath, expected.QuarantinePath) {
		return "", inspectionBoundaryError(IssuePathEscape, "layout does not match the project root")
	}

	info, err := os.Lstat(cleanRoot)
	if err != nil {
		return "", inspectionBoundaryError(IssueInvalidRoot, "project root is unavailable")
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", inspectionBoundaryError(IssueInvalidRoot, "project root must be a real directory")
	}
	canonical, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil {
		return "", inspectionBoundaryError(IssueInvalidRoot, "project root canonical identity is unavailable")
	}
	if !sameInspectionPath(canonical, cleanRoot) {
		return "", inspectionBoundaryError(IssueInvalidRoot, "project root must be canonical")
	}
	return cleanRoot, nil
}

func inspectionBoundaryError(code IssueCode, message string) error {
	return &ValidationErrors{Issues: []Issue{{Code: code, Path: "root", Message: message}}}
}

func sameInspectionPath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

type inspectedPath struct {
	exists    bool
	info      os.FileInfo
	canonical string
	unsafe    bool
	reason    InventoryProblemReason
}

// proveInspectionPath examines only the existing ancestors of candidate using
// Lstat.  Missing components are a safe absence.  No call to Stat, ReadDir,
// EvalSymlinks, or a file open happens until every existing ancestor is known
// to be a real directory inside boundary.
func proveInspectionPath(boundary, candidate string) inspectedPath {
	result := inspectedPath{}
	if !pathWithin(boundary, candidate) {
		result.unsafe = true
		result.reason = InventoryProblemRootOutsideProject
		return result
	}
	relative, err := filepath.Rel(filepath.Clean(boundary), filepath.Clean(candidate))
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		result.unsafe = true
		result.reason = InventoryProblemRootOutsideProject
		return result
	}
	if relative == "." {
		result.exists = true
		result.info, err = os.Lstat(boundary)
		if err != nil {
			result.unsafe = true
			result.reason = InventoryProblemRootUnreadable
			return result
		}
		if result.info.Mode()&os.ModeSymlink != 0 || !result.info.IsDir() {
			result.unsafe = true
			result.reason = InventoryProblemRootNotDirectory
			return result
		}
		result.canonical, err = filepath.EvalSymlinks(boundary)
		if err != nil {
			result.unsafe = true
			result.reason = InventoryProblemRootCanonicalUnavailable
			return result
		}
		if !pathWithin(boundary, result.canonical) {
			result.unsafe = true
			result.reason = InventoryProblemRootOutsideProject
		}
		return result
	}

	current := filepath.Clean(boundary)
	components := strings.Split(relative, string(filepath.Separator))
	for index, component := range components {
		current = filepath.Join(current, component)
		info, lstatErr := os.Lstat(current)
		if lstatErr != nil {
			if errors.Is(lstatErr, os.ErrNotExist) {
				// A missing ancestor means the modeled path is absent.  All
				// components already examined were safe real directories.
				return result
			}
			result.unsafe = true
			result.reason = InventoryProblemRootUnreadable
			return result
		}
		if info.Mode()&os.ModeSymlink != 0 {
			result.exists = index == len(components)-1
			result.info = info
			result.unsafe = true
			result.reason = InventoryProblemRootUnsafeAncestor
			return result
		}
		if index < len(components)-1 && !info.IsDir() {
			result.unsafe = true
			result.reason = InventoryProblemRootUnsafeAncestor
			return result
		}
		if index == len(components)-1 {
			result.exists = true
			result.info = info
		}
	}

	result.canonical, err = filepath.EvalSymlinks(candidate)
	if err != nil {
		result.unsafe = true
		result.reason = InventoryProblemRootCanonicalUnavailable
		return result
	}
	if !pathWithin(boundary, result.canonical) {
		result.unsafe = true
		result.reason = InventoryProblemRootOutsideProject
	}
	return result
}

func inspectProjectSkillsRoot(boundary string, target Target, path string) ProjectSkillsRootInventory {
	root := ProjectSkillsRootInventory{
		Target:   target,
		Path:     path,
		Entries:  make([]InventoryEntry, 0),
		Problems: make([]InventoryProblem, 0),
	}
	probe := proveInspectionPath(boundary, path)
	if !probe.exists && !probe.unsafe {
		root.Safe = true
		return root
	}
	root.Exists = probe.exists
	if probe.unsafe {
		root.Safe = false
		reason := probe.reason
		if reason == "" {
			reason = InventoryProblemRootUnsafeAncestor
		}
		root.Problems = append(root.Problems, InventoryProblem{Target: target, Path: path, Reason: reason})
		return root
	}
	if probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.IsDir() {
		root.Problems = append(root.Problems, InventoryProblem{
			Target: target, Path: path, Reason: InventoryProblemRootNotDirectory,
		})
		return root
	}
	if !pathWithin(boundary, probe.canonical) {
		root.Problems = append(root.Problems, InventoryProblem{
			Target: target, Path: path, Reason: InventoryProblemRootOutsideProject,
		})
		return root
	}
	root.Safe = true
	entries, readErr := os.ReadDir(path)
	if readErr != nil {
		root.Safe = false
		root.Problems = append(root.Problems, InventoryProblem{
			Target: target, Path: path, Reason: InventoryProblemRootUnreadable,
		})
		return root
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return compareUTF16(entries[i].Name(), entries[j].Name()) < 0
	})
	for _, directoryEntry := range entries {
		name := directoryEntry.Name()
		entryPath := filepath.Join(path, name)
		entry := inspectProjectEntry(target, entryPath, name)
		root.Entries = append(root.Entries, entry)
		if entry.Problem != "" {
			problem := InventoryProblem{Target: target, Name: name, Path: entryPath, Reason: entry.Problem}
			root.Problems = append(root.Problems, problem)
		}
	}
	return root
}

func inspectProjectEntry(target Target, path, name string) InventoryEntry {
	entry := InventoryEntry{
		Target: target,
		Name:   name,
		Path:   path,
		Kind:   InventoryEntryOther,
	}
	if !isPortableName(name) {
		entry.Problem = InventoryProblemEntryInvalidName
	}

	info, err := os.Lstat(path)
	if err != nil {
		if entry.Problem == "" {
			entry.Problem = InventoryProblemEntryUnreadable
		}
		return entry
	}
	if info.Mode()&os.ModeSymlink != 0 {
		entry.Kind = InventoryEntrySymlink
		linkTarget, readlinkErr := os.Readlink(path)
		if readlinkErr != nil {
			entry.Problem = InventoryProblemSymlinkTargetUnreadable
			return entry
		}
		entry.LinkTarget = linkTarget
		if entry.Problem == "" {
			entry.Problem = InventoryProblemSymlinkNotHashed
		}
		return entry
	}
	if info.IsDir() {
		entry.Kind = InventoryEntryDirectory
		hash, hashErr := HashSkillTree(path)
		if hashErr != nil {
			if entry.Problem == "" {
				entry.Problem = InventoryProblemEntryUnverifiable
			}
			return entry
		}
		entry.Hash = &hash
		return entry
	}
	if entry.Problem == "" {
		entry.Problem = InventoryProblemEntryUnsupportedType
	}
	return entry
}

func inspectProvenanceState(boundary, path string) (ProvenanceState, bool, []InventoryProblem) {
	missing := emptyTrustedProvenanceState()
	probe := proveInspectionPath(boundary, path)
	if !probe.exists && !probe.unsafe {
		return missing, true, nil
	}
	problem := InventoryProblem{Path: path}
	if probe.unsafe {
		switch probe.reason {
		case InventoryProblemRootOutsideProject:
			problem.Reason = InventoryProblemStateOutsideProject
		case InventoryProblemRootUnreadable, InventoryProblemRootCanonicalUnavailable:
			problem.Reason = InventoryProblemStateUnreadable
		default:
			problem.Reason = InventoryProblemStateUnsafeAncestor
		}
		return ProvenanceState{}, false, []InventoryProblem{problem}
	}
	if probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		problem.Reason = InventoryProblemStateNotRegular
		return ProvenanceState{}, false, []InventoryProblem{problem}
	}
	if !pathWithin(boundary, probe.canonical) {
		problem.Reason = InventoryProblemStateOutsideProject
		return ProvenanceState{}, false, []InventoryProblem{problem}
	}

	data, readErr := readBoundedInspectionFile(path, maxProvenanceStateBytes)
	if readErr != nil {
		problem.Reason = InventoryProblemStateUnreadable
		return ProvenanceState{}, false, []InventoryProblem{problem}
	}
	state, valid := decodeProvenanceState(data)
	if !valid {
		problem.Reason = InventoryProblemStateInvalid
		return ProvenanceState{}, false, []InventoryProblem{problem}
	}
	return state, true, nil
}

func emptyTrustedProvenanceState() ProvenanceState {
	return ProvenanceState{Version: ProvenanceStateVersion, Records: []ProvenanceRecord{}}
}

func readBoundedInspectionFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	reader := io.LimitReader(file, limit+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("state exceeds inspection bound")
	}
	return data, nil
}

var lowercaseDigestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func decodeProvenanceState(data []byte) (ProvenanceState, bool) {
	var state ProvenanceState
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return ProvenanceState{}, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ProvenanceState{}, false
	}
	if state.Version != ProvenanceStateVersion || state.Records == nil {
		return ProvenanceState{}, false
	}

	records := append([]ProvenanceRecord(nil), state.Records...)
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.Scope != ScopeProject ||
			record.Skill == "" || !isPortableName(record.Skill) ||
			(record.Target != TargetAgents && record.Target != TargetClaude) ||
			record.SourceIdentity == "" ||
			record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 ||
			!lowercaseDigestPattern.MatchString(record.TreeHash) ||
			record.RecordedAt.IsZero() {
			return ProvenanceState{}, false
		}
		key := string(record.Target) + "\x00" + record.Skill
		if _, duplicate := seen[key]; duplicate {
			return ProvenanceState{}, false
		}
		seen[key] = struct{}{}
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Target != records[j].Target {
			return records[i].Target < records[j].Target
		}
		return records[i].Skill < records[j].Skill
	})
	state.Records = records
	return state, true
}

func detachProjectInventory(inventory ProjectInventory) ProjectInventory {
	inventory.Roots = cloneInventoryRoots(inventory.Roots)
	for index := range inventory.Roots {
		inventory.Roots[index].Entries = cloneInventoryEntries(inventory.Roots[index].Entries)
		inventory.Roots[index].Problems = cloneInventoryProblems(inventory.Roots[index].Problems)
	}
	inventory.State.Records = cloneProvenanceRecords(inventory.State.Records)
	inventory.Problems = cloneInventoryProblems(inventory.Problems)
	return inventory
}

func cloneInventoryRoots(values []ProjectSkillsRootInventory) []ProjectSkillsRootInventory {
	if values == nil {
		return nil
	}
	result := make([]ProjectSkillsRootInventory, len(values))
	copy(result, values)
	return result
}

func cloneInventoryEntries(values []InventoryEntry) []InventoryEntry {
	if values == nil {
		return nil
	}
	result := make([]InventoryEntry, len(values))
	copy(result, values)
	for index := range result {
		if values[index].Hash != nil {
			hash := *values[index].Hash
			result[index].Hash = &hash
		}
	}
	return result
}

func cloneInventoryProblems(values []InventoryProblem) []InventoryProblem {
	if values == nil {
		return nil
	}
	result := make([]InventoryProblem, len(values))
	copy(result, values)
	return result
}

func cloneProvenanceRecords(values []ProvenanceRecord) []ProvenanceRecord {
	if values == nil {
		return nil
	}
	result := make([]ProvenanceRecord, len(values))
	copy(result, values)
	return result
}
