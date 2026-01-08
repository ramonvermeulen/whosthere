package version

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// Version and Commit are set via -ldflags at build time.
// Defaults are useful for local development builds.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

const (
	// ANSII color codes for labels
	cyan  = "\033[36m"
	reset = "\033[0m"
)

// Fprint writes version and runtime information to the provided writer.
func Fprint(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	_, _ = fmt.Fprintf(w, "%sOS:%s         %s/%s\n", cyan, reset, runtime.GOOS, runtime.GOARCH)
	_, _ = fmt.Fprintf(w, "%sVersion:%s    %s\n", cyan, reset, Version)
	_, _ = fmt.Fprintf(w, "%sCommit:%s     %s\n", cyan, reset, Commit)
	_, _ = fmt.Fprintf(w, "%sDate:%s       %s\n", cyan, reset, Date)
}
