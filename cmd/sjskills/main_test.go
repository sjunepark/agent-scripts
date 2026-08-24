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
	"time"

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
	return runCLIWithInputEnvironment(t, directory, overrides, "", args...)
}

func runCLIWithInputEnvironment(t *testing.T, directory string, overrides map[string]string, input string, args ...string) (int, string, string) {
	t.Helper()
	command, stdout, stderr := newCLICommand(t, directory, overrides, input, args...)
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

func newCLICommand(t *testing.T, directory string, overrides map[string]string, input string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
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
	command.Stdin = strings.NewReader(input)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	command.Stdout = stdout
	command.Stderr = stderr
	return command, stdout, stderr
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
  suffix="${SJSKILLS_FAKE_CONTENT-}"
  printf '# %s%s\n' "$skill" "$suffix" > "$target/SKILL.md"
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
	if !strings.Contains(stdout, `"operation":"plan"`) || !strings.Contains(stdout, `"plan":{"desired"`) || !strings.Contains(stdout, `"action":"install"`) || !strings.Contains(stdout, `"resolved 26 desired skills"`) || strings.Contains(stdout, `"warnings":null`) {
		t.Fatalf("plan output = %q", stdout)
	}

	code, stdout, stderr = runCLI(t, directory, "--json", "apply", "--global", "--yes")
	if code != 2 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"unavailable"`) || !strings.Contains(stdout, "global reconciliation is not implemented") {
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

func TestExternalProjectPlanClassifiesCopyPlacementsAndRemainsReadOnly(t *testing.T) {
	directory := t.TempDir()
	manifest := `version = 1
profiles = []

[[direct]]
name = "collision"
source = "example/collision"

[[direct]]
name = "fixture-skill"
source = "example/fixture-skill"

[[direct]]
name = "modified-skill"
source = "example/modified-skill"
`
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	discovered, err := sjskills.DiscoverProjectRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := sjskills.LayoutForProject(discovered.Root)
	if err != nil {
		t.Fatal(err)
	}
	expectedFixture := writePlanFixtureSkill(t, filepath.Join(directory, ".agents", "skills", "fixture-skill"), "# fixture-skill\n")
	oldFixture := writePlanFixtureSkill(t, filepath.Join(directory, ".claude", "skills", "fixture-skill"), "# old fixture\n")
	oldModifiedRoot := filepath.Join(t.TempDir(), "modified-skill")
	oldModified := writePlanFixtureSkill(t, oldModifiedRoot, "# old modified\n")
	writePlanFixtureSkill(t, filepath.Join(directory, ".agents", "skills", "modified-skill"), "# changed modified\n")
	writePlanFixtureSkill(t, filepath.Join(directory, ".agents", "skills", "collision"), "# collision\n")
	writePlanFixtureSkill(t, filepath.Join(directory, ".claude", "skills", "unknown"), "# unknown\n")
	if err := os.MkdirAll(layout.QuarantinePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.QuarantinePath, "sentinel"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	state := sjskills.ProvenanceState{Version: sjskills.ProvenanceStateVersion, Records: []sjskills.ProvenanceRecord{
		{Scope: sjskills.ScopeProject, Skill: "fixture-skill", Target: sjskills.TargetAgents, SourceIdentity: "github:example/fixture-skill", TreeHashAlgorithm: expectedFixture.Algorithm, TreeHash: expectedFixture.Digest, RecordedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{Scope: sjskills.ScopeProject, Skill: "fixture-skill", Target: sjskills.TargetClaude, SourceIdentity: "github:example/fixture-skill", TreeHashAlgorithm: oldFixture.Algorithm, TreeHash: oldFixture.Digest, RecordedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{Scope: sjskills.ScopeProject, Skill: "modified-skill", Target: sjskills.TargetAgents, SourceIdentity: "github:example/modified-skill", TreeHashAlgorithm: oldModified.Algorithm, TreeHash: oldModified.Digest, RecordedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}}
	stateData, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.ReconcilerStatePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.ReconcilerStatePath, stateData, 0o644); err != nil {
		t.Fatal(err)
	}
	before := map[string]fixtureTree{
		"agents":     captureFixtureTree(t, layout.AgentsSkillsPath),
		"claude":     captureFixtureTree(t, layout.ClaudeSkillsPath),
		"derived":    captureFixtureTree(t, layout.DerivedDirectoryPath),
		"quarantine": captureFixtureTree(t, layout.QuarantinePath),
	}
	globalHome, globalUserProfile, globalAgents, globalClaude := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	overrides := map[string]string{
		"HOME": globalHome, "USERPROFILE": globalUserProfile,
		"CODEX_HOME": globalAgents, "CLAUDE_CONFIG_DIR": globalClaude,
	}

	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "plan")
	if code != 0 || stderr != "" {
		t.Fatalf("project plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	envelope := decodeEnvelope(t, stdout)
	if envelope.Plan == nil || len(envelope.Plan.Operations) != 6 {
		t.Fatalf("plan operations = %#v, want six desired placements", envelope.Plan)
	}
	want := []struct {
		target sjskills.Target
		skill  string
		action string
		reason string
	}{
		{sjskills.TargetAgents, "collision", "blocked", "desired-path-unmanaged"},
		{sjskills.TargetAgents, "fixture-skill", "unchanged", "verified-exact"},
		{sjskills.TargetAgents, "modified-skill", "blocked", "local-modification"},
		{sjskills.TargetClaude, "collision", "install", "expected-entry-absent"},
		{sjskills.TargetClaude, "fixture-skill", "update", "verified-update"},
		{sjskills.TargetClaude, "modified-skill", "install", "expected-entry-absent"},
	}
	for index, operation := range envelope.Plan.Operations {
		if operation.Target != string(want[index].target) || operation.Skill != want[index].skill || operation.Action != want[index].action || operation.Reason != want[index].reason {
			t.Errorf("operation[%d] = %#v, want %s/%s/%s/%s", index, operation, want[index].target, want[index].skill, want[index].action, want[index].reason)
		}
	}
	if !hasWarning(envelope.Warnings, "unmanaged-preserved") {
		t.Fatalf("warnings = %#v, want preserved unknown entry warning", envelope.Warnings)
	}
	if strings.Contains(stdout, "sjskills-materialize-") || strings.Contains(stdout, filepath.Join(directory, ".agents", "skills")) || strings.Contains(stdout, filepath.Join(directory, ".claude", "skills")) {
		t.Fatalf("plan leaked staging or project placement path: %q", stdout)
	}
	if got := captureFixtureTree(t, layout.AgentsSkillsPath); !reflect.DeepEqual(got, before["agents"]) {
		t.Fatalf(".agents changed after plan: before=%#v after=%#v", before["agents"], got)
	}
	if got := captureFixtureTree(t, layout.ClaudeSkillsPath); !reflect.DeepEqual(got, before["claude"]) {
		t.Fatalf(".claude changed after plan: before=%#v after=%#v", before["claude"], got)
	}
	if got := captureFixtureTree(t, layout.DerivedDirectoryPath); !reflect.DeepEqual(got, before["derived"]) {
		t.Fatalf(".sjskills changed after plan: before=%#v after=%#v", before["derived"], got)
	}
	if got := captureFixtureTree(t, layout.QuarantinePath); !reflect.DeepEqual(got, before["quarantine"]) {
		t.Fatalf("quarantine changed after plan: before=%#v after=%#v", before["quarantine"], got)
	}

	code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes")
	if code != 2 || stderr != "" || !strings.Contains(stdout, `"result":"conflict"`) || !strings.Contains(stdout, `"code":"reconciliation_conflict"`) || !strings.Contains(stdout, `"action":"blocked"`) {
		t.Fatalf("project apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for name, root := range map[string]string{"agents": layout.AgentsSkillsPath, "claude": layout.ClaudeSkillsPath, "derived": layout.DerivedDirectoryPath, "quarantine": layout.QuarantinePath} {
		if got := captureFixtureTree(t, root); !reflect.DeepEqual(got, before[name]) {
			t.Fatalf("%s changed after apply: before=%#v after=%#v", name, before[name], got)
		}
	}
}

func TestExternalProjectApplyInstallsIdempotently(t *testing.T) {
	directory := t.TempDir()
	manifest := "version = 1\nprofiles = []\n\n[[direct]]\nname = \"fixture-skill\"\nsource = \"example/fixture-skill\"\n"
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	discovered, err := sjskills.DiscoverProjectRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := sjskills.LayoutForProject(discovered.Root)
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
		writePlanFixtureSkill(t, filepath.Join(root, "unknown"), "# unknown\n")
	}
	globalRoots := map[string]string{
		"HOME": t.TempDir(), "USERPROFILE": t.TempDir(),
		"CODEX_HOME": t.TempDir(), "CLAUDE_CONFIG_DIR": t.TempDir(),
	}
	globalBefore := make(map[string]fixtureTree, len(globalRoots))
	for name, root := range globalRoots {
		if err := os.WriteFile(filepath.Join(root, "sentinel"), []byte(name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		globalBefore[name] = captureFixtureTree(t, root)
	}

	beforeDeclined := captureFixtureTree(t, directory)
	code, stdout, stderr := runCLIWithEnvironment(t, directory, globalRoots, "--json", "apply")
	if code != 65 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"path":"apply.yes"`) {
		t.Fatalf("unconfirmed JSON apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, beforeDeclined) {
		t.Fatalf("JSON apply without --yes mutated project: before=%#v after=%#v", beforeDeclined, after)
	}

	code, stdout, stderr = runCLIWithEnvironment(t, directory, globalRoots, "--json", "apply", "--yes")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"success"`) || !strings.Contains(stdout, `"detail":"installed 2 project placements"`) {
		t.Fatalf("project apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, output := range []string{stdout, stderr} {
		for _, private := range []string{"sjskills-materialize-", ".sjskills-install-", layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
			if strings.Contains(output, private) {
				t.Fatalf("project apply leaked private placement state %q: %q", private, output)
			}
		}
	}
	for _, root := range []string{layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
		data, readErr := os.ReadFile(filepath.Join(root, "fixture-skill", "SKILL.md"))
		if readErr != nil || string(data) != "# fixture-skill\n" {
			t.Fatalf("installed fixture under %q: data=%q err=%v", root, data, readErr)
		}
		if data, readErr := os.ReadFile(filepath.Join(root, "unknown", "SKILL.md")); readErr != nil || string(data) != "# unknown\n" {
			t.Fatalf("unknown entry under %q changed: data=%q err=%v", root, data, readErr)
		}
	}
	stateInfo, err := os.Stat(layout.ReconcilerStatePath)
	if err != nil {
		t.Fatal(err)
	}
	if stateInfo.Mode().Perm() != 0o600 {
		t.Fatalf("provenance mode = %o, want 600", stateInfo.Mode().Perm())
	}
	inventory, err := sjskills.InspectProject(layout)
	if err != nil || !inventory.StateTrusted || len(inventory.State.Records) != 2 {
		t.Fatalf("strict provenance inventory=%#v err=%v", inventory, err)
	}
	firstApply := captureFixtureTree(t, directory)

	code, stdout, stderr = runCLIWithEnvironment(t, directory, globalRoots, "--json", "apply", "--yes")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"detail":"installed 0 project placements"`) {
		t.Fatalf("idempotent apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, firstApply) {
		t.Fatalf("idempotent apply changed project: before=%#v after=%#v", firstApply, after)
	}
	code, stdout, stderr = runCLIWithEnvironment(t, directory, globalRoots, "--json", "apply", "--global", "--yes")
	if code != 2 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"unavailable"`) || !strings.Contains(stdout, `"detail":"global managed roots unchanged"`) {
		t.Fatalf("global apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, firstApply) {
		t.Fatalf("global apply changed project: before=%#v after=%#v", firstApply, after)
	}
	for name, root := range globalRoots {
		if after := captureFixtureTree(t, root); !reflect.DeepEqual(after, globalBefore[name]) {
			t.Fatalf("global sentinel %s changed: before=%#v after=%#v", name, globalBefore[name], after)
		}
	}
}

func TestExternalHumanProjectApplyConfirmation(t *testing.T) {
	directory := t.TempDir()
	manifest := "version = 1\nprofiles = []\n\n[[direct]]\nname = \"fixture-skill\"\nsource = \"example/fixture-skill\"\n"
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	discovered, err := sjskills.DiscoverProjectRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := sjskills.LayoutForProject(discovered.Root)
	if err != nil {
		t.Fatal(err)
	}
	overrides := map[string]string{
		"HOME": t.TempDir(), "USERPROFILE": t.TempDir(),
		"CODEX_HOME": t.TempDir(), "CLAUDE_CONFIG_DIR": t.TempDir(),
	}
	before := captureFixtureTree(t, directory)

	code, stdout, stderr := runCLIWithInputEnvironment(t, directory, overrides, "\n", "apply")
	if code != 2 || !strings.Contains(stdout, "apply: unavailable") || strings.Count(stderr, "Apply 2 project skill installs?") != 1 || !strings.Contains(stderr, "project apply was not confirmed") {
		t.Fatalf("declined apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, before) {
		t.Fatalf("default-declined apply mutated project: before=%#v after=%#v", before, after)
	}

	code, stdout, stderr = runCLIWithInputEnvironment(t, directory, overrides, "yes\n", "apply")
	if code != 0 || !strings.Contains(stdout, "apply: success") || strings.Count(stderr, "Apply 2 project skill installs?") != 1 {
		t.Fatalf("confirmed apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, root := range []string{layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
		data, readErr := os.ReadFile(filepath.Join(root, "fixture-skill", "SKILL.md"))
		if readErr != nil || string(data) != "# fixture-skill\n" {
			t.Fatalf("confirmed placement under %q: data=%q err=%v", root, data, readErr)
		}
	}
}

func TestExternalHumanProjectUpdateConfirmation(t *testing.T) {
	directory, layout, _, overrides := newExternalUpdateFixture(t)
	before := captureFixtureTree(t, directory)

	code, stdout, stderr := runCLIWithInputEnvironment(t, directory, overrides, "\n", "apply")
	if code != 2 || !strings.Contains(stdout, "apply: unavailable") || strings.Count(stderr, "Apply 2 project skill updates?") != 1 || !strings.Contains(stderr, "project apply was not confirmed") {
		t.Fatalf("declined update code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, before) {
		t.Fatalf("default-declined update mutated project: before=%#v after=%#v", before, after)
	}

	code, stdout, stderr = runCLIWithInputEnvironment(t, directory, overrides, "yes\n", "apply")
	if code != 0 || !strings.Contains(stdout, "apply: success") || strings.Count(stderr, "Apply 2 project skill updates?") != 1 || !strings.Contains(stdout, "quarantine: id=") || !strings.Contains(stdout, " status=committed") {
		t.Fatalf("confirmed update code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, output := range []string{stdout, stderr} {
		if strings.Contains(output, layout.Root) || strings.Contains(output, "sjskills-materialize-") || strings.Contains(output, ".sjskills-install-") {
			t.Fatalf("confirmed update leaked a private path: %q", output)
		}
	}
}

func TestExternalJSONProjectUpdateAndIdempotence(t *testing.T) {
	directory, layout, oldHash, overrides := newExternalUpdateFixture(t)
	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("JSON update code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	envelope := decodeEnvelope(t, stdout)
	if envelope.Result != "success" || !hasSerializedEvidence(envelope.Evidence, "execution", "installed 0 project placements") || !hasSerializedEvidence(envelope.Evidence, "execution", "updated 2 project placements") {
		t.Fatalf("JSON update envelope=%#v", envelope)
	}
	quarantineDetail := serializedEvidenceDetail(envelope.Evidence, "quarantine")
	if !strings.HasPrefix(quarantineDetail, "id=") || !strings.HasSuffix(quarantineDetail, " status=committed") {
		t.Fatalf("quarantine evidence=%q", quarantineDetail)
	}
	id := strings.TrimSuffix(strings.TrimPrefix(quarantineDetail, "id="), " status=committed")
	if len(id) != 32 {
		t.Fatalf("quarantine id=%q", id)
	}
	manifestData, err := os.ReadFile(filepath.Join(layout.QuarantinePath, id, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, valid := sjskills.DecodeProjectQuarantineManifest(manifestData)
	if !valid || manifest.Status != sjskills.ProjectQuarantineCommitted || len(manifest.Entries) != 2 {
		t.Fatalf("committed manifest=%#v valid=%v", manifest, valid)
	}
	for _, entry := range manifest.Entries {
		if entry.OldTreeHash != oldHash.Digest {
			t.Fatalf("old hash=%q, want %q", entry.OldTreeHash, oldHash.Digest)
		}
		oldData, readErr := os.ReadFile(filepath.Join(layout.QuarantinePath, id, filepath.FromSlash(entry.QuarantinedPlacement), "SKILL.md"))
		if readErr != nil || string(oldData) != "# fixture-skill v1\n" {
			t.Fatalf("quarantined old bytes for %#v: data=%q err=%v", entry, oldData, readErr)
		}
	}
	for _, root := range []string{layout.AgentsSkillsPath, layout.ClaudeSkillsPath} {
		data, readErr := os.ReadFile(filepath.Join(root, "fixture-skill", "SKILL.md"))
		if readErr != nil || string(data) != "# fixture-skill v2\n" {
			t.Fatalf("updated placement under %q: data=%q err=%v", root, data, readErr)
		}
	}
	beforeRerun := captureFixtureTree(t, directory)
	code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes")
	if code != 0 || stderr != "" {
		t.Fatalf("idempotent JSON update code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	envelope = decodeEnvelope(t, stdout)
	if !hasSerializedEvidence(envelope.Evidence, "execution", "updated 0 project placements") || serializedEvidenceDetail(envelope.Evidence, "quarantine") != "" {
		t.Fatalf("idempotent update envelope=%#v", envelope)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(after, beforeRerun) {
		t.Fatalf("idempotent update changed project: before=%#v after=%#v", beforeRerun, after)
	}
}

func TestExternalHumanProjectUpdateReportsRecoveryRequired(t *testing.T) {
	directory, layout, _, overrides := newExternalUpdateFixture(t)
	destination := filepath.Join(layout.AgentsSkillsPath, "fixture-skill")
	command, stdout, stderr := newCLICommand(t, directory, overrides, "yes\n", "apply")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	processDone := make(chan error, 1)
	go func() { processDone <- command.Wait() }()
	intercepted := false
	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	for !intercepted {
		select {
		case err := <-processDone:
			t.Fatalf("update process exited before recovery race was injected: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
		case <-deadline.C:
			t.Fatalf("timed out waiting for quarantined destination: stdout=%q stderr=%q", stdout.String(), stderr.String())
		default:
			if _, err := os.Lstat(destination); os.IsNotExist(err) {
				if mkdirErr := os.Mkdir(destination, 0o755); mkdirErr == nil {
					if writeErr := os.WriteFile(filepath.Join(destination, "SKILL.md"), []byte("# external replacement\n"), 0o644); writeErr != nil {
						t.Fatal(writeErr)
					}
					intercepted = true
				}
			}
			time.Sleep(100 * time.Microsecond)
		}
	}
	err := <-processDone
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 2 {
		t.Fatalf("recovery-required update err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "quarantine: id=") || !strings.Contains(stdout.String(), " status=recovery-required") || strings.Count(stderr.String(), "Apply 2 project skill updates?") != 1 {
		t.Fatalf("recovery output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	for _, output := range []string{stdout.String(), stderr.String()} {
		if strings.Contains(output, layout.Root) || strings.Contains(output, "sjskills-materialize-") || strings.Contains(output, ".sjskills-install-") {
			t.Fatalf("recovery output leaked a private path: %q", output)
		}
	}
	data, readErr := os.ReadFile(filepath.Join(destination, "SKILL.md"))
	if readErr != nil || string(data) != "# external replacement\n" {
		t.Fatalf("external replacement was not preserved: data=%q err=%v", data, readErr)
	}
	runs, err := os.ReadDir(layout.QuarantinePath)
	if err != nil || len(runs) != 1 {
		t.Fatalf("quarantine runs=%d err=%v", len(runs), err)
	}
	manifestData, err := os.ReadFile(filepath.Join(layout.QuarantinePath, runs[0].Name(), "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, valid := sjskills.DecodeProjectQuarantineManifest(manifestData)
	if !valid || manifest.Status != sjskills.ProjectQuarantineRecoveryRequired {
		t.Fatalf("recovery manifest=%#v valid=%v", manifest, valid)
	}
	oldData, readErr := os.ReadFile(filepath.Join(layout.QuarantinePath, runs[0].Name(), "entries", string(sjskills.TargetAgents), "fixture-skill", "SKILL.md"))
	if readErr != nil || string(oldData) != "# fixture-skill v1\n" {
		t.Fatalf("recoverable old bytes: data=%q err=%v", oldData, readErr)
	}
}

func newExternalUpdateFixture(t *testing.T) (string, sjskills.DerivedLayout, sjskills.TreeHash, map[string]string) {
	t.Helper()
	directory := t.TempDir()
	manifest := "version = 1\nprofiles = []\n\n[[direct]]\nname = \"fixture-skill\"\nsource = \"example/fixture-skill\"\n"
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	discovered, err := sjskills.DiscoverProjectRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	layout, err := sjskills.LayoutForProject(discovered.Root)
	if err != nil {
		t.Fatal(err)
	}
	overrides := isolatedExternalHomes(t)
	overrides["SJSKILLS_FAKE_CONTENT"] = " v1"
	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes")
	if code != 0 || stderr != "" || !strings.Contains(stdout, `"detail":"installed 2 project placements"`) || strings.Contains(stdout, `"kind":"quarantine"`) {
		t.Fatalf("v1 fixture install code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	oldHash, err := sjskills.HashSkillTree(filepath.Join(layout.AgentsSkillsPath, "fixture-skill"))
	if err != nil {
		t.Fatal(err)
	}
	claudeHash, err := sjskills.HashSkillTree(filepath.Join(layout.ClaudeSkillsPath, "fixture-skill"))
	if err != nil || claudeHash != oldHash {
		t.Fatalf("v1 fixture hashes differ: agents=%#v claude=%#v err=%v", oldHash, claudeHash, err)
	}
	overrides["SJSKILLS_FAKE_CONTENT"] = " v2"
	return directory, layout, oldHash, overrides
}

func isolatedExternalHomes(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{
		"HOME": t.TempDir(), "USERPROFILE": t.TempDir(),
		"CODEX_HOME": t.TempDir(), "CLAUDE_CONFIG_DIR": t.TempDir(),
	}
}

func hasSerializedEvidence(evidence []sjskillsEvidence, kind, detail string) bool {
	return serializedEvidenceDetailMatching(evidence, kind, detail) != ""
}

func serializedEvidenceDetail(evidence []sjskillsEvidence, kind string) string {
	for _, item := range evidence {
		if item.Kind == kind {
			return item.Detail
		}
	}
	return ""
}

func serializedEvidenceDetailMatching(evidence []sjskillsEvidence, kind, detail string) string {
	for _, item := range evidence {
		if item.Kind == kind && item.Detail == detail {
			return item.Detail
		}
	}
	return ""
}

func writePlanFixtureSkill(t *testing.T, path, content string) sjskills.TreeHash {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := sjskills.HashSkillTree(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func hasWarning(warnings []sjskillsWarning, code string) bool {
	for _, warning := range warnings {
		if warning.Code == code {
			return true
		}
	}
	return false
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
	if globalEnvelope.Plan == nil || len(globalEnvelope.Plan.Operations) != 0 {
		t.Fatalf("global plan operations = %#v, want empty", globalEnvelope.Plan)
	}
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

	t.Run("project classification failure cleans up before return", func(t *testing.T) {
		directory := t.TempDir()
		if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"dev\"]\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		materializer, parent := testInjectedMaterializer(t)
		calls := 0
		var stagedRoot string
		app := &application{
			directory: directory,
			materialize: func(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				calls++
				materialized, err := materializer.Materialize(ctx, skills)
				if err != nil {
					return nil, err
				}
				stagedRoot = materialized.Root()
				snapshot := materialized.Snapshots()[0]
				unexpectedName := "unexpected-materialized-name"
				unexpectedPath := filepath.Join(filepath.Dir(snapshot.Path), unexpectedName)
				if err := os.Rename(snapshot.Path, unexpectedPath); err != nil {
					t.Fatalf("rename staged snapshot: %v", err)
				}
				snapshot.Skill.Name = unexpectedName
				snapshot.Path = unexpectedPath
				return materialized, nil
			},
		}
		envelope := app.plan(context.Background(), false)
		if calls != 1 || envelope.Result != sjskills.ResultUnavailable || envelope.Error == nil || envelope.Error.Code != sjskills.IssueMissingReference {
			t.Fatalf("calls=%d envelope=%#v", calls, envelope)
		}
		if stagedRoot == "" {
			t.Fatal("materializer did not return a staging root")
		}
		if _, err := os.Stat(stagedRoot); !os.IsNotExist(err) {
			t.Fatalf("staging root still exists after classification failure: %q err=%v", stagedRoot, err)
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after classification failure: entries=%d err=%v", len(entries), err)
		}
	})

	t.Run("project translation failure cleans up before return", func(t *testing.T) {
		directory := t.TempDir()
		if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"dev\"]\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		materializer, parent := testInjectedMaterializer(t)
		calls := 0
		var stagedRoot string
		app := &application{
			directory: directory,
			materialize: func(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				calls++
				materialized, err := materializer.Materialize(ctx, skills)
				if err == nil {
					stagedRoot = materialized.Root()
				}
				return materialized, err
			},
			translateProject: func(sjskills.Plan, sjskills.ProjectClassification) (sjskills.Plan, error) {
				return sjskills.Plan{}, errors.New("translation failed")
			},
		}
		envelope := app.plan(context.Background(), false)
		if calls != 1 || envelope.Result != sjskills.ResultUnavailable || envelope.Error == nil || envelope.Error.Code != sjskills.IssueUnavailable || envelope.Error.Message != "translation failed" {
			t.Fatalf("calls=%d envelope=%#v", calls, envelope)
		}
		if stagedRoot == "" {
			t.Fatal("materializer did not return a staging root")
		}
		if _, err := os.Stat(stagedRoot); !os.IsNotExist(err) {
			t.Fatalf("staging root still exists after translation failure: %q err=%v", stagedRoot, err)
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after translation failure: entries=%d err=%v", len(entries), err)
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

		applyEnvelope := app.apply(context.Background(), true, true)
		if calls != 2 || applyEnvelope.Result != sjskills.ResultUnavailable || applyEnvelope.Error == nil || !strings.Contains(applyEnvelope.Error.Message, "global reconciliation") {
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

func TestApplicationProjectApplyLifecycle(t *testing.T) {
	newProject := func(t *testing.T) string {
		t.Helper()
		directory := t.TempDir()
		manifest := "version = 1\nprofiles = []\n\n[[direct]]\nname = \"fixture-skill\"\nsource = \"example/fixture-skill\"\n"
		if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		return directory
	}

	t.Run("success owns one prepare apply and cleanup lifecycle", func(t *testing.T) {
		directory := newProject(t)
		materializer, parent := testInjectedMaterializer(t)
		var materializeCalls, verifyCalls, translateCalls, applyCalls, cleanupCalls int
		var prompt bytes.Buffer
		app := &application{
			directory:    directory,
			input:        strings.NewReader("yes\n"),
			promptOutput: &prompt,
			materialize: func(ctx context.Context, desired []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				materializeCalls++
				return materializer.Materialize(ctx, desired)
			},
			verifyMaterialized: func(plan *sjskills.MaterializationPlan) error {
				verifyCalls++
				return plan.Verify()
			},
			cleanupMaterialized: func(plan *sjskills.MaterializationPlan) error {
				cleanupCalls++
				return plan.Cleanup()
			},
			translateProject: func(plan sjskills.Plan, classification sjskills.ProjectClassification) (sjskills.Plan, error) {
				translateCalls++
				return sjskills.TranslateProjectClassification(plan, classification)
			},
			applyProject: func(_ context.Context, session *sjskills.ProjectApplySession, _ sjskills.ApplyDeps) (sjskills.ApplyResult, error) {
				applyCalls++
				return sjskills.ApplyResult{Plan: session.Plan, Installed: []sjskills.AppliedPlacement{{Skill: "fixture-skill", Target: sjskills.TargetAgents}}}, nil
			},
		}
		envelope := app.apply(context.Background(), false, false)
		if envelope.Result != sjskills.ResultSuccess || envelope.Plan == nil || envelope.Error != nil {
			t.Fatalf("envelope=%#v", envelope)
		}
		if materializeCalls != 1 || verifyCalls != 1 || translateCalls != 1 || applyCalls != 1 || cleanupCalls != 1 {
			t.Fatalf("calls materialize=%d verify=%d translate=%d apply=%d cleanup=%d", materializeCalls, verifyCalls, translateCalls, applyCalls, cleanupCalls)
		}
		if strings.Count(prompt.String(), "Apply 2 project skill installs?") != 1 {
			t.Fatalf("prompt=%q, want exactly one confirmation", prompt.String())
		}
		if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
			t.Fatalf("temporary parent after apply: entries=%d err=%v", len(entries), err)
		}
	})

	for _, test := range []struct {
		name         string
		failure      error
		wantResult   sjskills.Result
		wantCode     sjskills.IssueCode
		installed    []sjskills.AppliedPlacement
		updated      []sjskills.AppliedPlacement
		quarantine   *sjskills.ProjectQuarantineResult
		wantDetails  []string
		wantRecovery bool
	}{
		{name: "conflict", failure: &sjskills.ApplyError{Kind: sjskills.ApplyFailureConflict, Reason: "project state changed after planning"}, wantResult: sjskills.ResultConflict, wantCode: sjskills.IssueReconciliationConflict, wantDetails: []string{"no committed project placements were reported before apply failure"}},
		{name: "unavailable after commit", failure: &sjskills.ApplyError{Kind: sjskills.ApplyFailureUnavailable, Reason: "apply finalization preflight failed"}, wantResult: sjskills.ResultUnavailable, wantCode: sjskills.IssueUnavailable,
			installed:   []sjskills.AppliedPlacement{{Skill: "fixture-skill", Target: sjskills.TargetAgents}},
			updated:     []sjskills.AppliedPlacement{{Skill: "fixture-skill", Target: sjskills.TargetClaude}},
			quarantine:  &sjskills.ProjectQuarantineResult{ID: "0123456789abcdef0123456789abcdef", Status: sjskills.ProjectQuarantineRecoveryRequired},
			wantDetails: []string{"reported 1 committed installed project placements before apply failure", "reported 1 committed updated project placements before apply failure"}, wantRecovery: true},
	} {
		t.Run("apply failure maps "+test.name, func(t *testing.T) {
			materializer, parent := testInjectedMaterializer(t)
			app := &application{
				directory:   newProject(t),
				materialize: materializer.Materialize,
				applyProject: func(_ context.Context, session *sjskills.ProjectApplySession, _ sjskills.ApplyDeps) (sjskills.ApplyResult, error) {
					return sjskills.ApplyResult{Plan: session.Plan, Installed: test.installed, Updated: test.updated, Quarantine: test.quarantine}, test.failure
				},
			}
			envelope := app.apply(context.Background(), false, true)
			if envelope.Result != test.wantResult || envelope.Plan == nil || envelope.Error == nil || envelope.Error.Code != test.wantCode {
				t.Fatalf("envelope=%#v", envelope)
			}
			if strings.Contains(envelope.Error.Message, app.directory) || strings.Contains(envelope.Error.Message, "sjskills-materialize-") {
				t.Fatalf("apply error leaked a private path: %#v", envelope.Error)
			}
			for _, wantDetail := range test.wantDetails {
				if !hasEvidence(envelope.Evidence, "execution", wantDetail) {
					t.Fatalf("apply failure evidence=%#v, want %q", envelope.Evidence, wantDetail)
				}
			}
			if test.wantRecovery != hasEvidence(envelope.Evidence, "quarantine", "id=0123456789abcdef0123456789abcdef status=recovery-required") {
				t.Fatalf("apply failure quarantine evidence=%#v wantRecovery=%v", envelope.Evidence, test.wantRecovery)
			}
			if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
				t.Fatalf("temporary parent after apply failure: entries=%d err=%v", len(entries), err)
			}
		})
	}

	t.Run("cleanup failure overrides success without erasing execution evidence", func(t *testing.T) {
		materializer, _ := testInjectedMaterializer(t)
		var materialized *sjskills.MaterializationPlan
		cleanupCalls := 0
		app := &application{
			directory: newProject(t),
			materialize: func(ctx context.Context, desired []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
				var err error
				materialized, err = materializer.Materialize(ctx, desired)
				return materialized, err
			},
			cleanupMaterialized: func(*sjskills.MaterializationPlan) error {
				cleanupCalls++
				return errors.New("private cleanup path")
			},
			applyProject: func(_ context.Context, session *sjskills.ProjectApplySession, _ sjskills.ApplyDeps) (sjskills.ApplyResult, error) {
				return sjskills.ApplyResult{Plan: session.Plan, Installed: []sjskills.AppliedPlacement{{Skill: "fixture-skill", Target: sjskills.TargetAgents}}}, nil
			},
		}
		envelope := app.apply(context.Background(), false, true)
		if materialized != nil {
			defer materialized.Cleanup()
		}
		if cleanupCalls != 1 || envelope.Result != sjskills.ResultUnavailable || envelope.Error == nil || envelope.Error.Message != "materialization cleanup failed" {
			t.Fatalf("cleanupCalls=%d envelope=%#v", cleanupCalls, envelope)
		}
		if !hasEvidence(envelope.Evidence, "execution", "installed 1 project placements") || hasEvidenceKind(envelope.Evidence, "materialization") {
			t.Fatalf("cleanup failure evidence=%#v", envelope.Evidence)
		}
	})

	t.Run("blocked plan fails before confirmation and apply", func(t *testing.T) {
		directory := newProject(t)
		writePlanFixtureSkill(t, filepath.Join(directory, ".agents", "skills", "fixture-skill"), "# unmanaged\n")
		materializer, _ := testInjectedMaterializer(t)
		var prompt bytes.Buffer
		applyCalls := 0
		app := &application{
			directory:    directory,
			input:        strings.NewReader("yes\n"),
			promptOutput: &prompt,
			materialize:  materializer.Materialize,
			applyProject: func(context.Context, *sjskills.ProjectApplySession, sjskills.ApplyDeps) (sjskills.ApplyResult, error) {
				applyCalls++
				return sjskills.ApplyResult{}, nil
			},
		}
		envelope := app.apply(context.Background(), false, false)
		if envelope.Result != sjskills.ResultConflict || prompt.Len() != 0 || applyCalls != 0 {
			t.Fatalf("prompt=%q applyCalls=%d envelope=%#v", prompt.String(), applyCalls, envelope)
		}
		if _, err := os.Stat(filepath.Join(directory, ".sjskills")); !os.IsNotExist(err) {
			t.Fatalf("blocked plan wrote derived state: %v", err)
		}
	})
}

func TestConfirmProjectApplyVocabulary(t *testing.T) {
	for _, test := range []struct {
		input string
		want  bool
	}{
		{"y\n", true}, {"YES\n", true}, {"n\n", false}, {"\n", false}, {"", false}, {"sure\n", false},
	} {
		var output bytes.Buffer
		got, err := confirmProjectApply(strings.NewReader(test.input), &output, 2, 0)
		if err != nil || got != test.want || strings.Count(output.String(), "Apply 2 project skill installs?") != 1 {
			t.Errorf("input=%q got=%v err=%v output=%q", test.input, got, err, output.String())
		}
	}
}

func TestConfirmProjectApplyPromptShapes(t *testing.T) {
	for _, test := range []struct {
		installs int
		updates  int
		want     string
	}{
		{2, 0, "Apply 2 project skill installs? [y/N] "},
		{0, 2, "Apply 2 project skill updates? [y/N] "},
		{1, 2, "Apply 3 project skill changes (1 install, 2 updates)? [y/N] "},
	} {
		var output bytes.Buffer
		confirmed, err := confirmProjectApply(strings.NewReader("n\n"), &output, test.installs, test.updates)
		if err != nil || confirmed || output.String() != test.want {
			t.Errorf("installs=%d updates=%d confirmed=%v err=%v output=%q want=%q", test.installs, test.updates, confirmed, err, output.String(), test.want)
		}
	}
}

func hasEvidence(evidence []sjskills.Evidence, kind, detail string) bool {
	for _, item := range evidence {
		if item.Kind == kind && item.Detail == detail {
			return true
		}
	}
	return false
}

func hasEvidenceKind(evidence []sjskills.Evidence, kind string) bool {
	for _, item := range evidence {
		if item.Kind == kind {
			return true
		}
	}
	return false
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
	Operations []struct {
		Action  string `json:"action"`
		Skill   string `json:"skill"`
		Target  string `json:"target"`
		Reason  string `json:"reason"`
		Current struct {
			Kind   string `json:"kind"`
			Detail string `json:"detail"`
		} `json:"current"`
		Expected struct {
			Kind   string `json:"kind"`
			Detail string `json:"detail"`
		} `json:"expected"`
	} `json:"operations"`
}
