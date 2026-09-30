package sjskills

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// The binary embeds where the registry is published, never its contents.
// Reconciliation consumes only published content, so there is no local override.
const (
	registryRepository = releaseRepository
	registryBranchRef  = "refs/heads/main"
	registryFile       = "skill-registry.json"
	registryRefsURL    = "https://github.com/" + registryRepository + ".git/info/refs?service=git-upload-pack"
	registryRawBase    = "https://raw.githubusercontent.com/" + registryRepository + "/"

	maxRegistryBytes       = 1 << 20
	maxRefAdvertisement    = 4 << 20
	registryCacheSchema    = 1
	registryLatestFile     = "latest.json"
	registryLatestLockFile = "latest.lock"
)

var registryCommitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// PublishedRegistry is a validated registry together with the commit that
// supplied it and the SHA-256 of its exact published bytes. ResolvedAt is when
// the branch was last observed at Commit; it is zero for a registry loaded at an
// explicitly reviewed commit. Stale marks a cached registry used because a
// refresh failed or is cooling down; StaleReason says why.
type PublishedRegistry struct {
	Registry    Registry
	Commit      string
	SHA256      string
	ResolvedAt  time.Time
	Stale       bool
	StaleReason string
}

// RegistrySource resolves the published branch, fetches the registry at the
// resolved commit, and caches verified registries by commit. Resolution uses
// Git's anonymous smart-HTTP ref advertisement, which needs no credentials,
// Git executable, or GitHub API quota.
type RegistrySource struct {
	CacheRoot string
	Now       func() time.Time
	Client    *http.Client

	// Tests replace the published endpoints; production uses the fixed location.
	refsURL string
	rawBase string
}

type registryLatest struct {
	Schema     int       `json:"schema"`
	Source     string    `json:"source"`
	Ref        string    `json:"ref"`
	Commit     string    `json:"commit"`
	SHA256     string    `json:"sha256"`
	ResolvedAt time.Time `json:"resolvedAt"`
	RetryAt    time.Time `json:"retryAt"`
}

func (e registryLatest) resolved() bool { return e.Commit != "" }

func (s RegistrySource) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s RegistrySource) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: StatusRefreshBudget, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("registry redirect refused") }}
}

// Resolve observes the published branch now and never falls back to a cached
// registry. Plan and apply use it because they need current desired state.
func (s RegistrySource) Resolve(ctx context.Context) (PublishedRegistry, error) {
	directory, cacheErr := s.cacheDirectory()
	commit, err := s.resolveCommit(ctx)
	if err != nil {
		return PublishedRegistry{}, err
	}
	published, err := s.load(ctx, directory, cacheErr, commit)
	if err != nil {
		return PublishedRegistry{}, err
	}
	published.ResolvedAt = s.now()
	if cacheErr == nil {
		// The cache only accelerates later status; a write failure never
		// invalidates evidence that was just verified.
		_ = s.recordLatest(directory, registryLatest{Commit: commit, SHA256: published.SHA256, ResolvedAt: published.ResolvedAt})
	}
	return published, nil
}

// At loads the registry at an explicitly reviewed commit without consulting
// the branch, so a push after review cannot change the reviewed desired state.
func (s RegistrySource) At(ctx context.Context, commit string) (PublishedRegistry, error) {
	if !registryCommitPattern.MatchString(commit) {
		return PublishedRegistry{}, &Issue{Code: IssueMalformedInput, Path: "registry.commit", Message: "registry commit must be 40 lowercase hexadecimal characters"}
	}
	directory, cacheErr := s.cacheDirectory()
	return s.load(ctx, directory, cacheErr, commit)
}

// ForStatus reuses a registry resolved within the status refresh interval and
// otherwise refreshes, honoring the status retry cooldown. When refreshing is
// not possible it returns the last verified registry marked stale.
func (s RegistrySource) ForStatus(ctx context.Context) (PublishedRegistry, error) {
	return s.recent(ctx, statusRefreshInterval, true)
}

// ForSelection always attempts a refresh for commands that list or validate
// profiles, falling back to the last verified registry marked stale.
func (s RegistrySource) ForSelection(ctx context.Context) (PublishedRegistry, error) {
	return s.recent(ctx, 0, false)
}

func (s RegistrySource) recent(ctx context.Context, maxAge time.Duration, cooldown bool) (PublishedRegistry, error) {
	directory, err := s.cacheDirectory()
	if err != nil {
		return s.Resolve(ctx)
	}
	latest, _ := s.readLatest(directory)
	now := s.now()
	if latest.resolved() && maxAge > 0 && !latest.ResolvedAt.After(now) && now.Sub(latest.ResolvedAt) < maxAge {
		if cached, err := s.cachedAt(directory, latest); err == nil {
			return cached, nil
		}
	}
	if cooldown && !(statusCacheEntry{RetryAt: latest.RetryAt}).retryDue(now) {
		return s.staleOrError(directory, latest, errors.New("registry refresh cooldown active"))
	}
	published, resolveErr := s.Resolve(ctx)
	if resolveErr == nil {
		return published, nil
	}
	var issue *Issue
	var validation *ValidationErrors
	if errors.As(resolveErr, &issue) || errors.As(resolveErr, &validation) {
		// An unsupported or invalid published registry fails closed; an older
		// cached registry must not mask that the binary or catalog needs work.
		return PublishedRegistry{}, resolveErr
	}
	latest.RetryAt = now.Add(statusRetryInterval)
	_ = s.recordLatest(directory, latest)
	return s.staleOrError(directory, latest, resolveErr)
}

func (s RegistrySource) staleOrError(directory string, latest registryLatest, cause error) (PublishedRegistry, error) {
	if latest.resolved() {
		if cached, err := s.cachedAt(directory, latest); err == nil {
			cached.Stale = true
			cached.StaleReason = cause.Error()
			return cached, nil
		}
	}
	return PublishedRegistry{}, fmt.Errorf("skill registry unavailable: %w", cause)
}

func (s RegistrySource) cachedAt(directory string, latest registryLatest) (PublishedRegistry, error) {
	data, err := readStatusFile(filepath.Join(directory, latest.Commit+".json"))
	if err != nil {
		return PublishedRegistry{}, err
	}
	published, err := parsePublishedRegistry(latest.Commit, data)
	if err != nil {
		return PublishedRegistry{}, err
	}
	if published.SHA256 != latest.SHA256 {
		return PublishedRegistry{}, errors.New("cached registry digest mismatch")
	}
	published.ResolvedAt = latest.ResolvedAt
	return published, nil
}

func (s RegistrySource) load(ctx context.Context, directory string, cacheErr error, commit string) (PublishedRegistry, error) {
	path := filepath.Join(directory, commit+".json")
	if cacheErr == nil {
		if data, err := readStatusFile(path); err == nil {
			if published, err := parsePublishedRegistry(commit, data); err == nil {
				return published, nil
			}
		}
	}
	data, err := s.fetch(ctx, commit)
	if err != nil {
		return PublishedRegistry{}, err
	}
	published, err := parsePublishedRegistry(commit, data)
	if err != nil {
		return PublishedRegistry{}, err
	}
	if cacheErr == nil {
		_ = s.writeCommit(directory, commit, data)
	}
	return published, nil
}

func parsePublishedRegistry(commit string, data []byte) (PublishedRegistry, error) {
	if err := checkRegistryVersion(data); err != nil {
		return PublishedRegistry{}, err
	}
	registry, err := ParseRegistry(data)
	if err != nil {
		return PublishedRegistry{}, err
	}
	sum := sha256.Sum256(data)
	return PublishedRegistry{Registry: registry, Commit: commit, SHA256: hex.EncodeToString(sum[:])}, nil
}

// checkRegistryVersion names the required action for a schema this binary
// cannot interpret, before strict decoding reports unrelated unknown fields.
func checkRegistryVersion(data []byte) error {
	var header struct {
		Version *int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil || header.Version == nil || *header.Version == RegistryVersion {
		return nil
	}
	return &Issue{Code: IssueRegistryVersion, Path: "registry.version", Message: fmt.Sprintf(
		"published registry version %d is not supported by sjskills %s, which reads version %d; install the sjskills release that supports it",
		*header.Version, ToolVersion, RegistryVersion)}
}

func (s RegistrySource) resolveCommit(ctx context.Context) (string, error) {
	url := s.refsURL
	if url == "" {
		url = registryRefsURL
	}
	data, contentType, err := s.get(ctx, url, maxRefAdvertisement, "registry branch resolution")
	if err != nil {
		return "", err
	}
	if contentType != "application/x-git-upload-pack-advertisement" {
		return "", errors.New("registry branch resolution returned an unexpected content type")
	}
	return parseRefAdvertisement(data, registryBranchRef)
}

func (s RegistrySource) fetch(ctx context.Context, commit string) ([]byte, error) {
	base := s.rawBase
	if base == "" {
		base = registryRawBase
	}
	data, _, err := s.get(ctx, base+commit+"/"+registryFile, maxRegistryBytes, "registry fetch")
	return data, err
}

func (s RegistrySource) get(ctx context.Context, url string, limit int64, operation string) ([]byte, string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("User-Agent", "sjskills-registry")
	response, err := s.client().Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("%s failed (network or deadline)", operation)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%s returned HTTP %d", operation, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, "", fmt.Errorf("%s response unreadable or oversized", operation)
	}
	return data, strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]), nil
}

// parseRefAdvertisement reads Git's pkt-line ref advertisement and returns the
// commit for ref. Any framing error rejects the whole response.
func parseRefAdvertisement(data []byte, ref string) (string, error) {
	invalid := errors.New("registry branch resolution returned a malformed ref advertisement")
	for len(data) > 0 {
		if len(data) < 4 {
			return "", invalid
		}
		length, err := strconv.ParseUint(string(data[:4]), 16, 16)
		if err != nil {
			return "", invalid
		}
		if length == 0 {
			data = data[4:]
			continue
		}
		if length < 4 || int(length) > len(data) {
			return "", invalid
		}
		line := bytes.TrimSuffix(data[4:length], []byte("\n"))
		data = data[length:]
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		line, _, _ = bytes.Cut(line, []byte{0})
		hash, name, ok := strings.Cut(string(line), " ")
		if !ok || name != ref {
			continue
		}
		if !registryCommitPattern.MatchString(hash) {
			return "", invalid
		}
		return hash, nil
	}
	return "", fmt.Errorf("registry branch %s is not advertised", ref)
}

func (s RegistrySource) cacheDirectory() (string, error) {
	return disposableCacheDirectory(s.CacheRoot, "registry")
}

func (s RegistrySource) readLatest(directory string) (registryLatest, error) {
	data, err := readStatusFile(filepath.Join(directory, registryLatestFile))
	if err != nil {
		return registryLatest{}, err
	}
	var entry registryLatest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entry); err != nil {
		return registryLatest{}, err
	}
	if decoder.Decode(new(any)) != io.EOF || entry.Schema != registryCacheSchema || entry.Source != registryRepository || entry.Ref != registryBranchRef {
		return registryLatest{}, errors.New("incompatible registry cache")
	}
	resolved := registryCommitPattern.MatchString(entry.Commit) && lowercaseDigestPattern.MatchString(entry.SHA256) && !entry.ResolvedAt.IsZero()
	unresolved := entry.Commit == "" && entry.SHA256 == "" && entry.ResolvedAt.IsZero()
	if !resolved && !unresolved {
		return registryLatest{}, errors.New("invalid registry cache")
	}
	return entry, nil
}

// recordLatest publishes the branch observation under the namespace lock and
// keeps the newest successful observation when refreshes race.
func (s RegistrySource) recordLatest(directory string, entry registryLatest) error {
	entry.Schema, entry.Source, entry.Ref = registryCacheSchema, registryRepository, registryBranchRef
	unlock, err := lockStatusPath(filepath.Join(directory, registryLatestLockFile))
	if err != nil {
		return err
	}
	defer unlock()
	if current, err := s.readLatest(directory); err == nil && current.ResolvedAt.After(entry.ResolvedAt) && !current.ResolvedAt.After(s.now()) {
		if entry.resolved() {
			return nil
		}
		// A failed refresh keeps the newer success and records only its cooldown.
		current.RetryAt = entry.RetryAt
		entry = current
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := writeRegistryCacheFile(directory, registryLatestFile, data); err != nil {
		return err
	}
	s.prune(directory, entry.Commit)
	return nil
}

func (s RegistrySource) writeCommit(directory, commit string, data []byte) error {
	unlock, err := lockStatusPath(filepath.Join(directory, registryLatestLockFile))
	if err != nil {
		return err
	}
	defer unlock()
	return writeRegistryCacheFile(directory, commit+".json", data)
}

func writeRegistryCacheFile(directory, name string, data []byte) error {
	path := filepath.Join(directory, name)
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("unsafe registry cache destination")
	}
	file, err := os.CreateTemp(directory, name+".tmp-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

// prune runs under the namespace lock. It bounds directory work and removes
// expired commit registries and abandoned temporary files, keeping the latest.
func (s RegistrySource) prune(directory, keep string) {
	dir, err := os.Open(directory)
	if err != nil {
		return
	}
	defer dir.Close()
	entries, _ := dir.ReadDir(256)
	for _, entry := range entries {
		name := entry.Name()
		commit, isRegistry := strings.CutSuffix(name, ".json")
		isRegistry = isRegistry && registryCommitPattern.MatchString(commit)
		if (!isRegistry || commit == keep) && !strings.Contains(name, ".tmp-") {
			continue
		}
		path := filepath.Join(directory, name)
		info, err := os.Lstat(path)
		if err == nil && info.Mode().IsRegular() && s.now().Sub(info.ModTime()) >= statusRetention {
			_ = os.Remove(path)
		}
	}
}
