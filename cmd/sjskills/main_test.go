package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testBinary string

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "sjskills-cli-test-")
	if err != nil {
		panic(err)
	}
	testBinary = filepath.Join(directory, "sjskills")
	build := exec.Command("go", "build", "-o", testBinary, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic(fmt.Sprintf("build test binary: %v", err))
	}
	code := m.Run()
	_ = os.RemoveAll(directory)
	os.Exit(code)
}

func runCLI(t *testing.T, directory string, args ...string) (int, string, string) {
	t.Helper()
	command := exec.Command(testBinary, args...)
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), stdout.String(), stderr.String()
	}
	t.Fatalf("run %v: %v", args, err)
	return -1, "", ""
}

func TestExternalHelpAndVersion(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--help")
	if code != 0 || !strings.Contains(stdout, "sjskills <command>") || stderr != "" {
		t.Fatalf("help code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--version")
	if code != 0 || stdout != "sjskills 1.0.0\n" || stderr != "" {
		t.Fatalf("version code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalJSONEnvelopeAndMalformedInvocation(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--json", "profiles")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("profiles code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var envelope sjskillsEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("profiles JSON: %v", err)
	}
	if envelope.Operation != "profiles" || envelope.Result != "success" {
		t.Fatalf("profiles envelope = %#v", envelope)
	}
	var profileShape map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &profileShape); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"operation", "result", "error", "warnings", "evidence"} {
		if _, ok := profileShape[field]; !ok {
			t.Fatalf("profiles JSON missing stable field %q: %q", field, stdout)
		}
	}

	code, stdout, stderr = runCLI(t, directory, "--json", "unknown")
	if code != 64 || stderr != "" || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("malformed code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("malformed JSON: %v", err)
	}
	if envelope.Operation != "parse" || envelope.Result != "invalid" {
		t.Fatalf("malformed envelope = %#v", envelope)
	}

	code, stdout, stderr = runCLI(t, directory, "plan", "--dry")
	if code != 64 || stdout != "" || !strings.Contains(stderr, "unknown flag --dry") {
		t.Fatalf("unsupported flag code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanApplyAndRestoreContracts(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte("version = 1\nprofiles = [\"dev\", \"go\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, directory, "--json", "plan")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"operation":"plan"`) || !strings.Contains(stdout, `"plan":{"desired"`) || !strings.Contains(stdout, `"operations":[]`) || !strings.Contains(stdout, `"resolved 26 desired skills"`) {
		t.Fatalf("plan output = %q", stdout)
	}

	code, stdout, stderr = runCLI(t, directory, "--json", "apply", "--global")
	if code != 2 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"unavailable"`) || !strings.Contains(stdout, "not implemented") {
		t.Fatalf("apply code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "restore", "quarantine-1")
	if code != 2 || stderr != "" || !strings.Contains(stdout, `"result":"unavailable"`) {
		t.Fatalf("restore code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "restore")
	if code != 65 || stderr != "" || !strings.Contains(stdout, `"result":"invalid"`) {
		t.Fatalf("restore missing id code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanDiscoversNearestNestedProject(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "sjskills.toml")
	if err := os.WriteFile(manifestPath, []byte("version = 1\nprofiles = [\"kicpa\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	canonicalManifest, err := filepath.EvalSymlinks(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, nested, "--json", "plan")
	if code != 0 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"path":"`+canonicalManifest+`"`) {
		t.Fatalf("nested plan code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalPlanRejectsUnsupportedDirectFields(t *testing.T) {
	directory := t.TempDir()
	manifest := "version = 1\n\n[[direct]]\nname = \"third-party-review\"\nsource = \"example/third-party-review\"\nmanager = \"skills-cli\"\n"
	if err := os.WriteFile(filepath.Join(directory, "sjskills.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, directory, "--json", "plan")
	if code != 65 || stderr != "" || strings.Count(stdout, "\n") != 1 || !strings.Contains(stdout, `"result":"invalid"`) || !strings.Contains(stdout, `"code":"unknown_field"`) {
		t.Fatalf("unsupported direct field code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExternalInitRefusesOverwrite(t *testing.T) {
	directory := t.TempDir()
	code, stdout, stderr := runCLI(t, directory, "--json", "init", "dev", "go")
	if code != 0 || stderr != "" || !strings.Contains(stdout, `"operation":"init"`) {
		t.Fatalf("init code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	path := filepath.Join(directory, "sjskills.toml")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runCLI(t, directory, "--json", "init", "kicpa")
	if code != 2 || stderr != "" || !strings.Contains(stdout, `"result":"conflict"`) {
		t.Fatalf("second init code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("existing manifest was changed: before=%q after=%q", before, after)
	}
}

type sjskillsEnvelope struct {
	Operation string `json:"operation"`
	Result    string `json:"result"`
}
