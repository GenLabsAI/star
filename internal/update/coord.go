package update

import (
	"os"
	"path/filepath"
)

// Coordination between the core and the launcher that supervises it during an
// in-place self-update. This mirrors `launcher/src/coord.rs`; the two have to
// agree on the layout, so keep the constants here in step with that file.
//
// Updating means overwriting the `star-core` binary every running instance is
// currently executing, so each instance closes its core, records the session it
// was in, and reopens it on the new binary afterwards. Recording the session
// on *every* pause (not just the one the user clicked "Update Now" in) is what
// lets a session that was closed by somebody else's update come back on the
// same conversation instead of an empty one.

const (
	// requestFile is dropped by the core the user asked to update, and watched
	// by every other core so they can get out of the way.
	requestFile = "request"
	// sessionsDir holds one file per launcher pid, containing the session that
	// instance should resume with.
	sessionsDir = "sessions"
)

// Exit codes the launcher interprets as "pause and resume", as opposed to a
// real exit the user asked for.
const (
	ExitInitiator = 42
	ExitFollower  = 43
)

// CoordDir returns the directory the launcher coordinates updates through.
// The launcher passes it in; the fallback exists for a core run directly, where
// there is no update coordination to speak of.
func CoordDir() string {
	if dir := os.Getenv("STAR_UPDATE_DIR"); dir != "" {
		return dir
	}
	name := "star-update"
	if user := firstNonEmpty(os.Getenv("USER"), os.Getenv("USERNAME")); user != "" {
		name += "-" + user
	}
	return filepath.Join(os.TempDir(), name)
}

// PublishRequest asks every other running instance to pause so the binary can
// be replaced. Callers must only do this from the instance the user clicked
// "Update Now" in.
func PublishRequest() error {
	dir := CoordDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, requestFile), nil, 0o600)
}

// RequestPending reports whether an update was requested after this instance
// launched. Comparing nanosecond timestamps (rather than seconds) is what
// distinguishes "requested while I was running" from "left over from an
// earlier run" — a request left behind by an update that crashed would
// otherwise make every freshly started instance tear itself straight back down.
func RequestPending(launchedNanos int64) bool {
	info, err := os.Stat(filepath.Join(CoordDir(), requestFile))
	if err != nil {
		return false
	}
	return info.ModTime().UnixNano() > launchedNanos
}

// RecordSession stores the session this instance is in so the launcher can
// resume it after the pause. An empty session means there was nothing to
// resume and is not recorded.
func RecordSession(launcherPID, session string) error {
	if launcherPID == "" || session == "" {
		return nil
	}
	dir := filepath.Join(CoordDir(), sessionsDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, launcherPID), []byte(session), 0o600)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
