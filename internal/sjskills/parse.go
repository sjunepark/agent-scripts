package sjskills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// ParseRegistry decodes a version-4 registry and validates its complete
// desired-state contract. Unknown JSON fields are rejected before semantic
// validation.
func ParseRegistry(data []byte) (Registry, error) {
	var registry Registry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, decodeIssue("registry", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Registry{}, &ValidationErrors{Issues: []Issue{{
				Code: IssueMalformedInput, Path: "registry", Message: "multiple JSON documents are not allowed",
			}}}
		}
		return Registry{}, decodeIssue("registry", err)
	}
	if err := ValidateRegistry(registry); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func ReadRegistry(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, fmt.Errorf("read registry %s: %w", path, err)
	}
	return ParseRegistry(data)
}

// ParseManifest strictly decodes the committed project manifest. TOML's
// metadata records unknown keys, including keys nested in direct tables.
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	metadata, err := toml.Decode(string(data), &manifest)
	if err != nil {
		return Manifest{}, decodeIssue("manifest", err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		issues := make([]Issue, 0, len(undecoded))
		for _, key := range undecoded {
			issues = append(issues, Issue{
				Code:    IssueUnknownField,
				Path:    "manifest." + key.String(),
				Message: "unknown field",
			})
		}
		return Manifest{}, newValidationErrors(issues)
	}
	if err := ValidateManifestShape(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %s: %w", path, err)
	}
	return ParseManifest(data)
}

func decodeIssue(path string, err error) error {
	code := IssueMalformedInput
	message := err.Error()
	if strings.Contains(message, "unknown field") {
		code = IssueUnknownField
	}
	return newValidationErrors([]Issue{{Code: code, Path: path, Message: message}})
}
