package cmd

import (
	"context"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

// Ctrl+C while the splash is still up must leave. Dropping it made the splash
// eat a press the quit dialog never saw, so quitting took an extra press.
func TestSplashModelQuitsOnCtrlC(t *testing.T) {
	t.Parallel()

	m := newSplashModel(nil, "", false)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd, "ctrl+c on the splash must quit")
	require.IsType(t, tea.QuitMsg{}, cmd())
}

// An interrupt ends the launch instead of waiting for a handoff that will
// never come, which is what left the process parked on the splash.
func TestLauncherHandshakeAwaitReleaseHonorsCancellation(t *testing.T) {
	t.Parallel()

	h := &launcherHandshake{
		active:      true,
		releasePath: filepath.Join(t.TempDir(), "star-release-test"),
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, h.awaitRelease(ctx), context.Canceled)
}

func TestSplashModelAltScreenWhenLauncherOwnsIt(t *testing.T) {
	t.Parallel()

	m := newSplashModel(nil, "", false)
	require.True(t, m.View().AltScreen, "a directly launched core owns the alt screen")

	m.launcherOwnsAltScreen = true
	require.False(t, m.View().AltScreen,
		"the splash stays on the screen the launcher already opened")
}
