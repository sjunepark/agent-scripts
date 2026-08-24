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
	ProjectQuarantineRestoring        ProjectQuarantineStatus = "restoring"
	ProjectQuarantineRestored         ProjectQuarantineStatus = "restored"
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

// ProjectQuarantineEntryAction identifies whether a quarantined preimage is
// being replaced by a new desired tree or is being removed because it is no
// longer part of project intent.  The action is durable recovery evidence;
// entry status alone cannot distinguish a committed removal from an update.
type ProjectQuarantineEntryAction string

const (
	ProjectQuarantineEntryActionUpdate ProjectQuarantineEntryAction = "update"
	ProjectQuarantineEntryActionRemove ProjectQuarantineEntryAction = "remove"
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
	Action               ProjectQuarantineEntryAction `json:"action"`
	Skill                string                       `json:"skill"`
	Target               Target                       `json:"target"`
	OriginalPlacement    string                       `json:"originalPlacement"`
	QuarantinedPlacement string                       `json:"quarantinedPlacement"`
	OldSourceIdentity    string                       `json:"oldSourceIdentity"`
	NewSourceIdentity    string                       `json:"newSourceIdentity,omitempty"`
	TreeHashAlgorithm    string                       `json:"treeHashAlgorithm"`
	OldTreeHash          string                       `json:"oldTreeHash"`
	NewTreeHash          string                       `json:"newTreeHash,omitempty"`
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
	if !validProjectQuarantineJSONShape(data) {
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

func validProjectQuarantineJSONShape(data []byte) bool {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return false
	}
	entriesData, ok := object["entries"]
	if !ok {
		return true
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(entriesData, &entries); err != nil {
		return false
	}
	for _, entry := range entries {
		actionData, ok := entry["action"]
		if !ok {
			continue
		}
		var action ProjectQuarantineEntryAction
		if err := json.Unmarshal(actionData, &action); err != nil {
			return false
		}
		if action == ProjectQuarantineEntryActionRemove {
			if _, present := entry["newSourceIdentity"]; present {
				return false
			}
			if _, present := entry["newTreeHash"]; present {
				return false
			}
		}
	}
	return true
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
		if !validProjectQuarantineEntry(entry) {
			return false
		}
		key := projectPlacementKey(entry.Target, entry.Skill)
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
	}
	for _, entry := range manifest.Entries {
		if !validProjectQuarantineEntryState(manifest.Status, entry) {
			return false
		}
	}
	return sort.SliceIsSorted(manifest.Entries, func(i, j int) bool {
		if projectTargetRank(manifest.Entries[i].Target) != projectTargetRank(manifest.Entries[j].Target) {
			return projectTargetRank(manifest.Entries[i].Target) < projectTargetRank(manifest.Entries[j].Target)
		}
		return compareUTF16(manifest.Entries[i].Skill, manifest.Entries[j].Skill) < 0
	})
}

func validProjectQuarantineEntryState(manifestStatus ProjectQuarantineStatus, entry ProjectQuarantineManifestEntry) bool {
	switch manifestStatus {
	case ProjectQuarantinePrepared:
		return entry.Status == ProjectQuarantineEntryPending
	case ProjectQuarantineActive:
		if entry.Status == ProjectQuarantineEntryPending {
			return true
		}
		if entry.Action == ProjectQuarantineEntryActionUpdate {
			return entry.Status == ProjectQuarantineEntryQuarantined || entry.Status == ProjectQuarantineEntryReplaced
		}
		return entry.Status == ProjectQuarantineEntryQuarantined
	case ProjectQuarantineCommitted:
		if entry.Action == ProjectQuarantineEntryActionUpdate {
			return entry.Status == ProjectQuarantineEntryReplaced
		}
		return entry.Status == ProjectQuarantineEntryQuarantined
	case ProjectQuarantineRestoring:
		if entry.Status == ProjectQuarantineEntryRestored {
			return true
		}
		if entry.Action == ProjectQuarantineEntryActionUpdate {
			return entry.Status == ProjectQuarantineEntryReplaced
		}
		return entry.Status == ProjectQuarantineEntryQuarantined
	case ProjectQuarantineRestored:
		return entry.Status == ProjectQuarantineEntryRestored
	case ProjectQuarantineRolledBack:
		return entry.Status == ProjectQuarantineEntryPending || entry.Status == ProjectQuarantineEntryRestored
	case ProjectQuarantineRecoveryRequired:
		// Recovery may contain a mixture of untouched, restored, and ambiguous
		// entries. The entry action and bounded status vocabulary still make the
		// durable state unambiguous to a later restore operation.
		if entry.Action == ProjectQuarantineEntryActionUpdate {
			return entry.Status == ProjectQuarantineEntryPending ||
				entry.Status == ProjectQuarantineEntryQuarantined ||
				entry.Status == ProjectQuarantineEntryReplaced ||
				entry.Status == ProjectQuarantineEntryRestored ||
				entry.Status == ProjectQuarantineEntryRecoveryRequired
		}
		return entry.Status == ProjectQuarantineEntryPending ||
			entry.Status == ProjectQuarantineEntryQuarantined ||
			entry.Status == ProjectQuarantineEntryRestored ||
			entry.Status == ProjectQuarantineEntryRecoveryRequired
	default:
		return false
	}
}

func validProjectQuarantineEntry(entry ProjectQuarantineManifestEntry) bool {
	if !isPortableName(entry.Skill) ||
		(entry.Target != TargetAgents && entry.Target != TargetClaude) ||
		entry.OriginalPlacement != projectOriginalPlacement(entry.Target, entry.Skill) ||
		entry.QuarantinedPlacement != projectQuarantinedPlacement(entry.Target, entry.Skill) ||
		entry.TreeHashAlgorithm != TreeHashAlgorithmSHA256V2 ||
		!lowercaseDigestPattern.MatchString(entry.OldTreeHash) ||
		!validProjectQuarantineEntryStatus(entry.Status) {
		return false
	}
	switch entry.Action {
	case ProjectQuarantineEntryActionUpdate:
		return isCanonicalProjectSourceIdentity(entry.OldSourceIdentity) &&
			isCanonicalProjectSourceIdentity(entry.NewSourceIdentity) &&
			entry.NewSourceIdentity == entry.OldSourceIdentity &&
			lowercaseDigestPattern.MatchString(entry.NewTreeHash) &&
			entry.NewTreeHash != entry.OldTreeHash
	case ProjectQuarantineEntryActionRemove:
		// Empty replacement fields are intentionally omitted by encoding/json so
		// a remove entry cannot be mistaken for an update by restore.
		return isCanonicalProjectSourceIdentity(entry.OldSourceIdentity) &&
			entry.NewSourceIdentity == "" && entry.NewTreeHash == ""
	default:
		return false
	}
}

func validProjectQuarantineStatus(status ProjectQuarantineStatus) bool {
	switch status {
	case ProjectQuarantinePrepared, ProjectQuarantineActive, ProjectQuarantineCommitted,
		ProjectQuarantineRestoring, ProjectQuarantineRestored,
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
