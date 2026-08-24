package sjskills

import "time"

// ProvenanceState is machine-local derived state under DerivedLayout's
// .sjskills directory. It records verified placement identity only; the
// committed sjskills.toml remains the sole project intent and this state is
// not a second configuration source.
type ProvenanceState struct {
	Version int                `json:"version"`
	Records []ProvenanceRecord `json:"records"`
}

// ProvenanceRecord contains the minimum evidence needed to recognize a
// previously verified generated placement and recover it later. It does not
// contain rollout, lock, or execution fields.
type ProvenanceRecord struct {
	Scope             Scope     `json:"scope"`
	Skill             string    `json:"skill"`
	Target            Target    `json:"target"`
	SourceIdentity    string    `json:"sourceIdentity"`
	TreeHashAlgorithm string    `json:"treeHashAlgorithm"`
	TreeHash          string    `json:"treeHash"`
	RecordedAt        time.Time `json:"recordedAt"`
}
