package sjskills

import (
	_ "embed"
	"strings"
)

// versionFile is the release identity shared by the binary and packaging.
//
//go:embed VERSION
var versionFile string

var ToolVersion = strings.TrimSpace(versionFile)
