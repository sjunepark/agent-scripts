package sjskills

import (
	"sort"
	"strconv"
)

// Source identity migration lets a renamed or transferred source keep
// ownership of copies installed under its old repository identity. Only a
// registry-declared former identity of the placement's own desired source can
// migrate; the classifier still applies the ordinary tree-hash rules, and the
// next committed state records the desired source's current identity.
//
// A migrated record is treated as if it had always carried the current
// identity: quarantine manifests and recovery journal entries describe the
// mapped record, and the journal lists each mapping separately so recovery
// can verify the state transition without the registry.

// isFormerSourceIdentity reports whether identity is one of the desired copy's
// declared former identities. Direct, manual, and workflow entries never
// carry any.
func isFormerSourceIdentity(skill DesiredSkill, identity string) bool {
	if skill.Manager != ManagerSkillsCLI || identity == "" {
		return false
	}
	for _, former := range skill.FormerSourceIdentities {
		if former == identity {
			return true
		}
	}
	return false
}

type sourceIdentityMigration struct {
	Skill             string `json:"skill"`
	Target            Target `json:"target"`
	OldSourceIdentity string `json:"oldSourceIdentity"`
	NewSourceIdentity string `json:"newSourceIdentity"`
}

// migratedProvenanceRecord returns record under its desired placement's
// current identity when the record names a declared former identity.
func migratedProvenanceRecord(record ProvenanceRecord, desired map[string]DesiredSkill) (ProvenanceRecord, bool) {
	skill, wanted := desired[projectPlacementKey(record.Target, record.Skill)]
	if !wanted || skill.Mode != ModeCopy || !isFormerSourceIdentity(skill, record.SourceIdentity) {
		return record, false
	}
	current, ok := canonicalProjectSourceIdentity(skill.Source)
	if !ok || current == record.SourceIdentity {
		return record, false
	}
	record.SourceIdentity = current
	return record, true
}

// sourceIdentityMigrations lists, in placement order, every trusted record the
// next commit re-records under its desired source's current identity.
func sourceIdentityMigrations(state ProvenanceState, desired DesiredState) []sourceIdentityMigration {
	desiredByKey := desiredByPlacement(desired)
	result := make([]sourceIdentityMigration, 0)
	for _, record := range state.Records {
		if record.Scope != desired.Scope {
			continue
		}
		migrated, ok := migratedProvenanceRecord(record, desiredByKey)
		if !ok {
			continue
		}
		result = append(result, sourceIdentityMigration{Skill: record.Skill, Target: record.Target,
			OldSourceIdentity: record.SourceIdentity, NewSourceIdentity: migrated.SourceIdentity})
	}
	sortSourceIdentityMigrations(result)
	return result
}

func sortSourceIdentityMigrations(migrations []sourceIdentityMigration) {
	sort.SliceStable(migrations, func(i, j int) bool {
		if projectTargetRank(migrations[i].Target) != projectTargetRank(migrations[j].Target) {
			return projectTargetRank(migrations[i].Target) < projectTargetRank(migrations[j].Target)
		}
		return compareUTF16(migrations[i].Skill, migrations[j].Skill) < 0
	})
}

// effectiveProvenanceRecords maps each trusted record by placement after
// applying declared source identity migrations.
func effectiveProvenanceRecords(state ProvenanceState, desired DesiredState) map[string]ProvenanceRecord {
	desiredByKey := desiredByPlacement(desired)
	records := make(map[string]ProvenanceRecord, len(state.Records))
	for _, record := range state.Records {
		if record.Scope == desired.Scope {
			record, _ = migratedProvenanceRecord(record, desiredByKey)
		}
		records[projectPlacementKey(record.Target, record.Skill)] = record
	}
	return records
}

// validateDesiredFormerIdentities guards the pure classifier boundary, whose
// callers need not have come through registry validation.
func validateDesiredFormerIdentities(path string, skill DesiredSkill, issues *[]Issue) {
	if len(skill.FormerSourceIdentities) == 0 {
		return
	}
	current, ok := canonicalProjectSourceIdentity(skill.Source)
	if skill.Manager != ManagerSkillsCLI || skill.Origin == "direct" || !ok || IsLocalSource(skill.Source) {
		*issues = append(*issues, Issue{Code: IssueInvalidSource, Path: path + ".formerSourceIdentities", Message: "only a central remote skills-cli source may declare former identities"})
		return
	}
	seen := make(map[string]struct{}, len(skill.FormerSourceIdentities))
	for index, former := range skill.FormerSourceIdentities {
		entryPath := path + ".formerSourceIdentities[" + strconv.Itoa(index) + "]"
		if _, duplicate := seen[former]; duplicate {
			*issues = append(*issues, Issue{Code: IssueDuplicate, Path: entryPath, Message: "duplicate former source identity"})
			continue
		}
		seen[former] = struct{}{}
		if !isCanonicalSourceIdentityForScope(ScopeGlobal, former) || former == current {
			*issues = append(*issues, Issue{Code: IssueInvalidSource, Path: entryPath, Message: "must be a canonical remote identity other than the current source identity"})
		}
	}
}
