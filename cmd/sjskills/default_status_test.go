package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

// The original isolated reproduction exited 64, stdout="", stderr=
// sjskills: expected one of "init", "profiles", "plan", "apply", "restore".
func TestDefaultStatusCLI(t *testing.T) {
	for _, configured := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "configured"}[configured], func(t *testing.T) {
			manifest := ""
			if configured {
				manifest = "version = 1\nprofiles = [\"go\"]\n"
			}
			f := newStatusCLIFixture(t, manifest)
			var first sjskills.Envelope
			var calls int
			for i, args := range [][]string{{"--json"}, {"status", "--json"}, {"--json", "status"}} {
				code, out, errout := f.run(t, nil, args...)
				if code != 0 || errout != "" {
					t.Fatalf("%v: %d %s %s", args, code, out, errout)
				}
				e := decodeStatusEnvelope(t, out)
				want := sjskills.ProjectNotConfigured
				if configured {
					want = sjskills.ProjectConfigured
				}
				if e.Operation != sjskills.CommandOperationStatus || e.Status == nil || e.Status.ProjectConfiguration != want || e.Plan != nil || len(e.Evidence) != 0 {
					t.Fatalf("envelope %+v", e)
				}
				if configured && e.Status.ProjectRoot == "" {
					t.Fatal("missing root")
				}
				if i == 0 {
					first = e
					calls = len(f.calls(t))
					counts := map[string]int{}
					for _, name := range f.calls(t) {
						counts[name]++
					}
					// Each scope materializes each selected skill once, never an ancillary second pass.
					if counts["clarify"] != 1 || configured && counts["modern-go"] != 1 || !configured && counts["modern-go"] != 0 {
						t.Fatalf("duplicate or wrong refresh %v", counts)
					}
				} else {
					if len(f.calls(t)) != calls {
						t.Fatal("warm status refreshed upstream")
					}
					if !reflect.DeepEqual(e.Status, first.Status) {
						t.Fatal("different setup")
					}
					for j, v := range e.Advisories {
						if !v.Cached || v.Freshness != sjskills.AdvisoryFresh || !reflect.DeepEqual(v.Findings, first.Advisories[j].Findings) {
							t.Fatalf("warm evidence %+v", v)
						}
					}
				}
				if len(e.Advisories) != map[bool]int{false: 1, true: 2}[configured] {
					t.Fatalf("scopes %+v", e.Advisories)
				}
			}
			code, bare, errout := f.run(t, nil)
			code2, named, errout2 := f.run(t, nil, "status")
			if code != 0 || code2 != 0 || errout != "" || errout2 != "" || bare != named || !strings.Contains(bare, "Global:") {
				t.Fatalf("bare/named %d %d %q %q %q %q", code, code2, bare, named, errout, errout2)
			}
			if !configured {
				if !strings.Contains(bare, "Project: not configured") || !strings.Contains(bare, "sjskills init <profile>...") {
					t.Fatalf("setup %q", bare)
				}
				if _, err := os.Stat(filepath.Join(f.project, "sjskills.toml")); !os.IsNotExist(err) {
					t.Fatal("manifest created")
				}
			}
		})
	}
}

func TestDefaultStatusDispatchNoWork(t *testing.T) {
	for _, args := range [][]string{
		{"--no-status-check"}, {"status", "--no-status-check"}, {"--no-status-check", "status"},
		{"--json", "--no-status-check"}, {"status", "--json", "--no-status-check"}, {"--no-status-check", "--json", "status"},
		{"--help"}, {"status", "--help"}, {"--version"}, {"--json", "--version"}, {"--version", "--json"},
		{"unknown"}, {"--bad"}, {"status", "--bad"}, {"status", "extra"}, {"--json", "extra"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := newStatusCLIFixture(t, "")
			code, out, errout := f.run(t, nil, args...)
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "--no-status-check") {
				if code != 0 || errout != "" {
					t.Fatalf("disabled %d %s %s", code, out, errout)
				}
				if strings.Contains(joined, "--json") {
					e := decodeStatusEnvelope(t, out)
					if e.Status == nil || e.Status.ProjectConfiguration != sjskills.ProjectSkipped || e.Status.ProjectRoot != "" || len(e.Advisories) != 0 {
						t.Fatalf("disabled %+v", e)
					}
				} else if out != "Status checks disabled (--no-status-check).\n" {
					t.Fatalf("disabled %q", out)
				}
			} else if strings.Contains(joined, "--help") || strings.Contains(joined, "--version") {
				if code != 0 {
					t.Fatalf("shortcut %d %s %s", code, out, errout)
				}
				if strings.Contains(joined, "--help") && !strings.Contains(out, "status") {
					t.Fatal("missing status help")
				}
			} else if code != 64 {
				t.Fatalf("invalid %d %s %s", code, out, errout)
			}
			if len(f.calls(t)) != 0 {
				t.Fatal("unexpected upstream work")
			}
			for _, root := range []string{f.cache, f.home} {
				entries, err := os.ReadDir(root)
				if err != nil || len(entries) != 0 {
					t.Fatalf("unexpected writes %s %v %v", root, entries, err)
				}
			}
		})
	}
}

func TestDefaultStatusNestedAndUnsafeConfiguration(t *testing.T) {
	for _, kind := range []string{"nested", "malformed", "directory", "symlink", "unreadable"} {
		t.Run(kind, func(t *testing.T) {
			f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
			manifest := filepath.Join(f.project, "sjskills.toml")
			switch kind {
			case "nested":
				nested := filepath.Join(f.project, "child", "deeper")
				if err := os.MkdirAll(nested, 0755); err != nil {
					t.Fatal(err)
				}
				f.project = nested
			case "malformed":
				if err := os.WriteFile(manifest, []byte("invalid ["), 0644); err != nil {
					t.Fatal(err)
				}
			case "directory", "symlink":
				if err := os.Remove(manifest); err != nil {
					t.Fatal(err)
				}
				var err error
				if kind == "directory" {
					err = os.Mkdir(manifest, 0755)
				} else {
					err = os.Symlink(filepath.Join(f.project, "absent"), manifest)
				}
				if err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if runtime.GOOS == "windows" {
					t.Skip("POSIX permissions")
				}
				if err := os.Chmod(manifest, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(manifest, 0644) })
				if _, err := os.ReadFile(manifest); err == nil {
					t.Skip("privileged user bypasses file permissions")
				}
			}
			code, out, errout := f.run(t, nil, "--json")
			e := decodeStatusEnvelope(t, out)
			if code != 0 || errout != "" || advisoryFor(t, e, sjskills.ScopeGlobal).Freshness != sjskills.AdvisoryFresh {
				t.Fatalf("peer %d %s %s", code, out, errout)
			}
			if kind == "nested" {
				root, _ := filepath.EvalSymlinks(filepath.Dir(manifest))
				if e.Status.ProjectRoot != root || e.Status.ProjectConfiguration != sjskills.ProjectConfigured {
					t.Fatalf("nested %+v", e.Status)
				}
				if _, err := os.Stat(filepath.Join(f.project, "sjskills.toml")); !os.IsNotExist(err) {
					t.Fatal("nested manifest created")
				}
			} else if e.Status.ProjectConfiguration != sjskills.ProjectUnavailable || advisoryFor(t, e, sjskills.ScopeProject).Freshness != sjskills.AdvisoryUnavailable {
				t.Fatalf("unsafe %+v", e)
			}
			_, human, _ := f.run(t, nil)
			if kind != "nested" && (strings.Contains(human, "sjskills init") || !strings.Contains(human, "repair")) {
				t.Fatalf("unsafe guidance %q", human)
			}
		})
	}
}

func TestDefaultStatusDiscoveryFailuresAndOptOut(t *testing.T) {
	for _, kind := range []string{"deleted", "file", "inaccessible", "home", "registry", "registry-missing", "registry-malformed", "disabled"} {
		t.Run(kind, func(t *testing.T) {
			f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
			service := sjskills.StatusService{CacheRoot: f.cache, Refresh: func(context.Context, []sjskills.DesiredSkill) (sjskills.StatusSnapshot, error) {
				return sjskills.StatusSnapshot{}, errors.New("test upstream unavailable")
			}}
			app := &application{directory: f.project, homeDirectory: func() (string, error) { return f.home, nil }, statusService: &service}
			switch kind {
			case "deleted":
				app.directory = filepath.Join(f.project, "absent")
			case "file":
				app.directory = filepath.Join(f.project, "sjskills.toml")
			case "inaccessible":
				if runtime.GOOS == "windows" {
					t.Skip("POSIX permissions")
				}
				if err := os.Chmod(f.project, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(f.project, 0755) })
				if _, err := os.ReadDir(f.project); err == nil {
					t.Skip("privileged user bypasses permissions")
				}
			case "home":
				app.homeDirectory = func() (string, error) { return "", errors.New("no home") }
			case "registry", "registry-missing", "registry-malformed":
				app.loadRegistry = func() (sjskills.Registry, error) { return sjskills.Registry{}, errors.New("no registry") }
				if kind == "registry-missing" {
					_ = os.Remove(filepath.Join(f.project, "sjskills.toml"))
				}
				if kind == "registry-malformed" {
					_ = os.WriteFile(filepath.Join(f.project, "sjskills.toml"), []byte("invalid ["), 0644)
				}
			case "disabled":
				app.noStatusCheck = true
				app.directory = "invalid"
				app.homeDirectory = func() (string, error) { t.Fatal("home discovery during opt-out"); return "", nil }
				app.loadRegistry = func() (sjskills.Registry, error) { t.Fatal("registry during opt-out"); return sjskills.Registry{}, nil }
			}
			e := app.status(context.Background())
			if e.ExitStatus() != sjskills.ExitSuccess {
				t.Fatalf("report failure %+v", e)
			}
			want := sjskills.ProjectUnavailable
			switch kind {
			case "home":
				want = sjskills.ProjectConfigured
			case "registry-missing":
				want = sjskills.ProjectNotConfigured
			case "disabled":
				want = sjskills.ProjectSkipped
			}
			if e.Status.ProjectConfiguration != want {
				t.Fatalf("configuration %+v", e)
			}
			if kind != "disabled" {
				_ = advisoryFor(t, e, sjskills.ScopeGlobal)
			}
			if want == sjskills.ProjectUnavailable && advisoryFor(t, e, sjskills.ScopeProject).Freshness != sjskills.AdvisoryUnavailable {
				t.Fatal("not unavailable")
			}
		})
	}
}

func TestDefaultStatusFreshAndApprovalBoundary(t *testing.T) {
	f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
	code, out, errout := f.run(t, nil, "--json", "plan", "--global")
	if code != 0 {
		t.Fatalf("plan %d %s %s", code, out, errout)
	}
	original := out
	path, digest := writeReviewedPlan(t, []byte(out))
	code, out, errout = f.run(t, nil, "--json", "apply", "--global", "--yes", "--approved-plan", path, "--approved-plan-sha256", digest)
	if code != 0 {
		t.Fatalf("global apply %d %s %s", code, out, errout)
	}
	code, out, errout = f.run(t, nil, "apply", "--yes")
	if code != 0 {
		t.Fatalf("project apply %d %s %s", code, out, errout)
	}
	_, human, errout := f.run(t, nil)
	if strings.Count(human, "no drift detected") != 2 || errout != "" {
		t.Fatalf("fresh %q %q", human, errout)
	}
	_, _, errout = f.run(t, nil, "profiles")
	if errout != "" {
		t.Fatalf("incidental fresh %q", errout)
	}
	_, status, _ := f.run(t, nil, "--json")
	for _, artifact := range []string{status, strings.TrimSuffix(strings.TrimSpace(original), "}") + `,"status":{"projectConfiguration":"configured"}}`, strings.TrimSuffix(strings.TrimSpace(original), "}") + `,"status":null}`, strings.TrimSuffix(strings.TrimSpace(original), "}") + `,"Status":null}`, strings.TrimSuffix(strings.TrimSpace(original), "}") + `,"STATUS":null}`} {
		path, digest = writeReviewedPlan(t, []byte(artifact))
		before := len(f.calls(t))
		code, out, errout = f.run(t, nil, "--json", "apply", "--global", "--yes", "--approved-plan", path, "--approved-plan-sha256", digest)
		if code == 0 || len(f.calls(t)) != before {
			t.Fatalf("accepted or materialized status approval %d %s %s", code, out, errout)
		}
	}
}

func TestDefaultStatusPresentation(t *testing.T) {
	now := time.Now()
	old := now.Add(-2 * time.Hour)
	for _, freshness := range []sjskills.AdvisoryFreshness{sjskills.AdvisoryFresh, sjskills.AdvisoryStale, sjskills.AdvisoryUnavailable} {
		for _, drift := range []bool{false, true} {
			v := sjskills.Advisory{Scope: sjskills.ScopeProject, Freshness: freshness, ObservedAt: &old, Cached: true, ReviewCommand: "sjskills plan"}
			if freshness != sjskills.AdvisoryFresh {
				v.Error = "upstream unavailable"
			}
			if drift {
				for _, category := range []sjskills.AdvisoryCategory{sjskills.AdvisoryUpdate, sjskills.AdvisoryMissing, sjskills.AdvisoryExtra, sjskills.AdvisoryConflict} {
					for _, name := range []string{"a\nname", "b", "c", "d", "e", "f", "g"} {
						for _, target := range []sjskills.Target{sjskills.TargetAgents, sjskills.TargetClaude} {
							v.Findings = append(v.Findings, sjskills.AdvisoryFinding{Category: category, Skill: name, Target: target})
						}
					}
				}
			}
			var out bytes.Buffer
			renderStatusReport(&out, sjskills.StatusResult{ProjectConfiguration: sjskills.ProjectConfigured, ProjectRoot: "/project"}, []sjskills.Advisory{v}, now)
			text := out.String()
			if !strings.Contains(text, "checked 2h ago") {
				t.Fatalf("age %q", text)
			}
			if freshness == sjskills.AdvisoryStale && !strings.Contains(text, "against stale evidence") {
				t.Fatalf("stale %q", text)
			}
			if freshness == sjskills.AdvisoryUnavailable && (!strings.Contains(text, "inspection unavailable") || strings.Contains(text, "no drift detected")) {
				t.Fatalf("unavailable %q", text)
			}
			if drift && (strings.Count(text, "(+2 more)") != 4 || strings.Count(text, `"a\nname"`) != 4 || strings.Contains(text, "a\nname")) {
				t.Fatalf("categories %q", text)
			}
			data, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			var decoded sjskills.Advisory
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if len(decoded.Findings) != len(v.Findings) {
				t.Fatal("JSON truncated")
			}
		}
	}
}

func TestDefaultStatusAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	app := &application{loadRegistry: func() (sjskills.Registry, error) { t.Fatal("work after cancellation"); return sjskills.Registry{}, nil }}
	e := app.status(ctx)
	var out, errout bytes.Buffer
	code := emitEnvelope(&out, &errout, false, e)
	if code != int(sjskills.ExitExecutionFailure) || out.Len() != 0 || !strings.Contains(errout.String(), "cancelled") {
		t.Fatalf("cancel %d %q %q", code, out.String(), errout.String())
	}
}
