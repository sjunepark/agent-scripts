package sjskills

// Local direct sources let a project install a skill it develops in-repo.
// They are deliberately confined to project manifests: the global baseline is
// machine-independent, and registry sources stay remote. A local source is
// copied into the same verified staging snapshot that a remote fetch produces,
// so placement, provenance, quarantine, and rollback need no separate path.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

const localSourceIdentityPrefix = "local:"

// LocalSourceError is a configuration problem with one local direct source:
// the user fixes sjskills.toml or the source directory, and retrying cannot
// help. Its message names only the user's own manifest text, so it bypasses
// the credential redaction applied to fetched-source diagnostics.
type LocalSourceError struct {
	Skill   string
	Source  string
	Problem string
}

func (e *LocalSourceError) Error() string {
	return fmt.Sprintf("%s: local source %q %s", e.Skill, e.Source, e.Problem)
}

// managedDirectoryPairs are generated skill roots. A source inside one of them
// would be an installed copy rather than the skill's own source, in this or
// any other project or home.
var managedDirectoryPairs = [][2]string{{string(TargetAgents), ManagedSkillsDirectoryName}, {string(TargetClaude), ManagedSkillsDirectoryName}}

// IsLocalSource reports whether source is written as a filesystem path rather
// than a remote location. It is syntactic, so it classifies invalid local
// spellings (for example `~/x` or a backslash path) for a precise diagnostic.
func IsLocalSource(source string) bool {
	return strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") ||
		strings.Contains(source, `\`) || hasDriveLetter(source)
}

func hasDriveLetter(value string) bool {
	return len(value) >= 2 && ((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) && value[1] == ':'
}

// LocalSourceProblem validates the committed spelling of a local direct
// source. Paths use forward slashes on every platform so one manifest works
// everywhere; relative paths must be explicit (`./` or `../`).
func LocalSourceProblem(source string) string {
	switch {
	case source == "":
		return "source is empty"
	case strings.TrimSpace(source) != source || strings.IndexFunc(source, unicode.IsControl) >= 0:
		return "local path contains invalid characters"
	case strings.HasPrefix(source, "~"):
		return "home-relative paths are not supported; use a path relative to sjskills.toml"
	case strings.Contains(source, `\`):
		return "local paths must use forward slashes"
	}
	if hasDriveLetter(source) {
		if len(source) < 3 || source[2] != '/' {
			return "drive paths must be absolute, such as C:/path"
		}
	} else if !strings.HasPrefix(source, "/") && !strings.HasPrefix(source, "./") && !strings.HasPrefix(source, "../") {
		return "relative local paths must start with ./ or ../"
	}
	canonical, ok := canonicalLocalSource(source)
	if !ok || canonical == "./" || strings.Trim(strings.ReplaceAll(canonical, "..", ""), "/") == "" {
		return "local path must name a skill directory, not the project root or its parent"
	}
	return ""
}

// canonicalLocalSource normalizes a validated spelling into the stable form
// used in provenance. Relative sources keep their manifest-relative form, so
// moving or recloning the project does not change ownership.
func canonicalLocalSource(source string) (string, bool) {
	if source == "" || strings.ContainsRune(source, 0) || strings.Contains(source, `\`) || strings.HasPrefix(source, "~") {
		return "", false
	}
	if hasDriveLetter(source) {
		if len(source) < 3 || source[2] != '/' {
			return "", false
		}
		return strings.ToUpper(source[:1]) + ":" + path.Clean(source[2:]), true
	}
	if strings.HasPrefix(source, "/") {
		return path.Clean(source), true
	}
	if !strings.HasPrefix(source, "./") && !strings.HasPrefix(source, "../") {
		return "", false
	}
	cleaned := path.Clean(source)
	if cleaned == "." {
		return "./", true
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return cleaned, true
	}
	return "./" + cleaned, true
}

func localSourceIdentity(source string) (string, bool) {
	if LocalSourceProblem(source) != "" {
		return "", false
	}
	canonical, ok := canonicalLocalSource(source)
	if !ok {
		return "", false
	}
	return localSourceIdentityPrefix + canonical, true
}

func isCanonicalLocalSourceIdentity(identity string) bool {
	rest, ok := strings.CutPrefix(identity, localSourceIdentityPrefix)
	if !ok {
		return false
	}
	canonical, valid := localSourceIdentity(rest)
	return valid && canonical == identity
}

// isPortableLocalSource reports whether a validated source is written
// relative to the project and stays inside it, so the committed manifest
// works in every checkout.
func isPortableLocalSource(source string) bool {
	canonical, ok := canonicalLocalSource(source)
	return ok && strings.HasPrefix(canonical, "./")
}

// resolveLocalSourcePath maps a validated source onto this machine. Relative
// paths resolve against the directory containing sjskills.toml, never the
// process working directory.
func resolveLocalSourcePath(projectRoot, source string) (string, error) {
	canonical, ok := canonicalLocalSource(source)
	if !ok {
		return "", errors.New("local source is invalid")
	}
	if strings.HasPrefix(canonical, "./") || strings.HasPrefix(canonical, "../") {
		if projectRoot == "" || !filepath.IsAbs(projectRoot) {
			return "", errors.New("relative local source requires the absolute project root")
		}
		resolved := filepath.Join(projectRoot, filepath.FromSlash(canonical))
		if pathWithin(resolved, projectRoot) {
			return "", errors.New("local source must not contain the project root")
		}
		return resolved, nil
	}
	resolved := filepath.Clean(filepath.FromSlash(canonical))
	if !filepath.IsAbs(resolved) {
		return "", errors.New("absolute local source is not absolute on this platform")
	}
	if projectRoot != "" && pathWithin(resolved, projectRoot) {
		return "", errors.New("local source must not contain the project root")
	}
	return resolved, nil
}

// inspectLocalSkillSource proves that a local source is a real skill
// directory that the copy-mode placement contract can represent, and returns
// its tree hash. Every failure names the manifest source so the user can fix
// or remove the declaration.
func inspectLocalSkillSource(skill DesiredSkill, limits MaterializerLimits) (TreeHash, error) {
	fail := func(message string, cause error) (TreeHash, error) {
		if cause != nil {
			message += ": " + cause.Error()
		}
		return TreeHash{}, &LocalSourceError{Skill: safeSkillName(skill.Name), Source: skill.Source, Problem: message}
	}
	if skill.LocalPath == "" || !filepath.IsAbs(skill.LocalPath) {
		return fail("has no resolved absolute path", nil)
	}
	info, err := os.Lstat(skill.LocalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("does not exist; restore it or remove its [[direct]] entry from sjskills.toml", nil)
		}
		return fail("cannot be inspected", nil)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fail("must be a real directory, not a file or symlink", nil)
	}
	canonical, err := filepath.EvalSymlinks(skill.LocalPath)
	if err != nil {
		return fail("cannot be resolved", nil)
	}
	if insideManagedDirectory(canonical) {
		return fail("is inside a generated .agents/skills, .claude/skills, or .sjskills directory; point it at the skill's own source", nil)
	}
	if _, err := os.Lstat(filepath.Join(canonical, ManifestFileName)); err == nil {
		return fail("is a project root containing sjskills.toml, not a skill directory", nil)
	}
	if err := validateLocalSkillTree(canonical, limits); err != nil {
		return fail("cannot be installed", err)
	}
	declared, err := localSkillDeclaredName(filepath.Join(skill.LocalPath, "SKILL.md"))
	if err != nil {
		return fail("is not a skill", err)
	}
	if declared != skill.Name {
		return fail(fmt.Sprintf("declares SKILL.md name %q, not %q", declared, skill.Name), nil)
	}
	digest, err := hashSkillTree(skill.LocalPath, limits)
	if err != nil {
		return fail("cannot be hashed", err)
	}
	return digest, nil
}

func insideManagedDirectory(canonical string) bool {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(canonical)), "/")
	for index, part := range parts {
		if strings.EqualFold(part, DerivedDirectoryName) {
			return true
		}
		if index == 0 {
			continue
		}
		for _, pair := range managedDirectoryPairs {
			if strings.EqualFold(parts[index-1], pair[0]) && strings.EqualFold(part, pair[1]) {
				return true
			}
		}
	}
	return false
}

// validateLocalSkillTree enforces the copy contract before any byte is copied:
// project placements contain only real directories and regular files.
func validateLocalSkillTree(root string, limits MaterializerLimits) error {
	limits = normalizedLimits(limits)
	entries := 0
	var walk func(string, int) error
	walk = func(directory string, depth int) error {
		if depth > limits.MaxTreeDepth {
			return errors.New("directory depth exceeded its bound")
		}
		children, err := os.ReadDir(directory)
		if err != nil {
			return errors.New("a directory is unreadable")
		}
		for _, child := range children {
			entries++
			if entries > limits.MaxTreeEntries {
				return errors.New("entry count exceeded its bound")
			}
			childPath := filepath.Join(directory, child.Name())
			info, err := os.Lstat(childPath)
			if err != nil {
				return errors.New("an entry changed during inspection")
			}
			relative, _ := filepath.Rel(root, childPath)
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				return fmt.Errorf("%s is a symlink; copy-mode skills contain only regular files and directories", filepath.ToSlash(relative))
			case info.IsDir():
				if insideManagedDirectory(childPath) {
					return fmt.Errorf("%s is a generated skills or reconciler directory", filepath.ToSlash(relative))
				}
				if err := walk(childPath, depth+1); err != nil {
					return err
				}
			case !info.Mode().IsRegular():
				return fmt.Errorf("%s is not a regular file", filepath.ToSlash(relative))
			}
		}
		return nil
	}
	return walk(root, 0)
}

// localSkillDeclaredName reads the `name` field from SKILL.md frontmatter.
// Only the leading `---` block is read, and only a top-level scalar `name`.
func localSkillDeclaredName(skillFile string) (string, error) {
	info, err := os.Lstat(skillFile)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("SKILL.md is missing")
	}
	file, err := os.Open(skillFile)
	if err != nil {
		return "", errors.New("SKILL.md is unreadable")
	}
	defer file.Close()
	scanner := bufio.NewScanner(io.LimitReader(file, 64*1024))
	if !scanner.Scan() || strings.TrimRightFunc(strings.TrimPrefix(scanner.Text(), "\uFEFF"), unicode.IsSpace) != "---" {
		return "", errors.New("SKILL.md has no frontmatter")
	}
	for scanner.Scan() {
		line := strings.TrimRightFunc(scanner.Text(), unicode.IsSpace)
		if line == "---" {
			break
		}
		key, value, found := strings.Cut(line, ":")
		if !found || key != "name" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		} else if before, _, found := strings.Cut(value, " #"); found {
			value = strings.TrimSpace(before)
		}
		if value == "" {
			break
		}
		return value, nil
	}
	return "", errors.New("SKILL.md frontmatter does not declare a name")
}

// stageLocalSkill copies a verified local source into the plan's staging root
// at the same path a Skills CLI install would use, then proves the copy equals
// the source as inspected. A source edited mid-copy is reported, not adopted.
func stageLocalSkill(root string, skill DesiredSkill, limits MaterializerLimits) (string, TreeHash, error) {
	sourceHash, err := inspectLocalSkillSource(skill, limits)
	if err != nil {
		return "", TreeHash{}, err
	}
	destination := filepath.Join(root, ".agents", "skills", skill.Name)
	if !pathWithin(root, destination) {
		return "", TreeHash{}, materializationError("staging", "skill path escapes staging root", nil)
	}
	if err := os.Mkdir(destination, 0o700); err != nil {
		return "", TreeHash{}, materializationError(safeSkillName(skill.Name), "local source staging directory could not be created", err)
	}
	remaining := normalizedLimits(limits).MaxTreeBytes
	if err := copyLocalSkillTree(skill.LocalPath, destination, &remaining); err != nil {
		return "", TreeHash{}, &LocalSourceError{Skill: safeSkillName(skill.Name), Source: skill.Source, Problem: "could not be copied: " + err.Error()}
	}
	staged, err := locateStagedSkill(root, skill.Name)
	if err != nil {
		return "", TreeHash{}, err
	}
	stagedHash, err := hashSkillTree(staged, limits)
	if err != nil {
		return "", TreeHash{}, err
	}
	if stagedHash != sourceHash {
		return "", TreeHash{}, &LocalSourceError{Skill: safeSkillName(skill.Name), Source: skill.Source, Problem: "changed while it was copied; retry"}
	}
	return staged, stagedHash, nil
}

// copyLocalSkillTree copies at most *remaining bytes; the caller's post-copy
// hash comparison rejects any other concurrent change to the source.
func copyLocalSkillTree(source, destination string, remaining *int64) error {
	info, err := os.Lstat(source)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("source directory changed during copy")
	}
	if err := os.Chmod(destination, info.Mode().Perm()|0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		entryInfo, err := os.Lstat(sourcePath)
		if err != nil {
			return errors.New("source changed during copy")
		}
		switch {
		case entryInfo.Mode()&os.ModeSymlink != 0:
			return errors.New("source gained a symlink during copy")
		case entryInfo.IsDir():
			if err := os.Mkdir(destinationPath, 0o700); err != nil {
				return err
			}
			if err := copyLocalSkillTree(sourcePath, destinationPath, remaining); err != nil {
				return err
			}
		case entryInfo.Mode().IsRegular():
			if err := copyLocalSkillFile(sourcePath, destinationPath, entryInfo, remaining); err != nil {
				return err
			}
		default:
			return errors.New("source gained a special file during copy")
		}
	}
	return nil
}

// copyLocalSkillFile preserves the executable bits the tree hash records; the
// owner read/write bits only keep the staging copy removable.
func copyLocalSkillFile(source, destination string, expected os.FileInfo, remaining *int64) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	opened, err := input.Stat()
	if err != nil || !os.SameFile(expected, opened) {
		return errors.New("source changed during copy")
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(output, io.LimitReader(input, *remaining+1))
	*remaining -= written
	if copyErr == nil && *remaining < 0 {
		copyErr = errors.New("source exceeded its size bound during copy")
	}
	chmodErr := output.Chmod(expected.Mode().Perm() | 0o600)
	closeErr := output.Close()
	return errors.Join(copyErr, chmodErr, closeErr)
}
