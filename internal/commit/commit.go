// Package commit turns a reviewed working tree into a real git commit,
// optionally on a fresh branch, without ever pushing.
//
// The review half of change management is `star changes`: it shows what a
// task touched. This package is the other half, and it keeps two gates in
// front of the repository's history: a commit needs an explicit message,
// and files git does not track yet are only included when the caller asks
// for them by name (IncludeUntracked). Nothing here contacts a remote.
package commit

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/crush/internal/changes"
)

// Options describes the commit to create.
type Options struct {
	// Message is the commit message. It is required.
	Message string
	// Branch, when set, is checked out (created if missing) before the
	// commit, so work lands on its own branch instead of the current one.
	Branch string
	// IncludeUntracked also stages files git does not track yet. Without
	// it they block the commit, so new files never slip in unreviewed.
	IncludeUntracked bool
	// Approve satisfies a project approval policy that would otherwise
	// refuse the commit. It is the only way past that gate.
	Approve bool
}

// Record is what the commit actually produced: the new hash, the branch
// it landed on and the paths it contains.
type Record struct {
	Repository    bool           `json:"repository"`
	Root          string         `json:"root"`
	Branch        string         `json:"branch"`
	BranchCreated bool           `json:"branch_created,omitempty"`
	Hash          string         `json:"hash"`
	Short         string         `json:"short"`
	Message       string         `json:"message"`
	Files         []changes.File `json:"files"`
	Stat          string         `json:"stat,omitempty"`
	Policy        *PolicyRecord  `json:"policy,omitempty"`
}

// PolicyRecord is the approval policy that applied to this commit.
type PolicyRecord struct {
	Path              string   `json:"path,omitempty"`
	RequireApproval   bool     `json:"require_approval"`
	ProtectedBranches []string `json:"protected_branches,omitempty"`
	Approver          string   `json:"approver,omitempty"`
	ApprovalRequired  bool     `json:"approval_required"`
	Approved          bool     `json:"approved"`
}

// Commit stages and commits the working tree of dir.
func Commit(dir string, opts Options) (*Record, error) {
	message := strings.TrimSpace(opts.Message)
	if message == "" {
		return nil, fmt.Errorf("A commit message is required (--message)")
	}

	root, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("%s is not a git repository: %w", dir, err)
	}
	root = strings.TrimSpace(root)
	record := &Record{Repository: true, Root: root, Message: message}

	summary, err := changes.Collect(root)
	if err != nil {
		return nil, err
	}
	if summary.Clean() {
		return nil, fmt.Errorf("Nothing to commit: the working tree is clean")
	}

	var untracked []string
	for _, file := range summary.Files {
		if strings.TrimSpace(file.Status) == "??" {
			untracked = append(untracked, file.Path)
		}
	}
	if len(untracked) > 0 && !opts.IncludeUntracked {
		return nil, fmt.Errorf(
			"Refusing to commit %d untracked path(s) without approval: %s (re-run with --include-untracked to add them)",
			len(untracked), strings.Join(untracked, ", "),
		)
	}

	// The approval gate runs before anything is staged, so a refusal leaves
	// the working tree exactly as it was.
	policy, err := LoadPolicy(root)
	if err != nil {
		return nil, err
	}
	target := summary.Branch
	if opts.Branch != "" {
		target = opts.Branch
	}
	if policy.RequiresApproval(target) && !opts.Approve {
		return nil, &ApprovalRequiredError{Policy: policy, Branch: target}
	}
	if policy.Path != "" {
		record.Policy = &PolicyRecord{
			Path:              policy.Path,
			RequireApproval:   policy.RequireApproval,
			ProtectedBranches: policy.ProtectedBranches,
			Approver:          policy.Approver,
			ApprovalRequired:  policy.RequiresApproval(target),
			Approved:          opts.Approve,
		}
	}

	if opts.Branch != "" && opts.Branch != summary.Branch {
		created, err := switchBranch(root, opts.Branch)
		if err != nil {
			return nil, err
		}
		record.BranchCreated = created
	}

	if opts.IncludeUntracked {
		if _, err := git(root, "add", "-A"); err != nil {
			return nil, fmt.Errorf("Failed to stage changes: %w", err)
		}
	} else if _, err := git(root, "add", "-u"); err != nil {
		return nil, fmt.Errorf("Failed to stage changes: %w", err)
	}

	if _, err := git(root, "commit", "-m", message); err != nil {
		return nil, fmt.Errorf("Failed to commit: %w", err)
	}

	hash, err := git(root, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("Commit succeeded but its hash could not be read: %w", err)
	}
	record.Hash = strings.TrimSpace(hash)
	if short, err := git(root, "rev-parse", "--short", "HEAD"); err == nil {
		record.Short = strings.TrimSpace(short)
	}
	if branch, err := git(root, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		record.Branch = strings.TrimSpace(branch)
	}
	if stat, err := git(root, "show", "--stat", "--format=", "HEAD"); err == nil {
		record.Stat = strings.TrimRight(stat, "\n")
	}

	record.Files = summary.Files
	if !opts.IncludeUntracked {
		kept := record.Files[:0]
		for _, file := range record.Files {
			if strings.TrimSpace(file.Status) != "??" {
				kept = append(kept, file)
			}
		}
		record.Files = kept
	}
	return record, nil
}

// switchBranch checks out name, creating it first when it does not exist.
// It reports whether the branch was created.
func switchBranch(root, name string) (bool, error) {
	if _, err := git(root, "rev-parse", "--verify", "--quiet", "refs/heads/"+name); err == nil {
		if _, err := git(root, "checkout", name); err != nil {
			return false, fmt.Errorf("Failed to switch to branch %s: %w", name, err)
		}
		return false, nil
	}
	if _, err := git(root, "checkout", "-b", name); err != nil {
		return false, fmt.Errorf("Failed to create branch %s: %w", name, err)
	}
	return true, nil
}

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
