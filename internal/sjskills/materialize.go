package sjskills

// This file contains the boundary between the pure sjskills resolver and the
// third-party Skills CLI.  The adapter deliberately stops at a verified,
// temporary snapshot.  Placement, ownership, quarantine, and rollback belong
// to the reconciliation layer that consumes the snapshot.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"
)

const (
	// SkillsCLIVersion is intentionally a source-code constant.  Changing the
	// version changes the materialization process and therefore requires a
	// reviewed implementation change.
	SkillsCLIVersion = "1.5.23"
	SkillsCLICommand = "bunx"

	defaultMaterializeTimeout = 120 * time.Second
	defaultMaxStdoutBytes     = 16 * 1024 * 1024
	defaultMaxStderrBytes     = 4 * 1024 * 1024
	defaultMaxTreeBytes       = 128 * 1024 * 1024
	defaultMaxTreeEntries     = 100_000
	defaultMaxTreeDepth       = 64
	defaultMaxDiagnostic      = 8 * 1024
)

// ProcessResult is the bounded result returned by a SkillsCLI runner.  ExitCode
// is zero for a successful process.  A runner may return both a result and an
// error when the process exited unsuccessfully; the materializer includes the
// bounded output in its sanitized diagnostic.
type ProcessResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Runner is the only process dependency of Materializer.  command is always
// "bunx" for the production adapter; the versioned Skills CLI is the first
// argument.  env is a complete environment, preserving ordinary inherited
// values while replacing every home/config signal used by the CLI.
type Runner interface {
	Run(context.Context, string, []string, []string) (ProcessResult, error)
}

// TempRootFactory supplies a candidate directory for one plan. The adapter
// takes ownership only after verifying that it is a fresh, empty real
// directory. The default factory uses os.MkdirTemp; tests can inject a
// fixture directory without ever consulting a real user home.
type TempRootFactory func() (string, error)

// LookPathFunc is injected so preflight tests can model bunx availability
// without depending on the host installation.
type LookPathFunc func(string) (string, error)

// MaterializerLimits are hard bounds for one materialization plan.  Zero
// values are replaced by safe defaults by NewMaterializer.
type MaterializerLimits struct {
	CommandTimeout     time.Duration
	MaxStdoutBytes     int64
	MaxStderrBytes     int64
	MaxTreeBytes       int64
	MaxTreeEntries     int
	MaxTreeDepth       int
	MaxDiagnosticBytes int
}

// MaterializerConfig contains only process, environment, staging, and bound
// dependencies.  It intentionally has no inventory/apply/rollout knobs.
type MaterializerConfig struct {
	Runner Runner

	// TempRootFactory supplies a candidate directory. It is removed after the
	// session only once it passes the fresh-empty ownership check. Tests inject
	// a fixture here; production uses MkdirTemp.
	TempRootFactory TempRootFactory

	// BaseEnvironment is the inherited environment to preserve while replacing
	// the home, configuration, and common write-cache signals used by the CLI.
	// Empty means the current process environment.
	BaseEnvironment []string

	// LookPath is used for the explicit bunx availability preflight.  With a
	// custom Runner and no LookPath, the runner itself is the availability
	// oracle; production's default runner uses exec.LookPath.
	LookPath LookPathFunc

	// Platform is normally runtime.GOOS.  Tests may use "windows" to exercise
	// Windows HOME signal isolation on any host.
	Platform string
	Limits   MaterializerLimits
}

// Materializer invokes the pinned Skills CLI and returns verified staged
// snapshots.  A Materializer may be reused for multiple independent plans;
// snapshots are never reused across plans because moving remote sources must
// be revalidated for each plan.
type Materializer struct {
	runner   Runner
	tempRoot TempRootFactory
	baseEnv  []string
	lookPath LookPathFunc
	platform string
	limits   MaterializerLimits
}

// NewMaterializer constructs an isolated Skills CLI adapter.  No process,
// network, home, or filesystem operation occurs until a plan is materialized
// (apart from copying the inherited environment when no explicit environment
// was supplied).
func NewMaterializer(config MaterializerConfig) *Materializer {
	limits := config.Limits
	if limits.CommandTimeout <= 0 {
		limits.CommandTimeout = defaultMaterializeTimeout
	}
	if limits.MaxStdoutBytes <= 0 {
		limits.MaxStdoutBytes = defaultMaxStdoutBytes
	}
	if limits.MaxStderrBytes <= 0 {
		limits.MaxStderrBytes = defaultMaxStderrBytes
	}
	if limits.MaxTreeBytes <= 0 {
		limits.MaxTreeBytes = defaultMaxTreeBytes
	}
	if limits.MaxTreeEntries <= 0 {
		limits.MaxTreeEntries = defaultMaxTreeEntries
	}
	if limits.MaxTreeDepth <= 0 {
		limits.MaxTreeDepth = defaultMaxTreeDepth
	}
	if limits.MaxDiagnosticBytes <= 0 {
		limits.MaxDiagnosticBytes = defaultMaxDiagnostic
	}

	baseEnv := config.BaseEnvironment
	if baseEnv == nil {
		baseEnv = os.Environ()
	}
	baseEnv = append([]string(nil), baseEnv...)

	runner := config.Runner
	if runner == nil {
		runner = boundedExecRunner{limits: limits}
	}
	lookPath := config.LookPath
	if lookPath == nil {
		if config.Runner != nil {
			// A fake runner is intentionally self-contained.  The real runner
			// performs exec.LookPath before every command, so custom tests do
			// not need bunx installed just to exercise argv and env contracts.
			lookPath = func(name string) (string, error) { return name, nil }
		} else {
			lookPath = exec.LookPath
		}
	}

	tempRoot := config.TempRootFactory
	if tempRoot == nil {
		tempRoot = func() (string, error) {
			return os.MkdirTemp("", "sjskills-materialize-")
		}
	}

	platform := config.Platform
	if platform == "" {
		platform = runtime.GOOS
	}

	return &Materializer{
		runner:   runner,
		tempRoot: tempRoot,
		baseEnv:  baseEnv,
		lookPath: lookPath,
		platform: platform,
		limits:   limits,
	}
}

// Preflight verifies bunx and the exact pinned Skills CLI version without
// contacting a source.  It uses and cleans its own temporary root so callers
// can perform the check before deciding whether to build a plan.
func (m *Materializer) Preflight(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	root, err := m.newStagingRoot()
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	return m.preflightIn(ctx, root)
}

func (m *Materializer) preflightIn(ctx context.Context, root string) error {
	if _, err := m.lookPath(SkillsCLICommand); err != nil {
		return m.errorAt(root, "preflight", "bunx is unavailable", err)
	}

	// Keep the two checks separate.  `bunx --version` proves the pinned
	// launcher exists, while the second command proves that bunx resolved the
	// exact package version this adapter is built to use.  In particular, do
	// not collapse this into a package install invocation: callers rely on the
	// preflight being completed before any remote source is materialized.
	checks := []struct {
		args     []string
		name     string
		expected []byte
	}{
		{args: []string{"--version"}, name: "bunx"},
		{args: []string{"skills@" + SkillsCLIVersion, "--version"}, name: "Skills CLI", expected: []byte(SkillsCLIVersion + "\n")},
	}
	for _, check := range checks {
		result, err := m.run(ctx, root, check.args)
		if err != nil {
			return m.diagnosticErrorAt(root, "preflight", check.name+" version check failed", err, result)
		}
		if result.ExitCode != 0 {
			return m.diagnosticErrorAt(root, "preflight", fmt.Sprintf("%s version check exited with status %d", check.name, result.ExitCode), nil, result)
		}
		if len(check.expected) > 0 && !bytes.Equal(result.Stdout, check.expected) {
			return m.diagnosticErrorAt(root, "preflight", fmt.Sprintf("Skills CLI version output was not exactly %q", check.expected), nil, result)
		}
	}
	return nil
}

// skillsCLIAddArgs builds the only install command this adapter permits.
// Explicit skill selection and --global/--copy/--yes are always present;
// --full-depth is emitted only when requested.  No caller can introduce
// --all, an alternate package runner, or a different agent through this API.
func skillsCLIAddArgs(skill DesiredSkill) ([]string, error) {
	if skill.Name == "" || !isPortableName(skill.Name) {
		return nil, materializationError("command", "skill name is not portable", nil)
	}
	if problem := SkillsCLIPathProblem(skill.Source); problem != "" {
		return nil, materializationError("command", "source is not installable by Skills CLI: "+problem, nil)
	}
	args := []string{
		"skills@" + SkillsCLIVersion,
		"add",
		skill.Source,
		"--skill",
		skill.Name,
		"--copy",
		"--global",
		"--agent",
		"codex",
		"--yes",
	}
	if skill.FullDepth {
		args = append(args, "--full-depth")
	}
	return args, nil
}

// Materialize runs one plan.  Installable skills are keyed by identity (name,
// source, and full-depth behavior), so repeated desired placements share one
// verified snapshot and one subprocess invocation. Manual and workflow entries
// are returned as skipped and never reported as materialized successes.
func (m *Materializer) Materialize(ctx context.Context, skills []DesiredSkill) (*MaterializationPlan, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	installable, skipped, err := classifyMaterializationSkills(skills)
	if err != nil {
		return nil, m.sanitizeError(err, "")
	}
	plan := &MaterializationPlan{
		snapshots: make(map[string]*SkillSnapshot, len(installable)),
		skipped:   append([]DesiredSkill(nil), skipped...),
	}
	if len(installable) == 0 {
		return plan, nil
	}

	root, err := m.newStagingRoot()
	if err != nil {
		return nil, err
	}
	plan.root = root
	plan.cleanupRoot = root
	plan.limits = m.limits
	if err := m.preflightIn(ctx, root); err != nil {
		_ = plan.Cleanup()
		return nil, err
	}

	for _, skill := range installable {
		args, err := skillsCLIAddArgs(skill)
		if err != nil {
			_ = plan.Cleanup()
			return nil, m.errorAt(root, safeSkillName(skill.Name), "build Skills CLI command", err)
		}
		result, runErr := m.run(ctx, root, args)
		if runErr != nil {
			_ = plan.Cleanup()
			return nil, m.diagnosticErrorAt(root, safeSkillName(skill.Name), "Skills CLI command failed", runErr, result)
		}
		if result.ExitCode != 0 {
			_ = plan.Cleanup()
			return nil, m.diagnosticErrorAt(root, safeSkillName(skill.Name), fmt.Sprintf("Skills CLI command exited with status %d", result.ExitCode), nil, result)
		}
		path, err := locateStagedSkill(root, skill.Name)
		if err != nil {
			_ = plan.Cleanup()
			return nil, m.errorAt(root, safeSkillName(skill.Name), "locate staged skill", err)
		}
		digest, err := hashSkillTree(path, m.limits)
		if err != nil {
			_ = plan.Cleanup()
			return nil, m.errorAt(root, safeSkillName(skill.Name), "verify staged tree", err)
		}
		snapshot := &SkillSnapshot{
			Skill:     skill,
			Path:      path,
			Hash:      digest,
			plan:      plan,
			stageRoot: root,
			limits:    m.limits,
		}
		plan.snapshots[skill.Name] = snapshot
	}
	return plan, nil
}

func classifyMaterializationSkills(skills []DesiredSkill) ([]DesiredSkill, []DesiredSkill, error) {
	installable := make([]DesiredSkill, 0, len(skills))
	skipped := make([]DesiredSkill, 0)
	seen := make(map[string]DesiredSkill, len(skills))
	for _, skill := range skills {
		if skill.Name == "" || !isPortableName(skill.Name) {
			return nil, nil, materializationError("classify", "skill name is not portable", nil)
		}
		switch skill.Manager {
		case ManagerManual, ManagerWorkflow:
			if previous, ok := seen[skill.Name]; ok {
				if previous.Manager != skill.Manager {
					return nil, nil, materializationError("classify", "skill has contradictory managers", nil)
				}
				if !sameDesiredSkill(previous, skill) {
					return nil, nil, materializationError("classify", "skipped skill identity has contradictory source, workflow, or options", nil)
				}
				continue
			}
			if _, ok := seen[skill.Name]; !ok {
				seen[skill.Name] = skill
				skipped = append(skipped, skill)
			}
		case ManagerSkillsCLI:
			if previous, ok := seen[skill.Name]; ok {
				if previous.Manager != ManagerSkillsCLI || previous.Source != skill.Source || previous.FullDepth != skill.FullDepth {
					return nil, nil, materializationError("classify", "skill identity has contradictory source or options", nil)
				}
				// The exact identity was already scheduled.  Targets and mode
				// differ at placement time, not at materialization time.
				continue
			}
			if problem := SkillsCLIPathProblem(skill.Source); problem != "" {
				return nil, nil, materializationError(safeSkillName(skill.Name), "source is not installable by Skills CLI: "+problem, nil)
			}
			seen[skill.Name] = skill
			installable = append(installable, skill)
		default:
			return nil, nil, materializationError(safeSkillName(skill.Name), "unsupported installation manager", nil)
		}
	}
	return installable, skipped, nil
}

func sameDesiredSkill(left, right DesiredSkill) bool {
	if left.Name != right.Name || left.SourceID != right.SourceID || left.Source != right.Source ||
		left.Scope != right.Scope || left.Origin != right.Origin || left.Manager != right.Manager ||
		left.Mode != right.Mode || left.Workflow != right.Workflow || left.FullDepth != right.FullDepth ||
		len(left.Targets) != len(right.Targets) {
		return false
	}
	for index := range left.Targets {
		if left.Targets[index] != right.Targets[index] {
			return false
		}
	}
	return true
}

func (m *Materializer) newStagingRoot() (string, error) {
	root, err := m.tempRoot()
	if err != nil {
		return "", m.error("staging", "create temporary staging root", err)
	}
	if err := validateStagingRoot(root); err != nil {
		// Ownership is transferred only after the candidate is proven to be a
		// fresh, empty directory. Invalid candidates are never removed: a caller
		// may have supplied a shared sentinel, a filesystem root, or a path it
		// still owns.
		return "", m.sanitizeError(err, "")
	}
	root = filepath.Clean(root)
	if err := prepareStagingRoot(root); err != nil {
		// The empty-directory check transferred ownership immediately before
		// preparation, so this cleanup is confined to the now-owned root.
		_ = os.RemoveAll(root)
		return "", m.errorAt(root, "staging", "prepare isolated staging directories", err)
	}
	return root, nil
}

func validateStagingRoot(root string) error {
	if root == "" || strings.ContainsRune(root, 0) || !filepath.IsAbs(root) {
		return materializationError("staging", "temporary staging root must be an absolute path", nil)
	}
	clean := filepath.Clean(root)
	if clean == filepath.Dir(clean) {
		return materializationError("staging", "temporary staging root must not be a filesystem root", nil)
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return materializationError("staging", "temporary staging root is unavailable", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return materializationError("staging", "temporary staging root must be a real directory", nil)
	}
	entries, err := os.ReadDir(clean)
	if err != nil {
		return materializationError("staging", "temporary staging root cannot be inspected", err)
	}
	if len(entries) != 0 {
		return materializationError("staging", "temporary staging root must be empty", nil)
	}
	return nil
}

func (m *Materializer) run(ctx context.Context, root string, args []string) (ProcessResult, error) {
	if err := ensureContext(ctx); err != nil {
		return ProcessResult{}, err
	}
	commandCtx, cancel := context.WithTimeout(ctx, m.limits.CommandTimeout)
	defer cancel()
	env := isolatedEnvironment(m.baseEnv, root, m.platform)
	result, err := m.runner.Run(commandCtx, SkillsCLICommand, append([]string(nil), args...), env)
	stdoutExceeded := int64(len(result.Stdout)) > m.limits.MaxStdoutBytes
	stderrExceeded := int64(len(result.Stderr)) > m.limits.MaxStderrBytes
	if stdoutExceeded {
		// Do not carry an unbounded fake-runner result into diagnostics. The
		// production runner already retains only its configured bound, while a
		// test or alternate runner may return a larger slice.
		result.Stdout = nil
	}
	if stderrExceeded {
		result.Stderr = nil
	}
	if stdoutExceeded {
		return result, errors.New("process: stdout exceeded its bound")
	}
	if stderrExceeded {
		return result, errors.New("process: stderr exceeded its bound")
	}
	if commandCtx.Err() != nil {
		return result, commandCtx.Err()
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func ensureContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// isolatedEnvironment keeps ordinary inherited environment values (notably
// PATH and Git credential-helper variables) while replacing home/profile and
// common temp/cache/config signals that could make the CLI write outside the
// staging root.
func isolatedEnvironment(base []string, root, platform string) []string {
	values := make(map[string]string, len(base)+8)
	order := make([]string, 0, len(base)+8)
	for _, item := range base {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			continue
		}
		if _, exists := values[key]; !exists {
			order = append(order, key)
		}
		values[key] = value
	}
	windows := isWindowsPlatform(platform)
	redirects := isolatedPathValues(root)
	protected := make([]string, 0, len(redirects)+3)
	for key := range redirects {
		protected = append(protected, key)
	}
	protected = append(protected, "HOMEDRIVE", "HOMEPATH", "HOMESHARE")
	if windows {
		for key := range values {
			remove := false
			for _, candidate := range protected {
				if strings.EqualFold(key, candidate) {
					remove = true
					break
				}
			}
			if remove {
				delete(values, key)
			}
		}
		// Remove deleted spellings from the deterministic output order before
		// adding the canonical uppercase names below. Otherwise a Windows
		// environment containing `home` would receive duplicate HOME entries.
		order = order[:0]
		for key := range values {
			order = append(order, key)
		}
	}
	if !windows {
		// Unix environment names are case-sensitive, so only canonical names
		// need replacement. Windows-like environments were cleared above using
		// case-insensitive matching.
		for key := range redirects {
			delete(values, key)
		}
	}
	order = order[:0]
	for key := range values {
		order = append(order, key)
	}
	set := func(key, value string) {
		if _, exists := values[key]; !exists {
			order = append(order, key)
		}
		values[key] = value
	}
	for key, value := range redirects {
		set(key, value)
	}
	if windows {
		if drive, pathPart := windowsHomeParts(root); drive != "" {
			set("HOMEDRIVE", drive)
			set("HOMEPATH", pathPart)
		}
	}
	result := make([]string, 0, len(values))
	// Sorting makes fake-runner assertions deterministic while retaining all
	// ordinary inherited variables.
	sort.Strings(order)
	for _, key := range order {
		if value, ok := values[key]; ok {
			result = append(result, key+"="+value)
		}
	}
	return result
}

func isolatedPathValues(root string) map[string]string {
	return map[string]string{
		"HOME":                  root,
		"USERPROFILE":           root,
		"CODEX_HOME":            filepath.Join(root, ".agents"),
		"CLAUDE_CONFIG_DIR":     filepath.Join(root, ".claude"),
		"TMPDIR":                filepath.Join(root, ".tmp"),
		"TMP":                   filepath.Join(root, ".tmp"),
		"TEMP":                  filepath.Join(root, ".tmp"),
		"XDG_CACHE_HOME":        filepath.Join(root, ".cache"),
		"XDG_CONFIG_HOME":       filepath.Join(root, ".config"),
		"XDG_DATA_HOME":         filepath.Join(root, ".local", "share"),
		"NPM_CONFIG_CACHE":      filepath.Join(root, ".npm"),
		"BUN_INSTALL_CACHE_DIR": filepath.Join(root, ".bun", "install", "cache"),
	}
}

func prepareStagingRoot(root string) error {
	directories := []string{
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".claude"),
		filepath.Join(root, ".tmp"),
		filepath.Join(root, ".cache"),
		filepath.Join(root, ".config"),
		filepath.Join(root, ".local", "share"),
		filepath.Join(root, ".npm"),
		filepath.Join(root, ".bun", "install", "cache"),
	}
	for _, directory := range directories {
		if !pathWithin(root, directory) {
			return materializationError("staging", "isolated directory escapes staging root", nil)
		}
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return materializationError("staging", "isolated directory could not be created", err)
		}
		info, err := os.Lstat(directory)
		if err != nil {
			return materializationError("staging", "isolated directory is unavailable", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return materializationError("staging", "isolated directory is not a real directory", nil)
		}
	}
	return nil
}

func isWindowsPlatform(platform string) bool {
	return strings.EqualFold(platform, "windows") || strings.EqualFold(platform, "win32")
}

func windowsHomeParts(root string) (string, string) {
	if len(root) >= 2 && ((root[0] >= 'A' && root[0] <= 'Z') || (root[0] >= 'a' && root[0] <= 'z')) && root[1] == ':' {
		return root[:2], root[2:]
	}
	return "", ""
}

// MaterializationPlan owns the temporary root and every snapshot derived from
// it.  Cleanup is safe to call repeatedly, including through an individual
// snapshot, and removes only the factory-owned staging root.
type MaterializationPlan struct {
	root        string
	cleanupRoot string
	limits      MaterializerLimits
	snapshots   map[string]*SkillSnapshot
	skipped     []DesiredSkill
	cleanupMu   sync.Mutex
	cleanedFlag atomic.Bool
	cleanupErr  error
}

// Root returns the temporary plan root.  It is useful only while the plan is
// alive; callers must not retain it after Cleanup.
func (p *MaterializationPlan) Root() string {
	if p == nil {
		return ""
	}
	return p.root
}

// SnapshotFor returns the verified snapshot for one materialized skill.
func (p *MaterializationPlan) SnapshotFor(name string) (*SkillSnapshot, bool) {
	if p == nil {
		return nil, false
	}
	snapshot, ok := p.snapshots[name]
	return snapshot, ok
}

// Snapshots returns snapshots in deterministic skill-name order.
func (p *MaterializationPlan) Snapshots() []*SkillSnapshot {
	if p == nil {
		return nil
	}
	names := make([]string, 0, len(p.snapshots))
	for name := range p.snapshots {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]*SkillSnapshot, 0, len(names))
	for _, name := range names {
		result = append(result, p.snapshots[name])
	}
	return result
}

// Skipped returns manual/workflow entries that were deliberately left to their
// owning manager. The returned slice is detached from the plan.
func (p *MaterializationPlan) Skipped() []DesiredSkill {
	if p == nil {
		return nil
	}
	return append([]DesiredSkill(nil), p.skipped...)
}

// Verify detects any content, mode, path, or symlink-target change in every
// snapshot before a later placement operation consumes it.
func (p *MaterializationPlan) Verify() error {
	if p == nil {
		return materializationError("verify", "nil materialization plan", nil)
	}
	for _, snapshot := range p.Snapshots() {
		if err := snapshot.Verify(); err != nil {
			return sanitizeOwnedStageError(err, p.root, p.limits.MaxDiagnosticBytes)
		}
	}
	return nil
}

// Cleanup is idempotent.  A cleanup error is returned consistently on every
// subsequent call, and no second removal is attempted.
func (p *MaterializationPlan) Cleanup() error {
	if p == nil {
		return nil
	}
	p.cleanupMu.Lock()
	defer p.cleanupMu.Unlock()
	if p.cleanedFlag.Load() {
		return p.cleanupErr
	}
	// Publish the terminal lifecycle state before removal so Verify cannot
	// mistake a concurrently recreated path for the original snapshot.
	p.cleanedFlag.Store(true)
	if p.cleanupRoot != "" {
		p.cleanupErr = sanitizeOwnedStageError(os.RemoveAll(p.cleanupRoot), p.cleanupRoot, p.limits.MaxDiagnosticBytes)
	}
	return p.cleanupErr
}

// TreeHash is the reconciler-owned digest representation.  Algorithm is
// always tree-sha256-v2 for snapshots produced here.
type TreeHash struct {
	Algorithm string `json:"algorithm"`
	Digest    string `json:"digest"`
}

// SkillSnapshot is one verified staged tree.  Path remains inside the plan's
// isolated Codex target and is valid until the owning plan is cleaned up.
type SkillSnapshot struct {
	Skill DesiredSkill
	Path  string
	Hash  TreeHash

	plan      *MaterializationPlan
	stageRoot string
	limits    MaterializerLimits
}

// Verify rewalks the tree without following symlinks and compares its current
// digest and bounds to the captured digest.
func (s *SkillSnapshot) Verify() error {
	if s == nil {
		return materializationError("verify", "nil skill snapshot", nil)
	}
	if s.plan != nil && s.plan.cleaned() {
		return s.sanitizeError(materializationError(safeSkillName(s.Skill.Name), "snapshot has been cleaned up", nil))
	}
	if err := validateSnapshotPath(s.stageRoot, s.Path); err != nil {
		return s.sanitizeError(err)
	}
	expected := filepath.Join(s.stageRoot, ".agents", "skills", s.Skill.Name)
	if filepath.Clean(s.Path) != filepath.Clean(expected) {
		return s.sanitizeError(materializationError(safeSkillName(s.Skill.Name), "snapshot path is not the requested staged path", nil))
	}
	digest, err := hashSkillTree(s.Path, s.limits)
	if err != nil {
		return s.sanitizeError(fmt.Errorf("verify %s: %w", safeSkillName(s.Skill.Name), err))
	}
	if digest != s.Hash {
		return s.sanitizeError(materializationError(safeSkillName(s.Skill.Name), "staged snapshot content changed", nil))
	}
	return nil
}

func (s *SkillSnapshot) sanitizeError(err error) error {
	if s == nil {
		return err
	}
	return sanitizeOwnedStageError(err, s.stageRoot, s.limits.MaxDiagnosticBytes)
}

// Cleanup delegates to the owning plan so callers can use either lifecycle
// handle without accidentally removing only part of the shared staging tree.
func (s *SkillSnapshot) Cleanup() error {
	if s == nil || s.plan == nil {
		return nil
	}
	return s.plan.Cleanup()
}

func (p *MaterializationPlan) cleaned() bool {
	if p == nil {
		return true
	}
	return p.cleanedFlag.Load()
}

func validateSnapshotPath(stageRoot, skillPath string) error {
	if stageRoot == "" || skillPath == "" || !filepath.IsAbs(stageRoot) || !filepath.IsAbs(skillPath) {
		return materializationError("verify", "snapshot path is not absolute", nil)
	}
	if !pathWithin(stageRoot, skillPath) {
		return materializationError("verify", "snapshot path escapes staging root", nil)
	}
	if err := validateNoSymlinkPath(stageRoot, skillPath); err != nil {
		return err
	}
	info, err := os.Lstat(skillPath)
	if err != nil {
		return materializationError("verify", "snapshot path is missing", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return materializationError("verify", "snapshot path is not a real directory", nil)
	}
	return nil
}

func locateStagedSkill(root, name string) (string, error) {
	if name == "" || !isPortableName(name) {
		return "", materializationError("staging", "skill name is not portable", nil)
	}
	codexRoot := filepath.Join(root, ".agents")
	skillsRoot := filepath.Join(codexRoot, "skills")
	if err := validateNoSymlinkPath(root, skillsRoot); err != nil {
		return "", err
	}
	for label, path := range map[string]string{"Codex home": codexRoot, "Codex skills root": skillsRoot} {
		info, err := os.Lstat(path)
		if err != nil {
			return "", materializationError("staging", label+" is missing", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", materializationError("staging", label+" is malformed", nil)
		}
	}
	expected := filepath.Join(skillsRoot, name)
	if !pathWithin(root, expected) {
		return "", materializationError("staging", "skill path escapes staging root", nil)
	}

	info, err := os.Lstat(expected)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", materializationError("staging", "Skills CLI did not materialize the requested skill", nil)
		}
		return "", materializationError("staging", "staged skill disappeared", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", materializationError("staging", "staged skill must be a real directory", nil)
	}
	return expected, nil
}

// HashSkillTree computes the exact tree-sha256-v2 digest used by the legacy
// reconciler.  It rejects malformed trees and symlink escapes, and never
// follows a symlink while walking.
func HashSkillTree(root string) (TreeHash, error) {
	return hashSkillTree(root, MaterializerLimits{})
}

func hashSkillTree(root string, limits MaterializerLimits) (TreeHash, error) {
	limits = normalizedLimits(limits)
	if err := validateStagingTreeRoot(root); err != nil {
		return TreeHash{}, err
	}
	hash := sha256.New()
	stats := treeStats{}
	if err := hashPathInto(hash, root, ".", root, 0, limits, &stats); err != nil {
		return TreeHash{}, err
	}
	return TreeHash{Algorithm: TreeHashAlgorithmSHA256V2, Digest: hex.EncodeToString(hash.Sum(nil))}, nil
}

type treeStats struct {
	bytes   int64
	entries int
}

func hashPathInto(hash io.Writer, path, relative, root string, depth int, limits MaterializerLimits, stats *treeStats) error {
	info, err := os.Lstat(path)
	if err != nil {
		return materializationError("tree", "staged entry is unreadable", err)
	}
	if depth > limits.MaxTreeDepth {
		return materializationError("tree", "staged path depth exceeded its bound", nil)
	}
	if relative != "." {
		stats.entries++
		if stats.entries > limits.MaxTreeEntries {
			return materializationError("tree", "staged entry count exceeded its bound", nil)
		}
	}
	portable := filepath.ToSlash(relative)
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return materializationError("tree", "staged symlink target is unreadable", err)
		}
		if err := validateSymlinkTarget(path, target, root); err != nil {
			return err
		}
		updateHashField(hash, "symlink")
		updateHashField(hash, portable)
		updateHashField(hash, target)
		return nil
	}
	if info.Mode().IsRegular() {
		updateHashField(hash, "file")
		updateHashField(hash, portable)
		if info.Mode().Perm()&0o111 == 0 {
			updateHashField(hash, "non-executable")
		} else {
			updateHashField(hash, "executable")
		}
		content, err := readBoundedFile(path, limits.MaxTreeBytes-stats.bytes)
		if err != nil {
			return err
		}
		if err := addTreeBytes(stats, int64(len(content)), limits); err != nil {
			return err
		}
		updateHashField(hash, content)
		return nil
	}
	if !info.IsDir() {
		return materializationError("tree", "staged tree contains an unsupported entry type", nil)
	}
	updateHashField(hash, "directory")
	updateHashField(hash, portable)
	entries, err := os.ReadDir(path)
	if err != nil {
		return materializationError("tree", "staged directory is unreadable", err)
	}
	sort.Slice(entries, func(i, j int) bool { return compareUTF16(entries[i].Name(), entries[j].Name()) < 0 })
	for _, entry := range entries {
		child := filepath.Join(path, entry.Name())
		if !pathWithin(root, child) {
			return materializationError("tree", "staged entry escapes the skill tree", nil)
		}
		if err := hashPathInto(hash, child, filepath.Join(relative, entry.Name()), root, depth+1, limits, stats); err != nil {
			return err
		}
	}
	return nil
}

func readBoundedFile(path string, remaining int64) ([]byte, error) {
	if remaining < 0 {
		return nil, materializationError("tree", "staged tree bytes exceeded their bound", nil)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, materializationError("tree", "staged file is unreadable", err)
	}
	defer file.Close()
	readLimit := remaining
	if remaining < int64(^uint64(0)>>1) {
		readLimit++
	}
	reader := io.LimitReader(file, readLimit)
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, materializationError("tree", "staged file could not be read", err)
	}
	if int64(len(content)) > remaining {
		return nil, materializationError("tree", "staged tree bytes exceeded their bound", nil)
	}
	return content, nil
}

func addTreeBytes(stats *treeStats, count int64, limits MaterializerLimits) error {
	if count < 0 || stats.bytes > limits.MaxTreeBytes-count {
		return materializationError("tree", "staged tree bytes exceeded their bound", nil)
	}
	stats.bytes += count
	return nil
}

func updateHashField(hash io.Writer, value any) {
	var data []byte
	switch typed := value.(type) {
	case string:
		data = []byte(typed)
	case []byte:
		data = typed
	default:
		data = []byte(fmt.Sprint(typed))
	}
	_, _ = io.WriteString(hash, strconv.Itoa(len(data)))
	_, _ = io.WriteString(hash, ":")
	_, _ = hash.Write(data)
}

func validateStagingTreeRoot(root string) error {
	if root == "" || !filepath.IsAbs(root) {
		return materializationError("tree", "skill tree root must be an absolute path", nil)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return materializationError("tree", "skill tree root is missing", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return materializationError("tree", "skill tree root must be a real directory", nil)
	}
	return nil
}

func validateSymlinkTarget(linkPath, target, root string) error {
	return validateSymlinkChain(linkPath, target, root, make(map[string]struct{}))
}

func validateSymlinkChain(linkPath, target, root string, seen map[string]struct{}) error {
	if target == "" {
		return materializationError("tree", "staged symlink target is empty", nil)
	}
	if isAnyAbsolutePath(target) {
		return materializationError("tree", "absolute staged symlink targets are not allowed", nil)
	}
	linkPath = filepath.Clean(linkPath)
	if _, ok := seen[linkPath]; ok {
		return materializationError("tree", "staged symlink cycle is not allowed", nil)
	}
	seen[linkPath] = struct{}{}
	resolved, err := resolveSymlinkTargetPath(linkPath, target, root)
	if err != nil {
		return err
	}
	// Every existing component before the final target must be a real
	// directory. The final entry may be a symlink, in which case inspect its
	// raw target recursively without following it for hashing.
	if err := validateSymlinkParents(root, resolved); err != nil {
		return fmt.Errorf("validate symlink %s: %w", filepath.Base(linkPath), err)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return materializationError("tree", "staged symlink target is missing", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	next, err := os.Readlink(resolved)
	if err != nil {
		return materializationError("tree", "staged symlink target is unreadable", err)
	}
	return validateSymlinkChain(resolved, next, root, seen)
}

func resolveSymlinkTargetPath(linkPath, target, root string) (string, error) {
	current := filepath.Dir(linkPath)
	if !pathWithin(root, current) {
		return "", materializationError("tree", "staged symlink escapes the skill tree", nil)
	}
	if filepath.Separator == '\\' {
		target = strings.ReplaceAll(target, "/", `\`)
	}
	parts := strings.Split(target, string(filepath.Separator))
	for index, component := range parts {
		if component == "" {
			continue
		}
		if component == "." {
			continue
		}
		if component == ".." {
			current = filepath.Dir(current)
			if !pathWithin(root, current) {
				return "", materializationError("tree", "staged symlink escapes the skill tree", nil)
			}
			continue
		}
		next := filepath.Join(current, component)
		if !pathWithin(root, next) {
			return "", materializationError("tree", "staged symlink escapes the skill tree", nil)
		}
		info, err := os.Lstat(next)
		if err != nil {
			return "", materializationError("tree", "staged symlink target is missing", err)
		}
		if hasSymlinkTargetComponentAfter(parts, index) {
			if info.Mode()&os.ModeSymlink != 0 {
				return "", materializationError("tree", "staged symlink target traverses a symlink", nil)
			}
			if !info.IsDir() {
				return "", materializationError("tree", "staged symlink target ancestor is not a directory", nil)
			}
		}
		current = next
	}
	return filepath.Clean(current), nil
}

func hasSymlinkTargetComponentAfter(parts []string, index int) bool {
	for _, component := range parts[index+1:] {
		if component != "" {
			return true
		}
	}
	return false
}

func validateSymlinkParents(boundary, candidate string) error {
	if !pathWithin(boundary, candidate) {
		return materializationError("path", "path escapes its boundary", nil)
	}
	relative, err := filepath.Rel(boundary, candidate)
	if err != nil {
		return materializationError("path", "path relationship is unavailable", err)
	}
	if err := checkRealDirectory(boundary); err != nil {
		return err
	}
	if relative == "." {
		return nil
	}
	current := boundary
	components := strings.Split(relative, string(filepath.Separator))
	for index, component := range components {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return materializationError("tree", "staged symlink target is missing", err)
		}
		if index < len(components)-1 {
			if info.Mode()&os.ModeSymlink != 0 {
				return materializationError("tree", "staged symlink target traverses a symlink", nil)
			}
			if !info.IsDir() {
				return materializationError("tree", "staged symlink target ancestor is not a directory", nil)
			}
		}
	}
	return nil
}

func isAnyAbsolutePath(value string) bool {
	if filepath.IsAbs(value) || strings.HasPrefix(value, `\`) || strings.HasPrefix(value, `/`) {
		return true
	}
	return len(value) >= 2 && ((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) && value[1] == ':'
}

func validateNoSymlinkPath(boundary, candidate string) error {
	if !pathWithin(boundary, candidate) {
		return materializationError("path", "path escapes its boundary", nil)
	}
	relative, err := filepath.Rel(boundary, candidate)
	if err != nil {
		return materializationError("path", "path relationship is unavailable", err)
	}
	current := boundary
	if err := checkRealDirectory(current); err != nil {
		return err
	}
	if relative == "." {
		return nil
	}
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// Missing ancestors are allowed while the CLI is about to
				// create them; callers that need existence check the final path.
				return nil
			}
			return materializationError("path", "path ancestor is unreadable", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return materializationError("path", "path traverses a symlink", nil)
		}
		if component != filepath.Base(candidate) && !info.IsDir() {
			return materializationError("path", "path ancestor is not a directory", nil)
		}
	}
	return nil
}

func checkRealDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return materializationError("path", "path boundary is unreadable", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return materializationError("path", "path boundary is not a real directory", nil)
	}
	return nil
}

func pathWithin(parent, candidate string) bool {
	if !filepath.IsAbs(parent) || !filepath.IsAbs(candidate) {
		return false
	}
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(candidate))
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative))
}

func compareUTF16(left, right string) int {
	a, b := utf16.Encode([]rune(left)), utf16.Encode([]rune(right))
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func normalizedLimits(limits MaterializerLimits) MaterializerLimits {
	if limits.CommandTimeout <= 0 {
		limits.CommandTimeout = defaultMaterializeTimeout
	}
	if limits.MaxStdoutBytes <= 0 {
		limits.MaxStdoutBytes = defaultMaxStdoutBytes
	}
	if limits.MaxStderrBytes <= 0 {
		limits.MaxStderrBytes = defaultMaxStderrBytes
	}
	if limits.MaxTreeBytes <= 0 {
		limits.MaxTreeBytes = defaultMaxTreeBytes
	}
	if limits.MaxTreeEntries <= 0 {
		limits.MaxTreeEntries = defaultMaxTreeEntries
	}
	if limits.MaxTreeDepth <= 0 {
		limits.MaxTreeDepth = defaultMaxTreeDepth
	}
	if limits.MaxDiagnosticBytes <= 0 {
		limits.MaxDiagnosticBytes = defaultMaxDiagnostic
	}
	return limits
}

type boundedExecRunner struct {
	limits MaterializerLimits
}

const boundedExecWaitDelay = 250 * time.Millisecond

func (r boundedExecRunner) Run(ctx context.Context, command string, args []string, env []string) (ProcessResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = env
	cmd.Stdin = nil
	// A descendant can inherit stdout/stderr and keep those pipes open after
	// CommandContext kills the direct child. WaitDelay closes the pipes after a
	// short grace period so cancellation cannot leave this adapter blocked in
	// Wait forever.
	cmd.WaitDelay = boundedExecWaitDelay
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return ProcessResult{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return ProcessResult{}, err
	}
	if err := cmd.Start(); err != nil {
		return ProcessResult{}, err
	}
	var stdout, stderr boundedBuffer
	stdout.limit = r.limits.MaxStdoutBytes
	stderr.limit = r.limits.MaxStderrBytes
	copyDone := make(chan struct{}, 2)
	go func() { copyBounded(&stdout, stdoutPipe); copyDone <- struct{}{} }()
	go func() { copyBounded(&stderr, stderrPipe); copyDone <- struct{}{} }()
	waitErr := cmd.Wait()
	<-copyDone
	<-copyDone
	result := ProcessResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if stdout.exceeded {
		return result, materializationError("process", "stdout exceeded its bound", nil)
	}
	if stderr.exceeded {
		return result, materializationError("process", "stderr exceeded its bound", nil)
	}
	if waitErr == nil {
		return result, nil
	}
	if exitError, ok := waitErr.(*exec.ExitError); ok {
		result.ExitCode = exitError.ExitCode()
	}
	return result, waitErr
}

func copyBounded(destination *boundedBuffer, source io.Reader) {
	buffer := make([]byte, 32*1024)
	for {
		count, err := source.Read(buffer)
		if count > 0 {
			_, _ = destination.Write(buffer[:count])
		}
		if err != nil {
			return
		}
	}
}

type boundedBuffer struct {
	bytes.Buffer
	limit    int64
	exceeded bool
}

func (b *boundedBuffer) Write(data []byte) (int, error) {
	if b.limit < 0 {
		b.exceeded = true
		return len(data), nil
	}
	remaining := b.limit - int64(b.Len())
	if remaining <= 0 {
		b.exceeded = len(data) > 0
		return len(data), nil
	}
	if int64(len(data)) > remaining {
		_, _ = b.Buffer.Write(data[:remaining])
		b.exceeded = true
		return len(data), nil
	}
	_, _ = b.Buffer.Write(data)
	return len(data), nil
}

var (
	urlDiagnosticPattern    = regexp.MustCompile(`(?i)\b(https?|ssh|git)://[^\s<>"']+`)
	credentialPattern       = regexp.MustCompile(`(?i)\b(?:authorization|bearer|token|access[_-]?token|password|passwd|secret|api[_-]?key)\b(?:\s*[:=]\s*|\s+)(?:bearer\s+)?[^\s,;]+`)
	credentialPrefixPattern = regexp.MustCompile(`(?i)\b(?:ghp_|github_pat_|glpat-|xox[baprs]-|sk-)[A-Za-z0-9._~+/=-]+`)
)

func (m *Materializer) diagnosticErrorAt(root, scope, message string, processErr error, result ProcessResult) error {
	parts := make([]string, 0, 2)
	if scope != "" {
		parts = append(parts, scope)
	}
	parts = append(parts, message)
	if processErr != nil {
		parts = append(parts, processErr.Error())
	}
	if output := diagnosticText(result); output != "" {
		parts = append(parts, output)
	}
	return errors.New(sanitizeOwnedStageDiagnostic(strings.Join(parts, ": "), root, m.limits.MaxDiagnosticBytes))
}

func diagnosticText(result ProcessResult) string {
	parts := make([]string, 0, 2)
	if len(result.Stderr) > 0 {
		parts = append(parts, string(result.Stderr))
	}
	if len(result.Stdout) > 0 {
		parts = append(parts, string(result.Stdout))
	}
	return strings.Join(parts, " ")
}

func redactDiagnostic(value string, limit int) string {
	value = urlDiagnosticPattern.ReplaceAllStringFunc(value, func(candidate string) string {
		parsed, err := parseDiagnosticURL(candidate)
		if err != nil {
			return "<redacted-url>"
		}
		return parsed
	})
	value = credentialPattern.ReplaceAllString(value, "<redacted-secret>")
	value = credentialPrefixPattern.ReplaceAllString(value, "<redacted-token>")
	return truncateDiagnostic(value, limit)
}

func parseDiagnosticURL(value string) (string, error) {
	colon := strings.IndexByte(value, ':')
	if colon <= 0 {
		return "", errors.New("not a URL")
	}
	scheme := value[:colon]
	rest := value[colon+1:]
	if !strings.HasPrefix(rest, "//") {
		return scheme + "://<redacted-url>", nil
	}
	rest = strings.TrimPrefix(rest, "//")
	end := len(rest)
	for index, character := range rest {
		if character == '/' || character == '?' || character == '#' {
			end = index
			break
		}
	}
	host := rest[:end]
	if at := strings.LastIndexByte(host, '@'); at >= 0 {
		host = host[at+1:]
	}
	path := ""
	if end < len(rest) && rest[end] == '/' {
		pathEnd := len(rest)
		if query := strings.IndexAny(rest[end:], "?#"); query >= 0 {
			pathEnd = end + query
		}
		path = rest[end:pathEnd]
	}
	return scheme + "://" + host + path, nil
}

func truncateDiagnostic(value string, limit int) string {
	if limit <= 0 {
		limit = defaultMaxDiagnostic
	}
	data := []byte(value)
	if len(data) <= limit {
		return value
	}
	if limit <= len("...[truncated]") {
		return string(data[:limit])
	}
	return string(data[:limit-len("...[truncated]")]) + "...[truncated]"
}

func materializationError(scope, message string, cause error) error {
	return materializationErrorLimit(scope, message, cause, defaultMaxDiagnostic)
}

func materializationErrorLimit(scope, message string, cause error, limit int) error {
	parts := make([]string, 0, 2)
	if scope != "" {
		parts = append(parts, scope)
	}
	parts = append(parts, message)
	if cause != nil {
		parts = append(parts, cause.Error())
	}
	return errors.New(redactDiagnostic(strings.Join(parts, ": "), limit))
}

func (m *Materializer) error(scope, message string, cause error) error {
	return materializationErrorForRoot(scope, message, cause, "", m.limits.MaxDiagnosticBytes)
}

func (m *Materializer) errorAt(root, scope, message string, cause error) error {
	return materializationErrorForRoot(scope, message, cause, root, m.limits.MaxDiagnosticBytes)
}

func (m *Materializer) sanitizeError(err error, root string) error {
	return sanitizeOwnedStageError(err, root, m.limits.MaxDiagnosticBytes)
}

func materializationErrorForRoot(scope, message string, cause error, root string, limit int) error {
	parts := make([]string, 0, 2)
	if scope != "" {
		parts = append(parts, scope)
	}
	parts = append(parts, message)
	if cause != nil {
		parts = append(parts, cause.Error())
	}
	return errors.New(sanitizeOwnedStageDiagnostic(strings.Join(parts, ": "), root, limit))
}

func sanitizeOwnedStageError(err error, root string, limit int) error {
	if err == nil {
		return nil
	}
	return errors.New(sanitizeOwnedStageDiagnostic(err.Error(), root, limit))
}

func sanitizeOwnedStageDiagnostic(value, root string, limit int) string {
	return redactDiagnostic(redactOwnedStagePaths(value, root), limit)
}

var stagingRootPathPattern = regexp.MustCompile(`(?i)sjskills-materialize-[A-Za-z0-9._-]*`)

func redactOwnedStagePaths(value, root string) string {
	for _, variant := range stageRootVariants(root) {
		pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(variant))
		value = pattern.ReplaceAllString(value, "<staging-root>")
	}
	return stagingRootPathPattern.ReplaceAllString(value, "<staging-root>")
}

func stageRootVariants(root string) []string {
	if root == "" {
		return nil
	}
	seen := make(map[string]struct{}, 10)
	add := func(value string) {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	add(root)
	add(filepath.Clean(root))
	add(filepath.ToSlash(root))
	add(filepath.FromSlash(root))
	add(strings.ReplaceAll(root, "/", `\`))
	add(strings.ReplaceAll(root, `\`, "/"))
	for value := range seen {
		add(strings.ReplaceAll(value, `\`, `\\`))
	}
	variants := make([]string, 0, len(seen))
	for value := range seen {
		variants = append(variants, value)
	}
	sort.Slice(variants, func(i, j int) bool {
		if len(variants[i]) != len(variants[j]) {
			return len(variants[i]) > len(variants[j])
		}
		return variants[i] < variants[j]
	})
	return variants
}

func safeSkillName(name string) string {
	if name == "" {
		return "skill"
	}
	if isPortableName(name) {
		return name
	}
	return "skill"
}
