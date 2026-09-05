//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// launcherHandshake coordinates with the Rust launcher (star.exe) so that
// Bubble Tea never touches the terminal until the launcher's splash
// animation has fully released the alt-screen buffer. Without this gate,
// bubbletea's renderer and the launcher's animation thread both write to
// the same alt-screen buffer concurrently, causing visible flicker.
type launcherHandshake struct {
	active        bool
	readyEvent    windows.Handle
	releaseEvent  windows.Handle
	renderedEvent windows.Handle
}

func newLauncherHandshake() *launcherHandshake {
	if os.Getenv("STAR_LAUNCHER_HANDSHAKE") != "1" {
		return &launcherHandshake{}
	}
	pid := os.Getenv("STAR_LAUNCHER_PID")
	readyName, _ := windows.UTF16PtrFromString("Local\\star-ready-" + pid)
	releaseName, _ := windows.UTF16PtrFromString("Local\\star-release-" + pid)
	renderedName, _ := windows.UTF16PtrFromString("Local\\star-rendered-" + pid)
	readyEvent, err1 := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, readyName)
	releaseEvent, err2 := windows.OpenEvent(windows.SYNCHRONIZE, false, releaseName)
	renderedEvent, err3 := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, renderedName)
	if err1 != nil || err2 != nil || err3 != nil {
		return &launcherHandshake{}
	}
	return &launcherHandshake{
		active:        true,
		readyEvent:    readyEvent,
		releaseEvent:  releaseEvent,
		renderedEvent: renderedEvent,
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
	if err := windows.SetEvent(h.readyEvent); err != nil {
		return fmt.Errorf("notify launcher: %w", err)
	}
	if _, err := windows.WaitForSingleObject(h.releaseEvent, windows.INFINITE); err != nil {
		return fmt.Errorf("wait for launcher: %w", err)
	}
	_ = windows.CloseHandle(h.readyEvent)
	_ = windows.CloseHandle(h.releaseEvent)
	return nil
}

// notifyRendered tells the launcher that bubbletea has produced its first
// real frame, so the launcher process can finally exit. Safe to call
// multiple times; only the first call has any effect.
func (h *launcherHandshake) notifyRendered() {
	if !h.active {
		return
	}
	_ = windows.SetEvent(h.renderedEvent)
	_ = windows.CloseHandle(h.renderedEvent)
	h.active = false
}
