package sjskills

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestStagedTreeHashIsHostIndependent stages one committed skill holding a
// 100755 script and a symlink the way a Unix host and a Windows host would,
// and requires one tree-sha256-v3 digest. The Unix-style host enables
// symlinks and file modes; the Windows-style host is Git for Windows'
// default (core.symlinks=false, core.filemode=false, core.autocrlf=true) on a
// filesystem without executable bits, which the test models by clearing them.
func TestStagedTreeHashIsHostIndependent(t *testing.T) {
	git := requireGitWithSystemConfig(t)
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
	run := func(env []string, args ...string) string {
		t.Helper()
		cmd := exec.Command(git, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %q: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	// The index records the modes directly, so the source commits a 100755
	// script and a 120000 symlink even where the filesystem has neither.
	source := filepath.Join(scratch, "source")
	neutral := append(append([]string(nil), base...), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+empty,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	for name, content := range map[string]string{
		"skills/demo/SKILL.md":       "---\nname: demo\n---\nbody\n",
		"skills/demo/scripts/run.sh": "#!/bin/sh\necho run\n",
	} {
		path := filepath.Join(source, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	linkTarget := filepath.Join(scratch, "link-target")
	if err := os.WriteFile(linkTarget, []byte("SKILL.md"), 0o600); err != nil {
		t.Fatal(err)
	}
	run(neutral, "init", "--quiet", "--template=", source)
	run(neutral, "-C", source, "add", "--all")
	run(neutral, "-C", source, "update-index", "--chmod=+x", "skills/demo/scripts/run.sh")
	blob := run(neutral, "-C", source, "hash-object", "-w", "--", linkTarget)
	run(neutral, "-C", source, "update-index", "--add", "--cacheinfo", "120000,"+blob+",skills/demo/guide.md")
	run(neutral, "-C", source, "commit", "--quiet", "-m", "fixture")

	stage := func(label, config string, pinned bool) string {
		t.Helper()
		system := filepath.Join(scratch, label+".gitconfig")
		if err := os.WriteFile(system, []byte(config), 0o600); err != nil {
			t.Fatal(err)
		}
		env := append(append([]string(nil), base...), "GIT_CONFIG_SYSTEM="+system, "GIT_CONFIG_GLOBAL="+empty)
		if pinned {
			home := filepath.Join(scratch, label+"-home")
			if err := os.MkdirAll(home, 0o700); err != nil {
				t.Fatal(err)
			}
			env = isolatedEnvironment(env, home, runtime.GOOS)
		}
		target := filepath.Join(scratch, label)
		run(env, "clone", "--quiet", "--no-hardlinks", "--", source, target)
		return filepath.Join(target, "skills", "demo")
	}
	unixConfig := "[core]\n\tsymlinks = true\n\tfilemode = true\n"
	windowsConfig := "[core]\n\tsymlinks = false\n\tfilemode = false\n\tautocrlf = true\n"

	unix := stage("unix", unixConfig, true)
	windows := stage("windows", windowsConfig, true)
	// Windows reports no executable bit for any file.
	if err := filepath.WalkDir(windows, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.Type().IsRegular() {
			return err
		}
		return os.Chmod(path, 0o644)
	}); err != nil {
		t.Fatal(err)
	}

	for _, staged := range []string{unix, windows} {
		info, err := os.Lstat(filepath.Join(staged, "guide.md"))
		if err != nil || !info.Mode().IsRegular() {
			t.Fatalf("%s: committed symlink was not checked out as a regular file: %v %v", staged, info, err)
		}
		content, err := os.ReadFile(filepath.Join(staged, "guide.md"))
		if err != nil || string(content) != "SKILL.md" {
			t.Fatalf("%s: symlink file content = %q, %v", staged, content, err)
		}
	}
	unixHash, err := HashSkillTree(unix)
	if err != nil {
		t.Fatal(err)
	}
	windowsHash, err := HashSkillTree(windows)
	if err != nil {
		t.Fatal(err)
	}
	if unixHash != windowsHash || unixHash.Algorithm != TreeHashAlgorithmSHA256V3 {
		t.Fatalf("unix-style hash %#v != windows-style hash %#v", unixHash, windowsHash)
	}

	if runtime.GOOS == "windows" {
		return // The controls below need Unix modes and unprivileged symlinks.
	}
	// Controls: the fixture still models both host differences.
	if info, err := os.Stat(filepath.Join(unix, "scripts", "run.sh")); err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("unix-style checkout lost the executable bit: %v %v", info, err)
	}
	if legacyUnix, legacyWindows := legacyTreeHash(t, unix), legacyTreeHash(t, windows); legacyUnix == legacyWindows {
		t.Fatalf("tree-sha256-v2 no longer distinguishes the executable bit: %#v", legacyUnix)
	}
	unpinned := stage("unix-unpinned", unixConfig, false)
	if info, err := os.Lstat(filepath.Join(unpinned, "guide.md")); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("unpinned unix-style checkout did not create the symlink: %v %v", info, err)
	}
}

// requireGitWithSystemConfig returns git, skipping when it is missing or
// predates GIT_CONFIG_SYSTEM (2.32), which models the host configuration.
func requireGitWithSystemConfig(t *testing.T) string {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is unavailable")
	}
	version, err := exec.Command(git, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`(\d+)\.(\d+)`).FindSubmatch(version)
	if match == nil {
		t.Fatalf("unrecognized git version %q", version)
	}
	major, _ := strconv.Atoi(string(match[1]))
	minor, _ := strconv.Atoi(string(match[2]))
	if major < 2 || (major == 2 && minor < 32) {
		t.Skipf("git %s predates GIT_CONFIG_SYSTEM", match[0])
	}
	return git
}
