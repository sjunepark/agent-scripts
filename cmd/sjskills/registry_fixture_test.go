package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/sjunepark/agent-scripts/internal/sjskills"
)

// fixtureRegistryCommit is the commit the fake published source advertises.
const fixtureRegistryCommit = "0123456789abcdef0123456789abcdef01234567"

const fixtureRegistryPath = "../../internal/sjskills/testdata/registry-v4.json"

// registryFunc adapts one load result to every registry policy for
// in-process tests; the reviewed commit passed to At is ignored.
type registryFunc func() (sjskills.PublishedRegistry, error)

func (f registryFunc) Resolve(context.Context) (sjskills.PublishedRegistry, error) { return f() }
func (f registryFunc) At(context.Context, string) (sjskills.PublishedRegistry, error) {
	return f()
}
func (f registryFunc) ForStatus(context.Context) (sjskills.PublishedRegistry, error) { return f() }
func (f registryFunc) ForSelection(context.Context) (sjskills.PublishedRegistry, error) {
	return f()
}

// fixtureRegistry is the unpinned fixture registry used by directly
// constructed applications, whose materializers ignore source locations.
func fixtureRegistry() (sjskills.PublishedRegistry, error) {
	data, err := os.ReadFile(fixtureRegistryPath)
	if err != nil {
		return sjskills.PublishedRegistry{}, err
	}
	registry, err := sjskills.ParseRegistry(data)
	if err != nil {
		return sjskills.PublishedRegistry{}, err
	}
	sum := sha256.Sum256(data)
	return sjskills.PublishedRegistry{Registry: registry, Commit: fixtureRegistryCommit, SHA256: hex.EncodeToString(sum[:])}, nil
}

func init() { defaultRegistries = registryFunc(fixtureRegistry) }

func fixtureRegistryValue() (sjskills.Registry, error) {
	published, err := fixtureRegistry()
	return published.Registry, err
}

// startFixtureRegistryServer serves the fixture registry as the published
// source for CLI binaries built with the sjskillstest tag.
func startFixtureRegistryServer() (*httptest.Server, error) {
	data, err := os.ReadFile(fixtureRegistryPath)
	if err != nil {
		return nil, err
	}
	pkt := func(value string) string { return fmt.Sprintf("%04x%s", len(value)+4, value) }
	refs := pkt("# service=git-upload-pack\n") + "0000" + pkt(fixtureRegistryCommit+" refs/heads/main\n") + "0000"
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/refs":
			w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
			fmt.Fprint(w, refs)
		case r.URL.Path == "/raw/"+fixtureRegistryCommit+"/skill-registry.json":
			_, _ = w.Write(data)
		default:
			http.NotFound(w, r)
		}
	})), nil
}
