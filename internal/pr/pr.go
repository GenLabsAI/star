// Package pr turns a pushed branch into a pull request, or explains why it
// cannot.
//
// A pull request is a change landing in front of a reviewer, so the package
// keeps the same two habits as the commit gate: it refuses to guess (there
// is no title by default, and a branch that is not on the remote is not
// invented there), and it honours the project's approval policy before
// anything leaves the machine.
package pr

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/crush/internal/commit"
)

// Options describes the pull request to open.
type Options struct {
	// Title is the pull-request title. It is required.
	Title string
	// Body is the pull-request body. Empty is allowed.
	Body string
	// Branch is the branch the pull request comes from. Empty uses the
	// checked-out branch.
	Branch string
	// Base is the branch it lands on. Empty asks the remote for its
	// default branch.
	Base string
	// Draft opens the pull request as a draft.
	Draft bool
	// Push publishes the branch before opening the pull request. Without
	// it the branch must already be on the remote and match the local one.
	Push bool
	// Approve satisfies a project approval policy that would otherwise
	// refuse the pull request.
	Approve bool
}

// Record is what actually happened: which repository, which branches, and
// the pull request itself.
type Record struct {
	Repository bool                 `json:"repository"`
	Root       string               `json:"root"`
	Provider   string               `json:"provider"`
	Remote     string               `json:"remote"`
	RemoteURL  string               `json:"remote_url"`
	Host       string               `json:"host"`
	Owner      string               `json:"owner"`
	Name       string               `json:"name"`
	Branch     string               `json:"branch"`
	Base       string               `json:"base"`
	Title      string               `json:"title"`
	Draft      bool                 `json:"draft"`
	Pushed     bool                 `json:"pushed"`
	Number     int                  `json:"number"`
	URL        string               `json:"url"`
	Policy     *commit.PolicyRecord `json:"policy,omitempty"`
}

// ApprovalRequiredError is returned when the project's policy refuses a
// pull request that arrived without approval. Nothing has been pushed and
// no pull request exists when it is returned.
type ApprovalRequiredError struct {
	Policy commit.Policy
	Branch string
}

func (e *ApprovalRequiredError) Error() string {
	where := e.Policy.Path
	if where == "" {
		where = "the project's commit policy"
	}
	approver := ""
	if e.Policy.Approver != "" {
		approver = fmt.Sprintf(" for %s", e.Policy.Approver)
	}
	if e.Policy.RequireApproval {
		return fmt.Sprintf(
			"Refusing to open a pull request: %s requires approval%s for every change (re-run with --approve once it is approved)",
			where, approver,
		)
	}
	return fmt.Sprintf(
		"Refusing to open a pull request: %s protects branch %q%s (re-run with --approve once it is approved)",
		where, e.Branch, approver,
	)
}

// Create opens a pull request for dir's branch. It never changes the
// working tree, and it only pushes when the caller asked for it.
func Create(dir string, opts Options) (*Record, error) {
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return nil, fmt.Errorf("A pull-request title is required (--title)")
	}

	root, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("%s is not a git repository: %w", dir, err)
	}
	root = strings.TrimSpace(root)
	record := &Record{Repository: true, Root: root, Title: title, Draft: opts.Draft}

	branch := strings.TrimSpace(opts.Branch)
	if branch == "" {
		current, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("Failed to read the current branch: %w", err)
		}
		branch = strings.TrimSpace(current)
	}
	if branch == "HEAD" || branch == "" {
		return nil, fmt.Errorf("A detached HEAD has no branch to open a pull request from (check out a branch first)")
	}
	record.Branch = branch

	policy, err := commit.LoadPolicy(root)
	if err != nil {
		return nil, err
	}
	if policy.RequiresApproval(branch) && !opts.Approve {
		return nil, &ApprovalRequiredError{Policy: policy, Branch: branch}
	}
	if policy.Path != "" {
		record.Policy = &commit.PolicyRecord{
			Path:              policy.Path,
			RequireApproval:   policy.RequireApproval,
			ProtectedBranches: policy.ProtectedBranches,
			Approver:          policy.Approver,
			ApprovalRequired:  policy.RequiresApproval(branch),
			Approved:          opts.Approve,
		}
	}

	remote, remoteURL, err := remoteOf(root)
	if err != nil {
		return nil, err
	}
	record.Remote = remote
	record.RemoteURL = remoteURL
	host, owner, name, err := parseRemote(remoteURL)
	if err != nil {
		return nil, err
	}
	record.Host, record.Owner, record.Name = host, owner, name
	if host != "github.com" {
		return nil, fmt.Errorf("Opening a pull request is implemented for GitHub (found %s in %s)", host, remoteURL)
	}

	base := strings.TrimSpace(opts.Base)
	if base == "" {
		base = defaultBase(root, remote)
	}
	if base == branch {
		return nil, fmt.Errorf("Branch %q cannot be opened against itself: choose another --base", branch)
	}
	record.Base = base

	// The branch has to exist on the remote exactly as it is locally;
	// otherwise the pull request would describe different code than the
	// one the tree holds.
	local, err := git(root, "rev-parse", branch)
	if err != nil {
		return nil, fmt.Errorf("Unknown branch %q: %w", branch, err)
	}
	pushed := false
	if remoteHash, ok := branchOnRemote(root, remote, branch); !ok || remoteHash != strings.TrimSpace(local) {
		if !opts.Push {
			return nil, fmt.Errorf("Branch %q is not pushed to %s (re-run with --push, or push it yourself)", branch, remote)
		}
		if _, err := git(root, "push", "-u", remote, branch); err != nil {
			return nil, fmt.Errorf("Failed to push %q to %s: %w", branch, remote, err)
		}
		pushed = true
	}
	record.Pushed = pushed

	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("Opening a pull request needs the GitHub CLI (gh) on PATH")
	}
	args := []string{
		"pr", "create",
		"--repo", owner + "/" + name,
		"--head", branch,
		"--base", base,
		"--title", title,
		"--body", opts.Body,
	}
	if opts.Draft {
		args = append(args, "--draft")
	}
	out, err := gh(root, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to open the pull request: %w", err)
	}
	record.URL = strings.TrimSpace(out)
	record.Provider = "gh"
	record.Number = numberFromURL(record.URL)
	if record.Number == 0 {
		return nil, fmt.Errorf("The pull request was created but its number could not be read from %q", record.URL)
	}
	return record, nil
}

// branchOnRemote returns the remote's commit for branch, when it has one.
func branchOnRemote(root, remote, branch string) (string, bool) {
	out, err := git(root, "rev-parse", "--verify", "--quiet", "refs/remotes/"+remote+"/"+branch)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(out), true
}

// remoteOf returns the remote and URL to publish to. origin is the
// convention; a repository with one differently named remote still works.
func remoteOf(root string) (string, string, error) {
	for _, name := range []string{"origin"} {
		if url, err := git(root, "remote", "get-url", name); err == nil {
			return name, strings.TrimSpace(url), nil
		}
	}
	out, err := git(root, "remote")
	if err != nil {
		return "", "", fmt.Errorf("Failed to read the repository's remotes: %w", err)
	}
	names := strings.Fields(out)
	if len(names) == 0 {
		return "", "", fmt.Errorf("This repository has no remote to open a pull request against")
	}
	url, err := git(root, "remote", "get-url", names[0])
	if err != nil {
		return "", "", fmt.Errorf("Failed to read remote %q: %w", names[0], err)
	}
	return names[0], strings.TrimSpace(url), nil
}

// parseRemote reads the host, owner and repository out of the remote URL
// forms git accepts.
func parseRemote(url string) (host, owner, name string, err error) {
	url = strings.TrimSpace(url)
	switch {
	case strings.Contains(url, "://"):
		_, rest, _ := strings.Cut(url, "://")
		if at := strings.LastIndex(rest, "@"); at >= 0 {
			rest = rest[at+1:]
		}
		host, rest, _ = strings.Cut(rest, "/")
		owner, name = splitRepository(rest)
	case strings.Contains(url, "@") && strings.Contains(url, ":"):
		// ssh form: user@host:owner/name.git
		at := strings.LastIndex(url, "@")
		rest := url[at+1:]
		host, rest, _ = strings.Cut(rest, ":")
		owner, name = splitRepository(rest)
	default:
		return "", "", "", fmt.Errorf("Unrecognised remote URL %q", url)
	}
	if host == "" || owner == "" || name == "" {
		return "", "", "", fmt.Errorf("Unrecognised remote URL %q", url)
	}
	return host, owner, name, nil
}

// splitRepository cuts owner/name out of a remote path, dropping .git and
// any deeper path segments.
func splitRepository(path string) (string, string) {
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}

// defaultBase asks the remote for its default branch, falling back to main.
func defaultBase(root, remote string) string {
	out, err := git(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/"+remote+"/HEAD")
	if err != nil {
		return "main"
	}
	return strings.TrimPrefix(strings.TrimSpace(out), remote+"/")
}

// numberFromURL reads the pull-request number out of its URL.
func numberFromURL(url string) int {
	_, tail, ok := strings.Cut(url, "/pull/")
	if !ok {
		return 0
	}
	number, err := strconv.Atoi(strings.Trim(tail, "/"))
	if err != nil {
		return 0
	}
	return number
}

// gh runs the GitHub CLI and returns its standard output.
func gh(dir string, args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return stdout.String(), nil
}

// git runs one git command in dir.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
