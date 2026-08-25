package sjskills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

const maxReviewedPlanBytes = 16 << 20

// ReviewedPlan is the exact JSON artifact whose digest received operator
// approval. Apply compares its stable plan evidence with a fresh plan before
// using the still-live materialized content from that recheck.
type ReviewedPlan struct {
	envelope Envelope
	sha256   string
	verified bool
}

// SHA256 returns the normalized digest of the exact bytes read and verified by
// LoadReviewedPlan. The reviewed envelope remains opaque so callers cannot
// manufacture approval evidence without passing through that loader.
func (p ReviewedPlan) SHA256() string { return p.sha256 }

// LoadReviewedPlan reads the approval artifact once, binds it to the supplied
// digest, and accepts only a successful global plan emitted by sjskills.
func LoadReviewedPlan(path, approvedSHA256 string) (ReviewedPlan, error) {
	digest, err := normalizeApprovedSHA256(approvedSHA256)
	if err != nil {
		return ReviewedPlan{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return ReviewedPlan{}, fmt.Errorf("open reviewed plan: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return ReviewedPlan{}, fmt.Errorf("stat reviewed plan: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() > maxReviewedPlanBytes {
		return ReviewedPlan{}, &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlan", Message: "approved plan must be a bounded regular file"}
	}
	data, err := io.ReadAll(io.LimitReader(file, maxReviewedPlanBytes+1))
	if err != nil {
		return ReviewedPlan{}, fmt.Errorf("read reviewed plan: %w", err)
	}
	if len(data) > maxReviewedPlanBytes {
		return ReviewedPlan{}, &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlan", Message: "approved plan must be a bounded regular file"}
	}
	actual := sha256.Sum256(data)
	actualDigest := hex.EncodeToString(actual[:])
	if actualDigest != digest {
		return ReviewedPlan{}, &Issue{Code: IssueReconciliationConflict, Path: "apply.approvedPlanSha256", Message: "approved plan digest does not match the reviewed artifact"}
	}

	var envelope Envelope
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return ReviewedPlan{}, &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlan", Message: "approved plan is not a valid sjskills JSON envelope"}
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ReviewedPlan{}, &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlan", Message: "approved plan contains trailing JSON content"}
	}
	if envelope.Operation != CommandOperationPlan || envelope.Result != ResultSuccess || envelope.Error != nil || envelope.Plan == nil || envelope.Plan.Desired.Scope != ScopeGlobal {
		return ReviewedPlan{}, &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlan", Message: "approved artifact must be a successful global plan"}
	}
	return ReviewedPlan{envelope: envelope, sha256: digest, verified: true}, nil
}

func normalizeApprovedSHA256(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != sha256.Size*2 {
		return "", &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlanSha256", Message: "approved plan SHA-256 must contain 64 hexadecimal characters"}
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", &Issue{Code: IssueMalformedInput, Path: "apply.approvedPlanSha256", Message: "approved plan SHA-256 must contain 64 hexadecimal characters"}
	}
	return value, nil
}

// VerifyReviewedPlanRecheck requires the current global plan and all stable
// evidence to equal the approved artifact. Materialization cleanup evidence is
// excluded because it describes the completed plan command rather than the
// still-live apply session retained after this check.
func VerifyReviewedPlanRecheck(reviewed ReviewedPlan, current Envelope) error {
	current.Operation = CommandOperationPlan
	if !reviewed.verified || !reflect.DeepEqual(normalizeReviewedEnvelope(reviewed.envelope), normalizeReviewedEnvelope(current)) {
		return &Issue{Code: IssueReconciliationConflict, Path: "apply.approvedPlan", Message: "current global plan does not match the approved plan; review a new artifact"}
	}
	return nil
}

func normalizeReviewedEnvelope(envelope Envelope) Envelope {
	evidence := make([]Evidence, 0, len(envelope.Evidence))
	for _, item := range envelope.Evidence {
		if item.Kind != "materialization" {
			evidence = append(evidence, item)
		}
	}
	envelope.Evidence = evidence
	if envelope.Plan != nil {
		plan := *envelope.Plan
		plan.Warnings = nil
		plan.Evidence = nil
		envelope.Plan = &plan
	}
	return envelope
}
