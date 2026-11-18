// Package verify runs a project's detected build and test commands and
// records the outcome as an evidence bundle: command logs, exit codes,
// durations, the environment manifest and the working-tree diff.
//
// The point is evidence, not optimism. A reviewer reads verify.json or
// summary.md and sees what actually ran, what it printed, and what the
// tree looked like, instead of trusting an agent's summary.
package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/changes"
	"github.com/charmbracelet/crush/internal/datadir"
	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/shell"
)

// Verdicts reported by a run.
const (
	// VerdictPass means every selected command exited zero.
	VerdictPass = "pass"
	// VerdictFail means at least one selected command failed or timed out.
	VerdictFail = "fail"
	// VerdictFlaky means a command passed and failed across repeated
	// runs, so its result cannot be trusted either way.
	VerdictFlaky = "flaky"
	// VerdictNoCommands means the project had nothing to run.
	VerdictNoCommands = "no-commands"
)

// Defaults for a run.
const (
	DefaultTimeout   = 5 * time.Minute
	defaultTailBytes = 4000
	logDirName       = "logs"
)

// Spec is one command verification should run.
type Spec struct {
	// Kind is the command's role: install, build or test.
	Kind string
	// Command is the shell command line.
	Command string
}

// Attempt is one execution of a step, kept only when a run is repeated
// often enough that the same command can disagree with itself.
type Attempt struct {
	ExitCode   int    `json:"exit_code"`
	DurationMS int64  `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out,omitempty"`
	Log        string `json:"log,omitempty"`
}

// Step records the outcome of one command.
type Step struct {
	Kind       string `json:"kind"`
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	DurationMS int64  `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out,omitempty"`
	// Log is the command's log file, relative to the bundle directory.
	Log string `json:"log,omitempty"`
	// OutputTail is the end of the command's output, so a reader does
	// not have to open the log to see the failure.
	OutputTail string `json:"output_tail,omitempty"`
	// Attempts lists every execution when the run was repeated.
	Attempts []Attempt `json:"attempts,omitempty"`
	// Flaky marks a command whose attempts disagreed with each other.
	Flaky bool `json:"flaky,omitempty"`
}

// Outcomes renders the attempt exit codes in order, e.g. "pass, fail,
// pass", so a reader can see the disagreement without opening the JSON.
func (s Step) Outcomes() string {
	if len(s.Attempts) < 2 {
		return ""
	}
	parts := make([]string, 0, len(s.Attempts))
	for _, attempt := range s.Attempts {
		switch {
		case attempt.TimedOut:
			parts = append(parts, "timeout")
		case attempt.ExitCode != 0:
			parts = append(parts, "fail")
		default:
			parts = append(parts, "pass")
		}
	}
	return strings.Join(parts, ", ")
}

// Result is the full record of a verification run. Its JSON form is the
// bundle's verify.json.
type Result struct {
	Verdict    string             `json:"verdict"`
	StartedAt  time.Time          `json:"started_at"`
	DurationMS int64              `json:"duration_ms"`
	Workspace  envreport.Manifest `json:"workspace"`
	Changes    changes.Summary    `json:"changes"`
	Steps      []Step             `json:"steps"`
	// Coverage is present only when a threshold was requested.
	Coverage *Coverage `json:"coverage,omitempty"`
	Bundle   string    `json:"bundle"`
}

// Options configures a run.
type Options struct {
	// Timeout bounds each command. Zero uses DefaultTimeout.
	Timeout time.Duration
	// OutDir is the bundle directory. Zero uses a timestamped directory
	// under the workspace's data directory.
	OutDir string
	// TailBytes bounds each step's OutputTail in the JSON record. Zero
	// uses a small default.
	TailBytes int
	// MinCoverage, when above zero, requires measured statement coverage
	// of at least this percent. A project whose coverage cannot be
	// measured reports that explicitly instead of passing the gate.
	MinCoverage float64
	// Repeat runs every command this many times so a test that passes
	// and fails across runs is reported as flaky instead of a pass. Zero
	// or one runs everything once. Repeating also sets GOFLAGS=-count=1,
	// because Go's test cache would otherwise answer a second run with
	// the first run's result and hide exactly what repeat mode looks for.
	Repeat int
}

// Failed reports whether the verdict is a failure. A flaky run fails:
// something that only sometimes works is not something to ship.
func (r Result) Failed() bool {
	return r.Verdict == VerdictFail || r.Verdict == VerdictFlaky ||
		r.Verdict == VerdictUnderCovered || r.Verdict == VerdictCoverageUnavailable
}

// FlakySteps returns the commands whose attempts disagreed.
func (r Result) FlakySteps() []Step {
	var flaky []Step
	for _, step := range r.Steps {
		if step.Flaky {
			flaky = append(flaky, step)
		}
	}
	return flaky
}

// FailedSteps counts the steps that failed or timed out.
func (r Result) FailedSteps() int {
	failed := 0
	for _, step := range r.Steps {
		if step.ExitCode != 0 || step.TimedOut {
			failed++
		}
	}
	return failed
}

// Plan turns detected stacks into the commands to run. Development and
// service commands are skipped: verification builds and tests, it does
// not start long-running things.
func Plan(stacks []envreport.Stack, includeInstall bool) []Spec {
	var specs []Spec
	for _, stack := range stacks {
		for _, command := range stack.Commands {
			switch command.Kind {
			case "install":
				if !includeInstall {
					continue
				}
			case "build", "test":
			default:
				continue
			}
			specs = append(specs, Spec{Kind: command.Kind, Command: command.Command})
		}
	}
	return specs
}

// Run executes specs in dir and writes the evidence bundle.
func Run(ctx context.Context, dir string, specs []Spec, opts Options) (*Result, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("Nothing to verify: no build or test commands were detected")
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	tailBytes := opts.TailBytes
	if tailBytes <= 0 {
		tailBytes = defaultTailBytes
	}

	bundle := opts.OutDir
	if bundle == "" {
		bundle = filepath.Join(datadir.Resolve(dir), "verify", time.Now().UTC().Format("20060102T150405Z"))
	}
	if err := os.MkdirAll(filepath.Join(bundle, logDirName), 0o755); err != nil {
		return nil, fmt.Errorf("Failed to create evidence bundle: %w", err)
	}

	manifest, err := envreport.Build(dir)
	if err != nil {
		return nil, err
	}
	// A change summary is supporting evidence; a repository that cannot
	// be read must not block the build and test evidence. The bundle
	// itself is not a change, even when it lands inside the tree.
	summary, _ := changes.CollectWith(dir, changes.Options{Ignore: []string{bundle}})

	started := time.Now()
	result := &Result{
		StartedAt: started,
		Workspace: manifest,
		Changes:   summary,
		Bundle:    bundle,
	}

	repeat := opts.Repeat
	if repeat < 1 {
		repeat = 1
	}
	steps := make([]Step, len(specs))
	outputs := make([][]attemptOutput, len(specs))
	for i, spec := range specs {
		steps[i] = Step{Kind: spec.Kind, Command: spec.Command}
	}
	env := repeatEnv(repeat)
	for round := 1; round <= repeat; round++ {
		for i, spec := range specs {
			attempt, output := runAttempt(ctx, dir, spec, timeout, bundle, env, i+1, round, repeat)
			steps[i].Attempts = append(steps[i].Attempts, attempt)
			outputs[i] = append(outputs[i], output)
		}
	}
	for i := range steps {
		finalize(&steps[i], outputs[i], tailBytes)
	}
	result.Steps = steps
	if opts.MinCoverage > 0 {
		coverage := Coverage{Threshold: opts.MinCoverage}
		if spec, reason := PlanCoverage(manifest.Project.Stacks, bundle); reason != "" {
			coverage.Reason = reason
		} else {
			coverage = measureCoverage(ctx, dir, spec, opts.MinCoverage, timeout)
		}
		result.Coverage = &coverage
	}
	result.DurationMS = time.Since(started).Milliseconds()
	result.Verdict = gateCoverage(verdictFor(result.Steps), result.Coverage)

	if err := writeBundle(result); err != nil {
		return nil, err
	}
	return result, nil
}

// attemptOutput keeps one attempt's captured output so a failure can be
// quoted even when a later attempt passed.
type attemptOutput struct {
	log    string
	output string
	failed bool
}

// finalize folds a command's attempts into the single step a reader sees:
// the last attempt's numbers, a flaky mark when they disagreed, and the
// output of the attempt that failed.
func finalize(step *Step, outputs []attemptOutput, tailBytes int) {
	if len(outputs) == 0 {
		return
	}
	last := outputs[len(outputs)-1]
	final := step.Attempts[len(step.Attempts)-1]
	step.ExitCode = final.ExitCode
	step.DurationMS = final.DurationMS
	step.TimedOut = final.TimedOut
	step.Log = last.log

	quoted := last
	for _, output := range outputs {
		if output.failed {
			quoted = output
		}
	}
	step.OutputTail = tail(quoted.output, tailBytes)

	outcomes := make(map[bool]struct{}, 2)
	for _, attempt := range step.Attempts {
		outcomes[attempt.ExitCode != 0 || attempt.TimedOut] = struct{}{}
	}
	step.Flaky = len(outcomes) > 1
}

// repeatEnv is the environment a repeated run executes in. With more than
// one round it adds -count=1 to GOFLAGS so the Go test cache cannot answer
// a repeat with a cached result; without that, a flaky Go test looks
// stable and the whole point of repeating is lost. Other stacks ignore the
// variable, and a single run inherits the environment untouched.
func repeatEnv(repeat int) []string {
	env := os.Environ()
	if repeat <= 1 {
		return env
	}
	value := "-count=1"
	if current := os.Getenv("GOFLAGS"); current != "" {
		value = current + " -count=1"
	}
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if strings.HasPrefix(entry, "GOFLAGS=") {
			continue
		}
		out = append(out, entry)
	}
	return append(out, "GOFLAGS="+value)
}

func runAttempt(ctx context.Context, dir string, spec Spec, timeout time.Duration, bundle string, env []string, index, round, repeat int) (Attempt, attemptOutput) {
	attempt := Attempt{}
	slug := strings.ReplaceAll(strings.ToLower(spec.Kind), " ", "-")
	logName := fmt.Sprintf("%02d-%s.log", index, slug)
	if repeat > 1 {
		logName = fmt.Sprintf("%02d-%s-run%d.log", index, slug, round)
	}
	attempt.Log = filepath.ToSlash(filepath.Join(logDirName, logName))

	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	captured, err := shell.RunAndCapture(stepCtx, shell.RunOptions{
		Command: spec.Command,
		Cwd:     dir,
		Env:     env,
	})
	attempt.DurationMS = time.Since(started).Milliseconds()
	attempt.TimedOut = stepCtx.Err() != nil

	output := captured.Output
	attempt.ExitCode = captured.ExitCode
	if err != nil {
		// Execution could not start at all (parse or setup failure). That
		// is a failed attempt with the reason kept as evidence.
		attempt.ExitCode = 1
		output = strings.TrimSpace(output + "\n" + err.Error())
	}
	if attempt.TimedOut {
		attempt.ExitCode = 1
		output = strings.TrimSpace(output + fmt.Sprintf("\nTimed out after %s", timeout))
	}

	writeLog(bundle, attempt.Log, spec.Kind, spec.Command, attempt, output)
	return attempt, attemptOutput{
		log:    attempt.Log,
		output: output,
		failed: attempt.ExitCode != 0 || attempt.TimedOut,
	}
}

// verdictFor reports the worst thing that happened. A command that fails
// every time is a failure; one that passes and fails across runs is
// flaky, which is its own verdict because a retry would have hidden it.
func verdictFor(steps []Step) string {
	if len(steps) == 0 {
		return VerdictNoCommands
	}
	flaky := false
	for _, step := range steps {
		if step.Flaky {
			flaky = true
			continue
		}
		if step.ExitCode != 0 || step.TimedOut {
			return VerdictFail
		}
	}
	if flaky {
		return VerdictFlaky
	}
	return VerdictPass
}

// writeLog persists a command attempt's full output next to the JSON
// record. A log that cannot be written is not worth failing the run over:
// the output tail is in verify.json either way.
func writeLog(bundle, log, kind, command string, attempt Attempt, output string) {
	var b strings.Builder
	fmt.Fprintf(&b, "$ %s\n", command)
	fmt.Fprintf(&b, "# kind: %s\n", kind)
	fmt.Fprintf(&b, "# exit: %d\n", attempt.ExitCode)
	fmt.Fprintf(&b, "# duration_ms: %d\n\n", attempt.DurationMS)
	b.WriteString(output)
	if !strings.HasSuffix(output, "\n") {
		b.WriteString("\n")
	}
	_ = os.WriteFile(filepath.Join(bundle, filepath.FromSlash(log)), []byte(b.String()), 0o644)
}

// tail keeps the end of output, where failures live, and marks what was
// dropped.
// runsOf reports how many times the commands were executed.
func runsOf(steps []Step) int {
	runs := 0
	for _, step := range steps {
		if len(step.Attempts) > runs {
			runs = len(step.Attempts)
		}
	}
	return runs
}

func tail(output string, limit int) string {
	if len(output) <= limit {
		return output
	}
	return "…[truncated]\n" + output[len(output)-limit:]
}

// Latest reads the newest verification bundle under a project's data
// directory. It returns (nil, nil) when the project has never been
// verified. Bundle directories are named with UTC timestamps, so the
// newest name is the newest run.
func Latest(dataDir string) (*Result, error) {
	root := filepath.Join(dataDir, "verify")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to read verification history: %w", err)
	}
	newest := ""
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, entry.Name(), "verify.json")); err != nil {
			continue
		}
		if entry.Name() > newest {
			newest = entry.Name()
		}
	}
	if newest == "" {
		return nil, nil
	}
	data, err := os.ReadFile(filepath.Join(root, newest, "verify.json"))
	if err != nil {
		return nil, fmt.Errorf("Failed to read verification record: %w", err)
	}
	var result Result
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("Failed to decode verification record: %w", err)
	}
	return &result, nil
}

func writeBundle(result *Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to encode verification record: %w", err)
	}
	if err := os.WriteFile(filepath.Join(result.Bundle, "verify.json"), data, 0o644); err != nil {
		return fmt.Errorf("Failed to write verification record: %w", err)
	}
	if patch, err := changes.Diff(result.Workspace.Workspace.Dir); err == nil && len(patch) > 0 {
		_ = os.WriteFile(filepath.Join(result.Bundle, "changes.patch"), patch, 0o644)
	}
	if err := os.WriteFile(filepath.Join(result.Bundle, "summary.md"), []byte(SummaryMarkdown(*result)), 0o644); err != nil {
		return fmt.Errorf("Failed to write summary: %w", err)
	}
	return nil
}

// SummaryMarkdown renders the human-readable half of the bundle.
func SummaryMarkdown(result Result) string {
	var b strings.Builder
	b.WriteString("# Verification summary\n\n")
	fmt.Fprintf(&b, "- Verdict: **%s**\n", result.Verdict)
	fmt.Fprintf(&b, "- Started: %s\n", result.StartedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- Duration: %s\n", time.Duration(result.DurationMS)*time.Millisecond)
	fmt.Fprintf(&b, "- Workspace: `%s`\n", result.Workspace.Workspace.Dir)
	fmt.Fprintf(&b, "- Star: %s (%s/%s)\n", result.Workspace.Star.Version, result.Workspace.Star.OS, result.Workspace.Star.Arch)
	if runs := runsOf(result.Steps); runs > 1 {
		fmt.Fprintf(&b, "- Runs per command: %d\n", runs)
	}
	if result.Workspace.Git.Repository {
		branch := result.Workspace.Git.Branch
		if branch == "" {
			branch = "detached HEAD"
		}
		fmt.Fprintf(&b, "- Branch: %s\n", branch)
	}
	b.WriteString("\n## Steps\n\n")
	b.WriteString("| step | command | result | duration |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, step := range result.Steps {
		outcome := "pass"
		switch {
		case step.Flaky:
			outcome = "flaky (" + step.Outcomes() + ")"
		case step.TimedOut:
			outcome = "timed out"
		case step.ExitCode != 0:
			outcome = fmt.Sprintf("fail (exit %d)", step.ExitCode)
		}
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", step.Kind, step.Command, outcome, time.Duration(step.DurationMS)*time.Millisecond)
	}
	if result.Coverage != nil {
		b.WriteString("\n## Coverage\n\n")
		fmt.Fprintf(&b, "- Required: %.1f%% of statements\n", result.Coverage.Threshold)
		if result.Coverage.Available {
			fmt.Fprintf(&b, "- Measured: %.1f%% (%d of %d statements)\n", result.Coverage.Percent, result.Coverage.Covered, result.Coverage.Statements)
			fmt.Fprintf(&b, "- Gate: %s\n", map[bool]string{true: "passed", false: "failed"}[result.Coverage.Passed])
			if result.Coverage.Command != "" {
				fmt.Fprintf(&b, "- Command: `%s`\n", result.Coverage.Command)
			}
			if result.Coverage.Profile != "" {
				fmt.Fprintf(&b, "- Profile: `%s`\n", result.Coverage.Profile)
			}
		} else {
			fmt.Fprintf(&b, "- Unavailable: %s\n", result.Coverage.Reason)
			b.WriteString("- Gate: **not evaluated** (the run is not a pass until coverage can be measured)\n")
		}
	}

	b.WriteString("\n## Changes\n\n")
	if !result.Changes.Repository {
		b.WriteString("Not a git repository.\n")
	} else if result.Changes.Clean() {
		b.WriteString("Working tree clean.\n")
	} else {
		fmt.Fprintf(&b, "%d changed path(s), diff of %d bytes.\n\n", len(result.Changes.Files), result.Changes.PatchBytes)
		for _, file := range result.Changes.Files {
			fmt.Fprintf(&b, "- `%s` %s\n", strings.TrimSpace(file.Status), file.Path)
		}
	}
	b.WriteString("\n## Logs\n\n")
	for _, step := range result.Steps {
		// A repeated run keeps every attempt, so a reviewer can compare
		// the pass with the failure instead of trusting the summary.
		if len(step.Attempts) > 1 {
			for i, attempt := range step.Attempts {
				if attempt.Log != "" {
					fmt.Fprintf(&b, "- [%s attempt %d](%s)\n", step.Kind, i+1, attempt.Log)
				}
			}
			continue
		}
		if step.Log != "" {
			fmt.Fprintf(&b, "- [%s](%s)\n", step.Log, step.Log)
		}
	}
	return b.String()
}
