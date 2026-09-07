package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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
func (a *application) collectStatus(ctx context.Context) []sjskills.Advisory {
	if a.envelope.Result != sjskills.ResultSuccess || ctx.Err() != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, sjskills.StatusRefreshBudget)
	defer cancel()
	registry, err := a.registry()
	if err != nil {
		return []sjskills.Advisory{statusUnavailable(sjskills.ScopeGlobal, "skill registry unavailable")}
	}
	results := make([]*sjskills.Advisory, 2)
	var group sync.WaitGroup
	// There are exactly two independent scopes and one shared foreground budget.
	for index, global := range []bool{false, true} {
		group.Add(1)
		go func(index int, global bool) {
			defer group.Done()
			directory := a.directory
			scopeName := sjskills.ScopeProject
			if global {
				scopeName = sjskills.ScopeGlobal
				var homeErr error
				directory, homeErr = a.selectedGlobalHome()
				if homeErr != nil {
					v := statusUnavailable(scopeName, "global home unavailable")
					results[index] = &v
					return
				}
			}
			scope, resolveErr := sjskills.ResolveStatusScope(directory, registry, global)
			if resolveErr != nil {
				if !global && missingStatusManifest(resolveErr) {
					return
				}
				v := statusUnavailable(scopeName, "scope configuration could not be inspected")
				results[index] = &v
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
	values := []sjskills.Advisory{}
	for _, result := range results {
		if result != nil {
			values = append(values, *result)
		}
	}
	return values
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
		if value.Error != "" {
			fmt.Fprintf(output, "sjskills: %s — %s%s\n", value.Scope, value.Error, suffix)
		}
		// Only a fresh, explicit plan already renders this scope's full findings.
		if explicit != nil && explicit.Desired.Scope == value.Scope && value.Freshness == sjskills.AdvisoryFresh {
			continue
		}
		printed := false
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
			fmt.Fprintf(output, "sjskills: %s — %s: %s%s\n", value.Scope, label, list, suffix)
			printed = true
		}
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
