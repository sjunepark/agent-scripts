package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

// movingRegistries models main moving after a global plan was reviewed.
type movingRegistries struct {
	reviewed, moved sjskills.PublishedRegistry
	resolved        int
	at              []string
}

func (m *movingRegistries) Resolve(context.Context) (sjskills.PublishedRegistry, error) {
	m.resolved++
	return m.moved, nil
}
func (m *movingRegistries) At(_ context.Context, commit string) (sjskills.PublishedRegistry, error) {
	m.at = append(m.at, commit)
	if commit == m.reviewed.Commit {
		return m.reviewed, nil
	}
	return m.moved, nil
}
func (m *movingRegistries) ForStatus(ctx context.Context) (sjskills.PublishedRegistry, error) {
	return m.Resolve(ctx)
}
func (m *movingRegistries) ForSelection(ctx context.Context) (sjskills.PublishedRegistry, error) {
	return m.Resolve(ctx)
}

func TestGlobalApplyUsesReviewedRegistryCommit(t *testing.T) {
	reviewed, err := fixtureRegistry()
	if err != nil {
		t.Fatal(err)
	}
	moved := reviewed
	moved.Commit, moved.SHA256 = strings.Repeat("b", 40), strings.Repeat("c", 64)
	newApp := func(registries registryLoader) *application {
		materializer, _ := testInjectedMaterializer(t)
		home := t.TempDir()
		return &application{directory: t.TempDir(), homeDirectory: func() (string, error) { return home, nil }, materialize: materializer.Materialize, registries: registries}
	}

	planned := newApp(registryFunc(func() (sjskills.PublishedRegistry, error) { return reviewed, nil })).plan(context.Background(), true)
	if planned.Result != sjskills.ResultSuccess || planned.Evidence[0] != reviewed.Evidence() {
		t.Fatalf("plan %+v evidence=%+v", planned.Error, planned.Evidence)
	}
	approvedPlan, approvedSHA256 := writeReviewedPlanEnvelope(t, planned)

	registries := &movingRegistries{reviewed: reviewed, moved: moved}
	applied := newApp(registries).applyApproved(context.Background(), true, true, approvedPlan, approvedSHA256)
	if applied.Result != sjskills.ResultSuccess || registries.resolved != 0 || len(registries.at) != 1 || registries.at[0] != reviewed.Commit {
		t.Fatalf("apply %+v resolved=%d at=%v", applied.Error, registries.resolved, registries.at)
	}

	// A recorded commit whose registry differs from the reviewed digest fails.
	tampered := planned
	tampered.Evidence = append([]sjskills.Evidence(nil), planned.Evidence...)
	tampered.Evidence[0] = moved.Evidence()
	tampered.Evidence[0].Detail = strings.Replace(tampered.Evidence[0].Detail, moved.SHA256, reviewed.SHA256, 1)
	approvedPlan, approvedSHA256 = writeReviewedPlanEnvelope(t, tampered)
	registries = &movingRegistries{reviewed: reviewed, moved: moved}
	applied = newApp(registries).applyApproved(context.Background(), true, true, approvedPlan, approvedSHA256)
	if applied.Result != sjskills.ResultConflict || !strings.Contains(applied.Error.Message, "does not match the approved plan") {
		t.Fatalf("tampered apply %+v", applied)
	}

	// Plans from embedded-registry releases cannot be applied.
	legacy := planned
	legacy.Evidence = append([]sjskills.Evidence{{Kind: "registry", Detail: "embedded version 4"}}, planned.Evidence[1:]...)
	approvedPlan, approvedSHA256 = writeReviewedPlanEnvelope(t, legacy)
	registries = &movingRegistries{reviewed: reviewed, moved: moved}
	applied = newApp(registries).applyApproved(context.Background(), true, true, approvedPlan, approvedSHA256)
	if applied.Result != sjskills.ResultConflict || !strings.Contains(applied.Error.Message, "published registry commit") || len(registries.at)+registries.resolved != 0 {
		t.Fatalf("legacy apply %+v", applied)
	}
}

func TestRegistryFailuresAndStaleWarnings(t *testing.T) {
	stale, err := fixtureRegistry()
	if err != nil {
		t.Fatal(err)
	}
	stale.Stale, stale.StaleReason, stale.ResolvedAt = true, "registry branch resolution returned HTTP 503", time.Now().Add(-50*time.Hour)
	app := &application{directory: t.TempDir(), registries: registryFunc(func() (sjskills.PublishedRegistry, error) { return stale, nil })}
	envelope := app.profiles(context.Background())
	if envelope.Result != sjskills.ResultSuccess || len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != "registry-stale" {
		t.Fatalf("stale profiles %+v", envelope)
	}
	var out, errout bytes.Buffer
	emitEnvelope(&out, &errout, false, envelope)
	if !strings.Contains(errout.String(), "using cached skill registry at "+fixtureRegistryCommit[:12]+" resolved 2d ago; refresh failed: registry branch resolution returned HTTP 503") {
		t.Fatalf("stale rendering %q", errout.String())
	}

	unsupported := &sjskills.Issue{Code: sjskills.IssueRegistryVersion, Path: "registry.version", Message: "published registry version 5 is not supported"}
	for name, test := range map[string]struct {
		err    error
		result sjskills.Result
	}{
		"unreachable": {errors.New("skill registry unavailable: offline"), sjskills.ResultUnavailable},
		"unsupported": {unsupported, sjskills.ResultInvalid},
	} {
		app := &application{directory: t.TempDir(), registries: registryFunc(func() (sjskills.PublishedRegistry, error) { return sjskills.PublishedRegistry{}, test.err })}
		if envelope := app.plan(context.Background(), true); envelope.Result != test.result {
			t.Fatalf("%s plan %+v", name, envelope)
		}
	}
}

func TestExternalPlanPinsPublishedSourcesToRegistryCommit(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--json", "plan", "--global")
	if code != 0 || stderr != "" {
		t.Fatalf("plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	envelope := decodeStatusEnvelope(t, stdout)
	if envelope.Evidence[0].Kind != "registry" || !strings.HasPrefix(envelope.Evidence[0].Detail, "sjunepark/agent-scripts@"+fixtureRegistryCommit+" sha256:") {
		t.Fatalf("registry evidence %+v", envelope.Evidence)
	}
	for _, skill := range envelope.Plan.Desired.Skills {
		if skill.SourceID == "agent-scripts" && skill.Source != "https://github.com/sjunepark/agent-scripts/tree/"+fixtureRegistryCommit+"/skills" {
			t.Fatalf("unpinned source %+v", skill)
		}
	}

	code, stdout, _ = runCLIWithEnvironment(t, directory, map[string]string{"SJSKILLS_FAKE_PUBLIC_FETCH_FAIL": "1"}, "--json", "plan", "--global")
	if code == 0 || !strings.Contains(stdout, "public Git fetch failed") {
		t.Fatalf("failed fetch code=%d stdout=%q", code, stdout)
	}
}
