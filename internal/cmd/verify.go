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
	verifyJSON     bool
	verifyInstall  bool
	verifyTimeout  string
	verifyOut      string
	verifyRepeat   int
	verifyCoverage float64
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Build and test the project, and write an evidence bundle",
	Long: `Run the project's detected build and test commands and record what
actually happened: every command, its exit code, duration and output,
plus the environment manifest and the working-tree diff.

The bundle lands in the project's data directory (or --out) as
verify.json, summary.md, command logs, and changes.patch. Verification
fails, with a non-zero exit, when any command fails.

With --repeat N every command runs N times. A command that passes and
fails across those runs is reported as FLAKY and also exits non-zero,
because a result that only sometimes holds is not a pass.

With --min-coverage PERCENT the run also measures statement coverage and
fails when it is below the threshold, so green tests on thin coverage do
not pass the gate. Coverage is measured for Go projects today. When no
number can be measured the run exits non-zero and says why, because a
gate that silently passes when it could not measure is not a gate.`,
	Example: `
# Build and test, keeping an evidence bundle
star verify

# Also install dependencies first
star verify --install

# Machine-readable record on stdout
star verify --json

# Put the bundle somewhere a reviewer will find it
star verify --out artifacts/proof

# Run everything three times and catch flaky tests
star verify --repeat 3

# Require at least 80% statement coverage in a passing run
star verify --min-coverage 80
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the record is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}

		if verifyRepeat < 1 {
			return fmt.Errorf("Invalid --repeat: %d, want at least 1", verifyRepeat)
		}
		if verifyCoverage < 0 || verifyCoverage > 100 {
			return fmt.Errorf("Invalid --min-coverage: %g, want a percentage between 0 and 100", verifyCoverage)
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
			Timeout:     timeout,
			OutDir:      verifyOut,
			Repeat:      verifyRepeat,
			MinCoverage: verifyCoverage,
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

		if result.Verdict == verify.VerdictFlaky {
			return fmt.Errorf("Verification is flaky: %s passed and failed across runs (evidence: %s)", flakyList(result.FlakySteps()), result.Bundle)
		}
		// A coverage gate only ever downgrades a run that would otherwise
		// pass, so it is reported here only in those two verdicts; a failing
		// test is still reported as the failure it is.
		if coverage := result.Coverage; coverage != nil && (result.Verdict == verify.VerdictCoverageUnavailable || result.Verdict == verify.VerdictUnderCovered) {
			if !coverage.Available {
				return fmt.Errorf("Coverage gate cannot be evaluated: %s (evidence: %s)", coverage.Reason, result.Bundle)
			}
			if !coverage.Passed {
				return fmt.Errorf("Coverage gate failed: %.1f%% measured, %.1f%% required (%d of %d statements; evidence: %s)", coverage.Percent, coverage.Threshold, coverage.Covered, coverage.Statements, result.Bundle)
			}
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
	verifyCmd.Flags().IntVar(&verifyRepeat, "repeat", 1, "Run every command N times to detect flaky results")
	verifyCmd.Flags().Float64Var(&verifyCoverage, "min-coverage", 0, "Fail unless statement coverage is at least this percent")
}

func printVerifySummary(cmd *cobra.Command, result *verify.Result) {
	cmd.Printf("Verification: %s\n", strings.ToUpper(result.Verdict))

	envLine(cmd, "workspace", result.Workspace.Workspace.Dir)
	if result.Workspace.Git.Repository && result.Workspace.Git.Branch != "" {
		envLine(cmd, "branch", result.Workspace.Git.Branch)
	}
	envLine(cmd, "duration", (time.Duration(result.DurationMS) * time.Millisecond).String())
	if runs := len(result.Steps[0].Attempts); runs > 1 {
		envLine(cmd, "runs", fmt.Sprintf("%d per command", runs))
	}
	envLine(cmd, "bundle", result.Bundle)

	cmd.Println()
	cmd.Println("Steps")
	for _, step := range result.Steps {
		outcome := "pass"
		switch {
		case step.Flaky:
			outcome = "FLAKY"
		case step.TimedOut:
			outcome = "timeout"
		case step.ExitCode != 0:
			outcome = fmt.Sprintf("fail(%d)", step.ExitCode)
		}
		cmd.Printf("  %-8s %-6s %-34s %8s\n", outcome, step.Kind, step.Command, (time.Duration(step.DurationMS) * time.Millisecond).String())
		if outcomes := step.Outcomes(); outcomes != "" {
			cmd.Println("    runs: " + outcomes)
		}
		if step.ExitCode != 0 || step.TimedOut || step.Flaky {
			for _, line := range lastLines(step.OutputTail, 6) {
				cmd.Println("    " + line)
			}
		}
	}

	if coverage := result.Coverage; coverage != nil {
		cmd.Println()
		cmd.Println("Coverage")
		cmd.Printf("  required %.1f%%\n", coverage.Threshold)
		if coverage.Available {
			outcome := "pass"
			if !coverage.Passed {
				outcome = "FAIL"
			}
			cmd.Printf("  %-8s %.1f%% (%d of %d statements)\n", outcome, coverage.Percent, coverage.Covered, coverage.Statements)
			cmd.Printf("  command  %s\n", coverage.Command)
		} else {
			cmd.Printf("  unavailable: %s\n", coverage.Reason)
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

// flakyList names the commands whose attempts disagreed.
func flakyList(steps []verify.Step) string {
	names := make([]string, 0, len(steps))
	for _, step := range steps {
		names = append(names, step.Command)
	}
	return strings.Join(names, ", ")
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
