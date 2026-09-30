package sjskills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	testCommitA = "1111111111111111111111111111111111111111"
	testCommitB = "2222222222222222222222222222222222222222"
)

func pktLine(value string) string { return fmt.Sprintf("%04x%s", len(value)+4, value) }

func refAdvertisement(commit string) string {
	return pktLine("# service=git-upload-pack\n") + "0000" +
		pktLine(commit+" HEAD\x00multi_ack symref=HEAD:refs/heads/main\n") +
		pktLine("3333333333333333333333333333333333333333 refs/heads/dev\n") +
		pktLine(commit+" refs/heads/main\n") + "0000"
}

type registryServer struct {
	commit     atomic.Value
	registries map[string]string
	refs       atomic.Int32
	raw        atomic.Int32
	offline    atomic.Bool
	server     *httptest.Server
}

func newRegistryServer(t *testing.T, commit string, registries map[string]string) *registryServer {
	t.Helper()
	s := &registryServer{registries: registries}
	s.commit.Store(commit)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/refs" {
			s.refs.Add(1)
		} else {
			s.raw.Add(1)
		}
		if s.offline.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path == "/refs" {
			w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
			fmt.Fprint(w, refAdvertisement(s.commit.Load().(string)))
			return
		}
		commit, file, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/raw/"), "/")
		content, ok := s.registries[commit]
		if !ok || file != registryFile {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, content)
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *registryServer) source(t *testing.T, cacheRoot string, now *time.Time) RegistrySource {
	return RegistrySource{
		CacheRoot: cacheRoot,
		Now:       func() time.Time { return *now },
		Client:    s.server.Client(),
		refsURL:   s.server.URL + "/refs",
		rawBase:   s.server.URL + "/raw/",
	}
}

func digest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func TestParseRefAdvertisementFindsMain(t *testing.T) {
	commit, err := parseRefAdvertisement([]byte(refAdvertisement(testCommitA)), registryBranchRef)
	if err != nil || commit != testCommitA {
		t.Fatalf("commit=%q err=%v", commit, err)
	}
	for name, input := range map[string]string{
		"truncated":  refAdvertisement(testCommitA)[:30],
		"bad length": "zzzz" + testCommitA,
		"short hash": pktLine("abc refs/heads/main\n"),
		"upper hash": pktLine(strings.ToUpper("abcdef1234abcdef1234abcdef1234abcdef1234") + " refs/heads/main\n"),
		"missing":    pktLine(testCommitA+" refs/heads/dev\n") + "0000",
	} {
		if _, err := parseRefAdvertisement([]byte(input), registryBranchRef); err == nil {
			t.Fatalf("%s: accepted", name)
		}
	}
}

func TestRegistrySourceResolveFetchesAtResolvedCommitAndCaches(t *testing.T) {
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON)})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)

	published, err := source.Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if published.Commit != testCommitA || published.SHA256 != digest(string(registryV4JSON)) || !published.ResolvedAt.Equal(now) || published.Stale {
		t.Fatalf("published = %+v", published)
	}
	if published.Registry.Version != RegistryVersion {
		t.Fatalf("registry version = %d", published.Registry.Version)
	}
	// The commit cache avoids refetching content; the branch is always resolved.
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	if server.refs.Load() != 2 || server.raw.Load() != 1 {
		t.Fatalf("refs=%d raw=%d", server.refs.Load(), server.raw.Load())
	}
}

func TestRegistrySourceResolveFollowsBranchMoves(t *testing.T) {
	changed := strings.Replace(string(registryV4JSON), `"Version 4 desired-state registry for sjskills."`, `"Changed on main."`, 1)
	if changed == string(registryV4JSON) {
		t.Fatal("fixture description not found")
	}
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON), testCommitB: changed})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	server.commit.Store(testCommitB)
	published, err := source.Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if published.Commit != testCommitB || published.Registry.Description != "Changed on main." {
		t.Fatalf("published = %s %q", published.Commit, published.Registry.Description)
	}
	pinned, err := source.At(context.Background(), testCommitA)
	if err != nil || pinned.Registry.Description == "Changed on main." || !pinned.ResolvedAt.IsZero() {
		t.Fatalf("pinned = %+v err=%v", pinned, err)
	}
}

func TestRegistrySourceForStatusUsesCacheWithinInterval(t *testing.T) {
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON)})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	server.offline.Store(true)
	now = now.Add(statusRefreshInterval - time.Minute)
	published, err := source.ForStatus(context.Background())
	if err != nil || published.Stale || published.Commit != testCommitA {
		t.Fatalf("published = %+v err=%v", published, err)
	}
	if server.refs.Load() != 1 {
		t.Fatalf("status refreshed within interval: refs=%d", server.refs.Load())
	}
}

func TestRegistrySourceStaleFallbackAndCooldown(t *testing.T) {
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON)})
	resolvedAt := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	now := resolvedAt
	source := server.source(t, t.TempDir(), &now)
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	server.offline.Store(true)
	now = now.Add(statusRefreshInterval + time.Hour)

	published, err := source.ForStatus(context.Background())
	if err != nil || !published.Stale || published.Commit != testCommitA || !published.ResolvedAt.Equal(resolvedAt) || !strings.Contains(published.StaleReason, "HTTP 503") {
		t.Fatalf("published = %+v err=%v", published, err)
	}
	attempts := server.refs.Load()
	now = now.Add(statusRetryInterval / 2)
	published, err = source.ForStatus(context.Background())
	if err != nil || !published.Stale || !strings.Contains(published.StaleReason, "cooldown") {
		t.Fatalf("cooldown published = %+v err=%v", published, err)
	}
	if server.refs.Load() != attempts {
		t.Fatal("status refreshed during cooldown")
	}
	// Selection commands are explicit and always attempt a refresh.
	if published, err = source.ForSelection(context.Background()); err != nil || !published.Stale {
		t.Fatalf("selection published = %+v err=%v", published, err)
	}
	if server.refs.Load() != attempts+1 {
		t.Fatal("selection did not attempt a refresh")
	}

	server.offline.Store(false)
	now = now.Add(statusRetryInterval)
	if published, err = source.ForStatus(context.Background()); err != nil || published.Stale || !published.ResolvedAt.Equal(now) {
		t.Fatalf("recovered published = %+v err=%v", published, err)
	}
}

func TestRegistrySourceUnavailableWithoutCache(t *testing.T) {
	server := newRegistryServer(t, testCommitA, nil)
	server.offline.Store(true)
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)
	for name, load := range map[string]func(context.Context) (PublishedRegistry, error){
		"status": source.ForStatus, "selection": source.ForSelection, "resolve": source.Resolve,
	} {
		if _, err := load(context.Background()); err == nil {
			t.Fatalf("%s: loaded without network or cache", name)
		}
	}
}

func TestRegistrySourceUnsupportedVersionFailsClosed(t *testing.T) {
	future := strings.Replace(string(registryV4JSON), `"version": 4`, `"version": 5`, 1)
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON), testCommitB: future})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	server.commit.Store(testCommitB)
	now = now.Add(statusRefreshInterval + time.Hour)
	for name, load := range map[string]func(context.Context) (PublishedRegistry, error){
		"status": source.ForStatus, "selection": source.ForSelection, "resolve": source.Resolve,
	} {
		_, err := load(context.Background())
		var issue *Issue
		if !errors.As(err, &issue) || issue.Code != IssueRegistryVersion || !strings.Contains(issue.Message, "version 5") || !strings.Contains(issue.Message, "install the sjskills release") {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}

func TestRegistrySourceRejectsInvalidInputs(t *testing.T) {
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: `{"version": 4}`})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	source := server.source(t, t.TempDir(), &now)
	var validation *ValidationErrors
	if _, err := source.Resolve(context.Background()); !errors.As(err, &validation) {
		t.Fatalf("invalid registry err = %v", err)
	}
	if _, err := source.At(context.Background(), strings.ToUpper(testCommitA)); err == nil {
		t.Fatal("accepted non-canonical commit")
	}

	html := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, refAdvertisement(testCommitA))
	}))
	t.Cleanup(html.Close)
	source.refsURL = html.URL
	if _, err := source.Resolve(context.Background()); err == nil || !strings.Contains(err.Error(), "content type") {
		t.Fatalf("content type err = %v", err)
	}
}

func TestRegistrySourceIgnoresTamperedCommitCache(t *testing.T) {
	server := newRegistryServer(t, testCommitA, map[string]string{testCommitA: string(registryV4JSON)})
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	cacheRoot := t.TempDir()
	source := server.source(t, cacheRoot, &now)
	if _, err := source.Resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	directory, err := source.cacheDirectory()
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(registryV4JSON), `"Version 4 desired-state registry for sjskills."`, `"Tampered."`, 1)
	if err := writeRegistryCacheFile(directory, testCommitA+".json", []byte(tampered)); err != nil {
		t.Fatal(err)
	}
	server.offline.Store(true)
	// A digest mismatch makes the cached registry unusable rather than trusted.
	if _, err := source.ForStatus(context.Background()); err == nil {
		t.Fatal("status used a tampered cached registry")
	}
}
