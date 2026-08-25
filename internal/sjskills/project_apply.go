package sjskills

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	applyLockName       = "apply.lock"
	applyLockMarker     = "sjskills-project-lock-v1\n"
	applyInstallPattern = ".sjskills-install-"
	applyManifestName   = "manifest.json"
)

// ApplyFailureKind is the stable result class for a reconciliation
// transaction. The implementation deliberately does not expose filesystem
// or operating-system diagnostics: callers can report this error safely.
type ApplyFailureKind string

const (
	ApplyFailureUnavailable ApplyFailureKind = "unavailable"
	ApplyFailureConflict    ApplyFailureKind = "conflict"
)

// ApplyError is returned for an execution boundary failure.  Reason is a
// bounded vocabulary, not an underlying error or path.
type ApplyError struct {
	Kind   ApplyFailureKind
	Reason string
	Scope  Scope
}

func (e *ApplyError) Error() string {
	if e == nil {
		return "project apply failed"
	}
	scope := "project"
	if e.Scope == ScopeGlobal {
		scope = "global"
	}
	if e.Kind == ApplyFailureConflict {
		return scope + " apply conflict: " + e.Reason
	}
	return scope + " apply unavailable: " + e.Reason
}

func (e *ApplyError) Conflict() bool { return e != nil && e.Kind == ApplyFailureConflict }

func applyUnavailable(reason string) error {
	return &ApplyError{Kind: ApplyFailureUnavailable, Reason: reason}
}

func applyConflict(reason string) error {
	return &ApplyError{Kind: ApplyFailureConflict, Reason: reason}
}

// ProjectApplySession is the reviewed project plan plus its still-live,
// verified materialization session. ApplyProjectChanges owns the session
// only for the duration of the call; the CLI cleans it on every return path.
type ProjectApplySession struct {
	Layout       DerivedLayout
	Desired      DesiredState
	Plan         Plan
	Expected     map[string]TreeHash
	Materialized *MaterializationPlan
	global       *globalApplyBoundary
}

// GlobalApplySession is the reviewed fixed-baseline plan plus its verified
// materialization session. It uses the same transaction engine as projects,
// but carries the registry and fixed-home boundary needed for locked replans.
type GlobalApplySession struct {
	Layout       GlobalLayout
	Registry     Registry
	Desired      DesiredState
	Plan         Plan
	Expected     map[string]TreeHash
	Materialized *MaterializationPlan
	reviewedPlan *globalReviewedPlanBinding
}

type globalReviewedPlanBinding struct {
	artifactSHA256     string
	sessionFingerprint string
}

// BindReviewedPlan proves that this exact live materialization session
// reproduces the reviewed global plan before the session can cross the global
// mutation boundary.
func (s *GlobalApplySession) BindReviewedPlan(reviewed ReviewedPlan, current Envelope) error {
	if s == nil {
		return &Issue{Code: IssueUnavailable, Path: "apply.approvedPlan", Message: "global apply session is unavailable"}
	}
	if !reviewed.verified {
		return &Issue{Code: IssueUnavailable, Path: "apply.approvedPlan", Message: "reviewed plan binding is unavailable"}
	}
	if err := VerifyReviewedPlanRecheck(reviewed, current); err != nil {
		return err
	}
	binding, err := newGlobalReviewedPlanBinding(s, reviewed.sha256)
	if err != nil {
		return err
	}
	s.reviewedPlan = binding
	return nil
}

func newGlobalReviewedPlanBinding(s *GlobalApplySession, artifactSHA256 string) (*globalReviewedPlanBinding, error) {
	digest, err := normalizeApprovedSHA256(artifactSHA256)
	if err != nil {
		return nil, &Issue{Code: IssueUnavailable, Path: "apply.approvedPlan", Message: "reviewed plan binding is unavailable"}
	}
	fingerprint, err := globalApplySessionFingerprint(s)
	if err != nil {
		return nil, &Issue{Code: IssueUnavailable, Path: "apply.approvedPlan", Message: "global apply session could not be bound"}
	}
	return &globalReviewedPlanBinding{artifactSHA256: digest, sessionFingerprint: fingerprint}, nil
}

func globalApplySessionFingerprint(s *GlobalApplySession) (string, error) {
	if s == nil || s.Materialized == nil {
		return "", errors.New("global apply session is unavailable")
	}
	type namedHash struct {
		Name string   `json:"name"`
		Hash TreeHash `json:"hash"`
	}
	type boundSnapshot struct {
		Name  string       `json:"name"`
		Skill DesiredSkill `json:"skill"`
		Hash  TreeHash     `json:"hash"`
	}
	type boundSession struct {
		Layout    GlobalLayout        `json:"layout"`
		Registry  Registry            `json:"registry"`
		Desired   DesiredState        `json:"desired"`
		Plan      reviewedPlanContent `json:"plan"`
		Expected  []namedHash         `json:"expected"`
		Snapshots []boundSnapshot     `json:"snapshots"`
		Skipped   []DesiredSkill      `json:"skipped"`
	}
	expectedNames := make([]string, 0, len(s.Expected))
	for name := range s.Expected {
		expectedNames = append(expectedNames, name)
	}
	sort.Strings(expectedNames)
	expected := make([]namedHash, 0, len(expectedNames))
	for _, name := range expectedNames {
		expected = append(expected, namedHash{Name: name, Hash: s.Expected[name]})
	}
	snapshotNames := make([]string, 0, len(s.Materialized.snapshots))
	for name := range s.Materialized.snapshots {
		snapshotNames = append(snapshotNames, name)
	}
	sort.Strings(snapshotNames)
	snapshots := make([]boundSnapshot, 0, len(snapshotNames))
	for _, name := range snapshotNames {
		snapshot := s.Materialized.snapshots[name]
		if snapshot == nil {
			return "", errors.New("global materialization snapshot is unavailable")
		}
		snapshots = append(snapshots, boundSnapshot{Name: name, Skill: snapshot.Skill, Hash: snapshot.Hash})
	}
	payload := boundSession{
		Layout: s.Layout, Registry: s.Registry, Desired: cloneDesiredState(s.Desired),
		Plan: reviewedPlanContentFor(s.Plan), Expected: expected, Snapshots: snapshots,
		Skipped: s.Materialized.Skipped(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

type globalApplyBoundary struct {
	layout   GlobalLayout
	registry Registry
}

// AppliedPlacement is deliberately path-free so callers cannot accidentally
// report project or staging paths as plan evidence.
type AppliedPlacement struct {
	Skill  string
	Target Target
}

type ApplyResult struct {
	Plan        Plan
	Installed   []AppliedPlacement
	Updated     []AppliedPlacement
	Quarantined []AppliedPlacement
	Quarantine  *ProjectQuarantineResult
	Migrated    bool
}

// ProjectQuarantineResult is the path-free recovery handle returned when an
// update transaction created durable quarantine state.
type ProjectQuarantineResult struct {
	ID     string                  `json:"id"`
	Status ProjectQuarantineStatus `json:"status"`
}

// ApplyDeps contains the small set of platform and fault-injection seams.
// Ordinary bounded filesystem reads and copies remain in this package.
type ApplyDeps struct {
	Now                      func() time.Time
	MakeTempDir              func(parent, pattern string) (string, error)
	PublishNoReplace         func(source, destination string) error
	ReplaceFileAtomic        func(source, destination string) error
	SyncFile                 func(*os.File) error
	SyncDir                  func(path string) error
	BeforePublish            func(AppliedPlacement) error
	beforeQuarantine         func(AppliedPlacement) error
	newQuarantineID          func() (string, error)
	beforeRollback           func(AppliedPlacement) error
	beforeLock               func() error
	beforeCommit             func() error
	beforeUnlock             func() error
	beforeRestoreMove        func(AppliedPlacement) error
	beforeRestoreRollback    func(AppliedPlacement) error
	beforeRestoreManifest    func(ProjectQuarantineStatus) error
	beforeRestoreCommit      func() error
	publishStateNoReplace    func(source, destination string) error
	afterStateCommit         func()
	beforeStateRecoveryMove  func()
	afterQuarantineRunSync   func()
	beforeRecoveryTreeMove   func()
	afterRecoveryTreeHash    func()
	afterInitialManifestSync func()
	beforeJournalClear       func()
	afterJournalMove         func()
}

func defaultApplyDeps() ApplyDeps {
	return ApplyDeps{
		Now:                   time.Now,
		MakeTempDir:           os.MkdirTemp,
		PublishNoReplace:      publishNoReplace,
		ReplaceFileAtomic:     replaceFileAtomic,
		SyncFile:              func(file *os.File) error { return file.Sync() },
		SyncDir:               syncApplyDirectory,
		newQuarantineID:       newProjectQuarantineID,
		publishStateNoReplace: publishNoReplace,
	}
}

// ApplyProjectChanges is the project mutation seam. It validates the
// reviewed plan before taking the lock, refreshes inventory and translation,
// then copies verified snapshots into empty destinations, quarantines an
// unchanged managed preimage before verified replacement, or moves an
// unchanged managed placement whose desired identity was removed into durable
// quarantine. It commits one deterministic provenance state.
func ApplyProjectChanges(ctx context.Context, session *ProjectApplySession, deps ApplyDeps) (ApplyResult, error) {
	if session != nil && session.global != nil {
		return ApplyResult{}, applyUnavailable("project apply requires project scope")
	}
	return applyChanges(ctx, session, deps)
}

// ApplyGlobalChanges applies only the fixed desired baseline. Former profile
// placements remain report-only migration evidence and are never converted to
// removal operations by this boundary.
func ApplyGlobalChanges(ctx context.Context, session *GlobalApplySession, deps ApplyDeps) (ApplyResult, error) {
	if session == nil {
		return ApplyResult{}, scopeApplyFailure(applyUnavailable("materialization session is unavailable"), ScopeGlobal)
	}
	if session.reviewedPlan == nil {
		return ApplyResult{}, scopeApplyFailure(applyUnavailable("reviewed plan binding is unavailable"), ScopeGlobal)
	}
	if _, err := normalizeApprovedSHA256(session.reviewedPlan.artifactSHA256); err != nil {
		return ApplyResult{}, scopeApplyFailure(applyUnavailable("reviewed plan binding is unavailable"), ScopeGlobal)
	}
	fingerprint, err := globalApplySessionFingerprint(session)
	if err != nil {
		return ApplyResult{}, scopeApplyFailure(applyUnavailable("reviewed plan binding could not be verified"), ScopeGlobal)
	}
	if fingerprint != session.reviewedPlan.sessionFingerprint {
		return ApplyResult{}, scopeApplyFailure(applyConflict("reviewed plan binding changed"), ScopeGlobal)
	}
	internal := &ProjectApplySession{
		Layout:       session.Layout.mutationLayout(),
		Desired:      session.Desired,
		Plan:         session.Plan,
		Expected:     session.Expected,
		Materialized: session.Materialized,
		global:       &globalApplyBoundary{layout: session.Layout, registry: session.Registry},
	}
	result, err := applyChanges(ctx, internal, deps)
	return result, scopeApplyFailure(err, ScopeGlobal)
}

func scopeApplyFailure(err error, scope Scope) error {
	if err == nil {
		return nil
	}
	var failure *ApplyError
	if !errors.As(err, &failure) {
		return err
	}
	return &ApplyError{Kind: failure.Kind, Reason: failureReasonForScope(failure.Reason, scope), Scope: scope}
}

func failureReasonForScope(reason string, scope Scope) string {
	if scope != ScopeGlobal {
		return reason
	}
	return strings.NewReplacer(
		"another project apply", "another global apply",
		"project state", "global state",
		"project root", "global home",
		"project boundary", "global boundary",
		"project transaction", "global transaction",
		"project lock", "global lock",
		"project placement", "global placement",
		"the project", "the global home",
	).Replace(reason)
}

func applyChanges(ctx context.Context, session *ProjectApplySession, deps ApplyDeps) (ApplyResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateApplySession(session); err != nil {
		return ApplyResult{}, err
	}
	if err := contextErr(ctx); err != nil {
		return ApplyResult{}, applyUnavailable("context cancelled")
	}
	deps = normalizeApplyDeps(deps)
	rootInfo, err := applyRootInfo(session.Layout.Root)
	if err != nil {
		return ApplyResult{Plan: clonePlanForApply(session.Plan)}, err
	}

	lock, createdDerived, err := acquireApplyLock(session.Layout, rootInfo, deps)
	if err != nil {
		if cleanupErr := cleanupApplyDirs(createdDerived); cleanupErr != nil {
			return ApplyResult{Plan: clonePlanForApply(session.Plan)}, cleanupErr
		}
		return ApplyResult{Plan: clonePlanForApply(session.Plan)}, err
	}
	tx := applyTransaction{
		root:        session.Layout.Root,
		rootInfo:    rootInfo,
		ancestors:   make(map[Target][]*applyAncestor),
		createdDirs: createdDerived,
		published:   make([]publishedPlacement, 0),
		quarantined: make([]quarantinedPlacement, 0),
		lock:        lock,
		layout:      session.Layout,
		deps:        deps,
		scope:       session.Desired.Scope,
		global:      session.global,
	}
	if err := tx.checkLock(); err != nil {
		return finishApplySetup(&tx, clonePlanForApply(session.Plan), err)
	}
	if recovered, recoveryErr := tx.recoverInterruptedTransaction(); recoveryErr != nil {
		return finishApplySetup(&tx, clonePlanForApply(session.Plan), recoveryErr)
	} else if recovered {
		return finishApplySetup(&tx, clonePlanForApply(session.Plan), applyUnavailable("interrupted project transaction was recovered; rerun apply"))
	}
	freshPlan, err := refreshApplyPlan(session)
	if err != nil {
		return finishApplySetup(&tx, clonePlanForApply(session.Plan), err)
	}
	if !sameReviewedPlan(session.Plan, freshPlan, session.global != nil) {
		return finishApplySetup(&tx, freshPlan, applyConflict("project state changed after planning"))
	}
	reviewedMigration := session.global != nil && planHasEvidence(session.Plan, "provenance-migration")
	mutationOperations := reviewedMutationOperations(&ProjectApplySession{Plan: freshPlan})
	stateOnlyMigration := session.global != nil && planHasEvidence(freshPlan, "provenance-migration")
	if reviewedMigration != stateOnlyMigration {
		return finishApplySetup(&tx, freshPlan, applyConflict("global provenance migration changed after planning"))
	}
	if len(mutationOperations) == 0 && !stateOnlyMigration {
		if tx.deps.beforeCommit != nil {
			if hookErr := tx.deps.beforeCommit(); hookErr != nil {
				return finishApplySetup(&tx, freshPlan, applyUnavailable("apply verification preflight failed"))
			}
		}
		if err := tx.checkLock(); err != nil {
			return finishApplySetup(&tx, freshPlan, err)
		}
		verifiedPlan, err := refreshApplyPlan(session)
		if err != nil {
			return finishApplySetup(&tx, freshPlan, err)
		}
		if !sameReviewedPlan(freshPlan, verifiedPlan, session.global != nil) {
			return finishApplySetup(&tx, verifiedPlan, applyConflict("project state changed during apply verification"))
		}
		return finishApplySetup(&tx, verifiedPlan, nil)
	}
	if err := tx.syncCreatedDirectoryParents(createdDerived); err != nil {
		return finishApplySetup(&tx, freshPlan, err)
	}
	ancestors, err := captureApplyAncestors(session.Layout, mutationOperations)
	if err != nil {
		return finishApplySetup(&tx, freshPlan, err)
	}
	tx.ancestors = ancestors
	preimage, err := captureTransactionStatePreimage(session.Layout, session.Desired.Scope)
	if err != nil {
		return finishApplySetup(&tx, freshPlan, err)
	}
	result := ApplyResult{Plan: freshPlan, Installed: []AppliedPlacement{}, Updated: []AppliedPlacement{}, Quarantined: []AppliedPlacement{}}
	var primary error
	quarantineOperations := operationsWithActions(mutationOperations, PlanActionUpdate, PlanActionQuarantine)
	if len(quarantineOperations) > 0 {
		primary = tx.planQuarantine(session, preimage, quarantineOperations)
	}
	if primary == nil {
		primary = tx.prepareTransactionJournal(preimage, session)
	}
	if primary == nil && tx.quarantine != nil {
		primary = tx.publishPreparedQuarantine(quarantineOperations)
	}
	if primary == nil {
		primary = tx.applyOperations(ctx, session, preimage, mutationOperations)
	}
	var commit *stateCommit
	if primary == nil {
		commit, primary = tx.commitState(preimage, session)
	}
	if primary == nil && tx.deps.afterStateCommit != nil {
		tx.deps.afterStateCommit()
	}
	if primary == nil {
		primary = tx.commitQuarantine()
	}
	if primary == nil {
		primary = tx.clearTransactionJournal()
	}
	if primary == nil && stateOnlyMigration {
		result.Migrated = true
	}
	recoveryRequired := false
	if primary != nil && !tx.journalCleared && commit != nil && commit.replaced {
		if restoreErr := tx.restoreState(preimage, commit); restoreErr != nil {
			primary = applyUnavailable("provenance restoration could not be verified")
			recoveryRequired = true
		}
	}
	if primary != nil && !tx.journalCleared {
		if rollbackErr := tx.rollback(); rollbackErr != nil {
			var applyErr *ApplyError
			if !errors.As(primary, &applyErr) || !applyErr.Conflict() || !strings.Contains(applyErr.Reason, "lock changed") {
				primary = applyUnavailable("rollback could not be verified")
			}
			recoveryRequired = true
		}
	}
	if primary != nil && !tx.journalCleared && !recoveryRequired {
		if journalErr := tx.clearTransactionJournal(); journalErr != nil {
			recoveryRequired = true
		}
	}
	if recoveryRequired && tx.quarantine != nil && tx.quarantine.runInfo != nil {
		if statusErr := tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired); statusErr != nil {
			tx.quarantine.manifest.Status = ProjectQuarantineRecoveryRequired
		}
	}
	if tx.quarantine != nil && tx.quarantine.runInfo != nil {
		result.Quarantine = &ProjectQuarantineResult{ID: tx.quarantine.manifest.ID, Status: tx.quarantine.manifest.Status}
	}
	if primary == nil {
		for _, placement := range tx.published {
			value := AppliedPlacement{Skill: placement.skill, Target: placement.target}
			if placement.action == PlanActionUpdate {
				result.Updated = append(result.Updated, value)
			} else {
				result.Installed = append(result.Installed, value)
			}
		}
		for _, placement := range tx.quarantined {
			if placement.action == PlanActionQuarantine {
				result.Quarantined = append(result.Quarantined, AppliedPlacement{Skill: placement.skill, Target: placement.target})
			}
		}
		if tx.deps.beforeUnlock != nil {
			if hookErr := tx.deps.beforeUnlock(); hookErr != nil {
				primary = applyUnavailable("apply finalization preflight failed")
			}
		}
	}
	if unlockErr := tx.closeLock(); unlockErr != nil && primary == nil {
		primary = unlockErr
	}
	if cleanupErr := tx.cleanupCreatedDirs(primary == nil); cleanupErr != nil && primary == nil {
		primary = cleanupErr
	}
	if primary != nil {
		return result, primary
	}
	return result, nil
}

func planHasEvidence(plan Plan, kind string) bool {
	for _, evidence := range plan.Evidence {
		if evidence.Kind == kind {
			return true
		}
	}
	return false
}

func normalizeApplyDeps(deps ApplyDeps) ApplyDeps {
	defaults := defaultApplyDeps()
	if deps.Now == nil {
		deps.Now = defaults.Now
	}
	if deps.MakeTempDir == nil {
		deps.MakeTempDir = defaults.MakeTempDir
	}
	if deps.PublishNoReplace == nil {
		deps.PublishNoReplace = defaults.PublishNoReplace
	}
	if deps.ReplaceFileAtomic == nil {
		deps.ReplaceFileAtomic = defaults.ReplaceFileAtomic
	}
	if deps.SyncFile == nil {
		deps.SyncFile = defaults.SyncFile
	}
	if deps.SyncDir == nil {
		deps.SyncDir = defaults.SyncDir
	}
	if deps.newQuarantineID == nil {
		deps.newQuarantineID = defaults.newQuarantineID
	}
	if deps.publishStateNoReplace == nil {
		deps.publishStateNoReplace = defaults.publishStateNoReplace
	}
	return deps
}

func validateApplySession(session *ProjectApplySession) error {
	if session == nil {
		return applyUnavailable("materialization session is unavailable")
	}
	scope := session.Desired.Scope
	if scope != ScopeProject && scope != ScopeGlobal {
		return applyUnavailable("apply requires a supported scope")
	}
	if scope == ScopeProject && session.global != nil {
		return applyConflict("project apply boundary is inconsistent")
	}
	if scope == ScopeGlobal {
		if session.global == nil {
			return applyUnavailable("global apply requires the fixed global boundary")
		}
		resolved, err := ResolveGlobal(session.global.registry)
		if err != nil || !sameDesiredState(resolved, session.Desired) {
			return applyConflict("global desired identity changed")
		}
		expectedLayout, err := LayoutForGlobal(session.global.layout.Home)
		if err != nil || !reflect.DeepEqual(expectedLayout, session.global.layout) || !reflect.DeepEqual(session.Layout, expectedLayout.mutationLayout()) {
			return applyConflict("global boundary changed")
		}
	}
	if session.Plan.Desired.Scope != session.Desired.Scope || !sameDesiredState(session.Plan.Desired, session.Desired) {
		return applyConflict("reviewed desired identity changed")
	}
	if len(session.Expected) == 0 && hasSkillsCLIPlacement(session.Desired) {
		return applyUnavailable("expected materialized content is unavailable")
	}
	for _, skill := range session.Desired.Skills {
		if skill.Manager != ManagerSkillsCLI {
			continue
		}
		if skill.Mode != ModeCopy {
			return applyConflict("desired placement is not copy-owned")
		}
		expected, exists := session.Expected[skill.Name]
		if !exists {
			return applyUnavailable("expected materialized content is unavailable")
		}
		if session.Materialized == nil {
			return applyUnavailable("materialization session is unavailable")
		}
		snapshot, exists := session.Materialized.SnapshotFor(skill.Name)
		if !exists || snapshot == nil || !sameDesiredSkill(skill, snapshot.Skill) || snapshot.Hash != expected {
			return applyConflict("verified materialized snapshot identity changed")
		}
		if err := snapshot.Verify(); err != nil {
			return applyConflict("verified materialized snapshot changed")
		}
		if err := validateApplyCopyTree(snapshot.Path); err != nil {
			return err
		}
	}
	desired := desiredByPlacement(session.Desired)
	seen := make(map[string]struct{}, len(session.Plan.Operations))
	for _, operation := range session.Plan.Operations {
		key := projectPlacementKey(operation.Target, operation.Skill)
		skill, desiredPlacement := desired[key]
		switch operation.Action {
		case PlanActionUnchanged, PlanActionManual, PlanActionWorkflow:
			continue
		case PlanActionInstall, PlanActionUpdate:
			if !desiredPlacement || skill.Manager != ManagerSkillsCLI || skill.Mode != ModeCopy {
				return applyConflict("reviewed mutation has no copy-owned desired identity")
			}
			if operation.Manager != ManagerSkillsCLI || operation.SourceID != skill.SourceID || operation.Source != skill.Source {
				return applyConflict("reviewed mutation identity changed")
			}
			if operation.Expected.Kind != projectEvidenceTreeHash || operation.Expected.Detail != expectedEvidence(session.Expected[skill.Name]) {
				return applyConflict("reviewed expected content identity changed")
			}
			if operation.Action == PlanActionUpdate {
				if _, ok := treeHashFromPlanEvidence(operation.Current); !ok {
					return applyConflict("reviewed update preimage identity is invalid")
				}
			}
		case PlanActionQuarantine:
			if desiredPlacement {
				return applyConflict("reviewed removal still has a desired identity")
			}
			if operation.Manager != ManagerSkillsCLI || operation.SourceID != "" {
				return applyConflict("reviewed removal identity changed")
			}
			if !isCanonicalProjectSourceIdentity(operation.Source) {
				return applyConflict("reviewed removal source identity is not canonical")
			}
			oldHash, ok := treeHashFromPlanEvidence(operation.Current)
			if !ok {
				return applyConflict("reviewed removal preimage identity is invalid")
			}
			expectedHash, expectedOK := treeHashFromPlanEvidence(operation.Expected)
			if !expectedOK || expectedHash != oldHash {
				return applyConflict("reviewed removal provenance identity is invalid")
			}
		default:
			return applyConflict("reviewed plan contains an unsupported mutation")
		}
		if _, duplicate := seen[key]; duplicate {
			return applyConflict("reviewed plan contains a duplicate placement")
		}
		seen[key] = struct{}{}
	}
	for _, skill := range session.Desired.Skills {
		for _, target := range skill.Targets {
			key := projectPlacementKey(target, skill.Name)
			if skill.Manager == ManagerSkillsCLI {
				if _, exists := findPlanOperation(session.Plan.Operations, key); !exists {
					return applyConflict("reviewed plan is missing a desired placement")
				}
			}
		}
	}
	return nil
}

func validateApplyCopyTree(root string) error {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return applyConflict("materialized snapshot is not a real directory")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return applyUnavailable("materialized snapshot could not be read")
	}
	for _, entry := range entries {
		if !isApplyEntryName(entry.Name()) {
			return applyConflict("materialized snapshot contains an invalid entry name")
		}
		path := filepath.Join(root, entry.Name())
		entryInfo, err := os.Lstat(path)
		if err != nil {
			return applyConflict("materialized snapshot changed during validation")
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return applyConflict("materialized snapshot contains a symlink")
		}
		if entryInfo.IsDir() {
			if err := validateApplyCopyTree(path); err != nil {
				return err
			}
			continue
		}
		if !entryInfo.Mode().IsRegular() {
			return applyConflict("materialized snapshot contains a special file")
		}
	}
	return nil
}

func hasSkillsCLIPlacement(desired DesiredState) bool {
	for _, skill := range desired.Skills {
		if skill.Manager == ManagerSkillsCLI {
			return true
		}
	}
	return false
}

func desiredByPlacement(desired DesiredState) map[string]DesiredSkill {
	result := make(map[string]DesiredSkill)
	for _, skill := range desired.Skills {
		for _, target := range skill.Targets {
			result[projectPlacementKey(target, skill.Name)] = skill
		}
	}
	return result
}

func candidateRecordByKey(state ProvenanceState, key string) (ProvenanceRecord, bool) {
	for _, record := range state.Records {
		if projectPlacementKey(record.Target, record.Skill) == key {
			return record, true
		}
	}
	return ProvenanceRecord{}, false
}

func reviewedMutationOperations(session *ProjectApplySession) []PlanOperation {
	result := make([]PlanOperation, 0)
	for _, operation := range session.Plan.Operations {
		if operation.Action == PlanActionInstall || operation.Action == PlanActionUpdate || operation.Action == PlanActionQuarantine {
			result = append(result, operation)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if projectTargetRank(result[i].Target) != projectTargetRank(result[j].Target) {
			return projectTargetRank(result[i].Target) < projectTargetRank(result[j].Target)
		}
		return compareUTF16(result[i].Skill, result[j].Skill) < 0
	})
	return result
}

func operationsWithActions(operations []PlanOperation, actions ...PlanAction) []PlanOperation {
	result := make([]PlanOperation, 0)
	for _, operation := range operations {
		for _, action := range actions {
			if operation.Action == action {
				result = append(result, operation)
				break
			}
		}
	}
	return result
}

func expectedEvidence(hash TreeHash) string {
	return hash.Algorithm + ":" + hash.Digest
}

func treeHashFromPlanEvidence(evidence PlanEvidence) (TreeHash, bool) {
	const prefix = TreeHashAlgorithmSHA256V2 + ":"
	if evidence.Kind != projectEvidenceTreeHash || !strings.HasPrefix(evidence.Detail, prefix) {
		return TreeHash{}, false
	}
	digest := strings.TrimPrefix(evidence.Detail, prefix)
	if !lowercaseDigestPattern.MatchString(digest) {
		return TreeHash{}, false
	}
	return TreeHash{Algorithm: TreeHashAlgorithmSHA256V2, Digest: digest}, true
}

func findPlanOperation(operations []PlanOperation, key string) (PlanOperation, bool) {
	for _, operation := range operations {
		if projectPlacementKey(operation.Target, operation.Skill) == key {
			return operation, true
		}
	}
	return PlanOperation{}, false
}

func refreshApplyPlan(session *ProjectApplySession) (Plan, error) {
	if session.global != nil {
		inventory, err := InspectGlobal(session.global.layout)
		if err != nil {
			return Plan{}, applyConflict("global boundary changed after planning")
		}
		classification, err := ClassifyGlobal(session.global.registry, session.Plan.Desired, session.Expected, inventory)
		if err != nil {
			return Plan{}, applyConflict("global classification changed after planning")
		}
		base := clonePlanForApply(session.Plan)
		base.Operations = nil
		base.Warnings = staticGlobalWarnings(base.Warnings)
		base.Evidence = staticGlobalEvidence(base.Evidence)
		plan, err := TranslateGlobalClassification(base, classification)
		if err != nil {
			return Plan{}, applyConflict("global translation changed after planning")
		}
		return plan, nil
	}
	inventory, err := InspectProject(session.Layout)
	if err != nil {
		return Plan{}, applyConflict("project boundary changed after planning")
	}
	classification, err := ClassifyProject(session.Plan.Desired, session.Expected, inventory)
	if err != nil {
		return Plan{}, applyConflict("project classification changed after planning")
	}
	base := clonePlanForApply(session.Plan)
	base.Operations = nil
	base.Warnings = staticProjectWarnings(base.Warnings)
	plan, err := TranslateProjectClassification(base, classification)
	if err != nil {
		return Plan{}, applyConflict("project translation changed after planning")
	}
	return plan, nil
}

func staticGlobalEvidence(evidence []Evidence) []Evidence {
	result := make([]Evidence, 0, len(evidence))
	for _, item := range evidence {
		switch item.Kind {
		case "global-inventory", "provenance-migration":
			continue
		default:
			result = append(result, item)
		}
	}
	return result
}

func staticGlobalWarnings(warnings []Warning) []Warning {
	result := make([]Warning, 0, len(warnings))
	for _, warning := range warnings {
		switch warning.Code {
		case "global-state", "global-migration", "global-protected", "legacy-preserved", "legacy-protected", "unmanaged-preserved":
			continue
		default:
			result = append(result, warning)
		}
	}
	return result
}

func staticProjectWarnings(warnings []Warning) []Warning {
	result := make([]Warning, 0, len(warnings))
	for _, warning := range warnings {
		if warning.Code == "project-state" || warning.Code == "unmanaged-preserved" {
			continue
		}
		result = append(result, warning)
	}
	return result
}

func sameReviewedPlan(left, right Plan, includeDiagnostics bool) bool {
	leftContent := reviewedPlanContentFor(left)
	rightContent := reviewedPlanContentFor(right)
	if !includeDiagnostics {
		leftContent.Warnings, leftContent.Evidence = nil, nil
		rightContent.Warnings, rightContent.Evidence = nil, nil
	}
	leftData, leftErr := json.Marshal(leftContent)
	rightData, rightErr := json.Marshal(rightContent)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftData, rightData)
}

type reviewedPlanContent struct {
	Desired    DesiredState    `json:"desired"`
	Operations []PlanOperation `json:"operations"`
	Warnings   []Warning       `json:"warnings"`
	Evidence   []Evidence      `json:"evidence"`
}

func reviewedPlanContentFor(plan Plan) reviewedPlanContent {
	clone := clonePlanForApply(plan)
	return reviewedPlanContent{
		Desired: clone.Desired, Operations: clone.Operations,
		Warnings: clone.Warnings, Evidence: clone.Evidence,
	}
}

func sameDesiredState(left, right DesiredState) bool {
	leftData, leftErr := json.Marshal(left)
	rightData, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftData, rightData)
}

func clonePlanForApply(plan Plan) Plan {
	copyPlan := plan
	copyPlan.Desired = cloneDesiredState(plan.Desired)
	copyPlan.Operations = append([]PlanOperation(nil), plan.Operations...)
	copyPlan.Warnings = append([]Warning(nil), plan.Warnings...)
	copyPlan.Evidence = append([]Evidence(nil), plan.Evidence...)
	return copyPlan
}

func contextErr(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

type applyAncestor struct {
	path   string
	info   os.FileInfo
	exists bool
	owned  bool
}

type applyLock struct {
	file os.FileInfo
	hand *os.File
	path string
	held bool
}

type publishedPlacement struct {
	skill    string
	target   Target
	action   PlanAction
	dest     string
	info     os.FileInfo
	expected TreeHash
}

type quarantinedPlacement struct {
	skill         string
	target        Target
	action        PlanAction
	dest          string
	quarantined   string
	originalInfo  os.FileInfo
	oldHash       TreeHash
	manifestIndex int
}

type projectQuarantineTransaction struct {
	runPath      string
	runDurable   bool
	rootInfo     os.FileInfo
	runInfo      os.FileInfo
	ancestors    map[string]os.FileInfo
	manifestPath string
	manifestInfo os.FileInfo
	manifestData []byte
	manifest     ProjectQuarantineManifest
}

type applyStatePreimage struct {
	exists bool
	data   []byte
	state  ProvenanceState
	info   os.FileInfo
	mode   os.FileMode
}

type stateCommit struct {
	replaced bool
	data     []byte
	info     os.FileInfo
}

type applyTransaction struct {
	root           string
	rootInfo       os.FileInfo
	ancestors      map[Target][]*applyAncestor
	createdDirs    []*applyAncestor
	published      []publishedPlacement
	quarantined    []quarantinedPlacement
	quarantine     *projectQuarantineTransaction
	journal        *projectTransactionJournal
	journalData    []byte
	journalInfo    os.FileInfo
	journalCleared bool
	lock           *applyLock
	layout         DerivedLayout
	deps           ApplyDeps
	scope          Scope
	global         *globalApplyBoundary
}

func applyRootInfo(root string) (os.FileInfo, error) {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, applyConflict("project root changed after planning")
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil || !sameInspectionPath(canonical, filepath.Clean(root)) {
		return nil, applyConflict("project root changed after planning")
	}
	return info, nil
}

func captureApplyAncestors(layout DerivedLayout, operations []PlanOperation) (map[Target][]*applyAncestor, error) {
	result := make(map[Target][]*applyAncestor)
	seen := make(map[Target]struct{})
	for _, operation := range operations {
		if _, exists := seen[operation.Target]; exists {
			continue
		}
		seen[operation.Target] = struct{}{}
		path, err := layout.ManagedSkillsPath(operation.Target)
		if err != nil {
			return nil, applyConflict("target placement path changed after planning")
		}
		ancestors, err := captureApplyPathChain(layout.Root, path)
		if err != nil {
			return nil, err
		}
		result[operation.Target] = ancestors
	}
	return result, nil
}

func captureApplyPathChain(root, candidate string) ([]*applyAncestor, error) {
	if !pathWithin(root, candidate) {
		return nil, applyConflict("target placement path escapes the project")
	}
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, applyConflict("target placement path escapes the project")
	}
	current := filepath.Clean(root)
	parts := strings.Split(relative, string(filepath.Separator))
	result := make([]*applyAncestor, 0, len(parts))
	for index, part := range parts {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				result = append(result, &applyAncestor{path: current})
				for remaining := index + 1; remaining < len(parts); remaining++ {
					current = filepath.Join(current, parts[remaining])
					result = append(result, &applyAncestor{path: current})
				}
				return result, nil
			}
			return nil, applyConflict("target ancestor could not be inspected")
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, applyConflict("target ancestor is not a real directory")
		}
		result = append(result, &applyAncestor{path: current, info: info, exists: true})
	}
	return result, nil
}

func acquireApplyLock(layout DerivedLayout, rootInfo os.FileInfo, deps ApplyDeps) (*applyLock, []*applyAncestor, error) {
	created, err := ensureApplyDirectory(layout.Root, layout.DerivedDirectoryPath)
	if err != nil {
		return nil, created, err
	}
	if deps.beforeLock != nil {
		if hookErr := deps.beforeLock(); hookErr != nil {
			return nil, created, applyUnavailable("apply lock preflight failed")
		}
	}
	currentRoot, err := os.Lstat(layout.Root)
	if err != nil || !os.SameFile(rootInfo, currentRoot) {
		return nil, created, applyConflict("project root changed before lock acquisition")
	}
	path := filepath.Join(layout.DerivedDirectoryPath, applyLockName)
	handle, createdLock, err := openApplyLockFile(path)
	if err != nil {
		return nil, created, err
	}
	cleanup := func() {
		_ = handle.Close()
		if createdLock {
			_ = os.Remove(path)
		}
	}
	if lockErr := lockApplyFile(handle); lockErr != nil {
		cleanup()
		if errors.Is(lockErr, errApplyLockHeld) {
			return nil, created, applyConflict("another project apply is active")
		}
		return nil, created, applyUnavailable("apply lock could not be acquired")
	}
	info, statErr := handle.Stat()
	current, pathErr := os.Lstat(path)
	if statErr != nil || pathErr != nil || !os.SameFile(info, current) || current.Mode()&os.ModeSymlink != 0 || !current.Mode().IsRegular() || !validApplyPrivateFileMode(current.Mode()) {
		_ = unlockApplyFile(handle)
		cleanup()
		return nil, created, applyUnavailable("apply lock could not be verified")
	}
	return &applyLock{file: info, hand: handle, path: path, held: true}, created, nil
}

func openApplyLockFile(path string) (*os.File, bool, error) {
	handle, err := createApplyLockFile(path)
	if err == nil {
		if chmodErr := handle.Chmod(0o600); chmodErr != nil {
			_ = handle.Close()
			_ = os.Remove(path)
			return nil, false, applyUnavailable("apply lock could not be initialized")
		}
		if _, writeErr := handle.WriteString(applyLockMarker); writeErr != nil {
			_ = handle.Close()
			_ = os.Remove(path)
			return nil, false, applyUnavailable("apply lock could not be initialized")
		}
		if syncErr := handle.Sync(); syncErr != nil {
			_ = handle.Close()
			_ = os.Remove(path)
			return nil, false, applyUnavailable("apply lock could not be initialized")
		}
		return handle, true, nil
	}
	if !errors.Is(err, os.ErrExist) {
		return nil, false, applyUnavailable("apply lock could not be acquired")
	}
	handle, err = openExistingApplyLockFile(path)
	if err != nil {
		return nil, false, applyConflict("another project apply is active")
	}
	data, readErr := io.ReadAll(io.LimitReader(handle, int64(len(applyLockMarker)+1)))
	if readErr != nil || string(data) != applyLockMarker {
		_ = handle.Close()
		return nil, false, applyConflict("apply lock path is not reconciler-owned")
	}
	return handle, false, nil
}

func finishApplySetup(tx *applyTransaction, plan Plan, primary error) (ApplyResult, error) {
	if unlockErr := tx.closeLock(); unlockErr != nil && primary == nil {
		primary = unlockErr
	}
	if cleanupErr := tx.cleanupCreatedDirs(false); cleanupErr != nil && primary == nil {
		primary = cleanupErr
	}
	result := ApplyResult{Plan: plan, Installed: []AppliedPlacement{}}
	if tx.quarantine != nil && tx.quarantine.runInfo != nil {
		result.Quarantine = &ProjectQuarantineResult{ID: tx.quarantine.manifest.ID, Status: tx.quarantine.manifest.Status}
	}
	return result, primary
}

func cleanupApplyDirs(directories []*applyAncestor) error {
	failed := false
	for index := len(directories) - 1; index >= 0; index-- {
		directory := directories[index]
		info, err := os.Lstat(directory.path)
		if err != nil || !os.SameFile(directory.info, info) || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			failed = true
			continue
		}
		if removeErr := os.Remove(directory.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			failed = true
		}
	}
	if failed {
		return applyUnavailable("created directory cleanup could not be verified")
	}
	return nil
}

func ensureApplyDirectory(root, target string) ([]*applyAncestor, error) {
	if !pathWithin(root, target) {
		return nil, applyConflict("derived state path escapes the project")
	}
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, applyConflict("derived state path escapes the project")
	}
	current := filepath.Clean(root)
	created := make([]*applyAncestor, 0)
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return created, applyConflict("derived state directory could not be inspected")
		}
		if errors.Is(statErr, os.ErrNotExist) {
			if mkdirErr := os.Mkdir(current, 0o700); mkdirErr != nil {
				return created, applyConflict("derived state directory changed during lock acquisition")
			}
			info, statErr = os.Lstat(current)
			if statErr != nil {
				return created, applyUnavailable("derived state directory could not be verified")
			}
			created = append(created, &applyAncestor{path: current, info: info, exists: true, owned: true})
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return created, applyConflict("derived state directory is not a real directory")
		}
	}
	return created, nil
}

func (tx *applyTransaction) applyOperations(ctx context.Context, session *ProjectApplySession, preimage *applyStatePreimage, operations []PlanOperation) error {
	desired := desiredByPlacement(session.Desired)
	for _, operation := range operations {
		if err := contextErr(ctx); err != nil {
			return applyUnavailable("context cancelled")
		}
		if err := tx.checkRootAndAncestors(operation.Target); err != nil {
			return err
		}
		if err := tx.checkLock(); err != nil {
			return err
		}
		if operation.Action == PlanActionQuarantine {
			if tx.deps.beforeQuarantine != nil {
				if hookErr := tx.deps.beforeQuarantine(AppliedPlacement{Skill: operation.Skill, Target: operation.Target}); hookErr != nil {
					return applyUnavailable("quarantine preflight failed")
				}
			}
			root, err := tx.layout.ManagedSkillsPath(operation.Target)
			if err != nil {
				return applyConflict("target placement path changed after planning")
			}
			if err := tx.quarantineRemoval(session, preimage, operation, filepath.Join(root, operation.Skill)); err != nil {
				return err
			}
			continue
		}
		skill := desired[projectPlacementKey(operation.Target, operation.Skill)]
		if session.Materialized == nil {
			return applyUnavailable("materialized snapshot is unavailable")
		}
		snapshot, ok := session.Materialized.SnapshotFor(skill.Name)
		if !ok || snapshot == nil {
			return applyUnavailable("materialized snapshot is unavailable")
		}
		if err := snapshot.Verify(); err != nil {
			return applyConflict("materialized snapshot changed before placement")
		}
		root, err := tx.layout.ManagedSkillsPath(operation.Target)
		if err != nil {
			return applyConflict("target placement path changed after planning")
		}
		if tx.journal == nil {
			return applyUnavailable("transaction recovery evidence is unavailable")
		}
		stagingParent := filepath.Join(projectRecoveryRunPath(tx.layout, tx.journal.ID), "staging", string(operation.Target))
		if err := ensureRecoveryParent(tx, stagingParent); err != nil {
			return err
		}
		temp, err := tx.deps.MakeTempDir(stagingParent, applyInstallPattern)
		if err != nil {
			return applyUnavailable("temporary placement could not be created")
		}
		if copyErr := copyApplyTree(snapshot.Path, temp, tx.deps.SyncFile); copyErr != nil {
			_ = os.RemoveAll(temp)
			return copyErr
		}
		stagedHash, hashErr := HashSkillTree(temp)
		if hashErr != nil || stagedHash != snapshot.Hash {
			_ = os.RemoveAll(temp)
			return applyConflict("materialized snapshot changed during placement")
		}
		if syncErr := syncApplyTree(temp, tx.deps.SyncDir); syncErr != nil {
			_ = os.RemoveAll(temp)
			return applyUnavailable("temporary placement could not be synced")
		}
		if err := snapshot.Verify(); err != nil {
			_ = os.RemoveAll(temp)
			return applyConflict("materialized snapshot changed during placement")
		}
		if err := tx.checkRootAndAncestors(operation.Target); err != nil {
			_ = os.RemoveAll(temp)
			return err
		}
		if err := tx.checkLock(); err != nil {
			_ = os.RemoveAll(temp)
			return err
		}
		stagedInfo, err := os.Lstat(temp)
		if err != nil || stagedInfo.Mode()&os.ModeSymlink != 0 || !stagedInfo.IsDir() {
			_ = os.RemoveAll(temp)
			return applyUnavailable("temporary placement identity could not be verified")
		}
		destination := filepath.Join(root, skill.Name)
		if operation.Action == PlanActionUpdate {
			if tx.deps.beforeQuarantine != nil {
				if hookErr := tx.deps.beforeQuarantine(AppliedPlacement{Skill: skill.Name, Target: operation.Target}); hookErr != nil {
					_ = os.RemoveAll(temp)
					return applyUnavailable("quarantine preflight failed")
				}
			}
			if err := tx.quarantineUpdate(session, preimage, operation, destination); err != nil {
				_ = os.RemoveAll(temp)
				return err
			}
		}
		if tx.deps.BeforePublish != nil {
			if hookErr := tx.deps.BeforePublish(AppliedPlacement{Skill: skill.Name, Target: operation.Target}); hookErr != nil {
				_ = os.RemoveAll(temp)
				return applyUnavailable("placement preflight failed")
			}
		}
		if err := snapshot.Verify(); err != nil {
			_ = os.RemoveAll(temp)
			return applyConflict("materialized snapshot changed before publication")
		}
		if err := tx.checkRootAndAncestors(operation.Target); err != nil {
			_ = os.RemoveAll(temp)
			return err
		}
		if err := tx.checkLock(); err != nil {
			_ = os.RemoveAll(temp)
			return err
		}
		finalStagedHash, finalStagedErr := HashSkillTree(temp)
		if finalStagedErr != nil || finalStagedHash != snapshot.Hash {
			_ = os.RemoveAll(temp)
			return applyConflict("temporary placement changed before publication")
		}
		if publishErr := tx.deps.PublishNoReplace(temp, destination); publishErr != nil {
			_ = os.RemoveAll(temp)
			if errors.Is(publishErr, os.ErrExist) {
				return applyConflict("desired placement appeared during publication")
			}
			return applyUnavailable("desired placement could not be published")
		}
		info, statErr := os.Lstat(destination)
		if statErr != nil {
			return applyUnavailable("published placement could not be verified")
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !os.SameFile(stagedInfo, info) {
			return applyConflict("published placement identity changed during publication")
		}
		published := publishedPlacement{skill: skill.Name, target: operation.Target, action: operation.Action, dest: destination, info: info, expected: snapshot.Hash}
		tx.published = append(tx.published, published)
		if syncErr := tx.deps.SyncDir(root); syncErr != nil {
			return applyUnavailable("published placement could not be synced")
		}
		finalHash, finalErr := HashSkillTree(destination)
		if finalErr != nil || finalHash != snapshot.Hash {
			return applyConflict("published placement changed during verification")
		}
		if operation.Action == PlanActionUpdate {
			if err := tx.setQuarantineEntryStatus(operation.Target, operation.Skill, ProjectQuarantineEntryReplaced); err != nil {
				return err
			}
		}
	}
	return nil
}

func syncApplyTree(root string, syncDir func(string) error) error {
	if syncDir == nil {
		return nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	sort.SliceStable(entries, func(i, j int) bool { return compareUTF16(entries[i].Name(), entries[j].Name()) < 0 })
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("staged tree changed during sync")
		}
		if info.IsDir() {
			if err := syncApplyTree(path, syncDir); err != nil {
				return err
			}
		} else if !info.Mode().IsRegular() {
			return errors.New("staged tree contains a special file")
		}
	}
	return syncDir(root)
}

func (tx *applyTransaction) checkRootAndAncestors(target Target) error {
	info, err := os.Lstat(tx.root)
	if err != nil || !os.SameFile(tx.rootInfo, info) {
		return applyConflict("project root changed during placement")
	}
	values := tx.ancestors[target]
	for _, ancestor := range values {
		current, statErr := os.Lstat(ancestor.path)
		if ancestor.owned {
			if statErr != nil || !os.SameFile(ancestor.info, current) || current.Mode()&os.ModeSymlink != 0 || !current.IsDir() {
				return applyConflict("target ancestor changed during placement")
			}
			continue
		}
		if ancestor.exists {
			if statErr != nil || !os.SameFile(ancestor.info, current) || current.Mode()&os.ModeSymlink != 0 || !current.IsDir() {
				return applyConflict("target ancestor changed during placement")
			}
			continue
		}
		if statErr == nil {
			return applyConflict("target ancestor appeared during placement")
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return applyConflict("target ancestor could not be inspected")
		}
		if mkdirErr := os.Mkdir(ancestor.path, 0o755); mkdirErr != nil {
			return applyConflict("target ancestor appeared during placement")
		}
		createdInfo, verifyErr := os.Lstat(ancestor.path)
		if verifyErr != nil || !createdInfo.IsDir() || createdInfo.Mode()&os.ModeSymlink != 0 {
			return applyUnavailable("created target ancestor could not be verified")
		}
		ancestor.info = createdInfo
		ancestor.exists = true
		ancestor.owned = true
		tx.createdDirs = append(tx.createdDirs, ancestor)
		if syncErr := tx.deps.SyncDir(filepath.Dir(ancestor.path)); syncErr != nil {
			return applyUnavailable("created target ancestor could not be synced")
		}
	}
	return nil
}

func (tx *applyTransaction) syncCreatedDirectoryParents(directories []*applyAncestor) error {
	for _, directory := range directories {
		info, err := os.Lstat(directory.path)
		if err != nil || !os.SameFile(directory.info, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return applyConflict("created state directory changed before placement")
		}
		if syncErr := tx.deps.SyncDir(filepath.Dir(directory.path)); syncErr != nil {
			return applyUnavailable("created state directory could not be synced")
		}
	}
	return nil
}

func (tx *applyTransaction) checkLock() error {
	if tx.lock == nil {
		return applyConflict("apply lock is unavailable")
	}
	rootInfo, rootErr := os.Lstat(tx.root)
	if rootErr != nil || tx.rootInfo == nil || !os.SameFile(tx.rootInfo, rootInfo) || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return applyConflict("project root changed during transaction")
	}
	info, err := os.Lstat(tx.lock.path)
	if err != nil || !os.SameFile(tx.lock.file, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) {
		return applyConflict("apply lock changed during transaction")
	}
	return nil
}

func (tx *applyTransaction) rollback() error {
	if len(tx.published) == 0 && len(tx.quarantined) == 0 {
		if tx.quarantine != nil && tx.quarantine.runInfo != nil {
			return tx.setQuarantineStatus(ProjectQuarantineRolledBack)
		}
		return nil
	}
	var failed bool
	if tx.quarantine != nil && tx.quarantine.runInfo != nil {
		if err := tx.checkQuarantineBoundary(); err != nil {
			return applyUnavailable("rollback preserved recoverable project state")
		}
		if err := tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired); err != nil {
			failed = true
		}
	}
	var recoveryDir string
	if len(tx.published) > 0 {
		if tx.journal == nil {
			return applyUnavailable("rollback recovery could not be created")
		}
		recoveryDir = projectRecoveryRunPath(tx.layout, tx.journal.ID)
		created, err := ensureApplyDirectory(tx.root, recoveryDir)
		if err != nil {
			return applyUnavailable("rollback recovery could not be created")
		}
		if err := tx.syncCreatedDirectoryParents(created); err != nil {
			return err
		}
		recoveryInfo, err := os.Lstat(recoveryDir)
		if err != nil || !recoveryInfo.IsDir() || recoveryInfo.Mode()&os.ModeSymlink != 0 {
			return applyUnavailable("rollback recovery could not be verified")
		}
	}
	for index := len(tx.published) - 1; index >= 0; index-- {
		placement := tx.published[index]
		info, err := os.Lstat(placement.dest)
		if err != nil || !os.SameFile(placement.info, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			failed = true
			continue
		}
		hash, hashErr := HashSkillTree(placement.dest)
		if hashErr != nil || hash != placement.expected {
			failed = true
			continue
		}
		if tx.deps.beforeRollback != nil {
			if hookErr := tx.deps.beforeRollback(AppliedPlacement{Skill: placement.skill, Target: placement.target}); hookErr != nil {
				failed = true
				continue
			}
		}
		recoveryPath := filepath.Join(recoveryDir, string(placement.target), placement.skill)
		if err := ensureRecoveryParent(tx, filepath.Dir(recoveryPath)); err != nil {
			failed = true
			continue
		}
		if err := tx.checkRootAndAncestors(placement.target); err != nil || tx.checkLock() != nil {
			failed = true
			continue
		}
		info, err = os.Lstat(placement.dest)
		if err != nil || !os.SameFile(placement.info, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			failed = true
			continue
		}
		hash, hashErr = HashSkillTree(placement.dest)
		if hashErr != nil || hash != placement.expected {
			failed = true
			continue
		}
		if _, statErr := os.Lstat(recoveryPath); statErr == nil {
			failed = true
			continue
		}
		if moveErr := tx.deps.PublishNoReplace(placement.dest, recoveryPath); moveErr != nil {
			failed = true
			continue
		}
		movedInfo, statErr := os.Lstat(recoveryPath)
		movedHash, hashErr := HashSkillTree(recoveryPath)
		if statErr != nil || movedInfo.Mode()&os.ModeSymlink != 0 || !movedInfo.IsDir() || !os.SameFile(placement.info, movedInfo) || hashErr != nil || movedHash != placement.expected {
			failed = true
			continue
		}
	}
	for index := len(tx.quarantined) - 1; index >= 0; index-- {
		placement := tx.quarantined[index]
		entryIndex := placement.manifestIndex
		if tx.quarantine == nil || entryIndex < 0 || entryIndex >= len(tx.quarantine.manifest.Entries) {
			failed = true
			continue
		}
		if err := tx.checkQuarantineEntryAncestors(filepath.Dir(placement.quarantined)); err != nil {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		if err := tx.checkRootAndAncestors(placement.target); err != nil || tx.checkLock() != nil {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		if _, err := os.Lstat(placement.dest); err == nil || !errors.Is(err, os.ErrNotExist) {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		info, err := os.Lstat(placement.quarantined)
		hash, hashErr := HashSkillTree(placement.quarantined)
		if err != nil || !os.SameFile(placement.originalInfo, info) ||
			info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || hashErr != nil || hash != placement.oldHash {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		if moveErr := tx.deps.PublishNoReplace(placement.quarantined, placement.dest); moveErr != nil {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		restoredInfo, statErr := os.Lstat(placement.dest)
		restoredHash, restoredHashErr := HashSkillTree(placement.dest)
		if statErr != nil || !os.SameFile(placement.originalInfo, restoredInfo) || restoredInfo.Mode()&os.ModeSymlink != 0 ||
			!restoredInfo.IsDir() || restoredHashErr != nil || restoredHash != placement.oldHash {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		root, _ := tx.layout.ManagedSkillsPath(placement.target)
		if tx.deps.SyncDir(root) != nil || tx.deps.SyncDir(filepath.Dir(placement.quarantined)) != nil {
			failed = true
			_ = tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRecoveryRequired)
			continue
		}
		if err := tx.setQuarantineEntryStatus(placement.target, placement.skill, ProjectQuarantineEntryRestored); err != nil {
			failed = true
		}
	}
	if !failed && tx.quarantine != nil {
		if err := tx.setQuarantineStatus(ProjectQuarantineRolledBack); err != nil {
			failed = true
		}
	}
	if failed {
		return applyUnavailable("rollback preserved recoverable project state")
	}
	return nil
}

func captureApplyStatePreimage(layout DerivedLayout) (*applyStatePreimage, error) {
	return captureTransactionStatePreimage(layout, ScopeProject)
}

func captureTransactionStatePreimage(layout DerivedLayout, scope Scope) (*applyStatePreimage, error) {
	probe := proveInspectionPath(layout.Root, layout.ReconcilerStatePath)
	if probe.unsafe {
		return nil, applyConflict("provenance state changed after planning")
	}
	if !probe.exists {
		return &applyStatePreimage{state: emptyTransactionProvenanceState(scope)}, nil
	}
	if probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		return nil, applyConflict("provenance state is not a regular file")
	}
	data, err := readBoundedInspectionFile(layout.ReconcilerStatePath, maxProvenanceStateBytes)
	if err != nil {
		return nil, applyConflict("provenance state could not be read")
	}
	state, valid := decodeTransactionProvenanceState(data, scope, true)
	if !valid {
		return nil, applyConflict("provenance state is malformed")
	}
	return &applyStatePreimage{
		exists: true,
		data:   append([]byte(nil), data...),
		state:  state,
		info:   probe.info,
		mode:   probe.info.Mode().Perm(),
	}, nil
}

func emptyTransactionProvenanceState(scope Scope) ProvenanceState {
	if scope == ScopeGlobal {
		return ProvenanceState{Version: GlobalProvenanceStateVersion, Records: []ProvenanceRecord{}}
	}
	return emptyTrustedProvenanceState()
}

func decodeTransactionProvenanceState(data []byte, scope Scope, allowLegacyGlobal bool) (ProvenanceState, bool) {
	if scope == ScopeProject {
		return decodeProvenanceState(data)
	}
	if scope != ScopeGlobal {
		return ProvenanceState{}, false
	}
	state, format, valid := decodeGlobalProvenanceState(data)
	if !valid || (!allowLegacyGlobal && format != GlobalProvenanceCurrent) {
		return ProvenanceState{}, false
	}
	return ProvenanceState{Version: state.Version, Records: cloneProvenanceRecords(state.Records)}, true
}

func (tx *applyTransaction) captureStatePreimage() (*applyStatePreimage, error) {
	scope := tx.scope
	if scope == "" {
		scope = ScopeProject
	}
	return captureTransactionStatePreimage(tx.layout, scope)
}

func sameApplyState(left, right *applyStatePreimage) bool {
	if left == nil || right == nil || left.exists != right.exists {
		return false
	}
	if !left.exists {
		return true
	}
	return bytes.Equal(left.data, right.data) && os.SameFile(left.info, right.info)
}

func (tx *applyTransaction) commitState(preimage *applyStatePreimage, session *ProjectApplySession) (*stateCommit, error) {
	current, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(preimage, current) {
		return nil, applyConflict("provenance state changed before commit")
	}
	if tx.journal == nil {
		return nil, applyUnavailable("transaction recovery evidence is unavailable")
	}
	state, valid := decodeTransactionProvenanceState(tx.journal.CandidateState, tx.scope, false)
	if !valid {
		return nil, applyConflict("transaction recovery evidence is invalid")
	}
	data := append([]byte(nil), tx.journal.CandidateState...)
	temporaryPath, err := tx.prepareStateBytes(data, 0o600)
	if err != nil {
		return nil, err
	}
	defer os.Remove(temporaryPath)
	if tx.deps.beforeCommit != nil {
		if hookErr := tx.deps.beforeCommit(); hookErr != nil {
			return nil, applyUnavailable("provenance commit preflight failed")
		}
	}
	if err := tx.checkLock(); err != nil {
		return nil, err
	}
	if err := tx.revalidateCandidate(session, state); err != nil {
		return nil, err
	}
	current, err = tx.captureStatePreimage()
	if err != nil || !sameApplyState(preimage, current) {
		return nil, applyConflict("provenance state changed before commit")
	}
	if err := tx.checkLock(); err != nil {
		return nil, err
	}
	return tx.publishStateBytes(temporaryPath, data, preimage)
}

func (tx *applyTransaction) revalidateCandidate(session *ProjectApplySession, candidate ProvenanceState) error {
	if tx.global != nil {
		inventory, err := InspectGlobal(tx.global.layout)
		if err != nil || !inventory.StateTrusted {
			return applyConflict("global state changed before commit")
		}
		inventory.State = GlobalProvenanceState{Version: candidate.Version, Records: cloneProvenanceRecords(candidate.Records)}
		inventory.ProvenanceFormat = GlobalProvenanceCurrent
		inventory.MigrationRequired = false
		classification, err := ClassifyGlobal(tx.global.registry, session.Desired, session.Expected, inventory)
		if err != nil {
			return applyConflict("global classification changed before commit")
		}
		return tx.revalidateManagedCandidate(session, classification.Managed, candidate)
	}
	inventory, err := InspectProject(tx.layout)
	if err != nil || !inventory.StateTrusted {
		return applyConflict("project state changed before commit")
	}
	inventory.State = candidate
	classification, err := ClassifyProject(session.Desired, session.Expected, inventory)
	if err != nil {
		return applyConflict("project classification changed before commit")
	}
	for _, operation := range reviewedMutationOperations(session) {
		if operation.Action != PlanActionQuarantine {
			continue
		}
		root, ok := inventory.RootFor(operation.Target)
		if !ok || !root.Safe {
			return applyConflict("removed placement root changed before commit")
		}
	}
	return tx.revalidateManagedCandidate(session, classification, candidate)
}

func (tx *applyTransaction) revalidateManagedCandidate(session *ProjectApplySession, classification ProjectClassification, candidate ProvenanceState) error {
	desiredByKey := desiredByPlacement(session.Desired)
	seen := make(map[string]struct{}, len(desiredByKey))
	for _, state := range classification.States {
		key := projectPlacementKey(state.Target, state.Skill)
		skill, wanted := desiredByKey[key]
		if !wanted || skill.Manager != ManagerSkillsCLI {
			continue
		}
		seen[key] = struct{}{}
		if state.Kind != ProjectStateExact || state.Action != PlanActionUnchanged || state.Reason != ProjectStateReasonVerifiedExact {
			return applyConflict("managed placement changed before commit")
		}
	}
	for key, skill := range desiredByKey {
		if skill.Manager == ManagerSkillsCLI {
			if _, ok := seen[key]; !ok {
				return applyConflict("managed placement disappeared before commit")
			}
		}
	}
	for _, operation := range reviewedMutationOperations(session) {
		if operation.Action != PlanActionQuarantine {
			continue
		}
		key := projectPlacementKey(operation.Target, operation.Skill)
		if _, stillManaged := candidateRecordByKey(candidate, key); stillManaged {
			return applyConflict("removed placement remained managed before commit")
		}
		for _, state := range classification.States {
			if state.Target == operation.Target && state.Skill == operation.Skill && state.Action == PlanActionQuarantine {
				return applyConflict("removed placement remained managed before commit")
			}
		}
	}
	for _, placement := range tx.published {
		info, statErr := os.Lstat(placement.dest)
		if statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !os.SameFile(placement.info, info) {
			return applyConflict("published placement changed before commit")
		}
		hash, hashErr := HashSkillTree(placement.dest)
		if hashErr != nil || hash != placement.expected {
			return applyConflict("published placement changed before commit")
		}
	}
	return nil
}

func buildApplyState(previous ProvenanceState, session *ProjectApplySession, recordedAt time.Time) (ProvenanceState, error) {
	state := previous
	state.Version = ProvenanceStateVersion
	if session.Desired.Scope == ScopeGlobal {
		state.Version = GlobalProvenanceStateVersion
	}
	if state.Records == nil {
		state.Records = []ProvenanceRecord{}
	}
	records := make(map[string]ProvenanceRecord, len(state.Records)+len(session.Plan.Operations))
	for _, record := range state.Records {
		records[projectPlacementKey(record.Target, record.Skill)] = record
	}
	desired := desiredByPlacement(session.Desired)
	for _, operation := range reviewedMutationOperations(session) {
		key := projectPlacementKey(operation.Target, operation.Skill)
		if operation.Action == PlanActionQuarantine {
			record, managed := records[key]
			oldHash, oldHashOK := treeHashFromPlanEvidence(operation.Current)
			expectedHash, expectedHashOK := treeHashFromPlanEvidence(operation.Expected)
			if !managed || record.Scope != session.Desired.Scope || record.Skill != operation.Skill ||
				record.Target != operation.Target || record.SourceIdentity != operation.Source ||
				!isCanonicalProjectSourceIdentity(operation.Source) || !oldHashOK || !expectedHashOK || expectedHash != oldHash ||
				record.TreeHashAlgorithm != oldHash.Algorithm || record.TreeHash != oldHash.Digest {
				return ProvenanceState{}, applyConflict("removed provenance identity changed before commit")
			}
			delete(records, key)
			continue
		}
		skill, ok := desired[key]
		if !ok {
			return ProvenanceState{}, applyConflict("desired identity disappeared before provenance commit")
		}
		hash, ok := session.Expected[skill.Name]
		if !ok {
			return ProvenanceState{}, applyUnavailable("expected materialized content is unavailable")
		}
		sourceIdentity, ok := canonicalProjectSourceIdentity(skill.Source)
		if !ok {
			return ProvenanceState{}, applyConflict("desired source identity is not canonical")
		}
		records[projectPlacementKey(operation.Target, operation.Skill)] = ProvenanceRecord{
			Scope:             session.Desired.Scope,
			Skill:             skill.Name,
			Target:            operation.Target,
			SourceIdentity:    sourceIdentity,
			TreeHashAlgorithm: hash.Algorithm,
			TreeHash:          hash.Digest,
			RecordedAt:        recordedAt,
		}
	}
	state.Records = make([]ProvenanceRecord, 0, len(records))
	for _, record := range records {
		state.Records = append(state.Records, record)
	}
	sort.SliceStable(state.Records, func(i, j int) bool {
		if projectTargetRank(state.Records[i].Target) != projectTargetRank(state.Records[j].Target) {
			return projectTargetRank(state.Records[i].Target) < projectTargetRank(state.Records[j].Target)
		}
		return compareUTF16(state.Records[i].Skill, state.Records[j].Skill) < 0
	})
	return state, nil
}

func marshalApplyState(state ProvenanceState) ([]byte, error) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (tx *applyTransaction) writeStateBytes(expected *applyStatePreimage, data []byte, mode os.FileMode) (*stateCommit, error) {
	temporaryPath, err := tx.prepareStateBytes(data, mode)
	if err != nil {
		return nil, err
	}
	defer os.Remove(temporaryPath)
	return tx.publishStateBytes(temporaryPath, data, expected)
}

func (tx *applyTransaction) prepareStateBytes(data []byte, mode os.FileMode) (string, error) {
	parent := filepath.Dir(tx.layout.ReconcilerStatePath)
	temporary, err := os.CreateTemp(parent, ".sjskills-state-")
	if err != nil {
		return "", applyUnavailable("provenance temporary file could not be created")
	}
	temporaryPath := temporary.Name()
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(mode.Perm()); err != nil {
		_ = temporary.Close()
		return "", applyUnavailable("provenance temporary file permissions could not be set")
	}
	if written, err := temporary.Write(data); err != nil || written != len(data) {
		_ = temporary.Close()
		return "", applyUnavailable("provenance temporary file could not be written")
	}
	if err := tx.deps.SyncFile(temporary); err != nil {
		_ = temporary.Close()
		return "", applyUnavailable("provenance temporary file could not be synced")
	}
	if err := temporary.Close(); err != nil {
		return "", applyUnavailable("provenance temporary file could not be closed")
	}
	succeeded = true
	return temporaryPath, nil
}

func (tx *applyTransaction) publishStateBytes(temporaryPath string, data []byte, expected *applyStatePreimage) (*stateCommit, error) {
	parent := filepath.Dir(tx.layout.ReconcilerStatePath)
	if err := tx.checkLock(); err != nil {
		return nil, err
	}
	current, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(expected, current) {
		return nil, applyConflict("provenance state changed immediately before replacement")
	}
	var publishErr error
	if expected.exists {
		publishErr = tx.deps.ReplaceFileAtomic(temporaryPath, tx.layout.ReconcilerStatePath)
	} else {
		publishErr = tx.deps.publishStateNoReplace(temporaryPath, tx.layout.ReconcilerStatePath)
	}
	if publishErr != nil {
		if commit, landed := tx.observeStateCommit(data); landed {
			return commit, applyUnavailable("provenance state replacement was not confirmed")
		}
		if !expected.exists && errors.Is(publishErr, os.ErrExist) {
			return nil, applyConflict("provenance state appeared during publication")
		}
		return nil, applyUnavailable("provenance state could not be atomically replaced")
	}
	commit, landed := tx.observeStateCommit(data)
	if !landed {
		return &stateCommit{replaced: true, data: append([]byte(nil), data...)}, applyUnavailable("provenance state could not be verified")
	}
	if err := tx.deps.SyncDir(parent); err != nil {
		return commit, applyUnavailable("provenance directory could not be synced")
	}
	return commit, nil
}

func (tx *applyTransaction) observeStateCommit(data []byte) (*stateCommit, bool) {
	info, statErr := os.Lstat(tx.layout.ReconcilerStatePath)
	stored, readErr := readBoundedInspectionFile(tx.layout.ReconcilerStatePath, maxProvenanceStateBytes)
	if statErr != nil || readErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !bytes.Equal(stored, data) {
		return nil, false
	}
	return &stateCommit{replaced: true, data: append([]byte(nil), data...), info: info}, true
}

func (tx *applyTransaction) restoreState(preimage *applyStatePreimage, commit *stateCommit) error {
	if preimage == nil || commit == nil || !commit.replaced {
		return nil
	}
	current, err := tx.captureStatePreimage()
	if err != nil || !current.exists || !bytes.Equal(current.data, commit.data) || (commit.info != nil && !os.SameFile(current.info, commit.info)) {
		return applyUnavailable("provenance restoration could not prove state ownership")
	}
	if preimage.exists {
		_, err := tx.writeStateBytes(current, preimage.data, preimage.mode)
		return err
	}
	return tx.moveOwnedStateToRecovery(current, "new provenance state")
}

func (tx *applyTransaction) closeLock() error {
	if tx.lock == nil {
		return nil
	}
	info, statErr := os.Lstat(tx.lock.path)
	if statErr != nil || !os.SameFile(tx.lock.file, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) {
		_ = unlockApplyFile(tx.lock.hand)
		_ = tx.lock.hand.Close()
		return applyConflict("apply lock changed during transaction")
	}
	if removeErr := os.Remove(tx.lock.path); removeErr != nil {
		_ = unlockApplyFile(tx.lock.hand)
		_ = tx.lock.hand.Close()
		return applyUnavailable("apply lock could not be removed")
	}
	if tx.lock.held {
		if unlockErr := unlockApplyFile(tx.lock.hand); unlockErr != nil {
			_ = tx.lock.hand.Close()
			return applyUnavailable("apply lock could not be released")
		}
		tx.lock.held = false
	}
	closeErr := tx.lock.hand.Close()
	if closeErr != nil {
		return applyUnavailable("apply lock could not be closed")
	}
	tx.lock = nil
	return nil
}

func (tx *applyTransaction) cleanupCreatedDirs(success bool) error {
	if success {
		return nil
	}
	return cleanupApplyDirs(tx.createdDirs)
}

// copyApplyTree copies only regular files and real directories. Symlinks and
// special files are rejected even though the read-only tree hash can represent
// safe symlinks; v1 project placement is a directory-copy contract.
func copyApplyTree(source, destination string, syncFile func(*os.File) error) error {
	info, err := os.Lstat(source)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return applyConflict("materialized snapshot is not a real directory")
	}
	if err := os.Chmod(destination, info.Mode().Perm()); err != nil {
		return applyUnavailable("temporary placement permissions could not be set")
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return applyUnavailable("materialized snapshot could not be read")
	}
	sort.SliceStable(entries, func(i, j int) bool { return compareUTF16(entries[i].Name(), entries[j].Name()) < 0 })
	for _, entry := range entries {
		name := entry.Name()
		if !isApplyEntryName(name) {
			return applyConflict("materialized snapshot contains an invalid entry name")
		}
		sourcePath := filepath.Join(source, name)
		destinationPath := filepath.Join(destination, name)
		entryInfo, statErr := os.Lstat(sourcePath)
		if statErr != nil {
			return applyConflict("materialized snapshot changed during copy")
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return applyConflict("materialized snapshot contains a symlink")
		}
		switch {
		case entryInfo.IsDir():
			if mkdirErr := os.Mkdir(destinationPath, entryInfo.Mode().Perm()); mkdirErr != nil {
				return applyUnavailable("temporary placement directory could not be created")
			}
			if copyErr := copyApplyTree(sourcePath, destinationPath, syncFile); copyErr != nil {
				return copyErr
			}
		case entryInfo.Mode().IsRegular():
			if copyErr := copyApplyFile(sourcePath, destinationPath, entryInfo, syncFile); copyErr != nil {
				return copyErr
			}
		default:
			return applyConflict("materialized snapshot contains a special file")
		}
	}
	return nil
}

func isApplyEntryName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsRune(name, 0) && filepath.Base(name) == name
}

func copyApplyFile(source, destination string, expected os.FileInfo, syncFile func(*os.File) error) error {
	input, err := os.Open(source)
	if err != nil {
		return applyUnavailable("materialized file could not be opened")
	}
	defer input.Close()
	openedInfo, err := input.Stat()
	if err != nil || !os.SameFile(expected, openedInfo) {
		return applyConflict("materialized snapshot changed during copy")
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, expected.Mode().Perm())
	if err != nil {
		return applyUnavailable("temporary placement file could not be created")
	}
	copyErr := func() error {
		if _, err := io.Copy(output, input); err != nil {
			return err
		}
		if err := output.Chmod(expected.Mode().Perm()); err != nil {
			return err
		}
		return syncFile(output)
	}()
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil {
		return applyUnavailable("temporary placement file could not be copied")
	}
	return nil
}
