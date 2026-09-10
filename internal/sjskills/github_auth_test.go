package sjskills

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGHConfigReadOnlyMigrationBoundary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	for _, data := range []string{"version: 1\n", "version: \"1\"\n"} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if err := checkGHConfig(dir); err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(path)
		if string(after) != data {
			t.Fatal("config changed")
		}
	}
	for _, data := range []string{"", "git_protocol: https\n", "version: 2\n", "version: 1\nversion: 0\n", "version: []\n", "version: 1\n---\nversion: 0\n", "version: [\nsecret"} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		err := checkGHConfig(dir)
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatalf("config error: %v", err)
		}
		after, _ := os.ReadFile(path)
		if string(after) != data {
			t.Fatal("config changed")
		}
	}
}

func TestCredentialGateAndBoundsNeverInvokeGH(t *testing.T) {
	config := t.TempDir()
	if err := os.WriteFile(filepath.Join(config, "config.yml"), []byte("version: 1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	env := []string{"SJSKILLS_GH_REPOSITORY=owner/repo", "SJSKILLS_GH_EXECUTABLE=" + filepath.Join(t.TempDir(), "never-run"), "SJSKILLS_GH_CONFIG_DIR=" + config}
	for _, input := range []string{"protocol=http\nhost=github.com\npath=owner/repo.git\n", "protocol=https\nhost=evil.invalid\npath=owner/repo.git\n", "protocol=https\nhost=github.com\npath=owner/other.git\n", "protocol=https\nhost=github.com\nhost=github.com\npath=owner/repo.git\n", "protocol=https\nhost=github.com:443\npath=owner/repo.git\n", strings.Repeat("x", 8193)} {
		var out bytes.Buffer
		err := RunGitHubCredential(context.Background(), []string{"get"}, strings.NewReader(input), &out, env)
		if err == nil || out.Len() != 0 || err.Error() == "gh credential lookup failed" {
			t.Fatalf("gate failed %v %q", err, out.String())
		}
	}
	for _, action := range []string{"store", "erase"} {
		if err := RunGitHubCredential(context.Background(), []string{action}, strings.NewReader("password=secret\n"), &bytes.Buffer{}, nil); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	writer := credentialPipeWriter{writer: &out, remaining: 4}
	if _, err := writer.Write([]byte("12345")); err == nil || out.Len() != 0 {
		t.Fatal("pipe bound failed")
	}
}

func TestAuthenticatedFetchFailureCleanup(t *testing.T) {
	for _, mode := range []string{"missing-gh", "missing-git", "migration", "denied", "timeout", "cancel", "oversize", "survivor", "missing-skill"} {
		t.Run(mode, func(t *testing.T) {
			parent := t.TempDir()
			config := t.TempDir()
			configData := "version: 1\n"
			if mode == "migration" {
				configData = "git_protocol: https\n"
			}
			if err := os.WriteFile(filepath.Join(config, "config.yml"), []byte(configData), 0600); err != nil {
				t.Fatal(err)
			}
			runner := defaultMaterializeRunner()
			fallback := runner.invoke
			runner.invoke = func(ctx context.Context, command string, args, env []string) (ProcessResult, error) {
				if command == "git" {
					switch mode {
					case "denied":
						return ProcessResult{Stderr: []byte("sentinel-secret"), ExitCode: 1}, errors.New("sentinel-secret")
					case "timeout":
						<-ctx.Done()
						return ProcessResult{}, ctx.Err()
					case "cancel":
						return ProcessResult{}, context.Canceled
					case "oversize":
						return ProcessResult{Stdout: bytes.Repeat([]byte("sentinel-secret"), 100)}, nil
					case "survivor":
						return ProcessResult{}, errProcessTreeActive
					case "missing-skill":
						return ProcessResult{}, nil
					}
				}
				if len(args) > 1 && args[1] == "add" && mode == "missing-skill" {
					return ProcessResult{}, nil
				}
				return fallback(ctx, command, args, env)
			}
			m := NewMaterializer(MaterializerConfig{Runner: runner, BaseEnvironment: []string{"GH_CONFIG_DIR=" + config}, TempRootFactory: func() (string, error) { return os.MkdirTemp(parent, "stage-") }, LookPath: func(name string) (string, error) {
				if mode == "missing-"+name {
					return "", errors.New("not found")
				}
				return name, nil
			}, Limits: MaterializerLimits{CommandTimeout: 20 * time.Millisecond, MaxStdoutBytes: 100}})
			skill := desiredMaterializeSkill("fixture", "owner/repo")
			skill.Access = AccessGitHubAuthenticated
			plan, err := m.Materialize(context.Background(), []DesiredSkill{skill})
			if err == nil || plan != nil || strings.Contains(err.Error(), "sentinel-secret") {
				t.Fatalf("error %v plan %v", err, plan)
			}
			entries, _ := os.ReadDir(parent)
			if mode == "survivor" {
				if !errors.Is(err, errProcessTreeActive) || len(entries) == 0 {
					t.Fatal("unverified descendants staging removed")
				}
			} else if len(entries) != 0 {
				t.Fatalf("staging retained after %s", mode)
			}
			if mode == "missing-gh" || mode == "missing-git" || mode == "migration" {
				for _, call := range runner.calls {
					if call.command == "git" || call.command == "gh" {
						t.Fatal("preflight invoked auth process")
					}
				}
			}
		})
	}
}

func TestGHEnvironmentAndSourceParsing(t *testing.T) {
	original := t.TempDir()
	config := filepath.Join(original, "gh")
	env := []string{"HOME=" + original, "GH_CONFIG_DIR=" + config, "GIT_CONFIG_COUNT=2", "GIT_TRACE=1", "GH_DEBUG=api", "GH_TOKEN=sentinel", "GITHUB_TOKEN=sentinel", "PATH=/bin"}
	isolated := credentialEnvironment(env, t.TempDir(), "darwin")
	for _, name := range []string{"GIT_TRACE", "GIT_CONFIG_COUNT", "GH_DEBUG", "GH_TOKEN", "GITHUB_TOKEN", "GH_CONFIG_DIR"} {
		if environmentValue(isolated, name, "darwin") != "" {
			t.Fatalf("leaked %s", name)
		}
	}
	if environmentValue(isolated, "GIT_ALLOW_PROTOCOL", "darwin") != "https" {
		t.Fatal("protocol widened")
	}
	for _, tc := range []struct{ source, ref, path string }{{"owner/repo/skills/a", "", "skills/a"}, {"https://github.com/owner/repo/tree/main/skills", "main", "skills"}, {"https://github.com/owner/repo/tree/v1", "v1", ""}, {"https://github.com/owner/repo.git", "", ""}} {
		ref, path := githubSourcePath(tc.source)
		if ref != tc.ref || path != tc.path {
			t.Fatal(tc, ref, path)
		}
	}
	dir, err := ghConfigDirectory(env, "darwin")
	if err != nil || dir != config {
		t.Fatal(dir, err)
	}
	a := DesiredSkill{Name: "a", Source: "o/r", Access: AccessPublic}
	b := a
	b.Access = ""
	if !reflect.DeepEqual(normalizeReviewedEnvelope(Envelope{Plan: &Plan{Desired: DesiredState{Skills: []DesiredSkill{a}}}}), normalizeReviewedEnvelope(Envelope{Plan: &Plan{Desired: DesiredState{Skills: []DesiredSkill{b}}}})) {
		t.Fatal("legacy review compatibility")
	}
}
