//go:build sjskillstest

package sjskills

import "os"

// CLI integration tests build with -tags sjskillstest and serve a fixture
// registry at SJSKILLS_TEST_REGISTRY_URL (/refs and /raw/<commit>/<file>).
func init() {
	if base := os.Getenv("SJSKILLS_TEST_REGISTRY_URL"); base != "" {
		registryEndpoints = func() (string, string) { return base + "/refs", base + "/raw/" }
	}
}
