package main

import (
	"github.com/ramonvermeulen/whosthere/cmd"
	"github.com/ramonvermeulen/whosthere/internal/version"
)

// These are intended to be set by GoReleaser (or other build tooling) via -ldflags.
// By convention:
//   - main.version: current git tag (without leading v), or snapshot name
//   - main.commit:  current git commit SHA
//   - main.date:    build date in RFC3339 format
var (
	versionStr = "dev"
	commitStr  = "none"
	dateStr    = "unknown"
)

func main() {
	// Propagate build metadata into the internal version package so it can be
	// used consistently across the CLI, TUI header, and logs.
	version.Version = versionStr
	version.Commit = commitStr
	version.Date = dateStr

	cmd.Execute()
}
