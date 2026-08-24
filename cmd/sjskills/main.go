package main

import (
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
	Plan     planCommand     `cmd:"" help:"Resolve desired state without materializing it."`
	Apply    applyCommand    `cmd:"" help:"Resolve state and report execution availability."`
	Restore  restoreCommand  `cmd:"" help:"Restore a recorded quarantine entry."`
}

type initCommand struct {
	Profiles []string `arg:"" name:"profile" optional:"" help:"One or more project profiles."`
}

type profilesCommand struct{}

type planCommand struct {
	Global bool `name:"global" help:"Resolve the fixed global baseline."`
}

type applyCommand struct {
	Global bool `name:"global" help:"Resolve the fixed global baseline."`
	Yes    bool `name:"yes" help:"Do not prompt (execution is unavailable in v1 slice)."`
}

type restoreCommand struct {
	ID string `arg:"" name:"quarantine-id" optional:"" help:"Manifest-backed quarantine identifier."`
}

type commandContext struct {
	application *application
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
	ctx.application.envelope = ctx.application.plan(c.Global)
	return nil
}

func (c *applyCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.apply(c.Global)
	return nil
}

func (c *restoreCommand) Run(ctx *commandContext) error {
	ctx.application.envelope = ctx.application.restore(c.ID)
	return nil
}

type application struct {
	directory string
	envelope  sjskills.Envelope
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

func (a *application) plan(global bool) sjskills.Envelope {
	envelope := a.base(sjskills.CommandOperationPlan)
	registry, err := a.registry()
	if err != nil {
		return a.invalid(sjskills.CommandOperationPlan, err)
	}
	request := sjskills.ResolveRequest{Registry: registry, Global: global}
	if !global {
		project, discoverErr := sjskills.DiscoverProjectRoot(a.directory)
		if discoverErr != nil {
			return a.invalid(sjskills.CommandOperationPlan, discoverErr)
		}
		manifest, readErr := sjskills.ReadManifest(project.ManifestPath)
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				return a.invalid(sjskills.CommandOperationPlan, &sjskills.Issue{Code: sjskills.IssueManifestMissing, Path: project.ManifestPath, Message: "sjskills.toml was not found"})
			}
			return a.invalid(sjskills.CommandOperationPlan, readErr)
		}
		request.Manifest = &manifest
		envelope.Path = project.ManifestPath
	}
	plan, err := sjskills.BuildPlan(request)
	if err != nil {
		return a.invalid(sjskills.CommandOperationPlan, err)
	}
	envelope.Plan = &plan
	envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "registry", Detail: "embedded version 4"})
	envelope.Evidence = append(envelope.Evidence, plan.Evidence...)
	envelope.Warnings = append(envelope.Warnings, plan.Warnings...)
	return envelope
}

func (a *application) apply(global bool) sjskills.Envelope {
	envelope := a.plan(global)
	envelope.Operation = sjskills.CommandOperationApply
	if envelope.Result != sjskills.ResultSuccess {
		return envelope
	}
	envelope.Result = sjskills.ResultUnavailable
	envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "apply", Message: "materialization and managed-root mutation are not implemented in this slice"}
	envelope.Evidence = append(envelope.Evidence, sjskills.Evidence{Kind: "execution", Detail: "pure resolution only; no filesystem or subprocess side effects"})
	return envelope
}

func (a *application) restore(id string) sjskills.Envelope {
	if strings.TrimSpace(id) == "" {
		return a.invalid(sjskills.CommandOperationRestore, &sjskills.Issue{Code: sjskills.IssueMalformedInput, Path: "restore.quarantine-id", Message: "quarantine-id is required"})
	}
	return a.unavailable(sjskills.CommandOperationRestore, &sjskills.Issue{Code: sjskills.IssueUnavailable, Path: "restore", Message: "quarantine restore is not implemented in this slice"})
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
}

func execute(args []string, stdout, stderr io.Writer, directory string) int {
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
	app := &application{directory: directory}
	if err := parsed.Run(&commandContext{application: app}); err != nil {
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
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr, directory))
}
