package pr

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRemote(t *testing.T) {
	t.Parallel()

	cases := []struct {
		url           string
		host, owner   string
		repository    string
		expectFailure bool
	}{
		{url: "https://github.com/GenLabsAI/star.git", host: "github.com", owner: "GenLabsAI", repository: "star"},
		{url: "https://github.com/GenLabsAI/star", host: "github.com", owner: "GenLabsAI", repository: "star"},
		{url: "git@github.com:GenLabsAI/star.git", host: "github.com", owner: "GenLabsAI", repository: "star"},
		{url: "ssh://git@github.com/GenLabsAI/star.git", host: "github.com", owner: "GenLabsAI", repository: "star"},
		{url: "https://user:token@github.com/owner/name.git", host: "github.com", owner: "owner", repository: "name"},
		{url: "/tmp/a-repo.git", expectFailure: true},
	}
	for _, tc := range cases {
		host, owner, name, err := parseRemote(tc.url)
		if tc.expectFailure {
			require.Error(t, err, tc.url)
			continue
		}
		require.NoError(t, err, tc.url)
		require.Equal(t, tc.host, host, tc.url)
		require.Equal(t, tc.owner, owner, tc.url)
		require.Equal(t, tc.repository, name, tc.url)
	}
}

func TestNumberFromURL(t *testing.T) {
	t.Parallel()

	require.Equal(t, 42, numberFromURL("https://github.com/o/n/pull/42"))
	require.Equal(t, 7, numberFromURL("https://github.com/o/n/pull/7/"))
	require.Equal(t, 0, numberFromURL("https://github.com/o/n"))
}

func TestCreateRequiresTitle(t *testing.T) {
	t.Parallel()

	_, err := Create(t.TempDir(), Options{})
	require.ErrorContains(t, err, "title is required")
}

func TestCreateRefusesOutsideARepository(t *testing.T) {
	t.Parallel()

	_, err := Create(t.TempDir(), Options{Title: "fix: nothing"})
	require.ErrorContains(t, err, "not a git repository")
}

func TestCreateHonoursApprovalPolicy(t *testing.T) {
	t.Parallel()

	dir := initRepository(t)
	writePolicy(t, dir, `{"commit":{"require_approval":true,"approver":"the reviewer"}}`)

	_, err := Create(dir, Options{Title: "feat: gated"})
	var needsApproval *ApprovalRequiredError
	require.True(t, errors.As(err, &needsApproval))
	require.Equal(t, "main", needsApproval.Branch)
	require.Contains(t, err.Error(), "requires approval for the reviewer")
	require.Contains(t, err.Error(), "--approve")

	// Approval gets past the gate; the next thing to stop the run is the
	// missing remote, which proves the policy was not consulted twice.
	_, err = Create(dir, Options{Title: "feat: gated", Approve: true})
	require.ErrorContains(t, err, "no remote")
	require.False(t, errors.As(err, &needsApproval))
}

func TestCreateRefusesABranchThatIsNotPushed(t *testing.T) {
	t.Parallel()

	dir := initRepository(t)
	runGit(t, dir, "remote", "add", "origin", "https://github.com/example/repo.git")
	runGit(t, dir, "checkout", "-q", "-b", "feature")

	_, err := Create(dir, Options{Title: "feat: local only"})
	require.ErrorContains(t, err, `is not pushed`)
	require.ErrorContains(t, err, "--push")
}

func TestCreateRefusesDetachedHead(t *testing.T) {
	t.Parallel()

	dir := initRepository(t)
	runGit(t, dir, "checkout", "-q", "--detach")

	_, err := Create(dir, Options{Title: "feat: detached"})
	require.ErrorContains(t, err, "detached HEAD")
}

func initRepository(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("one\n"), 0o644))
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-q", "-m", "init")
	return dir
}

func writePolicy(t *testing.T, dir, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "crush.json"), []byte(body), 0o644))
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}
