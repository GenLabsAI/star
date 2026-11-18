package pairing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApprovalRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "approval.json")
	minted, err := SaveApproval(path, time.Minute)
	require.NoError(t, err)
	require.Len(t, minted.Code, 8)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	stored, err := ReadApproval(path)
	require.NoError(t, err)
	require.Equal(t, minted.Code, stored.Code)

	now := time.Now()
	require.True(t, stored.Valid(minted.Code, now))
	require.False(t, stored.Valid("WRONGCODE", now))
	require.False(t, stored.Valid("", now))
	require.False(t, stored.Valid(minted.Code, now.Add(2*time.Minute)))
}

func TestApprovalCodesDiffer(t *testing.T) {
	t.Parallel()

	first, err := NewApproval()
	require.NoError(t, err)
	second, err := NewApproval()
	require.NoError(t, err)
	require.NotEqual(t, first, second)
}

func TestReadApprovalRejectsGarbage(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "approval.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
	_, err := ReadApproval(path)
	require.ErrorContains(t, err, "holds no code")

	require.NoError(t, os.WriteFile(path, []byte("{oops"), 0o600))
	_, err = ReadApproval(path)
	require.ErrorContains(t, err, "Failed to decode approval")

	_, err = ReadApproval(filepath.Join(t.TempDir(), "missing.json"))
	require.ErrorContains(t, err, "Failed to read approval")
}
