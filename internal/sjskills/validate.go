package sjskills

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var portableNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var shorthandPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$`)

var supportedTargets = map[Target]struct{}{
	TargetAgents: {},
	TargetClaude: {},
}

var supportedManagers = map[Manager]struct{}{
	ManagerSkillsCLI: {},
	ManagerManual:    {},
	ManagerWorkflow:  {},
	ManagerNone:      {},
}

var supportedModes = map[InstallMode]struct{}{
	ModeCopy:    {},
	ModeSymlink: {},
}

var requiredProfiles = []string{"dev", "go", "kicpa", "rust"}

func isPortableName(value string) bool { return portableNamePattern.MatchString(value) }

func isSortedStrings(values []string) bool {
	return sort.SliceIsSorted(values, func(i, j int) bool { return values[i] < values[j] })
}

func isSortedTargets(values []Target) bool {
	return sort.SliceIsSorted(values, func(i, j int) bool { return values[i] < values[j] })
}

func addIssue(issues *[]Issue, code IssueCode, path, format string, args ...any) {
	*issues = append(*issues, Issue{Code: code, Path: path, Message: fmt.Sprintf(format, args...)})
}

// ValidateRegistry applies structural and semantic rules to a decoded v4
// registry. It never reads from disk, invokes a subprocess, or resolves a
// source.
func ValidateRegistry(registry Registry) error {
	var issues []Issue
	if registry.Version != RegistryVersion {
		addIssue(&issues, IssueRegistryVersion, "registry.version", "must be %d", RegistryVersion)
	}
	if strings.TrimSpace(registry.Description) == "" {
		addIssue(&issues, IssueMalformedInput, "registry.description", "must be a non-empty string")
	}
	validateDefaultTargets(registry.Defaults, &issues)
	validateSources(registry.Sources, &issues)
	validateSkillDeclarations(registry.Skills, registry.Sources, &issues)
	validateTargetExceptions(registry.TargetExceptions, registry.Skills, &issues)

	known := make(map[string]SkillDeclaration, len(registry.Skills))
	for _, skill := range registry.Skills {
		known[skill.Name] = skill
	}
	validateBaselineAndProfiles(registry.Global, registry.Profiles, known, &issues)

	usedSources := make(map[string]bool, len(registry.Sources))
	for _, skill := range registry.Skills {
		if _, ok := registry.Sources[skill.Source]; ok {
			usedSources[skill.Source] = true
		}
	}
	for _, sourceID := range sortedSourceIDs(registry.Sources) {
		if !usedSources[sourceID] {
			addIssue(&issues, IssueMissingReference, "registry.sources."+sourceID, "source is unused")
		}
	}
	return newValidationErrors(issues)
}

func validateDefaultTargets(defaults RegistryDefaults, issues *[]Issue) {
	if len(defaults.Targets) == 0 {
		addIssue(issues, IssueEmptySelection, "registry.defaults.targets", "must be a non-empty array")
		return
	}
	validateTargets("registry.defaults.targets", defaults.Targets, issues)
	expected := defaultTargets()
	if len(defaults.Targets) != len(expected) || (len(defaults.Targets) >= 2 && (defaults.Targets[0] != expected[0] || defaults.Targets[1] != expected[1])) {
		addIssue(issues, IssueInvalidTarget, "registry.defaults.targets", "must be exactly .agents and .claude")
	}
}

func validateSources(sources map[string]Source, issues *[]Issue) {
	if len(sources) == 0 {
		addIssue(issues, IssueEmptySelection, "registry.sources", "must define at least one source")
		return
	}
	for _, sourceID := range sortedSourceIDs(sources) {
		source := sources[sourceID]
		path := "registry.sources." + sourceID
		if !isPortableName(sourceID) {
			addIssue(issues, IssueInvalidName, path, "source id must contain lowercase letters, numbers, and single hyphens only")
		}
		switch source.Kind {
		case SourceRepository, SourceExternal:
		default:
			addIssue(issues, IssueInvalidSource, path+".kind", "unsupported source kind %q", source.Kind)
		}
		if source.Location == "" && source.Kind == SourceRepository {
			addIssue(issues, IssueInvalidSource, path+".location", "repository source must define a location")
		}
		if source.Location != "" && source.Kind == SourceRepository && !isRemoteSource(source.Location) {
			addIssue(issues, IssueInvalidSource, path+".location", "repository location must be a git shorthand or credential-free https source")
		}
		if source.Kind == SourceRepository && source.CatalogPath != "skills" {
			addIssue(issues, IssueInvalidSource, path+".catalogPath", "repository source must use catalogPath skills")
		}
		if source.Kind == SourceExternal && source.CatalogPath != "" {
			addIssue(issues, IssueInvalidSource, path+".catalogPath", "external source must not define catalogPath")
		}
	}
}

func validateSkillDeclarations(skills []SkillDeclaration, sources map[string]Source, issues *[]Issue) {
	if len(skills) == 0 {
		addIssue(issues, IssueEmptySelection, "registry.skills", "must define at least one skill")
		return
	}
	names := make(map[string]bool, len(skills))
	for index, skill := range skills {
		path := fmt.Sprintf("registry.skills[%d]", index)
		if skill.Name == "" || !isPortableName(skill.Name) {
			addIssue(issues, IssueInvalidName, path+".name", "must be a portable skill name")
		}
		if names[skill.Name] {
			addIssue(issues, IssueDuplicate, path+".name", "duplicate skill %q", skill.Name)
		}
		names[skill.Name] = true
		if index > 0 && skill.Name < skills[index-1].Name {
			addIssue(issues, IssueInvalidName, "registry.skills", "must be sorted by name")
		}
		source, ok := sources[skill.Source]
		if !ok {
			addIssue(issues, IssueMissingReference, path+".source", "references unknown source %q", skill.Source)
		}
		validateInstallation(path, skill.Manager, skill.Mode, skill.Workflow, skill.FullDepth, source, ok, issues)
	}
}

func validateInstallation(path string, manager Manager, mode InstallMode, workflow string, fullDepth bool, source Source, sourceKnown bool, issues *[]Issue) {
	if _, ok := supportedManagers[manager]; !ok {
		addIssue(issues, IssueInvalidManager, path+".manager", "unsupported installation manager %q", manager)
	}
	if mode != "" {
		if _, ok := supportedModes[mode]; !ok {
			addIssue(issues, IssueInvalidMode, path+".mode", "unsupported installation mode %q", mode)
		}
	}
	switch manager {
	case ManagerSkillsCLI:
		if !sourceKnown || source.Location == "" {
			addIssue(issues, IssueInvalidSource, path+".source", "skills-cli requires a source with a location")
		} else if problem := SkillsCLIPathProblem(source.Location); problem != "" {
			addIssue(issues, IssueInvalidSource, path+".source", "skills-cli requires a git shorthand or credential-free https source: %s", problem)
		}
		if mode == "" {
			addIssue(issues, IssueInvalidMode, path+".mode", "skills-cli installation must define mode")
		}
		if workflow != "" {
			addIssue(issues, IssueInvalidSource, path+".workflow", "skills-cli installation must not define workflow")
		}
	case ManagerManual, ManagerNone:
		if mode != "" {
			addIssue(issues, IssueInvalidMode, path+".mode", "%s installation must not define mode", manager)
		}
		if fullDepth {
			addIssue(issues, IssueInvalidSource, path+".fullDepth", "%s installation must not define fullDepth", manager)
		}
		if workflow != "" {
			addIssue(issues, IssueInvalidSource, path+".workflow", "%s installation must not define workflow", manager)
		}
	case ManagerWorkflow:
		if workflow == "" {
			addIssue(issues, IssueInvalidSource, path+".workflow", "workflow installation must name its workflow")
		}
		if mode != "" {
			addIssue(issues, IssueInvalidMode, path+".mode", "workflow installation must not define mode")
		}
		if fullDepth {
			addIssue(issues, IssueInvalidSource, path+".fullDepth", "workflow installation must not define fullDepth")
		}
	}
}

func validateTargetExceptions(exceptions map[string][]Target, skills []SkillDeclaration, issues *[]Issue) {
	known := make(map[string]bool, len(skills))
	for _, skill := range skills {
		known[skill.Name] = true
	}
	for _, name := range sortedStringKeys(exceptions) {
		targets := exceptions[name]
		path := "registry.targetExceptions." + name
		if !known[name] {
			addIssue(issues, IssueMissingReference, path, "references unknown skill")
		}
		if len(targets) == 0 {
			addIssue(issues, IssueEmptySelection, path, "must contain at least one target")
			continue
		}
		validateTargets(path, targets, issues)
	}
}

func validateTargets(path string, targets []Target, issues *[]Issue) {
	seen := make(map[Target]bool, len(targets))
	for _, target := range targets {
		if _, ok := supportedTargets[target]; !ok {
			addIssue(issues, IssueInvalidTarget, path, "unsupported target %q", target)
		}
		if seen[target] {
			addIssue(issues, IssueDuplicate, path, "duplicate target %q", target)
		}
		seen[target] = true
	}
	if !isSortedTargets(targets) {
		addIssue(issues, IssueInvalidTarget, path, "targets must be sorted")
	}
}

func validateBaselineAndProfiles(global GlobalRegistry, profiles map[string]Profile, known map[string]SkillDeclaration, issues *[]Issue) {
	if len(global.Baseline) == 0 {
		addIssue(issues, IssueEmptySelection, "registry.global.baseline", "must define at least one skill")
	}
	validateNamesList("registry.global.baseline", global.Baseline, issues)
	if len(profiles) != len(requiredProfiles) {
		addIssue(issues, IssueMissingReference, "registry.profiles", "must define exactly: %s", strings.Join(requiredProfiles, ", "))
	}
	for _, required := range requiredProfiles {
		profile, ok := profiles[required]
		if !ok {
			addIssue(issues, IssueMissingReference, "registry.profiles."+required, "required profile is missing")
			continue
		}
		if len(profile.Skills) == 0 {
			addIssue(issues, IssueEmptySelection, "registry.profiles."+required+".skills", "must define at least one skill")
		}
		validateNamesList("registry.profiles."+required+".skills", profile.Skills, issues)
	}
	for name := range profiles {
		if !contains(requiredProfiles, name) {
			addIssue(issues, IssueMissingReference, "registry.profiles."+name, "unsupported profile")
		}
	}

	membership := make(map[string][]string)
	for _, name := range global.Baseline {
		membership[name] = append(membership[name], "global baseline")
		if skill, ok := known[name]; !ok {
			addIssue(issues, IssueMissingReference, "registry.global.baseline."+name, "references unknown skill")
		} else if skill.Manager == ManagerNone {
			addIssue(issues, IssueInvalidManager, "registry.global.baseline."+name, "manager none cannot be selected")
		}
	}
	for _, profileName := range requiredProfiles {
		for _, name := range profiles[profileName].Skills {
			membership[name] = append(membership[name], "profile "+profileName)
			if _, ok := known[name]; !ok {
				addIssue(issues, IssueMissingReference, "registry.profiles."+profileName+".skills."+name, "references unknown skill")
			} else if known[name].Manager == ManagerNone {
				addIssue(issues, IssueInvalidManager, "registry.profiles."+profileName+".skills."+name, "manager none cannot be selected")
			}
		}
	}
	for name, origins := range membership {
		if len(origins) > 1 {
			addIssue(issues, IssueCollision, "registry.skills."+name, "skill appears in multiple desired sets: %s", strings.Join(sortStrings(origins), ", "))
		}
	}
}

// ValidateManifestShape validates fields that do not require a registry. It
// is useful for strict parser tests and is always followed by
// ValidateManifest against the canonical registry before resolution.
func ValidateManifestShape(manifest Manifest) error {
	var issues []Issue
	if manifest.Version != ManifestVersion {
		addIssue(&issues, IssueManifestVersion, "manifest.version", "must be %d", ManifestVersion)
	}
	if len(manifest.Profiles) == 0 && len(manifest.Direct) == 0 {
		addIssue(&issues, IssueEmptySelection, "manifest", "must select at least one profile or direct skill")
	}
	validateNamesList("manifest.profiles", manifest.Profiles, &issues)
	names := make(map[string]bool, len(manifest.Direct))
	for index, direct := range manifest.Direct {
		path := fmt.Sprintf("manifest.direct[%d]", index)
		if direct.Name == "" || !isPortableName(direct.Name) {
			addIssue(&issues, IssueInvalidName, path+".name", "must be a portable skill name")
		}
		if names[direct.Name] {
			addIssue(&issues, IssueDuplicate, path+".name", "duplicate direct skill %q", direct.Name)
		}
		names[direct.Name] = true
		if index > 0 && direct.Name < manifest.Direct[index-1].Name {
			addIssue(&issues, IssueInvalidName, "manifest.direct", "must be sorted by name")
		}
		if direct.Source == "" {
			addIssue(&issues, IssueInvalidSource, path+".source", "direct skill must record an installable source identity")
		} else if problem := SkillsCLIPathProblem(direct.Source); problem != "" {
			addIssue(&issues, IssueInvalidSource, path+".source", "direct skill source must be a git shorthand or credential-free https source: %s", problem)
		}
		// Direct entries intentionally have no manager, mode, target, or
		// workflow fields. Strict TOML decoding rejects those fields before
		// this semantic validation runs; v1 always resolves them to the
		// skills-cli/copy/default-target contract.
	}
	return newValidationErrors(issues)
}

// ValidateManifest applies registry references and cross-set collision rules.
func ValidateManifest(registry Registry, manifest Manifest) error {
	var issues []Issue
	if err := ValidateRegistry(registry); err != nil {
		if validation, ok := err.(*ValidationErrors); ok {
			issues = append(issues, validation.Issues...)
		} else {
			return err
		}
	}
	if err := ValidateManifestShape(manifest); err != nil {
		if validation, ok := err.(*ValidationErrors); ok {
			issues = append(issues, validation.Issues...)
		} else {
			return err
		}
	}
	known := make(map[string]SkillDeclaration, len(registry.Skills))
	for _, skill := range registry.Skills {
		known[skill.Name] = skill
	}
	selected := make(map[string]string)
	for _, profileName := range manifest.Profiles {
		profile, ok := registry.Profiles[profileName]
		if !ok {
			addIssue(&issues, IssueMissingReference, "manifest.profiles."+profileName, "unknown profile")
			continue
		}
		for _, name := range profile.Skills {
			if _, exists := selected[name]; exists {
				addIssue(&issues, IssueCollision, "manifest.profiles."+profileName+".skills."+name, "skill collides with %s", selected[name])
			} else {
				selected[name] = "profile " + profileName
			}
			if _, ok := known[name]; !ok {
				addIssue(&issues, IssueMissingReference, "manifest.profiles."+profileName+".skills."+name, "references unknown skill")
			}
		}
	}
	for index, direct := range manifest.Direct {
		path := fmt.Sprintf("manifest.direct[%d].name", index)
		if origin, exists := selected[direct.Name]; exists {
			addIssue(&issues, IssueCollision, path, "skill collides with %s", origin)
		}
		if contains(registry.Global.Baseline, direct.Name) {
			addIssue(&issues, IssueCollision, path, "direct skill collides with the fixed global baseline")
		}
		selected[direct.Name] = "direct"
	}
	return newValidationErrors(issues)
}

func validateNamesList(path string, names []string, issues *[]Issue) {
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if !isPortableName(name) {
			addIssue(issues, IssueInvalidName, path, "must contain only portable skill names; got %q", name)
		}
		if seen[name] {
			addIssue(issues, IssueDuplicate, path, "duplicate skill %q", name)
		}
		seen[name] = true
	}
	if !isSortedStrings(names) {
		addIssue(issues, IssueInvalidName, path, "must be sorted by name")
	}
}

// SkillsCLIPathProblem returns an explanatory problem or an empty string when
// location can be passed to the pinned Skills CLI materialization adapter.
func SkillsCLIPathProblem(location string) string {
	if location == "" {
		return "source is empty"
	}
	if strings.HasPrefix(location, ".") || strings.HasPrefix(location, "/") || strings.HasPrefix(location, "~") || strings.Contains(location, `\`) || (len(location) > 1 && location[1] == ':') {
		return "local paths are not allowed"
	}
	if shorthandPattern.MatchString(location) {
		for _, segment := range strings.Split(location, "/") {
			if segment == "." || segment == ".." || segment == "" {
				return "path traversal is not allowed"
			}
		}
		return ""
	}
	parsed, err := url.Parse(location)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Host == "" {
		return "source must be a git shorthand or credential-free https remote"
	}
	return ""
}

func isRemoteSource(location string) bool { return SkillsCLIPathProblem(location) == "" }

func sortedSourceIDs(sources map[string]Source) []string { return sortedStringKeys(sources) }

func sortedStringKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
