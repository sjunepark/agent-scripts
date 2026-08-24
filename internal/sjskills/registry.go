package sjskills

import (
	_ "embed"
)

// registryV4JSON is shipped with the command so an installed binary does not
// depend on the repository working tree for the policy catalog.
//
//go:embed data/registry-v4.json
var registryV4JSON []byte

// EmbeddedRegistry returns the canonical v4 registry used by the CLI.
func EmbeddedRegistry() (Registry, error) { return ParseRegistry(registryV4JSON) }
