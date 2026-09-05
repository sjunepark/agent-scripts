package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

func TestExternalStrictSyncRemovesUnownedExtras(t *testing.T) {
	for _, global := range []bool{false, true} {
		name := "project"
		if global {
			name = "global"
		}
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = []\n[[direct]]\nname = \"fixture-skill\"\nsource = \"example/fixture-skill\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			overrides := isolatedExternalHomes(t)
			root := directory
			if global {
				key := "HOME"
				if runtime.GOOS == "windows" {
					key = "USERPROFILE"
				}
				root = overrides[key]
			}
			for _, target := range []string{".agents", ".claude"} {
				path := filepath.Join(root, target, "skills", "project-status")
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("# Old Project Status\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--json", "plan"}
			if global {
				args = append(args, "--global")
			}
			code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, args...)
			if code != 0 {
				t.Fatalf("plan: code=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			plan := decodeEnvelope(t, stdout)
			removals := 0
			for _, op := range plan.Plan.Operations {
				if op.Skill == "project-status" && op.Action == string(sjskills.PlanActionQuarantine) {
					removals++
				}
			}
			if removals != 2 {
				t.Fatalf("strict sync requires 2 removals; got %d: %s", removals, stdout)
			}
			args = []string{"--json", "apply", "--yes"}
			if global {
				approved, digest := writeReviewedPlan(t, []byte(stdout))
				args = append(args, "--global", "--approved-plan", approved, "--approved-plan-sha256", digest)
			}
			code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, args...)
			if code != 0 {
				t.Fatalf("apply: code=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			detail := serializedEvidenceDetail(decodeEnvelope(t, stdout).Evidence, "quarantine")
			id := strings.TrimSuffix(strings.TrimPrefix(detail, "id="), " status=committed")
			if !validQuarantineID(id) {
				t.Fatalf("missing quarantine: %s", stdout)
			}
			for _, target := range []string{".agents", ".claude"} {
				if _, err := os.Lstat(filepath.Join(root, target, "skills", "project-status")); !os.IsNotExist(err) {
					t.Fatalf("extra remains active: %s %v", target, err)
				}
			}
			args = []string{"--json", "restore", id, "--yes"}
			if global {
				args = append(args, "--global")
			}
			code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, args...)
			if code != 0 {
				t.Fatalf("restore: code=%d stdout=%s stderr=%s", code, stdout, stderr)
			}
			for _, target := range []string{".agents", ".claude"} {
				data, err := os.ReadFile(filepath.Join(root, target, "skills", "project-status", "SKILL.md"))
				if err != nil || string(data) != "# Old Project Status\n" {
					t.Fatalf("restore bytes: %q %v", data, err)
				}
			}
		})
	}
}

func TestExternalStrictSyncBlocksUninspectableUnselectedTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires privileges on Windows")
	}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"kicpa\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "SKILL.md"), []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, ".claude", "skills")); err != nil {
		t.Fatal(err)
	}
	overrides := isolatedExternalHomes(t)
	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "plan")
	if code != 0 {
		t.Fatalf("plan: %d %s %s", code, stdout, stderr)
	}
	plan := decodeEnvelope(t, stdout)
	blocked := false
	for _, op := range plan.Plan.Operations {
		if op.Action == "blocked" && op.Target == ".claude" && op.Reason == "root-unavailable" {
			blocked = true
		}
	}
	if !blocked {
		t.Fatalf("unsafe target absent from plan: %s", stdout)
	}
	before := captureFixtureTree(t, directory)
	code, stdout, stderr = runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes")
	if code == 0 {
		t.Fatalf("unsafe target accepted: %s %s", stdout, stderr)
	}
	if after := captureFixtureTree(t, directory); !reflect.DeepEqual(before, after) {
		t.Fatal("blocked sync mutated project")
	}
	data, err := os.ReadFile(filepath.Join(outside, "SKILL.md"))
	if err != nil || string(data) != "external\n" {
		t.Fatalf("external content changed: %q %v", data, err)
	}
}

func TestRenderBlockedOperationEscapesUnsafeName(t *testing.T) {
	var output bytes.Buffer
	renderPlanOperations(&output, []sjskills.PlanOperation{{Action: sjskills.PlanActionBlocked, Skill: "bad\nname", Target: sjskills.TargetAgents, Reason: "current-entry-unverifiable"}})
	if strings.Count(output.String(), "\n") != 1 || !strings.Contains(output.String(), `bad\nname`) {
		t.Fatalf("unsafe output: %q", output.String())
	}
}
