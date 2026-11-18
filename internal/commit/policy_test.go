package commit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// withPolicy writes a project commit policy into the repo's config file and
// commits it so it is part of the tracked tree.
func withPolicy(t *testing.T, dir, policy string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "crush.json"), []byte(policy), 0o644))
	run(t, dir, "git", "add", "crush.json")
	run(t, dir, "git", "commit", "-q", "-m", "chore: add policy")
}

func headCount(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(run(t, dir, "git", "rev-list", "--count", "HEAD"))
}

func TestLoadPolicyFromProjectConfig(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	withPolicy(t, dir, `{"commit": {"require_approval": true, "protected_branches": ["main"], "approver": "the owner"}}`)

	policy, err := LoadPolicy(dir)
	require.NoError(t, err)
	require.True(t, policy.RequireApproval)
	require.Equal(t, []string{"main"}, policy.ProtectedBranches)
	require.Equal(t, "the owner", policy.Approver)
	require.Equal(t, filepath.Join(dir, "crush.json"), policy.Path)
	require.True(t, policy.RequiresApproval("anything"))
}

func TestLoadPolicyWithoutAConfigIsNoGate(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(newRepo(t))
	require.NoError(t, err)
	require.False(t, policy.RequireApproval)
	require.Empty(t, policy.ProtectedBranches)
	require.False(t, policy.RequiresApproval("main"))
}

func TestCommitRefusedByPolicyLeavesTreeUntouched(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	withPolicy(t, dir, `{"commit": {"require_approval": true, "approver": "the owner"}}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	before := headCount(t, dir)
	statusBefore := strings.TrimSpace(run(t, dir, "git", "status", "--porcelain"))

	_, err := Commit(dir, Options{Message: "feat: change"})
	require.Error(t, err)
	var refusal *ApprovalRequiredError
	require.True(t, errors.As(err, &refusal), "expected an ApprovalRequiredError, got %v", err)
	require.Contains(t, err.Error(), "requires approval")
	require.Contains(t, err.Error(), "--approve")

	require.Equal(t, before, headCount(t, dir))
	require.Equal(t, statusBefore, strings.TrimSpace(run(t, dir, "git", "status", "--porcelain")))
	require.Empty(t, strings.TrimSpace(run(t, dir, "git", "diff", "--cached", "--name-only")))
}

func TestCommitApprovedByPolicy(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	withPolicy(t, dir, `{"commit": {"require_approval": true}}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	before := headCount(t, dir)
	record, err := Commit(dir, Options{Message: "feat: change", Approve: true})
	require.NoError(t, err)
	require.NotNil(t, record.Policy)
	require.True(t, record.Policy.ApprovalRequired)
	require.True(t, record.Policy.Approved)
	require.Equal(t, strings.TrimSpace(run(t, dir, "git", "rev-parse", "HEAD")), record.Hash)
	require.Equal(t, "3", headCount(t, dir))
	require.Equal(t, "2", before)
	require.Empty(t, strings.TrimSpace(run(t, dir, "git", "status", "--porcelain")))
}

func TestPolicyProtectsOnlyListedBranches(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	withPolicy(t, dir, `{"commit": {"protected_branches": ["main"]}}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	_, err := Commit(dir, Options{Message: "edit on main"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "protects branch \"main\"")

	record, err := Commit(dir, Options{Message: "edit on a branch", Branch: "feat/approved"})
	require.NoError(t, err)
	require.Equal(t, "feat/approved", record.Branch)
	require.True(t, record.BranchCreated)
}

func TestUnreadablePolicyIsAnErrorNotABypass(t *testing.T) {
	t.Parallel()

	dir := newRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "crush.json"), []byte("{not json"), 0o644))
	run(t, dir, "git", "add", "crush.json")
	run(t, dir, "git", "commit", "-q", "-m", "chore: broken config")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0o644))

	_, err := Commit(dir, Options{Message: "feat: change"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "commit policy")
	require.Equal(t, "2", headCount(t, dir))
}
