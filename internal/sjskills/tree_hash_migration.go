package sjskills

import "fmt"

// Provenance written before tree-sha256-v3 records tree-sha256-v2 digests,
// which hash each file's executable bit. A v2 record still proves ownership of
// the exact tree it describes, so inventory recognizes it by recomputing the
// v2 digest and, when that matches, treats the record as the v3 digest of the
// same walk. The two digests come from one walk, and a v2 digest identifies
// one tree, so the rewrite neither widens nor narrows what the record proves.
//
// Plans classify with the rewritten records and report how many there are;
// apply persists them before any other state change (persistTreeHashUpgrade),
// so every later transaction, journal, and quarantine sees only v3 records
// apart from v2 records that no longer match their placement. Those stay as
// they are and verify nothing, exactly as a changed tree would.

const treeHashMigrationEvidenceKind = "tree-hash-migration"

// upgradeLegacyTreeHashRecords returns a copy of records in which every
// tree-sha256-v2 record whose placement still holds the recorded tree is
// rewritten as that tree's tree-sha256-v3 digest, plus the rewrite count.
// Only hashed directory entries of safe roots are considered, and an entry's
// current digest must equal the one inventory already recorded for it.
func upgradeLegacyTreeHashRecords(records []ProvenanceRecord, roots []ProjectSkillsRootInventory) ([]ProvenanceRecord, int) {
	result := cloneProvenanceRecords(records)
	upgraded := 0
	for index, record := range result {
		if record.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 {
			continue
		}
		entry, ok := hashedInventoryEntry(roots, record.Target, record.Skill)
		if !ok {
			continue
		}
		digests, err := hashSkillTreeDigests(entry.Path, MaterializerLimits{})
		if err != nil || digests.legacy.Digest != record.TreeHash || digests.current != *entry.Hash {
			continue
		}
		result[index].TreeHashAlgorithm = digests.current.Algorithm
		result[index].TreeHash = digests.current.Digest
		upgraded++
	}
	return result, upgraded
}

// persistPlannedTreeHashUpgrade persists the upgrade only when the verified
// plan reported it, so apply never writes state its plan did not show.
func (tx *applyTransaction) persistPlannedTreeHashUpgrade(plan Plan) (bool, error) {
	if !planHasEvidence(plan, treeHashMigrationEvidenceKind) {
		return false, nil
	}
	return tx.persistTreeHashUpgrade()
}

func hasLegacyTreeHashRecord(records []ProvenanceRecord) bool {
	for _, record := range records {
		if record.TreeHashAlgorithm == TreeHashAlgorithmSHA256V2 {
			return true
		}
	}
	return false
}

func hashedInventoryEntry(roots []ProjectSkillsRootInventory, target Target, skill string) (InventoryEntry, bool) {
	for _, root := range roots {
		if root.Target != target || !root.Safe {
			continue
		}
		for _, entry := range root.Entries {
			if entry.Name == skill && entry.Kind == InventoryEntryDirectory && entry.Problem == "" && entry.Hash != nil {
				return entry, true
			}
		}
	}
	return InventoryEntry{}, false
}

func treeHashMigrationEvidence(count int) Evidence {
	return Evidence{
		Kind:   treeHashMigrationEvidenceKind,
		Detail: fmt.Sprintf("%d trusted provenance records verified with tree-sha256-v2 are rewritten as tree-sha256-v3", count),
	}
}

// persistTreeHashUpgrade writes the records the plan already classified as
// tree-sha256-v3. It runs under the apply lock after the reviewed plan was
// re-verified and before any placement or transaction journal changes. The
// rewritten state describes the same trees, so a crash on either side of
// the atomic replacement leaves valid, equivalent provenance and needs no
// journal. Legacy version 1 global state decodes into the current format, so
// writing it here also performs that format migration; the reviewed
// state-only migration transaction that follows then rewrites equal records.
func (tx *applyTransaction) persistTreeHashUpgrade() (bool, error) {
	preimage, err := tx.captureStatePreimage()
	if err != nil {
		return false, err
	}
	if !preimage.exists || !hasLegacyTreeHashRecord(preimage.state.Records) {
		return false, nil
	}
	var roots []ProjectSkillsRootInventory
	if tx.global != nil {
		inventory, inspectErr := InspectGlobal(tx.global.layout)
		if inspectErr != nil {
			return false, applyConflict("global boundary changed after planning")
		}
		roots = inventory.Roots
	} else {
		inventory, inspectErr := InspectProject(tx.layout)
		if inspectErr != nil {
			return false, applyConflict("project boundary changed after planning")
		}
		roots = inventory.Roots
	}
	records, upgraded := upgradeLegacyTreeHashRecords(preimage.state.Records, roots)
	if upgraded == 0 {
		return false, nil
	}
	state := preimage.state
	state.Records = records
	data, err := marshalApplyState(state)
	if err != nil {
		return false, applyUnavailable("provenance tree hash upgrade could not be encoded")
	}
	if _, err := tx.writeStateBytes(preimage, data, preimage.mode); err != nil {
		return false, err
	}
	return true, nil
}
