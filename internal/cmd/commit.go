package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/crush/internal/commit"
	"github.com/spf13/cobra"
)

var (
	commitMessage          string
	commitBranch           string
	commitJSON             bool
	commitIncludeUntracked bool
	commitApprove          bool
	commitShowPolicy       bool
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit the working tree, optionally on a new branch",
	Long: `Turn the changes shown by ` + "`star changes`" + ` into a git commit.

Two gates stand between a working tree and the repository's history: the
commit needs an explicit --message, and files git does not track yet block
the commit until you approve them with --include-untracked. Nothing is
pushed; the commit is local, and the remote stays untouched.

A third gate can come from the project itself: a "commit" block in the
project's crush.json or .crush.json can require approval for every commit
or for protected branches. Such a commit is refused, leaving the tree and
the history untouched, until it is re-run with --approve. "star commit
--policy" prints the effective policy.

With --branch the commit lands on that branch, which is created when it
does not exist yet, so work stays isolated until review.`,
	Example: `
# Commit reviewed edits to tracked files
star commit --message "fix: handle the empty config"

# Include new files, on a branch of its own
star commit -m "feat: add the importer" --branch feat/importer --include-untracked

# Machine-readable record
star commit -m "chore: tidy docs" --json

# See (and satisfy) a project's approval policy
star commit --policy
star commit -m "fix: handle the empty config" --approve
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the record is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		if commitShowPolicy {
			policy, err := commit.LoadPolicy(dir)
			if err != nil {
				return err
			}
			if commitJSON {
				data, err := json.MarshalIndent(policy, "", "  ")
				if err != nil {
					return err
				}
				cmd.Println(string(data))
				return nil
			}
			path := policy.Path
			if path == "" {
				path = "none (no project policy)"
			}
			envLine(cmd, "policy", path)
			envLine(cmd, "require approval", fmt.Sprintf("%t", policy.RequireApproval))
			protected := strings.Join(policy.ProtectedBranches, ", ")
			if protected == "" {
				protected = "none"
			}
			envLine(cmd, "protected branches", protected)
			if policy.Approver != "" {
				envLine(cmd, "approver", policy.Approver)
			}
			return nil
		}

		record, err := commit.Commit(dir, commit.Options{
			Message:          commitMessage,
			Branch:           commitBranch,
			IncludeUntracked: commitIncludeUntracked,
			Approve:          commitApprove,
		})
		if err != nil {
			return err
		}

		if commitJSON {
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}

		envLine(cmd, "repository", record.Root)
		branch := record.Branch
		if record.BranchCreated {
			branch += " (created)"
		}
		envLine(cmd, "branch", branch)
		envLine(cmd, "commit", record.Short)
		envLine(cmd, "message", record.Message)
		envLine(cmd, "paths", fmt.Sprintf("%d committed", len(record.Files)))
		if policy := record.Policy; policy != nil {
			approval := "not required"
			if policy.ApprovalRequired {
				approval = "approved"
			}
			envLine(cmd, "policy", fmt.Sprintf("%s (%s)", policy.Path, approval))
		}
		for _, file := range record.Files {
			path := file.Path
			if file.Original != "" {
				path = file.Original + " -> " + path
			}
			cmd.Printf("  %-3s %s\n", strings.TrimSpace(file.Status), path)
		}
		if record.Stat != "" {
			cmd.Println()
			cmd.Println(strings.TrimSpace(record.Stat))
		}
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (required)")
	commitCmd.Flags().StringVar(&commitBranch, "branch", "", "Commit on this branch, creating it when missing")
	commitCmd.Flags().BoolVar(&commitIncludeUntracked, "include-untracked", false, "Also commit files git does not track yet")
	commitCmd.Flags().BoolVar(&commitJSON, "json", false, "Output the commit record as JSON")
	commitCmd.Flags().BoolVar(&commitApprove, "approve", false, "Record that this commit was approved, satisfying a project policy")
	commitCmd.Flags().BoolVar(&commitShowPolicy, "policy", false, "Print the project's effective commit policy and exit")
}
