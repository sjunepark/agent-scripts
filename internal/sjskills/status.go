package sjskills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

// AdvisoryFreshness describes upstream evidence, independently of local drift.
type AdvisoryFreshness string

const (
	AdvisoryFresh       AdvisoryFreshness = "fresh"
	AdvisoryStale       AdvisoryFreshness = "stale"
	AdvisoryUnavailable AdvisoryFreshness = "unavailable"
)

type AdvisoryCategory string

const (
	AdvisoryUpdate   AdvisoryCategory = "update"
	AdvisoryMissing  AdvisoryCategory = "missing"
	AdvisoryExtra    AdvisoryCategory = "extra"
	AdvisoryConflict AdvisoryCategory = "conflict"
)

type AdvisoryFinding struct {
	Category AdvisoryCategory `json:"category"`
	Skill    string           `json:"skill"`
	Target   Target           `json:"target"`
	Reason   string           `json:"reason"`
}

// Advisory carries no authority to apply. Findings always describe a new local
// inventory; ObservedAt refers only to the complete expected-content snapshot.
type Advisory struct {
	Scope         Scope             `json:"scope"`
	Freshness     AdvisoryFreshness `json:"freshness"`
	ObservedAt    *time.Time        `json:"observedAt,omitempty"`
	Cached        bool              `json:"cached"`
	Findings      []AdvisoryFinding `json:"findings"`
	Error         string            `json:"error,omitempty"`
	ReviewCommand string            `json:"reviewCommand"`
}

func newAdvisory(scope Scope) Advisory {
	command := "sjskills plan"
	if scope == ScopeGlobal {
		command += " --global"
	}
	return Advisory{Scope: scope, Freshness: AdvisoryUnavailable, Findings: []AdvisoryFinding{}, ReviewCommand: command}
}
func validateAdvisories(values []Advisory) error {
	seen := map[Scope]bool{}
	for _, a := range values {
		if (a.Scope != ScopeGlobal && a.Scope != ScopeProject) || seen[a.Scope] {
			return errors.New("invalid advisory scope")
		}
		seen[a.Scope] = true
		if a.ReviewCommand != newAdvisory(a.Scope).ReviewCommand {
			return errors.New("invalid advisory review command")
		}
		switch a.Freshness {
		case AdvisoryFresh, AdvisoryStale:
			if a.ObservedAt == nil || a.ObservedAt.IsZero() {
				return errors.New("missing advisory observation")
			}
		case AdvisoryUnavailable:
			if a.ObservedAt != nil || len(a.Findings) != 0 || a.Error == "" {
				return errors.New("invalid unavailable advisory")
			}
		default:
			return errors.New("invalid advisory freshness")
		}
		for _, f := range a.Findings {
			if f.Category != AdvisoryUpdate && f.Category != AdvisoryMissing && f.Category != AdvisoryExtra && f.Category != AdvisoryConflict {
				return errors.New("invalid advisory category")
			}
			if f.Target != "" && f.Target != TargetAgents && f.Target != TargetClaude {
				return errors.New("invalid advisory target")
			}
			if f.Reason == "" {
				return errors.New("missing advisory reason")
			}
		}
	}
	return nil
}
func summarizeStatus(plan Plan) []AdvisoryFinding {
	findings := []AdvisoryFinding{}
	for _, op := range plan.Operations {
		var category AdvisoryCategory
		switch op.Action {
		case PlanActionInstall:
			category = AdvisoryMissing
		case PlanActionUpdate:
			category = AdvisoryUpdate
		case PlanActionQuarantine:
			category = AdvisoryExtra
		case PlanActionBlocked:
			category = AdvisoryConflict
		default:
			continue
		}
		findings = append(findings, AdvisoryFinding{category, op.Skill, op.Target, op.Reason})
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Skill != b.Skill {
			return a.Skill < b.Skill
		}
		if a.Target != b.Target {
			return a.Target < b.Target
		}
		return a.Reason < b.Reason
	})
	return findings
}

// StatusScope is resolved afresh before looking up advisory evidence. Its
// identity includes every desired/source/install input and the embedded registry.
type StatusScope struct {
	Root     string
	Registry Registry
	Plan     Plan
}

func ResolveStatusScope(directory string, registry Registry, global bool) (StatusScope, error) {
	request := ResolveRequest{Registry: registry, Global: global}
	root := directory
	if !global {
		project, err := DiscoverProjectRoot(directory)
		if err != nil {
			return StatusScope{}, err
		}
		manifest, err := ReadManifest(project.ManifestPath)
		if err != nil {
			return StatusScope{}, err
		}
		request.Manifest = &manifest
		root = project.Root
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return StatusScope{}, err
	}
	root, err = filepath.Abs(canonical)
	if err != nil {
		return StatusScope{}, err
	}
	plan, err := BuildPlan(request)
	if err != nil {
		return StatusScope{}, err
	}
	return StatusScope{root, registry, plan}, nil
}

// Matches checks whether command-produced hashes belong to this exact scope.
func (s StatusScope) Matches(other StatusScope) bool { return s.identity() == other.identity() }

func (s StatusScope) identity() string {
	data, _ := json.Marshal(struct {
		Root      string
		Registry  Registry
		Desired   DesiredState
		CLI, Hash string
		Format    int
	}{s.Root, s.Registry, s.Plan.Desired, SkillsCLIVersion, TreeHashAlgorithmSHA256V2, statusCacheVersion})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (s StatusScope) cacheKey() string {
	sum := sha256.Sum256([]byte(string(s.Plan.Desired.Scope) + "\x00" + s.Root))
	return hex.EncodeToString(sum[:])
}

// InspectStatusPlan reuses the reconciliation inventory, ownership classifier,
// and plan translator. No cached findings or provenance are accepted here.
func InspectStatusPlan(s StatusScope, expected map[string]TreeHash) (Plan, error) {
	if s.Plan.Desired.Scope == ScopeGlobal {
		layout, err := LayoutForGlobal(s.Root)
		if err != nil {
			return Plan{}, err
		}
		inventory, err := InspectGlobal(layout)
		if err != nil {
			return Plan{}, err
		}
		classification, err := ClassifyGlobal(s.Registry, s.Plan.Desired, expected, inventory)
		if err != nil {
			return Plan{}, err
		}
		return TranslateGlobalClassification(s.Plan, classification)
	}
	layout, err := LayoutForProject(s.Root)
	if err != nil {
		return Plan{}, err
	}
	inventory, err := InspectProject(layout)
	if err != nil {
		return Plan{}, err
	}
	classification, err := ClassifyProject(s.Plan.Desired, expected, inventory)
	if err != nil {
		return Plan{}, err
	}
	return TranslateProjectClassification(s.Plan, classification)
}

const StatusRefreshBudget = 30 * time.Second
const statusRefreshInterval = 24 * time.Hour
const statusRetryInterval = 15 * time.Minute

// StatusSnapshot is disposable expected-content evidence, never an apply session.
type StatusSnapshot struct {
	Expected   map[string]TreeHash
	ObservedAt time.Time
}
type StatusRefresh func(context.Context, []DesiredSkill) (StatusSnapshot, error)
type StatusService struct {
	CacheRoot string
	Now       func() time.Time
	Refresh   StatusRefresh
}

func (s StatusService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
func RefreshStatusSnapshot(ctx context.Context, skills []DesiredSkill) (StatusSnapshot, error) {
	materialized, err := NewMaterializer(MaterializerConfig{}).Materialize(ctx, skills)
	if err != nil {
		if materialized != nil {
			_ = materialized.Cleanup()
		}
		return StatusSnapshot{}, err
	}
	if materialized == nil {
		return StatusSnapshot{}, errors.New("no snapshot")
	}
	verifyErr := materialized.Verify()
	expected := map[string]TreeHash{}
	if verifyErr == nil {
		for _, snapshot := range materialized.Snapshots() {
			if snapshot == nil {
				verifyErr = errors.New("nil snapshot")
				break
			}
			expected[snapshot.Skill.Name] = snapshot.Hash
		}
	}
	cleanupErr := materialized.Cleanup()
	if verifyErr != nil || cleanupErr != nil {
		return StatusSnapshot{}, errors.New("snapshot verification or cleanup failed")
	}
	return StatusSnapshot{expected, time.Now().UTC()}, nil
}
func validStatusExpected(desired DesiredState, expected map[string]TreeHash) bool {
	count := 0
	for _, skill := range desired.Skills {
		if skill.Manager != ManagerSkillsCLI {
			continue
		}
		count++
		h, ok := expected[skill.Name]
		if !ok || h.Algorithm != TreeHashAlgorithmSHA256V2 || !lowercaseDigestPattern.MatchString(h.Digest) {
			return false
		}
	}
	return len(expected) == count
}
func (s StatusService) Check(ctx context.Context, scope StatusScope, reusable *StatusSnapshot) Advisory {
	result := newAdvisory(scope.Plan.Desired.Scope)
	now := s.now()
	entry, err := s.read(scope)
	if err != nil {
		entry = statusCacheEntry{Version: statusCacheVersion, Identity: scope.identity()}
	}
	result.Cached = true
	if reusable != nil && validStatusExpected(scope.Plan.Desired, reusable.Expected) && !reusable.ObservedAt.IsZero() && !reusable.ObservedAt.After(now) {
		entry = statusCacheEntry{Version: statusCacheVersion, Identity: scope.identity(), Expected: reusable.Expected, ObservedAt: reusable.ObservedAt}
		result.Cached = false
		if err := s.publish(scope, entry); err != nil {
			result.Error = "status cache could not be written"
		}
	} else if !entry.fresh(now) && entry.retryDue(now) {
		unlock, lockErr := s.lock(scope)
		if lockErr != nil {
			result.Error = "upstream status could not be refreshed (cache unavailable or another check active)"
		} else {
			// A peer may have completed either a refresh or a failed attempt between
			// our lookup and lock acquisition. Both snapshot and cooldown are reused.
			if latest, readErr := s.read(scope); readErr == nil {
				entry = latest
			}
			now = s.now()
			if !entry.fresh(now) {
				if !entry.retryDue(now) {
					result.Error = "upstream status could not be refreshed; retry cooldown active"
				} else {
					refresh := s.Refresh
					if refresh == nil {
						refresh = RefreshStatusSnapshot
					}
					snapshot, refreshErr := refresh(ctx, scope.Plan.Desired.Skills)
					finished := s.now()
					if refreshErr == nil && ctx.Err() == nil && validStatusExpected(scope.Plan.Desired, snapshot.Expected) && !snapshot.ObservedAt.IsZero() && !snapshot.ObservedAt.After(finished) {
						entry = statusCacheEntry{Version: statusCacheVersion, Identity: scope.identity(), Expected: snapshot.Expected, ObservedAt: snapshot.ObservedAt}
						result.Cached = false
					} else {
						entry.RetryAt = finished.Add(statusRetryInterval)
						result.Error = "upstream status could not be refreshed"
					}
					if err := s.write(scope, entry); err != nil {
						if result.Error != "" {
							result.Error += "; status cache could not be written"
						} else {
							result.Error = "status cache could not be written"
						}
					}
					s.prune(scope)
				}
			}
			unlock()
		}
	} else if !entry.fresh(now) {
		result.Error = "upstream status could not be refreshed; retry cooldown active"
	}
	if entry.ObservedAt.IsZero() || !validStatusExpected(scope.Plan.Desired, entry.Expected) {
		result.Cached = false
		if result.Error == "" {
			result.Error = "upstream status unavailable"
		}
		return result
	}
	result.ObservedAt = &entry.ObservedAt
	result.Freshness = AdvisoryFresh
	if !entry.fresh(s.now()) {
		result.Freshness = AdvisoryStale
		result.Cached = true
	}
	plan, err := InspectStatusPlan(scope, entry.Expected)
	if err != nil {
		result.Freshness = AdvisoryUnavailable
		result.ObservedAt = nil
		result.Error = fmt.Sprintf("%s local status could not be inspected", scope.Plan.Desired.Scope)
		return result
	}
	result.Findings = summarizeStatus(plan)
	return result
}
