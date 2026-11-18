package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/charmbracelet/crush/internal/pr"
	"github.com/spf13/cobra"
)

var (
	prTitle   string
	prBody    string
	prBranch  string
	prBase    string
	prDraft   bool
	prPush    bool
	prApprove bool
	prJSON    bool
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Open a pull request for a branch",
	Long: `Open a GitHub pull request for the branch in front of a reviewer.

The branch has to be on the remote exactly as it is locally, so the pull
request describes the code the tree holds; --push publishes it first. The
command never touches the working tree, and it never opens a pull request
with a guessed title.

A project approval policy (the "commit" block in crush.json or .crush.json)
gates this too: a branch that needs approval is refused before anything is
pushed, and --approve satisfies it.`,
	Example: `
# Open a pull request for the current branch (already pushed)
star pr --title "fix: handle an empty config"

# Publish the branch first
star pr --push --title "feat: pairing scopes" --body "Closes #12"

# Target another base branch, as a draft
star pr --base release --draft --title "chore: backport the fix"
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the record is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}

		record, err := pr.Create(dir, pr.Options{
			Title:   prTitle,
			Body:    prBody,
			Branch:  prBranch,
			Base:    prBase,
			Draft:   prDraft,
			Push:    prPush,
			Approve: prApprove,
		})
		if err != nil {
			var needsApproval *pr.ApprovalRequiredError
			if errors.As(err, &needsApproval) {
				return fmt.Errorf("%w (policy: %s)", err, needsApproval.Policy.Path)
			}
			return err
		}

		if prJSON {
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}

		cmd.Printf("Pull request #%d\n", record.Number)
		envLine(cmd, "title", record.Title)
		envLine(cmd, "branch", record.Branch+" -> "+record.Base)
		envLine(cmd, "repository", record.Owner+"/"+record.Name+" ("+record.Remote+")")
		envLine(cmd, "url", record.URL)
		if record.Draft {
			envLine(cmd, "draft", "yes")
		}
		if record.Pushed {
			envLine(cmd, "pushed", record.Branch+" to "+record.Remote)
		}
		if record.Policy != nil {
			state := "not required"
			if record.Policy.ApprovalRequired {
				state = "required and approved"
			}
			envLine(cmd, "approval", state)
		}
		return nil
	},
}

func init() {
	prCmd.Flags().StringVar(&prTitle, "title", "", "Pull-request title (required)")
	prCmd.Flags().StringVar(&prBody, "body", "", "Pull-request body")
	prCmd.Flags().StringVar(&prBranch, "branch", "", "Branch to open the pull request from (default: the checked-out branch)")
	prCmd.Flags().StringVar(&prBase, "base", "", "Branch the pull request lands on (default: the remote's default branch)")
	prCmd.Flags().BoolVar(&prDraft, "draft", false, "Open the pull request as a draft")
	prCmd.Flags().BoolVar(&prPush, "push", false, "Push the branch to the remote first")
	prCmd.Flags().BoolVar(&prApprove, "approve", false, "Satisfy a project approval policy that would refuse the pull request")
	prCmd.Flags().BoolVar(&prJSON, "json", false, "Output the pull-request record as JSON")
}
