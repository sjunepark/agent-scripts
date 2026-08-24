// Package sjskills contains the pure desired-state contracts used by the
// sjskills command.  Parsing and resolution deliberately do not know how a
// skill is materialized or where a managed root lives.
package sjskills

import (
	"fmt"
	"sort"
	"strings"
)

const (
	RegistryVersion = 4
	ManifestVersion = 1
	ToolVersion     = "1.0.0"
)

// Target is a supported installation target. Targets are intentionally not
// harness names: the two supported project/global roots are the public model.
type Target string

const (
	TargetAgents Target = ".agents"
	TargetClaude Target = ".claude"
)

// Manager identifies who is responsible for provisioning a desired skill.
type Manager string

const (
	ManagerSkillsCLI Manager = "skills-cli"
	ManagerManual    Manager = "manual"
	ManagerWorkflow  Manager = "workflow"
	ManagerNone      Manager = "none"
)

// InstallMode controls reconciler-owned placement semantics. Version 1 uses
// verified directory copies exclusively.
type InstallMode string

const (
	ModeCopy InstallMode = "copy"
)

// SourceKind distinguishes a repository-backed catalog from an external
// source. A source location is optional for manual entries because their
// provisioning procedure is intentionally owned elsewhere.
type SourceKind string

const (
	SourceRepository SourceKind = "repository"
	SourceExternal   SourceKind = "external"
)

// Scope is the scope of a resolved desired skill. Registry declarations are
// single-source records; baseline/profile membership determines their use.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// Registry is the strict version-4 desired-state catalog. The Skills slice is
// the sole declaration of a skill's source and installation metadata. Global
// and profile fields contain only names referring to those declarations.
type Registry struct {
	Version          int                 `json:"version"`
	Description      string              `json:"description"`
	Defaults         RegistryDefaults    `json:"defaults"`
	Global           GlobalRegistry      `json:"global"`
	Profiles         map[string]Profile  `json:"profiles"`
	Sources          map[string]Source   `json:"sources"`
	TargetExceptions map[string][]Target `json:"targetExceptions,omitempty"`
	Skills           []SkillDeclaration  `json:"skills"`
}

// RegistryDefaults contains the exact default target set. Exceptions are
// kept centrally so a normal manifest never has to select harnesses.
type RegistryDefaults struct {
	Targets []Target `json:"targets"`
}

type GlobalRegistry struct {
	Baseline []string `json:"baseline"`
}

type Profile struct {
	Skills []string `json:"skills"`
}

// Source describes a source/catalog record. Source IDs are used only by the
// central registry; direct manifest declarations carry a location directly.
type Source struct {
	Kind        SourceKind `json:"kind"`
	Location    string     `json:"location,omitempty"`
	CatalogPath string     `json:"catalogPath,omitempty"`
}

// SkillDeclaration is the one central declaration for a skill. A missing
// target exception means Registry.Defaults.Targets. A missing fullDepth is
// false; all other manager-specific fields are validated strictly.
type SkillDeclaration struct {
	Name      string      `json:"name"`
	Source    string      `json:"source"`
	Manager   Manager     `json:"manager"`
	Mode      InstallMode `json:"mode,omitempty"`
	Workflow  string      `json:"workflow,omitempty"`
	FullDepth bool        `json:"fullDepth,omitempty"`
}

// DirectSkill is a project-local third-party declaration. Version 1 fixes its
// ownership to Skills CLI, its placement to copy mode, and its targets to the
// registry defaults; the manifest carries only a portable name, installable
// source, and optional full-depth hint.
type DirectSkill struct {
	Name      string `toml:"name" json:"name"`
	Source    string `toml:"source" json:"source"`
	FullDepth bool   `toml:"full_depth" json:"fullDepth,omitempty"`
}

// Manifest is the committed project intent. Project scope is implicit: the
// selected profile and direct sets are never global installations.
type Manifest struct {
	Version  int           `toml:"version" json:"version"`
	Profiles []string      `toml:"profiles" json:"profiles"`
	Direct   []DirectSkill `toml:"direct" json:"direct,omitempty"`
}

// DesiredSkill is a fully resolved, deterministic desired entry. It carries
// manager boundaries explicitly so later adapters cannot claim a manual or
// workflow entry was installed.
type DesiredSkill struct {
	Name      string      `json:"name"`
	SourceID  string      `json:"sourceId,omitempty"`
	Source    string      `json:"source"`
	Scope     Scope       `json:"scope"`
	Origin    string      `json:"origin"`
	Manager   Manager     `json:"manager"`
	Mode      InstallMode `json:"mode,omitempty"`
	Workflow  string      `json:"workflow,omitempty"`
	FullDepth bool        `json:"fullDepth"`
	Targets   []Target    `json:"targets"`
}

type DesiredState struct {
	Scope  Scope          `json:"scope"`
	Skills []DesiredSkill `json:"skills"`
}

// ResolveRequest is the pure boundary consumed by planning adapters. Global
// requests ignore Manifest and always select the fixed baseline; project
// requests require it and select only its profile/direct union.
type ResolveRequest struct {
	Registry Registry
	Manifest *Manifest
	Global   bool
}

// PlanAction is the stable action vocabulary shared by pure resolution,
// inventory-backed planning, and execution.
type PlanAction string

const (
	PlanActionInstall    PlanAction = "install"
	PlanActionUpdate     PlanAction = "update"
	PlanActionQuarantine PlanAction = "quarantine"
	PlanActionUnchanged  PlanAction = "unchanged"
	PlanActionManual     PlanAction = "manual"
	PlanActionWorkflow   PlanAction = "workflow"
	PlanActionBlocked    PlanAction = "blocked"
)

// PlanEvidence is deliberately source-agnostic. Inventory-backed project plans
// populate Current and verified materialization populates Expected.
type PlanEvidence struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// PlanOperation is one future reconciliation action. SourceID and Source
// preserve central/direct provenance, while Current and Expected carry bounded
// exact-state evidence.
type PlanOperation struct {
	Action   PlanAction   `json:"action"`
	Skill    string       `json:"skill"`
	Target   Target       `json:"target"`
	Manager  Manager      `json:"manager"`
	SourceID string       `json:"sourceId"`
	Source   string       `json:"source"`
	Reason   string       `json:"reason"`
	Current  PlanEvidence `json:"current"`
	Expected PlanEvidence `json:"expected"`
}

// Plan is the stable plan payload. Operations is non-nil even when the pure
// resolver has no filesystem inventory and therefore emits no operations.
// Warnings and Evidence remain process-level concerns and are copied to the
// outer envelope by the CLI; they are not duplicated inside this payload.
type Plan struct {
	Desired    DesiredState    `json:"desired"`
	Operations []PlanOperation `json:"operations"`
	Warnings   []Warning       `json:"-"`
	Evidence   []Evidence      `json:"-"`
}

// CommandOperation names the user-facing CLI operation. It is intentionally
// distinct from PlanAction: a command such as plan or apply may carry a plan
// containing any of the reconciliation actions above.
type CommandOperation string

const (
	CommandOperationInit     CommandOperation = "init"
	CommandOperationProfiles CommandOperation = "profiles"
	CommandOperationPlan     CommandOperation = "plan"
	CommandOperationApply    CommandOperation = "apply"
	CommandOperationRestore  CommandOperation = "restore"
	CommandOperationParse    CommandOperation = "parse"
	CommandOperationVersion  CommandOperation = "version"
)

type Result string

const (
	ResultSuccess     Result = "success"
	ResultInvalid     Result = "invalid"
	ResultUnavailable Result = "unavailable"
	ResultConflict    Result = "conflict"
)

// ExitStatus is intentionally small and stable. 2 means a valid invocation
// could not establish/apply the requested state; 64 and 65 distinguish usage
// from validly parsed but invalid input.
type ExitStatus int

const (
	ExitSuccess           ExitStatus = 0
	ExitExecutionFailure  ExitStatus = 2
	ExitInvalidInvocation ExitStatus = 64
	ExitInvalidInput      ExitStatus = 65
)

type IssueCode string

const (
	IssueMalformedInput   IssueCode = "malformed_input"
	IssueUnknownField     IssueCode = "unknown_field"
	IssueInvalidName      IssueCode = "invalid_name"
	IssueInvalidSource    IssueCode = "invalid_source"
	IssueInvalidTarget    IssueCode = "invalid_target"
	IssueInvalidManager   IssueCode = "invalid_manager"
	IssueInvalidMode      IssueCode = "invalid_mode"
	IssueEmptySelection   IssueCode = "empty_selection"
	IssueDuplicate        IssueCode = "duplicate"
	IssueCollision        IssueCode = "collision"
	IssueMissingReference IssueCode = "missing_reference"
	IssueRegistryVersion  IssueCode = "registry_version"
	IssueManifestVersion  IssueCode = "manifest_version"
	IssueManifestMissing  IssueCode = "manifest_missing"
	IssueInvalidRoot      IssueCode = "invalid_root"
	IssuePathEscape       IssueCode = "path_escape"
	IssueAlreadyExists    IssueCode = "already_exists"
	IssueUnavailable      IssueCode = "unavailable"
)

// Issue is the serializable stable diagnostic shape used in validation and
// JSON process envelopes.
type Issue struct {
	Code    IssueCode `json:"code"`
	Path    string    `json:"path,omitempty"`
	Message string    `json:"message"`
}

func (i Issue) Error() string {
	if i.Path == "" {
		return i.Message
	}
	return fmt.Sprintf("%s: %s", i.Path, i.Message)
}

// ValidationErrors retains every deterministic validation issue found before
// resolution. No operation should proceed when this error is returned.
type ValidationErrors struct {
	Issues []Issue `json:"issues"`
}

func (e *ValidationErrors) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, issue.Error())
	}
	return strings.Join(parts, "; ")
}

func (e *ValidationErrors) Unwrap() error { return nil }

func newValidationErrors(issues []Issue) error {
	if len(issues) == 0 {
		return nil
	}
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		if issues[i].Code != issues[j].Code {
			return issues[i].Code < issues[j].Code
		}
		return issues[i].Message < issues[j].Message
	})
	return &ValidationErrors{Issues: issues}
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Evidence struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// Envelope is the process-facing JSON shape. The fields error, warnings, and
// evidence are always emitted (including null/empty values) for stable
// automation consumption.
type Envelope struct {
	Operation CommandOperation `json:"operation"`
	Result    Result           `json:"result"`
	Error     *Issue           `json:"error"`
	Warnings  []Warning        `json:"warnings"`
	Evidence  []Evidence       `json:"evidence"`
	Plan      *Plan            `json:"plan,omitempty"`
	Profiles  []ProfileInfo    `json:"profiles,omitempty"`
	Path      string           `json:"path,omitempty"`
}

type ProfileInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (e Envelope) ExitStatus() ExitStatus {
	switch e.Result {
	case ResultSuccess:
		return ExitSuccess
	case ResultInvalid:
		return ExitInvalidInput
	case ResultUnavailable, ResultConflict:
		return ExitExecutionFailure
	default:
		return ExitExecutionFailure
	}
}

func defaultTargets() []Target { return []Target{TargetAgents, TargetClaude} }

func copyTargets(targets []Target) []Target {
	return append([]Target(nil), targets...)
}

func sortStrings(values []string) []string {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return copyValues
}
