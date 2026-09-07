package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

type statusCLIFixture struct{ project, home, cache, log string }

func newStatusCLIFixture(t *testing.T, manifest string) statusCLIFixture {
	t.Helper()
	root := t.TempDir()
	f := statusCLIFixture{filepath.Join(root, "project"), filepath.Join(root, "home"), filepath.Join(root, "cache"), filepath.Join(root, "calls")}
	for _, dir := range []string{f.project, f.home, f.cache} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if manifest != "" {
		if err := os.WriteFile(filepath.Join(f.project, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return f
}
func (f statusCLIFixture) command(t *testing.T, overrides map[string]string, args ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	env := append([]string(nil), os.Environ()...)
	values := map[string]string{"HOME": f.home, "USERPROFILE": f.home, "LOCALAPPDATA": f.cache, "XDG_CACHE_HOME": f.cache, "SJSKILLS_FAKE_LOG": f.log, "PATH": filepath.Dir(testBinary) + string(os.PathListSeparator) + os.Getenv("PATH")}
	for k, v := range overrides {
		values[k] = v
	}
	for k, v := range values {
		setEnvironmentValue(&env, k, v)
	}
	command := exec.Command(testBinary, args...)
	command.Env = env
	command.Dir = f.project
	out, errout := &bytes.Buffer{}, &bytes.Buffer{}
	command.Stdout = out
	command.Stderr = errout
	return command, out, errout
}
func (f statusCLIFixture) run(t *testing.T, overrides map[string]string, args ...string) (int, string, string) {
	t.Helper()
	cmd, out, errout := f.command(t, overrides, args...)
	err := cmd.Run()
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return code, out.String(), errout.String()
}
func (f statusCLIFixture) calls(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(f.log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(string(data))
}
func decodeStatusEnvelope(t *testing.T, output string) sjskills.Envelope {
	t.Helper()
	var envelope sjskills.Envelope
	decoder := json.NewDecoder(strings.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatalf("decode %v: %s", err, output)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("not exactly one document: %v", err)
	}
	return envelope
}
func advisoryFor(t *testing.T, envelope sjskills.Envelope, scope sjskills.Scope) sjskills.Advisory {
	t.Helper()
	for _, value := range envelope.Advisories {
		if value.Scope == scope {
			return value
		}
	}
	t.Fatalf("missing %s advisory: %+v", scope, envelope.Advisories)
	return sjskills.Advisory{}
}
func TestStatusCLIColdWarmLocalDriftAndJSON(t *testing.T) {
	f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
	code, out, errout := f.run(t, nil, "--json", "profiles")
	if code != 0 || errout != "" {
		t.Fatalf("code=%d stderr=%q out=%s", code, errout, out)
	}
	envelope := decodeStatusEnvelope(t, out)
	for _, scope := range []sjskills.Scope{sjskills.ScopeProject, sjskills.ScopeGlobal} {
		value := advisoryFor(t, envelope, scope)
		if value.Freshness != sjskills.AdvisoryFresh || len(value.Findings) == 0 || value.Cached {
			t.Fatalf("cold %+v", value)
		}
	}
	coldCalls := len(f.calls(t))
	if coldCalls == 0 {
		t.Fatal("cold check did not materialize")
	}
	code, out, errout = f.run(t, nil, "profiles")
	if code != 0 || !strings.Contains(out, "go (") || !strings.Contains(errout, "project — missing: modern-go, write-go-docs") || !strings.Contains(errout, "global — missing:") || !strings.Contains(errout, "(+4 more)") || !strings.Contains(errout, "checked <1m ago") {
		t.Fatalf("warm code=%d out=%q err=%q", code, out, errout)
	}
	if len(f.calls(t)) != coldCalls {
		t.Fatal("warm check started upstream process")
	}
	path := filepath.Join(f.project, ".agents", "skills", "modern-go")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("# local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, out, _ = f.run(t, nil, "--json", "profiles")
	value := advisoryFor(t, decodeStatusEnvelope(t, out), sjskills.ScopeProject)
	found := false
	for _, finding := range value.Findings {
		if finding.Skill == "modern-go" && finding.Target == sjskills.TargetAgents && finding.Category == sjskills.AdvisoryConflict && finding.Reason == "desired-path-unmanaged" {
			found = true
		}
	}
	if !found || len(f.calls(t)) != coldCalls {
		t.Fatalf("local drift %+v", value)
	}
}
func TestStatusCLIEligibility(t *testing.T) {
	cases := []struct {
		name, manifest string
		args           []string
		input          string
		check          bool
	}{
		{name: "help", args: []string{"--help"}}, {name: "version", args: []string{"--version"}},
		{name: "JSON version", args: []string{"--json", "--version"}},
		{name: "unknown", args: []string{"unknown"}}, {name: "invalid flag", args: []string{"profiles", "--bad"}},
		{name: "opt out", args: []string{"--no-status-check", "profiles"}},
		{name: "opt out after", args: []string{"profiles", "--no-status-check"}},
		{name: "invalid manifest", manifest: "invalid", args: []string{"plan"}},
		{name: "init failure", args: []string{"init", "no-profile"}},
		{name: "restore failure", args: []string{"restore", "invalid"}},
		{name: "cancelled", manifest: "version = 1\nprofiles = [\"go\"]\n", args: []string{"apply"}, input: "no\n"},
		{name: "init", args: []string{"--json", "init", "go"}, check: true},
		{name: "profiles", args: []string{"--json", "profiles"}, check: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newStatusCLIFixture(t, test.manifest)
			cmd, out, errout := f.command(t, nil, test.args...)
			cmd.Stdin = strings.NewReader(test.input)
			_ = cmd.Run()
			// Cancelled apply necessarily verifies its selected scope, but must not
			// inspect global state or publish either scope's advisory cache.
			calls := f.calls(t)
			if test.check {
				if len(calls) == 0 || len(decodeStatusEnvelope(t, out.String()).Advisories) == 0 {
					t.Fatalf("no check: %s %s", out, errout)
				}
			} else {
				if test.name != "cancelled" && len(calls) != 0 {
					t.Fatalf("ancillary processes %v", calls)
				}
				if test.name == "cancelled" {
					for _, call := range calls {
						if call == "clarify" {
							t.Fatal("global ancillary refresh on cancelled apply")
						}
					}
				}
				cache := f.cache
				if runtime.GOOS == "darwin" {
					cache = filepath.Join(f.home, "Library", "Caches")
				}
				entries, _ := os.ReadDir(cache)
				if len(entries) != 0 {
					t.Fatalf("ineligible cache writes: %v", entries)
				}
			}
		})
	}
}
func TestStatusCLIIndependentMalformedAndMissingProject(t *testing.T) {
	for _, manifest := range []string{"", "invalid"} {
		t.Run(manifest, func(t *testing.T) {
			f := newStatusCLIFixture(t, manifest)
			code, out, errout := f.run(t, nil, "--json", "profiles")
			if code != 0 || errout != "" {
				t.Fatalf("result %d %q", code, errout)
			}
			envelope := decodeStatusEnvelope(t, out)
			global := advisoryFor(t, envelope, sjskills.ScopeGlobal)
			if global.Freshness != sjskills.AdvisoryFresh {
				t.Fatalf("global %+v", global)
			}
			if manifest == "" {
				if len(envelope.Advisories) != 1 {
					t.Fatalf("missing project %+v", envelope)
				}
			} else {
				project := advisoryFor(t, envelope, sjskills.ScopeProject)
				if project.Freshness != sjskills.AdvisoryUnavailable || project.Error == "" {
					t.Fatalf("malformed project %+v", project)
				}
			}
		})
	}
	f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
	code, out, errout := f.run(t, map[string]string{"SJSKILLS_FAKE_FAIL_SKILL": "modern-go"}, "--json", "profiles")
	if code != 0 || errout != "" {
		t.Fatalf("failed advisory changed command: %d %q", code, errout)
	}
	envelope := decodeStatusEnvelope(t, out)
	if advisoryFor(t, envelope, sjskills.ScopeProject).Freshness != sjskills.AdvisoryUnavailable || advisoryFor(t, envelope, sjskills.ScopeGlobal).Freshness != sjskills.AdvisoryFresh {
		t.Fatalf("independence %+v", envelope.Advisories)
	}
	count := len(f.calls(t))
	f.run(t, nil, "profiles")
	if len(f.calls(t)) != count {
		t.Fatal("cold failure cooldown retried immediately")
	}
}
func TestStatusCLIPlanApplyReuseAndPostMutation(t *testing.T) {
	f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
	code, out, errout := f.run(t, nil, "plan")
	if code != 0 || !strings.Contains(out, "install:") || strings.Contains(errout, "project — missing:") || !strings.Contains(errout, "global — missing:") {
		t.Fatalf("plan %d %q %q", code, out, errout)
	}
	countName := func(name string) int {
		count := 0
		for _, call := range f.calls(t) {
			if call == name {
				count++
			}
		}
		return count
	}
	if countName("modern-go") != 1 {
		t.Fatalf("plan was materialized twice: %v", f.calls(t))
	}
	code, out, errout = f.run(t, nil, "--json", "apply", "--yes")
	if code != 0 || errout != "" {
		t.Fatalf("apply %d %q %q", code, out, errout)
	}
	project := advisoryFor(t, decodeStatusEnvelope(t, out), sjskills.ScopeProject)
	if project.Freshness != sjskills.AdvisoryFresh || len(project.Findings) != 0 || countName("modern-go") != 2 {
		t.Fatalf("post apply %+v calls=%v", project, f.calls(t))
	}
	if countName("clarify") != 1 {
		t.Fatal("apply refreshed warm global scope")
	}
	// Removing an undeclared skill produces a real quarantine. Restore should
	// immediately report the restored extra using the same expected snapshot.
	extra := filepath.Join(f.project, ".agents", "skills", "extra")
	if err := os.MkdirAll(extra, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extra, "SKILL.md"), []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errout = f.run(t, nil, "--json", "apply", "--yes")
	if code != 0 {
		t.Fatalf("quarantine %d %q %q", code, out, errout)
	}
	envelope := decodeStatusEnvelope(t, out)
	id := ""
	for _, evidence := range envelope.Evidence {
		if evidence.Kind == "quarantine" {
			id = strings.TrimPrefix(strings.Fields(evidence.Detail)[0], "id=")
		}
	}
	if len(id) != 32 {
		t.Fatalf("missing quarantine: %+v", envelope.Evidence)
	}
	before := len(f.calls(t))
	code, out, errout = f.run(t, nil, "--json", "restore", id, "--yes")
	if code != 0 {
		t.Fatalf("restore %d %q %q", code, out, errout)
	}
	project = advisoryFor(t, decodeStatusEnvelope(t, out), sjskills.ScopeProject)
	found := false
	for _, finding := range project.Findings {
		if finding.Skill == "extra" && finding.Category == sjskills.AdvisoryExtra {
			found = true
		}
	}
	if !found || len(f.calls(t)) != before {
		t.Fatalf("restore state %+v calls=%v", project, f.calls(t))
	}
}
func TestStatusCLIGlobalApprovalIgnoresAdvisorySemanticsOnly(t *testing.T) {
	f := newStatusCLIFixture(t, "")
	code, out, errout := f.run(t, nil, "--json", "plan", "--global")
	if code != 0 {
		t.Fatalf("plan %d %s %s", code, out, errout)
	}
	envelope := decodeStatusEnvelope(t, out)
	if len(envelope.Advisories) != 1 {
		t.Fatal("global advisory absent")
	}
	path, digest := writeReviewedPlan(t, []byte(out))
	code, out, errout = f.run(t, nil, "--json", "apply", "--global", "--yes", "--approved-plan", path, "--approved-plan-sha256", digest)
	if code != 0 || errout != "" {
		t.Fatalf("reviewed apply %d %s %s", code, out, errout)
	}
	global := advisoryFor(t, decodeStatusEnvelope(t, out), sjskills.ScopeGlobal)
	if len(global.Findings) != 0 {
		t.Fatalf("global post apply %+v", global)
	}
}
func TestStatusCommandSnapshotWaitsForSuccessfulCleanup(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"go\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	materializer, _ := testInjectedMaterializer(t)
	app := &application{directory: directory, materialize: materializer.Materialize, cleanupMaterialized: func(plan *sjskills.MaterializationPlan) error { _ = plan.Cleanup(); return context.Canceled }}
	value := app.plan(context.Background(), false)
	if value.Result == sjskills.ResultSuccess || app.statusSnapshot != nil {
		t.Fatalf("failed cleanup published snapshot: %+v", app.statusSnapshot)
	}
}
func TestStatusScopesShareDeadlineAndCancellation(t *testing.T) {
	f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
	var calls atomic.Int32
	entered := make(chan time.Time, 2)
	service := sjskills.StatusService{CacheRoot: f.cache, Refresh: func(ctx context.Context, _ []sjskills.DesiredSkill) (sjskills.StatusSnapshot, error) {
		calls.Add(1)
		deadline, _ := ctx.Deadline()
		entered <- deadline
		<-ctx.Done()
		return sjskills.StatusSnapshot{}, ctx.Err()
	}}
	app := &application{directory: f.project, homeDirectory: func() (string, error) { return f.home, nil }, envelope: sjskills.Envelope{Result: sjskills.ResultSuccess}, statusService: &service}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan []sjskills.Advisory, 1)
	go func() { done <- app.collectStatus(ctx) }()
	d1, d2 := <-entered, <-entered
	if !d1.Equal(d2) || time.Until(d1) > sjskills.StatusRefreshBudget {
		t.Fatal("scopes did not share refresh deadline")
	}
	cancel()
	select {
	case values := <-done:
		if len(values) != 2 || calls.Load() != 2 {
			t.Fatalf("values %+v", values)
		}
	case <-time.After(time.Second):
		t.Fatal("refresh cancellation was not propagated")
	}
}
func TestStatusHumanRendererDeterminismAndSilence(t *testing.T) {
	now := time.Now()
	value := sjskills.Advisory{Scope: sjskills.ScopeProject, Freshness: sjskills.AdvisoryFresh, ObservedAt: &now, ReviewCommand: "sjskills plan"}
	var output bytes.Buffer
	renderStatus(&output, []sjskills.Advisory{value}, nil, now)
	if output.Len() != 0 {
		t.Fatalf("exact notice %q", output.String())
	}
	value.Findings = []sjskills.AdvisoryFinding{{Category: sjskills.AdvisoryUpdate, Skill: "same", Target: sjskills.TargetAgents}, {Category: sjskills.AdvisoryUpdate, Skill: "same", Target: sjskills.TargetClaude}, {Category: sjskills.AdvisoryConflict, Skill: "same", Target: sjskills.TargetClaude}}
	original := append([]sjskills.AdvisoryFinding(nil), value.Findings...)
	renderStatus(&output, []sjskills.Advisory{value}, nil, now)
	if strings.Count(output.String(), "updates available: same") != 1 || !strings.Contains(output.String(), "conflicts need attention: same") || !reflect.DeepEqual(original, value.Findings) {
		t.Fatalf("renderer %q", output.String())
	}
}

func TestStatusRegistryFailurePreservesConfiguredScopes(t *testing.T) {
	for _, manifest := range []string{"", "version = 1\nprofiles = [\"go\"]\n", "invalid ["} {
		t.Run(manifest, func(t *testing.T) {
			f := newStatusCLIFixture(t, manifest)
			values := unavailableRegistryStatus(f.project)
			want := []sjskills.Scope{sjskills.ScopeGlobal}
			if manifest != "" {
				want = []sjskills.Scope{sjskills.ScopeProject, sjskills.ScopeGlobal}
			}
			if len(values) != len(want) {
				t.Fatalf("scopes %+v", values)
			}
			for i, value := range values {
				if value.Scope != want[i] || value.Freshness != sjskills.AdvisoryUnavailable || value.Error != "skill registry unavailable" || value.ObservedAt != nil || len(value.Findings) != 0 {
					t.Fatalf("advisory %+v", value)
				}
			}
		})
	}
}
