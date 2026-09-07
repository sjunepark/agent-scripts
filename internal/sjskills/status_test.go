package sjskills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func statusFixture(t *testing.T, global bool) (StatusScope, StatusService, *time.Time, *atomic.Int32) {
	t.Helper()
	root := canonicalTempHome(t)
	registry := minimalGlobalRegistry(t)
	if !global {
		if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte("version = 1\nprofiles = [\"dev\"]\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	scope, err := ResolveStatusScope(root, registry, global)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 7, 1, 0, 0, 0, time.UTC)
	calls := &atomic.Int32{}
	service := StatusService{CacheRoot: filepath.Join(t.TempDir(), "status"), Now: func() time.Time { return now }}
	service.Refresh = func(ctx context.Context, desired []DesiredSkill) (StatusSnapshot, error) {
		calls.Add(1)
		expected := map[string]TreeHash{}
		for _, skill := range desired {
			if skill.Manager == ManagerSkillsCLI {
				expected[skill.Name] = classificationHash('a')
			}
		}
		return StatusSnapshot{expected, now}, nil
	}
	return scope, service, &now, calls
}
func requireStatusFinding(t *testing.T, a Advisory, category AdvisoryCategory, skill string, target Target, reason string) {
	t.Helper()
	for _, f := range a.Findings {
		if f.Category == category && f.Skill == skill && f.Target == target && f.Reason == reason {
			return
		}
	}
	t.Fatalf("missing %s %s %s %s in %+v", category, skill, target, reason, a)
}
func TestStatusFreshnessRetryAndIdentity(t *testing.T) {
	scope, service, now, calls := statusFixture(t, false)
	cold := service.Check(context.Background(), scope, nil)
	if cold.Freshness != AdvisoryFresh || cold.Cached || cold.Error != "" || calls.Load() != 1 {
		t.Fatalf("cold %+v calls %d", cold, calls.Load())
	}
	*now = now.Add(time.Hour)
	warm := service.Check(context.Background(), scope, nil)
	if warm.Freshness != AdvisoryFresh || !warm.Cached || calls.Load() != 1 {
		t.Fatalf("warm %+v", warm)
	}
	*now = now.Add(24 * time.Hour)
	service.Refresh = func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
		calls.Add(1)
		return StatusSnapshot{}, errors.New("secret raw subprocess output")
	}
	stale := service.Check(context.Background(), scope, nil)
	if stale.Freshness != AdvisoryStale || stale.Error == "" || strings.Contains(stale.Error, "secret") || calls.Load() != 2 {
		t.Fatalf("stale %+v", stale)
	}
	*now = now.Add(time.Minute)
	again := service.Check(context.Background(), scope, nil)
	if again.Freshness != AdvisoryStale || calls.Load() != 2 {
		t.Fatalf("cooldown %+v calls %d", again, calls.Load())
	}
	*now = now.Add(statusRetryInterval)
	service.Check(context.Background(), scope, nil)
	if calls.Load() != 3 {
		t.Fatalf("retry calls %d", calls.Load())
	}
	changed := scope
	changed.Plan.Desired = cloneDesiredState(scope.Plan.Desired)
	changed.Plan.Desired.Skills[0].FullDepth = true
	unavailable := service.Check(context.Background(), changed, nil)
	if unavailable.Freshness != AdvisoryUnavailable || unavailable.ObservedAt != nil || len(unavailable.Findings) != 0 || calls.Load() != 4 {
		t.Fatalf("identity %+v calls %d", unavailable, calls.Load())
	}
	service.Check(context.Background(), changed, nil)
	if calls.Load() != 4 {
		t.Fatalf("cold failure cooldown lost: %d", calls.Load())
	}
}
func TestStatusWarmCacheReinspectsLocalOwnershipAndContent(t *testing.T) {
	for _, global := range []bool{false, true} {
		t.Run(map[bool]string{false: "project", true: "global"}[global], func(t *testing.T) {
			scope, service, _, calls := statusFixture(t, global)
			skill := scope.Plan.Desired.Skills[0]
			root := filepath.Join(scope.Root, string(TargetAgents), "skills")
			old := writeGlobalSkill(t, root, skill.Name, "old\n")
			record := classificationRecord(skill.Name, TargetAgents, "github:example/skills", old)
			record.Scope = scope.Plan.Desired.Scope
			record.RecordedAt = service.now()
			var statePath string
			var state any
			if global {
				layout, _ := LayoutForGlobal(scope.Root)
				statePath = layout.ProvenanceStatePath
				state = GlobalProvenanceState{Version: GlobalProvenanceStateVersion, Records: []ProvenanceRecord{record}}
			} else {
				layout, _ := LayoutForProject(scope.Root)
				statePath = layout.ReconcilerStatePath
				state = ProvenanceState{Version: ProvenanceStateVersion, Records: []ProvenanceRecord{record}}
			}
			data, _ := json.Marshal(state)
			if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(statePath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			cold := service.Check(context.Background(), scope, nil)
			requireStatusFinding(t, cold, AdvisoryUpdate, skill.Name, TargetAgents, "verified-update")
			requireStatusFinding(t, cold, AdvisoryMissing, skill.Name, TargetClaude, "expected-entry-absent")
			writeGlobalSkill(t, root, skill.Name, "local edit\n")
			writeGlobalSkill(t, root, "extra", "extra\n")
			warm := service.Check(context.Background(), scope, nil)
			requireStatusFinding(t, warm, AdvisoryConflict, skill.Name, TargetAgents, "local-modification")
			requireStatusFinding(t, warm, AdvisoryExtra, "extra", TargetAgents, "not-desired")
			if calls.Load() != 1 {
				t.Fatalf("warm materialized %d", calls.Load())
			}
			if err := os.RemoveAll(filepath.Join(root, skill.Name)); err != nil {
				t.Fatal(err)
			}
			missing := service.Check(context.Background(), scope, nil)
			requireStatusFinding(t, missing, AdvisoryMissing, skill.Name, TargetAgents, "expected-entry-absent")
			// Recreating a byte-identical directory without provenance cannot grant ownership.
			writeGlobalSkill(t, root, skill.Name, "old\n")
			if err := os.Remove(statePath); err != nil {
				t.Fatal(err)
			}
			unowned := service.Check(context.Background(), scope, nil)
			requireStatusFinding(t, unowned, AdvisoryConflict, skill.Name, TargetAgents, "desired-path-unmanaged")
		})
	}
}
func TestStatusCompleteSnapshotsAndHostileCache(t *testing.T) {
	for _, kind := range []string{"oversized", "symlink", "malformed", "wrong-hash", "partial", "wrong-identity", "version", "future", "trailing"} {
		t.Run(kind, func(t *testing.T) {
			scope, service, now, calls := statusFixture(t, false)
			service.Check(context.Background(), scope, nil)
			directory, _ := service.cacheDirectory()
			path := filepath.Join(directory, scope.cacheKey()+".json")
			entry, _ := service.read(scope)
			switch kind {
			case "oversized":
				if err := os.WriteFile(path, []byte(strings.Repeat("x", maxStatusCacheBytes+1)), 0o600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				outside := filepath.Join(t.TempDir(), "outside")
				if err := os.WriteFile(outside, []byte("sentinel"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
				defer func() {
					data, _ := os.ReadFile(outside)
					if string(data) != "sentinel" {
						t.Fatal("followed cache symlink")
					}
				}()
			case "malformed":
				_ = os.WriteFile(path, []byte("{"), 0o600)
			case "trailing":
				data, _ := os.ReadFile(path)
				_ = os.WriteFile(path, append(data, []byte("{}")...), 0o600)
			default:
				switch kind {
				case "wrong-hash":
					entry.Expected[scope.Plan.Desired.Skills[0].Name] = TreeHash{Algorithm: TreeHashAlgorithmSHA256V2, Digest: "invalid"}
				case "partial":
					entry.Expected = map[string]TreeHash{}
				case "wrong-identity":
					entry.Identity = strings.Repeat("b", 64)
				case "version":
					entry.Version = 99
				case "future":
					entry.ObservedAt = now.Add(time.Hour)
				}
				data, _ := json.Marshal(entry)
				_ = os.WriteFile(path, data, 0o600)
			}
			service.Refresh = func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
				calls.Add(1)
				return StatusSnapshot{}, errors.New("offline")
			}
			value := service.Check(context.Background(), scope, nil)
			if kind == "future" {
				if value.Freshness != AdvisoryStale {
					t.Fatalf("future %+v", value)
				}
			} else if value.Freshness != AdvisoryUnavailable || len(value.Findings) != 0 {
				t.Fatalf("unsafe cache %+v", value)
			}
			if value.Error == "" {
				t.Fatalf("missing cache failure %+v", value)
			}
		})
	}
	scope, service, _, _ := statusFixture(t, false)
	service.Refresh = func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
		return StatusSnapshot{map[string]TreeHash{}, service.now()}, nil
	}
	value := service.Check(context.Background(), scope, nil)
	if value.Freshness != AdvisoryUnavailable || len(value.Findings) != 0 {
		t.Fatalf("partial refresh %+v", value)
	}
}
func TestStatusConcurrentRefreshAndSnapshotOrdering(t *testing.T) {
	scope, service, now, calls := statusFixture(t, false)
	entered := make(chan struct{})
	release := make(chan struct{})
	original := service.Refresh
	service.Refresh = func(ctx context.Context, skills []DesiredSkill) (StatusSnapshot, error) {
		close(entered)
		<-release
		return original(ctx, skills)
	}
	done := make(chan Advisory, 1)
	go func() { done <- service.Check(context.Background(), scope, nil) }()
	<-entered
	start := time.Now()
	contended := service.Check(context.Background(), scope, nil)
	if time.Since(start) > time.Second || contended.Freshness != AdvisoryUnavailable || contended.Error == "" {
		t.Fatalf("contention %+v", contended)
	}
	close(release)
	fresh := <-done
	if fresh.Freshness != AdvisoryFresh || calls.Load() != 1 {
		t.Fatalf("fresh %+v calls %d", fresh, calls.Load())
	}
	entry, _ := service.read(scope)
	old := entry
	old.ObservedAt = now.Add(-time.Hour)
	old.Expected = map[string]TreeHash{scope.Plan.Desired.Skills[0].Name: classificationHash('b')}
	if err := service.publish(scope, old); err != nil {
		t.Fatal(err)
	}
	saved, _ := service.read(scope)
	if !reflect.DeepEqual(saved, entry) {
		t.Fatal("older snapshot replaced newer evidence")
	}
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() { defer group.Done(); service.Check(context.Background(), scope, nil) }()
	}
	group.Wait()
	if calls.Load() != 1 {
		t.Fatalf("warm concurrency refreshed %d", calls.Load())
	}
}
func TestStatusClockRollbackAndIdentityInputs(t *testing.T) {
	scope, service, now, calls := statusFixture(t, false)
	service.Check(context.Background(), scope, nil)
	*now = now.Add(-time.Hour)
	value := service.Check(context.Background(), scope, nil)
	if value.Freshness != AdvisoryFresh || calls.Load() != 2 || !value.ObservedAt.Equal(*now) {
		t.Fatalf("rollback %+v", value)
	}
	changes := []func(*StatusScope){
		func(s *StatusScope) { s.Root += "-other" },
		func(s *StatusScope) { s.Registry.Description += " changed" },
		func(s *StatusScope) { s.Plan.Desired.Skills[0].Source = "other/repo" },
		func(s *StatusScope) { s.Plan.Desired.Skills[0].Targets = []Target{TargetAgents} },
		func(s *StatusScope) { s.Plan.Desired.Skills[0].FullDepth = !s.Plan.Desired.Skills[0].FullDepth },
	}
	for _, change := range changes {
		changed := scope
		changed.Plan.Desired = cloneDesiredState(scope.Plan.Desired)
		change(&changed)
		if changed.Matches(scope) {
			t.Fatal("changed identity matched")
		}
	}
}

func TestStatusCacheDirectoryLockAndPruningSafety(t *testing.T) {
	scope, service, now, _ := statusFixture(t, false)
	directory, err := service.cacheDirectory()
	if err != nil {
		t.Fatal(err)
	}
	oldKey := strings.Repeat("d", 64)
	for _, name := range []string{oldKey + ".json", oldKey + ".tmp-abandoned", oldKey + ".lock", scope.cacheKey() + ".tmp-abandoned", "unrelated"} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte("sentinel"), 0600); err != nil {
			t.Fatal(err)
		}
		old := now.Add(-statusRetention - time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	// A stale-looking lock can still be held by a live process. Pruning must
	// honor the OS lock rather than treating its age as abandoned ownership.
	lockPath := filepath.Join(directory, oldKey+".lock")
	lock, err := openStatusFile(lockPath, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := lockApplyFile(lock); err != nil {
		t.Fatal(err)
	}
	value := service.Check(context.Background(), scope, nil)
	if value.Error != "" {
		t.Fatal(value.Error)
	}
	if _, err := os.Stat(filepath.Join(directory, oldKey+".json")); err != nil {
		t.Fatal("pruned active scope")
	}
	if _, err := os.Stat(filepath.Join(directory, scope.cacheKey()+".tmp-abandoned")); !os.IsNotExist(err) {
		t.Fatal("abandoned own temporary file not pruned")
	}
	_ = unlockApplyFile(lock)
	_ = lock.Close()
	*now = now.Add(statusRefreshInterval)
	service.Check(context.Background(), scope, nil)
	for _, name := range []string{oldKey + ".json", oldKey + ".tmp-abandoned", oldKey + ".lock"} {
		if _, err := os.Lstat(filepath.Join(directory, name)); !os.IsNotExist(err) {
			t.Fatalf("abandoned file not pruned: %s %v", name, err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(directory, "unrelated"))
	if string(data) != "sentinel" {
		t.Fatal("unrelated file pruned")
	}
}
func TestStatusPruningPreservesReplacementAndTouchedSnapshot(t *testing.T) {
	_, service, now, _ := statusFixture(t, false)
	for _, replace := range []bool{false, true} {
		t.Run(fmt.Sprint(replace), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "snapshot.json")
			if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			old := now.Add(-statusRetention - time.Hour)
			if err := os.Chtimes(path, old, old); err != nil {
				t.Fatal(err)
			}
			candidate, err := lstatIdentity(path)
			if err != nil {
				t.Fatal(err)
			}
			if replace {
				temp := path + ".new"
				if err := os.WriteFile(temp, []byte("fresh"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(temp, path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Chtimes(path, *now, *now); err != nil {
				t.Fatal(err)
			}
			// Simulate the publication occurring after selection but before the
			// scope lock is acquired. The protected deletion must recheck it.
			service.removeExpiredStatusFile(path, candidate)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("fresh snapshot deleted: %v", err)
			}
		})
	}
}

func TestStatusUnsafeCacheIsOnlyAnAdvisoryFailure(t *testing.T) {
	for _, kind := range []string{"directory-file", "directory-link", "lock-link"} {
		t.Run(kind, func(t *testing.T) {
			scope, service, _, calls := statusFixture(t, false)
			outside := t.TempDir()
			switch kind {
			case "directory-file":
				if err := os.WriteFile(service.CacheRoot, []byte("sentinel"), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory-link":
				if err := os.Symlink(outside, service.CacheRoot); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			case "lock-link":
				directory, _ := service.cacheDirectory()
				if err := os.Symlink(filepath.Join(outside, "not-created"), filepath.Join(directory, scope.cacheKey()+".lock")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			value := service.Check(context.Background(), scope, nil)
			if value.Freshness != AdvisoryUnavailable || value.Error == "" || calls.Load() != 0 {
				t.Fatalf("unsafe cache result=%+v calls=%d", value, calls.Load())
			}
			entries, _ := os.ReadDir(outside)
			if len(entries) != 0 {
				t.Fatalf("unsafe cache wrote outside: %v", entries)
			}
		})
	}
}
func TestStatusManualWorkflowAndProtectedPathsHaveNoInventedUpdates(t *testing.T) {
	scope, service, _, _ := statusFixture(t, true)
	registry := scope.Registry
	registry.Skills = append([]SkillDeclaration(nil), registry.Skills...)
	registry.Skills[0].Manager = ManagerManual
	registry.Skills[0].Mode = ""
	registry.Skills[0].Source = "manual"
	registry.Sources = map[string]Source{"fixture": {Kind: SourceExternal, Location: "example/skills"}, "manual": {Kind: SourceExternal}}
	changed, err := ResolveStatusScope(scope.Root, registry, true)
	if err != nil {
		t.Fatal(err)
	}
	layout, _ := LayoutForGlobal(scope.Root)
	// Global inspection observes these roots without recursively classifying
	// their content as missing/outdated/extra skills.
	writeGlobalSkill(t, layout.CodexSystemSkillsPath, "vendor-only", "vendor")
	writeGlobalSkill(t, layout.CodexPluginCachePath, "plugin-only", "plugin")
	writeGlobalSkill(t, layout.LegacyPiSkillsPath, "legacy-only", "legacy")
	result := service.Check(context.Background(), changed, nil)
	if result.Freshness != AdvisoryFresh || len(result.Findings) != 0 {
		t.Fatalf("manual/protected status %+v", result)
	}
	registry.Skills[0].Manager = ManagerWorkflow
	registry.Skills[0].Workflow = "manual workflow"
	changed, err = ResolveStatusScope(scope.Root, registry, true)
	if err != nil {
		t.Fatal(err)
	}
	result = service.Check(context.Background(), changed, nil)
	if result.Freshness != AdvisoryFresh || len(result.Findings) != 0 {
		t.Fatalf("workflow status %+v", result)
	}
}
