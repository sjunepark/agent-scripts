package sjskills

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLocalSourceSpellingAndIdentity(t *testing.T) {
	valid := map[string]string{
		"./skills/team-tool":   "local:./skills/team-tool",
		"./skills//team-tool/": "local:./skills/team-tool",
		"../shared/team-tool":  "local:../shared/team-tool",
		"/opt/skills/a/../b":   "local:/opt/skills/b",
		"c:/skills/team-tool":  "local:C:/skills/team-tool",
	}
	for source, want := range valid {
		if problem := LocalSourceProblem(source); problem != "" {
			t.Fatalf("LocalSourceProblem(%q) = %q", source, problem)
		}
		identity, ok := canonicalProjectSourceIdentity(source)
		if !ok || identity != want || !isCanonicalProjectSourceIdentity(identity) {
			t.Fatalf("identity(%q) = %q %v, want canonical %q", source, identity, ok, want)
		}
	}
	for _, source := range []string{"~/skills/x", `.\skills\x`, "C:\\skills\\x", "C:skills", ".", "./", "./skills/..", "skills/x/."} {
		if IsLocalSource(source) && LocalSourceProblem(source) == "" {
			t.Fatalf("LocalSourceProblem(%q) accepted an invalid local source", source)
		}
		if _, ok := canonicalProjectSourceIdentity(source); ok && IsLocalSource(source) {
			t.Fatalf("invalid local source %q received an identity", source)
		}
	}
	for _, identity := range []string{"local:skills/x", "local:./", "local:./a/../b", "local:c:/x", "local:~/x"} {
		if isCanonicalProjectSourceIdentity(identity) {
			t.Fatalf("non-canonical identity %q was accepted", identity)
		}
	}
	if IsLocalSource("owner/repo/skills/x") || IsLocalSource("https://github.com/o/r") {
		t.Fatal("remote sources were classified as local")
	}
}

func TestManifestLocalSourceValidation(t *testing.T) {
	manifest, err := ParseManifest([]byte("version = 1\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n"))
	if err != nil || manifest.Direct[0].Source != "./skills/team-tool" {
		t.Fatalf("local manifest = %#v err=%v", manifest, err)
	}
	for _, extra := range []string{"access = \"github-authenticated\"\n", "full_depth = true\n"} {
		_, err := ParseManifest([]byte("version = 1\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n" + extra))
		if err == nil || !issueCode(err, IssueInvalidSource) {
			t.Fatalf("local source with %q error = %v, want invalid source", extra, err)
		}
	}
	if err := ValidateRegistry(func() Registry {
		registry := fixtureRegistry(t)
		for id, source := range registry.Sources {
			if source.Kind == SourceRepository {
				source.Location = "./skills"
				registry.Sources[id] = source
			}
		}
		return registry
	}()); err == nil || !issueCode(err, IssueInvalidSource) {
		t.Fatalf("registry local source error = %v, want invalid source", err)
	}
}

func TestResolveLocalSourceAgainstProjectRoot(t *testing.T) {
	root := t.TempDir()
	registry := fixtureRegistry(t)
	manifest := Manifest{Version: 1, Direct: []DirectSkill{{Name: "team-tool", Source: "./skills/team-tool"}}}
	plan, err := BuildPlan(ResolveRequest{Registry: registry, Manifest: &manifest, ProjectRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	skill := plan.Desired.Skills[0]
	if skill.LocalPath != filepath.Join(root, "skills", "team-tool") || skill.Source != "./skills/team-tool" || skill.Manager != ManagerSkillsCLI || skill.Mode != ModeCopy {
		t.Fatalf("resolved local skill = %#v", skill)
	}
	if len(plan.Warnings) != 0 {
		t.Fatalf("portable source warnings = %#v", plan.Warnings)
	}
	if _, err := ResolveProject(registry, manifest, ""); err == nil || !issueCode(err, IssueInvalidSource) {
		t.Fatalf("missing root error = %v", err)
	}
	outside := Manifest{Version: 1, Direct: []DirectSkill{{Name: "team-tool", Source: "../shared/team-tool"}}}
	plan, err = BuildPlan(ResolveRequest{Registry: registry, Manifest: &outside, ProjectRoot: root})
	if err != nil || len(plan.Warnings) != 1 || plan.Warnings[0].Code != "machine-specific-source" {
		t.Fatalf("outside source plan warnings = %#v err=%v", plan.Warnings, err)
	}
	containing := Manifest{Version: 1, Direct: []DirectSkill{{Name: "team-tool", Source: filepath.ToSlash(filepath.Dir(root))}}}
	if _, err := ResolveProject(registry, containing, root); err == nil || !strings.Contains(err.Error(), "must not contain the project root") {
		t.Fatalf("containing source error = %v", err)
	}
}

func writeLocalSkill(t *testing.T, path, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(path, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: \"Fixture\"\n---\n# Fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "scripts", "run.sh"), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func localDesiredSkill(path string) DesiredSkill {
	return DesiredSkill{Access: AccessPublic, Name: "team-tool", Source: "./skills/team-tool", LocalPath: path, Scope: ScopeProject,
		Origin: "direct", Manager: ManagerSkillsCLI, Mode: ModeCopy, Targets: defaultTargets()}
}

func TestMaterializeLocalSourceWithoutSkillsCLI(t *testing.T) {
	source := filepath.Join(t.TempDir(), "skills", "team-tool")
	writeLocalSkill(t, source, "team-tool")
	var calls atomic.Int32
	materializer := NewMaterializer(MaterializerConfig{Runner: runnerFunc(func(context.Context, string, []string, []string) (ProcessResult, error) {
		calls.Add(1)
		return ProcessResult{}, nil
	})})
	plan, err := materializer.Materialize(context.Background(), []DesiredSkill{localDesiredSkill(source)})
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Cleanup()
	if calls.Load() != 0 {
		t.Fatalf("local-only materialization ran %d processes", calls.Load())
	}
	snapshot, ok := plan.SnapshotFor("team-tool")
	if !ok {
		t.Fatal("missing local snapshot")
	}
	want, err := HashSkillTree(source)
	if err != nil || snapshot.Hash != want {
		t.Fatalf("snapshot hash = %#v, source hash = %#v err=%v", snapshot.Hash, want, err)
	}
	if err := plan.Verify(); err != nil {
		t.Fatal(err)
	}
	if err := validateApplyCopyTree(snapshot.Path); err != nil {
		t.Fatalf("staged copy violates the placement contract: %v", err)
	}
}

func TestMaterializeLocalSourceBlocksInvalidSources(t *testing.T) {
	cases := map[string]struct {
		prepare func(t *testing.T, path string) string
		want    string
	}{
		"missing": {func(t *testing.T, path string) string { return path }, "does not exist; restore it or remove its [[direct]] entry"},
		"no skill file": {func(t *testing.T, path string) string {
			writeLocalSkill(t, path, "team-tool")
			if err := os.Remove(filepath.Join(path, "SKILL.md")); err != nil {
				t.Fatal(err)
			}
			return path
		}, "SKILL.md is missing"},
		"name mismatch": {func(t *testing.T, path string) string {
			writeLocalSkill(t, path, "other-tool")
			return path
		}, `declares SKILL.md name "other-tool", not "team-tool"`},
		"managed root": {func(t *testing.T, path string) string {
			managed := filepath.Join(filepath.Dir(path), ".claude", "skills", "team-tool")
			writeLocalSkill(t, managed, "team-tool")
			return managed
		}, "is inside a generated"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			path := test.prepare(t, filepath.Join(t.TempDir(), "team-tool"))
			plan, err := NewMaterializer(MaterializerConfig{Runner: runnerFunc(func(context.Context, string, []string, []string) (ProcessResult, error) {
				t.Fatal("local source invoked a process")
				return ProcessResult{}, nil
			})}).Materialize(context.Background(), []DesiredSkill{localDesiredSkill(path)})
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), "./skills/team-tool") {
				t.Fatalf("plan=%v err=%v, want %q", plan, err, test.want)
			}
		})
	}
}

func TestMaterializeLocalSourceRejectsSymlinks(t *testing.T) {
	source := filepath.Join(t.TempDir(), "team-tool")
	writeLocalSkill(t, source, "team-tool")
	if err := os.Symlink("SKILL.md", filepath.Join(source, "alias.md")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	_, err := NewMaterializer(MaterializerConfig{}).Materialize(context.Background(), []DesiredSkill{localDesiredSkill(source)})
	if err == nil || !strings.Contains(err.Error(), "alias.md is a symlink") {
		t.Fatalf("symlink error = %v", err)
	}
}

func TestStatusHashesLocalSourcesOnEveryCheck(t *testing.T) {
	root := canonicalTempHome(t)
	writeLocalSkill(t, filepath.Join(root, "skills", "team-tool"), "team-tool")
	if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte("version = 1\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scope, err := ResolveStatusScope(root, minimalGlobalRegistry(t), false)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 6, 1, 0, 0, 0, time.UTC)
	service := StatusService{CacheRoot: filepath.Join(t.TempDir(), "status"), Now: func() time.Time { return now }}
	service.Refresh = func(_ context.Context, desired []DesiredSkill) (StatusSnapshot, error) {
		if len(desired) != 0 {
			t.Fatalf("local sources reached the upstream refresh: %#v", desired)
		}
		return StatusSnapshot{map[string]TreeHash{}, now}, nil
	}
	advisory := service.Check(context.Background(), scope, nil)
	if advisory.Error != "" || len(advisory.Findings) != 2 {
		t.Fatalf("advisory = %+v", advisory)
	}
	requireStatusFinding(t, advisory, AdvisoryMissing, "team-tool", TargetAgents, string(ProjectStateReasonExpectedEntryAbsent))

	if err := os.RemoveAll(filepath.Join(root, "skills", "team-tool")); err != nil {
		t.Fatal(err)
	}
	advisory = service.Check(context.Background(), scope, nil)
	if advisory.Freshness != AdvisoryUnavailable || !strings.Contains(advisory.Error, "does not exist") || validateAdvisories([]Advisory{advisory}) != nil {
		t.Fatalf("missing-source advisory = %+v", advisory)
	}
}

type runnerFunc func(context.Context, string, []string, []string) (ProcessResult, error)

func (f runnerFunc) Run(ctx context.Context, command string, args, env []string) (ProcessResult, error) {
	return f(ctx, command, args, env)
}

func TestLocalSourceRejectsParentOnlyAndProjectRoots(t *testing.T) {
	for _, source := range []string{"../", "../..", "./skills/../..", "/"} {
		if LocalSourceProblem(source) == "" {
			t.Fatalf("LocalSourceProblem(%q) accepted a parent-only path", source)
		}
	}
	project := filepath.Join(t.TempDir(), "nested")
	writeLocalSkill(t, project, "team-tool")
	if err := os.WriteFile(filepath.Join(project, ManifestFileName), []byte("version = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := inspectLocalSkillSource(localDesiredSkill(project), MaterializerLimits{})
	if err == nil || !strings.Contains(err.Error(), "sjskills.toml is a project manifest") {
		t.Fatalf("project-root source error = %v", err)
	}
	if err := os.Remove(filepath.Join(project, ManifestFileName)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(project, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err = inspectLocalSkillSource(localDesiredSkill(project), MaterializerLimits{})
	if err == nil || !strings.Contains(err.Error(), ".claude/skills is a generated") {
		t.Fatalf("source containing a managed root error = %v", err)
	}
}

func TestLocalSkillDeclaredNameToleratesYAMLDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(path, []byte("---\r\nname: team-tool # owner: platform\r\n---  \r\nname: body\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if name, err := localSkillDeclaredName(path); err != nil || name != "team-tool" {
		t.Fatalf("declared name = %q err=%v", name, err)
	}
}

func TestMaterializeMixedLocalAndRemoteSources(t *testing.T) {
	source := filepath.Join(t.TempDir(), "team-tool")
	writeLocalSkill(t, source, "team-tool")
	runner := defaultMaterializeRunner()
	materializer, _ := testMaterializer(t, runner, MaterializerLimits{})
	plan, err := materializer.Materialize(context.Background(), []DesiredSkill{localDesiredSkill(source), desiredMaterializeSkill("remote-tool", "example/remote-catalog")})
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Cleanup()
	if len(plan.Snapshots()) != 2 {
		t.Fatalf("snapshots = %d, want 2", len(plan.Snapshots()))
	}
	for _, call := range runner.calls {
		for _, arg := range call.args {
			if strings.Contains(arg, "team-tool") {
				t.Fatalf("local skill reached Skills CLI: %q", call.args)
			}
		}
	}
	if len(runner.calls) != 3 {
		t.Fatalf("calls = %d, want preflight twice and one add", len(runner.calls))
	}
}

func TestStatusReportsLocalSourceEditAsUpdate(t *testing.T) {
	root := canonicalTempHome(t)
	source := filepath.Join(root, "skills", "team-tool")
	writeLocalSkill(t, source, "team-tool")
	if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte("version = 1\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scope, err := ResolveStatusScope(root, minimalGlobalRegistry(t), false)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := LayoutForProject(scope.Root)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := NewMaterializer(MaterializerConfig{}).Materialize(context.Background(), scope.Plan.Desired.Skills)
	if err != nil {
		t.Fatal(err)
	}
	defer plan.Cleanup()
	session := &ProjectApplySession{Layout: layout, Desired: scope.Plan.Desired, Expected: map[string]TreeHash{"team-tool": plan.snapshots["team-tool"].Hash}, Materialized: plan}
	if session.Plan, err = InspectStatusPlan(scope, session.Expected); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyProjectChanges(context.Background(), session, ApplyDeps{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("---\nname: team-tool\n---\n# edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 6, 1, 0, 0, 0, time.UTC)
	service := StatusService{CacheRoot: filepath.Join(t.TempDir(), "status"), Now: func() time.Time { return now },
		Refresh: func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
			return StatusSnapshot{map[string]TreeHash{}, now}, nil
		}}
	advisory := service.Check(context.Background(), scope, nil)
	requireStatusFinding(t, advisory, AdvisoryUpdate, "team-tool", TargetClaude, string(ProjectStateReasonVerifiedUpdate))
}

func TestLocalSkillDeclaredNameStripsCommentAfterQuotedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	for content, want := range map[string]string{
		"---\nname: \"team-tool\" # maintained locally\n---\n": "team-tool",
		"---\nname: 'team-tool'\n---\n":                        "team-tool",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if name, err := localSkillDeclaredName(path); err != nil || name != want {
			t.Fatalf("%q: name = %q err=%v", content, name, err)
		}
	}
}

func TestMaterializeReportsBrokenLocalSourceBeforeRemoteTooling(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "team-tool")
	runner := &materializeRunner{invoke: func(context.Context, string, []string, []string) (ProcessResult, error) {
		return ProcessResult{}, errors.New("bunx unavailable")
	}}
	materializer, _ := testMaterializer(t, runner, MaterializerLimits{})
	_, err := materializer.Materialize(context.Background(), []DesiredSkill{localDesiredSkill(missing), desiredMaterializeSkill("remote-tool", "example/remote-catalog")})
	var localErr *LocalSourceError
	if !errors.As(err, &localErr) || len(runner.calls) != 0 {
		t.Fatalf("error = %v, calls = %d; want a local source error before any process", err, len(runner.calls))
	}
}

func TestLocalOnlyStatusDoesNotDependOnUpstreamCache(t *testing.T) {
	root := canonicalTempHome(t)
	writeLocalSkill(t, filepath.Join(root, "skills", "team-tool"), "team-tool")
	if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte("version = 1\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scope, err := ResolveStatusScope(root, minimalGlobalRegistry(t), false)
	if err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	service := StatusService{CacheRoot: blocked, Refresh: func(context.Context, []DesiredSkill) (StatusSnapshot, error) {
		t.Fatal("local-only status refreshed upstream evidence")
		return StatusSnapshot{}, nil
	}}
	advisory := service.Check(context.Background(), scope, nil)
	if advisory.Freshness != AdvisoryFresh || advisory.Error != "" || validateAdvisories([]Advisory{advisory}) != nil {
		t.Fatalf("advisory = %+v", advisory)
	}
	requireStatusFinding(t, advisory, AdvisoryMissing, "team-tool", TargetAgents, string(ProjectStateReasonExpectedEntryAbsent))
}

func TestRestoreRejectsLocalIdentityInGlobalScope(t *testing.T) {
	if !isCanonicalSourceIdentityForScope(ScopeProject, "local:./skills/team-tool") ||
		isCanonicalSourceIdentityForScope(ScopeGlobal, "local:./skills/team-tool") ||
		!isCanonicalSourceIdentityForScope(ScopeGlobal, "github:owner/repo") {
		t.Fatal("scope-aware identity validation is wrong")
	}
	entry := restoreEntry{
		entry:   ProjectQuarantineManifestEntry{Skill: "team-tool", Target: TargetAgents, Action: ProjectQuarantineEntryActionRemove, OldSourceIdentity: "local:./skills/team-tool"},
		oldHash: classificationHash('a'),
	}
	if _, err := buildRestoreProvenanceState(ProvenanceState{}, []restoreEntry{entry}, time.Now(), ScopeGlobal); err == nil {
		t.Fatal("global restore accepted a local source identity")
	}
	if _, err := buildRestoreProvenanceState(ProvenanceState{}, []restoreEntry{entry}, time.Now(), ScopeProject); err != nil {
		t.Fatalf("project restore rejected a local source identity: %v", err)
	}
}
