package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/crush/internal/changes"
	"github.com/spf13/cobra"
)

var (
	changesJSON  bool
	changesPatch string
)

var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Summarize the working tree and write a reviewable patch",
	Long: `List every path a task touched, with the branch it is isolated on and
the size of the diff, and optionally write the tracked diff to a patch
file. Untracked files are listed but not part of a git diff, so a
reviewer sees them by name.

Nothing is committed or pushed: this is the review half of change
management.`,
	Example: `
# See what changed
star changes

# Save a patch for review
star changes --patch artifacts/task.patch

# Machine-readable summary
star changes --json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the summary is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		summary, err := changes.Collect(dir)
		if err != nil {
			return err
		}

		var patchNote string
		if changesPatch != "" {
			if !summary.Repository {
				return fmt.Errorf("Cannot write a patch: %s is not a git repository", dir)
			}
			patch, err := changes.Diff(dir)
			if err != nil {
				return err
			}
			if err := os.WriteFile(changesPatch, patch, 0o644); err != nil {
				return fmt.Errorf("Failed to write patch: %w", err)
			}
			patchNote = fmt.Sprintf("%d bytes written to %s", len(patch), changesPatch)
		}

		if changesJSON {
			data, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
		} else if summary.Repository {
			printChangesSummary(cmd, summary, patchNote)
		}

		if !summary.Repository {
			return fmt.Errorf("%s is not a git repository", dir)
		}
		return nil
	},
}

func init() {
	changesCmd.Flags().BoolVar(&changesJSON, "json", false, "Output the summary as JSON")
	changesCmd.Flags().StringVar(&changesPatch, "patch", "", "Write the tracked diff to this file")
}

func printChangesSummary(cmd *cobra.Command, summary changes.Summary, patchNote string) {
	envLine(cmd, "repository", summary.Root)
	branch := summary.Branch
	if branch == "" {
		branch = "detached HEAD"
	}
	if summary.Upstream != "" {
		branch = fmt.Sprintf("%s (%d ahead, %d behind %s)", branch, summary.Ahead, summary.Behind, summary.Upstream)
	}
	envLine(cmd, "branch", branch)

	if summary.Clean() {
		cmd.Println("Working tree clean.")
		return
	}

	envLine(cmd, "paths", fmt.Sprintf("%d changed", len(summary.Files)))
	for _, file := range summary.Files {
		path := file.Path
		if file.Original != "" {
			path = file.Original + " -> " + file.Path
		}
		cmd.Printf("  %-3s %s\n", strings.TrimSpace(file.Status), path)
	}
	envLine(cmd, "patch", fmt.Sprintf("%d bytes", summary.PatchBytes))
	if patchNote != "" {
		envLine(cmd, "written", patchNote)
	}
}
