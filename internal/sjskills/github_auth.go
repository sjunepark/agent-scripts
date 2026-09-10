package sjskills

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const GitHubCredentialCommand = "__github-credential"

// ghConfigDirectory resolves only the original gh configuration location. The
// materializer never copies a user's home, Git configuration, or credentials.
func ghConfigDirectory(env []string, platform string) (string, error) {
	get := func(name string) string { return environmentValue(env, name, platform) }
	directory := get("GH_CONFIG_DIR")
	if directory == "" && get("XDG_CONFIG_HOME") != "" {
		directory = filepath.Join(get("XDG_CONFIG_HOME"), "gh")
	}
	if directory == "" && isWindowsPlatform(platform) && get("APPDATA") != "" {
		directory = filepath.Join(get("APPDATA"), "GitHub CLI")
	}
	if directory == "" {
		home := get("HOME")
		if home == "" && isWindowsPlatform(platform) {
			home = get("USERPROFILE")
		}
		if home != "" {
			directory = filepath.Join(home, ".config", "gh")
		}
	}
	if directory == "" || !filepath.IsAbs(directory) {
		return "", errors.New("cannot locate absolute gh configuration; set GH_CONFIG_DIR")
	}
	return directory, nil
}

func environmentValue(env []string, name, platform string) string {
	value := ""
	for _, entry := range env {
		key, candidate, ok := strings.Cut(entry, "=")
		if ok && (key == name || isWindowsPlatform(platform) && strings.EqualFold(key, name)) {
			value = candidate
		}
	}
	return value
}

// gh migrates legacy configuration even for credential-helper invocations.
// Refuse it before starting gh: authentication must never perform that write.
// Decode YAML rather than guessing at quoting, aliases, or duplicate keys.
func checkGHConfig(directory string) error {
	data, err := readBoundedFile(filepath.Join(directory, "config.yml"), 1024*1024)
	if err != nil {
		return errors.New("gh configuration is missing or unreadable; run gh auth status to complete login/configuration setup, then retry")
	}
	var config map[string]any
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&config); err != nil {
		return errors.New("gh configuration is malformed; repair it using gh before retrying")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("gh configuration must contain one YAML document")
	}
	// gh's current multi-account migration has post-version 1. Future schemas
	// need a reviewed compatibility change, rather than assuming migration safety.
	if config["version"] != "1" && config["version"] != 1 {
		return errors.New("gh configuration requires migration; run gh auth status yourself, then retry (supported configuration version: 1)")
	}
	return nil
}

// credentialEnvironment excludes inherited Git configuration and diagnostics:
// only our exact repository helper may supply credentials, with no redirects,
// fallback protocols, caller hooks, or persisted Git configuration.
func credentialEnvironment(base []string, root, platform string) []string {
	filtered := make([]string, 0, len(base))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "GIT_") || strings.HasPrefix(upper, "GH_") || strings.HasPrefix(upper, "SJSKILLS_GH_") || upper == "SSH_ASKPASS" || upper == "GITHUB_TOKEN" {
			continue
		}
		filtered = append(filtered, entry)
	}
	env := isolatedEnvironment(filtered, root, platform)
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+filepath.Join(root, ".gitconfig"), "GIT_TERMINAL_PROMPT=0", "GIT_ALLOW_PROTOCOL=https", "GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "GH_NO_EXTENSION_UPDATE_NOTIFIER=1", "GH_TELEMETRY_DISABLED=1")
}

func gitShellQuote(value string) string {
	// Git executes ! helpers with its POSIX shell, including Git for Windows.
	return "'" + strings.ReplaceAll(filepath.ToSlash(value), "'", "'\\''") + "'"
}

func (m *Materializer) authenticatedSource(ctx context.Context, root, source string) (string, error) {
	repo, err := githubRepository(source)
	if err != nil {
		return "", err
	}
	gh, err := m.lookPath("gh")
	if err != nil {
		return "", errors.New("gh is unavailable; install GitHub CLI and sign in before fetching authenticated skills")
	}
	git, err := m.lookPath("git")
	if err != nil {
		return "", errors.New("git is unavailable for authenticated fetching")
	}
	config, err := ghConfigDirectory(m.baseEnv, m.platform)
	token := environmentValue(m.baseEnv, "GH_TOKEN", m.platform)
	if token == "" {
		token = environmentValue(m.baseEnv, "GITHUB_TOKEN", m.platform)
	}
	if token != "" {
		// Environment-backed gh login needs no persistent config. This temporary
		// config contains only a schema version, never the credential itself.
		config = filepath.Join(root, ".gh-auth")
		if err = os.MkdirAll(config, 0o700); err == nil {
			err = os.WriteFile(filepath.Join(config, "config.yml"), []byte("version: 1\n"), 0o600)
		}
	}
	if err != nil {
		return "", errors.New("cannot prepare gh configuration; set GH_CONFIG_DIR or sign in with gh")
	}
	if err = checkGHConfig(config); err != nil {
		return "", err
	}
	helper := m.helperExecutable
	if helper == "" {
		helper, err = os.Executable()
	}
	if err != nil || !filepath.IsAbs(helper) {
		return "", errors.New("cannot locate sjskills credential helper")
	}
	clone, err := os.MkdirTemp(root, "github-")
	if err != nil {
		return "", errors.New("cannot prepare private Git staging")
	}
	// Git clone expects an absent or empty destination. Each fetch has a distinct
	// directory; its lifetime and cleanup belong to the materialization plan.
	ref, subpath := githubSourcePath(source)
	env := credentialEnvironment(m.baseEnv, root, m.platform)
	if token != "" {
		env = append(env, "GH_TOKEN="+token)
	}
	env = append(env, "SJSKILLS_GH_EXECUTABLE="+gh, "SJSKILLS_GH_CONFIG_DIR="+config, "SJSKILLS_GH_REPOSITORY="+repo)
	helperCommand := "!" + gitShellQuote(helper) + " " + GitHubCredentialCommand
	args := []string{"-c", "credential.helper=", "-c", "credential.helper=" + helperCommand, "-c", "credential.useHttpPath=true", "-c", "http.followRedirects=false", "-c", "submodule.recurse=false", "-c", "core.hooksPath=" + filepath.Join(root, ".empty-hooks"), "clone", "--depth=1", "--single-branch", "--no-recurse-submodules"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, "--", "https://github.com/"+repo+".git", clone)
	result, err := m.runCommand(ctx, git, args, env)
	if err != nil || result.ExitCode != 0 {
		return "", privateProcessError("authenticated Git fetch failed; verify gh login, repository access, ref, and network", err)
	}
	local := filepath.Join(clone, filepath.FromSlash(subpath))
	if err := validateSymlinkParents(clone, local); err != nil {
		return "", errors.New("authenticated source subpath is missing or unsafe")
	}
	if err := checkRealDirectory(local); err != nil {
		return "", errors.New("authenticated source subpath is not a real directory")
	}
	return local, nil
}

// Match the supported subset of pinned Skills CLI source parsing. Shorthand
// suffixes are subpaths; GitHub /tree/<ref>/ URLs select a ref and optional path.
func githubSourcePath(source string) (ref, subpath string) {
	path := strings.TrimSuffix(strings.TrimPrefix(source, "https://github.com/"), "/")
	parts := strings.Split(path, "/")
	if strings.HasPrefix(source, "https://") {
		if len(parts) >= 4 && parts[2] == "tree" {
			return parts[3], strings.Join(parts[4:], "/")
		}
		return "", ""
	}
	return "", strings.Join(parts[2:], "/")
}

func privateProcessError(message string, err error) error {
	// External output is never diagnostic evidence for authenticated processes.
	// In particular a failed helper must not echo credentials through Git/CLI.
	switch {
	case errors.Is(err, errProcessTreeActive):
		return fmt.Errorf("%s: %w", message, errProcessTreeActive)
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", message, context.DeadlineExceeded)
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("%s: %w", message, context.Canceled)
	default:
		return errors.New(message)
	}
}

// RunGitHubCredential is the private Git helper entry point. The only output
// is Git's credential protocol, written directly into Git's private pipe.
// It intentionally emits no diagnostics and never handles store or erase.
func RunGitHubCredential(ctx context.Context, args []string, input io.Reader, output io.Writer, env []string) error {
	if len(args) != 1 {
		return errors.New("invalid credential operation")
	}
	if args[0] == "store" || args[0] == "erase" {
		return nil
	}
	if args[0] != "get" {
		return errors.New("invalid credential operation")
	}
	fields := map[string]string{}
	scanner := bufio.NewScanner(io.LimitReader(input, 8193))
	total := 0
	for scanner.Scan() {
		line := scanner.Text()
		total += len(line) + 1
		if total > 8192 {
			return errors.New("credential request exceeds bound")
		}
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New("invalid credential request")
		}
		if _, exists := fields[key]; exists {
			return errors.New("duplicate credential field")
		}
		fields[key] = value
	}
	if scanner.Err() != nil {
		return errors.New("invalid credential request")
	}
	get := func(name string) string { return environmentValue(env, name, "windows") }
	repo := get("SJSKILLS_GH_REPOSITORY")
	valid, err := githubRepository(repo)
	if err != nil || valid != repo || fields["protocol"] != "https" || fields["host"] != "github.com" || strings.TrimSuffix(fields["path"], ".git") != repo {
		return errors.New("credential request outside selected repository")
	}
	gh, config := get("SJSKILLS_GH_EXECUTABLE"), get("SJSKILLS_GH_CONFIG_DIR")
	if !filepath.IsAbs(gh) || !filepath.IsAbs(config) {
		return errors.New("invalid credential helper configuration")
	}
	if err := checkGHConfig(config); err != nil {
		return err
	}
	helperEnv := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "GIT_") || strings.HasPrefix(upper, "GH_") || strings.HasPrefix(upper, "SJSKILLS_GH_") {
			continue
		}
		helperEnv = append(helperEnv, entry)
	}
	if token := get("GH_TOKEN"); token != "" {
		helperEnv = append(helperEnv, "GH_TOKEN="+token)
	}
	helperEnv = append(helperEnv, "GH_CONFIG_DIR="+config, "GH_HOST=github.com", "GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "GH_NO_EXTENSION_UPDATE_NOTIFIER=1", "GH_TELEMETRY_DISABLED=1")
	helperCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(helperCtx, gh, "auth", "git-credential", "get")
	cmd.Env = helperEnv
	cmd.Stdin = strings.NewReader("protocol=https\nhost=github.com\npath=" + repo + ".git\n\n")
	cmd.Stdout = &credentialPipeWriter{writer: output, remaining: 64 * 1024}
	cmd.Stderr = io.Discard
	cmd.WaitDelay = boundedExecWaitDelay
	// Stay in the enclosing Git process group/job so outer cancellation and
	// cleanup verification cover gh as well as Git and the helper itself.
	if err := cmd.Run(); err != nil {
		return errors.New("gh credential lookup failed")
	}
	return nil
}

// Bound the private pipe without retaining a second copy of credential output.
type credentialPipeWriter struct {
	writer    io.Writer
	remaining int
}

func (w *credentialPipeWriter) Write(data []byte) (int, error) {
	if len(data) > w.remaining {
		return 0, errors.New("credential response exceeds bound")
	}
	n, err := w.writer.Write(data)
	w.remaining -= n
	return n, err
}
