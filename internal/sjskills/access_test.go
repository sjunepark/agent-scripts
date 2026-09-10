package sjskills

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAccessProfileSelectionAndPublicCompatibility(t *testing.T) {
	r := fixtureRegistry(t)
	public, err := ResolveProject(r, Manifest{Version: 1, Profiles: []string{"kicpa"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(public.Skills) != 5 {
		t.Fatalf("public profile changed: %+v", public)
	}
	for _, s := range public.Skills {
		if s.Access != AccessPublic {
			t.Fatal(s)
		}
	}
	mixed, err := ResolveProject(r, Manifest{Version: 1, Profiles: []string{"kicpa", "kicpa-private"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(mixed.Skills) != 12 {
		t.Fatal(mixed)
	}
	for _, s := range mixed.Skills {
		if (s.Access == AccessGitHubAuthenticated) != (s.SourceID == "kicpa") {
			t.Fatal(s)
		}
	}
	global, err := ResolveGlobal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range global.Skills {
		if s.Access != AccessPublic {
			t.Fatal(s)
		}
	}
	for name, profile := range r.Profiles {
		if profile.Access == AccessPublic {
			profile.Access = ""
			r.Profiles[name] = profile
		}
	}
	omitted, err := ResolveProject(r, Manifest{Version: 1, Profiles: []string{"kicpa"}})
	if err != nil {
		t.Fatal(err)
	}
	for i := range omitted.Skills {
		if !sameDesiredSkill(omitted.Skills[i], public.Skills[i]) {
			t.Fatal("omitted public changed identity")
		}
	}
}

func TestAccessDecodingRejectsMalformedAndPreservesDirect(t *testing.T) {
	for _, input := range []string{`null`, `""`, `"private"`, `true`, `1`, `[]`, `{}`} {
		var profile Profile
		if err := json.Unmarshal([]byte(`{"access":`+input+`,"skills":["test"]}`), &profile); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
	for _, value := range []string{`""`, `"private"`, `true`, `1`, `[]`} {
		_, err := ParseManifest([]byte("version=1\n[[direct]]\nname=\"fixture\"\nsource=\"owner/repo\"\naccess=" + value + "\n"))
		if err == nil {
			t.Fatalf("accepted TOML %s", value)
		}
	}
	for _, access := range []string{"", "access=\"public\"\n", "access=\"github-authenticated\"\n"} {
		m, err := ParseManifest([]byte("version=1\n[[direct]]\nname=\"fixture\"\nsource=\"owner/repo\"\n" + access))
		if err != nil {
			t.Fatal(err)
		}
		desired, err := ResolveProject(fixtureRegistry(t), m)
		if err != nil {
			t.Fatal(err)
		}
		if (desired.Skills[0].Access == AccessGitHubAuthenticated) != strings.Contains(access, "github-authenticated") {
			t.Fatal(desired)
		}
	}
}

func TestAuthenticatedSourcesAreBoundedToGitHub(t *testing.T) {
	for _, source := range []string{"owner/repo", "owner/repo/skills/fixture", "https://github.com/owner/repo.git", "https://github.com/owner/repo/tree/main/skills"} {
		if repo, err := githubRepository(source); err != nil || repo != "owner/repo" {
			t.Fatalf("%s: %s %v", source, repo, err)
		}
	}
	for _, source := range []string{"https://example.com/o/r", "https://github.com:443/o/r", "https://github.com.evil/o/r", "https://u:secret@github.com/o/r", "https://github.com/o/r?token=secret", "https://github.com/o/r#main", "https://github.com/o/r/%2e%2e/a", "https://github.com/o/r/../a", "https://github.com/o/r/a%2Fb", "https://github.com/o/r\\a", "./o/r", "owner/..", "owner/.git", "https://github.com/o/r/blob/main/skills", "https://github.com/o/r/tree", "https://github.com/o/r/issues"} {
		if _, err := githubRepository(source); err == nil {
			t.Fatalf("accepted %q", source)
		}
	}
	r := fixtureRegistry(t)
	p := r.Profiles["kicpa"]
	p.Access = AccessGitHubAuthenticated
	r.Profiles["kicpa"] = p
	if ValidateRegistry(r) == nil {
		t.Fatal("required public profile allowed authentication")
	}
	r = fixtureRegistry(t)
	p = r.Profiles["kicpa-private"]
	delete(r.Profiles, "kicpa-private")
	r.Profiles["other-private"] = p
	if err := ValidateRegistry(r); err != nil {
		t.Fatal(err)
	}
	r.Profiles["bad_name"] = p
	if ValidateRegistry(r) == nil {
		t.Fatal("accepted malformed additional profile")
	}
}

func TestAccessSeparatesFetchAndReviewEvidenceButNotOwnership(t *testing.T) {
	a := DesiredSkill{Name: "fixture", Source: "owner/repo", Manager: ManagerSkillsCLI, Mode: ModeCopy, Access: AccessPublic}
	b := a
	b.Access = AccessGitHubAuthenticated
	if sameDesiredSkill(a, b) {
		t.Fatal("access omitted from desired identity")
	}
	if _, _, err := classifyMaterializationSkills([]DesiredSkill{a, b}); err == nil {
		t.Fatal("contradictory fetch policy accepted")
	}
	b.Name = "other"
	if len(materializationBatches([]DesiredSkill{a, b})) != 2 {
		t.Fatal("access batches combined")
	}
	sa := StatusScope{Plan: Plan{Desired: DesiredState{Skills: []DesiredSkill{a}}}}
	b.Name = a.Name
	sb := StatusScope{Plan: Plan{Desired: DesiredState{Skills: []DesiredSkill{b}}}}
	if sa.cacheKey() == sb.cacheKey() {
		t.Fatal("authentication shares public cache evidence")
	}
	legacy := a
	legacy.Access = ""
	sl := StatusScope{Plan: Plan{Desired: DesiredState{Skills: []DesiredSkill{legacy}}}}
	if sa.cacheKey() != sl.cacheKey() {
		t.Fatal("omitted public changes cache key")
	}
	ea := Envelope{Plan: &sa.Plan}
	el := Envelope{Plan: &sl.Plan}
	na := normalizeReviewedEnvelope(ea)
	nl := normalizeReviewedEnvelope(el)
	if !sameDesiredSkill(na.Plan.Desired.Skills[0], nl.Plan.Desired.Skills[0]) {
		t.Fatal("legacy review not normalized")
	}
	if el.Plan.Desired.Skills[0].Access != "" {
		t.Fatal("normalization mutated input")
	}
}

func TestAccessStatusSnapshotsNormalizePublic(t *testing.T) {
	scope, _, _, _ := statusFixture(t, false)
	other := scope
	other.Registry.Profiles = map[string]Profile{}
	for name, profile := range scope.Registry.Profiles {
		profile.Access = ""
		other.Registry.Profiles[name] = profile
	}
	other.Plan.Desired = cloneDesiredState(scope.Plan.Desired)
	for i := range other.Plan.Desired.Skills {
		other.Plan.Desired.Skills[i].Access = ""
	}
	if !scope.Matches(other) {
		t.Fatal("omitted public changed command snapshot identity")
	}
	other.Plan.Desired.Skills[0].Access = AccessGitHubAuthenticated
	if scope.Matches(other) {
		t.Fatal("authenticated policy reused command snapshot")
	}
}

func TestAuthenticatedNonInstallableSkillsAreSkipped(t *testing.T) {
	for _, manager := range []Manager{ManagerManual, ManagerWorkflow} {
		installed, skipped, err := classifyMaterializationSkills([]DesiredSkill{{Name: "manual-fixture", Manager: manager, Access: AccessGitHubAuthenticated}})
		if err != nil || len(installed) != 0 || len(skipped) != 1 {
			t.Fatalf("%s: installed=%v skipped=%v err=%v", manager, installed, skipped, err)
		}
	}
}
