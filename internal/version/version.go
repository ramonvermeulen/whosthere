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

// Fprint writes version and runtime information to the provided writer.
func Fprint(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	_, _ = fmt.Fprintf(w, "whosthere version: %s\n", Version)
	_, _ = fmt.Fprintf(w, "Git commit:        %s\n", Commit)
	_, _ = fmt.Fprintf(w, "Build date:       %s\n", Date)
	_, _ = fmt.Fprintf(w, "%s/%s\n", runtime.GOOS, runtime.GOARCH)
}
