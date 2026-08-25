package sjskills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fixtureRegistry(t *testing.T) Registry {
	t.Helper()
	registry, err := EmbeddedRegistry()
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func fixtureManifest(t *testing.T, name string) Manifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "manifests", name))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func issueCode(err error, code IssueCode) bool {
	var validation *ValidationErrors
	if !errors.As(err, &validation) {
		return false
	}
	for _, issue := range validation.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func TestCanonicalRegistryAndProfiles(t *testing.T) {
	registry := fixtureRegistry(t)
	if registry.Version != RegistryVersion {
		t.Fatalf("version = %d", registry.Version)
	}
	if len(registry.Skills) != 44 {
		t.Fatalf("skills = %d, want 44", len(registry.Skills))
	}
	global, err := ResolveGlobal(registry)
	if err != nil {
		t.Fatal(err)
	}
	if len(global.Skills) != 9 {
		t.Fatalf("global baseline = %d, want 9", len(global.Skills))
	}
	if got := global.Skills[0].Name; got != "brainstorming" {
		t.Fatalf("first global skill = %q", got)
	}
	if got := global.Skills[0].Targets; len(got) != 1 || got[0] != TargetAgents {
		t.Fatalf("brainstorming targets = %#v", got)
	}
}

func TestProjectResolutionDevGoAndDirect(t *testing.T) {
	registry := fixtureRegistry(t)
	manifest := fixtureManifest(t, "dev-go.toml")
	state, err := ResolveProject(registry, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 27 {
		t.Fatalf("dev+go skills = %d, want 27", len(state.Skills))
	}
	for _, skill := range state.Skills {
		if skill.Scope != ScopeProject {
			t.Errorf("%s scope = %s", skill.Name, skill.Scope)
		}
	}

	manifest = fixtureManifest(t, "direct-third-party.toml")
	state, err = ResolveProject(registry, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 1 || state.Skills[0].Source != "example/third-party-skills/review" {
		t.Fatalf("direct result = %#v", state.Skills)
	}
	if !state.Skills[0].FullDepth || len(state.Skills[0].Targets) != 2 {
		t.Fatalf("direct metadata = %#v", state.Skills[0])
	}
}

func TestCanonicalManifestIncludesDirectSourceIdentity(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "sjskills.toml"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Profiles) != 2 || len(manifest.Direct) != 1 {
		t.Fatalf("manifest = %#v", manifest)
	}
	state, err := ResolveProject(fixtureRegistry(t), manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 28 {
		t.Fatalf("canonical desired state = %#v", state)
	}
	var direct *DesiredSkill
	for index := range state.Skills {
		if state.Skills[index].Name == "third-party-review" {
			direct = &state.Skills[index]
			break
		}
	}
	if direct == nil || direct.Source != "example/third-party-skills/review" || direct.Manager != ManagerSkillsCLI || direct.Mode != ModeCopy {
		t.Fatalf("canonical direct desired state = %#v", direct)
	}
}

func TestBuildPlanKeepsWarningsAndEvidenceAtPureBoundary(t *testing.T) {
	registry := fixtureRegistry(t)
	manifest := fixtureManifest(t, "dev-go.toml")
	plan, err := BuildPlan(ResolveRequest{Registry: registry, Manifest: &manifest})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Evidence) != 1 || plan.Evidence[0].Kind != "resolution" {
		t.Fatalf("evidence = %#v", plan.Evidence)
	}
	if plan.Operations == nil || len(plan.Operations) != 0 {
		t.Fatalf("pure resolution operations = %#v, want a stable empty list", plan.Operations)
	}
	if len(plan.Warnings) != 2 {
		t.Fatalf("warnings = %#v", plan.Warnings)
	}
	if _, err := BuildPlan(ResolveRequest{Registry: registry, Global: true, Manifest: &manifest}); err != nil {
		t.Fatal(err)
	}
}

func TestKicpaAndManagerBoundaries(t *testing.T) {
	state, err := ResolveProject(fixtureRegistry(t), fixtureManifest(t, "kicpa.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 1 || state.Skills[0].Name != "windows-cleanup" || state.Skills[0].Targets[0] != TargetAgents {
		t.Fatalf("kicpa = %#v", state.Skills)
	}

	registry := fixtureRegistry(t)
	dev := fixtureManifest(t, "dev-go.toml")
	state, err = ResolveProject(registry, dev)
	if err != nil {
		t.Fatal(err)
	}
	var manual int
	for _, skill := range state.Skills {
		if skill.Manager == ManagerManual {
			manual++
		}
	}
	if manual != 2 {
		t.Fatalf("manual skills = %d, want 2", manual)
	}
	registry = fixtureRegistry(t)
	for _, declaration := range registry.Skills {
		if declaration.Name == "impeccable" && declaration.Manager != ManagerWorkflow {
			t.Fatalf("impeccable manager = %q, want workflow", declaration.Manager)
		}
	}
	for _, skill := range state.Skills {
		if skill.Manager == ManagerWorkflow {
			t.Fatalf("catalog workflow unexpectedly selected: %#v", skill)
		}
	}
}

func TestManifestAndRegistryInvalidCases(t *testing.T) {
	registry := fixtureRegistry(t)
	for _, name := range []string{"collision.toml", "source.toml"} {
		data, err := os.ReadFile(filepath.Join("testdata", "invalid", name))
		if err != nil {
			t.Fatal(err)
		}
		manifest, err := ParseManifest(data)
		if name == "source.toml" {
			if err == nil || !issueCode(err, IssueInvalidSource) {
				t.Fatalf("%s parse error = %v, want invalid source", name, err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ResolveProject(registry, manifest); err == nil || !issueCode(err, IssueCollision) {
			t.Fatalf("%s resolution error = %v, want collision", name, err)
		}
	}
	unknown, err := os.ReadFile(filepath.Join("testdata", "invalid", "registry-unknown-field.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseRegistry(unknown); err == nil || !issueCode(err, IssueUnknownField) {
		t.Fatalf("unknown registry error = %v", err)
	}
	unknownManifest, err := os.ReadFile(filepath.Join("testdata", "invalid", "unknown-field.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseManifest(unknownManifest); err == nil || !issueCode(err, IssueUnknownField) {
		t.Fatalf("unknown manifest error = %v", err)
	}
}

func TestValidationRejectsSourceAndCollisionsBeforeResolution(t *testing.T) {
	registry := fixtureRegistry(t)
	registry.Profiles["go"] = Profile{Skills: []string{"modern-go", "modern-rust"}}
	if err := ValidateRegistry(registry); err == nil || !issueCode(err, IssueCollision) {
		t.Fatalf("profile collision error = %v", err)
	}
	registry = fixtureRegistry(t)
	registry.Skills[0].Manager = ManagerSkillsCLI
	registry.Sources[registry.Skills[0].Source] = Source{Kind: SourceExternal, Location: "npm:unsafe@latest"}
	if err := ValidateRegistry(registry); err == nil || !issueCode(err, IssueInvalidSource) {
		t.Fatalf("invalid source error = %v", err)
	}
	unknownDirect, err := os.ReadFile(filepath.Join("testdata", "invalid", "direct-unsupported-field.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseManifest(unknownDirect); err == nil || !issueCode(err, IssueUnknownField) {
		t.Fatalf("unsupported direct field error = %v", err)
	}
}

func TestValidationRejectsLegacySymlinkMode(t *testing.T) {
	registry := fixtureRegistry(t)
	for index := range registry.Skills {
		if registry.Skills[index].Manager == ManagerSkillsCLI {
			registry.Skills[index].Mode = InstallMode("symlink")
			break
		}
	}
	if err := ValidateRegistry(registry); err == nil || !issueCode(err, IssueInvalidMode) {
		t.Fatalf("legacy symlink mode error = %v, want invalid mode", err)
	}

	desired := DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{{
		Name: "legacy", Source: "owner/repo", Scope: ScopeProject,
		Manager: ManagerSkillsCLI, Mode: InstallMode("symlink"), Targets: []Target{TargetAgents},
	}}}
	root := t.TempDir()
	layout, err := LayoutForProject(root)
	if err != nil {
		t.Fatal(err)
	}
	inventory := ProjectInventory{
		Root: root, StatePath: layout.ReconcilerStatePath, StateTrusted: true,
		Roots: []ProjectSkillsRootInventory{
			{Target: TargetAgents, Path: layout.AgentsSkillsPath, Safe: true},
			{Target: TargetClaude, Path: layout.ClaudeSkillsPath, Safe: true},
		},
	}
	if _, err := ClassifyProject(desired, map[string]TreeHash{"legacy": classificationHash('1')}, inventory); err == nil || !issueCode(err, IssueInvalidMode) {
		t.Fatalf("legacy desired symlink mode error = %v, want invalid mode", err)
	}
}

func TestSelectionScopedCollisionsAndUnselectedDirectCatalog(t *testing.T) {
	registry := fixtureRegistry(t)
	registry.Profiles["dev"] = Profile{Skills: append([]string{"brainstorming"}, registry.Profiles["dev"].Skills...)}
	if err := ValidateRegistry(registry); err == nil || !issueCode(err, IssueCollision) {
		t.Fatalf("baseline/profile collision error = %v", err)
	}

	registry = fixtureRegistry(t)
	registry.Profiles["rust"] = Profile{Skills: []string{"modern-go", "modern-rust"}}
	if err := ValidateRegistry(registry); err == nil || !issueCode(err, IssueCollision) {
		t.Fatalf("profile/profile collision error = %v", err)
	}

	registry = fixtureRegistry(t)
	manifest := Manifest{Version: ManifestVersion, Profiles: []string{"go"}, Direct: []DirectSkill{{Name: "modern-go", Source: "example/modern-go"}}}
	if err := ValidateManifest(registry, manifest); err == nil || !issueCode(err, IssueCollision) {
		t.Fatalf("direct/profile collision error = %v", err)
	}

	manifest = Manifest{Version: ManifestVersion, Direct: []DirectSkill{{Name: "brainstorming", Source: "example/brainstorming"}}}
	if err := ValidateManifest(registry, manifest); err == nil || !issueCode(err, IssueCollision) {
		t.Fatalf("direct/baseline collision error = %v", err)
	}

	manifest = Manifest{Version: ManifestVersion, Direct: []DirectSkill{
		{Name: "third-party-review", Source: "example/third-party-review"},
		{Name: "third-party-review", Source: "example/third-party-review"},
	}}
	if err := ValidateManifestShape(manifest); err == nil || !issueCode(err, IssueDuplicate) {
		t.Fatalf("duplicate direct error = %v", err)
	}

	manifest = Manifest{Version: ManifestVersion, Direct: []DirectSkill{{Name: "impeccable", Source: "example/impeccable"}}}
	if err := ValidateManifest(registry, manifest); err != nil {
		t.Fatalf("unselected catalog direct should be allowed: %v", err)
	}
	state, err := ResolveProject(registry, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 1 || state.Skills[0].Manager != ManagerSkillsCLI || state.Skills[0].Mode != ModeCopy || len(state.Skills[0].Targets) != 2 {
		t.Fatalf("unselected direct resolution = %#v", state.Skills)
	}
}

func TestPlanContractFreezesActionsAndPayloadShape(t *testing.T) {
	wantActions := map[PlanAction]string{
		PlanActionInstall:    "install",
		PlanActionUpdate:     "update",
		PlanActionQuarantine: "quarantine",
		PlanActionUnchanged:  "unchanged",
		PlanActionManual:     "manual",
		PlanActionWorkflow:   "workflow",
		PlanActionBlocked:    "blocked",
	}
	for action, want := range wantActions {
		if string(action) != want {
			t.Fatalf("plan action %q = %q, want %q", action, action, want)
		}
	}
	plan := Plan{
		Desired:    DesiredState{Scope: ScopeProject, Skills: []DesiredSkill{}},
		Operations: []PlanOperation{},
	}
	envelope := Envelope{
		Operation: CommandOperationPlan,
		Result:    ResultSuccess,
		Warnings:  []Warning{},
		Evidence:  []Evidence{},
		Plan:      &plan,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(data, &shape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"operation", "result", "error", "warnings", "evidence", "plan"} {
		if _, ok := shape[field]; !ok {
			t.Fatalf("envelope missing %q: %s", field, data)
		}
	}
	var planShape map[string]json.RawMessage
	if err := json.Unmarshal(shape["plan"], &planShape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"desired", "operations"} {
		if _, ok := planShape[field]; !ok {
			t.Fatalf("plan missing %q: %s", field, shape["plan"])
		}
	}
	operation := PlanOperation{
		Action: PlanActionInstall, Skill: "example", Target: TargetAgents,
		Manager: ManagerSkillsCLI, SourceID: "catalog", Source: "example/catalog",
		Reason: "missing", Current: PlanEvidence{Kind: "inventory", Detail: "absent"},
		Expected: PlanEvidence{Kind: "desired", Detail: "source example/catalog"},
	}
	operationJSON, err := json.Marshal(operation)
	if err != nil {
		t.Fatal(err)
	}
	var operationShape map[string]json.RawMessage
	if err := json.Unmarshal(operationJSON, &operationShape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"action", "skill", "target", "manager", "sourceId", "source", "reason", "current", "expected"} {
		if _, ok := operationShape[field]; !ok {
			t.Fatalf("operation missing %q: %s", field, operationJSON)
		}
	}
}
