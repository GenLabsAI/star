//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// launcherHandshake coordinates with the Rust launcher (star binary) so
// that Bubble Tea never touches the terminal until the launcher's splash
// animation has fully released the alt-screen buffer. It mirrors
// splash_windows.go but uses sentinel files in the OS temp directory
// instead of Win32 named events, since Unix has no portable equivalent.
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

func waitForFile(path string) {
	for {
		if _, err := os.Stat(path); err == nil {
			_ = os.Remove(path)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// awaitRelease notifies the launcher that Go has finished initializing and
// blocks until the launcher has stopped animating and given up ownership of
// the terminal. It must be called before tea.NewProgram/Run so bubbletea
// never writes to the terminal while the launcher still owns it.
func (h *launcherHandshake) awaitRelease() error {
	if !h.active {
		return nil
	}
	if err := touchFile(h.readyPath); err != nil {
		return fmt.Errorf("notify launcher: %w", err)
	}
	waitForFile(h.releasePath)
	return nil
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
