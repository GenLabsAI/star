package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/crush/internal/overview"
	"github.com/spf13/cobra"
)

var overviewJSON bool

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "One view of every tracked workspace",
	Long: `Join the local picture: which workspaces Star knows about, the branch
each is on, how many paths changed, whether the workspace was verified
and how that went, and which cloud credentials are present.

Workspaces come from the same list as "star projects"; verification
results come from the bundles "star verify" writes.`,
	Example: `
# Human-readable overview
star overview

# Machine-readable overview
star overview --json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the report is the product.
		cmd.SetOut(cmd.OutOrStdout())

		report, err := overview.Build()
		if err != nil {
			return err
		}
		if overviewJSON {
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}
		printOverview(cmd, report)
		return nil
	},
}

func init() {
	overviewCmd.Flags().BoolVar(&overviewJSON, "json", false, "Output the overview as JSON")
}

func printOverview(cmd *cobra.Command, report overview.Report) {
	cmd.Printf("Workspaces (%d)\n", len(report.Projects))
	if len(report.Projects) == 0 {
		cmd.Println("  none tracked yet")
	}
	for _, entry := range report.Projects {
		branch := entry.Branch
		if branch == "" {
			branch = "-"
		}
		state := fmt.Sprintf("%d changed", entry.ChangedPaths)
		if entry.ChangedPaths == 0 {
			state = "clean"
		}
		verified := "never verified"
		if entry.Verification != nil {
			verified = fmt.Sprintf("%s %s", entry.Verification.Verdict, age(entry.Verification.FinishedAt))
		}
		if entry.Problem != "" {
			verified = entry.Problem
		}
		cmd.Printf("  %-10s %-12s %-26s %s\n", branch, state, verified, entry.Path)
	}

	cmd.Println()
	cmd.Println("Cloud")
	envLine(cmd, "api base url", presence(report.Cloud.APIBaseURLSet))
	envLine(cmd, "user credential", presence(report.Cloud.UserCredential))
	envLine(cmd, "worker token", presence(report.Cloud.WorkerToken))
	envLine(cmd, "model", presence(report.Cloud.ModelReady))

	cmd.Println()
	cmd.Println("Scheduled work: `star wakeup list`; details: `star env`.")
}

// age renders how long ago something finished, at a glance.
func age(t time.Time) string {
	if t.IsZero() {
		return "at an unknown time"
	}
	delta := time.Since(t)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	}
}
