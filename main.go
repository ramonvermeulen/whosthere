package main

import (
	"github.com/ramonvermeulen/whosthere/internal/cmd"
	"github.com/ramonvermeulen/whosthere/internal/core/version"
)

// These are intended to be set by GoReleaser (or other build tooling) via -ldflags.
// By convention:
//   - main.version: current git tag (without leading v), or snapshot name
//   - main.commit:  current git commit SHA
//   - main.date:    build date in RFC3339 format
var (
	versionStr = "dev"
	commitStr  = "unknown"
	dateStr    = "unknown"
)

func main() {
	version.Version = versionStr
	version.Commit = commitStr
	version.Date = dateStr

	cmd.Execute()
}
