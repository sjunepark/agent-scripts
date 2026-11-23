package sjskills

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// assertCanonicalCheckoutConfig requires env to carry exactly
// canonicalCheckoutConfig as its command-scope Git configuration.
func assertCanonicalCheckoutConfig(t *testing.T, env []string) {
	t.Helper()
	if got := envValue(env, "GIT_CONFIG_COUNT"); got != strconv.Itoa(len(canonicalCheckoutConfig)) {
		t.Fatalf("GIT_CONFIG_COUNT = %q in %q", got, env)
	}
	for index, entry := range canonicalCheckoutConfig {
		keyName, valueName := fmt.Sprintf("GIT_CONFIG_KEY_%d", index), fmt.Sprintf("GIT_CONFIG_VALUE_%d", index)
		key, value := envValue(env, keyName), envValue(env, valueName)
		if key != entry[0] || value != entry[1] || !envHas(env, valueName) {
			t.Fatalf("command config %d = %s=%s, want %s=%s", index, key, value, entry[0], entry[1])
		}
	}
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if strings.EqualFold(key, "GIT_CONFIG_PARAMETERS") {
			t.Fatalf("inherited %s reached the process", key)
		}
	}
}

func TestIsolatedEnvironmentReplacesInheritedCommandConfig(t *testing.T) {
	for _, platform := range []string{"darwin", "windows"} {
		env := isolatedEnvironment([]string{
			"PATH=/bin", "git_config_count=9", "GIT_CONFIG_KEY_8=core.autocrlf", "GIT_CONFIG_VALUE_8=true",
			"GIT_CONFIG_PARAMETERS='core.eol'='crlf'", "GIT_ASKPASS=/askpass",
		}, t.TempDir(), platform)
		assertCanonicalCheckoutConfig(t, env)
		for _, name := range []string{"git_config_count", "GIT_CONFIG_KEY_8", "GIT_CONFIG_VALUE_8"} {
			if envHas(env, name) {
				t.Fatalf("%s: inherited %s survived: %q", platform, name, env)
			}
		}
		if envValue(env, "GIT_ASKPASS") != "/askpass" || envValue(env, "PATH") != "/bin" {
			t.Fatalf("%s: ordinary inherited values were dropped: %q", platform, env)
		}
	}
}

// TestCanonicalCheckoutBytesIgnoreHostConfig runs real Git under system
// configurations that differ by host (Git for Windows ships
// core.autocrlf=true and a required LFS filter) and requires the same staged
// bytes and tree hash from each, for both the clone used by named refs and the
// init/fetch/checkout sequence used by commit pins. The expected bytes are the
// committed blobs plus the source's own eol attribute.
func TestCanonicalCheckoutBytesIgnoreHostConfig(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is unavailable")
	}
	// GIT_CONFIG_SYSTEM, which models the host here, needs Git 2.32; the
	// GIT_CONFIG_COUNT mechanism under test needs 2.31.
	version, err := exec.Command(git, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	if match := regexp.MustCompile(`(\d+)\.(\d+)`).FindSubmatch(version); match == nil {
		t.Fatalf("unrecognized git version %q", version)
	} else if major, _ := strconv.Atoi(string(match[1])); major < 2 {
		t.Skipf("git %s predates GIT_CONFIG_SYSTEM", match[0])
	} else if minor, _ := strconv.Atoi(string(match[2])); major == 2 && minor < 32 {
		t.Skipf("git %s predates GIT_CONFIG_SYSTEM", match[0])
	}

	scratch := t.TempDir()
	empty := filepath.Join(scratch, "empty.gitconfig")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	base := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		if key, _, _ := strings.Cut(item, "="); !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			base = append(base, item)
		}
	}
	run := func(env []string, args ...string) {
		t.Helper()
		cmd := exec.Command(git, args...)
		cmd.Env = env
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %q: %v\n%s", args, err, output)
		}
	}
	output := func(env []string, args ...string) string {
		t.Helper()
		cmd := exec.Command(git, args...)
		cmd.Env = env
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %q: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}

	// The source commits LF blobs without any filter. Its .gitattributes marks
	// one path as text (subject to core.eol), one as eol=crlf (a source
	// decision every host must honor), and one as LFS-filtered.
	source := filepath.Join(scratch, "source")
	neutral := append(append([]string(nil), base...), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+empty,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	files := map[string]string{
		".gitattributes":        "skills/demo/text.md text\nskills/demo/crlf.md text eol=crlf\nskills/demo/blob.bin filter=lfs\n",
		"skills/demo/SKILL.md":  "---\nname: demo\n---\nline one\nline two\n",
		"skills/demo/text.md":   "text one\ntext two\n",
		"skills/demo/crlf.md":   "crlf one\ncrlf two\n",
		"skills/demo/blob.bin":  "pointer-like content\n",
		"skills/demo/data.json": "{\n  \"a\": 1\n}\n",
	}
	for name, content := range files {
		path := filepath.Join(source, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	run(neutral, "init", "--quiet", "--template=", source)
	run(neutral, "-C", source, "add", "--all")
	run(neutral, "-C", source, "commit", "--quiet", "-m", "fixture")
	commit := output(neutral, "-C", source, "rev-parse", "HEAD")

	want := map[string]string{}
	for name, content := range files {
		if strings.HasPrefix(name, "skills/demo/") {
			want[strings.TrimPrefix(name, "skills/demo/")] = content
		}
	}
	want["crlf.md"] = strings.ReplaceAll(want["crlf.md"], "\n", "\r\n")

	// Each hostile host rewrites a named file when the canonical environment is
	// absent; the control below proves the fixture still models that host.
	hosts := []struct {
		label, config, rewritten string
		inherited                []string
	}{
		{label: "windows-default", config: "[core]\n\tautocrlf = true\n", rewritten: "SKILL.md"},
		{label: "eol-crlf", config: "[core]\n\teol = crlf\n", rewritten: "text.md"},
		{label: "lfs-filter", config: "[filter \"lfs\"]\n\tsmudge = git hash-object --stdin\n\trequired = true\n", rewritten: "blob.bin"},
		{label: "caller-parameters", rewritten: "SKILL.md", inherited: []string{"GIT_CONFIG_PARAMETERS='core.autocrlf'='true'"}},
		{label: "unset"},
	}
	stage := func(label string, env []string, pinned bool) string {
		t.Helper()
		target := filepath.Join(scratch, label)
		if pinned {
			run(env, "init", "--quiet", "--template=", "--", target)
			run(env, "-C", target, "fetch", "--quiet", "--depth=1", "--no-tags", "--no-recurse-submodules", "--", source, commit)
			run(env, "-C", target, "checkout", "--quiet", "--detach", "--force", "FETCH_HEAD", "--")
		} else {
			run(env, "clone", "--quiet", "--no-hardlinks", "--", source, target)
		}
		return filepath.Join(target, "skills", "demo")
	}
	var reference *TreeHash
	for _, host := range hosts {
		system := filepath.Join(scratch, host.label+".gitconfig")
		if err := os.WriteFile(system, []byte(host.config), 0o600); err != nil {
			t.Fatal(err)
		}
		hostEnv := append(append(append([]string(nil), base...), "GIT_CONFIG_SYSTEM="+system, "GIT_CONFIG_GLOBAL="+empty), host.inherited...)
		if host.rewritten != "" {
			control := stage(host.label+"-control", hostEnv, false)
			got, err := os.ReadFile(filepath.Join(control, host.rewritten))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(got, []byte(want[host.rewritten])) {
				t.Fatalf("%s: control kept %s unchanged; the fixture no longer models this host", host.label, host.rewritten)
			}
		}
		for _, pinned := range []bool{false, true} {
			label := fmt.Sprintf("%s-pinned-%t", host.label, pinned)
			root := filepath.Join(scratch, label+"-home")
			if err := os.MkdirAll(root, 0o700); err != nil {
				t.Fatal(err)
			}
			staged := stage(label, isolatedEnvironment(hostEnv, root, "linux"), pinned)
			for name, content := range want {
				got, err := os.ReadFile(filepath.Join(staged, name))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, []byte(content)) {
					t.Errorf("%s: %s = %q, want %q", label, name, got, content)
				}
			}
			digest, err := hashSkillTree(staged, MaterializerLimits{})
			if err != nil {
				t.Fatal(err)
			}
			if reference == nil {
				reference = &digest
			} else if digest != *reference {
				t.Errorf("%s: tree hash %v differs from %v", label, digest, *reference)
			}
		}
	}
}
