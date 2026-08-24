package sjskills

import (
	"fmt"
	"sort"
)

// Resolve dispatches the two pure desired-state scopes behind one stable
// request boundary for later filesystem/process adapters.
func Resolve(request ResolveRequest) (DesiredState, error) {
	if request.Global {
		return ResolveGlobal(request.Registry)
	}
	if request.Manifest == nil {
		return DesiredState{}, &ValidationErrors{Issues: []Issue{{
			Code: IssueEmptySelection, Path: "manifest", Message: "project resolution requires a manifest",
		}}}
	}
	return ResolveProject(request.Registry, *request.Manifest)
}

// BuildPlan derives stable warnings/evidence without reading files, invoking
// processes, using credentials, or contacting a source.
func BuildPlan(request ResolveRequest) (Plan, error) {
	desired, err := Resolve(request)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{
		Desired:    desired,
		Operations: []PlanOperation{},
		Warnings:   []Warning{},
		Evidence: []Evidence{
			{Kind: "resolution", Detail: fmt.Sprintf("resolved %d desired skills", len(desired.Skills))},
		},
	}
	for _, skill := range desired.Skills {
		switch skill.Manager {
		case ManagerManual:
			plan.Warnings = append(plan.Warnings, Warning{Code: "manual-action", Message: fmt.Sprintf("%s requires its recorded manual provisioning procedure", skill.Name)})
		case ManagerWorkflow:
			plan.Warnings = append(plan.Warnings, Warning{Code: "workflow-action", Message: fmt.Sprintf("%s is provided by workflow %s", skill.Name, skill.Workflow)})
		}
	}
	return plan, nil
}

// ResolveGlobal selects only the fixed baseline. It is intentionally
// independent of project manifests and contains no machine/profile input.
func ResolveGlobal(registry Registry) (DesiredState, error) {
	if err := ValidateRegistry(registry); err != nil {
		return DesiredState{}, err
	}
	state := DesiredState{Scope: ScopeGlobal, Skills: make([]DesiredSkill, 0, len(registry.Global.Baseline))}
	seen := make(map[string]string, len(registry.Global.Baseline))
	for _, name := range registry.Global.Baseline {
		if previous, exists := seen[name]; exists {
			return DesiredState{}, collisionError("global baseline", name, previous)
		}
		seen[name] = "global baseline"
		declaration := declarationByName(registry, name)
		state.Skills = append(state.Skills, resolveCentralSkill(registry, declaration, ScopeGlobal, "global baseline"))
	}
	sortDesiredSkills(state.Skills)
	return state, nil
}

// ResolveProject unions selected profile sets and direct entries. Every name
// is checked before a desired result is returned, so contradictory/duplicate
// sources cannot reach a later materialization adapter.
func ResolveProject(registry Registry, manifest Manifest) (DesiredState, error) {
	if err := ValidateManifest(registry, manifest); err != nil {
		return DesiredState{}, err
	}
	state := DesiredState{Scope: ScopeProject}
	selected := make(map[string]string)
	declarations := make(map[string]SkillDeclaration, len(registry.Skills))
	for _, declaration := range registry.Skills {
		declarations[declaration.Name] = declaration
	}
	for _, profileName := range manifest.Profiles {
		for _, name := range registry.Profiles[profileName].Skills {
			origin := "profile:" + profileName
			if previous, exists := selected[name]; exists {
				return DesiredState{}, collisionError(origin, name, previous)
			}
			selected[name] = origin
			state.Skills = append(state.Skills, resolveCentralSkill(registry, declarations[name], ScopeProject, origin))
		}
	}
	for _, direct := range manifest.Direct {
		if previous, exists := selected[direct.Name]; exists {
			return DesiredState{}, collisionError("direct", direct.Name, previous)
		}
		selected[direct.Name] = "direct"
		state.Skills = append(state.Skills, resolveDirectSkill(registry, direct))
	}
	sortDesiredSkills(state.Skills)
	return state, nil
}

func declarationByName(registry Registry, name string) SkillDeclaration {
	for _, declaration := range registry.Skills {
		if declaration.Name == name {
			return declaration
		}
	}
	return SkillDeclaration{Name: name}
}

func resolveCentralSkill(registry Registry, declaration SkillDeclaration, scope Scope, origin string) DesiredSkill {
	source := registry.Sources[declaration.Source]
	return DesiredSkill{
		Name:      declaration.Name,
		SourceID:  declaration.Source,
		Source:    source.Location,
		Scope:     scope,
		Origin:    origin,
		Manager:   declaration.Manager,
		Mode:      declaration.Mode,
		Workflow:  declaration.Workflow,
		FullDepth: declaration.FullDepth,
		Targets:   effectiveTargets(registry, declaration.Name),
	}
}

func resolveDirectSkill(registry Registry, direct DirectSkill) DesiredSkill {
	return DesiredSkill{
		Name:      direct.Name,
		Source:    direct.Source,
		Scope:     ScopeProject,
		Origin:    "direct",
		Manager:   ManagerSkillsCLI,
		Mode:      ModeCopy,
		FullDepth: direct.FullDepth,
		Targets:   copyTargets(registry.Defaults.Targets),
	}
}

func effectiveTargets(registry Registry, name string) []Target {
	if targets, ok := registry.TargetExceptions[name]; ok {
		return copyTargets(targets)
	}
	return copyTargets(registry.Defaults.Targets)
}

func sortDesiredSkills(skills []DesiredSkill) {
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Name != skills[j].Name {
			return skills[i].Name < skills[j].Name
		}
		return skills[i].Origin < skills[j].Origin
	})
	for index := range skills {
		skills[index].Targets = copyTargets(skills[index].Targets)
	}
}

func collisionError(origin, name, previous string) error {
	return &ValidationErrors{Issues: []Issue{{
		Code:    IssueCollision,
		Path:    fmt.Sprintf("desired.%s", name),
		Message: fmt.Sprintf("%s collides with %s", origin, previous),
	}}}
}
