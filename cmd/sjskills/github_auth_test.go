package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

func githubFixtureGitEnvironment(root string) []string {
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			env = append(env, entry)
		}
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+filepath.Join(root, "no-gitconfig"), "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
}

func githubFixture(t *testing.T) (string, map[string]string) {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "project")
	config := filepath.Join(root, "original-gh")
	bin := filepath.Join(root, "user's 한글 tools")
	for _, d := range []string{project, config, bin} {
		if e := os.MkdirAll(d, 0700); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(config, "config.yml"), []byte("version: 1\n"), 0600); e != nil {
		t.Fatal(e)
	}
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	data, e := os.ReadFile(filepath.Join(filepath.Dir(testBinary), "fakegithub"+suffix))
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"git", "gh"} {
		if e := os.WriteFile(filepath.Join(bin, name+suffix), data, 0700); e != nil {
			t.Fatal(e)
		}
	}
	realGit, e := exec.LookPath("git")
	if e != nil {
		t.Fatal(e)
	}
	repo := filepath.Join(root, "repo")
	command := func(args ...string) {
		t.Helper()
		c := exec.Command(realGit, args...)
		c.Env = githubFixtureGitEnvironment(root)
		if out, e := c.CombinedOutput(); e != nil {
			t.Fatalf("fixture git: %v %s", e, out)
		}
	}
	command("init", repo)
	if e := os.MkdirAll(filepath.Join(repo, "skills", "private-fixture"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(repo, "skills", "private-fixture", "SKILL.md"), []byte("---\nname: private-fixture\ndescription: Private fixture\n---\nFixture\n"), 0600); e != nil {
		t.Fatal(e)
	}
	command("-C", repo, "add", ".")
	command("-C", repo, "-c", "core.hooksPath="+filepath.Join(root, "no-hooks"), "commit", "-m", "fixture")
	env := isolatedExternalHomes(t)
	env["PATH"] = bin + string(os.PathListSeparator) + filepath.Dir(testBinary) + string(os.PathListSeparator) + os.Getenv("PATH")
	env["GH_CONFIG_DIR"] = config
	env["GH_TOKEN"] = ""
	env["GITHUB_TOKEN"] = ""
	env["SJSKILLS_EXPECT_GH_CONFIG"] = config
	env["SJSKILLS_ORIGINAL_HOME"] = env["HOME"]
	env["SJSKILLS_GIT_FIXTURE"] = repo
	env["SJSKILLS_REAL_GIT"] = realGit
	env["SJSKILLS_FAKE_GITHUB_LOG"] = filepath.Join(root, "auth.log")
	env["GIT_TRACE"] = "1"
	env["GH_DEBUG"] = "api"
	env["GIT_CONFIG_COUNT"] = "1"
	env["GIT_CONFIG_KEY_0"] = "credential.helper"
	env["GIT_CONFIG_VALUE_0"] = "untrusted-helper"
	return project, env
}

func writeAuthenticatedManifest(t *testing.T, project string, private bool) {
	t.Helper()
	text := "version=1\nprofiles=[]\n"
	if private {
		text += "\n[[direct]]\nname=\"private-fixture\"\nsource=\"fixture/private/skills\"\naccess=\"github-authenticated\"\n"
	}
	text += "\n[[direct]]\nname=\"public-fixture\"\nsource=\"fixture/public\"\n"
	if e := os.WriteFile(filepath.Join(project, "sjskills.toml"), []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
}

func TestExternalAuthenticatedLifecycleAndCredentialBoundary(t *testing.T) {
	project, env := githubFixture(t)
	writeAuthenticatedManifest(t, project, true)
	before := captureFixtureTree(t, env["GH_CONFIG_DIR"])
	run := func(args ...string) sjskills.Envelope {
		t.Helper()
		code, out, stderr := runCLIWithEnvironment(t, project, env, append([]string{"--json"}, args...)...)
		if code != 0 || stderr != "" {
			t.Fatalf("%v: code=%d out=%s err=%s", args, code, out, stderr)
		}
		if strings.Contains(out, "sjskills-sentinel-credential") || strings.Contains(out, "/github-") || strings.Contains(out, `\\github-`) {
			t.Fatal("credential/staging leaked")
		}
		var e sjskills.Envelope
		if err := json.Unmarshal([]byte(out), &e); err != nil {
			t.Fatal(err)
		}
		return e
	}
	plan := run("plan")
	if plan.Plan.Desired.Skills[0].Access != sjskills.AccessGitHubAuthenticated || plan.Plan.Desired.Skills[0].Source != "fixture/private/skills" {
		t.Fatal(plan)
	}
	run("apply", "--yes")
	for _, target := range []string{".agents", ".claude"} {
		if _, e := os.Stat(filepath.Join(project, target, "skills", "private-fixture", "SKILL.md")); e != nil {
			t.Fatal(e)
		}
	}
	next := run("apply", "--yes")
	for _, op := range next.Plan.Operations {
		if op.Action != sjskills.PlanActionUnchanged {
			t.Fatal(op)
		}
	}
	env["SJSKILLS_FAKE_CONTENT"] = " updated"
	privateFile := filepath.Join(env["SJSKILLS_GIT_FIXTURE"], "skills", "private-fixture", "SKILL.md")
	if err := os.WriteFile(privateFile, []byte("# private fixture upstream update\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "core.hooksPath=" + t.TempDir(), "commit", "-m", "upstream update"}} {
		cmd := exec.Command(env["SJSKILLS_REAL_GIT"], append([]string{"-C", env["SJSKILLS_GIT_FIXTURE"]}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+filepath.Join(t.TempDir(), "empty"))
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("update fixture: %v %s", err, out)
		}
	}
	run("apply", "--yes")
	installed, err := os.ReadFile(filepath.Join(project, ".agents", "skills", "private-fixture", "SKILL.md"))
	if err != nil || string(installed) != "# private fixture upstream update\n" {
		t.Fatalf("upstream update missing: %s %v", installed, err)
	}
	// Access-only changes retain remote ownership; an exact tree remains managed.
	manifest, e := os.ReadFile(filepath.Join(project, "sjskills.toml"))
	if e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(project, "sjskills.toml"), []byte(strings.ReplaceAll(string(manifest), "github-authenticated", "public")), 0600); e != nil {
		t.Fatal(e)
	}
	public := run("plan")
	for _, op := range public.Plan.Operations {
		if op.Action != sjskills.PlanActionUnchanged {
			t.Fatal(op)
		}
	}
	writeAuthenticatedManifest(t, project, false)
	removed := run("apply", "--yes")
	id := ""
	for _, e := range removed.Evidence {
		if e.Kind == "quarantine" {
			id = strings.TrimSuffix(strings.TrimPrefix(e.Detail, "id="), " status=committed")
		}
	}
	if id == "" {
		t.Fatalf("no quarantine: %+v", removed.Evidence)
	}
	run("restore", id, "--yes")
	if !reflect.DeepEqual(before, captureFixtureTree(t, env["GH_CONFIG_DIR"])) {
		t.Fatal("gh configuration modified")
	}
	log, e := os.ReadFile(env["SJSKILLS_FAKE_GITHUB_LOG"])
	if e != nil {
		t.Fatal(e)
	}
	if strings.Count(string(log), "git\n") != strings.Count(string(log), "gh\n") {
		t.Fatalf("wrong-host/path invoked gh: %s", log)
	}
}

func TestExternalAuthenticationFailurePreservesInstallations(t *testing.T) {
	for _, mode := range []string{"SJSKILLS_FAKE_AUTH_FAIL", "SJSKILLS_FAKE_AUTH_OVERSIZE"} {
		t.Run(mode, func(t *testing.T) {
			project, env := githubFixture(t)
			writeAuthenticatedManifest(t, project, true)
			code, out, stderr := runCLIWithEnvironment(t, project, env, "--json", "apply", "--yes")
			if code != 0 {
				t.Fatalf("initial %d %s %s", code, out, stderr)
			}
			before := captureFixtureTree(t, project)
			env[mode] = "1"
			code, out, stderr = runCLIWithEnvironment(t, project, env, "--json", "apply", "--yes")
			if code == 0 || strings.Contains(out+stderr, "sjskills-sentinel-credential") || !strings.Contains(out, "verify gh login") {
				t.Fatalf("failure %d %s %s", code, out, stderr)
			}
			if !reflect.DeepEqual(before, captureFixtureTree(t, project)) {
				t.Fatal("failed auth changed managed project")
			}
		})
	}
}

func TestExternalPublicSelectionDoesNotUseAuthentication(t *testing.T) {
	project, env := githubFixture(t)
	writeAuthenticatedManifest(t, project, false)
	env["SJSKILLS_FAKE_AUTH_FAIL"] = "1"
	code, out, stderr := runCLIWithEnvironment(t, project, env, "--json", "plan")
	if code != 0 {
		t.Fatalf("public %d %s %s", code, out, stderr)
	}
	if _, err := os.Stat(env["SJSKILLS_FAKE_GITHUB_LOG"]); !os.IsNotExist(err) {
		t.Fatal("public invoked auth tools")
	}
}

func TestAuthenticatedManifestRenderingAndProfiles(t *testing.T) {
	m := sjskills.Manifest{Version: 1, Direct: []sjskills.DirectSkill{{Name: "fixture", Source: "fixture/private", Access: sjskills.AccessGitHubAuthenticated}}}
	parsed, err := sjskills.ParseManifest([]byte(renderManifest(m)))
	if err != nil || parsed.Direct[0].Access != sjskills.AccessGitHubAuthenticated {
		t.Fatalf("render %v %+v", err, parsed)
	}
	code, out, stderr := runCLI(t, t.TempDir(), "profiles")
	if code != 0 || !strings.Contains(out, "kicpa-private (7 skills, github-authenticated)") {
		t.Fatalf("profiles %d %s %s", code, out, stderr)
	}
}

func TestNativeAuthenticatedDescendantsStopBeforeCleanup(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			_, env := githubFixture(t)
			env["SJSKILLS_FAKE_AUTH_BLOCK"] = "1"
			env["SJSKILLS_FAKE_AUTH_HEARTBEAT"] = filepath.Join(t.TempDir(), "heartbeat")
			t.Setenv("PATH", env["PATH"])
			base := append([]string{}, os.Environ()...)
			for key, value := range env {
				setEnvironmentValue(&base, key, value)
			}
			stages := t.TempDir()
			materializer := sjskills.NewMaterializer(sjskills.MaterializerConfig{HelperExecutable: testBinary, BaseEnvironment: base, TempRootFactory: func() (string, error) { return os.MkdirTemp(stages, "stage-") }, Limits: sjskills.MaterializerLimits{CommandTimeout: 3 * time.Second}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				plan, err := materializer.Materialize(ctx, []sjskills.DesiredSkill{{Name: "private-fixture", Source: "fixture/private/skills", Manager: sjskills.ManagerSkillsCLI, Mode: sjskills.ModeCopy, Access: sjskills.AccessGitHubAuthenticated}})
				if plan != nil {
					plan.Cleanup()
				}
				done <- err
			}()
			deadline := time.Now().Add(5 * time.Second)
			for {
				if _, err := os.Stat(env["SJSKILLS_FAKE_AUTH_HEARTBEAT"]); err == nil {
					break
				}
				if time.Now().After(deadline) {
					cancel()
					t.Fatal("gh descendant never reached blocked fixture")
				}
				time.Sleep(20 * time.Millisecond)
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("blocked authentication succeeded")
				}
			case <-time.After(6 * time.Second):
				cancel()
				t.Fatal("authentication did not stop")
			}
			before, err := os.ReadFile(env["SJSKILLS_FAKE_AUTH_HEARTBEAT"])
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
			after, _ := os.ReadFile(env["SJSKILLS_FAKE_AUTH_HEARTBEAT"])
			if string(before) != string(after) {
				t.Fatal("gh descendant remained active after materializer returned")
			}
			entries, err := os.ReadDir(stages)
			if err != nil || len(entries) != 0 {
				t.Fatalf("staging cleanup %v %v", entries, err)
			}
		})
	}
}

func TestExternalTokenAuthenticationNeedsNoOriginalConfig(t *testing.T) {
	for _, name := range []string{"GH_TOKEN", "GITHUB_TOKEN"} {
		t.Run(name, func(t *testing.T) {
			project, env := githubFixture(t)
			writeAuthenticatedManifest(t, project, true)
			if err := os.WriteFile(filepath.Join(env["GH_CONFIG_DIR"], "config.yml"), []byte("git_protocol: https\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before := captureFixtureTree(t, env["GH_CONFIG_DIR"])
			env[name] = "sjskills-sentinel-credential-never-log"
			env["SJSKILLS_EXPECT_TOKEN"] = "1"
			code, out, stderr := runCLIWithEnvironment(t, project, env, "--json", "apply", "--yes")
			if code != 0 || strings.Contains(out+stderr, env[name]) {
				t.Fatalf("token auth %d %s %s", code, out, stderr)
			}
			for _, content := range captureFixtureTree(t, project).Files {
				if strings.Contains(string(content), env[name]) {
					t.Fatal("token persisted in project")
				}
			}
			if !reflect.DeepEqual(before, captureFixtureTree(t, env["GH_CONFIG_DIR"])) {
				t.Fatal("token login migrated original configuration")
			}
		})
	}
}

func TestExternalAuthenticatedCommitPin(t *testing.T) {
	project, env := githubFixture(t)
	repo := env["SJSKILLS_GIT_FIXTURE"]
	c := exec.Command(env["SJSKILLS_REAL_GIT"], "-C", repo, "rev-parse", "HEAD")
	c.Env = githubFixtureGitEnvironment(filepath.Dir(repo))
	pin, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(repo, "skills", "private-fixture", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "skills", "private-fixture", "SKILL.md"), []byte("newer unpinned content"), 0600); err != nil {
		t.Fatal(err)
	}
	c = exec.Command(env["SJSKILLS_REAL_GIT"], "-C", repo, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "core.hooksPath="+filepath.Join(repo, "no-hooks"), "commit", "-am", "newer commit")
	c.Env = githubFixtureGitEnvironment(filepath.Dir(repo))
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	writeAuthenticatedManifest(t, project, true)
	manifest := filepath.Join(project, "sjskills.toml")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	source := "https://github.com/fixture/private/tree/" + strings.TrimSpace(string(pin)) + "/skills"
	if err := os.WriteFile(manifest, []byte(strings.ReplaceAll(string(data), "fixture/private/skills", source)), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCLIWithEnvironment(t, project, env, "--json", "apply", "--yes")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d out=%s stderr=%s", code, out, stderr)
	}
	installed, err := os.ReadFile(filepath.Join(project, ".agents", "skills", "private-fixture", "SKILL.md"))
	if err != nil || string(installed) != string(original) {
		t.Fatalf("pin mismatch: %q err=%v", installed, err)
	}
}
