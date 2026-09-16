package version

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

// Build-time parameters set via -ldflags.

var (
	Version = "devel"
	Commit  = "unknown"
	// BuildID is a unique identifier for this build. For release builds it
	// equals Commit; for development builds (go run / go build without
	// ldflags) it is derived from the executable's modification time, which
	// changes on every recompilation.
	BuildID = ""
)

// A user may install crush using `go install github.com/charmbracelet/crush@latest`.
// without -ldflags, in which case the version above is unset. As a workaround
// we use the embedded build version that *is* set when using `go install` (and
// is only set for `go install` and not for `go build`).
func init() {
	info, ok := debug.ReadBuildInfo()
	if ok && Version == "devel" {
		mainVersion := info.Main.Version
		// Only use the build info version if it looks like a clean
		// semver tag (e.g. "v1.2.3"), not a pseudo-version like
		// "v0.91.1-0.20260915221056-83e735f6...". Pseudo-versions
		// are confusing to display and break the update checker.
		if mainVersion != "" && mainVersion != "(devel)" && !strings.Contains(mainVersion, "-0.") {
			Version = mainVersion
		}
	}

	// Derive BuildID when not set via ldflags.
	if BuildID == "" {
		BuildID = deriveBuildID()
	}
}

// deriveBuildID uses the running executable's modification time as a unique
// build fingerprint. This changes on every recompilation (including `go run`),
// making it reliable for detecting stale servers during development.
func deriveBuildID() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	fi, err := os.Stat(exe)
	if err != nil {
		return "unknown"
	}
	return strconv.FormatInt(fi.ModTime().UnixNano(), 36)
}
