package sjskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RestoreFailureKind is the stable result class for a restore transaction.
// Restore errors deliberately do not retain operating-system diagnostics or
// filesystem paths: a caller can safely surface them as process evidence.
type RestoreFailureKind string

const (
	RestoreFailureUnavailable RestoreFailureKind = "unavailable"
	RestoreFailureConflict    RestoreFailureKind = "conflict"
)

// RestoreError is returned for a restore boundary failure. Reason is a
// bounded vocabulary owned by this package, rather than an underlying error.
type RestoreError struct {
	Kind   RestoreFailureKind
	Reason string
	Scope  Scope
}

func (e *RestoreError) Error() string {
	if e == nil {
		return "project restore failed"
	}
	scope := "project"
	if e.Scope == ScopeGlobal {
		scope = "global"
	}
	if e.Kind == RestoreFailureConflict {
		return scope + " restore conflict: " + e.Reason
	}
	return scope + " restore unavailable: " + e.Reason
}

func (e *RestoreError) Conflict() bool { return e != nil && e.Kind == RestoreFailureConflict }

func restoreUnavailable(reason string) error {
	return &RestoreError{Kind: RestoreFailureUnavailable, Reason: reason}
}

func restoreConflict(reason string) error {
	return &RestoreError{Kind: RestoreFailureConflict, Reason: reason}
}

// RestoreResult contains only the quarantine handle and modeled placements.
// It never contains a project, run, manifest, or staging path.
type RestoreResult struct {
	ID       string                  `json:"id"`
	Status   ProjectQuarantineStatus `json:"status"`
	Restored []AppliedPlacement      `json:"restored"`
}

type restoreBoundary struct {
	identities map[string]os.FileInfo
}

type restoreEntry struct {
	entry          ProjectQuarantineManifestEntry
	destination    string
	quarantined    string
	quarantineDir  string
	oldHash        TreeHash
	quarantineInfo os.FileInfo
}

type restoredPlacement struct {
	entryIndex  int
	skill       string
	target      Target
	source      string
	destination string
	info        os.FileInfo
	oldHash     TreeHash
}

type restoreTransaction struct {
	applyTransaction

	boundary *restoreBoundary

	originalManifest     ProjectQuarantineManifest
	originalManifestData []byte
	originalManifestInfo os.FileInfo

	entries    []restoreEntry
	moved      []restoredPlacement
	ambiguous  map[int]bool
	rolledBack map[int]bool

	statePreimage      *applyStatePreimage
	stateCandidateData []byte
	stateCommit        *stateCommit

	manifestOwnedData [][]byte
	createdManaged    []*applyAncestor
}

// RestoreProjectQuarantine restores one committed project quarantine run.
// The run is addressed only by its exact lower-hex identifier; callers cannot
// supply an arbitrary manifest path.
func RestoreProjectQuarantine(ctx context.Context, layout DerivedLayout, id string, deps ApplyDeps) (RestoreResult, error) {
	return restoreQuarantine(ctx, layout, nil, id, deps)
}

// RestoreGlobalQuarantine restores a committed fixed-home quarantine run
// without overwriting a global placement.
func RestoreGlobalQuarantine(ctx context.Context, layout GlobalLayout, id string, deps ApplyDeps) (RestoreResult, error) {
	result, err := restoreQuarantine(ctx, layout.mutationLayout(), &globalApplyBoundary{layout: layout}, id, deps)
	return result, scopeRestoreFailure(err, ScopeGlobal)
}

func scopeRestoreFailure(err error, scope Scope) error {
	if err == nil {
		return nil
	}
	var failure *RestoreError
	if !errors.As(err, &failure) {
		return err
	}
	return &RestoreError{Kind: failure.Kind, Reason: failureReasonForScope(failure.Reason, scope), Scope: scope}
}

func restoreQuarantine(ctx context.Context, layout DerivedLayout, global *globalApplyBoundary, id string, deps ApplyDeps) (RestoreResult, error) {
	result := RestoreResult{Restored: []AppliedPlacement{}}
	if !projectQuarantineIDPattern.MatchString(id) {
		return result, restoreConflict("quarantine identifier is invalid")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := contextErr(ctx); err != nil {
		return result, restoreUnavailable("context cancelled")
	}
	var boundaryErr error
	if global != nil {
		_, boundaryErr = establishGlobalInspectionBoundary(global.layout)
	} else {
		_, boundaryErr = establishInspectionBoundary(layout)
	}
	if boundaryErr != nil {
		return result, restoreConflict("project boundary is invalid")
	}
	deps = normalizeApplyDeps(deps)
	rootInfo, err := applyRootInfo(layout.Root)
	if err != nil {
		return result, restoreConflict("project boundary is invalid")
	}
	lock, createdDerived, err := acquireApplyLock(layout, rootInfo, deps)
	if err != nil {
		if cleanupErr := cleanupApplyDirs(createdDerived); cleanupErr != nil {
			return result, restoreUnavailable("restore setup could not be cleaned up")
		}
		return result, restoreApplyError(err, "project lock is unavailable", "project lock could not be acquired")
	}
	tx := restoreTransaction{
		applyTransaction: applyTransaction{
			root:        layout.Root,
			rootInfo:    rootInfo,
			ancestors:   make(map[Target][]*applyAncestor),
			createdDirs: createdDerived,
			lock:        lock,
			layout:      layout,
			deps:        deps,
			scope:       ScopeProject,
			global:      global,
		},
		ambiguous:  make(map[int]bool),
		rolledBack: make(map[int]bool),
	}
	if global != nil {
		tx.scope = ScopeGlobal
	}
	if err := tx.checkRestoreLockAndRoot(); err != nil {
		return tx.finish(result, err, true)
	}
	if recovered, recoveryErr := tx.recoverInterruptedTransaction(); recoveryErr != nil {
		if tx.quarantine != nil {
			result.ID = tx.quarantine.manifest.ID
			result.Status = tx.quarantine.manifest.Status
		}
		return tx.finish(result, restoreApplyError(recoveryErr, "interrupted project transaction is ambiguous", "interrupted project transaction could not be recovered"), true)
	} else if recovered {
		if tx.quarantine != nil {
			result.ID = tx.quarantine.manifest.ID
			result.Status = tx.quarantine.manifest.Status
		}
		return tx.finish(result, restoreUnavailable("interrupted project transaction was recovered; rerun restore"), true)
	}
	if err := tx.loadManifest(id); err != nil {
		if tx.quarantine != nil {
			result.ID = tx.quarantine.manifest.ID
			result.Status = tx.quarantine.manifest.Status
		}
		return tx.finish(result, err, true)
	}
	result.ID = tx.quarantine.manifest.ID
	result.Status = tx.quarantine.manifest.Status
	if err := tx.preflight(ctx); err != nil {
		return tx.finish(result, err, true)
	}
	if tx.quarantine.manifest.Status == ProjectQuarantineRestored {
		if err := tx.finishSuccess(&result); err != nil {
			return result, err
		}
		return result, nil
	}
	if err := tx.prepareRestoreJournal(); err != nil {
		return tx.finish(result, err, false)
	}

	var primary error
	if err := tx.beginRestore(); err != nil {
		primary = err
	}
	if primary == nil {
		primary = tx.restoreEntries(ctx)
	}
	if primary == nil {
		primary = tx.commitRestoredProvenance()
	}
	if primary == nil && tx.deps.afterStateCommit != nil {
		tx.deps.afterStateCommit()
	}
	if primary == nil {
		primary = tx.finalizeRestoreManifest()
	}
	if primary == nil && tx.deps.beforeUnlock != nil {
		if hookErr := tx.deps.beforeUnlock(); hookErr != nil {
			primary = restoreUnavailable("restore finalization preflight failed")
		}
	}
	if primary == nil {
		primary = tx.verifyMovedPlacements()
	}
	if primary == nil {
		if journalErr := tx.clearTransactionJournal(); journalErr != nil {
			primary = restoreApplyError(journalErr, "restore recovery evidence changed", "restore recovery evidence could not be finalized")
		}
	}
	if primary != nil && !tx.journalCleared {
		if rollbackErr := tx.rollbackRestore(); rollbackErr != nil {
			result.Status = ProjectQuarantineRecoveryRequired
			_ = tx.closeRestoreLock()
			return result, restoreUnavailable("restore rollback could not be verified")
		}
		if journalErr := tx.clearTransactionJournal(); journalErr != nil {
			result.Status = ProjectQuarantineRecoveryRequired
			_ = tx.closeRestoreLock()
			return result, restoreUnavailable("restore recovery evidence could not be finalized")
		}
		result.Status = ProjectQuarantineCommitted
		if cleanupErr := tx.cleanupManagedDirectories(); cleanupErr != nil {
			return result, cleanupErr
		}
		if closeErr := tx.closeRestoreLock(); closeErr != nil {
			return result, closeErr
		}
		return result, primary
	}

	result.Status = ProjectQuarantineRestored
	result.Restored = make([]AppliedPlacement, 0, len(tx.entries))
	for _, entry := range tx.entries {
		result.Restored = append(result.Restored, AppliedPlacement{Skill: entry.entry.Skill, Target: entry.entry.Target})
	}
	if err := tx.closeRestoreLock(); err != nil {
		return result, err
	}
	if primary != nil {
		return result, primary
	}
	return result, nil
}

func restoreApplyError(err error, conflictReason, unavailableReason string) error {
	if err == nil {
		return nil
	}
	var restoreErr *RestoreError
	if errors.As(err, &restoreErr) {
		return err
	}
	var applyErr *ApplyError
	if errors.As(err, &applyErr) && applyErr.Conflict() {
		return restoreConflict(conflictReason)
	}
	return restoreUnavailable(unavailableReason)
}

func (tx *restoreTransaction) finish(result RestoreResult, primary error, cleanup bool) (RestoreResult, error) {
	if closeErr := tx.closeRestoreLock(); closeErr != nil && primary == nil {
		primary = closeErr
	}
	if cleanup {
		if cleanupErr := cleanupApplyDirs(tx.createdDirs); cleanupErr != nil && primary == nil {
			primary = restoreUnavailable("restore setup could not be cleaned up")
		}
	}
	return result, primary
}

func (tx *restoreTransaction) finishSuccess(result *RestoreResult) error {
	result.Status = ProjectQuarantineRestored
	if err := tx.closeRestoreLock(); err != nil {
		return err
	}
	return nil
}

func (tx *restoreTransaction) closeRestoreLock() error {
	if tx.lock == nil {
		return nil
	}
	info, statErr := os.Lstat(tx.lock.path)
	if statErr != nil || !os.SameFile(tx.lock.file, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) {
		_ = unlockApplyFile(tx.lock.hand)
		_ = tx.lock.hand.Close()
		return restoreConflict("project lock changed during restore")
	}
	if removeErr := os.Remove(tx.lock.path); removeErr != nil {
		_ = unlockApplyFile(tx.lock.hand)
		_ = tx.lock.hand.Close()
		return restoreUnavailable("project lock could not be removed")
	}
	if tx.lock.held {
		if unlockErr := unlockApplyFile(tx.lock.hand); unlockErr != nil {
			_ = tx.lock.hand.Close()
			return restoreUnavailable("project lock could not be released")
		}
		tx.lock.held = false
	}
	closeErr := tx.lock.hand.Close()
	if closeErr != nil {
		return restoreUnavailable("project lock could not be closed")
	}
	tx.lock = nil
	return nil
}

func (tx *restoreTransaction) checkRestoreLockAndRoot() error {
	if err := tx.checkLock(); err != nil {
		return restoreApplyError(err, "project lock changed during restore", "project lock could not be verified")
	}
	info, err := os.Lstat(tx.root)
	if err != nil || !os.SameFile(tx.rootInfo, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return restoreConflict("project root changed during restore")
	}
	return nil
}

func (tx *restoreTransaction) loadManifest(id string) error {
	quarantineRootInfo, err := os.Lstat(tx.layout.QuarantinePath)
	if err != nil || !quarantineRootInfo.IsDir() || quarantineRootInfo.Mode()&os.ModeSymlink != 0 {
		return restoreConflict("quarantine root is unavailable")
	}
	identities, err := captureRestoreDirectoryChain(tx.root, tx.layout.QuarantinePath)
	if err != nil {
		return err
	}
	runPath := filepath.Join(tx.layout.QuarantinePath, id)
	if !pathWithin(tx.layout.QuarantinePath, runPath) {
		return restoreConflict("quarantine run path is invalid")
	}
	runInfo, err := os.Lstat(runPath)
	if err != nil || !runInfo.IsDir() || runInfo.Mode()&os.ModeSymlink != 0 {
		return restoreConflict("quarantine run is unavailable")
	}
	runIdentities, err := captureRestoreDirectoryChain(tx.layout.QuarantinePath, runPath)
	if err != nil {
		return err
	}
	for path, info := range runIdentities {
		identities[path] = info
	}
	manifestPath := filepath.Join(runPath, applyManifestName)
	probe := proveInspectionPath(tx.root, manifestPath)
	if probe.unsafe || !probe.exists || probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		return restoreConflict("quarantine manifest is unavailable")
	}
	data, err := readBoundedInspectionFile(manifestPath, maxProjectQuarantineManifestSize)
	if err != nil {
		return restoreConflict("quarantine manifest is malformed")
	}
	manifest, ok := DecodeProjectQuarantineManifest(data)
	if !ok {
		return restoreConflict("quarantine manifest is malformed")
	}
	if manifest.ID != id {
		return restoreConflict("quarantine manifest identity does not match the request")
	}
	if manifest.Status != ProjectQuarantineCommitted && manifest.Status != ProjectQuarantineRestored {
		return restoreConflict("quarantine manifest requires a different recovery workflow")
	}
	for _, entry := range manifest.Entries {
		quarantined := filepath.Join(runPath, filepath.FromSlash(entry.QuarantinedPlacement))
		if !pathWithin(runPath, quarantined) || filepath.Clean(quarantined) != filepath.Join(runPath, filepath.FromSlash(projectQuarantinedPlacement(entry.Target, entry.Skill))) {
			return restoreConflict("quarantine entry path is invalid")
		}
		parent := filepath.Dir(quarantined)
		entryIdentities, entryErr := captureRestoreDirectoryChain(runPath, parent)
		if entryErr != nil {
			return entryErr
		}
		for path, info := range entryIdentities {
			if existing, exists := identities[path]; exists && !os.SameFile(existing, info) {
				return restoreConflict("quarantine entry boundary changed")
			}
			identities[path] = info
		}
	}
	tx.boundary = &restoreBoundary{identities: identities}
	tx.quarantine = &projectQuarantineTransaction{
		runPath: runPath, runDurable: true, rootInfo: quarantineRootInfo, runInfo: runInfo,
		ancestors: make(map[string]os.FileInfo), manifestPath: manifestPath,
		manifestInfo: probe.info, manifestData: append([]byte(nil), data...), manifest: manifest,
	}
	tx.originalManifest = manifest
	tx.originalManifestData = append([]byte(nil), data...)
	tx.originalManifestInfo = probe.info
	return tx.checkRestoreBoundary()
}

func captureRestoreDirectoryChain(boundary, candidate string) (map[string]os.FileInfo, error) {
	if !pathWithin(boundary, candidate) {
		return nil, restoreConflict("restore path escapes its modeled boundary")
	}
	relative, err := filepath.Rel(filepath.Clean(boundary), filepath.Clean(candidate))
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, restoreConflict("restore path escapes its modeled boundary")
	}
	current := filepath.Clean(boundary)
	result := make(map[string]os.FileInfo)
	parts := splitApplyPath(relative)
	if len(parts) == 0 {
		info, statErr := os.Lstat(current)
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, restoreConflict("restore boundary is not a real directory")
		}
		result[current] = info
		return result, nil
	}
	for _, part := range parts {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, restoreConflict("restore boundary is not a real directory")
		}
		result[current] = info
	}
	return result, nil
}

func (tx *restoreTransaction) checkRestoreBoundary() error {
	if err := tx.checkRestoreLockAndRoot(); err != nil {
		return err
	}
	if tx.boundary == nil || tx.quarantine == nil {
		return restoreConflict("restore boundary is unavailable")
	}
	for path, expected := range tx.boundary.identities {
		current, err := os.Lstat(path)
		if err != nil || !os.SameFile(expected, current) || current.Mode()&os.ModeSymlink != 0 || !current.IsDir() {
			return restoreConflict("restore boundary changed during transaction")
		}
	}
	manifestInfo, err := os.Lstat(tx.quarantine.manifestPath)
	if err != nil || tx.quarantine.manifestInfo == nil || !os.SameFile(tx.quarantine.manifestInfo, manifestInfo) || manifestInfo.Mode()&os.ModeSymlink != 0 || !manifestInfo.Mode().IsRegular() {
		return restoreConflict("quarantine manifest changed during restore")
	}
	return nil
}

func (tx *restoreTransaction) preflight(ctx context.Context) error {
	state, err := tx.captureStatePreimage()
	if err != nil {
		return restoreConflict("provenance state is unavailable")
	}
	if !state.exists {
		state.state = emptyTransactionProvenanceState(tx.scope)
	}
	for _, record := range state.state.Records {
		if !isCanonicalProjectSourceIdentity(record.SourceIdentity) {
			return restoreConflict("provenance identity is incompatible")
		}
	}
	tx.statePreimage = state
	tx.entries = make([]restoreEntry, 0, len(tx.quarantine.manifest.Entries))
	for _, manifestEntry := range tx.quarantine.manifest.Entries {
		if err := contextErr(ctx); err != nil {
			return restoreUnavailable("context cancelled")
		}
		managedRoot, err := tx.layout.ManagedSkillsPath(manifestEntry.Target)
		if err != nil || filepath.Clean(managedRoot) != filepath.Join(tx.root, string(manifestEntry.Target), ManagedSkillsDirectoryName) {
			return restoreConflict("managed placement path is invalid")
		}
		ancestors, err := captureApplyPathChain(tx.root, managedRoot)
		if err != nil {
			return restoreApplyError(err, "managed placement boundary changed", "managed placement boundary is unavailable")
		}
		if existing, ok := tx.ancestors[manifestEntry.Target]; ok {
			if !sameRestoreAncestorChains(existing, ancestors) {
				return restoreConflict("managed placement boundary changed")
			}
		} else {
			tx.ancestors[manifestEntry.Target] = ancestors
		}
		destination := filepath.Join(managedRoot, manifestEntry.Skill)
		if filepath.Clean(destination) != filepath.Join(filepath.Clean(managedRoot), manifestEntry.Skill) || !pathWithin(tx.root, destination) {
			return restoreConflict("managed placement path is invalid")
		}
		quarantined := filepath.Join(tx.quarantine.runPath, filepath.FromSlash(manifestEntry.QuarantinedPlacement))
		oldHash := TreeHash{Algorithm: manifestEntry.TreeHashAlgorithm, Digest: manifestEntry.OldTreeHash}
		entry := restoreEntry{
			entry: manifestEntry, destination: destination,
			quarantined: quarantined, quarantineDir: filepath.Dir(quarantined), oldHash: oldHash,
		}
		if tx.quarantine.manifest.Status == ProjectQuarantineCommitted {
			info, statErr := os.Lstat(quarantined)
			if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return restoreConflict("quarantined placement is unavailable")
			}
			hash, hashErr := HashSkillTree(quarantined)
			if hashErr != nil || hash != oldHash {
				return restoreConflict("quarantined placement content changed")
			}
			entry.quarantineInfo = info
			if _, statErr := os.Lstat(destination); statErr == nil {
				return restoreConflict("restore destination already exists")
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return restoreConflict("restore destination is unavailable")
			}
		} else {
			if _, statErr := os.Lstat(quarantined); statErr == nil {
				return restoreConflict("restored quarantine still contains an entry")
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return restoreConflict("quarantined placement is unavailable")
			}
			info, statErr := os.Lstat(destination)
			if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return restoreConflict("restored destination is unavailable")
			}
			hash, hashErr := HashSkillTree(destination)
			if hashErr != nil || hash != oldHash {
				return restoreConflict("restored destination content changed")
			}
		}
		if err := validateRestoreProvenance(state.state, manifestEntry, tx.quarantine.manifest.Status, tx.scope); err != nil {
			return err
		}
		tx.entries = append(tx.entries, entry)
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	current, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(tx.statePreimage, current) {
		return restoreConflict("provenance state changed during restore preflight")
	}
	return nil
}

func sameRestoreAncestorChains(left, right []*applyAncestor) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].path != right[index].path || left[index].exists != right[index].exists {
			return false
		}
		if left[index].exists && (left[index].info == nil || right[index].info == nil || !os.SameFile(left[index].info, right[index].info)) {
			return false
		}
	}
	return true
}

func validateRestoreProvenance(state ProvenanceState, entry ProjectQuarantineManifestEntry, status ProjectQuarantineStatus, scope Scope) error {
	record, exists := candidateRecordByKey(state, projectPlacementKey(entry.Target, entry.Skill))
	if status == ProjectQuarantineCommitted {
		if entry.Action == ProjectQuarantineEntryActionRemove {
			if exists {
				return restoreConflict("removal provenance unexpectedly exists")
			}
			return nil
		}
		if !exists || record.Scope != scope || record.Target != entry.Target || record.Skill != entry.Skill ||
			record.SourceIdentity != entry.NewSourceIdentity || record.TreeHashAlgorithm != entry.TreeHashAlgorithm || record.TreeHash != entry.NewTreeHash ||
			!isCanonicalProjectSourceIdentity(record.SourceIdentity) {
			return restoreConflict("update provenance does not match the manifest")
		}
		return nil
	}
	oldSource := entry.OldSourceIdentity
	if !exists || record.Scope != scope || record.Target != entry.Target || record.Skill != entry.Skill ||
		record.SourceIdentity != oldSource || record.TreeHashAlgorithm != entry.TreeHashAlgorithm || record.TreeHash != entry.OldTreeHash ||
		!isCanonicalProjectSourceIdentity(record.SourceIdentity) {
		return restoreConflict("restored provenance does not match the manifest")
	}
	return nil
}

func (tx *restoreTransaction) beginRestore() error {
	if tx.quarantine == nil || tx.quarantine.manifest.Status != ProjectQuarantineCommitted {
		return restoreConflict("quarantine manifest requires a different recovery workflow")
	}
	manifest := tx.quarantine.manifest
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	manifest.Status = ProjectQuarantineRestoring
	return tx.writeRestoreManifest(manifest, true)
}

func (tx *restoreTransaction) restoreEntries(ctx context.Context) error {
	for index := range tx.entries {
		if err := contextErr(ctx); err != nil {
			return restoreUnavailable("context cancelled")
		}
		if err := tx.verifyMovedPlacements(); err != nil {
			return err
		}
		if err := tx.reproveRestoreEntry(index); err != nil {
			return err
		}
		entry := tx.entries[index]
		if tx.deps.beforeRestoreMove != nil {
			if hookErr := tx.deps.beforeRestoreMove(AppliedPlacement{Skill: entry.entry.Skill, Target: entry.entry.Target}); hookErr != nil {
				return restoreUnavailable("restore move preflight failed")
			}
		}
		if err := tx.reproveRestoreEntry(index); err != nil {
			return err
		}
		if err := tx.ensureManagedAncestors(entry.entry.Target); err != nil {
			return err
		}
		if err := tx.reproveRestoreEntry(index); err != nil {
			return err
		}
		moveErr := tx.deps.PublishNoReplace(entry.quarantined, entry.destination)
		movedInfo, moved := tx.proveMovedEntry(index)
		if moved {
			tx.moved = append(tx.moved, restoredPlacement{entryIndex: index, skill: entry.entry.Skill, target: entry.entry.Target, source: entry.quarantined, destination: entry.destination, info: movedInfo, oldHash: entry.oldHash})
		}
		if moveErr != nil {
			if moved {
				return restoreUnavailable("restore move could not be verified")
			}
			if tx.restoreMoveBecameAmbiguous(entry) {
				tx.ambiguous[index] = true
				tx.moved = append(tx.moved, restoredPlacement{entryIndex: index, skill: entry.entry.Skill, target: entry.entry.Target, source: entry.quarantined, destination: entry.destination, info: entry.quarantineInfo, oldHash: entry.oldHash})
				return restoreUnavailable("restore move ownership could not be verified")
			}
			if errors.Is(moveErr, os.ErrExist) {
				return restoreConflict("restore destination appeared during move")
			}
			return restoreUnavailable("restore move could not be completed")
		}
		if !moved {
			if tx.restoreMoveBecameAmbiguous(entry) {
				tx.ambiguous[index] = true
				tx.moved = append(tx.moved, restoredPlacement{entryIndex: index, skill: entry.entry.Skill, target: entry.entry.Target, source: entry.quarantined, destination: entry.destination, info: entry.quarantineInfo, oldHash: entry.oldHash})
				return restoreUnavailable("restore move ownership could not be verified")
			}
			return restoreUnavailable("restored placement could not be verified")
		}
		if err := tx.syncRestoredPlacement(entry); err != nil {
			return err
		}
		if err := tx.verifyMovedPlacements(); err != nil {
			return err
		}
		manifest := tx.quarantine.manifest
		manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
		manifest.Entries[index].Status = ProjectQuarantineEntryRestored
		if err := tx.writeRestoreManifest(manifest, true); err != nil {
			return err
		}
	}
	return nil
}

func (tx *restoreTransaction) reproveRestoreEntry(index int) error {
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if index < 0 || index >= len(tx.entries) {
		return restoreConflict("restore entry identity is invalid")
	}
	entry := tx.entries[index]
	if err := tx.checkManagedAncestors(entry.entry.Target); err != nil {
		return err
	}
	currentState, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(tx.statePreimage, currentState) {
		return restoreConflict("provenance state changed during restore")
	}
	info, statErr := os.Lstat(entry.quarantined)
	if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || entry.quarantineInfo == nil || !os.SameFile(entry.quarantineInfo, info) {
		return restoreConflict("quarantined placement changed during restore")
	}
	hash, hashErr := HashSkillTree(entry.quarantined)
	if hashErr != nil || hash != entry.oldHash {
		return restoreConflict("quarantined placement content changed during restore")
	}
	if _, statErr := os.Lstat(entry.destination); statErr == nil {
		return restoreConflict("restore destination appeared during restore")
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return restoreConflict("restore destination is unavailable")
	}
	return nil
}

func (tx *restoreTransaction) checkManagedAncestors(target Target) error {
	values := tx.ancestors[target]
	for _, ancestor := range values {
		current, statErr := os.Lstat(ancestor.path)
		if ancestor.exists || ancestor.owned {
			if statErr != nil || ancestor.info == nil || !os.SameFile(ancestor.info, current) || current.Mode()&os.ModeSymlink != 0 || !current.IsDir() {
				return restoreConflict("managed placement boundary changed during restore")
			}
			continue
		}
		if statErr == nil {
			return restoreConflict("managed placement boundary appeared during restore")
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return restoreConflict("managed placement boundary is unavailable")
		}
	}
	return nil
}

func (tx *restoreTransaction) restoreMoveBecameAmbiguous(entry restoreEntry) bool {
	_, sourceErr := os.Lstat(entry.quarantined)
	return sourceErr != nil
}

func (tx *restoreTransaction) ensureManagedAncestors(target Target) error {
	before := len(tx.createdDirs)
	err := tx.checkRootAndAncestors(target)
	tx.createdManaged = append(tx.createdManaged, tx.createdDirs[before:]...)
	if err != nil {
		return restoreApplyError(err, "managed placement boundary changed during restore", "managed placement boundary could not be prepared")
	}
	if _, err := tx.layout.ManagedSkillsPath(target); err != nil {
		return restoreConflict("managed placement path is invalid")
	}
	return nil
}

func (tx *restoreTransaction) proveMovedEntry(index int) (os.FileInfo, bool) {
	if index < 0 || index >= len(tx.entries) {
		return nil, false
	}
	entry := tx.entries[index]
	info, err := os.Lstat(entry.destination)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || entry.quarantineInfo == nil || !os.SameFile(entry.quarantineInfo, info) {
		return info, false
	}
	if _, err := os.Lstat(entry.quarantined); !errors.Is(err, os.ErrNotExist) {
		return info, false
	}
	hash, err := HashSkillTree(entry.destination)
	if err != nil || hash != entry.oldHash {
		return info, false
	}
	return info, true
}

func (tx *restoreTransaction) verifyMovedPlacements() error {
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	for _, moved := range tx.moved {
		if tx.ambiguous[moved.entryIndex] {
			continue
		}
		if err := tx.checkManagedAncestors(moved.target); err != nil {
			return err
		}
		info, statErr := os.Lstat(moved.destination)
		if statErr != nil || moved.info == nil || !os.SameFile(moved.info, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return restoreConflict("restored placement ownership changed during restore")
		}
		hash, hashErr := HashSkillTree(moved.destination)
		if hashErr != nil || hash != moved.oldHash {
			return restoreConflict("restored placement content changed during restore")
		}
		if _, sourceErr := os.Lstat(moved.source); !errors.Is(sourceErr, os.ErrNotExist) {
			return restoreConflict("quarantine placement changed during restore")
		}
	}
	return nil
}

func (tx *restoreTransaction) syncRestoredPlacement(entry restoreEntry) error {
	if tx.deps.SyncDir == nil {
		return nil
	}
	root, err := tx.layout.ManagedSkillsPath(entry.entry.Target)
	if err != nil {
		return restoreConflict("managed placement path is invalid")
	}
	if err := tx.checkManagedAncestors(entry.entry.Target); err != nil {
		return err
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if syncErr := tx.deps.SyncDir(root); syncErr != nil {
		return restoreUnavailable("managed placement directory could not be synced")
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if err := tx.checkManagedAncestors(entry.entry.Target); err != nil {
		return err
	}
	if syncErr := tx.deps.SyncDir(entry.quarantineDir); syncErr != nil {
		return restoreUnavailable("quarantine entry directory could not be synced")
	}
	return tx.checkRestoreBoundary()
}

func (tx *restoreTransaction) writeRestoreManifest(manifest ProjectQuarantineManifest, invokeHook bool) error {
	if tx.quarantine == nil {
		return restoreUnavailable("quarantine manifest is unavailable")
	}
	if invokeHook && tx.deps.beforeRestoreManifest != nil {
		if hookErr := tx.deps.beforeRestoreManifest(manifest.Status); hookErr != nil {
			return restoreUnavailable("quarantine manifest write preflight failed")
		}
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	data, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		return restoreApplyError(err, "quarantine manifest identity is invalid", "quarantine manifest could not be encoded")
	}
	temporary, err := os.CreateTemp(tx.quarantine.runPath, ".manifest-restore-")
	if err != nil {
		return restoreUnavailable("quarantine manifest temporary file could not be created")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return restoreUnavailable("quarantine manifest permissions could not be set")
	}
	if written, err := temporary.Write(data); err != nil || written != len(data) {
		_ = temporary.Close()
		return restoreUnavailable("quarantine manifest could not be written")
	}
	if err := tx.deps.SyncFile(temporary); err != nil {
		_ = temporary.Close()
		return restoreUnavailable("quarantine manifest could not be synced")
	}
	if err := temporary.Close(); err != nil {
		return restoreUnavailable("quarantine manifest could not be closed")
	}
	currentInfo, statErr := os.Lstat(tx.quarantine.manifestPath)
	currentData, readErr := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	if statErr != nil || readErr != nil || tx.quarantine.manifestInfo == nil || !os.SameFile(tx.quarantine.manifestInfo, currentInfo) || currentInfo.Mode()&os.ModeSymlink != 0 || !currentInfo.Mode().IsRegular() || !bytes.Equal(currentData, tx.quarantine.manifestData) {
		return restoreConflict("quarantine manifest changed during restore")
	}
	if replaceErr := tx.deps.ReplaceFileAtomic(temporaryPath, tx.quarantine.manifestPath); replaceErr != nil {
		// An atomic replacement may land even when its wrapper reports an error.
		// Adopt exact landed bytes so rollback retains their ownership evidence.
		_ = tx.observeManifestData(data)
		return restoreUnavailable("quarantine manifest could not be atomically replaced")
	}
	storedInfo, storedData, valid := tx.readOwnedManifest(data, manifest.ID)
	if !valid {
		return restoreUnavailable("quarantine manifest could not be verified")
	}
	tx.quarantine.manifestInfo = storedInfo
	tx.quarantine.manifestData = append([]byte(nil), storedData...)
	tx.quarantine.manifest = manifest
	tx.manifestOwnedData = append(tx.manifestOwnedData, append([]byte(nil), storedData...))
	if err := tx.deps.SyncDir(tx.quarantine.runPath); err != nil {
		return restoreUnavailable("quarantine manifest directory could not be synced")
	}
	return nil
}

func (tx *restoreTransaction) observeManifestData(expected []byte) bool {
	info, err := os.Lstat(tx.quarantine.manifestPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false
	}
	data, err := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	if err != nil || !bytes.Equal(data, expected) {
		return false
	}
	tx.quarantine.manifestInfo = info
	tx.quarantine.manifestData = append([]byte(nil), data...)
	tx.manifestOwnedData = append(tx.manifestOwnedData, append([]byte(nil), data...))
	return true
}

func (tx *restoreTransaction) readOwnedManifest(expected []byte, id string) (os.FileInfo, []byte, bool) {
	info, err := os.Lstat(tx.quarantine.manifestPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, nil, false
	}
	data, err := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	if err != nil || !bytes.Equal(data, expected) {
		return nil, nil, false
	}
	manifest, ok := DecodeProjectQuarantineManifest(data)
	return info, data, ok && manifest.ID == id
}

func (tx *restoreTransaction) commitRestoredProvenance() error {
	if tx.statePreimage == nil {
		return restoreConflict("provenance state is unavailable")
	}
	if tx.deps.beforeRestoreCommit != nil {
		if hookErr := tx.deps.beforeRestoreCommit(); hookErr != nil {
			return restoreUnavailable("provenance commit preflight failed")
		}
	}
	if tx.deps.beforeCommit != nil {
		if hookErr := tx.deps.beforeCommit(); hookErr != nil {
			return restoreUnavailable("provenance commit preflight failed")
		}
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if err := tx.verifyMovedPlacements(); err != nil {
		return err
	}
	for _, entry := range tx.entries {
		if err := tx.checkManagedAncestors(entry.entry.Target); err != nil {
			return err
		}
	}
	current, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(tx.statePreimage, current) {
		return restoreConflict("provenance state changed before commit")
	}
	if tx.journal == nil || tx.journal.Kind != projectTransactionKindRestore {
		return restoreConflict("restore recovery evidence is unavailable")
	}
	data := append([]byte(nil), tx.journal.CandidateState...)
	tx.stateCandidateData = append([]byte(nil), data...)
	mode := os.FileMode(0o600)
	if tx.statePreimage.exists && tx.statePreimage.mode.Perm() != 0 {
		mode = tx.statePreimage.mode.Perm()
	}
	temporaryPath, err := tx.prepareStateBytes(data, mode)
	if err != nil {
		return restoreApplyError(err, "provenance state could not be prepared", "provenance state could not be prepared")
	}
	defer os.Remove(temporaryPath)
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	current, err = tx.captureStatePreimage()
	if err != nil || !sameApplyState(tx.statePreimage, current) {
		return restoreConflict("provenance state changed before commit")
	}
	commit, publishErr := tx.publishStateBytes(temporaryPath, data, tx.statePreimage)
	if commit != nil {
		tx.stateCommit = commit
	}
	if publishErr != nil {
		return restoreApplyError(publishErr, "provenance state changed during commit", "provenance state could not be committed")
	}
	stored, err := tx.captureStatePreimage()
	if err != nil || !stored.exists || !bytes.Equal(stored.data, data) || commit == nil || commit.info == nil || !os.SameFile(commit.info, stored.info) {
		return restoreUnavailable("provenance state could not be verified")
	}
	return nil
}

func buildRestoreProvenanceState(previous ProvenanceState, entries []restoreEntry, recordedAt time.Time, scope Scope) (ProvenanceState, error) {
	state := previous
	state.Version = ProvenanceStateVersion
	if scope == ScopeGlobal {
		state.Version = GlobalProvenanceStateVersion
	}
	if state.Records == nil {
		state.Records = []ProvenanceRecord{}
	}
	records := make(map[string]ProvenanceRecord, len(state.Records)+len(entries))
	for _, record := range state.Records {
		records[projectPlacementKey(record.Target, record.Skill)] = record
	}
	for _, entry := range entries {
		if !isCanonicalProjectSourceIdentity(entry.entry.OldSourceIdentity) {
			return ProvenanceState{}, restoreConflict("quarantine source identity is incompatible")
		}
		records[projectPlacementKey(entry.entry.Target, entry.entry.Skill)] = ProvenanceRecord{
			Scope: scope, Skill: entry.entry.Skill, Target: entry.entry.Target,
			SourceIdentity: entry.entry.OldSourceIdentity, TreeHashAlgorithm: entry.oldHash.Algorithm,
			TreeHash: entry.oldHash.Digest, RecordedAt: recordedAt,
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

func (tx *restoreTransaction) finalizeRestoreManifest() error {
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if err := tx.verifyMovedPlacements(); err != nil {
		return err
	}
	for _, entry := range tx.entries {
		if err := tx.checkManagedAncestors(entry.entry.Target); err != nil {
			return err
		}
	}
	manifest := tx.quarantine.manifest
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	for _, entry := range manifest.Entries {
		if entry.Status != ProjectQuarantineEntryRestored {
			return restoreConflict("quarantine manifest progress is incomplete")
		}
	}
	manifest.Status = ProjectQuarantineRestored
	if err := tx.writeRestoreManifest(manifest, true); err != nil {
		return err
	}
	return tx.verifyMovedPlacements()
}

func (tx *restoreTransaction) rollbackRestore() error {
	failed := false
	for index := len(tx.moved) - 1; index >= 0; index-- {
		moved := tx.moved[index]
		if tx.ambiguous[moved.entryIndex] {
			failed = true
			continue
		}
		if err := tx.rollbackMovedPlacement(moved); err != nil {
			failed = true
			tx.ambiguous[moved.entryIndex] = true
		}
	}
	if err := tx.restoreCommittedProvenance(); err != nil {
		failed = true
	}
	if err := tx.restoreOriginalManifest(); err != nil {
		failed = true
	}
	if err := tx.cleanupManagedDirectories(); err != nil {
		failed = true
	}
	if err := tx.verifyCommittedPreimage(); err != nil {
		failed = true
	}
	if failed {
		tx.persistRecoveryEvidence()
		return restoreUnavailable("restore rollback preserved recoverable state")
	}
	return nil
}

// verifyCommittedPreimage is the final rollback gate. Restoring only the
// placements that moved is insufficient: an untouched quarantine entry or an
// absent destination may have changed while the transaction was in progress.
// Recovery evidence must be durable whenever the complete committed preimage
// cannot be proved again.
func (tx *restoreTransaction) verifyCommittedPreimage() error {
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	failed := false
	for index, entry := range tx.entries {
		info, statErr := os.Lstat(entry.quarantined)
		if statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || entry.quarantineInfo == nil || !os.SameFile(entry.quarantineInfo, info) {
			tx.ambiguous[index] = true
			failed = true
			continue
		}
		hash, hashErr := HashSkillTree(entry.quarantined)
		if hashErr != nil || hash != entry.oldHash {
			tx.ambiguous[index] = true
			failed = true
		}
		if _, destErr := os.Lstat(entry.destination); destErr == nil || !errors.Is(destErr, os.ErrNotExist) {
			tx.ambiguous[index] = true
			failed = true
		}
	}
	currentState, stateErr := tx.captureStatePreimage()
	if stateErr != nil || !sameRestoreStateContent(tx.statePreimage, currentState) {
		failed = true
	}
	manifestInfo, manifestStatErr := os.Lstat(tx.quarantine.manifestPath)
	manifestData, manifestReadErr := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	manifest, manifestValid := DecodeProjectQuarantineManifest(manifestData)
	if manifestStatErr != nil || manifestReadErr != nil || manifestInfo.Mode()&os.ModeSymlink != 0 || !manifestInfo.Mode().IsRegular() || !bytes.Equal(manifestData, tx.originalManifestData) || !manifestValid || manifest.ID != tx.originalManifest.ID || manifest.Status != ProjectQuarantineCommitted {
		failed = true
	}
	if failed {
		return restoreUnavailable("restore committed preimage could not be verified")
	}
	return nil
}

func (tx *restoreTransaction) rollbackMovedPlacement(moved restoredPlacement) error {
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if err := tx.checkManagedAncestors(moved.target); err != nil {
		return err
	}
	if tx.deps.beforeRestoreRollback != nil {
		if hookErr := tx.deps.beforeRestoreRollback(AppliedPlacement{Skill: moved.skill, Target: moved.target}); hookErr != nil {
			return restoreUnavailable("restore rollback preflight failed")
		}
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	info, err := os.Lstat(moved.destination)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !os.SameFile(moved.info, info) {
		return restoreConflict("restore destination ownership changed during rollback")
	}
	hash, err := HashSkillTree(moved.destination)
	if err != nil || hash != moved.oldHash {
		return restoreConflict("restore destination content changed during rollback")
	}
	if _, err := os.Lstat(moved.source); err == nil {
		return restoreConflict("quarantine destination appeared during rollback")
	} else if !errors.Is(err, os.ErrNotExist) {
		return restoreConflict("quarantine destination is unavailable during rollback")
	}
	moveErr := tx.deps.PublishNoReplace(moved.destination, moved.source)
	if moveErr != nil {
		if tx.proveRolledBackPlacement(moved) {
			tx.rolledBack[moved.entryIndex] = true
			return tx.syncRollbackPlacement(moved)
		}
		return restoreApplyError(moveErr, "restore rollback destination appeared", "restore rollback move could not be completed")
	}
	if !tx.proveRolledBackPlacement(moved) {
		return restoreUnavailable("restore rollback could not verify the original tree")
	}
	tx.rolledBack[moved.entryIndex] = true
	return tx.syncRollbackPlacement(moved)
}

func (tx *restoreTransaction) proveRolledBackPlacement(moved restoredPlacement) bool {
	info, err := os.Lstat(moved.source)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !os.SameFile(moved.info, info) {
		return false
	}
	hash, err := HashSkillTree(moved.source)
	if err != nil || hash != moved.oldHash {
		return false
	}
	_, destErr := os.Lstat(moved.destination)
	return errors.Is(destErr, os.ErrNotExist)
}

func (tx *restoreTransaction) syncRollbackPlacement(moved restoredPlacement) error {
	root, err := tx.layout.ManagedSkillsPath(moved.target)
	if err != nil {
		return restoreConflict("managed placement path is invalid")
	}
	if syncErr := tx.deps.SyncDir(root); syncErr != nil {
		return restoreUnavailable("managed placement directory could not be synced during rollback")
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	if err := tx.checkManagedAncestors(moved.target); err != nil {
		return err
	}
	if syncErr := tx.deps.SyncDir(filepath.Dir(moved.source)); syncErr != nil {
		return restoreUnavailable("quarantine directory could not be synced during rollback")
	}
	return tx.checkRestoreBoundary()
}

func (tx *restoreTransaction) restoreCommittedProvenance() error {
	if tx.statePreimage == nil || len(tx.stateCandidateData) == 0 {
		return nil
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	current, err := tx.captureStatePreimage()
	if err != nil {
		return restoreUnavailable("provenance rollback could not be inspected")
	}
	if sameRestoreStateContent(tx.statePreimage, current) {
		return nil
	}
	if !current.exists || !bytes.Equal(current.data, tx.stateCandidateData) || tx.stateCommit == nil || tx.stateCommit.info == nil || !os.SameFile(tx.stateCommit.info, current.info) {
		return restoreConflict("provenance rollback ownership changed")
	}
	if tx.statePreimage.exists {
		if err := tx.checkRestoreBoundary(); err != nil {
			return err
		}
		commit, writeErr := tx.writeStateBytes(current, tx.statePreimage.data, tx.statePreimage.mode.Perm())
		if commit != nil && writeErr == nil {
			stored, readErr := tx.captureStatePreimage()
			if readErr != nil || !stored.exists || !bytes.Equal(stored.data, tx.statePreimage.data) {
				return restoreUnavailable("provenance rollback could not be verified")
			}
			if err := tx.checkRestoreBoundary(); err != nil {
				return err
			}
		}
		if writeErr != nil {
			return restoreApplyError(writeErr, "provenance rollback ownership changed", "provenance rollback could not be completed")
		}
		return nil
	}
	if removeErr := os.Remove(tx.layout.ReconcilerStatePath); removeErr != nil {
		return restoreUnavailable("provenance rollback could not remove state")
	}
	if syncErr := tx.deps.SyncDir(filepath.Dir(tx.layout.ReconcilerStatePath)); syncErr != nil {
		return restoreUnavailable("provenance rollback directory could not be synced")
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	final, statErr := tx.captureStatePreimage()
	if statErr != nil || final.exists {
		return restoreUnavailable("provenance rollback could not be verified")
	}
	return nil
}

func sameRestoreStateContent(left, right *applyStatePreimage) bool {
	if left == nil || right == nil || left.exists != right.exists {
		return false
	}
	if !left.exists {
		return true
	}
	return bytes.Equal(left.data, right.data)
}

func (tx *restoreTransaction) restoreOriginalManifest() error {
	if tx.quarantine == nil {
		return restoreUnavailable("quarantine manifest is unavailable")
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return err
	}
	currentInfo, err := os.Lstat(tx.quarantine.manifestPath)
	if err != nil || currentInfo.Mode()&os.ModeSymlink != 0 || !currentInfo.Mode().IsRegular() {
		return restoreConflict("quarantine manifest ownership changed")
	}
	currentData, err := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	if err != nil {
		return restoreConflict("quarantine manifest ownership changed")
	}
	if bytes.Equal(currentData, tx.originalManifestData) {
		if tx.originalManifestInfo == nil || !os.SameFile(tx.originalManifestInfo, currentInfo) {
			return restoreConflict("quarantine manifest ownership changed")
		}
		tx.quarantine.manifestInfo = currentInfo
		tx.quarantine.manifestData = append([]byte(nil), currentData...)
		tx.quarantine.manifest = tx.originalManifest
		return nil
	}
	owned := false
	for _, data := range tx.manifestOwnedData {
		if bytes.Equal(currentData, data) {
			owned = true
			break
		}
	}
	if !owned {
		return restoreConflict("quarantine manifest ownership changed")
	}
	if tx.quarantine.manifestInfo == nil || !os.SameFile(tx.quarantine.manifestInfo, currentInfo) {
		return restoreConflict("quarantine manifest ownership changed")
	}
	tx.quarantine.manifestInfo = currentInfo
	tx.quarantine.manifestData = append([]byte(nil), currentData...)
	if err := tx.writeRestoreManifest(tx.originalManifest, false); err != nil {
		return err
	}
	if !bytes.Equal(tx.quarantine.manifestData, tx.originalManifestData) {
		return restoreUnavailable("quarantine manifest rollback could not be verified")
	}
	return nil
}

func (tx *restoreTransaction) persistRecoveryEvidence() {
	if tx.quarantine == nil {
		return
	}
	if err := tx.checkRestoreBoundary(); err != nil {
		return
	}
	data, err := readBoundedInspectionFile(tx.quarantine.manifestPath, maxProjectQuarantineManifestSize)
	info, statErr := os.Lstat(tx.quarantine.manifestPath)
	if err != nil || statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !bytes.Equal(data, tx.quarantine.manifestData) {
		return
	}
	manifest, ok := DecodeProjectQuarantineManifest(data)
	if !ok || manifest.ID != tx.originalManifest.ID {
		return
	}
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	for index := range manifest.Entries {
		if tx.ambiguous[index] || (!tx.rolledBack[index] && index < len(tx.entries) && tx.entryWasMoved(index)) {
			manifest.Entries[index].Status = ProjectQuarantineEntryRecoveryRequired
			continue
		}
		if index < len(tx.originalManifest.Entries) {
			manifest.Entries[index].Status = tx.originalManifest.Entries[index].Status
		}
	}
	manifest.Status = ProjectQuarantineRecoveryRequired
	previousInfo, previousData := tx.quarantine.manifestInfo, tx.quarantine.manifestData
	tx.quarantine.manifestInfo, tx.quarantine.manifestData = info, append([]byte(nil), data...)
	if err := tx.writeRestoreManifest(manifest, false); err != nil {
		tx.quarantine.manifestInfo, tx.quarantine.manifestData = previousInfo, previousData
	}
}

func (tx *restoreTransaction) entryWasMoved(index int) bool {
	for _, moved := range tx.moved {
		if moved.entryIndex == index {
			return true
		}
	}
	return false
}

func (tx *restoreTransaction) cleanupManagedDirectories() error {
	failed := false
	for index := len(tx.createdManaged) - 1; index >= 0; index-- {
		created := tx.createdManaged[index]
		if err := tx.checkRestoreBoundary(); err != nil {
			failed = true
			continue
		}
		info, err := os.Lstat(created.path)
		if err != nil || created.info == nil || !os.SameFile(created.info, info) || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			failed = true
			continue
		}
		if removeErr := os.Remove(created.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			failed = true
			continue
		}
		if syncErr := tx.deps.SyncDir(filepath.Dir(created.path)); syncErr != nil {
			failed = true
		}
	}
	if failed {
		return restoreUnavailable("restore-created directory cleanup could not be verified")
	}
	return nil
}
