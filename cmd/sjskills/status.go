package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

type commandStatusSnapshot struct {
	scope    sjskills.StatusScope
	snapshot sjskills.StatusSnapshot
}

func (a *application) finishPrepared(p *preparedPlan, envelope sjskills.Envelope, stage string) sjskills.Envelope {
	envelope = p.finish(envelope, stage)
	if envelope.Result == sjskills.ResultSuccess && p.verified && p.cleaned {
		registry, err := a.registry()
		if err == nil {
			root := ""
			if p.project != nil {
				root = p.project.Layout.Root
			}
			if p.global != nil {
				root = p.global.Layout.Home
			}
			a.statusSnapshot = &commandStatusSnapshot{sjskills.StatusScope{Root: root, Registry: registry, Plan: p.plan}, sjskills.StatusSnapshot{Expected: copyExpected(p.expected), ObservedAt: p.observedAt}}
		}
	}
	return envelope
}

// statusCollection keeps discovery metadata and scope evidence together; only the
// status operation exposes metadata, while incidental callers use advisories.
type statusCollection struct {
	result     sjskills.StatusResult
	advisories []sjskills.Advisory
}

func (a *application) status(ctx context.Context) sjskills.Envelope {
	envelope := a.base(sjskills.CommandOperationStatus)
	if a.noStatusCheck {
		envelope.Status = &sjskills.StatusResult{ProjectConfiguration: sjskills.ProjectSkipped}
		return envelope
	}
	if ctx.Err() == nil {
		report := a.collectStatus(ctx)
		envelope.Status = &report.result
		envelope.Advisories = report.advisories
	}
	if ctx.Err() != nil {
		envelope.Result = sjskills.ResultUnavailable
		envelope.Error = &sjskills.Issue{Code: sjskills.IssueUnavailable, Message: "status inspection cancelled"}
	}
	return envelope
}

func discoverStatusProject(directory string) sjskills.StatusResult {
	result := sjskills.StatusResult{ProjectConfiguration: sjskills.ProjectUnavailable}
	// DiscoverProjectRoot accepts files and classifies invalid starts as missing.
	// A status invocation must establish an inspectable directory before setup advice.
	file, err := os.Open(directory)
	if err != nil {
		return result
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.IsDir() {
		return result
	}
	if _, err := file.Readdirnames(1); err != nil && err != io.EOF {
		return result
	}
	project, err := sjskills.DiscoverProjectRoot(directory)
	if missingStatusManifest(err) {
		result.ProjectConfiguration = sjskills.ProjectNotConfigured
	} else if err == nil {
		result.ProjectRoot = project.Root
		if _, err := sjskills.ReadManifest(project.ManifestPath); err == nil {
			result.ProjectConfiguration = sjskills.ProjectConfigured
		}
	}
	return result
}

func (a *application) collectStatus(ctx context.Context) statusCollection {
	ctx, cancel := context.WithTimeout(ctx, sjskills.StatusRefreshBudget)
	defer cancel()
	report := statusCollection{result: discoverStatusProject(a.directory)}
	registry, registryErr := a.registry()
	results := make([]*sjskills.Advisory, 2)
	var group sync.WaitGroup
	// There are exactly two independent scopes and one shared foreground budget.
	for index, global := range []bool{false, true} {
		group.Add(1)
		go func(index int, global bool) {
			defer group.Done()
			directory := a.directory
			scopeName := sjskills.ScopeProject
			unavailable := func(message string) {
				v := statusUnavailable(scopeName, message)
				results[index] = &v
			}
			if global {
				scopeName = sjskills.ScopeGlobal
				var homeErr error
				directory, homeErr = a.selectedGlobalHome()
				if homeErr != nil {
					unavailable("global home unavailable")
					return
				}
			} else {
				switch report.result.ProjectConfiguration {
				case sjskills.ProjectNotConfigured:
					return
				case sjskills.ProjectUnavailable:
					unavailable("scope configuration could not be inspected")
					return
				}
			}
			if registryErr != nil {
				if !global {
					report.result.ProjectConfiguration = sjskills.ProjectUnavailable
				}
				unavailable("skill registry unavailable")
				return
			}
			scope, resolveErr := sjskills.ResolveStatusScope(directory, registry, global)
			if resolveErr != nil {
				if !global {
					report.result.ProjectConfiguration = sjskills.ProjectUnavailable
				}
				unavailable("scope configuration could not be inspected")
				return
			}
			var reusable *sjskills.StatusSnapshot
			if a.statusSnapshot != nil && scope.Matches(a.statusSnapshot.scope) {
				reusable = &a.statusSnapshot.snapshot
			}
			service := sjskills.StatusService{}
			if a.statusService != nil {
				service = *a.statusService
			}
			v := service.Check(ctx, scope, reusable)
			results[index] = &v
		}(index, global)
	}
	group.Wait()
	for _, result := range results {
		if result != nil {
			report.advisories = append(report.advisories, *result)
		}
	}
	return report
}
func missingStatusManifest(err error) bool {
	var issues *sjskills.ValidationErrors
	return errors.As(err, &issues) && len(issues.Issues) == 1 && issues.Issues[0].Code == sjskills.IssueManifestMissing
}
func statusUnavailable(scope sjskills.Scope, message string) sjskills.Advisory {
	command := "sjskills plan"
	if scope == sjskills.ScopeGlobal {
		command += " --global"
	}
	return sjskills.Advisory{Scope: scope, Freshness: sjskills.AdvisoryUnavailable, Findings: []sjskills.AdvisoryFinding{}, Error: message, ReviewCommand: command}
}
func renderStatus(output io.Writer, values []sjskills.Advisory, explicit *sjskills.Plan, now time.Time) {
	for _, value := range values {
		suffix := statusSuffix(value, now)
		if value.Error != "" {
			fmt.Fprintf(output, "sjskills: %s — %s%s\n", value.Scope, value.Error, suffix)
		}
		// Only a fresh, explicit plan already renders this scope's full findings.
		if explicit != nil && explicit.Desired.Scope == value.Scope && value.Freshness == sjskills.AdvisoryFresh {
			continue
		}
		printed := false
		renderStatusFindings(value, func(line string) {
			fmt.Fprintf(output, "sjskills: %s — %s%s\n", value.Scope, line, suffix)
			printed = true
		})
		if printed {
			fmt.Fprintf(output, "sjskills: review with `%s`.\n", value.ReviewCommand)
		}
	}
}
func statusAge(age time.Duration) string {
	if age < time.Minute {
		return "<1m"
	}
	if age < time.Hour {
		return fmt.Sprintf("%dm", int(age/time.Minute))
	}
	if age < 24*time.Hour {
		return fmt.Sprintf("%dh", int(age/time.Hour))
	}
	return fmt.Sprintf("%dd", int(age/(24*time.Hour)))
}

func statusSuffix(value sjskills.Advisory, now time.Time) string {
	suffix := ""
	if value.Cached && value.ObservedAt != nil {
		age := now.Sub(*value.ObservedAt)
		if age < 0 {
			suffix = " (cached observation is in the future)"
		} else {
			suffix = " (checked " + statusAge(age) + " ago)"
		}
	}
	if value.Freshness == sjskills.AdvisoryStale {
		suffix += " (stale upstream evidence)"
	}
	return suffix
}

func renderStatusFindings(value sjskills.Advisory, emit func(string)) {
	for _, category := range []sjskills.AdvisoryCategory{sjskills.AdvisoryUpdate, sjskills.AdvisoryMissing, sjskills.AdvisoryExtra, sjskills.AdvisoryConflict} {
		names := map[string]bool{}
		for _, finding := range value.Findings {
			if finding.Category == category {
				name := finding.Skill
				if name == "" {
					name = "scope state"
				}
				names[name] = true
			}
		}
		if len(names) == 0 {
			continue
		}
		ordered := make([]string, 0, len(names))
		for name := range names {
			ordered = append(ordered, name)
		}
		sort.Strings(ordered)
		count := len(ordered)
		if count > 5 {
			ordered = ordered[:5]
		}
		for index, name := range ordered {
			quoted := strconv.QuoteToASCII(name)
			if quoted != "\""+name+"\"" {
				ordered[index] = quoted
			}
		}
		list := strings.Join(ordered, ", ")
		if count > 5 {
			list += fmt.Sprintf(" (+%d more)", count-5)
		}
		label := map[sjskills.AdvisoryCategory]string{sjskills.AdvisoryUpdate: "updates available", sjskills.AdvisoryMissing: "missing", sjskills.AdvisoryExtra: "undeclared extras", sjskills.AdvisoryConflict: "conflicts need attention"}[category]
		emit(label + ": " + list)
	}
}

func renderStatusReport(output io.Writer, result sjskills.StatusResult, values []sjskills.Advisory, now time.Time) {
	if result.ProjectConfiguration == sjskills.ProjectSkipped {
		fmt.Fprintln(output, "Status checks disabled (--no-status-check).")
		return
	}
	if result.ProjectConfiguration == sjskills.ProjectNotConfigured {
		fmt.Fprintln(output, "Project: not configured\n  Run `sjskills profiles` to choose profiles, then `sjskills init <profile>...`.")
	}
	for _, value := range values {
		label := "Global"
		if value.Scope == sjskills.ScopeProject {
			label = "Project"
			if result.ProjectRoot != "" {
				label += " (" + strconv.QuoteToASCII(result.ProjectRoot) + ")"
			}
		}
		summary := "drift detected"
		switch {
		case value.Freshness == sjskills.AdvisoryUnavailable:
			summary = "inspection unavailable"
		case len(value.Findings) == 0:
			summary = "no drift detected"
		}
		if value.Freshness == sjskills.AdvisoryStale {
			summary += " against stale evidence"
		}
		fmt.Fprintf(output, "%s: %s%s\n", label, summary, statusSuffix(value, now))
		if value.Error != "" {
			fmt.Fprintf(output, "  %s\n", value.Error)
		}
		renderStatusFindings(value, func(line string) { fmt.Fprintf(output, "  %s\n", line) })
		if len(value.Findings) != 0 || value.Freshness != sjskills.AdvisoryFresh {
			fmt.Fprintf(output, "  Review with `%s`.\n", value.ReviewCommand)
		}
		if value.Scope == sjskills.ScopeProject && result.ProjectConfiguration == sjskills.ProjectUnavailable {
			fmt.Fprintln(output, "  Review the project directory and repair any existing sjskills.toml configuration.")
		}
	}
}
