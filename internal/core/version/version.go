package version

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
)

// Version and Commit are set via -ldflags at build time.
// Defaults are useful for local development builds.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
	// ANSII color codes for labels
	magenta = "\x1b[35m"
	reset   = "\033[0m"
)

// Fprint writes version and runtime information to the provided writer.
func Fprint(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if theme.IsNoColor() {
		magenta = ""
		reset = ""
	}
	_, _ = fmt.Fprintf(w, "%sOS:%s         %s/%s\n", magenta, reset, runtime.GOOS, runtime.GOARCH)
	_, _ = fmt.Fprintf(w, "%sVersion:%s    %s\n", magenta, reset, Version)
	_, _ = fmt.Fprintf(w, "%sCommit:%s     %s\n", magenta, reset, Commit)
	dateStr := Date
	if len(Date) >= 10 {
		dateStr = Date[:10]
	}
	_, _ = fmt.Fprintf(w, "%sDate:%s       %s\n", magenta, reset, dateStr)
}
