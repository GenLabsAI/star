// Policy is the project's commit approval gate, read from the same project
// config file the rest of the CLI uses (crush.json / .crush.json).
//
// The point is to make "who may land a commit" a property of the project
// rather than a habit: a policy can require approval for every commit, or
// only for branches it protects, and a refused commit never touches the
// working tree, the index or the history.
package commit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/crush/internal/config"
)

// Policy is the effective commit policy for a project. Missing values mean
// "no gate": a project without a policy commits exactly as before.
type Policy struct {
	// RequireApproval demands --approve for every commit in the project.
	RequireApproval bool `json:"require_approval"`
	// ProtectedBranches demands --approve for commits landing on any of
	// these branches, whether or not RequireApproval is set.
	ProtectedBranches []string `json:"protected_branches,omitempty"`
	// Approver names who has to approve, for the refusal message. Nothing
	// verifies the identity; it documents the expectation.
	Approver string `json:"approver,omitempty"`
	// Path is the config file the policy was read from.
	Path string `json:"path,omitempty"`
}

// RequiresApproval reports whether a commit on branch needs --approve.
func (p Policy) RequiresApproval(branch string) bool {
	if p.RequireApproval {
		return true
	}
	return branch != "" && slices.Contains(p.ProtectedBranches, branch)
}

// LoadPolicy reads the commit policy from the project's JSON config files,
// in the CLI's usual order (.crush.json beats crush.json). Policies merge:
// approval is required if any file requires it, and protected branches are
// the union. A config file that cannot be parsed is an error, never a
// silent bypass of the gate.
func LoadPolicy(root string) (Policy, error) {
	var policy Policy
	for _, path := range config.ProjectConfigs(root) {
		if !strings.HasSuffix(path, ".json") {
			continue
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			// Global and out-of-project configs do not gate this project.
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var wrapper struct {
			Commit *Policy `json:"commit"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return Policy{}, fmt.Errorf("Failed to read the commit policy in %s: %w", path, err)
		}
		if wrapper.Commit == nil {
			continue
		}
		policy.RequireApproval = policy.RequireApproval || wrapper.Commit.RequireApproval
		for _, branch := range wrapper.Commit.ProtectedBranches {
			if !slices.Contains(policy.ProtectedBranches, branch) {
				policy.ProtectedBranches = append(policy.ProtectedBranches, branch)
			}
		}
		if policy.Approver == "" {
			policy.Approver = wrapper.Commit.Approver
		}
		if policy.Path == "" {
			policy.Path = path
		}
	}
	return policy, nil
}

// ApprovalRequiredError is returned when the project's policy refuses a
// commit that arrived without approval. Nothing has been staged or
// committed when it is returned.
type ApprovalRequiredError struct {
	Policy Policy
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
			"Refusing to commit: %s requires approval%s for every commit (re-run with --approve once it is approved)",
			where, approver,
		)
	}
	return fmt.Sprintf(
		"Refusing to commit: %s protects branch %q%s (re-run with --approve once it is approved)",
		where, e.Branch, approver,
	)
}
