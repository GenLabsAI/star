//go:build !windows
// +build !windows

package cmd

type launcherHandshake struct{}

func newLauncherHandshake() *launcherHandshake {
	return &launcherHandshake{}
}

func (h *launcherHandshake) awaitRelease() error {
	return nil
}

func (h *launcherHandshake) notifyRendered() {}
