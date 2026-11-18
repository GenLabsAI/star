// Coverage measurement for the `--min-coverage` gate.
//
// Only a stack with a known, deterministic coverage command is measured.
// Anything else reports itself as unavailable with a reason: a gate that
// silently passes when it could not measure would be worse than no gate.
package verify

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/shell"
)

// Verdicts produced by the coverage gate.
const (
	// VerdictUnderCovered means the tests passed but coverage is below the
	// threshold the caller asked for.
	VerdictUnderCovered = "under-covered"
	// VerdictCoverageUnavailable means a coverage gate was requested but
	// no number could be measured, so the gate cannot be evaluated.
	VerdictCoverageUnavailable = "coverage-unavailable"
)

// Coverage is the measured statement coverage of a run and how it compares
// with the requested threshold.
type Coverage struct {
	// Threshold is the minimum the caller asked for, in percent.
	Threshold float64 `json:"threshold"`
	// Available reports whether a number was actually measured.
	Available bool `json:"available"`
	// Percent is the measured statement coverage, in percent.
	Percent float64 `json:"percent"`
	// Statements and Covered are the raw counts behind Percent.
	Statements int `json:"statements,omitempty"`
	Covered    int `json:"covered,omitempty"`
	// Command is the coverage command that produced the number.
	Command string `json:"command,omitempty"`
	// Profile is the coverage profile file, relative to the bundle.
	Profile string `json:"profile,omitempty"`
	// Passed reports whether a measured number met the threshold.
	Passed bool `json:"passed"`
	// Reason explains an unavailable measurement, and why.
	Reason string `json:"reason,omitempty"`
}

// coverageSpec is the command that measures coverage for a stack, plus the
// file it writes the profile to.
type coverageSpec struct {
	command string
	profile string
}

// PlanCoverage returns the coverage command for the detected stacks, or a
// reason when no stack has one. Go is the only stack with a coverage tool
// this package can run deterministically today; the rest say so out loud
// instead of pretending.
func PlanCoverage(stacks []envreport.Stack, bundle string) (coverageSpec, string) {
	for _, stack := range stacks {
		if stack.Name != "Go" {
			continue
		}
		profile := filepath.Join(bundle, "coverage.out")
		return coverageSpec{
			command: fmt.Sprintf("go test -coverprofile=%q ./...", profile),
			profile: profile,
		}, ""
	}
	names := make([]string, 0, len(stacks))
	for _, stack := range stacks {
		names = append(names, stack.Name)
	}
	if len(names) == 0 {
		return coverageSpec{}, "no stack was detected, so there is no coverage command to run"
	}
	return coverageSpec{}, fmt.Sprintf("no coverage provider is known for the detected stack(s): %s", strings.Join(names, ", "))
}

// measureCoverage runs the coverage command and reads the profile it
// wrote. Every failure to produce a number comes back as an unavailable
// Coverage with a reason, never as a zero percent.
func measureCoverage(ctx context.Context, dir string, spec coverageSpec, threshold float64, timeout time.Duration) Coverage {
	coverage := Coverage{Threshold: threshold, Command: spec.command, Profile: filepath.Base(spec.profile)}

	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	captured, err := shell.RunAndCapture(stepCtx, shell.RunOptions{
		Command: spec.command,
		Cwd:     dir,
	})
	if err != nil || captured.ExitCode != 0 {
		// A failing coverage run can still leave a usable profile, so read
		// it before giving up; the tests' own failure is reported by the
		// steps, not here.
		if _, statErr := os.Stat(spec.profile); statErr != nil {
			coverage.Reason = fmt.Sprintf("the coverage command could not run: %s", firstLine(captured.Output, err))
			return coverage
		}
	}

	data, err := os.ReadFile(spec.profile)
	if err != nil {
		coverage.Reason = fmt.Sprintf("the coverage command wrote no profile: %s", firstLine(captured.Output, err))
		return coverage
	}
	covered, total, err := parseCoverageProfile(data)
	if err != nil {
		coverage.Reason = err.Error()
		return coverage
	}
	if total == 0 {
		coverage.Reason = "the coverage profile contains no statements to measure"
		return coverage
	}
	coverage.Available = true
	coverage.Covered = covered
	coverage.Statements = total
	coverage.Percent = roundPercent(100 * float64(covered) / float64(total))
	coverage.Passed = coverage.Percent >= threshold
	return coverage
}

// parseCoverageProfile totals the statement counts in a Go coverage
// profile:
//
//	mode: set
//	example.com/p/sum.go:5.1,7.2 2 1
//
// The third field is the number of statements, the fourth the execution
// count. A file that appears once with count 0 is uncovered.
func parseCoverageProfile(data []byte) (covered, total int, err error) {
	seenMode := false
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "mode:") {
			seenMode = true
			continue
		}
		if !seenMode {
			return 0, 0, fmt.Errorf("the coverage profile does not start with a mode line")
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return 0, 0, fmt.Errorf("unreadable coverage profile line: %q", line)
		}
		statements, err := strconv.Atoi(fields[len(fields)-2])
		if err != nil {
			return 0, 0, fmt.Errorf("unreadable statement count in %q", line)
		}
		count, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return 0, 0, fmt.Errorf("unreadable execution count in %q", line)
		}
		total += statements
		if count > 0 {
			covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("Failed to read the coverage profile: %w", err)
	}
	return covered, total, nil
}

// gateCoverage downgrades a passing verdict when the coverage gate was not
// met or could not be measured. It never touches a failing or flaky run: a
// failing test is still reported as the failure it is.
func gateCoverage(verdict string, coverage *Coverage) string {
	if coverage == nil || verdict != VerdictPass {
		return verdict
	}
	switch {
	case !coverage.Available:
		return VerdictCoverageUnavailable
	case !coverage.Passed:
		return VerdictUnderCovered
	}
	return verdict
}

func roundPercent(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}

// firstLine picks something short and human out of a failed command's
// output, falling back to the error itself.
func firstLine(output string, err error) string {
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	if err != nil {
		return err.Error()
	}
	return "no output"
}
