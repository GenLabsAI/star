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
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit the working tree, optionally on a new branch",
	Long: `Turn the changes shown by ` + "`star changes`" + ` into a git commit.

Two gates stand between a working tree and the repository's history: the
commit needs an explicit --message, and files git does not track yet block
the commit until you approve them with --include-untracked. Nothing is
pushed; the commit is local, and the remote stays untouched.

With --branch the commit lands on that branch, which is created when it
does not exist yet, so work stays isolated until review.`,
	Example: `
# Commit reviewed edits to tracked files
star commit --message "fix: handle the empty config"

# Include new files, on a branch of its own
star commit -m "feat: add the importer" --branch feat/importer --include-untracked

# Machine-readable record
star commit -m "chore: tidy docs" --json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the record is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		record, err := commit.Commit(dir, commit.Options{
			Message:          commitMessage,
			Branch:           commitBranch,
			IncludeUntracked: commitIncludeUntracked,
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
}
