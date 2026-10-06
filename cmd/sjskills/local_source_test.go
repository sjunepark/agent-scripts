package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLocalFixtureSkill(t *testing.T, path, name, body string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Fixture\n---\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func localPlanActions(t *testing.T, directory string, overrides map[string]string) (map[string]string, sjskillsEnvelope) {
	t.Helper()
	code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "plan")
	envelope := decodeEnvelope(t, stdout)
	if code != 0 {
		return nil, envelope
	}
	if stderr != "" {
		t.Fatalf("plan stderr=%q", stderr)
	}
	actions := map[string]string{}
	for _, operation := range envelope.Plan.Operations {
		actions[operation.Skill+"@"+operation.Target] = operation.Action + "/" + operation.Reason
	}
	return actions, envelope
}

func TestExternalLocalDirectSourceLifecycle(t *testing.T) {
	project := t.TempDir()
	source := filepath.Join(project, "skills", "team-tool")
	writeLocalFixtureSkill(t, source, "team-tool", "# v1\n")
	if err := os.WriteFile(filepath.Join(project, "sjskills.toml"), []byte("version = 1\n\n[[direct]]\nname = \"team-tool\"\nsource = \"./skills/team-tool\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	overrides := isolatedExternalHomes(t)
	// Run from a nested directory: relative sources resolve against the
	// manifest, never the process working directory.
	directory := filepath.Join(project, "skills")

	expectActions := func(want string) {
		t.Helper()
		actions, envelope := localPlanActions(t, directory, overrides)
		if actions == nil {
			t.Fatalf("plan failed: %#v", envelope.Error)
		}
		if len(actions) != 2 || actions["team-tool@.agents"] != want || actions["team-tool@.claude"] != want {
			t.Fatalf("plan actions = %#v, want %s for both targets", actions, want)
		}
	}
	apply := func() {
		t.Helper()
		if code, stdout, stderr := runCLIWithEnvironment(t, directory, overrides, "--json", "apply", "--yes"); code != 0 {
			t.Fatalf("apply code=%d stdout=%s stderr=%s", code, stdout, stderr)
		}
	}

	expectActions("install/expected-entry-absent")
	apply()
	expectActions("unchanged/verified-exact")
	for _, target := range []string{".agents", ".claude"} {
		data, err := os.ReadFile(filepath.Join(project, target, "skills", "team-tool", "SKILL.md"))
		if err != nil || !strings.HasSuffix(string(data), "# v1\n") {
			t.Fatalf("%s placement = %q err=%v", target, data, err)
		}
	}
	var state struct {
		Records []struct {
			SourceIdentity string `json:"sourceIdentity"`
		} `json:"records"`
	}
	data, err := os.ReadFile(filepath.Join(project, ".sjskills", "state.json"))
	if err != nil || json.Unmarshal(data, &state) != nil || len(state.Records) != 2 || state.Records[0].SourceIdentity != "local:./skills/team-tool" {
		t.Fatalf("provenance = %s err=%v", data, err)
	}

	writeLocalFixtureSkill(t, source, "team-tool", "# v2\n")
	expectActions("update/verified-update")
	apply()
	expectActions("unchanged/verified-exact")

	if err := os.RemoveAll(source); err != nil {
		t.Fatal(err)
	}
	actions, envelope := localPlanActions(t, directory, overrides)
	if actions != nil || envelope.Result != "invalid" || envelope.Error == nil || envelope.Error.Code != "invalid_source" ||
		!strings.Contains(envelope.Error.Message, `local source "./skills/team-tool" does not exist`) || !strings.Contains(envelope.Error.Message, "[[direct]]") {
		t.Fatalf("missing source plan = %#v error=%#v", envelope, envelope.Error)
	}
	if _, err := os.Stat(filepath.Join(project, ".claude", "skills", "team-tool", "SKILL.md")); err != nil {
		t.Fatalf("blocked plan changed the placement: %v", err)
	}
}
