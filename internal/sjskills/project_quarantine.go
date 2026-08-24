package sjskills

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"sort"
	"time"
)

const (
	ProjectQuarantineManifestVersion = 1
	maxProjectQuarantineManifestSize = 1 << 20
	maxProjectQuarantineEntries      = 2048
)

type ProjectQuarantineStatus string

const (
	ProjectQuarantinePrepared         ProjectQuarantineStatus = "prepared"
	ProjectQuarantineActive           ProjectQuarantineStatus = "active"
	ProjectQuarantineCommitted        ProjectQuarantineStatus = "committed"
	ProjectQuarantineRolledBack       ProjectQuarantineStatus = "rolled-back"
	ProjectQuarantineRecoveryRequired ProjectQuarantineStatus = "recovery-required"
)

type ProjectQuarantineEntryStatus string

const (
	ProjectQuarantineEntryPending          ProjectQuarantineEntryStatus = "pending"
	ProjectQuarantineEntryQuarantined      ProjectQuarantineEntryStatus = "quarantined"
	ProjectQuarantineEntryReplaced         ProjectQuarantineEntryStatus = "replaced"
	ProjectQuarantineEntryRestored         ProjectQuarantineEntryStatus = "restored"
	ProjectQuarantineEntryRecoveryRequired ProjectQuarantineEntryStatus = "recovery-required"
)

// ProjectQuarantineManifest is durable recovery evidence for one project
// update transaction. Placement identifiers are relative modeled identities,
// never process-private absolute filesystem paths.
type ProjectQuarantineManifest struct {
	Version   int                              `json:"version"`
	ID        string                           `json:"id"`
	CreatedAt time.Time                        `json:"createdAt"`
	Status    ProjectQuarantineStatus          `json:"status"`
	Entries   []ProjectQuarantineManifestEntry `json:"entries"`
}

type ProjectQuarantineManifestEntry struct {
	Skill                string                       `json:"skill"`
	Target               Target                       `json:"target"`
	OriginalPlacement    string                       `json:"originalPlacement"`
	QuarantinedPlacement string                       `json:"quarantinedPlacement"`
	OldSourceIdentity    string                       `json:"oldSourceIdentity"`
	NewSourceIdentity    string                       `json:"newSourceIdentity"`
	TreeHashAlgorithm    string                       `json:"treeHashAlgorithm"`
	OldTreeHash          string                       `json:"oldTreeHash"`
	NewTreeHash          string                       `json:"newTreeHash"`
	Status               ProjectQuarantineEntryStatus `json:"status"`
}

var projectQuarantineIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func newProjectQuarantineID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func projectOriginalPlacement(target Target, skill string) string {
	return string(target) + "/skills/" + skill
}

func projectQuarantinedPlacement(target Target, skill string) string {
	return "entries/" + string(target) + "/" + skill
}

// DecodeProjectQuarantineManifest applies the same strict bounded contract
// that a later overwrite-refusing restore command will consume.
func DecodeProjectQuarantineManifest(data []byte) (ProjectQuarantineManifest, bool) {
	if len(data) == 0 || len(data) > maxProjectQuarantineManifestSize {
		return ProjectQuarantineManifest{}, false
	}
	var manifest ProjectQuarantineManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return ProjectQuarantineManifest{}, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ProjectQuarantineManifest{}, false
	}
	if !validProjectQuarantineManifest(manifest) {
		return ProjectQuarantineManifest{}, false
	}
	manifest.Entries = append([]ProjectQuarantineManifestEntry(nil), manifest.Entries...)
	return manifest, true
}

func validProjectQuarantineManifest(manifest ProjectQuarantineManifest) bool {
	if manifest.Version != ProjectQuarantineManifestVersion ||
		!projectQuarantineIDPattern.MatchString(manifest.ID) ||
		manifest.CreatedAt.IsZero() || manifest.CreatedAt.Location() != time.UTC ||
		!validProjectQuarantineStatus(manifest.Status) ||
		manifest.Entries == nil || len(manifest.Entries) == 0 || len(manifest.Entries) > maxProjectQuarantineEntries {
		return false
	}
	seen := make(map[string]struct{}, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if !isPortableName(entry.Skill) ||
			(entry.Target != TargetAgents && entry.Target != TargetClaude) ||
			entry.OriginalPlacement != projectOriginalPlacement(entry.Target, entry.Skill) ||
			entry.QuarantinedPlacement != projectQuarantinedPlacement(entry.Target, entry.Skill) ||
			!isCanonicalProjectSourceIdentity(entry.OldSourceIdentity) || entry.NewSourceIdentity != entry.OldSourceIdentity ||
			entry.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 ||
			!lowercaseDigestPattern.MatchString(entry.OldTreeHash) ||
			!lowercaseDigestPattern.MatchString(entry.NewTreeHash) ||
			entry.NewTreeHash == entry.OldTreeHash ||
			!validProjectQuarantineEntryStatus(entry.Status) {
			return false
		}
		key := projectPlacementKey(entry.Target, entry.Skill)
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
	}
	for _, entry := range manifest.Entries {
		switch manifest.Status {
		case ProjectQuarantinePrepared:
			if entry.Status != ProjectQuarantineEntryPending {
				return false
			}
		case ProjectQuarantineCommitted:
			if entry.Status != ProjectQuarantineEntryReplaced {
				return false
			}
		case ProjectQuarantineRolledBack:
			if entry.Status != ProjectQuarantineEntryPending && entry.Status != ProjectQuarantineEntryRestored {
				return false
			}
		}
	}
	return sort.SliceIsSorted(manifest.Entries, func(i, j int) bool {
		if projectTargetRank(manifest.Entries[i].Target) != projectTargetRank(manifest.Entries[j].Target) {
			return projectTargetRank(manifest.Entries[i].Target) < projectTargetRank(manifest.Entries[j].Target)
		}
		return compareUTF16(manifest.Entries[i].Skill, manifest.Entries[j].Skill) < 0
	})
}

func validProjectQuarantineStatus(status ProjectQuarantineStatus) bool {
	switch status {
	case ProjectQuarantinePrepared, ProjectQuarantineActive, ProjectQuarantineCommitted,
		ProjectQuarantineRolledBack, ProjectQuarantineRecoveryRequired:
		return true
	default:
		return false
	}
}

func validProjectQuarantineEntryStatus(status ProjectQuarantineEntryStatus) bool {
	switch status {
	case ProjectQuarantineEntryPending, ProjectQuarantineEntryQuarantined,
		ProjectQuarantineEntryReplaced, ProjectQuarantineEntryRestored,
		ProjectQuarantineEntryRecoveryRequired:
		return true
	default:
		return false
	}
}

func marshalProjectQuarantineManifest(manifest ProjectQuarantineManifest) ([]byte, error) {
	if !validProjectQuarantineManifest(manifest) {
		return nil, applyConflict("quarantine manifest identity is invalid")
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if len(data) > maxProjectQuarantineManifestSize {
		return nil, applyUnavailable("quarantine manifest exceeds its size bound")
	}
	return data, nil
}
