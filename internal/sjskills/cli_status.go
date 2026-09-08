package sjskills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const releaseRepository = "sjunepark/agent-scripts"
const releaseAPI = "https://api.github.com/repos/" + releaseRepository + "/releases"
const releasePage = "https://github.com/" + releaseRepository + "/releases/tag/sjskills-v"

var stableVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

type CLIComparison string

const (
	CLIUpdate                  CLIComparison = "update"
	CLIEqual                   CLIComparison = "equal"
	CLIAhead                   CLIComparison = "ahead"
	CLINoRelease               CLIComparison = "no-release"
	CLIUncomparable            CLIComparison = "uncomparable"
	CLIDistributionUnavailable CLIComparison = "distribution-unavailable"
	CLIUnknown                 CLIComparison = "unknown"
)

// CLIAdvisory compares embedded versions, never source commits or binary integrity.
// It is disposable evidence and grants no reconciliation or installation authority.
type CLIAdvisory struct {
	RunningVersion   string            `json:"runningVersion"`
	AvailableVersion string            `json:"availableVersion,omitempty"`
	Comparison       CLIComparison     `json:"comparison"`
	Freshness        AdvisoryFreshness `json:"freshness"`
	ObservedAt       *time.Time        `json:"observedAt,omitempty"`
	Cached           bool              `json:"cached"`
	ReleaseURL       string            `json:"releaseURL,omitempty"`
	Error            string            `json:"error,omitempty"`
}

func validateCLIAdvisory(a *CLIAdvisory) error {
	if a == nil {
		return nil
	}
	invalid := errors.New("invalid CLI advisory")
	if a.RunningVersion == "" || len(a.RunningVersion) > 256 || len(a.Error) > 1024 {
		return invalid
	}
	switch a.Freshness {
	case AdvisoryUnavailable:
		if a.ObservedAt != nil || a.Cached || a.Comparison != CLIUnknown || a.AvailableVersion != "" || a.ReleaseURL != "" || a.Error == "" {
			return invalid
		}
		return nil
	case AdvisoryFresh, AdvisoryStale:
		if a.ObservedAt == nil || a.ObservedAt.IsZero() {
			return invalid
		}
		if a.Freshness == AdvisoryStale && (!a.Cached || a.Error == "") {
			return invalid
		}
	default:
		return invalid
	}
	if a.AvailableVersion != "" && (!validCLIVersion(a.AvailableVersion) || a.ReleaseURL != releasePage+a.AvailableVersion) {
		return invalid
	}
	switch a.Comparison {
	case CLINoRelease:
		if a.AvailableVersion != "" || a.ReleaseURL != "" {
			return invalid
		}
	case CLIUncomparable:
		if validCLIVersion(a.RunningVersion) || a.AvailableVersion == "" {
			return invalid
		}
	case CLIDistributionUnavailable:
		if a.AvailableVersion == "" || a.Error == "" {
			return invalid
		}
	case CLIUpdate, CLIEqual, CLIAhead:
		if !validCLIVersion(a.RunningVersion) || a.AvailableVersion == "" || compareCLI(a.RunningVersion, a.AvailableVersion) != a.Comparison {
			return invalid
		}
	default:
		return invalid
	}
	return nil
}
func validCLIVersion(v string) bool { return len(v) <= 128 && stableVersion.MatchString(v) }
func compareCLI(running, available string) CLIComparison {
	a, b := strings.Split(running, "."), strings.Split(available, ".")
	// Compare decimal components without integer overflow, matching packaging's grammar.
	for i := range a {
		if len(a[i]) < len(b[i]) || len(a[i]) == len(b[i]) && a[i] < b[i] {
			return CLIUpdate
		}
		if len(a[i]) > len(b[i]) || len(a[i]) == len(b[i]) && a[i] > b[i] {
			return CLIAhead
		}
	}
	return CLIEqual
}

type CLIStatusService struct {
	CacheRoot string
	Now       func() time.Time
	Client    *http.Client
	// Platform defaults to this executable's target; tests can exercise packaged targets.
	Platform string
}
type cliReleaseCache struct {
	Schema      int       `json:"schema"`
	Source      string    `json:"source"`
	Platform    string    `json:"platform"`
	Version     string    `json:"version"`
	Installable bool      `json:"installable"`
	ObservedAt  time.Time `json:"observedAt"`
	RetryAt     time.Time `json:"retryAt"`
}

func (s CLIStatusService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
func (s CLIStatusService) platform() string {
	if s.Platform != "" {
		return s.Platform
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}
func (e cliReleaseCache) fresh(now time.Time) bool {
	return (statusCacheEntry{ObservedAt: e.ObservedAt}).fresh(now)
}
func (e cliReleaseCache) retryDue(now time.Time) bool {
	return (statusCacheEntry{RetryAt: e.RetryAt}).retryDue(now)
}
func (s CLIStatusService) read(path string) (cliReleaseCache, error) {
	var e cliReleaseCache
	data, err := readStatusFile(path)
	if err != nil {
		return e, err
	}
	// Empty version is a successful absence observation only when explicitly
	// recorded. Missing/null fields are corruption, not zero-valued evidence.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return e, err
	}
	for _, name := range []string{"schema", "source", "platform", "version", "installable", "observedAt", "retryAt"} {
		value, ok := fields[name]
		if !ok || strings.TrimSpace(string(value)) == "null" {
			return e, errors.New("incomplete release cache")
		}
	}
	d := json.NewDecoder(strings.NewReader(string(data)))
	d.DisallowUnknownFields()
	if err = d.Decode(&e); err != nil {
		return e, err
	}
	if d.Decode(new(any)) != io.EOF || e.Schema != 1 || e.Source != releaseRepository || e.Platform != s.platform() || (e.Version != "" && !validCLIVersion(e.Version)) || (e.Version == "" && e.Installable) || (e.ObservedAt.IsZero() && e.Version != "") {
		return cliReleaseCache{}, errors.New("invalid release cache")
	}
	return e, nil
}
func writeCLIReleaseCache(path string, e cliReleaseCache) error {
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("unsafe release cache")
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), "release.tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func (s CLIStatusService) Check(ctx context.Context) CLIAdvisory {
	a := CLIAdvisory{RunningVersion: ToolVersion, Comparison: CLIUnknown, Freshness: AdvisoryUnavailable}
	directory, err := disposableCacheDirectory(s.CacheRoot, "cli-status")
	if err != nil {
		a.Error = "CLI release cache unavailable"
		return a
	}
	// A single fixed entry bounds retention and is isolated from skill-cache pruning.
	path := filepath.Join(directory, "release.json")
	entry, err := s.read(path)
	if err != nil {
		entry = cliReleaseCache{Schema: 1, Source: releaseRepository, Platform: s.platform()}
	}
	a.Cached = true
	if !entry.fresh(s.now()) {
		if !entry.retryDue(s.now()) {
			a.Error = "CLI release refresh cooldown active"
		} else {
			unlock, err := lockStatusPath(filepath.Join(directory, "release.lock"))
			if err != nil {
				a.Error = "CLI release cache unavailable or another check active"
			} else {
				defer unlock()
				s.pruneTemporaryFiles(directory)
				if latest, err := s.read(path); err == nil {
					entry = latest
				}
				if !entry.fresh(s.now()) {
					if !entry.retryDue(s.now()) {
						a.Error = "CLI release refresh cooldown active"
					} else {
						version, installable, err := s.lookup(ctx)
						if err != nil {
							a.Error = err.Error()
							entry.RetryAt = s.now().Add(statusRetryInterval)
						} else {
							entry = cliReleaseCache{Schema: 1, Source: releaseRepository, Platform: s.platform(), Version: version, Installable: installable, ObservedAt: s.now()}
							a.Cached = false
						}
						if err := writeCLIReleaseCache(path, entry); err != nil {
							if a.Error != "" {
								a.Error += "; "
							}
							a.Error += "CLI release cache could not be written"
						}
					}
				}
			}
		}
	}
	if entry.ObservedAt.IsZero() {
		a.Cached = false
		if a.Error == "" {
			a.Error = "CLI release evidence unavailable"
		}
		return a
	}
	a.ObservedAt = &entry.ObservedAt
	a.Freshness = AdvisoryFresh
	if !entry.fresh(s.now()) {
		a.Freshness = AdvisoryStale
		a.Cached = true
		if a.Error == "" {
			a.Error = "CLI release evidence expired"
		}
	}
	a.AvailableVersion = entry.Version
	if entry.Version == "" {
		a.Comparison = CLINoRelease
		return a
	}
	a.ReleaseURL = releasePage + entry.Version
	switch {
	case !entry.Installable:
		a.Comparison = CLIDistributionUnavailable
		if a.Error != "" {
			a.Error += "; "
		}
		a.Error += "release lacks required assets for this platform"
	case !validCLIVersion(ToolVersion):
		a.Comparison = CLIUncomparable
	default:
		a.Comparison = compareCLI(ToolVersion, entry.Version)
	}
	return a
}

type githubRelease struct {
	Tag         *string    `json:"tag_name"`
	Draft       *bool      `json:"draft"`
	Prerelease  *bool      `json:"prerelease"`
	PublishedAt *time.Time `json:"published_at"`
	Assets      []struct {
		Name  string `json:"name"`
		State string `json:"state"`
		Size  int64  `json:"size"`
		URL   string `json:"browser_download_url"`
	} `json:"assets"`
}

func (s CLIStatusService) lookup(ctx context.Context) (string, bool, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: StatusRefreshBudget, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("release redirect refused") }}
	}
	best := ""
	installable := false
	for page := 1; page <= 10; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?per_page=100&page=%d", releaseAPI, page), nil)
		if err != nil {
			return "", false, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("User-Agent", "sjskills-version-status")
		response, err := client.Do(req)
		if err != nil {
			return "", false, errors.New("CLI release lookup failed (network or deadline)")
		}
		data, readErr := io.ReadAll(io.LimitReader(response.Body, 4<<20+1))
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return "", false, fmt.Errorf("CLI release lookup returned HTTP %d", response.StatusCode)
		}
		if readErr != nil || len(data) > 4<<20 {
			return "", false, errors.New("CLI release response unreadable or oversized")
		}
		var releases []githubRelease
		if err := json.Unmarshal(data, &releases); err != nil || releases == nil || len(releases) > 100 {
			return "", false, errors.New("invalid CLI release response")
		}
		for _, r := range releases {
			if r.Tag == nil || r.Draft == nil || r.Prerelease == nil {
				return "", false, errors.New("incomplete CLI release response")
			}
			v := strings.TrimPrefix(*r.Tag, "sjskills-v")
			if *r.Draft || *r.Prerelease || !strings.HasPrefix(*r.Tag, "sjskills-v") || !validCLIVersion(v) {
				continue
			}
			if r.PublishedAt == nil || r.PublishedAt.IsZero() || r.PublishedAt.After(s.now()) {
				return "", false, errors.New("invalid CLI release publication")
			}
			if best == "" || compareCLI(best, v) == CLIUpdate {
				best = v
				installable = s.installable(r, v)
			}
		}
		if len(releases) < 100 && !strings.Contains(response.Header.Get("Link"), `rel="next"`) {
			return best, installable, nil
		}
	}
	return "", false, errors.New("CLI release pagination limit exceeded")
}
func (s CLIStatusService) installable(r githubRelease, v string) bool {
	var archive, installer string
	switch s.platform() {
	case "darwin/arm64":
		archive = "darwin_arm64.tar.gz"
		installer = "install.sh"
	case "darwin/amd64":
		archive = "darwin_amd64.tar.gz"
		installer = "install.sh"
	case "windows/amd64":
		archive = "windows_amd64.zip"
		installer = "install.ps1"
	default:
		return false
	}
	needed := map[string]bool{"sjskills_" + v + "_" + archive: false, "SHA256SUMS": false, installer: false}
	for _, a := range r.Assets {
		if _, ok := needed[a.Name]; ok && a.State == "uploaded" && a.Size > 0 && a.URL == "https://github.com/"+releaseRepository+"/releases/download/sjskills-v"+v+"/"+a.Name {
			needed[a.Name] = true
		}
	}
	for _, ok := range needed {
		if !ok {
			return false
		}
	}
	return true
}

// Called while holding the release lock; abandoned temporary files are the only
// accumulating entries in this fixed, separate namespace.
func (s CLIStatusService) pruneTemporaryFiles(directory string) {
	dir, err := os.Open(directory)
	if err != nil {
		return
	}
	defer dir.Close()
	entries, _ := dir.ReadDir(256)
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "release.tmp-") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		info, err := os.Lstat(path)
		if err == nil && info.Mode().IsRegular() && s.now().Sub(info.ModTime()) >= statusRetention {
			_ = os.Remove(path)
		}
	}
}
