package sjskills

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	projectTransactionJournalVersion = 1
	projectTransactionJournalName    = "transaction.json"
	projectRecoveryDirectoryName     = "recovery"
	maxProjectTransactionJournalSize = 6 << 20
	projectTransactionKindApply      = "apply"
	projectTransactionKindRestore    = "restore"
)

type projectJournalState struct {
	Exists bool   `json:"exists"`
	Mode   uint32 `json:"mode,omitempty"`
	Data   []byte `json:"data,omitempty"`
}

type projectJournalEntry struct {
	Action               PlanAction `json:"action"`
	Skill                string     `json:"skill"`
	Target               Target     `json:"target"`
	ManagedParentExisted bool       `json:"managedParentExisted"`
	ManagedRootExisted   bool       `json:"managedRootExisted"`
	OldSourceIdentity    string     `json:"oldSourceIdentity,omitempty"`
	NewSourceIdentity    string     `json:"newSourceIdentity,omitempty"`
	TreeHashAlgorithm    string     `json:"treeHashAlgorithm"`
	OldTreeHash          string     `json:"oldTreeHash,omitempty"`
	NewTreeHash          string     `json:"newTreeHash,omitempty"`
}

// projectTransactionJournal is private crash-recovery evidence. It contains
// only modeled relative identities and bounded provenance bytes; absolute
// process paths and materialization paths are never persisted.
type projectTransactionJournal struct {
	Version        int                   `json:"version"`
	ID             string                `json:"id"`
	Kind           string                `json:"kind"`
	QuarantineID   string                `json:"quarantineId,omitempty"`
	PreState       projectJournalState   `json:"preState"`
	CandidateState []byte                `json:"candidateState"`
	PreManifest    []byte                `json:"preManifest,omitempty"`
	Entries        []projectJournalEntry `json:"entries"`
}

func projectTransactionJournalPath(layout DerivedLayout) string {
	return filepath.Join(layout.DerivedDirectoryPath, projectTransactionJournalName)
}

func projectRecoveryRunPath(layout DerivedLayout, id string) string {
	return filepath.Join(layout.DerivedDirectoryPath, projectRecoveryDirectoryName, id)
}

func newApplyJournal(preimage *applyStatePreimage, session *ProjectApplySession, candidateData []byte, id, quarantineID string) (projectTransactionJournal, error) {
	if preimage == nil || session == nil {
		return projectTransactionJournal{}, applyUnavailable("transaction recovery evidence is unavailable")
	}
	journal := projectTransactionJournal{
		Version:        projectTransactionJournalVersion,
		ID:             id,
		Kind:           projectTransactionKindApply,
		QuarantineID:   quarantineID,
		CandidateState: append([]byte(nil), candidateData...),
		Entries:        make([]projectJournalEntry, 0),
	}
	journal.PreState.Exists = preimage.exists
	if preimage.exists {
		journal.PreState.Mode = uint32(preimage.mode.Perm())
		journal.PreState.Data = append([]byte(nil), preimage.data...)
	}
	candidate, valid := decodeTransactionProvenanceState(candidateData, session.Desired.Scope, false)
	if !valid {
		return projectTransactionJournal{}, applyConflict("transaction candidate state is invalid")
	}
	preRecords := recoveryRecordMap(preimage.state)
	candidateRecords := recoveryRecordMap(candidate)
	for _, operation := range reviewedMutationOperations(session) {
		entry := projectJournalEntry{Action: operation.Action, Skill: operation.Skill, Target: operation.Target}
		key := projectPlacementKey(operation.Target, operation.Skill)
		if oldHash, ok := treeHashFromPlanEvidence(operation.Current); ok && operation.Action != PlanActionInstall {
			entry.TreeHashAlgorithm = oldHash.Algorithm
			entry.OldTreeHash = oldHash.Digest
			entry.OldSourceIdentity = preRecords[key].SourceIdentity
		}
		if operation.Action != PlanActionQuarantine {
			newHash, ok := session.Expected[operation.Skill]
			if !ok {
				return projectTransactionJournal{}, applyUnavailable("transaction recovery hash is unavailable")
			}
			if entry.TreeHashAlgorithm == "" {
				entry.TreeHashAlgorithm = newHash.Algorithm
			}
			entry.NewTreeHash = newHash.Digest
			entry.NewSourceIdentity = candidateRecords[key].SourceIdentity
		}
		journal.Entries = append(journal.Entries, entry)
	}
	if !validProjectTransactionJournal(journal) {
		return projectTransactionJournal{}, applyConflict("transaction recovery evidence is invalid")
	}
	return journal, nil
}

func (tx *applyTransaction) prepareTransactionJournal(preimage *applyStatePreimage, session *ProjectApplySession) error {
	if tx.journal != nil {
		return applyConflict("transaction recovery evidence already exists")
	}
	now := tx.deps.Now()
	if now.IsZero() {
		return applyUnavailable("provenance timestamp is unavailable")
	}
	candidate, err := buildApplyState(preimage.state, session, now.UTC())
	if err != nil {
		return err
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		return applyUnavailable("provenance state could not be encoded")
	}
	transactionID := ""
	quarantineID := ""
	if tx.quarantine != nil {
		transactionID = tx.quarantine.manifest.ID
		quarantineID = transactionID
	} else {
		transactionID, err = tx.deps.newQuarantineID()
		if err != nil || !projectQuarantineIDPattern.MatchString(transactionID) {
			return applyUnavailable("transaction recovery identity could not be created")
		}
	}
	journal, err := newApplyJournal(preimage, session, candidateData, transactionID, quarantineID)
	if err != nil {
		return err
	}
	for index := range journal.Entries {
		journal.Entries[index].ManagedParentExisted, journal.Entries[index].ManagedRootExisted = applyManagedAncestorExistence(tx.ancestors[journal.Entries[index].Target])
	}
	return tx.writeTransactionJournal(journal)
}

func (tx *restoreTransaction) prepareRestoreJournal() error {
	if tx.statePreimage == nil || tx.quarantine == nil {
		return restoreConflict("restore recovery evidence is unavailable")
	}
	now := tx.deps.Now().UTC()
	if now.IsZero() {
		return restoreUnavailable("provenance timestamp is unavailable")
	}
	candidate, err := buildRestoreProvenanceState(tx.statePreimage.state, tx.entries, now, tx.scope)
	if err != nil {
		return err
	}
	candidateData, err := marshalApplyState(candidate)
	if err != nil {
		return restoreUnavailable("provenance state could not be encoded")
	}
	journal := projectTransactionJournal{
		Version: projectTransactionJournalVersion, ID: tx.quarantine.manifest.ID,
		Kind: projectTransactionKindRestore, QuarantineID: tx.quarantine.manifest.ID,
		CandidateState: candidateData, PreManifest: append([]byte(nil), tx.originalManifestData...),
		Entries: make([]projectJournalEntry, 0, len(tx.entries)),
	}
	journal.PreState.Exists = tx.statePreimage.exists
	if tx.statePreimage.exists {
		journal.PreState.Mode = uint32(tx.statePreimage.mode.Perm())
		journal.PreState.Data = append([]byte(nil), tx.statePreimage.data...)
	}
	for _, restoreEntry := range tx.entries {
		parentExisted, rootExisted := applyManagedAncestorExistence(tx.ancestors[restoreEntry.entry.Target])
		entry := projectJournalEntry{
			Skill: restoreEntry.entry.Skill, Target: restoreEntry.entry.Target,
			ManagedParentExisted: parentExisted,
			ManagedRootExisted:   rootExisted,
			TreeHashAlgorithm:    restoreEntry.entry.TreeHashAlgorithm,
			OldTreeHash:          restoreEntry.entry.OldTreeHash,
			OldSourceIdentity:    restoreEntry.entry.OldSourceIdentity,
		}
		if restoreEntry.entry.Action == ProjectQuarantineEntryActionUpdate {
			entry.Action = PlanActionUpdate
			entry.NewTreeHash = restoreEntry.entry.NewTreeHash
			entry.NewSourceIdentity = restoreEntry.entry.NewSourceIdentity
		} else {
			entry.Action = PlanActionQuarantine
		}
		journal.Entries = append(journal.Entries, entry)
	}
	if !validProjectTransactionJournal(journal) {
		return restoreConflict("restore recovery evidence is invalid")
	}
	if err := tx.writeTransactionJournal(journal); err != nil {
		return restoreApplyError(err, "restore recovery evidence changed", "restore recovery evidence could not be committed")
	}
	return nil
}

func applyManagedAncestorExistence(ancestors []*applyAncestor) (bool, bool) {
	if len(ancestors) == 0 {
		return false, false
	}
	return ancestors[0].exists, ancestors[len(ancestors)-1].exists
}

func validProjectTransactionJournal(journal projectTransactionJournal) bool {
	if journal.Version != projectTransactionJournalVersion ||
		!projectQuarantineIDPattern.MatchString(journal.ID) ||
		(journal.QuarantineID != "" && !projectQuarantineIDPattern.MatchString(journal.QuarantineID)) ||
		len(journal.CandidateState) == 0 || int64(len(journal.CandidateState)) > maxProvenanceStateBytes ||
		journal.Entries == nil || len(journal.Entries) > maxProjectQuarantineEntries {
		return false
	}
	switch journal.Kind {
	case projectTransactionKindApply:
		if len(journal.PreManifest) != 0 {
			return false
		}
	case projectTransactionKindRestore:
		if journal.QuarantineID == "" || len(journal.PreManifest) == 0 || len(journal.PreManifest) > maxProjectQuarantineManifestSize {
			return false
		}
		manifest, valid := DecodeProjectQuarantineManifest(journal.PreManifest)
		if !valid || manifest.ID != journal.QuarantineID || manifest.Status != ProjectQuarantineCommitted {
			return false
		}
	default:
		return false
	}
	candidateScope, candidateState, valid := decodeJournalCandidateState(journal.CandidateState)
	if !valid {
		return false
	}
	if len(journal.Entries) == 0 {
		if journal.Kind != projectTransactionKindApply || candidateScope != ScopeGlobal || !journal.PreState.Exists || journal.QuarantineID != "" {
			return false
		}
		_, format, legacyValid := decodeGlobalProvenanceState(journal.PreState.Data)
		if !legacyValid || format != GlobalProvenanceLegacyV1 {
			return false
		}
	}
	preState := ProvenanceState{Version: candidateState.Version, Records: []ProvenanceRecord{}}
	if journal.PreState.Exists {
		if len(journal.PreState.Data) == 0 || int64(len(journal.PreState.Data)) > maxProvenanceStateBytes || journal.PreState.Mode > 0o777 {
			return false
		}
		decoded, valid := decodeTransactionProvenanceState(journal.PreState.Data, candidateScope, true)
		if !valid {
			return false
		}
		preState = decoded
	} else if journal.PreState.Mode != 0 || len(journal.PreState.Data) != 0 {
		return false
	}
	seen := make(map[string]struct{}, len(journal.Entries))
	type managedAncestorExistence struct {
		parent bool
		root   bool
	}
	ancestorExistenceByTarget := make(map[Target]managedAncestorExistence)
	requiresQuarantine := false
	for _, entry := range journal.Entries {
		if !isPortableName(entry.Skill) {
			return false
		}
		if _, ok := supportedTargets[entry.Target]; !ok || entry.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 {
			return false
		}
		if entry.ManagedRootExisted && !entry.ManagedParentExisted {
			return false
		}
		existence := managedAncestorExistence{parent: entry.ManagedParentExisted, root: entry.ManagedRootExisted}
		if previous, ok := ancestorExistenceByTarget[entry.Target]; ok && previous != existence {
			return false
		}
		ancestorExistenceByTarget[entry.Target] = existence
		switch entry.Action {
		case PlanActionInstall:
			if entry.OldTreeHash != "" || entry.OldSourceIdentity != "" || !validTreeDigest(entry.NewTreeHash) || !isCanonicalProjectSourceIdentity(entry.NewSourceIdentity) {
				return false
			}
		case PlanActionUpdate:
			requiresQuarantine = true
			if !validTreeDigest(entry.OldTreeHash) || !validTreeDigest(entry.NewTreeHash) || entry.OldTreeHash == entry.NewTreeHash ||
				!isCanonicalProjectSourceIdentity(entry.OldSourceIdentity) || !isCanonicalProjectSourceIdentity(entry.NewSourceIdentity) ||
				entry.OldSourceIdentity != entry.NewSourceIdentity {
				return false
			}
		case PlanActionQuarantine:
			requiresQuarantine = true
			if !validTreeDigest(entry.OldTreeHash) || entry.NewTreeHash != "" || !isCanonicalProjectSourceIdentity(entry.OldSourceIdentity) || entry.NewSourceIdentity != "" {
				return false
			}
		default:
			return false
		}
		key := projectPlacementKey(entry.Target, entry.Skill)
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
	}
	if requiresQuarantine != (journal.QuarantineID != "") {
		return false
	}
	if !journalStateTransitionMatches(journal, preState, candidateState, candidateScope) {
		return false
	}
	if journal.Kind == projectTransactionKindRestore && !journalMatchesQuarantine(journal, mustDecodeRecoveryManifest(journal.PreManifest)) {
		return false
	}
	return sort.SliceIsSorted(journal.Entries, func(i, k int) bool {
		if projectTargetRank(journal.Entries[i].Target) != projectTargetRank(journal.Entries[k].Target) {
			return projectTargetRank(journal.Entries[i].Target) < projectTargetRank(journal.Entries[k].Target)
		}
		return compareUTF16(journal.Entries[i].Skill, journal.Entries[k].Skill) < 0
	})
}

func recoveryRecordMap(state ProvenanceState) map[string]ProvenanceRecord {
	records := make(map[string]ProvenanceRecord, len(state.Records))
	for _, record := range state.Records {
		records[projectPlacementKey(record.Target, record.Skill)] = record
	}
	return records
}

func decodeJournalCandidateState(data []byte) (Scope, ProvenanceState, bool) {
	if state, valid := decodeTransactionProvenanceState(data, ScopeProject, false); valid {
		return ScopeProject, state, true
	}
	if state, valid := decodeTransactionProvenanceState(data, ScopeGlobal, false); valid {
		return ScopeGlobal, state, true
	}
	return "", ProvenanceState{}, false
}

func journalStateTransitionMatches(journal projectTransactionJournal, preState, candidateState ProvenanceState, scope Scope) bool {
	if preState.Version != candidateState.Version {
		return false
	}
	preRecords := recoveryRecordMap(preState)
	candidateRecords := recoveryRecordMap(candidateState)
	changed := make(map[string]struct{}, len(journal.Entries))
	for _, entry := range journal.Entries {
		key := projectPlacementKey(entry.Target, entry.Skill)
		changed[key] = struct{}{}
		pre, preExists := preRecords[key]
		candidate, candidateExists := candidateRecords[key]
		oldMatches := preExists && recoveryRecordMatches(pre, entry, false, scope)
		newMatches := candidateExists && recoveryRecordMatches(candidate, entry, true, scope)
		if journal.Kind == projectTransactionKindApply {
			switch entry.Action {
			case PlanActionInstall:
				if preExists || !newMatches {
					return false
				}
			case PlanActionUpdate:
				if !oldMatches || !newMatches {
					return false
				}
			case PlanActionQuarantine:
				if !oldMatches || candidateExists {
					return false
				}
			}
			continue
		}
		switch entry.Action {
		case PlanActionUpdate:
			if !preExists || pre.SourceIdentity != entry.NewSourceIdentity || pre.TreeHashAlgorithm != entry.TreeHashAlgorithm || pre.TreeHash != entry.NewTreeHash || !oldRecoveryRecordMatches(candidate, candidateExists, entry, scope) {
				return false
			}
		case PlanActionQuarantine:
			if preExists || !oldRecoveryRecordMatches(candidate, candidateExists, entry, scope) {
				return false
			}
		default:
			return false
		}
	}
	for key, pre := range preRecords {
		if _, ok := changed[key]; ok {
			continue
		}
		candidate, ok := candidateRecords[key]
		if !ok || candidate != pre {
			return false
		}
	}
	for key := range candidateRecords {
		if _, ok := changed[key]; ok {
			continue
		}
		if _, ok := preRecords[key]; !ok {
			return false
		}
	}
	return true
}

func recoveryRecordMatches(record ProvenanceRecord, entry projectJournalEntry, useNew bool, scope Scope) bool {
	source, hash := entry.OldSourceIdentity, entry.OldTreeHash
	if useNew {
		source, hash = entry.NewSourceIdentity, entry.NewTreeHash
	}
	return record.Scope == scope && record.Skill == entry.Skill && record.Target == entry.Target &&
		record.SourceIdentity == source && record.TreeHashAlgorithm == entry.TreeHashAlgorithm && record.TreeHash == hash
}

func oldRecoveryRecordMatches(record ProvenanceRecord, exists bool, entry projectJournalEntry, scope Scope) bool {
	return exists && recoveryRecordMatches(record, entry, false, scope)
}

func mustDecodeRecoveryManifest(data []byte) ProjectQuarantineManifest {
	manifest, _ := DecodeProjectQuarantineManifest(data)
	return manifest
}

func validTreeDigest(value string) bool {
	return lowercaseDigestPattern.MatchString(value)
}

func marshalProjectTransactionJournal(journal projectTransactionJournal) ([]byte, error) {
	if !validProjectTransactionJournal(journal) {
		return nil, applyConflict("transaction recovery evidence is invalid")
	}
	data, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return nil, applyUnavailable("transaction recovery evidence could not be encoded")
	}
	data = append(data, '\n')
	if len(data) > maxProjectTransactionJournalSize {
		return nil, applyUnavailable("transaction recovery evidence exceeds its size bound")
	}
	return data, nil
}

func decodeProjectTransactionJournal(data []byte) (projectTransactionJournal, bool) {
	if len(data) == 0 || len(data) > maxProjectTransactionJournalSize {
		return projectTransactionJournal{}, false
	}
	var journal projectTransactionJournal
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&journal); err != nil {
		return projectTransactionJournal{}, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF || !validProjectTransactionJournal(journal) {
		return projectTransactionJournal{}, false
	}
	journal.PreState.Data = append([]byte(nil), journal.PreState.Data...)
	journal.CandidateState = append([]byte(nil), journal.CandidateState...)
	journal.PreManifest = append([]byte(nil), journal.PreManifest...)
	journal.Entries = append([]projectJournalEntry(nil), journal.Entries...)
	return journal, true
}

func (tx *applyTransaction) writeTransactionJournal(journal projectTransactionJournal) error {
	path := projectTransactionJournalPath(tx.layout)
	if err := tx.checkLock(); err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil || !errors.Is(err, os.ErrNotExist) {
		return applyConflict("transaction recovery evidence already exists")
	}
	data, err := marshalProjectTransactionJournal(journal)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(tx.layout.DerivedDirectoryPath, ".transaction-")
	if err != nil {
		return applyUnavailable("transaction recovery evidence could not be created")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return applyUnavailable("transaction recovery evidence permissions could not be set")
	}
	if written, err := temporary.Write(data); err != nil || written != len(data) {
		_ = temporary.Close()
		return applyUnavailable("transaction recovery evidence could not be written")
	}
	if err := tx.deps.SyncFile(temporary); err != nil {
		_ = temporary.Close()
		return applyUnavailable("transaction recovery evidence could not be synced")
	}
	temporaryInfo, err := temporary.Stat()
	if err != nil {
		_ = temporary.Close()
		return applyUnavailable("transaction recovery evidence identity could not be captured")
	}
	if err := temporary.Close(); err != nil {
		return applyUnavailable("transaction recovery evidence could not be closed")
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	if err := publishNoReplace(temporaryPath, path); err != nil {
		if stored, readErr := readBoundedInspectionFile(path, maxProjectTransactionJournalSize); readErr == nil && bytes.Equal(stored, data) {
			info, statErr := os.Lstat(path)
			if statErr != nil || !os.SameFile(temporaryInfo, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) {
				return applyUnavailable("transaction recovery evidence publication was not confirmed")
			}
			tx.journal = &journal
			tx.journalData = append([]byte(nil), data...)
			tx.journalInfo = info
			return applyUnavailable("transaction recovery evidence publication was not confirmed")
		}
		if errors.Is(err, os.ErrExist) {
			return applyConflict("transaction recovery evidence already exists")
		}
		return applyUnavailable("transaction recovery evidence could not be published")
	}
	stored, readErr := readBoundedInspectionFile(path, maxProjectTransactionJournalSize)
	info, statErr := os.Lstat(path)
	if readErr != nil || statErr != nil || !os.SameFile(temporaryInfo, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) || !bytes.Equal(stored, data) {
		return applyUnavailable("transaction recovery evidence could not be verified")
	}
	if err := tx.deps.SyncDir(tx.layout.DerivedDirectoryPath); err != nil {
		return applyUnavailable("transaction recovery directory could not be synced")
	}
	tx.journal = &journal
	tx.journalData = append([]byte(nil), data...)
	tx.journalInfo = info
	return nil
}

func (tx *applyTransaction) loadTransactionJournal() (projectTransactionJournal, bool, error) {
	path := projectTransactionJournalPath(tx.layout)
	probe := proveInspectionPath(tx.root, path)
	if probe.unsafe {
		return projectTransactionJournal{}, false, applyConflict("transaction recovery evidence is unsafe")
	}
	if !probe.exists {
		return projectTransactionJournal{}, false, nil
	}
	if probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() || !validApplyPrivateFileMode(probe.info.Mode()) {
		return projectTransactionJournal{}, false, applyConflict("transaction recovery evidence is not a regular file")
	}
	data, err := readBoundedInspectionFile(path, maxProjectTransactionJournalSize)
	if err != nil {
		return projectTransactionJournal{}, false, applyConflict("transaction recovery evidence could not be read")
	}
	journal, ok := decodeProjectTransactionJournal(data)
	if !ok {
		return projectTransactionJournal{}, false, applyConflict("transaction recovery evidence is malformed")
	}
	tx.journal = &journal
	tx.journalData = append([]byte(nil), data...)
	tx.journalInfo = probe.info
	return journal, true, nil
}

func (tx *applyTransaction) clearTransactionJournal() error {
	if tx.journal == nil {
		return nil
	}
	path := projectTransactionJournalPath(tx.layout)
	digest := sha256.Sum256(tx.journalData)
	tombstoneName := "transaction-" + tx.journal.Kind + "-" + hex.EncodeToString(digest[:]) + ".cleared"
	tombstone := filepath.Join(projectRecoveryRunPath(tx.layout, tx.journal.ID), tombstoneName)
	if err := ensureRecoveryParent(tx, filepath.Dir(tombstone)); err != nil {
		return err
	}
	data, err := readBoundedInspectionFile(path, maxProjectTransactionJournalSize)
	if err != nil || !bytes.Equal(data, tx.journalData) {
		return applyConflict("transaction recovery evidence changed")
	}
	if tx.deps.beforeJournalClear != nil {
		tx.deps.beforeJournalClear()
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil || tx.journalInfo == nil || !os.SameFile(tx.journalInfo, info) || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || !validApplyPrivateFileMode(info.Mode()) {
		return applyConflict("transaction recovery evidence changed")
	}
	if _, err := os.Lstat(tombstone); err == nil || !errors.Is(err, os.ErrNotExist) {
		return applyConflict("transaction recovery finalization path is occupied")
	}
	moveErr := publishNoReplace(path, tombstone)
	if tx.deps.afterJournalMove != nil {
		tx.deps.afterJournalMove()
	}
	landedData, landedReadErr := readBoundedInspectionFile(tombstone, maxProjectTransactionJournalSize)
	landedInfo, landedErr := os.Lstat(tombstone)
	_, activeErr := os.Lstat(path)
	landedOwned := landedReadErr == nil && landedErr == nil && tx.journalInfo != nil &&
		os.SameFile(tx.journalInfo, landedInfo) && landedInfo.Mode()&os.ModeSymlink == 0 &&
		landedInfo.Mode().IsRegular() && validApplyPrivateFileMode(landedInfo.Mode()) && bytes.Equal(landedData, tx.journalData)
	activeAbsent := errors.Is(activeErr, os.ErrNotExist)
	if !landedOwned || !activeAbsent {
		if landedErr == nil && activeAbsent {
			// The pathname was swapped after validation. Return the unknown file
			// to the active location when possible and retain the journal.
			_ = publishNoReplace(tombstone, path)
		}
		return applyConflict("transaction recovery evidence changed during finalization")
	}
	// Moving the owned journal away from its active pathname is the transaction
	// commit point. The private tombstone is retained so finalization never
	// unlinks a byte-identical external replacement by pathname.
	tx.journalCleared = true
	tx.journal = nil
	tx.journalData = nil
	tx.journalInfo = nil
	if moveErr != nil {
		return applyUnavailable("transaction recovery evidence finalization was not confirmed")
	}
	if err := tx.deps.SyncDir(tx.layout.DerivedDirectoryPath); err != nil {
		return applyUnavailable("transaction recovery directory could not be synced")
	}
	if err := tx.deps.SyncDir(filepath.Dir(tombstone)); err != nil {
		return applyUnavailable("transaction recovery finalization directory could not be synced")
	}
	return nil
}

func (tx *applyTransaction) recoverInterruptedTransaction() (bool, error) {
	journal, exists, err := tx.loadTransactionJournal()
	if err != nil || !exists {
		return false, err
	}
	if journal.Kind == projectTransactionKindRestore {
		return true, tx.recoverInterruptedRestore(journal)
	}
	current, err := tx.captureStatePreimage()
	if err != nil {
		return true, applyConflict("interrupted provenance state is unavailable")
	}
	if err := tx.loadRecoveryQuarantine(journal); err != nil {
		return true, err
	}
	preStateMatches := journalStateMatches(current, journal.PreState)
	candidateMatches := journalCandidateStateMatches(current, journal.CandidateState)
	if !preStateMatches && !candidateMatches {
		if tx.quarantine != nil {
			_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
		}
		return true, applyConflict("interrupted provenance state is ambiguous")
	}
	for index := len(journal.Entries) - 1; index >= 0; index-- {
		if err := tx.recoverApplyEntry(journal, journal.Entries[index]); err != nil {
			if tx.quarantine != nil {
				_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
			}
			return true, err
		}
	}
	if err := tx.cleanupInterruptedManagedRoots(journal); err != nil {
		return true, err
	}
	if candidateMatches {
		if err := tx.restoreJournalPreState(journal); err != nil {
			if tx.quarantine != nil {
				_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
			}
			return true, err
		}
	}
	if tx.quarantine != nil {
		manifest := tx.quarantine.manifest
		manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
		for index := range manifest.Entries {
			manifest.Entries[index].Status = ProjectQuarantineEntryRestored
		}
		manifest.Status = ProjectQuarantineRolledBack
		if err := tx.writeQuarantineManifest(tx.quarantine, manifest); err != nil {
			return true, err
		}
		tx.quarantine.manifest = manifest
	}
	if err := tx.clearTransactionJournal(); err != nil {
		return true, err
	}
	recoveryRun := projectRecoveryRunPath(tx.layout, journal.ID)
	if err := removeOwnedRecoveryRun(tx.layout, recoveryRun); err != nil {
		return true, err
	}
	return true, nil
}

func (tx *applyTransaction) recoverInterruptedRestore(journal projectTransactionJournal) error {
	current, err := tx.captureStatePreimage()
	if err != nil {
		return applyConflict("interrupted provenance state is unavailable")
	}
	if err := tx.loadRecoveryQuarantine(journal); err != nil {
		return err
	}
	preStateMatches := journalStateMatches(current, journal.PreState)
	candidateMatches := journalCandidateStateMatches(current, journal.CandidateState)
	if !preStateMatches && !candidateMatches {
		if tx.quarantine != nil {
			_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
		}
		return applyConflict("interrupted provenance state is ambiguous")
	}
	for index := len(journal.Entries) - 1; index >= 0; index-- {
		entry := journal.Entries[index]
		root, pathErr := tx.layout.ManagedSkillsPath(entry.Target)
		if pathErr != nil {
			return applyConflict("interrupted restore placement path is invalid")
		}
		destination := filepath.Join(root, entry.Skill)
		quarantined := filepath.Join(tx.layout.QuarantinePath, journal.QuarantineID, filepath.FromSlash(projectQuarantinedPlacement(entry.Target, entry.Skill)))
		oldHash := TreeHash{Algorithm: entry.TreeHashAlgorithm, Digest: entry.OldTreeHash}
		destinationHash, destinationExists, destinationErr := inspectRecoveryTree(tx.root, destination)
		quarantineHash, quarantineExists, quarantineErr := inspectRecoveryTree(tx.layout.QuarantinePath, quarantined)
		if destinationErr != nil || quarantineErr != nil {
			return firstRecoveryError(destinationErr, quarantineErr)
		}
		if quarantineExists {
			if quarantineHash != oldHash || destinationExists {
				_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
				return applyConflict("interrupted restore placement is ambiguous")
			}
			continue
		}
		if !destinationExists || destinationHash != oldHash {
			_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
			return applyConflict("interrupted restored placement changed")
		}
		if err := ensureRecoveryParent(tx, filepath.Dir(quarantined)); err != nil {
			return err
		}
		if err := moveRecoveryTree(tx, destination, quarantined, oldHash); err != nil {
			_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
			return err
		}
	}
	if err := tx.cleanupInterruptedManagedRoots(journal); err != nil {
		return err
	}
	if candidateMatches {
		if err := tx.restoreJournalPreState(journal); err != nil {
			_ = tx.setQuarantineStatus(ProjectQuarantineRecoveryRequired)
			return err
		}
	}
	preManifest, valid := DecodeProjectQuarantineManifest(journal.PreManifest)
	if !valid {
		return applyConflict("restore recovery manifest is malformed")
	}
	if err := tx.writeQuarantineManifest(tx.quarantine, preManifest); err != nil {
		return err
	}
	tx.quarantine.manifest = preManifest
	return tx.clearTransactionJournal()
}

func firstRecoveryError(left, right error) error {
	if left != nil {
		return left
	}
	return right
}

func (tx *applyTransaction) cleanupInterruptedManagedRoots(journal projectTransactionJournal) error {
	seen := make(map[Target]struct{})
	for _, entry := range journal.Entries {
		if entry.ManagedRootExisted {
			continue
		}
		if _, duplicate := seen[entry.Target]; duplicate {
			continue
		}
		seen[entry.Target] = struct{}{}
		root, err := tx.layout.ManagedSkillsPath(entry.Target)
		if err != nil {
			return applyConflict("interrupted managed root path is invalid")
		}
		if err := removeInterruptedEmptyDirectory(root); err != nil {
			return applyUnavailable("interrupted managed root could not be cleaned up")
		}
		parent := filepath.Dir(root)
		if !entry.ManagedParentExisted {
			if err := removeInterruptedEmptyDirectory(parent); err != nil {
				return applyUnavailable("interrupted managed ancestor could not be cleaned up")
			}
		}
		if err := tx.deps.SyncDir(filepath.Dir(parent)); err != nil {
			return applyUnavailable("interrupted managed ancestor directory could not be synced")
		}
	}
	return nil
}

func removeInterruptedEmptyDirectory(path string) error {
	if err := os.Remove(path); err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	entries, readErr := os.ReadDir(path)
	if readErr == nil && len(entries) > 0 {
		return nil
	}
	return applyUnavailable("interrupted directory could not be removed")
}

func journalStateMatches(current *applyStatePreimage, expected projectJournalState) bool {
	if current == nil || current.exists != expected.Exists {
		return false
	}
	return !expected.Exists || current.mode.Perm() == os.FileMode(expected.Mode).Perm() && bytes.Equal(current.data, expected.Data)
}

func journalCandidateStateMatches(current *applyStatePreimage, data []byte) bool {
	return current != nil && current.exists && validApplyPrivateFileMode(current.mode) && bytes.Equal(current.data, data)
}

func (tx *applyTransaction) loadRecoveryQuarantine(journal projectTransactionJournal) error {
	if journal.QuarantineID == "" {
		return nil
	}
	runPath := filepath.Join(tx.layout.QuarantinePath, journal.QuarantineID)
	manifestPath := filepath.Join(runPath, applyManifestName)
	if !pathWithin(tx.layout.QuarantinePath, runPath) || !pathWithin(runPath, manifestPath) {
		return applyConflict("interrupted quarantine boundary is invalid")
	}
	rootInfo, rootErr := os.Lstat(tx.layout.QuarantinePath)
	runInfo, runErr := os.Lstat(runPath)
	if errors.Is(runErr, os.ErrNotExist) {
		if journal.Kind == projectTransactionKindRestore {
			return applyConflict("interrupted restore quarantine disappeared")
		}
		// The journal is committed before the prepared quarantine is published.
		// Exact pre-state placements below prove that no managed mutation began.
		if errors.Is(rootErr, os.ErrNotExist) {
			return nil
		}
		if rootErr != nil || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
			return applyConflict("interrupted quarantine boundary is unsafe")
		}
		return nil
	}
	probe := proveInspectionPath(tx.root, manifestPath)
	if rootErr != nil || runErr != nil || !rootInfo.IsDir() || !runInfo.IsDir() ||
		rootInfo.Mode()&os.ModeSymlink != 0 || runInfo.Mode()&os.ModeSymlink != 0 || probe.unsafe {
		return applyConflict("interrupted quarantine is unavailable")
	}
	if !probe.exists {
		if journal.Kind == projectTransactionKindRestore {
			return applyConflict("interrupted restore quarantine manifest disappeared")
		}
		return tx.removeEmptyInterruptedQuarantineRun(journal.ID, rootInfo, runInfo, runPath)
	}
	if probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		return applyConflict("interrupted quarantine manifest is unsafe")
	}
	data, err := readBoundedInspectionFile(manifestPath, maxProjectQuarantineManifestSize)
	if err != nil {
		return applyConflict("interrupted quarantine manifest is unavailable")
	}
	manifest, valid := DecodeProjectQuarantineManifest(data)
	if !valid || manifest.ID != journal.QuarantineID || !journalMatchesQuarantine(journal, manifest) {
		return applyConflict("interrupted quarantine manifest is incompatible")
	}
	tx.quarantine = &projectQuarantineTransaction{
		runPath: runPath, runDurable: true, rootInfo: rootInfo, runInfo: runInfo,
		ancestors: make(map[string]os.FileInfo), manifestPath: manifestPath,
		manifestInfo: probe.info, manifestData: append([]byte(nil), data...), manifest: manifest,
	}
	return tx.checkQuarantineBoundary()
}

func (tx *applyTransaction) removeEmptyInterruptedQuarantineRun(id string, rootInfo, runInfo os.FileInfo, runPath string) error {
	entries, err := os.ReadDir(runPath)
	if err != nil || len(entries) != 0 {
		return applyConflict("interrupted quarantine preparation is ambiguous")
	}
	if err := tx.preserveInterruptedManifestStaging(id); err != nil {
		return err
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	currentRoot, rootErr := os.Lstat(tx.layout.QuarantinePath)
	currentRun, runErr := os.Lstat(runPath)
	if rootErr != nil || runErr != nil || !os.SameFile(rootInfo, currentRoot) || !os.SameFile(runInfo, currentRun) ||
		currentRoot.Mode()&os.ModeSymlink != 0 || currentRun.Mode()&os.ModeSymlink != 0 || !currentRoot.IsDir() || !currentRun.IsDir() {
		return applyConflict("interrupted quarantine preparation changed")
	}
	if err := os.Remove(runPath); err != nil {
		return applyConflict("interrupted quarantine preparation could not be removed")
	}
	if err := tx.deps.SyncDir(tx.layout.QuarantinePath); err != nil {
		return applyUnavailable("interrupted quarantine root could not be synced")
	}
	return nil
}

func (tx *applyTransaction) preserveInterruptedManifestStaging(id string) error {
	stagedPath := filepath.Join(tx.layout.DerivedDirectoryPath, ".quarantine-manifest-"+id+"-staged")
	probe := proveInspectionPath(tx.root, stagedPath)
	if probe.unsafe {
		return applyConflict("interrupted manifest staging is unsafe")
	}
	if !probe.exists {
		return nil
	}
	if probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.Mode().IsRegular() {
		return applyConflict("interrupted manifest staging is not a regular file")
	}
	recoveryPath := filepath.Join(projectRecoveryRunPath(tx.layout, id), "prepared-manifest.partial")
	if err := ensureRecoveryParent(tx, filepath.Dir(recoveryPath)); err != nil {
		return err
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	current, statErr := os.Lstat(stagedPath)
	if statErr != nil || !os.SameFile(probe.info, current) || current.Mode()&os.ModeSymlink != 0 || !current.Mode().IsRegular() {
		return applyConflict("interrupted manifest staging changed")
	}
	if err := publishNoReplace(stagedPath, recoveryPath); err != nil {
		return applyUnavailable("interrupted manifest staging could not be preserved")
	}
	landed, landedErr := os.Lstat(recoveryPath)
	if landedErr != nil || !os.SameFile(probe.info, landed) {
		return applyUnavailable("interrupted manifest staging could not be verified")
	}
	if err := tx.deps.SyncDir(tx.layout.DerivedDirectoryPath); err != nil {
		return applyUnavailable("interrupted manifest staging source could not be synced")
	}
	if err := tx.deps.SyncDir(filepath.Dir(recoveryPath)); err != nil {
		return applyUnavailable("interrupted manifest staging recovery could not be synced")
	}
	return nil
}

func journalMatchesQuarantine(journal projectTransactionJournal, manifest ProjectQuarantineManifest) bool {
	expected := make(map[string]projectJournalEntry)
	for _, entry := range journal.Entries {
		if entry.Action != PlanActionInstall {
			expected[projectPlacementKey(entry.Target, entry.Skill)] = entry
		}
	}
	if len(expected) != len(manifest.Entries) {
		return false
	}
	for _, entry := range manifest.Entries {
		journalEntry, ok := expected[projectPlacementKey(entry.Target, entry.Skill)]
		if !ok || journalEntry.TreeHashAlgorithm != entry.TreeHashAlgorithm || journalEntry.OldTreeHash != entry.OldTreeHash ||
			journalEntry.OldSourceIdentity != entry.OldSourceIdentity {
			return false
		}
		if journalEntry.Action == PlanActionUpdate {
			if entry.Action != ProjectQuarantineEntryActionUpdate || journalEntry.NewTreeHash != entry.NewTreeHash ||
				journalEntry.NewSourceIdentity != entry.NewSourceIdentity {
				return false
			}
		} else if journalEntry.Action != PlanActionQuarantine || entry.Action != ProjectQuarantineEntryActionRemove {
			return false
		}
	}
	return true
}

func (tx *applyTransaction) recoverApplyEntry(journal projectTransactionJournal, entry projectJournalEntry) error {
	root, err := tx.layout.ManagedSkillsPath(entry.Target)
	if err != nil {
		return applyConflict("interrupted placement path is invalid")
	}
	destination := filepath.Join(root, entry.Skill)
	if filepath.Clean(destination) != filepath.Join(filepath.Clean(root), entry.Skill) || !pathWithin(tx.root, destination) {
		return applyConflict("interrupted placement path is invalid")
	}
	recoveryPath := filepath.Join(projectRecoveryRunPath(tx.layout, journal.ID), string(entry.Target), entry.Skill)
	if !pathWithin(tx.layout.DerivedDirectoryPath, recoveryPath) {
		return applyConflict("interrupted recovery path is invalid")
	}
	if entry.Action != PlanActionQuarantine {
		newHash := TreeHash{Algorithm: entry.TreeHashAlgorithm, Digest: entry.NewTreeHash}
		destinationHash, destinationExists, inspectErr := inspectRecoveryTree(tx.root, destination)
		if inspectErr != nil {
			return inspectErr
		}
		recoveryHash, recoveryExists, recoveryErr := inspectRecoveryTree(tx.layout.DerivedDirectoryPath, recoveryPath)
		if recoveryErr != nil {
			return recoveryErr
		}
		if destinationExists && destinationHash == newHash {
			if recoveryExists {
				return applyConflict("interrupted recovery destination is occupied")
			}
			if err := ensureRecoveryParent(tx, filepath.Dir(recoveryPath)); err != nil {
				return err
			}
			if err := moveRecoveryTree(tx, destination, recoveryPath, newHash); err != nil {
				return err
			}
			recoveryExists = true
		} else if destinationExists && entry.Action == PlanActionInstall {
			return applyConflict("interrupted installed placement changed")
		} else if recoveryExists && recoveryHash != newHash {
			return applyConflict("interrupted recovery content changed")
		}
	}
	if entry.Action == PlanActionInstall {
		return nil
	}
	oldHash := TreeHash{Algorithm: entry.TreeHashAlgorithm, Digest: entry.OldTreeHash}
	quarantined := filepath.Join(tx.layout.QuarantinePath, journal.QuarantineID, filepath.FromSlash(projectQuarantinedPlacement(entry.Target, entry.Skill)))
	destinationHash, destinationExists, inspectErr := inspectRecoveryTree(tx.root, destination)
	if inspectErr != nil {
		return inspectErr
	}
	quarantineHash, quarantineExists, quarantineErr := inspectRecoveryTree(tx.layout.QuarantinePath, quarantined)
	if quarantineErr != nil {
		return quarantineErr
	}
	if destinationExists {
		if destinationHash == oldHash && !quarantineExists {
			return nil
		}
		return applyConflict("interrupted managed placement is ambiguous")
	}
	if !quarantineExists || quarantineHash != oldHash {
		return applyConflict("interrupted quarantined placement changed")
	}
	if err := ensureRecoveryManagedAncestors(tx, entry.Target); err != nil {
		return err
	}
	return moveRecoveryTree(tx, quarantined, destination, oldHash)
}

func inspectRecoveryTree(boundary, path string) (TreeHash, bool, error) {
	hash, _, exists, err := inspectRecoveryTreeIdentity(boundary, path, nil)
	return hash, exists, err
}

func inspectRecoveryTreeIdentity(boundary, path string, afterHash func()) (TreeHash, os.FileInfo, bool, error) {
	if !pathWithin(boundary, path) {
		return TreeHash{}, nil, false, applyConflict("interrupted recovery path escapes its boundary")
	}
	probe := proveInspectionPath(boundary, path)
	if probe.unsafe {
		return TreeHash{}, nil, false, applyConflict("interrupted recovery path is unsafe")
	}
	if !probe.exists {
		return TreeHash{}, nil, false, nil
	}
	if probe.info == nil || probe.info.Mode()&os.ModeSymlink != 0 || !probe.info.IsDir() {
		return TreeHash{}, probe.info, true, applyConflict("interrupted recovery entry is not a real directory")
	}
	hash, err := HashSkillTree(path)
	if err != nil {
		return TreeHash{}, probe.info, true, applyConflict("interrupted recovery content could not be verified")
	}
	if afterHash != nil {
		afterHash()
	}
	current, err := os.Lstat(path)
	if err != nil || !os.SameFile(probe.info, current) || current.Mode()&os.ModeSymlink != 0 || !current.IsDir() {
		return TreeHash{}, probe.info, true, applyConflict("interrupted recovery entry changed during verification")
	}
	return hash, probe.info, true, nil
}

func ensureRecoveryParent(tx *applyTransaction, parent string) error {
	created, err := ensureApplyDirectory(tx.root, parent)
	if err != nil {
		return err
	}
	return tx.syncCreatedDirectoryParents(created)
}

func ensureRecoveryManagedAncestors(tx *applyTransaction, target Target) error {
	root, err := tx.layout.ManagedSkillsPath(target)
	if err != nil {
		return applyConflict("interrupted managed placement path is invalid")
	}
	created, err := ensureApplyDirectory(tx.root, root)
	if err != nil {
		return err
	}
	return tx.syncCreatedDirectoryParents(created)
}

func moveRecoveryTree(tx *applyTransaction, source, destination string, expected TreeHash) error {
	if err := tx.checkLock(); err != nil {
		return err
	}
	before, sourceInfo, exists, err := inspectRecoveryTreeIdentity(tx.root, source, tx.deps.afterRecoveryTreeHash)
	if err != nil || !exists || before != expected {
		return applyConflict("interrupted recovery source changed before move")
	}
	if _, exists, err := inspectRecoveryTree(tx.root, destination); err != nil || exists {
		if err != nil {
			return err
		}
		return applyConflict("interrupted recovery destination appeared")
	}
	if tx.deps.beforeRecoveryTreeMove != nil {
		tx.deps.beforeRecoveryTreeMove()
	}
	if err := publishNoReplace(source, destination); err != nil {
		landed, landedExists, _ := inspectRecoveryTree(tx.root, destination)
		landedInfo, landedStatErr := os.Lstat(destination)
		_, sourceExists, _ := inspectRecoveryTree(tx.root, source)
		if landedExists && landed == expected && !sourceExists {
			if landedStatErr == nil && os.SameFile(sourceInfo, landedInfo) {
				return applyUnavailable("interrupted recovery move was not confirmed")
			}
			_ = publishNoReplace(destination, source)
			return applyConflict("interrupted recovery tree changed during move")
		}
		if errors.Is(err, os.ErrExist) {
			return applyConflict("interrupted recovery destination appeared")
		}
		return applyUnavailable("interrupted recovery move could not be completed")
	}
	landed, landedExists, err := inspectRecoveryTree(tx.root, destination)
	landedInfo, landedStatErr := os.Lstat(destination)
	_, sourceExists, sourceErr := inspectRecoveryTree(tx.root, source)
	if err != nil || sourceErr != nil || landedStatErr != nil || !landedExists || landed != expected || sourceExists || !os.SameFile(sourceInfo, landedInfo) {
		if landedExists && !sourceExists {
			// The moved tree was not ours. Return it to the public source path
			// when that path remains absent; otherwise retain both copies.
			_ = publishNoReplace(destination, source)
		}
		return applyConflict("interrupted recovery tree changed during move")
	}
	if err := tx.deps.SyncDir(filepath.Dir(source)); err != nil {
		return applyUnavailable("interrupted recovery source directory could not be synced")
	}
	if err := tx.deps.SyncDir(filepath.Dir(destination)); err != nil {
		return applyUnavailable("interrupted recovery destination directory could not be synced")
	}
	return nil
}

func (tx *applyTransaction) restoreJournalPreState(journal projectTransactionJournal) error {
	current, err := tx.captureStatePreimage()
	if err != nil || !journalCandidateStateMatches(current, journal.CandidateState) {
		return applyConflict("interrupted provenance state changed before recovery")
	}
	if journal.PreState.Exists {
		commit, err := tx.writeStateBytes(current, journal.PreState.Data, os.FileMode(journal.PreState.Mode))
		if err != nil || commit == nil || !commit.replaced {
			return applyUnavailable("interrupted provenance state could not be restored")
		}
		return nil
	}
	return tx.moveOwnedStateToRecovery(current, "interrupted provenance state")
}

func (tx *applyTransaction) moveOwnedStateToRecovery(expected *applyStatePreimage, description string) error {
	if expected == nil || !expected.exists || expected.info == nil || tx.journal == nil {
		return applyUnavailable(description + " ownership is unavailable")
	}
	if err := tx.checkLock(); err != nil {
		return err
	}
	current, err := tx.captureStatePreimage()
	if err != nil || !current.exists || !bytes.Equal(current.data, expected.data) || !os.SameFile(current.info, expected.info) {
		return applyConflict(description + " changed before removal")
	}
	recoveryPath := filepath.Join(projectRecoveryRunPath(tx.layout, tx.journal.ID), "provenance.json")
	if !pathWithin(tx.layout.DerivedDirectoryPath, recoveryPath) {
		return applyConflict(description + " recovery path is invalid")
	}
	if err := ensureRecoveryParent(tx, filepath.Dir(recoveryPath)); err != nil {
		return err
	}
	if _, err := os.Lstat(recoveryPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		return applyConflict(description + " recovery destination appeared")
	}
	if tx.deps.beforeStateRecoveryMove != nil {
		tx.deps.beforeStateRecoveryMove()
	}
	if err := publishNoReplace(tx.layout.ReconcilerStatePath, recoveryPath); err != nil {
		return applyUnavailable(description + " could not be moved to recovery")
	}
	stored, readErr := readBoundedInspectionFile(recoveryPath, maxProvenanceStateBytes)
	info, statErr := os.Lstat(recoveryPath)
	if readErr != nil || statErr != nil || !bytes.Equal(stored, expected.data) || !os.SameFile(info, expected.info) {
		// A raced replacement is not ours to retain. Restore it to the public
		// path when that path is still absent; otherwise keep the private copy.
		if _, sourceErr := os.Lstat(tx.layout.ReconcilerStatePath); errors.Is(sourceErr, os.ErrNotExist) {
			_ = publishNoReplace(recoveryPath, tx.layout.ReconcilerStatePath)
		}
		return applyConflict(description + " changed during recovery move")
	}
	if err := tx.deps.SyncDir(filepath.Dir(tx.layout.ReconcilerStatePath)); err != nil {
		return applyUnavailable(description + " source directory could not be synced")
	}
	if err := tx.deps.SyncDir(filepath.Dir(recoveryPath)); err != nil {
		return applyUnavailable(description + " recovery directory could not be synced")
	}
	return nil
}

func removeOwnedRecoveryRun(layout DerivedLayout, runPath string) error {
	if !pathWithin(filepath.Join(layout.DerivedDirectoryPath, projectRecoveryDirectoryName), runPath) {
		return applyConflict("interrupted recovery cleanup path is invalid")
	}
	if _, err := os.Lstat(runPath); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return applyUnavailable("interrupted recovery cleanup is unavailable")
	}
	// Published snapshots are retained privately after rollback. Removing a
	// non-empty tree recursively would reintroduce a swap-and-delete race.
	// Empty scaffolding can be removed with the non-recursive operation.
	_ = os.Remove(runPath)
	_ = os.Remove(filepath.Dir(runPath))
	return nil
}
