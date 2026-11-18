package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// restoreTerminalSeq hands the terminal back after a launch we interrupted
// ourselves. It mirrors the launcher's own restore sequence.
const restoreTerminalSeq = "\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r"

// restoreTerminal leaves the alternate screen and shows the cursor again.
// Used when an interrupt ends the launch before Bubble Tea or the launcher
// could restore the terminal, so the user is not stranded on the splash.
func restoreTerminal() {
	fmt.Fprint(os.Stdout, restoreTerminalSeq)
}

// launcherHandshake coordinates with the Rust launcher (star binary) so
// that Bubble Tea never touches the terminal until the launcher's splash
// animation has fully released the alt-screen buffer. It uses sentinel
// files in the OS temp directory as a cross-platform IPC mechanism.
type launcherHandshake struct {
	active       bool
	readyPath    string
	releasePath  string
	renderedPath string
}

func newLauncherHandshake() *launcherHandshake {
	if os.Getenv("STAR_LAUNCHER_HANDSHAKE") != "1" {
		return &launcherHandshake{}
	}
	pid := os.Getenv("STAR_LAUNCHER_PID")
	if pid == "" {
		return &launcherHandshake{}
	}
	dir := os.TempDir()
	return &launcherHandshake{
		active:       true,
		readyPath:    filepath.Join(dir, "star-ready-"+pid),
		releasePath:  filepath.Join(dir, "star-release-"+pid),
		renderedPath: filepath.Join(dir, "star-rendered-"+pid),
	}
}

func touchFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}

// waitForFile blocks until the launcher creates the sentinel file, or ctx is
// done. It never times out on its own: proceeding to render Bubble Tea before
// the launcher released the terminal breaks the terminal state. Cancellation
// is different — an interrupt means nobody is going to hand anything over, so
// waiting would only hang the process on the splash.
func waitForFile(ctx context.Context, path string) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(path)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// signalReady tells the launcher that the core is alive and initializing.
// It must be called before awaitRelease, and before Bubble Tea is started, so
// the launcher can keep owning the terminal until it hands it over.
func (h *launcherHandshake) signalReady() error {
	if !h.active {
		return nil
	}
	if err := touchFile(h.readyPath); err != nil {
		return fmt.Errorf("notify launcher: %w", err)
	}
	return nil
}

// awaitRelease blocks until the launcher has stopped animating and given up
// ownership of the terminal. It must be called before tea.NewProgram/Run so
// bubbletea never writes to the terminal while the launcher still owns it.
// It returns ctx.Err() when the context is cancelled first, which is how a
// Ctrl+C pressed while the splash is up ends the launch instead of being
// swallowed.
func (h *launcherHandshake) awaitRelease(ctx context.Context) error {
	if !h.active {
		return nil
	}
	return waitForFile(ctx, h.releasePath)
}

// notifyRendered tells the launcher that bubbletea has produced its first
// real frame, so the launcher process can finally exit. Safe to call
// multiple times; only the first call has any effect.
func (h *launcherHandshake) notifyRendered() {
	if !h.active {
		return
	}
	_ = touchFile(h.renderedPath)
	h.active = false
}
