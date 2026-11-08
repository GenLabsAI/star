package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplashModelAltScreenWhenLauncherOwnsIt(t *testing.T) {
	t.Parallel()

	m := newSplashModel(nil, "", false)
	require.True(t, m.View().AltScreen, "a directly launched core owns the alt screen")

	m.launcherOwnsAltScreen = true
	require.False(t, m.View().AltScreen,
		"a launcher-driven core must not request the alt screen a second time")
}
