package sjskills

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func (tx *applyTransaction) planQuarantine(session *ProjectApplySession, preimage *applyStatePreimage, operations []PlanOperation) error {
	if tx.quarantine != nil || preimage == nil || len(operations) == 0 || len(operations) > maxProjectQuarantineEntries {
		return applyConflict("quarantine request is invalid")
	}
	createdAt := tx.deps.Now().UTC()
	if createdAt.IsZero() {
		return applyUnavailable("quarantine timestamp is unavailable")
	}
	id, err := tx.deps.newQuarantineID()
	if err != nil || !projectQuarantineIDPattern.MatchString(id) {
		return applyUnavailable("quarantine identity is unavailable")
	}
	records := make(map[string]ProvenanceRecord, len(preimage.state.Records))
	for _, record := range preimage.state.Records {
		records[projectPlacementKey(record.Target, record.Skill)] = record
	}
	desired := desiredByPlacement(session.Desired)
	entries := make([]ProjectQuarantineManifestEntry, 0, len(operations))
	for _, operation := range operations {
		key := projectPlacementKey(operation.Target, operation.Skill)
		record, managed := records[key]
		oldHash, oldHashOK := treeHashFromPlanEvidence(operation.Current)
		expectedHash, expectedHashOK := treeHashFromPlanEvidence(operation.Expected)
		if !managed || !oldHashOK ||
			record.Scope != session.Desired.Scope || record.Skill != operation.Skill || record.Target != operation.Target ||
			!isCanonicalProjectSourceIdentity(record.SourceIdentity) || record.TreeHashAlgorithm != oldHash.Algorithm || record.TreeHash != oldHash.Digest {
			return applyConflict("reviewed quarantine provenance identity changed")
		}
		entry := ProjectQuarantineManifestEntry{
			Action:               ProjectQuarantineEntryActionRemove,
			Skill:                operation.Skill,
			Target:               operation.Target,
			OriginalPlacement:    projectOriginalPlacement(operation.Target, operation.Skill),
			QuarantinedPlacement: projectQuarantinedPlacement(operation.Target, operation.Skill),
			OldSourceIdentity:    record.SourceIdentity,
			TreeHashAlgorithm:    oldHash.Algorithm,
			OldTreeHash:          oldHash.Digest,
			Status:               ProjectQuarantineEntryPending,
		}
		switch operation.Action {
		case PlanActionUpdate:
			skill, wanted := desired[key]
			newHash, newHashOK := session.Expected[skill.Name]
			newSource, sourceOK := canonicalProjectSourceIdentity(skill.Source)
			if !wanted || skill.Manager != ManagerSkillsCLI || skill.Mode != ModeCopy || !newHashOK || !sourceOK ||
				operation.Manager != ManagerSkillsCLI || operation.Source != skill.Source ||
				record.SourceIdentity != newSource || newHash != expectedHash || newHash.Algorithm != oldHash.Algorithm || newHash.Digest == oldHash.Digest {
				return applyConflict("reviewed update provenance identity changed")
			}
			entry.Action = ProjectQuarantineEntryActionUpdate
			entry.NewSourceIdentity = newSource
			entry.NewTreeHash = newHash.Digest
		case PlanActionQuarantine:
			if _, wanted := desired[key]; wanted || operation.Manager != ManagerSkillsCLI || operation.SourceID != "" || operation.Source != record.SourceIdentity ||
				!expectedHashOK || expectedHash != oldHash {
				return applyConflict("reviewed removal provenance identity changed")
			}
		default:
			return applyConflict("unsupported quarantine operation")
		}
		entries = append(entries, entry)
	}
	runPath := filepath.Join(tx.layout.QuarantinePath, id)
	if !pathWithin(tx.layout.QuarantinePath, runPath) {
		return applyConflict("quarantine path escapes derived state")
	}
	tx.quarantine = &projectQuarantineTransaction{
		runPath: runPath, ancestors: make(map[string]os.FileInfo), manifestPath: filepath.Join(runPath, applyManifestName),
		manifest: ProjectQuarantineManifest{Version: ProjectQuarantineManifestVersion, ID: id, CreatedAt: createdAt, Status: ProjectQuarantinePrepared, Entries: entries},
	}
	return nil
}

func (tx *applyTransaction) publishPreparedQuarantine(operations []PlanOperation) error {
	if tx.quarantine == nil || tx.quarantine.runDurable {
		return applyConflict("prepared quarantine is unavailable")
	}
	created, err := ensureApplyDirectory(tx.root, tx.layout.QuarantinePath)
	tx.createdDirs = append(tx.createdDirs, created...)
	if err != nil {
		return err
	}
	if err := tx.syncCreatedDirectoryParents(created); err != nil {
		return err
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	quarantineRootInfo, err := os.Lstat(tx.layout.QuarantinePath)
	if err != nil || !quarantineRootInfo.IsDir() || quarantineRootInfo.Mode()&os.ModeSymlink != 0 {
		return applyConflict("quarantine root changed before run creation")
	}
	runPath := tx.quarantine.runPath
	if err := os.Mkdir(runPath, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return applyConflict("quarantine identity already exists")
		}
		return applyUnavailable("quarantine run could not be created")
	}
	runInfo, err := os.Lstat(runPath)
	if err != nil || !runInfo.IsDir() || runInfo.Mode()&os.ModeSymlink != 0 {
		return applyUnavailable("quarantine run could not be verified")
	}
	quarantine := tx.quarantine
	quarantine.rootInfo = quarantineRootInfo
	quarantine.runInfo = runInfo
	if err := tx.deps.SyncDir(tx.layout.QuarantinePath); err != nil {
		return applyUnavailable("quarantine root could not be synced")
	}
	quarantine.runDurable = true
	if tx.deps.afterQuarantineRunSync != nil {
		tx.deps.afterQuarantineRunSync()
	}
	if err := tx.writeQuarantineManifest(quarantine, quarantine.manifest); err != nil {
		return err
	}
	for _, operation := range operations {
		parent := filepath.Join(runPath, "entries", string(operation.Target))
		if err := ensureOwnedQuarantineDirectory(runPath, parent, tx.deps.SyncDir); err != nil {
			return err
		}
		if err := captureQuarantineEntryAncestors(runPath, parent, quarantine.ancestors); err != nil {
			return err
		}
	}
	return nil
}

func ensureOwnedQuarantineDirectory(runPath, target string, syncDir func(string) error) error {
	if !pathWithin(runPath, target) {
		return applyConflict("quarantine entry path escapes its run")
	}
	relative, err := filepath.Rel(runPath, target)
	if err != nil {
		return applyConflict("quarantine entry path is invalid")
	}
	current := runPath
	for _, part := range splitApplyPath(relative) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			if mkdirErr := os.Mkdir(current, 0o700); mkdirErr != nil {
				return applyUnavailable("quarantine entry directory could not be created")
			}
			info, statErr = os.Lstat(current)
			if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return applyUnavailable("quarantine entry directory could not be verified")
			}
			if syncErr := syncDir(filepath.Dir(current)); syncErr != nil {
				return applyUnavailable("quarantine entry directory could not be synced")
			}
			continue
		}
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return applyConflict("quarantine entry directory changed")
		}
	}
	return nil
}

func splitApplyPath(relative string) []string {
	if relative == "." || relative == "" {
		return nil
	}
	result := make([]string, 0)
	for relative != "." && relative != "" {
		directory, base := filepath.Split(relative)
		if base != "" {
			result = append([]string{base}, result...)
		}
		relative = filepath.Clean(directory)
	}
	return result
}

func (tx *applyTransaction) quarantineUpdate(session *ProjectApplySession, preimage *applyStatePreimage, operation PlanOperation, destination string) error {
	return tx.quarantineExisting(session, preimage, operation, destination, ProjectQuarantineEntryActionUpdate)
}

func (tx *applyTransaction) quarantineRemoval(session *ProjectApplySession, preimage *applyStatePreimage, operation PlanOperation, destination string) error {
	return tx.quarantineExisting(session, preimage, operation, destination, ProjectQuarantineEntryActionRemove)
}

func (tx *applyTransaction) quarantineExisting(session *ProjectApplySession, preimage *applyStatePreimage, operation PlanOperation, destination string, action ProjectQuarantineEntryAction) error {
	if tx.quarantine == nil {
		return applyUnavailable("quarantine is unavailable")
	}
	if err := tx.checkQuarantineBoundary(); err != nil {
		return err
	}
	index := tx.quarantineEntryIndex(operation.Target, operation.Skill)
	if index < 0 || tx.quarantine.manifest.Entries[index].Action != action || tx.quarantine.manifest.Entries[index].Status != ProjectQuarantineEntryPending {
		return applyConflict("quarantine entry identity changed")
	}
	currentState, err := tx.captureStatePreimage()
	if err != nil || !sameApplyState(preimage, currentState) {
		return applyConflict("provenance state changed before quarantine")
	}
	if err := tx.checkRootAndAncestors(operation.Target); err != nil {
		return err
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	oldHash, ok := treeHashFromPlanEvidence(operation.Current)
	if !ok {
		return applyConflict("reviewed quarantine preimage identity changed")
	}
	expectedHash, expectedOK := treeHashFromPlanEvidence(operation.Expected)
	record, recordOK := quarantineRecord(preimage.state, operation.Target, operation.Skill)
	entry := tx.quarantine.manifest.Entries[index]
	if !recordOK || record.Scope != session.Desired.Scope || record.Skill != operation.Skill || record.Target != operation.Target ||
		record.SourceIdentity != entry.OldSourceIdentity || record.TreeHashAlgorithm != oldHash.Algorithm || record.TreeHash != oldHash.Digest ||
		(action == ProjectQuarantineEntryActionRemove &&
			(operation.Manager != ManagerSkillsCLI || operation.SourceID != "" || operation.Source != record.SourceIdentity || !expectedOK || expectedHash != oldHash)) {
		return applyConflict("quarantine provenance identity changed before move")
	}
	if action == ProjectQuarantineEntryActionUpdate {
		skill, wanted := desiredByPlacement(session.Desired)[projectPlacementKey(operation.Target, operation.Skill)]
		newSource, sourceOK := canonicalProjectSourceIdentity(skill.Source)
		newHash, hashOK := session.Expected[skill.Name]
		if !wanted || skill.Manager != ManagerSkillsCLI || skill.Mode != ModeCopy || !sourceOK || !hashOK ||
			operation.SourceID != skill.SourceID || operation.Source != skill.Source ||
			newSource != entry.NewSourceIdentity || newHash.Digest != entry.NewTreeHash || newHash.Algorithm != entry.TreeHashAlgorithm {
			return applyConflict("update provenance identity changed before move")
		}
	}
	managedRoot, rootErr := tx.layout.ManagedSkillsPath(operation.Target)
	if rootErr != nil || filepath.Clean(destination) != filepath.Join(filepath.Clean(managedRoot), operation.Skill) || !pathWithin(tx.root, destination) {
		return applyConflict("quarantine destination path changed before move")
	}
	originalInfo, err := os.Lstat(destination)
	if err != nil || originalInfo.Mode()&os.ModeSymlink != 0 || !originalInfo.IsDir() {
		return applyConflict("managed quarantine placement changed before move")
	}
	currentHash, err := HashSkillTree(destination)
	if err != nil || currentHash != oldHash {
		return applyConflict("managed quarantine content changed before move")
	}
	quarantined := filepath.Join(tx.quarantine.runPath, filepath.FromSlash(entry.QuarantinedPlacement))
	if !pathWithin(tx.quarantine.runPath, quarantined) {
		return applyConflict("quarantine destination escapes its run")
	}
	if err := tx.checkQuarantineEntryAncestors(filepath.Dir(quarantined)); err != nil {
		return err
	}
	if _, err := os.Lstat(quarantined); err == nil || !errors.Is(err, os.ErrNotExist) {
		return applyConflict("quarantine destination already exists")
	}
	if err := tx.checkRootAndAncestors(operation.Target); err != nil {
		return err
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	if err := tx.deps.PublishNoReplace(destination, quarantined); err != nil {
		if errors.Is(err, os.ErrExist) {
			return applyConflict("quarantine destination already exists")
		}
		return applyUnavailable("managed placement could not be quarantined")
	}
	quarantineInfo, statErr := os.Lstat(quarantined)
	movedHash, hashErr := HashSkillTree(quarantined)
	placement := quarantinedPlacement{
		skill: operation.Skill, target: operation.Target, dest: destination, quarantined: quarantined,
		originalInfo: originalInfo, oldHash: oldHash, manifestIndex: index, action: operation.Action,
	}
	tx.quarantined = append(tx.quarantined, placement)
	if statErr != nil || quarantineInfo.Mode()&os.ModeSymlink != 0 || !quarantineInfo.IsDir() ||
		!os.SameFile(originalInfo, quarantineInfo) || hashErr != nil || movedHash != oldHash {
		return applyUnavailable("quarantined placement could not be verified")
	}
	root, _ := tx.layout.ManagedSkillsPath(operation.Target)
	if err := tx.deps.SyncDir(root); err != nil {
		return applyUnavailable("managed root could not be synced after quarantine")
	}
	if err := tx.deps.SyncDir(filepath.Dir(quarantined)); err != nil {
		return applyUnavailable("quarantine entry could not be synced")
	}
	return tx.setQuarantineEntryStatus(operation.Target, operation.Skill, ProjectQuarantineEntryQuarantined)
}

func captureQuarantineEntryAncestors(runPath, parent string, ancestors map[string]os.FileInfo) error {
	if ancestors == nil {
		return applyConflict("quarantine entry directory identity is unavailable")
	}
	if !pathWithin(runPath, parent) {
		return applyConflict("quarantine entry path escapes its run")
	}
	relative, err := filepath.Rel(filepath.Clean(runPath), filepath.Clean(parent))
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return applyConflict("quarantine entry path is invalid")
	}
	current := filepath.Clean(runPath)
	for _, part := range splitApplyPath(relative) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return applyConflict("quarantine entry directory changed")
		}
		if expected, exists := ancestors[current]; exists && !os.SameFile(expected, info) {
			return applyConflict("quarantine entry directory changed")
		}
		ancestors[current] = info
	}
	return nil
}

func (tx *applyTransaction) checkQuarantineEntryAncestors(parent string) error {
	if tx.quarantine == nil {
		return applyUnavailable("quarantine is unavailable")
	}
	if err := captureQuarantineEntryAncestors(tx.quarantine.runPath, parent, tx.quarantine.ancestors); err != nil {
		return err
	}
	return nil
}

func (tx *applyTransaction) quarantineEntryIndex(target Target, skill string) int {
	if tx.quarantine == nil {
		return -1
	}
	for index, entry := range tx.quarantine.manifest.Entries {
		if entry.Target == target && entry.Skill == skill {
			return index
		}
	}
	return -1
}

func quarantineRecord(state ProvenanceState, target Target, skill string) (ProvenanceRecord, bool) {
	for _, record := range state.Records {
		if record.Target == target && record.Skill == skill {
			return record, true
		}
	}
	return ProvenanceRecord{}, false
}

func (tx *applyTransaction) setQuarantineEntryStatus(target Target, skill string, status ProjectQuarantineEntryStatus) error {
	index := tx.quarantineEntryIndex(target, skill)
	if index < 0 {
		return applyConflict("quarantine manifest entry disappeared")
	}
	manifest := tx.quarantine.manifest
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	manifest.Entries[index].Status = status
	if manifest.Status == ProjectQuarantinePrepared {
		manifest.Status = ProjectQuarantineActive
	}
	if err := tx.writeQuarantineManifest(tx.quarantine, manifest); err != nil {
		return err
	}
	tx.quarantine.manifest = manifest
	return nil
}

func (tx *applyTransaction) setQuarantineStatus(status ProjectQuarantineStatus) error {
	if tx.quarantine == nil {
		return nil
	}
	manifest := tx.quarantine.manifest
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	manifest.Status = status
	if err := tx.writeQuarantineManifest(tx.quarantine, manifest); err != nil {
		return err
	}
	tx.quarantine.manifest = manifest
	return nil
}

func (tx *applyTransaction) commitQuarantine() error {
	if tx.quarantine == nil {
		return nil
	}
	for _, entry := range tx.quarantine.manifest.Entries {
		want := ProjectQuarantineEntryReplaced
		if entry.Action == ProjectQuarantineEntryActionRemove {
			want = ProjectQuarantineEntryQuarantined
		}
		if entry.Status != want {
			return applyConflict("quarantine manifest is incomplete before commit")
		}
	}
	return tx.setQuarantineStatus(ProjectQuarantineCommitted)
}

func (tx *applyTransaction) writeQuarantineManifest(quarantine *projectQuarantineTransaction, manifest ProjectQuarantineManifest) error {
	if err := tx.checkLock(); err != nil {
		return err
	}
	if err := tx.checkQuarantineBoundaryFor(quarantine); err != nil {
		return err
	}
	data, err := marshalProjectQuarantineManifest(manifest)
	if err != nil {
		return err
	}
	temporaryParent := quarantine.runPath
	temporaryPattern := ".manifest-"
	if quarantine.manifestInfo == nil {
		// Keep initial-publication staging outside the durable run. A process
		// death can then leave only an exactly empty run, which the journal can
		// recover without deleting an unknown entry.
		temporaryParent = tx.layout.DerivedDirectoryPath
		temporaryPattern = ".quarantine-manifest-" + manifest.ID + "-"
	}
	var temporary *os.File
	if quarantine.manifestInfo == nil {
		temporary, err = os.OpenFile(filepath.Join(temporaryParent, temporaryPattern+"staged"), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	} else {
		temporary, err = os.CreateTemp(temporaryParent, temporaryPattern)
	}
	if err != nil {
		return applyUnavailable("quarantine manifest temporary file could not be created")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return applyUnavailable("quarantine manifest permissions could not be set")
	}
	if written, err := temporary.Write(data); err != nil || written != len(data) {
		_ = temporary.Close()
		return applyUnavailable("quarantine manifest could not be written")
	}
	if err := tx.deps.SyncFile(temporary); err != nil {
		_ = temporary.Close()
		return applyUnavailable("quarantine manifest could not be synced")
	}
	if err := temporary.Close(); err != nil {
		return applyUnavailable("quarantine manifest could not be closed")
	}
	if quarantine.manifestInfo == nil && tx.deps.afterInitialManifestSync != nil {
		tx.deps.afterInitialManifestSync()
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	if err := tx.checkQuarantineBoundaryFor(quarantine); err != nil {
		return err
	}
	if quarantine.manifestInfo == nil {
		if _, err := os.Lstat(quarantine.manifestPath); err == nil || !errors.Is(err, os.ErrNotExist) {
			return applyConflict("quarantine manifest destination appeared")
		}
		if err := tx.deps.PublishNoReplace(temporaryPath, quarantine.manifestPath); err != nil {
			if tx.observeApplyQuarantineManifest(quarantine, manifest, data) {
				return applyUnavailable("quarantine manifest publication was not confirmed")
			}
			if errors.Is(err, os.ErrExist) {
				return applyConflict("quarantine manifest destination appeared")
			}
			return applyUnavailable("quarantine manifest could not be published")
		}
	} else {
		currentInfo, statErr := os.Lstat(quarantine.manifestPath)
		currentData, readErr := readBoundedInspectionFile(quarantine.manifestPath, maxProjectQuarantineManifestSize)
		if statErr != nil || currentInfo.Mode()&os.ModeSymlink != 0 || !currentInfo.Mode().IsRegular() ||
			!os.SameFile(quarantine.manifestInfo, currentInfo) || readErr != nil || !bytes.Equal(currentData, quarantine.manifestData) {
			return applyConflict("quarantine manifest changed during transaction")
		}
		if err := tx.deps.ReplaceFileAtomic(temporaryPath, quarantine.manifestPath); err != nil {
			if tx.observeApplyQuarantineManifest(quarantine, manifest, data) {
				return applyUnavailable("quarantine manifest replacement was not confirmed")
			}
			return applyUnavailable("quarantine manifest could not be atomically replaced")
		}
	}
	stored, err := readBoundedInspectionFile(quarantine.manifestPath, maxProjectQuarantineManifestSize)
	storedInfo, statErr := os.Lstat(quarantine.manifestPath)
	decoded, valid := DecodeProjectQuarantineManifest(stored)
	if err != nil || statErr != nil || storedInfo.Mode()&os.ModeSymlink != 0 || !storedInfo.Mode().IsRegular() ||
		!valid || decoded.ID != manifest.ID || !bytes.Equal(stored, data) {
		return applyUnavailable("quarantine manifest could not be verified")
	}
	quarantine.manifestInfo = storedInfo
	quarantine.manifestData = append([]byte(nil), data...)
	if err := tx.deps.SyncDir(quarantine.runPath); err != nil {
		return applyUnavailable("quarantine manifest directory could not be synced")
	}
	if !quarantine.runDurable {
		if err := tx.deps.SyncDir(tx.layout.QuarantinePath); err != nil {
			return applyUnavailable("quarantine root could not be synced")
		}
		quarantine.runDurable = true
	}
	return nil
}

func (tx *applyTransaction) observeApplyQuarantineManifest(quarantine *projectQuarantineTransaction, manifest ProjectQuarantineManifest, expected []byte) bool {
	stored, readErr := readBoundedInspectionFile(quarantine.manifestPath, maxProjectQuarantineManifestSize)
	info, statErr := os.Lstat(quarantine.manifestPath)
	decoded, valid := DecodeProjectQuarantineManifest(stored)
	if readErr != nil || statErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() ||
		!valid || decoded.ID != manifest.ID || !bytes.Equal(stored, expected) {
		return false
	}
	quarantine.manifestInfo = info
	quarantine.manifestData = append([]byte(nil), expected...)
	quarantine.manifest = manifest
	return true
}

func (tx *applyTransaction) checkQuarantineBoundary() error {
	if tx.quarantine == nil {
		return applyUnavailable("update quarantine is unavailable")
	}
	return tx.checkQuarantineBoundaryFor(tx.quarantine)
}

func (tx *applyTransaction) checkQuarantineBoundaryFor(quarantine *projectQuarantineTransaction) error {
	if quarantine == nil || !pathWithin(tx.layout.QuarantinePath, quarantine.runPath) ||
		!pathWithin(quarantine.runPath, quarantine.manifestPath) {
		return applyConflict("quarantine boundary is invalid")
	}
	rootInfo, rootErr := os.Lstat(tx.layout.QuarantinePath)
	runInfo, runErr := os.Lstat(quarantine.runPath)
	if rootErr != nil || runErr != nil || quarantine.rootInfo == nil || quarantine.runInfo == nil ||
		!os.SameFile(quarantine.rootInfo, rootInfo) || !os.SameFile(quarantine.runInfo, runInfo) ||
		rootInfo.Mode()&os.ModeSymlink != 0 || runInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() || !runInfo.IsDir() {
		return applyConflict("quarantine boundary changed during transaction")
	}
	return nil
}
