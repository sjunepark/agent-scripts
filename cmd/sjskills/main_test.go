package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

var testBinary string

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "sjskills-cli-test-")
	if err != nil {
		panic(err)
	}
	testBinary = filepath.Join(directory, "sjskills")
	build := exec.Command("go", "build", "-o", testBinary, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic(fmt.Sprintf("build test binary: %v", err))
	}
	code := m.Run()
	_ = os.RemoveAll(directory)
	os.Exit(code)
}

func runCLI(t *testing.T, directory string, args ...string) (int, string, string) {
	t.Helper()
	return runCLIWithEnvironment(t, directory, nil, args...)
}

func runCLIWithEnvironment(t *testing.T, directory string, overrides map[string]string, args ...string) (int, string, string) {
	t.Helper()
	fakeDirectory := t.TempDir()
	fakeBunx := filepath.Join(fakeDirectory, "bunx")
	if err := os.WriteFile(fakeBunx, []byte(fakeBunxScript), 0o755); err != nil {
		t.Fatalf("write fake bunx: %v", err)
	}
	sentinelHome := t.TempDir()
	sentinelUserProfile := t.TempDir()
	environment := append([]string(nil), os.Environ()...)
	setEnvironmentValue(&environment, "HOME", sentinelHome)
	setEnvironmentValue(&environment, "USERPROFILE", sentinelUserProfile)
	pathValue := os.Getenv("PATH")
	setEnvironmentValue(&environment, "PATH", fakeDirectory+string(os.PathListSeparator)+pathValue)
	for key, value := range overrides {
		setEnvironmentValue(&environment, key, value)
	}
	command := exec.Command(testBinary, args...)
	command.Dir = directory
	command.Env = environment
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), stdout.String(), stderr.String()
	}
	t.Fatalf("run %v: %v", args, err)
	return -1, "", ""
}

func setEnvironmentValue(environment *[]string, key, value string) {
	for index, item := range *environment {
		name, _, ok := strings.Cut(item, "=")
		if ok && name == key {
			(*environment)[index] = key + "=" + value
			return
		}
	}
	*environment = append(*environment, key+"="+value)
}

const fakeBunxScript = `#!/bin/sh
set -eu

if [ "$#" -eq 1 ] && [ "$1" = "--version" ]; then
  printf 'bunx 1\n'
  exit 0
fi
if [ "$#" -eq 2 ] && [ "$1" = "skills@1.5.23" ] && [ "$2" = "--version" ]; then
  printf '1.5.23\n'
  exit 0
fi
if [ "$#" -ge 2 ] && [ "$1" = "skills@1.5.23" ] && [ "$2" = "add" ]; then
  skill=""
  previous=""
  for argument in "$@"; do
    if [ "$previous" = "--skill" ]; then
      skill="$argument"
      break
    fi
    previous="$argument"
  done
  if [ -z "$skill" ]; then
    exit 3
  fi
  target="$CODEX_HOME/skills/$skill"
  mkdir -p "$target"
  printf '# %s\n' "$skill" > "$target/SKILL.md"
  exit 0
fi
exit 4
`

func TestExternalHelpAndVersion(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--help")
	if code != 0 || !strings.Contains(stdout, "sjskills <command>") || stderr != "" {
		t.Fatalf("help code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--version")
	if code != 0 || stdout != "sjskills 1.0.0\n" || stderr != "" {
		t.Fatalf("version code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalJSONEnvelopeAndMalformedInvocation(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--json", "profiles")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("profiles code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var envelope sjskillsEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("profiles JSON: %v", err)
	}
	if envelope.Operation != "profiles" || envelope.Result != "success" {
		t.Fatalf("profiles envelope = %#v", envelope)
	}
	var profileShape map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &profileShape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"operation", "result", "error", "warnings", "evidence"} {
		if _, ok := profileShape[field]; !ok {
			t.Fatalf("profiles JSON missing stable field %q: %q", field, stdout)
		}
	}

	code, stdout, stderr = runCLI(t, directory, "--json", "unknown")
	if code != 64 || stderr != "" || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("malformed code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("malformed JSON: %v", err)
	}
	if envelope.Operation != "parse" || envelope.Result != "invalid" {
		t.Fatalf("malformed envelope = %#v", envelope)
	}

	code, stdout, stderr = runCLI(t, directory, "plan", "--dry")
	if code != 64 || stdout != "" || !strings.Contains(stderr, "unknown flag --dry") {
		t.Fatalf("unsupported flag code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanApplyAndRestoreContracts(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"dev\", \"go\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, directory, "--json", "plan")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"operation":"plan"`) || !strings.Contains(stdout, `"plan":{"desired"`) || !strings.Contains(stdout, `"operations":[]`) || !strings.Contains(stdout, `"resolved 26 desired skills"`) {
		t.Fatalf("plan output = %q", stdout)
	}

	code, stdout, stderr = runCLI(t, directory, "--json", "apply", "--global")
	if code != 2 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"unavailable"`) || !strings.Contains(stdout, "managed-root reconciliation is not implemented") || strings.Contains(stdout, "materialization and managed-root mutation are not implemented") {
		t.Fatalf("apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "restore", "quarantine-1")
	if code != 2 || stderr != "" || !strings.Contains(stdout, `"result":"unavailable"`) {
		t.Fatalf("restore code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "restore")
	if code != 65 || stderr != "" || !strings.Contains(stdout, `"result":"invalid"`) {
		t.Fatalf("restore missing id code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanMaterializesExpectedContentReadOnly(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"dev\", \"go\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	projectAgents := filepath.Join(directory, ".agents", "skills")
	projectClaude := filepath.Join(directory, ".claude", "skills")
	for _, root := range []string{projectAgents, projectClaude} {
		if err := os.MkdirAll(filepath.Join(root, "existing"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "existing", "SKILL.md"), []byte("project fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projectAgentsBefore := captureFixtureTree(t, projectAgents)
	projectClaudeBefore := captureFixtureTree(t, projectClaude)

	globalHome := t.TempDir()
	globalUserProfile := t.TempDir()
	globalAgents := filepath.Join(t.TempDir(), "agents")
	globalClaude := filepath.Join(t.TempDir(), "claude")
	for _, root := range []string{globalAgents, globalClaude} {
		if err := os.MkdirAll(root, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "fixture.txt"), []byte("global fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	globalHomeBefore := captureFixtureTree(t, globalHome)
	globalUserProfileBefore := captureFixtureTree(t, globalUserProfile)
	globalAgentsBefore := captureFixtureTree(t, globalAgents)
	globalClaudeBefore := captureFixtureTree(t, globalClaude)
	overrides := map[string]string{
		"HOME":              globalHome,
		"USERPROFILE":       globalUserProfile,
		"CODEX_HOME":        globalAgents,
		"CLAUDE_CONFIG_DIR": globalClaude,
	}

	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "plan")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("project plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	projectEnvelope := decodeEnvelope(t, stdout)
	assertExpectedContentEvidence(t, projectEnvelope)
	if len(projectEnvelope.Warnings) == 0 || projectEnvelope.Warnings[0].Code != "manual-action" {
		t.Fatalf("project warnings = %#v, want authoritative manual warning", projectEnvelope.Warnings)
	}
	for _, output := range []string{stdout, stderr} {
		if strings.Contains(output, "sjskills-materialize-") || strings.Contains(output, string(filepath.Separator)+".agents"+string(filepath.Separator)+"skills"+string(filepath.Separator)) {
			t.Fatalf("materialization path leaked in output: %q", output)
		}
	}
	if got := captureFixtureTree(t, projectAgents); !reflect.DeepEqual(got, projectAgentsBefore) {
		t.Fatalf("project .agents root changed: before=%#v after=%#v", projectAgentsBefore, got)
	}
	if got := captureFixtureTree(t, projectClaude); !reflect.DeepEqual(got, projectClaudeBefore) {
		t.Fatalf("project .claude root changed: before=%#v after=%#v", projectClaudeBefore, got)
	}
	for path, before := range map[string]fixtureTree{
		"HOME": globalHomeBefore, "USERPROFILE": globalUserProfileBefore,
		"CODEX_HOME": globalAgentsBefore, "CLAUDE_CONFIG_DIR": globalClaudeBefore,
	} {
		var root string
		switch path {
		case "HOME":
			root = globalHome
		case "USERPROFILE":
			root = globalUserProfile
		case "CODEX_HOME":
			root = globalAgents
		case "CLAUDE_CONFIG_DIR":
			root = globalClaude
		}
		if got := captureFixtureTree(t, root); !reflect.DeepEqual(got, before) {
			t.Fatalf("sentinel %s changed: before=%#v after=%#v", path, before, got)
		}
	}

	code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, "--json", "plan", "--global")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("global plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	globalEnvelope := decodeEnvelope(t, stdout)
	assertExpectedContentEvidence(t, globalEnvelope)
	for _, output := range []string{stdout, stderr} {
		if strings.Contains(output, "sjskills-materialize-") {
			t.Fatalf("global materialization path leaked in output: %q", output)
		}
	}
	for path, before := range map[string]fixtureTree{
		"HOME": globalHomeBefore, "USERPROFILE": globalUserProfileBefore,
		"CODEX_HOME": globalAgentsBefore, "CLAUDE_CONFIG_DIR": globalClaudeBefore,
	} {
		var root string
		switch path {
		case "HOME":
			root = globalHome
		case "USERPROFILE":
			root = globalUserProfile
		case "CODEX_HOME":
			root = globalAgents
		case "CLAUDE_CONFIG_DIR":
			root = globalClaude
		}
		if got := captureFixtureTree(t, root); !reflect.DeepEqual(got, before) {
			t.Fatalf("global sentinel %s changed: before=%#v after=%#v", path, before, got)
		}
	}
}

func TestApplicationMaterializationFailuresAndLifecycle(t *testing.T) {
	registry, err := sjskills.EmbeddedRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(registry.Global.Baseline) == 0 {
		t.Fatal("embedded registry has no global baseline")
	}

	t.Run("materialize failure preserves plan", func(t *testing.T) {
		calls := 0
		app := &application{
			directory: t.TempDir(),
			materialize: func(context.Context, []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				calls++
				return nil, errors.New("materialize failed")
			},
		}
		envelope := app.plan(context.Background(), true)
		if calls != 1 || envelope.Result != sjskills.ResultUnavailable || envelope.Plan == nil || envelope.Error == nil || envelope.Error.Code != sjskills.IssueUnavailable {
			t.Fatalf("calls=%d envelope=%#v", calls, envelope)
		}
		if !strings.Contains(envelope.Error.Message, "materialize failed") {
			t.Fatalf("error = %#v", envelope.Error)
		}
	})

	t.Run("partial materialize failure preserves error and cleans up", func(t *testing.T) {
		materializer, parent := testInjectedMaterializer(t)
		var stagedRoot string
		const materializeMessage = "materialize failed after staging"
		app := &application{
			directory: t.TempDir(),
			materialize: func(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				materialized, err := materializer.Materialize(ctx, skills)
				if err != nil {
					return nil, err
				}
				stagedRoot = materialized.Root()
				return materialized, errors.New(materializeMessage)
			},
		}
		envelope := app.plan(context.Background(), true)
		if envelope.Result != sjskills.ResultUnavailable || envelope.Error == nil || envelope.Error.Code != sjskills.IssueUnavailable || envelope.Error.Message != materializeMessage {
			t.Fatalf("envelope=%#v", envelope)
		}
		if _, err := os.Stat(stagedRoot); !os.IsNotExist(err) {
			t.Fatalf("staging root still exists after partial materialize failure: %q err=%v", stagedRoot, err)
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after partial materialize failure: entries=%d err=%v", len(entries), err)
		}
	})

	t.Run("verify failure cleans up before return", func(t *testing.T) {
		materializer, parent := testInjectedMaterializer(t)
		var stagedRoot string
		app := &application{
			directory: t.TempDir(),
			materialize: func(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				materialized, err := materializer.Materialize(ctx, skills)
				if err != nil {
					return nil, err
				}
				stagedRoot = materialized.Root()
				snapshot := materialized.Snapshots()[0]
				if err := os.WriteFile(filepath.Join(snapshot.Path, "SKILL.md"), []byte("tampered\n"), 0o644); err != nil {
					t.Fatalf("tamper staged snapshot: %v", err)
				}
				return materialized, nil
			},
		}
		envelope := app.plan(context.Background(), true)
		if envelope.Result != sjskills.ResultUnavailable || envelope.Error == nil || envelope.Error.Code != sjskills.IssueUnavailable {
			t.Fatalf("envelope=%#v", envelope)
		}
		if stagedRoot == "" {
			t.Fatal("materializer did not return a staging root")
		}
		if _, err := os.Stat(stagedRoot); !os.IsNotExist(err) {
			t.Fatalf("staging root still exists after verify failure: %q err=%v", stagedRoot, err)
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after verify failure: entries=%d err=%v", len(entries), err)
		}
	})

	t.Run("successful plan materializes once and cleans up", func(t *testing.T) {
		materializer, parent := testInjectedMaterializer(t)
		calls := 0
		var stagedRoot string
		app := &application{
			directory: t.TempDir(),
			materialize: func(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				calls++
				materialized, err := materializer.Materialize(ctx, skills)
				if err == nil {
					stagedRoot = materialized.Root()
				}
				return materialized, err
			},
		}
		envelope := app.plan(context.Background(), true)
		if calls != 1 || envelope.Result != sjskills.ResultSuccess || stagedRoot == "" {
			t.Fatalf("calls=%d stagedRoot=%q envelope=%#v", calls, stagedRoot, envelope)
		}
		if _, err := os.Stat(stagedRoot); !os.IsNotExist(err) {
			t.Fatalf("staging root still exists after plan: %q err=%v", stagedRoot, err)
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after plan: entries=%d err=%v", len(entries), err)
		}

		applyEnvelope := app.apply(context.Background(), true)
		if calls != 2 || applyEnvelope.Result != sjskills.ResultUnavailable || applyEnvelope.Error == nil || !strings.Contains(applyEnvelope.Error.Message, "managed-root reconciliation") {
			t.Fatalf("calls=%d apply=%#v", calls, applyEnvelope)
		}
	})

	t.Run("non-planning operations do not materialize", func(t *testing.T) {
		calls := 0
		app := &application{directory: t.TempDir(), materialize: func(context.Context, []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
			calls++
			return nil, errors.New("unexpected materialization")
		}}
		if envelope := app.profiles(); envelope.Result != sjskills.ResultSuccess {
			t.Fatalf("profiles=%#v", envelope)
		}
		if envelope := app.init([]string{"dev"}); envelope.Result != sjskills.ResultSuccess {
			t.Fatalf("init=%#v", envelope)
		}
		if envelope := app.restore("quarantine-1"); envelope.Result != sjskills.ResultUnavailable {
			t.Fatalf("restore=%#v", envelope)
		}
		if calls != 0 {
			t.Fatalf("materialize calls=%d, want zero", calls)
		}
	})
}

func TestLifecycleErrorShapes(t *testing.T) {
	sanitizedPrimary := errors.New("sanitized primary diagnostic")
	tests := []struct {
		name            string
		stage           string
		primary         error
		cleanup         error
		want            string
		preservePrimary bool
	}{
		{
			name:            "primary with successful cleanup preserves sanitized error",
			stage:           "verify",
			primary:         sanitizedPrimary,
			want:            "sanitized primary diagnostic",
			preservePrimary: true,
		},
		{
			name:    "cleanup-only failure is concise and generic",
			stage:   "cleanup",
			cleanup: errors.New("cleanup path /tmp/private-stage"),
			want:    "materialization cleanup failed",
		},
		{
			name:    "primary and cleanup failure hide both causes",
			stage:   "verify",
			primary: errors.New("primary secret /private/source"),
			cleanup: errors.New("cleanup path /tmp/private-stage"),
			want:    "materialization verify failed and cleanup failed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := lifecycleError(test.stage, test.primary, test.cleanup)
			if got == nil || got.Error() != test.want {
				t.Fatalf("lifecycleError() = %v, want %q", got, test.want)
			}
			if test.preservePrimary && got != test.primary {
				t.Fatalf("lifecycleError() returned a different primary error: got=%p want=%p", got, test.primary)
			}
			if strings.Contains(got.Error(), "private") || strings.Contains(got.Error(), "secret") {
				t.Fatalf("lifecycleError() leaked a cause: %q", got.Error())
			}
		})
	}
}

func TestExternalPlanDiscoversNearestNestedProject(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "sjskills.toml")
	if err := os.WriteFile(manifestPath, []byte("version = 1\nprofiles = [\"kicpa\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	canonicalManifest, err := filepath.EvalSymlinks(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, nested, "--json", "plan")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"path":"`+canonicalManifest+`"`) {
		t.Fatalf("nested plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanRejectsUnsupportedDirectFields(t *testing.T) {
	directory := t.TempDir()
	manifest := "version = 1\n\n[[direct]]\nname = \"third-party-review\"\nsource = \"example/third-party-review\"\nmanager = \"skills-cli\"\n"
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, directory, "--json", "plan")
	if code != 65 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"invalid"`) || !strings.Contains(stdout, `"code":"unknown_field"`) {
		t.Fatalf("unsupported direct field code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalInitRefusesOverwrite(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--json", "init", "dev", "go")
	if code != 0 || stderr != "" || !strings.Contains(stdout, `"operation":"init"`) {
		t.Fatalf("init code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	path := filepath.Join(directory, "sjskills.toml")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "init", "kicpa")
	if code != 2 || stderr != "" || !strings.Contains(stdout, `"result":"conflict"`) {
		t.Fatalf("second init code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("existing manifest was changed: before=%q after=%q", before, after)
	}
}

type fixtureTree struct {
	Exists bool
	Files  map[string][]byte
}

func captureFixtureTree(t *testing.T, root string) fixtureTree {
	t.Helper()
	tree := fixtureTree{Files: map[string][]byte{}}
	if _, err := os.Lstat(root); err != nil {
		if os.IsNotExist(err) {
			return tree
		}
		t.Fatalf("stat fixture root %q: %v", root, err)
	}
	tree.Exists = true
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		tree.Files[filepath.ToSlash(relative)] = append([]byte(nil), data...)
		return nil
	}); err != nil {
		t.Fatalf("capture fixture root %q: %v", root, err)
	}
	return tree
}

func decodeEnvelope(t *testing.T, stdout string) sjskillsEnvelope {
	t.Helper()
	var envelope sjskillsEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode envelope: %v; stdout=%q", err, stdout)
	}
	return envelope
}

func assertExpectedContentEvidence(t *testing.T, envelope sjskillsEnvelope) {
	t.Helper()
	if envelope.Operation != "plan" || envelope.Result != "success" || envelope.Plan == nil {
		t.Fatalf("envelope = %#v", envelope)
	}
	wantNames := make([]string, 0, len(envelope.Plan.Desired.Skills))
	for _, skill := range envelope.Plan.Desired.Skills {
		if skill.Manager == "skills-cli" {
			wantNames = append(wantNames, skill.Name)
		}
	}
	sort.Strings(wantNames)
	gotDetails := make([]string, 0, len(wantNames))
	var summaries []string
	for _, evidence := range envelope.Evidence {
		switch evidence.Kind {
		case "expected-content":
			gotDetails = append(gotDetails, evidence.Detail)
		case "materialization":
			summaries = append(summaries, evidence.Detail)
		}
	}
	if len(gotDetails) != len(wantNames) {
		t.Fatalf("expected evidence count=%d, want %d: %#v", len(gotDetails), len(wantNames), gotDetails)
	}
	for index, name := range wantNames {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		hash, err := sjskills.HashSkillTree(root)
		if err != nil {
			t.Fatal(err)
		}
		want := fmt.Sprintf("%s %s:%s", name, sjskills.TreeHashAlgorithmSHA256V2, hash.Digest)
		if gotDetails[index] != want {
			t.Errorf("expected evidence[%d]=%q, want %q", index, gotDetails[index], want)
		}
	}
	wantSummary := fmt.Sprintf("skills@%s verified %d snapshots; temporary cleanup successful", sjskills.SkillsCLIVersion, len(wantNames))
	if len(summaries) != 1 || summaries[0] != wantSummary {
		t.Fatalf("materialization summaries=%#v, want [%q]", summaries, wantSummary)
	}
}

type injectedMaterializeRunner struct{}

func (injectedMaterializeRunner) Run(_ context.Context, _ string, args []string, environment []string) (sjskills.ProcessResult, error) {
	if len(args) == 1 && args[0] == "--version" {
		return sjskills.ProcessResult{Stdout: []byte("bunx 1\n")}, nil
	}
	if len(args) == 2 && args[0] == "skills@"+sjskills.SkillsCLIVersion && args[1] == "--version" {
		return sjskills.ProcessResult{Stdout: []byte(sjskills.SkillsCLIVersion + "\n")}, nil
	}
	if len(args) < 2 || args[0] != "skills@"+sjskills.SkillsCLIVersion || args[1] != "add" {
		return sjskills.ProcessResult{}, fmt.Errorf("unexpected fake runner argv: %q", args)
	}
	name := ""
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "--skill" {
			name = args[index+1]
			break
		}
	}
	if name == "" {
		return sjskills.ProcessResult{}, errors.New("fake runner received no skill")
	}
	root := environmentValue(environment, "CODEX_HOME")
	path := filepath.Join(root, "skills", name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return sjskills.ProcessResult{}, err
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
		return sjskills.ProcessResult{}, err
	}
	return sjskills.ProcessResult{}, nil
}

func environmentValue(environment []string, key string) string {
	for _, item := range environment {
		name, value, ok := strings.Cut(item, "=")
		if ok && name == key {
			return value
		}
	}
	return ""
}

func testInjectedMaterializer(t *testing.T) (*sjskills.Materializer, string) {
	t.Helper()
	parent := t.TempDir()
	materializer := sjskills.NewMaterializer(sjskills.MaterializerConfig{
		Runner: injectedMaterializeRunner{},
		TempRootFactory: func() (string, error) {
			return os.MkdirTemp(parent, "stage-")
		},
		BaseEnvironment: []string{"PATH=/bin"},
	})
	return materializer, parent
}

type sjskillsEnvelope struct {
	Operation string                `json:"operation"`
	Result    string                `json:"result"`
	Error     *sjskillsIssue        `json:"error"`
	Warnings  []sjskillsWarning     `json:"warnings"`
	Evidence  []sjskillsEvidence    `json:"evidence"`
	Plan      *sjskillsPlanEnvelope `json:"plan"`
}

type sjskillsIssue struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type sjskillsWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type sjskillsEvidence struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

type sjskillsPlanEnvelope struct {
	Desired struct {
		Skills []struct {
			Name    string `json:"name"`
			Manager string `json:"manager"`
		} `json:"skills"`
	} `json:"desired"`
}
