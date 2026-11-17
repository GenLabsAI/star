package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/verify"
	"github.com/spf13/cobra"
)

var (
	verifyJSON    bool
	verifyInstall bool
	verifyTimeout string
	verifyOut     string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Build and test the project, and write an evidence bundle",
	Long: `Run the project's detected build and test commands and record what
actually happened: every command, its exit code, duration and output,
plus the environment manifest and the working-tree diff.

The bundle lands in the project's data directory (or --out) as
verify.json, summary.md, command logs, and changes.patch. Verification
fails, with a non-zero exit, when any command fails.`,
	Example: `
# Build and test, keeping an evidence bundle
star verify

# Also install dependencies first
star verify --install

# Machine-readable record on stdout
star verify --json

# Put the bundle somewhere a reviewer will find it
star verify --out artifacts/proof
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the record is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}

		var timeout time.Duration
		if verifyTimeout != "" {
			timeout, err = time.ParseDuration(verifyTimeout)
			if err != nil {
				return fmt.Errorf("Invalid --timeout: %w", err)
			}
		}

		specs := verify.Plan(envreport.DetectStacks(dir), verifyInstall)
		result, err := verify.Run(cmd.Context(), dir, specs, verify.Options{
			Timeout: timeout,
			OutDir:  verifyOut,
		})
		if err != nil {
			return err
		}

		if verifyJSON {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
		} else {
			printVerifySummary(cmd, result)
		}

		if result.Failed() {
			return fmt.Errorf("Verification failed: %d of %d steps failed (evidence: %s)", result.FailedSteps(), len(result.Steps), result.Bundle)
		}
		return nil
	},
}

func init() {
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output the verification record as JSON")
	verifyCmd.Flags().BoolVar(&verifyInstall, "install", false, "Run detected install commands first")
	verifyCmd.Flags().StringVar(&verifyTimeout, "timeout", "5m", "Per-command timeout, e.g. 90s or 10m")
	verifyCmd.Flags().StringVar(&verifyOut, "out", "", "Evidence bundle directory")
}

func printVerifySummary(cmd *cobra.Command, result *verify.Result) {
	cmd.Printf("Verification: %s\n", strings.ToUpper(result.Verdict))

	envLine(cmd, "workspace", result.Workspace.Workspace.Dir)
	if result.Workspace.Git.Repository && result.Workspace.Git.Branch != "" {
		envLine(cmd, "branch", result.Workspace.Git.Branch)
	}
	envLine(cmd, "duration", (time.Duration(result.DurationMS) * time.Millisecond).String())
	envLine(cmd, "bundle", result.Bundle)

	cmd.Println()
	cmd.Println("Steps")
	for _, step := range result.Steps {
		outcome := "pass"
		switch {
		case step.TimedOut:
			outcome = "timeout"
		case step.ExitCode != 0:
			outcome = fmt.Sprintf("fail(%d)", step.ExitCode)
		}
		cmd.Printf("  %-8s %-6s %-34s %8s\n", outcome, step.Kind, step.Command, (time.Duration(step.DurationMS) * time.Millisecond).String())
		if step.ExitCode != 0 || step.TimedOut {
			for _, line := range lastLines(step.OutputTail, 6) {
				cmd.Println("    " + line)
			}
		}
	}

	if result.Changes.Repository {
		cmd.Println()
		cmd.Println("Changes")
		switch {
		case result.Changes.Clean():
			cmd.Println("  working tree clean")
		default:
			cmd.Printf("  %d path(s), patch %d bytes\n", len(result.Changes.Files), result.Changes.PatchBytes)
		}
	}

	cmd.Println()
	cmd.Println("Evidence: " + result.Bundle + "/summary.md")
}

// lastLines returns the last n lines of output for a compact failure
// excerpt in the terminal.
func lastLines(output string, n int) []string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
