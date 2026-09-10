package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func initRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("hi"), 0o644))
	run("add", ".")
	run("commit", "-m", "init")
	return root
}

func TestManagerCreateAndRemove(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := initRepo(t)
	base := filepath.Join(root, ".crush", "worktrees")
	mgr := NewManager(root, base)

	path, err := mgr.Create(t.Context(), "session-1", "")
	require.NoError(t, err)
	require.DirExists(t, path)

	repeated, err := mgr.Create(t.Context(), "session-1", "")
	require.NoError(t, err)
	require.Equal(t, path, repeated)

	require.NoError(t, mgr.Remove(t.Context(), "session-1"))
	require.NoDirExists(t, path)
	require.NoError(t, mgr.Remove(t.Context(), "unknown"))
}

func TestManagerCreateWithBranchAndCleanup(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := initRepo(t)
	base := filepath.Join(root, ".crush", "worktrees")
	mgr := NewManager(root, base)

	path, err := mgr.Create(t.Context(), "branchy", "feature/x")
	require.NoError(t, err)
	require.DirExists(t, path)

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(out), "feature/x")

	require.NoError(t, mgr.Cleanup(t.Context()))
	require.NoDirExists(t, path)
	require.NoError(t, mgr.Cleanup(t.Context()))
}
