package sjskills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type releaseTransport func(*http.Request) (*http.Response, error)

func (f releaseTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func releaseFixture(version string) map[string]any {
	assets := []map[string]any{}
	for _, name := range []string{"sjskills_" + version + "_darwin_arm64.tar.gz", "SHA256SUMS", "install.sh"} {
		assets = append(assets, map[string]any{"name": name, "state": "uploaded", "size": 100, "browser_download_url": "https://github.com/" + releaseRepository + "/releases/download/sjskills-v" + version + "/" + name})
	}
	return map[string]any{"tag_name": "sjskills-v" + version, "draft": false, "prerelease": false, "published_at": "2026-01-01T00:00:00Z", "assets": assets}
}
func releaseTestService(t *testing.T, handler func(*http.Request) (int, string)) CLIStatusService {
	t.Helper()
	return CLIStatusService{CacheRoot: filepath.Join(t.TempDir(), "cache"), Platform: "darwin/arm64", Now: func() time.Time { return time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC) }, Client: &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(r.URL.String(), releaseAPI+"?") || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected request %s", r.URL)
		}
		status, body := handler(r)
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
}
func releaseJSON(values ...map[string]any) string {
	if values == nil {
		return "[]"
	}
	b, _ := json.Marshal(values)
	return string(b)
}

func TestCLIReleaseSelection(t *testing.T) {
	cases := []struct {
		name, body string
		want       CLIComparison
		version    string
	}{
		{"none", "[]", CLINoRelease, ""},
		{"equal", releaseJSON(releaseFixture(ToolVersion)), CLIEqual, ToolVersion},
		{"older", releaseJSON(releaseFixture("0.9.0")), CLIAhead, "0.9.0"},
		{"numeric order", releaseJSON(releaseFixture("1.9.0"), releaseFixture("1.10.0")), CLIUpdate, "1.10.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := releaseTestService(t, func(*http.Request) (int, string) { return 200, tc.body })
			a := s.Check(context.Background())
			if a.Comparison != tc.want || a.AvailableVersion != tc.version || a.Freshness != AdvisoryFresh || a.Cached || a.Error != "" {
				t.Fatalf("%+v", a)
			}
			if err := validateCLIAdvisory(&a); err != nil {
				t.Fatal(err, a)
			}
		})
	}
	for _, field := range []string{"draft", "prerelease"} {
		r := releaseFixture("99.0.0")
		r[field] = true
		s := releaseTestService(t, func(*http.Request) (int, string) { return 200, releaseJSON(r, releaseFixture("1.10.0")) })
		if a := s.Check(context.Background()); a.AvailableVersion != "1.10.0" {
			t.Fatal(a)
		}
	}
	for _, tag := range []string{"other-v99.0.0", "sjskills-v01.2.3", "sjskills-v2.0.0-rc1"} {
		r := releaseFixture("99.0.0")
		r["tag_name"] = tag
		s := releaseTestService(t, func(*http.Request) (int, string) { return 200, releaseJSON(r) })
		if a := s.Check(context.Background()); a.Comparison != CLINoRelease {
			t.Fatal(a)
		}
	}
}
func TestCLIReleasePaginationAndErrors(t *testing.T) {
	first := make([]map[string]any, 100)
	for i := range first {
		first[i] = releaseFixture("1.9.0")
	}
	calls := 0
	s := releaseTestService(t, func(r *http.Request) (int, string) {
		calls++
		if r.URL.Query().Get("page") == "1" {
			return 200, releaseJSON(first...)
		}
		return 200, releaseJSON(releaseFixture("1.10.0"))
	})
	if a := s.Check(context.Background()); a.AvailableVersion != "1.10.0" || calls != 2 {
		t.Fatal(a, calls)
	}
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"rate limit", 403, "[]"}, {"server", 500, "[]"}, {"null", 200, "null"}, {"malformed", 200, "{"}, {"missing fields", 200, `[{"tag_name":"sjskills-v2.0.0"}]`}, {"oversized", 200, strings.Repeat(" ", 4<<20+1)}, {"pagination limit", 200, releaseJSON(first...)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := releaseTestService(t, func(*http.Request) (int, string) { return tc.status, tc.body })
			a := s.Check(context.Background())
			if a.Freshness != AdvisoryUnavailable || a.Error == "" || a.Comparison != CLIUnknown {
				t.Fatal(a)
			}
			if err := validateCLIAdvisory(&a); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestCLIReleaseAssetsAndIdentity(t *testing.T) {
	r := releaseFixture("2.0.0")
	r["assets"] = []any{}
	s := releaseTestService(t, func(*http.Request) (int, string) { return 200, releaseJSON(r) })
	a := s.Check(context.Background())
	if a.Comparison != CLIDistributionUnavailable {
		t.Fatal(a)
	}
	s = releaseTestService(t, func(*http.Request) (int, string) { return 200, releaseJSON(releaseFixture("2.0.0")) })
	before := ToolVersion
	defer func() { ToolVersion = before }()
	ToolVersion = "checkout\nunknown"
	a = s.Check(context.Background())
	if a.Comparison != CLIUncomparable {
		t.Fatal(a)
	}
	ToolVersion = "2.0.0"
	a = s.Check(context.Background())
	if a.Comparison != CLIEqual || !a.Cached {
		t.Fatal(a)
	}
	ToolVersion = "99999999999999999999999999999999999.0.0"
	a = s.Check(context.Background())
	if a.Comparison != CLIAhead {
		t.Fatal(a)
	}
}
func TestCLIReleaseCacheLifecycle(t *testing.T) {
	calls := 0
	status := 200
	s := releaseTestService(t, func(*http.Request) (int, string) { calls++; return status, releaseJSON(releaseFixture("2.0.0")) })
	now := s.now()
	s.Now = func() time.Time { return now }
	first := s.Check(context.Background())
	for i := 0; i < 5; i++ {
		a := s.Check(context.Background())
		if !a.Cached || a.Freshness != AdvisoryFresh {
			t.Fatal(a)
		}
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	now = now.Add(25 * time.Hour)
	status = 503
	a := s.Check(context.Background())
	if a.Freshness != AdvisoryStale || a.Comparison != CLIUpdate || !a.ObservedAt.Equal(*first.ObservedAt) {
		t.Fatal(a)
	}
	_ = s.Check(context.Background())
	if calls != 2 {
		t.Fatal(calls)
	}
	now = now.Add(16 * time.Minute)
	status = 200
	a = s.Check(context.Background())
	if a.Freshness != AdvisoryFresh || a.Cached || calls != 3 {
		t.Fatal(a, calls)
	}
	now = now.Add(-48 * time.Hour)
	a = s.Check(context.Background())
	if a.Freshness != AdvisoryFresh || calls != 4 {
		t.Fatal(a, calls)
	}
}
func TestCLIReleaseEmptyAndColdFailureCache(t *testing.T) {
	for _, status := range []int{200, 403} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			s := releaseTestService(t, func(*http.Request) (int, string) { calls++; return status, "[]" })
			a := s.Check(context.Background())
			b := s.Check(context.Background())
			if calls != 1 || a.Comparison != b.Comparison {
				t.Fatal(a, b, calls)
			}
			if status == 200 && (!b.Cached || b.Comparison != CLINoRelease) {
				t.Fatal(b)
			}
			if status == 403 && (b.Cached || b.ObservedAt != nil || b.Freshness != AdvisoryUnavailable) {
				t.Fatal(b)
			}
		})
	}
}
func TestCLIReleaseUnsafeCacheAndCorruption(t *testing.T) {
	s := releaseTestService(t, func(*http.Request) (int, string) { return 200, "[]" })
	if err := os.MkdirAll(s.CacheRoot, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.CacheRoot, "release.json")
	for _, data := range []string{"{", `{"schema":99}`, strings.Repeat("x", maxStatusCacheBytes+1)} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if a := s.Check(context.Background()); a.Comparison != CLINoRelease || a.Cached {
			t.Fatal(a)
		}
	}
	victim := filepath.Join(t.TempDir(), "victim")
	os.WriteFile(victim, []byte("preserved"), 0600)
	os.Remove(path)
	if err := os.Symlink(victim, path); err != nil {
		t.Skip(err)
	}
	a := s.Check(context.Background())
	if !strings.Contains(a.Error, "cache could not be written") {
		t.Fatal(a)
	}
	data, _ := os.ReadFile(victim)
	if string(data) != "preserved" {
		t.Fatal("symlink target changed")
	}
}
func TestCLIReleaseDeadlineAndConcurrentCheck(t *testing.T) {
	entered := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); close(entered); <-r.Context().Done() }))
	defer server.Close()
	s := releaseTestService(t, func(*http.Request) (int, string) { panic("unused") })
	s.Client = &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		target := *r.URL
		target.Scheme = "http"
		target.Host = strings.TrimPrefix(server.URL, "http://")
		r = r.Clone(r.Context())
		r.URL = &target
		return http.DefaultTransport.RoundTrip(r)
	})}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done := make(chan CLIAdvisory, 1)
	start := time.Now()
	go func() { done <- s.Check(ctx) }()
	<-entered
	peer := s.Check(context.Background())
	if peer.Freshness != AdvisoryUnavailable || !strings.Contains(peer.Error, "another check") {
		t.Fatal(peer)
	}
	a := <-done
	if a.Freshness != AdvisoryUnavailable || time.Since(start) > time.Second || calls.Load() != 1 {
		t.Fatal(a, calls.Load())
	}
}
func TestReviewedCLIAdvisory(t *testing.T) {
	now := time.Now().UTC()
	e := reviewedPlanFixture()
	e.CLIAdvisory = &CLIAdvisory{RunningVersion: "1.0.0", AvailableVersion: "2.0.0", Comparison: CLIUpdate, Freshness: AdvisoryFresh, ObservedAt: &now, ReleaseURL: releasePage + "2.0.0"}
	load := func(data []byte) (ReviewedPlan, error) {
		path := filepath.Join(t.TempDir(), "plan.json")
		os.WriteFile(path, data, 0600)
		sum := sha256.Sum256(data)
		return LoadReviewedPlan(path, hex.EncodeToString(sum[:]))
	}
	data, _ := json.Marshal(e)
	reviewed, err := load(data)
	if err != nil {
		t.Fatal(err)
	}
	current := e
	current.CLIAdvisory = nil
	if err := VerifyReviewedPlanRecheck(reviewed, current); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*CLIAdvisory){func(a *CLIAdvisory) { a.Comparison = CLIEqual }, func(a *CLIAdvisory) { a.ReleaseURL = "https://evil.example" }, func(a *CLIAdvisory) { a.Freshness = AdvisoryUnavailable }, func(a *CLIAdvisory) { a.ObservedAt = nil }, func(a *CLIAdvisory) { a.AvailableVersion = "02.0.0" }} {
		copy := *e.CLIAdvisory
		mutate(&copy)
		current = e
		current.CLIAdvisory = &copy
		data, _ := json.Marshal(current)
		if _, err := load(data); err == nil {
			t.Fatal("accepted invalid CLI metadata", copy)
		}
	}
}

func BenchmarkCLIStatusWarm(b *testing.B) {
	s := CLIStatusService{CacheRoot: filepath.Join(b.TempDir(), "cache"), Platform: "darwin/arm64", Client: &http.Client{Transport: releaseTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("[]")), Header: make(http.Header)}, nil
	})}}
	s.Check(context.Background())
	s.Client = &http.Client{Transport: releaseTransport(func(*http.Request) (*http.Response, error) { b.Fatal("warm HTTP request"); return nil, nil })}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if a := s.Check(context.Background()); !a.Cached || a.Comparison != CLINoRelease {
			b.Fatal(a)
		}
	}
}

func TestCLIReleaseCacheRequiresExplicitObservation(t *testing.T) {
	for _, field := range []string{"schema", "source", "platform", "version", "installable", "observedAt", "retryAt"} {
		for _, null := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/null=%v", field, null), func(t *testing.T) {
				calls := 0
				s := releaseTestService(t, func(*http.Request) (int, string) { calls++; return 200, releaseJSON(releaseFixture("2.0.0")) })
				directory, err := disposableCacheDirectory(s.CacheRoot, "cli-status")
				if err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(cliReleaseCache{Schema: 1, Source: releaseRepository, Platform: s.platform(), ObservedAt: s.now()})
				var fields map[string]any
				json.Unmarshal(data, &fields)
				if null {
					fields[field] = nil
				} else {
					delete(fields, field)
				}
				data, _ = json.Marshal(fields)
				os.WriteFile(filepath.Join(directory, "release.json"), data, 0600)
				a := s.Check(context.Background())
				if a.Comparison != CLIUpdate || a.Cached || calls != 1 {
					t.Fatal(a, calls)
				}
			})
		}
	}
}

func TestCLIReleasePackagingContract(t *testing.T) {
	data, err := os.ReadFile("../../packaging/targets.json")
	if err != nil {
		t.Fatal(err)
	}
	var targets []struct{ OS, Arch, Archive string }
	if err = json.Unmarshal(data, &targets); err != nil {
		t.Fatal(err)
	}
	for _, target := range targets {
		t.Run(target.OS+"/"+target.Arch, func(t *testing.T) {
			version := "2.0.0"
			r := releaseFixture(version)
			installer := "install.sh"
			if target.OS == "windows" {
				installer = "install.ps1"
			}
			assets := []map[string]any{}
			for _, name := range []string{fmt.Sprintf("sjskills_%s_%s_%s.%s", version, target.OS, target.Arch, target.Archive), "SHA256SUMS", installer} {
				assets = append(assets, map[string]any{"name": name, "state": "uploaded", "size": 100, "browser_download_url": "https://github.com/" + releaseRepository + "/releases/download/sjskills-v" + version + "/" + name})
			}
			r["assets"] = assets
			s := releaseTestService(t, func(*http.Request) (int, string) { return 200, releaseJSON(r) })
			s.Platform = target.OS + "/" + target.Arch
			if a := s.Check(context.Background()); a.Comparison != CLIUpdate {
				t.Fatal(a)
			}
			assets[0]["browser_download_url"] = "https://example.com/archive"
			s.CacheRoot = filepath.Join(t.TempDir(), "cache")
			if a := s.Check(context.Background()); a.Comparison != CLIDistributionUnavailable {
				t.Fatal(a)
			}
		})
	}
}
