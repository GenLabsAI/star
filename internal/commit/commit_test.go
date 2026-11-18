package commit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newRepo creates a git repository with one committed file and returns
// its path.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q", "-b", "main")
	run(t, dir, "git", "config", "user.email", "proof@example.com")
	run(t, dir, "git", "config", "user.name", "Proof")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o644))
	run(t, dir, "git", "add", "-A")
	run(t, dir, "git", "commit", "-q", "-m", "init")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func headMessage(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(run(t, dir, "git", "log", "-1", "--format=%s"))
}

func TestCommitNeedsAMessage(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	_, err := Commit(dir, Options{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "commit message is required")
	require.Equal(t, "init", headMessage(t, dir))
}

func TestCommitRefusesUntrackedWithoutApproval(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o644))

	_, err := Commit(dir, Options{Message: "feat: change"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "untracked")
	require.Contains(t, err.Error(), "new.txt")
	require.Equal(t, "init", headMessage(t, dir), "a refused commit must not create one")
}

func TestCommitStagesTrackedChangesOnly(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	record, err := Commit(dir, Options{Message: "docs: change the readme"})
	require.NoError(t, err)
	require.Equal(t, "docs: change the readme", headMessage(t, dir))
	require.NotEmpty(t, record.Short)
	require.Len(t, record.Files, 1)
	require.Equal(t, "README.md", record.Files[0].Path)

	// The commit is real: HEAD contains the new bytes.
	require.Equal(t, "changed", strings.TrimSpace(run(t, dir, "git", "show", "HEAD:README.md")))
	require.Equal(t, "", strings.TrimSpace(run(t, dir, "git", "status", "--porcelain")))
}

func TestCommitIncludesUntrackedWhenApproved(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o644))

	record, err := Commit(dir, Options{Message: "feat: add new", IncludeUntracked: true})
	require.NoError(t, err)
	require.Len(t, record.Files, 2)
	require.Equal(t, "new", strings.TrimSpace(run(t, dir, "git", "show", "HEAD:new.txt")))
	require.Contains(t, record.Stat, "new.txt")
}

func TestCommitCreatesAndReusesBranch(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("one\n"), 0o644))

	record, err := Commit(dir, Options{Message: "feat: work", Branch: "feat/one"})
	require.NoError(t, err)
	require.True(t, record.BranchCreated)
	require.Equal(t, "feat/one", record.Branch)
	require.Equal(t, "feat/one", strings.TrimSpace(run(t, dir, "git", "branch", "--show-current")))
	require.Equal(t, "main", strings.TrimSpace(run(t, dir, "git", "rev-parse", "--abbrev-ref", "main")))

	// Committing again on the same branch reuses it instead of failing.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("two\n"), 0o644))
	again, err := Commit(dir, Options{Message: "feat: more", Branch: "feat/one"})
	require.NoError(t, err)
	require.False(t, again.BranchCreated)
	require.Equal(t, "feat/one", again.Branch)
}

func TestCommitNothingToCommit(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	_, err := Commit(dir, Options{Message: "chore: empty"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Nothing to commit")
	require.Equal(t, "init", headMessage(t, dir))
}

func TestCommitOutsideRepository(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := Commit(dir, Options{Message: "chore: nope"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a git repository")
}
