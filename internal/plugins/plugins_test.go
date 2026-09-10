package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeManifest(t *testing.T, dir string, manifest Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ManifestFile), data, 0o644))
}

func TestInstallListAndRemove(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "source")
	writeManifest(t, source, Manifest{Name: "test-plugin", Version: "1.0.0"})
	require.NoError(t, os.WriteFile(filepath.Join(source, "asset.txt"), []byte("ok"), 0o644))
	directory := filepath.Join(t.TempDir(), "plugins")

	installed, err := Install(source, directory)
	require.NoError(t, err)
	require.Equal(t, "test-plugin", installed.Name)

	listed, err := List(directory)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.True(t, listed[0].Enabled)

	require.NoError(t, SetEnabled("test-plugin", directory, false))
	listed, err = List(directory)
	require.NoError(t, err)
	require.False(t, listed[0].Enabled)

	require.NoError(t, SetEnabled("test-plugin", directory, true))
	require.NoError(t, Remove("test-plugin", directory))
	listed, err = List(directory)
	require.NoError(t, err)
	require.Empty(t, listed)
}

func TestLoadRejectsInvalidManifest(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeManifest(t, directory, Manifest{Name: "../bad"})
	_, err := Load(directory)
	require.Error(t, err)
}
