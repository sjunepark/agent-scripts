package sjskills

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type materializeCall struct {
	command string
	args    []string
	env     []string
}

type materializeRunner struct {
	calls  []materializeCall
	invoke func(context.Context, string, []string, []string) (ProcessResult, error)
}

func (r *materializeRunner) Run(ctx context.Context, command string, args []string, env []string) (ProcessResult, error) {
	r.calls = append(r.calls, materializeCall{
		command: command,
		args:    append([]string(nil), args...),
		env:     append([]string(nil), env...),
	})
	if r.invoke == nil {
		return ProcessResult{}, nil
	}
	return r.invoke(ctx, command, args, env)
}

func defaultMaterializeRunner() *materializeRunner {
	runner := &materializeRunner{}
	runner.invoke = func(_ context.Context, _ string, args []string, env []string) (ProcessResult, error) {
		switch {
		case reflect.DeepEqual(args, []string{"--version"}):
			return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
		case reflect.DeepEqual(args, []string{"skills@" + SkillsCLIVersion, "--version"}):
			return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
		case len(args) >= 2 && args[1] == "add":
			name := args[4]
			root := envValue(env, "CODEX_HOME")
			if err := os.MkdirAll(filepath.Join(root, "skills", name), 0o755); err != nil {
				return ProcessResult{}, err
			}
			if err := os.WriteFile(filepath.Join(root, "skills", name, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
				return ProcessResult{}, err
			}
			return ProcessResult{}, nil
		default:
			return ProcessResult{}, fmt.Errorf("unexpected args: %q", args)
		}
	}
	return runner
}

func testMaterializer(t *testing.T, runner *materializeRunner, limits MaterializerLimits) (*Materializer, func() string) {
	t.Helper()
	parent := t.TempDir()
	factory := func() (string, error) { return os.MkdirTemp(parent, "stage-") }
	m := NewMaterializer(MaterializerConfig{
		Runner:          runner,
		TempRootFactory: factory,
		BaseEnvironment: []string{"PATH=/bin"},
		Limits:          limits,
	})
	return m, func() string {
		entries, err := os.ReadDir(parent)
		if err != nil || len(entries) == 0 {
			return ""
		}
		return filepath.Join(parent, entries[0].Name())
	}
}

func desiredMaterializeSkill(name, source string) DesiredSkill {
	return DesiredSkill{
		Name: name, Source: source, Manager: ManagerSkillsCLI, Mode: ModeCopy,
		Targets: []Target{TargetAgents, TargetClaude},
	}
}

func envValue(env []string, key string) string {
	for _, item := range env {
		name, value, ok := strings.Cut(item, "=")
		if ok && name == key {
			return value
		}
	}
	return ""
}

func envHas(env []string, key string) bool {
	for _, item := range env {
		name, _, ok := strings.Cut(item, "=")
		if ok && name == key {
			return true
		}
	}
	return false
}

func TestMaterializeUsesPinnedPreflightAndExactInstallArguments(t *testing.T) {
	runner := defaultMaterializeRunner()
	m, _ := testMaterializer(t, runner, MaterializerLimits{})
	skill := desiredMaterializeSkill("demo", "example/catalog")
	skill.FullDepth = true
	plan, err := m.Materialize(context.Background(), []DesiredSkill{skill, skill})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := plan.Cleanup(); err != nil {
			t.Errorf("cleanup materialization: %v", err)
		}
	}()
	if len(runner.calls) != 3 {
		t.Fatalf("calls = %d, want preflight twice plus one install", len(runner.calls))
	}
	want := [][]string{
		{"--version"},
		{"skills@" + SkillsCLIVersion, "--version"},
		{"skills@" + SkillsCLIVersion, "add", "example/catalog", "--skill", "demo", "--copy", "--global", "--agent", "codex", "--yes", "--full-depth"},
	}
	for i, call := range runner.calls {
		if call.command != "bunx" || !reflect.DeepEqual(call.args, want[i]) {
			t.Errorf("call %d = %s %q, want bunx %q", i, call.command, call.args, want[i])
		}
	}
	if len(plan.Snapshots()) != 1 {
		t.Fatalf("snapshots = %d, want one for duplicate desired identity", len(plan.Snapshots()))
	}
}

func TestMaterializeRejectsNonEmptyAndFilesystemRootFactoriesWithoutRemovingThem(t *testing.T) {
	sentinelParent := t.TempDir()
	sentinel := filepath.Join(sentinelParent, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("owned by caller"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		root string
		keep func(t *testing.T)
	}{
		{
			name: "non-empty",
			root: sentinelParent,
			keep: func(t *testing.T) {
				if _, err := os.Stat(sentinel); err != nil {
					t.Fatalf("sentinel was removed: %v", err)
				}
			},
		},
		{
			name: "filesystem root",
			root: string(filepath.Separator),
			keep: func(t *testing.T) {
				if _, err := os.Stat(string(filepath.Separator)); err != nil {
					t.Fatalf("filesystem root became unavailable: %v", err)
				}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := NewMaterializer(MaterializerConfig{
				Runner:          defaultMaterializeRunner(),
				TempRootFactory: func() (string, error) { return test.root, nil },
				BaseEnvironment: []string{"PATH=/bin"},
			})
			if err := m.Preflight(context.Background()); err == nil {
				t.Fatal("invalid staging root unexpectedly accepted")
			}
			test.keep(t)
		})
	}
}

func TestMaterializeFullDepthFalseDoesNotAddFlag(t *testing.T) {
	runner := defaultMaterializeRunner()
	m, _ := testMaterializer(t, runner, MaterializerLimits{})
	plan, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := plan.Cleanup(); err != nil {
			t.Errorf("cleanup materialization: %v", err)
		}
	}()
	install := runner.calls[len(runner.calls)-1].args
	if containsArg(install, "--full-depth") {
		t.Fatalf("unexpected --full-depth in default install args: %q", install)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func TestMaterializePreflightMismatchFailureAndTimeoutCleanUp(t *testing.T) {
	for _, test := range []struct {
		name   string
		invoke func(context.Context, string, []string, []string) (ProcessResult, error)
		want   string
	}{
		{
			name: "version mismatch",
			invoke: func(_ context.Context, _ string, args []string, _ []string) (ProcessResult, error) {
				if reflect.DeepEqual(args, []string{"--version"}) {
					return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
				}
				return ProcessResult{Stdout: []byte("1.5.22\n"), Stderr: []byte("https://user:password@example.test/repo?token=secret#fragment")}, nil
			},
			want: "version output",
		},
		{
			name: "bunx failure",
			invoke: func(_ context.Context, _ string, _ []string, _ []string) (ProcessResult, error) {
				return ProcessResult{Stderr: []byte("authorization: Bearer very-secret")}, errors.New("launcher failed")
			},
			want: "preflight",
		},
		{
			name: "timeout",
			invoke: func(ctx context.Context, _ string, _ []string, _ []string) (ProcessResult, error) {
				<-ctx.Done()
				return ProcessResult{}, ctx.Err()
			},
			want: "deadline",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := &materializeRunner{invoke: test.invoke}
			m, stage := testMaterializer(t, runner, MaterializerLimits{CommandTimeout: 10 * time.Millisecond, MaxDiagnosticBytes: 96})
			_, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
			if strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "very-secret") || strings.Contains(err.Error(), "user:") {
				t.Fatalf("diagnostic leaked a credential: %v", err)
			}
			if len([]byte(err.Error())) > 96 {
				t.Fatalf("diagnostic length = %d, want <= 96", len([]byte(err.Error())))
			}
			if root := stage(); root != "" {
				if _, statErr := os.Lstat(root); !os.IsNotExist(statErr) {
					t.Fatalf("failed materialization left staging root %q: %v", root, statErr)
				}
			}
		})
	}
}

func TestMaterializeRejectsUnavailableBunxBeforeRunner(t *testing.T) {
	runner := defaultMaterializeRunner()
	m, stage := testMaterializer(t, runner, MaterializerLimits{})
	m.lookPath = func(string) (string, error) { return "", errors.New("bunx not found") }
	if _, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")}); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("error = %v, want bunx unavailable", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner calls = %d, want none when lookpath fails", len(runner.calls))
	}
	if root := stage(); root != "" {
		if _, statErr := os.Lstat(root); !os.IsNotExist(statErr) {
			t.Fatalf("unavailable bunx left staging root %q: %v", root, statErr)
		}
	}
}

func TestMaterializeInstallFailureCleansAndRedactsDiagnostics(t *testing.T) {
	runner := &materializeRunner{invoke: func(_ context.Context, _ string, args []string, _ []string) (ProcessResult, error) {
		if reflect.DeepEqual(args, []string{"--version"}) {
			return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
		}
		if reflect.DeepEqual(args, []string{"skills@" + SkillsCLIVersion, "--version"}) {
			return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
		}
		return ProcessResult{
			ExitCode: 1,
			Stderr:   []byte("authorization: Bearer install-secret https://user:password@example.test/repo?token=query#fragment"),
		}, errors.New("install failed token=raw-secret")
	}}
	m, stage := testMaterializer(t, runner, MaterializerLimits{MaxDiagnosticBytes: 160})
	_, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err == nil || !strings.Contains(err.Error(), "command failed") {
		t.Fatalf("error = %v, want install failure", err)
	}
	for _, secret := range []string{"install-secret", "raw-secret", "password", "query", "fragment", "user:"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("diagnostic leaked %q: %v", secret, err)
		}
	}
	if root := stage(); root != "" {
		if _, statErr := os.Lstat(root); !os.IsNotExist(statErr) {
			t.Fatalf("install failure left staging root %q: %v", root, statErr)
		}
	}
}

func TestMaterializeRedactsOwnedStagingRootFromPostRootFailure(t *testing.T) {
	parent := t.TempDir()
	sentinel := filepath.Join(parent, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	var root string
	runner := &materializeRunner{invoke: func(_ context.Context, _ string, args []string, env []string) (ProcessResult, error) {
		root = envValue(env, "HOME")
		switch {
		case reflect.DeepEqual(args, []string{"--version"}):
			return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
		case reflect.DeepEqual(args, []string{"skills@" + SkillsCLIVersion, "--version"}):
			return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
		default:
			path := filepath.Join(root, ".agents", "skills", "demo")
			return ProcessResult{Stderr: []byte("staged path " + path)}, fmt.Errorf("cannot read %s: %w", path, os.ErrPermission)
		}
	}}
	m := NewMaterializer(MaterializerConfig{
		Runner: runner,
		TempRootFactory: func() (string, error) {
			return os.MkdirTemp(parent, "sjskills-materialize-")
		},
		BaseEnvironment: []string{"PATH=/bin"},
		Limits:          MaterializerLimits{MaxDiagnosticBytes: 128},
	})
	if _, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")}); err == nil {
		t.Fatal("post-root install failure unexpectedly succeeded")
	} else {
		message := err.Error()
		for _, leaked := range []string{root, filepath.Base(root), "sjskills-materialize-"} {
			if leaked != "" && strings.Contains(message, leaked) {
				t.Fatalf("error leaked staging path component %q: %v", leaked, err)
			}
		}
		if len([]byte(message)) > 128 {
			t.Fatalf("error length = %d, want <= 128", len([]byte(message)))
		}
	}
	if root == "" {
		t.Fatal("runner did not observe an owned staging root")
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("failed materialization left root %q: %v", root, err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("cleanup escaped its root and removed sentinel: %v", err)
	}
}

func TestMaterializeSanitizesBeforeDiagnosticTruncation(t *testing.T) {
	parent := t.TempDir()
	sentinel := filepath.Join(parent, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	var root string
	padding := strings.Repeat("x", 1024)
	runner := &materializeRunner{invoke: func(_ context.Context, _ string, args []string, env []string) (ProcessResult, error) {
		root = envValue(env, "HOME")
		switch {
		case reflect.DeepEqual(args, []string{"--version"}):
			return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
		case reflect.DeepEqual(args, []string{"skills@" + SkillsCLIVersion, "--version"}):
			return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
		default:
			return ProcessResult{Stderr: []byte(root + padding)}, fmt.Errorf("%s%s", root, padding)
		}
	}}
	const diagnosticLimit = 64
	m := NewMaterializer(MaterializerConfig{
		Runner: runner,
		TempRootFactory: func() (string, error) {
			return os.MkdirTemp(parent, "sjskills-materialize-")
		},
		BaseEnvironment: []string{"PATH=/bin"},
		Limits:          MaterializerLimits{MaxDiagnosticBytes: diagnosticLimit},
	})
	_, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err == nil {
		t.Fatal("oversized path diagnostic unexpectedly succeeded")
	}
	message := err.Error()
	if !strings.Contains(message, "demo") || !strings.Contains(message, "Skills CLI command failed") {
		t.Fatalf("diagnostic lost useful stage context: %v", err)
	}
	for _, leaked := range []string{root, filepath.Base(root), "sjskills-materialize-", filepath.Base(filepath.Dir(root))} {
		if leaked != "" && strings.Contains(message, leaked) {
			t.Fatalf("diagnostic leaked path component %q: %v", leaked, err)
		}
	}
	for _, component := range strings.FieldsFunc(filepath.Clean(root), func(r rune) bool { return r == '/' || r == '\\' }) {
		if len(component) >= 4 && strings.Contains(message, component) {
			t.Fatalf("diagnostic leaked partial root component %q: %v", component, err)
		}
	}
	if len([]byte(message)) > diagnosticLimit {
		t.Fatalf("diagnostic length = %d, want <= %d", len([]byte(message)), diagnosticLimit)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("failed materialization left root %q: %v", root, err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("cleanup escaped its root and removed sentinel: %v", err)
	}
}

func TestSkillSnapshotVerifyRedactsOwnedStagingRoot(t *testing.T) {
	parent := t.TempDir()
	sentinel := filepath.Join(parent, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := defaultMaterializeRunner()
	m := NewMaterializer(MaterializerConfig{
		Runner: runner,
		TempRootFactory: func() (string, error) {
			return os.MkdirTemp(parent, "sjskills-materialize-")
		},
		BaseEnvironment: []string{"PATH=/bin"},
		Limits:          MaterializerLimits{MaxDiagnosticBytes: 128},
	})
	plan, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err != nil {
		t.Fatal(err)
	}
	root := plan.Root()
	snapshot, ok := plan.SnapshotFor("demo")
	if !ok {
		t.Fatal("materialized snapshot missing")
	}
	if err := os.RemoveAll(snapshot.Path); err != nil {
		t.Fatal(err)
	}
	for name, verify := range map[string]func() error{
		"snapshot": snapshot.Verify,
		"plan":     plan.Verify,
	} {
		t.Run(name, func(t *testing.T) {
			err := verify()
			if err == nil {
				t.Fatal("verification unexpectedly succeeded")
			}
			message := err.Error()
			for _, leaked := range []string{root, filepath.Base(root), "sjskills-materialize-"} {
				if strings.Contains(message, leaked) {
					t.Fatalf("error leaked staging path component %q: %v", leaked, err)
				}
			}
			if len([]byte(message)) > 128 {
				t.Fatalf("error length = %d, want <= 128", len([]byte(message)))
			}
		})
	}
	if err := plan.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if err := plan.Cleanup(); err != nil {
		t.Fatalf("second cleanup failed: %v", err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("cleanup left root %q: %v", root, err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("cleanup escaped its root and removed sentinel: %v", err)
	}
}

func TestOwnedStageDiagnosticRedactsSlashAndBackslashForms(t *testing.T) {
	root := filepath.Join(t.TempDir(), "sjskills-materialize-abcdef")
	slash := filepath.ToSlash(root)
	backslash := strings.ReplaceAll(slash, "/", `\`)
	escaped := strings.ReplaceAll(backslash, `\`, `\\`)
	message := strings.Join([]string{root, slash, backslash, escaped}, " ")
	redacted := sanitizeOwnedStageError(errors.New(message), root, 96).Error()
	for _, leaked := range []string{root, filepath.Base(root), "sjskills-materialize-"} {
		if strings.Contains(redacted, leaked) {
			t.Fatalf("redacted diagnostic leaked %q: %s", leaked, redacted)
		}
	}
	if len([]byte(redacted)) > 96 {
		t.Fatalf("redacted diagnostic length = %d, want <= 96", len([]byte(redacted)))
	}
}

func TestHashSkillTreeDoesNotRedactStandaloneCallerRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "arbitrary-root")
	_, err := HashSkillTree(root)
	if err == nil || !strings.Contains(err.Error(), root) {
		t.Fatalf("standalone hash error = %v, want caller root %q", err, root)
	}
}

func TestMaterializeBoundsBothOutputStreams(t *testing.T) {
	for _, stream := range []string{"stdout", "stderr"} {
		t.Run(stream, func(t *testing.T) {
			runner := &materializeRunner{invoke: func(_ context.Context, _ string, args []string, _ []string) (ProcessResult, error) {
				result := ProcessResult{}
				if reflect.DeepEqual(args, []string{"--version"}) {
					if stream == "stdout" {
						result.Stdout = []byte("123456789")
					} else {
						result.Stdout = []byte("bunx 1\n")
						result.Stderr = []byte("123456789")
					}
				}
				return result, nil
			}}
			m, _ := testMaterializer(t, runner, MaterializerLimits{MaxStdoutBytes: 4, MaxStderrBytes: 4, MaxDiagnosticBytes: 64})
			if _, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")}); err == nil {
				t.Fatal("oversized process output unexpectedly succeeded")
			}
		})
	}
}

func TestMaterializeIsolatesHomeSignalsAndPreservesCredentialHelpers(t *testing.T) {
	runner := defaultMaterializeRunner()
	parent := t.TempDir()
	m := NewMaterializer(MaterializerConfig{
		Runner: runner,
		TempRootFactory: func() (string, error) {
			return os.MkdirTemp(parent, "stage-")
		},
		Platform: "windows",
		BaseEnvironment: []string{
			"PATH=/custom/bin", "HOME=/real/home", "USERPROFILE=C:\\real",
			"CODEX_HOME=/real/.agents", "CLAUDE_CONFIG_DIR=/real/.claude",
			"HOMEDRIVE=C:", "HOMEPATH=\\Users\\real", "HOMESHARE=\\\\server\\share",
			"TMPDIR=/real/tmp", "TMP=/real/tmp", "TEMP=C:\\real\\temp",
			"XDG_CACHE_HOME=/real/cache", "XDG_CONFIG_HOME=/real/config", "XDG_DATA_HOME=/real/data",
			"NPM_CONFIG_CACHE=/real/npm", "BUN_INSTALL_CACHE_DIR=/real/bun",
			"GIT_ASKPASS=/usr/bin/askpass", "SSH_AUTH_SOCK=/tmp/ssh.sock", "KEEP=value",
		},
	})
	plan, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := plan.Cleanup(); err != nil {
			t.Errorf("cleanup materialization: %v", err)
		}
	}()
	env := runner.calls[0].env
	root := plan.Root()
	want := map[string]string{
		"PATH": "/custom/bin", "HOME": root, "USERPROFILE": root,
		"CODEX_HOME": filepath.Join(root, ".agents"), "CLAUDE_CONFIG_DIR": filepath.Join(root, ".claude"),
		"TMPDIR": filepath.Join(root, ".tmp"), "TMP": filepath.Join(root, ".tmp"), "TEMP": filepath.Join(root, ".tmp"),
		"XDG_CACHE_HOME": filepath.Join(root, ".cache"), "XDG_CONFIG_HOME": filepath.Join(root, ".config"),
		"XDG_DATA_HOME": filepath.Join(root, ".local", "share"), "NPM_CONFIG_CACHE": filepath.Join(root, ".npm"),
		"BUN_INSTALL_CACHE_DIR": filepath.Join(root, ".bun", "install", "cache"),
		"GIT_ASKPASS":           "/usr/bin/askpass", "SSH_AUTH_SOCK": "/tmp/ssh.sock", "KEEP": "value",
	}
	for key, value := range want {
		if got := envValue(env, key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
	for _, key := range []string{"HOMEDRIVE", "HOMEPATH", "HOMESHARE"} {
		if drive, pathPart := windowsHomeParts(root); drive != "" && key != "HOMESHARE" {
			expected := drive
			if key == "HOMEPATH" {
				expected = pathPart
			}
			if got := envValue(env, key); got != expected {
				t.Errorf("%s = %q, want isolated home component %q", key, got, expected)
			}
			continue
		}
		if envHas(env, key) {
			t.Errorf("%s unexpectedly preserved in Windows-like environment", key)
		}
	}
	for _, key := range []string{"TMPDIR", "TMP", "TEMP", "XDG_CACHE_HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "NPM_CONFIG_CACHE", "BUN_INSTALL_CACHE_DIR"} {
		if got := envValue(env, key); got == "" || !pathWithin(root, got) {
			t.Errorf("%s = %q, want path under %q", key, got, root)
		}
	}
	for _, directory := range []string{
		filepath.Join(root, ".agents", "skills"), filepath.Join(root, ".claude"),
		filepath.Join(root, ".tmp"), filepath.Join(root, ".cache"), filepath.Join(root, ".config"),
		filepath.Join(root, ".local", "share"), filepath.Join(root, ".npm"), filepath.Join(root, ".bun", "install", "cache"),
	} {
		info, err := os.Lstat(directory)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("isolated directory %q is not a real directory: info=%v err=%v", directory, info, err)
		}
		if !pathWithin(root, directory) {
			t.Errorf("isolated directory %q escaped root %q", directory, root)
		}
	}
	driveEnv := isolatedEnvironment([]string{"HOMEDRIVE=C:", "HOMEPATH=\\old", "HOMESHARE=\\share"}, `C:\\new\\home`, "windows")
	if envValue(driveEnv, "HOMEDRIVE") != "C:" || envValue(driveEnv, "HOMEPATH") != `\\new\\home` || envHas(driveEnv, "HOMESHARE") {
		t.Fatalf("drive environment = %#v", driveEnv)
	}
	lowercase := isolatedEnvironment([]string{"home=/old", "HOME=/old"}, "/tmp/new", "windows")
	if envValue(lowercase, "HOME") != "/tmp/new" {
		t.Fatalf("case-insensitive HOME override = %#v", lowercase)
	}
	count := 0
	for _, item := range lowercase {
		if strings.HasPrefix(item, "HOME=") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("HOME appears %d times: %#v", count, lowercase)
	}
}

func TestMaterializeSkipsManagerOwnedEntriesAndRejectsUnsupportedSources(t *testing.T) {
	runner := defaultMaterializeRunner()
	m, _ := testMaterializer(t, runner, MaterializerLimits{})
	plan, err := m.Materialize(context.Background(), []DesiredSkill{
		{Name: "manual-entry", Manager: ManagerManual},
		{Name: "manual-entry", Manager: ManagerManual},
		{Name: "workflow-entry", Manager: ManagerWorkflow, Workflow: "release"},
		{Name: "workflow-entry", Manager: ManagerWorkflow, Workflow: "release"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Skipped()) != 2 || len(plan.Snapshots()) != 0 || len(runner.calls) != 0 {
		t.Fatalf("skipped=%#v snapshots=%#v calls=%d", plan.Skipped(), plan.Snapshots(), len(runner.calls))
	}
	if _, err := m.Materialize(context.Background(), []DesiredSkill{{Name: "unowned", Manager: ManagerNone}}); err == nil {
		t.Fatal("manager none unexpectedly treated as a skipped owner")
	}
	if _, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("bad", "https://user:pass@example.test/repo")}); err == nil {
		t.Fatal("credential-bearing source unexpectedly accepted")
	}
	legacyMode := desiredMaterializeSkill("legacy-mode", "example/legacy-mode")
	legacyMode.Mode = InstallMode("symlink")
	if _, err := m.Materialize(context.Background(), []DesiredSkill{legacyMode}); err == nil || !strings.Contains(err.Error(), "copy mode") {
		t.Fatalf("legacy mode error = %v, want copy-mode rejection", err)
	}
	if _, err := m.Materialize(context.Background(), []DesiredSkill{
		desiredMaterializeSkill("same", "example/one"),
		desiredMaterializeSkill("same", "example/two"),
	}); err == nil {
		t.Fatal("contradictory duplicate source unexpectedly accepted")
	}
}

func TestMaterializeRejectsContradictorySkippedIdentity(t *testing.T) {
	m, _ := testMaterializer(t, defaultMaterializeRunner(), MaterializerLimits{})
	cases := []struct {
		name   string
		first  DesiredSkill
		second DesiredSkill
	}{
		{
			name:   "manual source",
			first:  DesiredSkill{Name: "same", Manager: ManagerManual, Source: "first"},
			second: DesiredSkill{Name: "same", Manager: ManagerManual, Source: "second"},
		},
		{
			name:   "workflow",
			first:  DesiredSkill{Name: "same", Manager: ManagerWorkflow, Workflow: "first"},
			second: DesiredSkill{Name: "same", Manager: ManagerWorkflow, Workflow: "second"},
		},
		{
			name:   "options",
			first:  DesiredSkill{Name: "same", Manager: ManagerManual, Targets: []Target{TargetAgents}},
			second: DesiredSkill{Name: "same", Manager: ManagerManual, Targets: []Target{TargetClaude}},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := m.Materialize(context.Background(), []DesiredSkill{test.first, test.second}); err == nil {
				t.Fatal("contradictory skipped identity unexpectedly accepted")
			}
		})
	}
}

func TestMaterializeRejectsMissingNonDirectoryAndSpecialStagedEntries(t *testing.T) {
	for _, mode := range []string{"missing", "non-directory", "special"} {
		t.Run(mode, func(t *testing.T) {
			runner := &materializeRunner{invoke: func(_ context.Context, _ string, args []string, env []string) (ProcessResult, error) {
				if reflect.DeepEqual(args, []string{"--version"}) {
					return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
				}
				if reflect.DeepEqual(args, []string{"skills@" + SkillsCLIVersion, "--version"}) {
					return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
				}
				if len(args) < 2 || args[1] != "add" {
					return ProcessResult{}, fmt.Errorf("unexpected args: %q", args)
				}
				name := args[4]
				root := envValue(env, "CODEX_HOME")
				skillsRoot := filepath.Join(root, "skills")
				if err := os.MkdirAll(skillsRoot, 0o755); err != nil {
					return ProcessResult{}, err
				}
				switch mode {
				case "non-directory":
					return ProcessResult{}, os.WriteFile(filepath.Join(skillsRoot, name), []byte("not a directory"), 0o644)
				case "special":
					if !createSpecialEntry(t, filepath.Join(skillsRoot, name)) {
						t.Skip("special entries unavailable on this platform")
					}
					return ProcessResult{}, nil
				default:
					return ProcessResult{}, nil
				}
			}}
			m, stage := testMaterializer(t, runner, MaterializerLimits{})
			if _, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")}); err == nil {
				t.Fatalf("%s staged output unexpectedly succeeded", mode)
			}
			if root := stage(); root != "" {
				if _, statErr := os.Lstat(root); !os.IsNotExist(statErr) {
					t.Fatalf("failed materialization left staging root %q: %v", root, statErr)
				}
			}
		})
	}
}

func TestHashSkillTreeMatchesLegacyV2VectorAndDetectsTampering(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.sh"), []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "z"), []byte("é"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("a.txt", filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	hash, err := HashSkillTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if hash.Algorithm != TreeHashAlgorithmSHA256V2 || hash.Digest != "1a1731f91aeff9ae9af8aae368d794bd55fe555e4fb6bf40f66e00f0cfde6718" {
		t.Fatalf("hash = %#v", hash)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := HashSkillTree(root)
	if err != nil || changed.Digest == hash.Digest {
		t.Fatalf("tampered hash = %#v, err = %v", changed, err)
	}
}

func TestHashSkillTreePermitsInTreeSymlinkChain(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("file", filepath.Join(root, "link-b")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink("link-b", filepath.Join(root, "link-a")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := HashSkillTree(root); err != nil {
		t.Fatalf("in-tree symlink chain rejected: %v", err)
	}
}

func TestHashSkillTreeRejectsSymlinkAncestorHiddenByDotDot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink("escape/../file", filepath.Join(root, "tricky")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := HashSkillTree(root); err == nil {
		t.Fatal("symlink ancestor hidden by .. unexpectedly accepted")
	}
}

func TestHashSkillTreeRejectsSymlinkEscapeSpecialEntriesAndLimits(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := hashSkillTree(root, MaterializerLimits{MaxTreeBytes: 4}); err == nil || got.Digest != "" {
		t.Fatalf("byte limit result = %#v, err = %v", got, err)
	}
	if got, err := hashSkillTree(root, MaterializerLimits{MaxTreeEntries: 0}); err != nil || got.Digest == "" {
		t.Fatalf("default entry limit result = %#v, err = %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "deep", "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := hashSkillTree(root, MaterializerLimits{MaxTreeDepth: 1}); err == nil {
		t.Fatal("depth limit unexpectedly succeeded")
	}
	if _, err := hashSkillTree(root, MaterializerLimits{MaxTreeEntries: 1}); err == nil {
		t.Fatal("entry limit unexpectedly succeeded")
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "absolute-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := HashSkillTree(root); err == nil {
		t.Fatal("absolute symlink escape unexpectedly accepted")
	}
	if err := os.Remove(filepath.Join(root, "absolute-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(root, "relative-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := HashSkillTree(root); err == nil {
		t.Fatal("relative symlink escape unexpectedly accepted")
	}
	if createSpecialEntry(t, filepath.Join(root, "fifo")) {
		if _, err := HashSkillTree(root); err == nil {
			t.Fatal("special entry unexpectedly accepted")
		}
	}
}

func TestMaterializationVerifyAndCleanupLifecycle(t *testing.T) {
	runner := defaultMaterializeRunner()
	m, _ := testMaterializer(t, runner, MaterializerLimits{})
	plan, err := m.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("demo", "example/catalog")})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok := plan.SnapshotFor("demo")
	if !ok || snapshot.Path != filepath.Join(plan.Root(), ".agents", "skills", "demo") {
		t.Fatalf("snapshot = %#v, root = %q", snapshot, plan.Root())
	}
	if err := plan.Verify(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot.Path, "SKILL.md"), []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := plan.Verify(); err == nil {
		t.Fatal("tampering was not detected")
	}
	if err := plan.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(plan.Root()); !os.IsNotExist(err) {
		t.Fatalf("cleanup left original root: %v", err)
	}
	if err := os.Mkdir(plan.Root(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := plan.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(plan.Root()); err != nil {
		t.Fatalf("idempotent cleanup removed recreated path: %v", err)
	}
	if err := snapshot.Verify(); err == nil {
		t.Fatal("snapshot unexpectedly verified after cleanup")
	}
}

const (
	boundedRunnerHelperEnv  = "SJSKILLS_BOUNDED_RUNNER_HELPER"
	boundedRunnerHelperMode = "SJSKILLS_BOUNDED_RUNNER_MODE"
)

func TestBoundedExecRunnerHelperProcess(t *testing.T) {
	if os.Getenv(boundedRunnerHelperEnv) != "1" {
		runner := boundedExecRunner{limits: MaterializerLimits{MaxStdoutBytes: 8, MaxStderrBytes: 7}}
		output, err := runner.Run(context.Background(), os.Args[0], []string{"-test.run=TestBoundedExecRunnerHelperProcess"}, append(os.Environ(), boundedRunnerHelperEnv+"=1", boundedRunnerHelperMode+"=output"))
		if err == nil {
			t.Fatalf("oversized helper output unexpectedly succeeded: stdout=%q stderr=%q", output.Stdout, output.Stderr)
		}
		if len(output.Stdout) != 8 || len(output.Stderr) != 7 {
			t.Fatalf("retained output lengths = %d/%d, want 8/7", len(output.Stdout), len(output.Stderr))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		started := time.Now()
		_, err = runner.Run(ctx, os.Args[0], []string{"-test.run=TestBoundedExecRunnerHelperProcess"}, append(os.Environ(), boundedRunnerHelperEnv+"=1", boundedRunnerHelperMode+"=hang"))
		if err == nil {
			t.Fatal("canceled helper unexpectedly succeeded")
		}
		if elapsed := time.Since(started); elapsed > 2*time.Second {
			t.Fatalf("canceled helper returned after %s", elapsed)
		}
		return
	}

	switch os.Getenv(boundedRunnerHelperMode) {
	case "output":
		_, _ = os.Stdout.Write([]byte(strings.Repeat("o", 64)))
		_, _ = os.Stderr.Write([]byte(strings.Repeat("e", 64)))
	case "hang":
		child := exec.Command(os.Args[0], "-test.run=TestBoundedExecRunnerHelperProcess")
		child.Env = append(os.Environ(), boundedRunnerHelperEnv+"=1", boundedRunnerHelperMode+"=child")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		select {}
	case "child":
		time.Sleep(400 * time.Millisecond)
	}
}
