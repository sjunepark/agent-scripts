package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

type cli struct {
	JSON    bool `name:"json" help:"Emit one JSON result document."`
	Version bool `name:"version" help:"Print the sjskills version."`

	Init     initCommand     `cmd:"" help:"Create a project manifest without overwriting one."`
	Profiles profilesCommand `cmd:"" help:"List selectable project profiles."`
	Plan     planCommand     `cmd:"" help:"Resolve desired state and verified expected content without changing managed roots."`
	Apply    applyCommand    `cmd:"" help:"Apply verified project skill installs, updates, and removals to quarantine."`
	Restore  restoreCommand  `cmd:"" help:"Restore a committed project quarantine without overwriting managed placements."`
}

type initCommand struct {
	Profiles []string `arg:"" name:"profile" optional:"" help:"One or more project profiles."`
}

type profilesCommand struct{}

type planCommand struct {
	Global bool `name:"global" help:"Resolve the fixed global baseline."`
}

type applyCommand struct {
	Global bool `name:"global" help:"Resolve the fixed global baseline without applying it."`
	Yes    bool `name:"yes" help:"Apply without prompting for confirmation."`
}

type restoreCommand struct {
	ID  string `arg:"" name:"quarantine-id" help:"Manifest-backed quarantine identifier."`
	Yes bool   `name:"yes" help:"Restore without prompting for confirmation."`
}

type commandContext struct {
	application *application
	context     context.Context
}

func (c *initCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.init(c.Profiles)
	return nil
}

func (c *profilesCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.profiles()
	return nil
}

func (c *planCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.plan(ctx.context, c.Global)
	return nil
}

func (c *applyCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.apply(ctx.context, c.Global, c.Yes)
	return nil
}

func (c *restoreCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.restore(ctx.context, c.ID, c.Yes)
	return nil
}

type application struct {
	directory           string
	envelope            sjskills.Envelope
	jsonMode            bool
	input               io.Reader
	promptOutput        io.Writer
	materialize         materializeFunc
	verifyMaterialized  materializationVerifyFunc
	cleanupMaterialized materializationCleanupFunc
	translateProject    projectTranslationFunc
	applyProject        projectApplyFunc
	restoreProject      projectRestoreFunc
}

// materializeFunc is the one process and temporary-state seam owned by the
// CLI. It deliberately mirrors Materializer.Materialize so the lifecycle can
// retain one verified snapshot session until planning has copied its hashes.
type materializeFunc func(context.Context, []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error)

type materializationVerifyFunc func(*sjskills.MaterializationPlan) error

type materializationCleanupFunc func(*sjskills.MaterializationPlan) error

// projectTranslationFunc keeps the pure project-to-plan boundary injectable
// for lifecycle tests without introducing a second plan model.
type projectTranslationFunc func(sjskills.Plan, sjskills.ProjectClassification) (sjskills.Plan, error)

type projectApplyFunc func(context.Context, *sjskills.ProjectApplySession, sjskills.ApplyDeps) (sjskills.ApplyResult, error)

type projectRestoreFunc func(context.Context, sjskills.DerivedLayout, string, sjskills.ApplyDeps) (sjskills.RestoreResult, error)

// productionMaterialize keeps construction lazy: profiles, init, help, and
// version never construct or invoke the Skills CLI adapter.
func productionMaterialize(ctx context.Context, skills []sjskills.DesiredSkill) (*sjskills.MaterializationPlan, error) {
	return sjskills.NewMaterializer(sjskills.MaterializerConfig{}).Materialize(ctx, skills)
}

func (a *application) registry() (sjskills.Registry, error) {
	return sjskills.EmbeddedRegistry()
}

func (a *application) base(operation sjskills.CommandOperation) sjskills.Envelope {
	return sjskills.Envelope{
		Operation: operation,
		Result:    sjskills.ResultSuccess,
		Warnings:  []sjskills.Warning{},
		Evidence:  []sjskills.Evidence{},
	}
}

func (a *application) invalid(operation sjskills.CommandOperation, err error) sjskills.Envelope {
	envelope := a.base(operation)
	envelope.Result = sjskills.ResultInvalid
	envelope.Error = issueFromError(err, sjskills.IssueMalformedInput)
	return envelope
}

func (a *application) unavailable(operation sjskills.CommandOperation, err error) sjskills.Envelope {
	envelope := a.base(operation)
	envelope.Result = sjskills.ResultUnavailable
	envelope.Error = issueFromError(err, sjskills.IssueUnavailable)
	return envelope
}

func (a *application) conflict(operation sjskills.CommandOperation, err error) sjskills.Envelope {
	envelope := a.base(operation)
	envelope.Result = sjskills.ResultConflict
	envelope.Error = issueFromError(err, sjskills.IssueAlreadyExists)
	return envelope
}

func (a *application) profiles() sjskills.Envelope {
	envelope := a.base(sjskills.CommandOperationProfiles)
	registry, err := a.registry()
	if err != nil {
		return a.invalid(sjskills.CommandOperationProfiles, err)
	}
	profiles := make([]string, 0, len(registry.Profiles))
	for name := range registry.Profiles {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	for _, name := range profiles {
		envelope.Profiles = append(envelope.Profiles, sjskills.ProfileInfo{Name: name, Count: len(registry.Profiles[name].Skills)})
	}
	envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "registry", Detail: "embedded version 4"})
	return envelope
}

func (a *application) init(profileNames []string) sjskills.Envelope {
	envelope := a.base(sjskills.CommandOperationInit)
	registry, err := a.registry()
	if err != nil {
		return a.invalid(sjskills.CommandOperationInit, err)
	}
	if len(profileNames) == 0 {
		return a.invalid(sjskills.CommandOperationInit, &sjskills.Issue{Code: sjskills.IssueEmptySelection, Path: "init.profile", Message: "at least one profile is required"})
	}
	profileNames = append([]string(nil), profileNames...)
	sort.Strings(profileNames)
	for index := 1; index < len(profileNames); index++ {
		if profileNames[index] == profileNames[index-1] {
			return a.invalid(sjskills.CommandOperationInit, &sjskills.Issue{Code: sjskills.IssueDuplicate, Path: "init.profile", Message: fmt.Sprintf("duplicate profile %q", profileNames[index])})
		}
	}
	manifest := sjskills.Manifest{Version: sjskills.ManifestVersion, Profiles: profileNames}
	if err := sjskills.ValidateManifest(registry, manifest); err != nil {
		return a.invalid(sjskills.CommandOperationInit, err)
	}
	manifestPath := filepath.Join(a.directory, sjskills.ManifestFileName)
	content := renderManifest(manifest)
	file, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return a.conflict(sjskills.CommandOperationInit, &sjskills.Issue{Code: sjskills.IssueAlreadyExists, Path: manifestPath, Message: "manifest already exists; refusing to overwrite"})
		}
		return a.unavailable(sjskills.CommandOperationInit, fmt.Errorf("create manifest: %w", err))
	}
	cleanup := func(primary error) sjskills.Envelope {
		_ = file.Close()
		_ = os.Remove(manifestPath)
		return a.unavailable(sjskills.CommandOperationInit, primary)
	}
	if written, writeErr := file.WriteString(content); writeErr != nil || written != len(content) {
		if writeErr == nil {
			writeErr = io.ErrShortWrite
		}
		return cleanup(fmt.Errorf("write manifest: %w", writeErr))
	}
	if syncErr := file.Sync(); syncErr != nil {
		return cleanup(fmt.Errorf("sync manifest: %w", syncErr))
	}
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(manifestPath)
		return a.unavailable(sjskills.CommandOperationInit, fmt.Errorf("close manifest: %w", closeErr))
	}
	envelope.Path = manifestPath
	envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "manifest", Detail: fmt.Sprintf("created %s", manifestPath)})
	return envelope
}

type preparedPlan struct {
	envelope     sjskills.Envelope
	plan         sjskills.Plan
	materialized *sjskills.MaterializationPlan
	snapshots    []*sjskills.SkillSnapshot
	expected     map[string]sjskills.TreeHash
	project      *sjskills.ProjectApplySession
	cleanup      materializationCleanupFunc
	cleaned      bool
	verified     bool
}

func (a *application) plan(ctx context.Context, global bool) sjskills.Envelope {
	prepared, envelope := a.prepare(ctx, global, sjskills.CommandOperationPlan)
	if prepared == nil {
		return envelope
	}
	return prepared.finish(envelope, "cleanup")
}

func (a *application) prepare(ctx context.Context, global bool, operation sjskills.CommandOperation) (*preparedPlan, sjskills.Envelope) {
	envelope := a.base(operation)
	registry, err := a.registry()
	if err != nil {
		return nil, a.invalid(operation, err)
	}
	request := sjskills.ResolveRequest{Registry: registry, Global: global}
	var project *sjskills.ProjectRoot
	if !global {
		discovered, discoverErr := sjskills.DiscoverProjectRoot(a.directory)
		if discoverErr != nil {
			return nil, a.invalid(operation, discoverErr)
		}
		manifest, readErr := sjskills.ReadManifest(discovered.ManifestPath)
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				return nil, a.invalid(operation, &sjskills.Issue{Code: sjskills.IssueManifestMissing, Path: discovered.ManifestPath, Message: "sjskills.toml was not found"})
			}
			return nil, a.invalid(operation, readErr)
		}
		request.Manifest = &manifest
		project = &discovered
		envelope.Path = discovered.ManifestPath
	}
	plan, err := sjskills.BuildPlan(request)
	if err != nil {
		return nil, a.invalid(operation, err)
	}
	prepared := &preparedPlan{envelope: envelope, plan: plan, expected: map[string]sjskills.TreeHash{}}
	prepared.syncPlan(plan)

	materialize := a.materialize
	if materialize == nil {
		materialize = productionMaterialize
	}
	if ctx == nil {
		ctx = context.Background()
	}
	materialized, materializeErr := materialize(ctx, plan.Desired.Skills)
	if materializeErr != nil {
		if materialized != nil {
			prepared.materialized = materialized
			prepared.cleanup = a.materializationCleanup()
			envelope = unavailableWithPlan(prepared.envelope, materializeErr)
			return nil, prepared.finish(envelope, "materialize")
		}
		return nil, unavailableWithPlan(prepared.envelope, materializeErr)
	}
	if materialized == nil {
		return nil, unavailableWithPlan(prepared.envelope, errors.New("materialization returned no session"))
	}
	prepared.materialized = materialized
	prepared.cleanup = a.materializationCleanup()

	if verifyErr := a.materializationVerify()(materialized); verifyErr != nil {
		envelope = unavailableWithPlan(prepared.envelope, verifyErr)
		return nil, prepared.finish(envelope, "verify")
	}
	snapshots := materialized.Snapshots()
	prepared.snapshots = snapshots
	for _, snapshot := range snapshots {
		if snapshot == nil {
			envelope = unavailableWithPlan(prepared.envelope, errors.New("materialization returned a nil snapshot"))
			return nil, prepared.finish(envelope, "verify")
		}
		prepared.expected[snapshot.Skill.Name] = snapshot.Hash
	}
	prepared.verified = true
	prepared.syncPlan(plan)

	if !global {
		layout, layoutErr := sjskills.LayoutForProject(project.Root)
		if layoutErr != nil {
			envelope = unavailableWithPlan(prepared.envelope, layoutErr)
			return nil, prepared.finish(envelope, "layout")
		}
		inventory, inspectErr := sjskills.InspectProject(layout)
		if inspectErr != nil {
			envelope = unavailableWithPlan(prepared.envelope, inspectErr)
			return nil, prepared.finish(envelope, "inspect")
		}
		classification, classifyErr := sjskills.ClassifyProject(plan.Desired, prepared.expected, inventory)
		if classifyErr != nil {
			envelope = unavailableWithPlan(prepared.envelope, classifyErr)
			return nil, prepared.finish(envelope, "classify")
		}
		translateProject := a.translateProject
		if translateProject == nil {
			translateProject = sjskills.TranslateProjectClassification
		}
		translated, translateErr := translateProject(plan, classification)
		if translateErr != nil {
			envelope = unavailableWithPlan(prepared.envelope, translateErr)
			return nil, prepared.finish(envelope, "translate")
		}
		plan = translated
		prepared.syncPlan(plan)
		prepared.project = &sjskills.ProjectApplySession{
			Layout:       layout,
			Desired:      plan.Desired,
			Plan:         plan,
			Expected:     copyExpected(prepared.expected),
			Materialized: materialized,
		}
	}
	return prepared, prepared.envelope
}

func (a *application) apply(ctx context.Context, global, yes bool) sjskills.Envelope {
	if a.jsonMode && !yes {
		return a.invalid(sjskills.CommandOperationApply, &sjskills.Issue{Code: sjskills.IssueMalformedInput, Path: "apply.yes", Message: "JSON apply requires --yes"})
	}
	prepared, envelope := a.prepare(ctx, global, sjskills.CommandOperationApply)
	if prepared == nil {
		return envelope
	}
	// Apply output is an operator result, not a path-discovery response. Keep
	// the manifest path out of normal and recovery output alongside the other
	// path-free execution evidence.
	envelope.Path = ""
	prepared.envelope.Path = ""
	if global {
		envelope.Result = sjskills.ResultUnavailable
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: "global reconciliation is not implemented in this slice"}
		envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "global managed roots unchanged"})
		return prepared.finish(envelope, "apply")
	}
	if reason := unsupportedApplyAction(prepared.plan); reason != "" {
		envelope.Result = sjskills.ResultConflict
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueReconciliationConflict, Path: "apply", Message: reason}
		envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "project managed roots unchanged"})
		return prepared.finish(envelope, "apply")
	}
	installCount := planActionCount(prepared.plan, sjskills.PlanActionInstall)
	updateCount := planActionCount(prepared.plan, sjskills.PlanActionUpdate)
	quarantineCount := planActionCount(prepared.plan, sjskills.PlanActionQuarantine)
	if installCount+updateCount+quarantineCount > 0 && !yes {
		confirmed, confirmErr := confirmProjectApply(a.input, a.promptOutput, installCount, updateCount, quarantineCount)
		if confirmErr != nil {
			envelope.Result = sjskills.ResultUnavailable
			envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: "confirmation could not be read"}
		} else if !confirmed {
			envelope.Result = sjskills.ResultUnavailable
			envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: "project apply was not confirmed"}
		}
		if envelope.Result != sjskills.ResultSuccess {
			envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "project managed roots unchanged"})
			return prepared.finish(envelope, "apply")
		}
	}
	applyProject := a.applyProject
	if applyProject == nil {
		applyProject = sjskills.ApplyProjectChanges
	}
	result, applyErr := applyProject(ctx, prepared.project, sjskills.ApplyDeps{})
	if result.Plan.Desired.Scope != "" {
		prepared.syncPlan(result.Plan)
	}
	envelope = prepared.envelope
	envelope.Evidence = append(envelope.Evidence, applyExecutionEvidence(result, applyErr)...)
	if evidence, ok := projectQuarantineEvidence(result.Quarantine); ok {
		envelope.Evidence = append(envelope.Evidence, evidence)
	}
	if applyErr != nil {
		setApplyFailure(&envelope, applyErr)
	}
	return prepared.finish(envelope, "apply")
}

func applyExecutionEvidence(result sjskills.ApplyResult, err error) []sjskills.Evidence {
	if err == nil {
		return []sjskills.Evidence{
			{Kind: "execution", Detail: fmt.Sprintf("installed %d project placements", len(result.Installed))},
			{Kind: "execution", Detail: fmt.Sprintf("updated %d project placements", len(result.Updated))},
			{Kind: "execution", Detail: fmt.Sprintf("quarantined %d removed project placements", len(result.Quarantined))},
		}
	}
	if len(result.Installed)+len(result.Updated)+len(result.Quarantined) > 0 {
		return []sjskills.Evidence{
			{Kind: "execution", Detail: fmt.Sprintf("reported %d committed installed project placements before apply failure", len(result.Installed))},
			{Kind: "execution", Detail: fmt.Sprintf("reported %d committed updated project placements before apply failure", len(result.Updated))},
			{Kind: "execution", Detail: fmt.Sprintf("reported %d committed quarantined removed project placements before apply failure", len(result.Quarantined))},
		}
	}
	return []sjskills.Evidence{{Kind: "execution", Detail: "no committed project placements were reported before apply failure"}}
}

func projectQuarantineEvidence(result *sjskills.ProjectQuarantineResult) (sjskills.Evidence, bool) {
	if result == nil || !validQuarantineID(result.ID) || !validQuarantineStatus(result.Status) {
		return sjskills.Evidence{}, false
	}
	return sjskills.Evidence{Kind: "quarantine", Detail: fmt.Sprintf("id=%s status=%s", result.ID, result.Status)}, true
}

func validQuarantineID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for _, char := range id {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func validQuarantineStatus(status sjskills.ProjectQuarantineStatus) bool {
	switch status {
	case sjskills.ProjectQuarantinePrepared, sjskills.ProjectQuarantineActive,
		sjskills.ProjectQuarantineCommitted, sjskills.ProjectQuarantineRestoring,
		sjskills.ProjectQuarantineRestored, sjskills.ProjectQuarantineRolledBack,
		sjskills.ProjectQuarantineRecoveryRequired:
		return true
	default:
		return false
	}
}

func (a *application) materializationVerify() materializationVerifyFunc {
	if a.verifyMaterialized != nil {
		return a.verifyMaterialized
	}
	return func(plan *sjskills.MaterializationPlan) error { return plan.Verify() }
}

func (a *application) materializationCleanup() materializationCleanupFunc {
	if a.cleanupMaterialized != nil {
		return a.cleanupMaterialized
	}
	return func(plan *sjskills.MaterializationPlan) error { return plan.Cleanup() }
}

func (p *preparedPlan) syncPlan(plan sjskills.Plan) {
	p.plan = plan
	p.envelope.Plan = &p.plan
	p.envelope.Warnings = append([]sjskills.Warning{}, plan.Warnings...)
	p.envelope.Evidence = append([]sjskills.Evidence{{Kind: "registry", Detail: "embedded version 4"}}, plan.Evidence...)
	for _, snapshot := range p.snapshots {
		p.envelope.Evidence = append(p.envelope.Evidence, sjskills.Evidence{
			Kind:   "expected-content",
			Detail: fmt.Sprintf("%s %s:%s", snapshot.Skill.Name, sjskills.TreeHashAlgorithmSHA256V2, snapshot.Hash.Digest),
		})
	}
}

func (p *preparedPlan) finish(envelope sjskills.Envelope, stage string) sjskills.Envelope {
	if !p.cleaned && p.materialized != nil {
		p.cleaned = true
		if cleanupErr := p.cleanup(p.materialized); cleanupErr != nil {
			envelope.Result = sjskills.ResultUnavailable
			if envelope.Error != nil {
				envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Message: fmt.Sprintf("materialization cleanup failed after %s failure", stage)}
				return envelope
			}
			envelope.Error = issueFromError(lifecycleError("cleanup", nil, cleanupErr), sjskills.IssueUnavailable)
			return envelope
		}
	}
	if p.verified {
		envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{
			Kind:   "materialization",
			Detail: fmt.Sprintf("skills@%s verified %d snapshots; temporary cleanup successful", sjskills.SkillsCLIVersion, len(p.snapshots)),
		})
	}
	return envelope
}

func copyExpected(expected map[string]sjskills.TreeHash) map[string]sjskills.TreeHash {
	result := make(map[string]sjskills.TreeHash, len(expected))
	for name, hash := range expected {
		result[name] = hash
	}
	return result
}

func unsupportedApplyAction(plan sjskills.Plan) string {
	for _, operation := range plan.Operations {
		switch operation.Action {
		case sjskills.PlanActionInstall, sjskills.PlanActionUpdate, sjskills.PlanActionQuarantine, sjskills.PlanActionUnchanged, sjskills.PlanActionManual, sjskills.PlanActionWorkflow:
			continue
		case sjskills.PlanActionBlocked:
			return "reviewed plan contains a blocked project placement"
		default:
			return "reviewed plan contains an unsupported project action"
		}
	}
	return ""
}

func planActionCount(plan sjskills.Plan, action sjskills.PlanAction) int {
	count := 0
	for _, operation := range plan.Operations {
		if operation.Action == action {
			count++
		}
	}
	return count
}

func confirmProjectApply(input io.Reader, output io.Writer, installs, updates, removals int) (bool, error) {
	if input == nil {
		input = strings.NewReader("")
	}
	if output == nil {
		output = io.Discard
	}
	var prompt string
	count := installs + updates + removals
	switch {
	case count == 0:
		return false, nil
	case updates == 0 && removals == 0:
		prompt = fmt.Sprintf("Apply %d project skill installs? [y/N] ", installs)
	case installs == 0 && removals == 0:
		prompt = fmt.Sprintf("Apply %d project skill updates? [y/N] ", updates)
	case installs == 0 && updates == 0:
		prompt = fmt.Sprintf("Apply %d project skill removals to quarantine? [y/N] ", removals)
	default:
		categories := make([]string, 0, 3)
		if installs > 0 {
			categories = append(categories, mutationCount(installs, "install"))
		}
		if updates > 0 {
			categories = append(categories, mutationCount(updates, "update"))
		}
		if removals > 0 {
			categories = append(categories, mutationCount(removals, "removal")+" to quarantine")
		}
		prompt = fmt.Sprintf("Apply %d project skill changes (%s)? [y/N] ", count, strings.Join(categories, ", "))
	}
	if _, err := io.WriteString(output, prompt); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64), 64)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

func mutationCount(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}

func setApplyFailure(envelope *sjskills.Envelope, err error) {
	var applyErr *sjskills.ApplyError
	if errors.As(err, &applyErr) && applyErr.Conflict() {
		envelope.Result = sjskills.ResultConflict
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueReconciliationConflict, Path: "apply", Message: applyErr.Error()}
		return
	}
	envelope.Result = sjskills.ResultUnavailable
	if errors.As(err, &applyErr) {
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: applyErr.Error()}
		return
	}
	envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: "project apply unavailable"}
}

func unavailableWithPlan(envelope sjskills.Envelope, err error) sjskills.Envelope {
	envelope.Result = sjskills.ResultUnavailable
	envelope.Error = issueFromError(err, sjskills.IssueUnavailable)
	return envelope
}

func lifecycleError(stage string, primary, cleanup error) error {
	if cleanup == nil {
		if primary != nil {
			// Materializer diagnostics are already bounded and redacted. Keep the
			// original error so the CLI does not re-expand sensitive details.
			return primary
		}
		return fmt.Errorf("materialization %s failed", stage)
	}
	if primary == nil {
		return fmt.Errorf("materialization %s failed", stage)
	}
	return fmt.Errorf("materialization %s failed and cleanup failed", stage)
}

func (a *application) restore(ctx context.Context, id string, yes bool) sjskills.Envelope {
	const invalidIDMessage = "quarantine-id must be exactly 32 lowercase hexadecimal characters"
	if !validQuarantineID(id) {
		return a.invalid(sjskills.CommandOperationRestore, &sjskills.Issue{
			Code:    sjskills.IssueMalformedInput,
			Path:    "restore.quarantine-id",
			Message: invalidIDMessage,
		})
	}
	if a.jsonMode && !yes {
		return a.invalid(sjskills.CommandOperationRestore, &sjskills.Issue{
			Code:    sjskills.IssueMalformedInput,
			Path:    "restore.yes",
			Message: "JSON restore requires --yes",
		})
	}

	// Restore is project-scoped but does not need the manifest contents or a
	// remote materialization. Discovery proves the canonical root; the derived
	// layout then supplies the internal transaction's bounded paths.
	discovered, discoverErr := sjskills.DiscoverProjectRoot(a.directory)
	if discoverErr != nil {
		return restoreErrorEnvelope(sjskills.ResultInvalid, sjskills.IssueInvalidRoot, "canonical project root is required for restore")
	}
	layout, layoutErr := sjskills.LayoutForProject(discovered.Root)
	if layoutErr != nil {
		return restoreErrorEnvelope(sjskills.ResultInvalid, sjskills.IssueInvalidRoot, "canonical project root is required for restore")
	}

	if !yes {
		confirmed, confirmErr := confirmProjectRestore(a.input, a.promptOutput, id)
		if confirmErr != nil {
			envelope := restoreErrorEnvelope(sjskills.ResultUnavailable, sjskills.IssueUnavailable, "confirmation could not be read")
			envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "project managed roots unchanged"})
			return envelope
		}
		if !confirmed {
			envelope := restoreErrorEnvelope(sjskills.ResultUnavailable, sjskills.IssueUnavailable, "project restore was not confirmed")
			envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "project managed roots unchanged"})
			return envelope
		}
	}

	projectRestore := a.restoreProject
	if projectRestore == nil {
		projectRestore = sjskills.RestoreProjectQuarantine
	}
	result, restoreErr := projectRestore(ctx, layout, id, sjskills.ApplyDeps{})
	envelope := a.base(sjskills.CommandOperationRestore)
	// Restore output is an operator result. The canonical root and all derived
	// paths remain private implementation details, including on failures.
	envelope.Path = ""
	envelope.Evidence = append(envelope.Evidence, restoreExecutionEvidence(result, restoreErr)...)
	if result.ID != "" {
		if evidence, ok := projectQuarantineEvidence(&sjskills.ProjectQuarantineResult{ID: result.ID, Status: result.Status}); ok {
			envelope.Evidence = append(envelope.Evidence, evidence)
		}
	}
	if restoreErr != nil {
		setRestoreFailure(&envelope, restoreErr)
	}
	return envelope
}

func restoreErrorEnvelope(result sjskills.Result, code sjskills.IssueCode, message string) sjskills.Envelope {
	return sjskills.Envelope{
		Operation: sjskills.CommandOperationRestore,
		Result:    result,
		Error:     &sjskills.Issue{Code: code, Path: "restore", Message: message},
		Warnings:  []sjskills.Warning{},
		Evidence:  []sjskills.Evidence{},
	}
}

func confirmProjectRestore(input io.Reader, output io.Writer, id string) (bool, error) {
	if !validQuarantineID(id) {
		return false, errors.New("invalid quarantine identifier")
	}
	if input == nil {
		input = strings.NewReader("")
	}
	if output == nil {
		output = io.Discard
	}
	if _, err := fmt.Fprintf(output, "Quarantined project placements for %s will be restored without overwrite? [y/N] ", id); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64), 64)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

func restoreExecutionEvidence(result sjskills.RestoreResult, err error) []sjskills.Evidence {
	if err == nil {
		return []sjskills.Evidence{{Kind: "execution", Detail: fmt.Sprintf("restored %d project placements", len(result.Restored))}}
	}
	if len(result.Restored) > 0 {
		return []sjskills.Evidence{{Kind: "execution", Detail: fmt.Sprintf("reported %d committed restored project placements before restore failure", len(result.Restored))}}
	}
	return []sjskills.Evidence{{Kind: "execution", Detail: "no committed project placements were reported before restore failure"}}
}

func setRestoreFailure(envelope *sjskills.Envelope, err error) {
	var restoreErr *sjskills.RestoreError
	if errors.As(err, &restoreErr) {
		if restoreErr.Conflict() {
			envelope.Result = sjskills.ResultConflict
			envelope.Error = &sjskills.Issue{Code: sjskills.IssueReconciliationConflict, Path: "restore", Message: restoreErr.Error()}
			return
		}
		envelope.Result = sjskills.ResultUnavailable
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "restore", Message: restoreErr.Error()}
		return
	}
	envelope.Result = sjskills.ResultUnavailable
	envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "restore", Message: "project restore unavailable"}
}

func renderManifest(manifest sjskills.Manifest) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "version = %d\n", manifest.Version)
	fmt.Fprintf(&builder, "profiles = [%s]\n", quoteStrings(manifest.Profiles))
	for _, direct := range manifest.Direct {
		builder.WriteString("\n[[direct]]\n")
		fmt.Fprintf(&builder, "name = %q\nsource = %q\n", direct.Name, direct.Source)
		if direct.FullDepth {
			builder.WriteString("full_depth = true\n")
		}
	}
	return builder.String()
}

func quoteStrings[T ~string](values []T) string {
	quoted := make([]string, len(values))
	for index, value := range values {
		quoted[index] = fmt.Sprintf("%q", string(value))
	}
	return strings.Join(quoted, ", ")
}

func issueFromError(err error, fallback sjskills.IssueCode) *sjskills.Issue {
	if err == nil {
		return &sjskills.Issue{Code: fallback, Message: "unknown error"}
	}
	var issue *sjskills.Issue
	if errors.As(err, &issue) {
		return issue
	}
	var validation *sjskills.ValidationErrors
	if errors.As(err, &validation) && len(validation.Issues) > 0 {
		first := validation.Issues[0]
		return &first
	}
	return &sjskills.Issue{Code: fallback, Message: err.Error()}
}

func isJSONRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func emitEnvelope(stdout, stderr io.Writer, jsonMode bool, envelope sjskills.Envelope) int {
	if jsonMode {
		// Encoder.Encode is deliberately used so every JSON response is exactly
		// one newline-terminated document.
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintf(stderr, "sjskills: write JSON result: %v\n", err)
			return int(sjskills.ExitExecutionFailure)
		}
		return int(envelope.ExitStatus())
	}
	renderHuman(stdout, stderr, envelope)
	return int(envelope.ExitStatus())
}

func renderHuman(stdout, stderr io.Writer, envelope sjskills.Envelope) {
	if envelope.Error != nil {
		fmt.Fprintf(stderr, "sjskills: %s\n", envelope.Error.Error())
	}
	switch envelope.Operation {
	case sjskills.CommandOperationProfiles:
		for _, profile := range envelope.Profiles {
			fmt.Fprintf(stdout, "%s (%d skills)\n", profile.Name, profile.Count)
		}
	case sjskills.CommandOperationPlan, sjskills.CommandOperationApply:
		if envelope.Plan != nil {
			fmt.Fprintf(stdout, "%s: %s (%d skills)\n", envelope.Operation, envelope.Result, len(envelope.Plan.Desired.Skills))
		} else {
			fmt.Fprintf(stdout, "%s: %s\n", envelope.Operation, envelope.Result)
		}
	case sjskills.CommandOperationInit:
		fmt.Fprintf(stdout, "init: %s", envelope.Result)
		if envelope.Path != "" {
			fmt.Fprintf(stdout, " (%s)", envelope.Path)
		}
		fmt.Fprintln(stdout)
	case sjskills.CommandOperationRestore:
		fmt.Fprintf(stdout, "restore: %s\n", envelope.Result)
	}
	for _, warning := range envelope.Warnings {
		fmt.Fprintf(stderr, "sjskills warning: %s\n", warning.Message)
	}
	for _, evidence := range envelope.Evidence {
		if envelope.Operation == sjskills.CommandOperationRestore && evidence.Kind == "execution" {
			fmt.Fprintf(stdout, "execution: %s\n", evidence.Detail)
		}
		if evidence.Kind == "quarantine" {
			fmt.Fprintf(stdout, "quarantine: %s\n", evidence.Detail)
		}
	}
}

func execute(ctx context.Context, args []string, stdout, stderr io.Writer, directory string) int {
	return executeWithInput(ctx, args, strings.NewReader(""), stdout, stderr, directory)
}

func executeWithInput(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, directory string) int {
	if ctx == nil {
		ctx = context.Background()
	}
	jsonMode := isJSONRequested(args)
	if isExactVersionRequest(args) {
		if jsonMode {
			envelope := sjskills.Envelope{Operation: sjskills.CommandOperationVersion, Result: sjskills.ResultSuccess, Warnings: []sjskills.Warning{}, Evidence: []sjskills.Evidence{{Kind: "version", Detail: sjskills.ToolVersion}}}
			return emitEnvelope(stdout, stderr, true, envelope)
		}
		fmt.Fprintf(stdout, "sjskills %s\n", sjskills.ToolVersion)
		return int(sjskills.ExitSuccess)
	}
	var commands cli
	kongParser, err := kong.New(&commands, kong.Name("sjskills"), kong.Description("Resolve reproducible project and global skills."))
	if err != nil {
		fmt.Fprintf(stderr, "sjskills: configure parser: %v\n", err)
		return int(sjskills.ExitExecutionFailure)
	}
	parsed, err := kongParser.Parse(args)
	if err != nil {
		if jsonMode {
			envelope := sjskills.Envelope{Operation: sjskills.CommandOperationParse, Result: sjskills.ResultInvalid, Error: &sjskills.Issue{Code: sjskills.IssueMalformedInput, Path: "arguments", Message: err.Error()}, Warnings: []sjskills.Warning{}, Evidence: []sjskills.Evidence{}}
			_ = json.NewEncoder(stdout).Encode(envelope)
		} else {
			fmt.Fprintf(stderr, "sjskills: %v\n", err)
		}
		return int(sjskills.ExitInvalidInvocation)
	}
	app := &application{
		directory:    directory,
		jsonMode:     commands.JSON,
		input:        stdin,
		promptOutput: stderr,
		materialize:  productionMaterialize,
	}
	if err := parsed.Run(&commandContext{application: app, context: ctx}); err != nil {
		if jsonMode {
			envelope := sjskills.Envelope{Operation: sjskills.CommandOperationParse, Result: sjskills.ResultInvalid, Error: &sjskills.Issue{Code: sjskills.IssueMalformedInput, Path: "command", Message: err.Error()}, Warnings: []sjskills.Warning{}, Evidence: []sjskills.Evidence{}}
			_ = json.NewEncoder(stdout).Encode(envelope)
		} else {
			fmt.Fprintf(stderr, "sjskills: %v\n", err)
		}
		return int(sjskills.ExitExecutionFailure)
	}
	return emitEnvelope(stdout, stderr, commands.JSON, app.envelope)
}

func isExactVersionRequest(args []string) bool {
	if len(args) == 1 && args[0] == "--version" {
		return true
	}
	return len(args) == 2 && ((args[0] == "--json" && args[1] == "--version") || (args[0] == "--version" && args[1] == "--json"))
}

func main() {
	directory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sjskills: get working directory: %v\n", err)
		os.Exit(int(sjskills.ExitExecutionFailure))
	}
	os.Exit(executeWithInput(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, directory))
}
