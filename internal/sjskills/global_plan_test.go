package sjskills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func canonicalTempHome(t *testing.T) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return home
}

func minimalGlobalRegistry(t *testing.T) Registry {
	t.Helper()
	registry := Registry{
		Version:     RegistryVersion,
		Description: "global fixture",
		Defaults:    RegistryDefaults{Targets: []Target{TargetAgents, TargetClaude}},
		Global:      GlobalRegistry{Baseline: []string{"base"}},
		Profiles: map[string]Profile{
			"dev":   {Skills: []string{"former"}},
			"go":    {Skills: []string{"go-skill"}},
			"kicpa": {Skills: []string{"kicpa-skill"}},
			"rust":  {Skills: []string{"rust-skill"}},
		},
		Sources: map[string]Source{
			"fixture": {Kind: SourceExternal, Location: "example/skills"},
		},
		Skills: []SkillDeclaration{
			{Name: "base", Source: "fixture", Manager: ManagerSkillsCLI, Mode: ModeCopy},
			{Name: "former", Source: "fixture", Manager: ManagerSkillsCLI, Mode: ModeCopy},
			{Name: "go-skill", Source: "fixture", Manager: ManagerSkillsCLI, Mode: ModeCopy},
			{Name: "kicpa-skill", Source: "fixture", Manager: ManagerSkillsCLI, Mode: ModeCopy},
			{Name: "rust-skill", Source: "fixture", Manager: ManagerSkillsCLI, Mode: ModeCopy},
		},
	}
	if err := ValidateRegistry(registry); err != nil {
		t.Fatal(err)
	}
	return registry
}

func writeGlobalSkill(t *testing.T, root, name, content string) TreeHash {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := HashSkillTree(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

type legacyGlobalRecordFixture struct {
	Root          string    `json:"root"`
	Skill         string    `json:"skill"`
	Source        string    `json:"source"`
	HashAlgorithm string    `json:"hashAlgorithm"`
	Hash          string    `json:"hash"`
	RecordedAt    time.Time `json:"recordedAt"`
}

func writeLegacyGlobalState(t *testing.T, path string, records []legacyGlobalRecordFixture) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(struct {
		Version int                         `json:"version"`
		Records []legacyGlobalRecordFixture `json:"records"`
	}{Version: 1, Records: records})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGlobalLayoutFixesManagedAndProtectedPaths(t *testing.T) {
	home := canonicalTempHome(t)
	layout, err := LayoutForGlobal(home)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"agents":     filepath.Join(home, ".agents", "skills"),
		"claude":     filepath.Join(home, ".claude", "skills"),
		"state":      filepath.Join(home, ".agents", ".global-skill-state.json"),
		"pi":         filepath.Join(home, ".pi", "agent", "skills"),
		"quarantine": filepath.Join(home, ".skill-quarantine"),
		"system":     filepath.Join(home, ".codex", "skills", ".system"),
		"cache":      filepath.Join(home, ".codex", "plugins", "cache"),
	}
	got := map[string]string{
		"agents": layout.AgentsSkillsPath, "claude": layout.ClaudeSkillsPath,
		"state": layout.ProvenanceStatePath, "pi": layout.LegacyPiSkillsPath,
		"quarantine": layout.LegacyQuarantinePath, "system": layout.CodexSystemSkillsPath,
		"cache": layout.CodexPluginCachePath,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("layout = %#v, want %#v", got, want)
	}
	if _, err := LayoutForGlobal("relative"); err == nil {
		t.Fatal("relative global home accepted")
	}
	if _, err := layout.ManagedSkillsPath(Target("other")); err == nil {
		t.Fatal("unsupported global target accepted")
	}
}

func TestInspectGlobalUsesOnlyFixedRootsAndMigratesStrictLegacyState(t *testing.T) {
	home := canonicalTempHome(t)
	layout, err := LayoutForGlobal(home)
	if err != nil {
		t.Fatal(err)
	}
	agentsHash := writeGlobalSkill(t, layout.AgentsSkillsPath, "base", "agents\n")
	_ = writeGlobalSkill(t, layout.ClaudeSkillsPath, "base", "claude\n")
	_ = writeGlobalSkill(t, layout.LegacyPiSkillsPath, "legacy", "legacy\n")
	for _, path := range []string{layout.LegacyQuarantinePath, layout.CodexSystemSkillsPath, layout.CodexPluginCachePath} {
		if err := os.MkdirAll(filepath.Join(path, "must-not-be-enumerated"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{layout.AgentsVendorLockPath, layout.ClaudeVendorLockPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("not parsed as ownership\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	recordedAt := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	writeLegacyGlobalState(t, layout.ProvenanceStatePath, []legacyGlobalRecordFixture{{
		Root: "shared", Skill: "base", Source: "example/skills", HashAlgorithm: agentsHash.Algorithm, Hash: agentsHash.Digest, RecordedAt: recordedAt,
	}})

	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.ProvenanceFormat != GlobalProvenanceLegacyV1 || !inventory.MigrationRequired || !inventory.StateTrusted {
		t.Fatalf("provenance = format %q migration=%t trusted=%t", inventory.ProvenanceFormat, inventory.MigrationRequired, inventory.StateTrusted)
	}
	if len(inventory.State.Records) != 1 || inventory.State.Records[0].Scope != ScopeGlobal || inventory.State.Records[0].Target != TargetAgents || inventory.State.Records[0].SourceIdentity != "github:example/skills" {
		t.Fatalf("migrated records = %#v", inventory.State.Records)
	}
	if len(inventory.LegacyRoot.Entries) != 1 || inventory.LegacyRoot.Entries[0].Name != "legacy" {
		t.Fatalf("legacy inventory = %#v", inventory.LegacyRoot)
	}
	if len(inventory.Protected) != 5 {
		t.Fatalf("protected locations = %#v", inventory.Protected)
	}
	for _, location := range inventory.Protected {
		if !location.Exists || !location.Safe {
			t.Fatalf("protected location = %#v", location)
		}
		if strings.Contains(location.ID, "must-not-be-enumerated") {
			t.Fatalf("protected child leaked: %#v", location)
		}
	}
}

func TestInspectGlobalRejectsMalformedOrUnsafeProvenanceAsAWhole(t *testing.T) {
	validHash := strings.Repeat("a", 64)
	validTime := "2026-08-24T00:00:00Z"
	tests := []struct {
		name string
		data string
	}{
		{name: "unknown-field", data: `{"version":1,"records":[],"extra":true}`},
		{name: "trailing", data: `{"version":1,"records":[]} {}`},
		{name: "invalid-root", data: `{"version":1,"records":[{"root":"pi","skill":"base","source":"example/skills","hashAlgorithm":"tree-sha256-v2","hash":"` + validHash + `","recordedAt":"` + validTime + `"}]}`},
		{name: "invalid-time", data: `{"version":1,"records":[{"root":"shared","skill":"base","source":"example/skills","hashAlgorithm":"tree-sha256-v2","hash":"` + validHash + `","recordedAt":"not-time"}]}`},
		{name: "duplicate", data: `{"version":1,"records":[{"root":"shared","skill":"base","source":"example/skills","hashAlgorithm":"tree-sha256-v2","hash":"` + validHash + `","recordedAt":"` + validTime + `"},{"root":"shared","skill":"base","source":"example/skills","hashAlgorithm":"tree-sha256-v2","hash":"` + validHash + `","recordedAt":"` + validTime + `"}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := canonicalTempHome(t)
			layout, err := LayoutForGlobal(home)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(layout.ProvenanceStatePath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(layout.ProvenanceStatePath, []byte(test.data), 0o600); err != nil {
				t.Fatal(err)
			}
			inventory, err := InspectGlobal(layout)
			if err != nil {
				t.Fatal(err)
			}
			if inventory.StateTrusted || len(inventory.State.Records) != 0 || inventory.MigrationRequired {
				t.Fatalf("malformed state trusted: %#v", inventory)
			}
		})
	}

	t.Run("oversized", func(t *testing.T) {
		home := canonicalTempHome(t)
		layout, _ := LayoutForGlobal(home)
		if err := os.MkdirAll(filepath.Dir(layout.ProvenanceStatePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(layout.ProvenanceStatePath, []byte(strings.Repeat("x", int(maxProvenanceStateBytes)+1)), 0o600); err != nil {
			t.Fatal(err)
		}
		inventory, err := InspectGlobal(layout)
		if err != nil {
			t.Fatal(err)
		}
		if inventory.StateTrusted {
			t.Fatal("oversized global state trusted")
		}
	})
}

func TestInspectGlobalReadsStrictCurrentProvenance(t *testing.T) {
	home := canonicalTempHome(t)
	layout, err := LayoutForGlobal(home)
	if err != nil {
		t.Fatal(err)
	}
	recordedAt := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	state := GlobalProvenanceState{
		Version: GlobalProvenanceStateVersion,
		Records: []ProvenanceRecord{
			{
				Scope: ScopeGlobal, Skill: "base", Target: TargetClaude,
				SourceIdentity: "github:example/skills", TreeHashAlgorithm: TreeHashAlgorithmSHA256V2,
				TreeHash: strings.Repeat("a", 64), RecordedAt: recordedAt,
			},
		},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.ProvenanceStatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.ProvenanceStatePath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	if !inventory.StateTrusted || inventory.ProvenanceFormat != GlobalProvenanceCurrent || inventory.MigrationRequired || !reflect.DeepEqual(inventory.State, state) {
		t.Fatalf("current provenance = %#v", inventory)
	}
}

func TestInspectGlobalRefusesSymlinkedManagedRootWithoutReadingOutside(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("symlink creation requires privileges on some Windows hosts")
	}
	home := canonicalTempHome(t)
	layout, _ := LayoutForGlobal(home)
	outside := canonicalTempHome(t)
	if err := os.MkdirAll(filepath.Join(outside, "secret"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.AgentsSkillsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, layout.AgentsSkillsPath); err != nil {
		t.Fatal(err)
	}
	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	root, ok := globalRootFor(inventory, TargetAgents)
	if !ok || root.Safe || len(root.Entries) != 0 {
		t.Fatalf("symlinked root = %#v", root)
	}
}

func TestGlobalClassificationMigratesOwnershipWithoutTrustingBytesOrVendorLocks(t *testing.T) {
	registry := minimalGlobalRegistry(t)
	desired, err := ResolveGlobal(registry)
	if err != nil {
		t.Fatal(err)
	}
	home := canonicalTempHome(t)
	layout, _ := LayoutForGlobal(home)
	agentsBase := writeGlobalSkill(t, layout.AgentsSkillsPath, "base", "base-v1\n")
	claudeBase := writeGlobalSkill(t, layout.ClaudeSkillsPath, "base", "base-v1\n")
	former := writeGlobalSkill(t, layout.AgentsSkillsPath, "former", "former\n")
	_ = writeGlobalSkill(t, layout.AgentsSkillsPath, "unknown", "unknown\n")
	_ = writeGlobalSkill(t, layout.LegacyPiSkillsPath, "base", "legacy\n")
	if err := os.WriteFile(layout.AgentsVendorLockPath, []byte(`{"skills":{"base":{"source":"example/skills","hashAlgorithm":"tree-sha256-v2","contentHash":"`+agentsBase.Digest+`"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	recordedAt := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	writeLegacyGlobalState(t, layout.ProvenanceStatePath, []legacyGlobalRecordFixture{
		{Root: "shared", Skill: "base", Source: "example/skills", HashAlgorithm: agentsBase.Algorithm, Hash: agentsBase.Digest, RecordedAt: recordedAt},
		{Root: "shared", Skill: "former", Source: "example/skills", HashAlgorithm: former.Algorithm, Hash: former.Digest, RecordedAt: recordedAt},
		{Root: "claude", Skill: "base", Source: "example/skills", HashAlgorithm: claudeBase.Algorithm, Hash: claudeBase.Digest, RecordedAt: recordedAt},
	})
	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]TreeHash{"base": agentsBase}
	classification, err := ClassifyGlobal(registry, desired, expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := TranslateGlobalClassification(Plan{Desired: desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	assertGlobalAction(t, plan, PlanActionUnchanged, "base", TargetAgents)
	assertGlobalAction(t, plan, PlanActionUnchanged, "base", TargetClaude)
	if hasPlanAction(plan, PlanActionQuarantine) {
		t.Fatalf("read-only migration proposed quarantine: %#v", plan.Operations)
	}
	if !hasPlanWarning(plan.Warnings, "legacy-preserved", "pi/base") || !hasPlanWarning(plan.Warnings, "unmanaged-preserved", "unknown") || !hasPlanWarning(plan.Warnings, "global-migration", "former") {
		t.Fatalf("migration warnings = %#v", plan.Warnings)
	}
	if !hasPlanEvidence(plan.Evidence, "provenance-migration") {
		t.Fatalf("migration evidence = %#v", plan.Evidence)
	}

	// Removing the reconciler state leaves identical bytes and a matching
	// vendor lock, neither of which is ownership evidence in sjskills v1.
	if err := os.Remove(layout.ProvenanceStatePath); err != nil {
		t.Fatal(err)
	}
	inventory, err = InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err = ClassifyGlobal(registry, desired, expected, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err = TranslateGlobalClassification(Plan{Desired: desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	assertGlobalAction(t, plan, PlanActionBlocked, "base", TargetAgents)
	assertGlobalAction(t, plan, PlanActionBlocked, "base", TargetClaude)
	if !hasPlanWarning(plan.Warnings, "global-migration", "former-global-skill-preserved") {
		t.Fatalf("former-profile warning = %#v", plan.Warnings)
	}
}

func TestGlobalClassificationDistinguishesOutdatedModifiedAndStaleProvenance(t *testing.T) {
	registry := minimalGlobalRegistry(t)
	desired, _ := ResolveGlobal(registry)
	home := canonicalTempHome(t)
	layout, _ := LayoutForGlobal(home)
	oldAgents := writeGlobalSkill(t, layout.AgentsSkillsPath, "base", "old\n")
	oldClaude := writeGlobalSkill(t, layout.ClaudeSkillsPath, "base", "old\n")
	recordedAt := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	writeLegacyGlobalState(t, layout.ProvenanceStatePath, []legacyGlobalRecordFixture{
		{Root: "shared", Skill: "base", Source: "example/skills", HashAlgorithm: oldAgents.Algorithm, Hash: oldAgents.Digest, RecordedAt: recordedAt},
		{Root: "claude", Skill: "base", Source: "example/skills", HashAlgorithm: oldClaude.Algorithm, Hash: oldClaude.Digest, RecordedAt: recordedAt},
		{Root: "shared", Skill: "former", Source: "example/skills", HashAlgorithm: oldAgents.Algorithm, Hash: oldAgents.Digest, RecordedAt: recordedAt},
	})
	if err := os.WriteFile(filepath.Join(layout.AgentsSkillsPath, "base", "SKILL.md"), []byte("modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	expectedDir := filepath.Join(t.TempDir(), "expected")
	expected := writeGlobalSkill(t, expectedDir, "base", "new\n")
	inventory, err := InspectGlobal(layout)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := ClassifyGlobal(registry, desired, map[string]TreeHash{"base": expected}, inventory)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := TranslateGlobalClassification(Plan{Desired: desired, Operations: []PlanOperation{}, Warnings: []Warning{}, Evidence: []Evidence{}}, classification)
	if err != nil {
		t.Fatal(err)
	}
	assertGlobalAction(t, plan, PlanActionBlocked, "base", TargetAgents)
	assertGlobalAction(t, plan, PlanActionUpdate, "base", TargetClaude)
	if hasPlanAction(plan, PlanActionQuarantine) || hasGlobalOperation(plan, "former", TargetAgents) {
		t.Fatalf("stale former placement became an operation: %#v", plan.Operations)
	}
	if !hasPlanWarning(plan.Warnings, "global-migration", "former") {
		t.Fatalf("stale former placement was not reported: %#v", plan.Warnings)
	}
}

func globalRootFor(inventory GlobalInventory, target Target) (ProjectSkillsRootInventory, bool) {
	for _, root := range inventory.Roots {
		if root.Target == target {
			return root, true
		}
	}
	return ProjectSkillsRootInventory{}, false
}

func assertGlobalAction(t *testing.T, plan Plan, action PlanAction, skill string, target Target) {
	t.Helper()
	for _, operation := range plan.Operations {
		if operation.Action == action && operation.Skill == skill && operation.Target == target {
			return
		}
	}
	t.Fatalf("missing %s %s @ %s in %#v", action, skill, target, plan.Operations)
}

func hasPlanAction(plan Plan, action PlanAction) bool {
	for _, operation := range plan.Operations {
		if operation.Action == action {
			return true
		}
	}
	return false
}

func hasGlobalOperation(plan Plan, skill string, target Target) bool {
	for _, operation := range plan.Operations {
		if operation.Skill == skill && operation.Target == target {
			return true
		}
	}
	return false
}

func hasPlanWarning(warnings []Warning, code, fragment string) bool {
	for _, warning := range warnings {
		if warning.Code == code && strings.Contains(warning.Message, fragment) {
			return true
		}
	}
	return false
}

func hasPlanEvidence(evidence []Evidence, kind string) bool {
	for _, item := range evidence {
		if item.Kind == kind {
			return true
		}
	}
	return false
}
