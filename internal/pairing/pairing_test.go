package pairing

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTokenIsRandomAndURLSafe(t *testing.T) {
	t.Parallel()

	first, err := NewToken()
	require.NoError(t, err)
	second, err := NewToken()
	require.NoError(t, err)

	require.NotEqual(t, first, second)
	require.Len(t, first, 43, "32 bytes of entropy in raw URL-safe base64")
	for _, r := range first {
		valid := r == '-' || r == '_' ||
			(r >= '0' && r <= '9') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z')
		require.True(t, valid, "unexpected character %q in token", r)
	}
}

func TestSaveAndRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "token")
	require.NoError(t, Save(path, "abc123"))

	token, err := Read(path)
	require.NoError(t, err)
	require.Equal(t, "abc123", token)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestReadRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte("  \n"), 0o600))

	_, err := Read(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

func TestResolvePrefersEnv(t *testing.T) {
	t.Setenv(EnvToken, "from-env")
	t.Setenv(EnvTokenFile, "")

	token, source := Resolve()
	require.Equal(t, "from-env", token)
	require.Equal(t, "env "+EnvToken, source)
}

func TestResolveReadsTokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, Save(path, "from-file"))

	t.Setenv(EnvToken, "")
	t.Setenv(EnvTokenFile, path)

	token, source := Resolve()
	require.Equal(t, "from-file", token)
	require.Equal(t, path, source)
}

func TestResolveEmptyWhenUnpaired(t *testing.T) {
	t.Setenv(EnvToken, "")
	t.Setenv(EnvTokenFile, filepath.Join(t.TempDir(), "missing"))

	token, source := Resolve()
	require.Empty(t, token)
	require.Empty(t, source)
}

func TestMatch(t *testing.T) {
	t.Parallel()

	require.True(t, Match("token", "token"))
	require.False(t, Match("token", "other"))
	require.False(t, Match("", ""))
	require.False(t, Match("token", ""))
	require.False(t, Match("", "token"))
	require.False(t, Match("token-longer", "token"))
}
