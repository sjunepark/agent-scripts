package sjskills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// BenchmarkStatusWarmDevGoAndGlobal models the CLI's two scopes with the real
// embedded selections, eight 8 KiB files per placement, and trusted provenance.
// It measures cache lookup plus fresh local hashing/classification, not network.
func BenchmarkStatusWarmDevGoAndGlobal(b *testing.B) {
	registry, err := EmbeddedRegistry()
	if err != nil {
		b.Fatal(err)
	}
	service := StatusService{CacheRoot: filepath.Join(b.TempDir(), "status"), Refresh: func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
		b.Error("warm status invoked an upstream dependency")
		return StatusSnapshot{}, context.Canceled
	}}
	scopes := make([]StatusScope, 2)
	placements := 0
	for index, global := range []bool{false, true} {
		root := b.TempDir()
		if !global {
			if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte("version = 1\nprofiles = [\"dev\", \"go\"]\n"), 0644); err != nil {
				b.Fatal(err)
			}
		}
		scope, err := ResolveStatusScope(root, registry, global)
		if err != nil {
			b.Fatal(err)
		}
		scopes[index] = scope
		records := []ProvenanceRecord{}
		expected := map[string]TreeHash{}
		for _, skill := range scope.Plan.Desired.Skills {
			if skill.Manager != ManagerSkillsCLI {
				continue
			}
			for _, target := range skill.Targets {
				path := filepath.Join(scope.Root, string(target), "skills", skill.Name)
				if err := os.MkdirAll(path, 0755); err != nil {
					b.Fatal(err)
				}
				for file := range 8 {
					name := fmt.Sprintf("reference-%d.md", file)
					if file == 0 {
						name = "SKILL.md"
					}
					if err := os.WriteFile(filepath.Join(path, name), []byte(strings.Repeat("x", 8<<10)), 0644); err != nil {
						b.Fatal(err)
					}
				}
				hash, err := HashSkillTree(path)
				if err != nil {
					b.Fatal(err)
				}
				expected[skill.Name] = hash
				placements++
				source, ok := canonicalProjectSourceIdentity(skill.Source)
				if !ok {
					b.Fatal("invalid source")
				}
				records = append(records, ProvenanceRecord{Scope: scope.Plan.Desired.Scope, Skill: skill.Name, Target: target, SourceIdentity: source, TreeHashAlgorithm: hash.Algorithm, TreeHash: hash.Digest, RecordedAt: time.Now().UTC()})
			}
		}
		var path string
		var state any
		if global {
			layout, _ := LayoutForGlobal(scope.Root)
			path = layout.ProvenanceStatePath
			state = GlobalProvenanceState{Version: GlobalProvenanceStateVersion, Records: records}
		} else {
			layout, _ := LayoutForProject(scope.Root)
			path = layout.ReconcilerStatePath
			state = ProvenanceState{Version: ProvenanceStateVersion, Records: records}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			b.Fatal(err)
		}
		data, _ := json.Marshal(state)
		if err := os.WriteFile(path, data, 0600); err != nil {
			b.Fatal(err)
		}
		observed := time.Now().UTC()
		value := service.Check(context.Background(), scope, &StatusSnapshot{expected, observed})
		if value.Freshness != AdvisoryFresh || len(value.Findings) != 0 || value.Error != "" {
			b.Fatalf("fixture %+v", value)
		}
	}
	b.SetBytes(int64(placements * 8 * 8 << 10))
	b.ResetTimer()
	for range b.N {
		var group sync.WaitGroup
		for _, scope := range scopes {
			group.Add(1)
			go func(scope StatusScope) {
				defer group.Done()
				value := service.Check(context.Background(), scope, nil)
				if value.Freshness != AdvisoryFresh || len(value.Findings) != 0 || value.Error != "" {
					b.Errorf("warm %+v", value)
				}
			}(scope)
		}
		group.Wait()
	}
	b.ReportMetric(float64(placements), "placements")
	b.ReportMetric(float64(placements*8), "files")
}
