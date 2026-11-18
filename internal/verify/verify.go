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
	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/shell"
)

// Verdicts reported by a run.
const (
	// VerdictPass means every selected command exited zero.
	VerdictPass = "pass"
	// VerdictFail means at least one selected command failed or timed out.
	VerdictFail = "fail"
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
	Bundle     string             `json:"bundle"`
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
}

// Failed reports whether the verdict is a failure.
func (r Result) Failed() bool {
	return r.Verdict == VerdictFail
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
		bundle = filepath.Join(dir, ".crush", "verify", time.Now().UTC().Format("20060102T150405Z"))
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

	for i, spec := range specs {
		step := runStep(ctx, dir, spec, timeout, tailBytes, bundle, i+1)
		result.Steps = append(result.Steps, step)
	}
	result.DurationMS = time.Since(started).Milliseconds()
	result.Verdict = verdictFor(result.Steps)

	if err := writeBundle(result); err != nil {
		return nil, err
	}
	return result, nil
}

func runStep(ctx context.Context, dir string, spec Spec, timeout time.Duration, tailBytes int, bundle string, index int) Step {
	step := Step{Kind: spec.Kind, Command: spec.Command}
	slug := strings.ReplaceAll(strings.ToLower(spec.Kind), " ", "-")
	step.Log = filepath.ToSlash(filepath.Join(logDirName, fmt.Sprintf("%02d-%s.log", index, slug)))

	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	captured, err := shell.RunAndCapture(stepCtx, shell.RunOptions{
		Command: spec.Command,
		Cwd:     dir,
	})
	step.DurationMS = time.Since(started).Milliseconds()
	step.TimedOut = stepCtx.Err() != nil

	output := captured.Output
	step.ExitCode = captured.ExitCode
	if err != nil {
		// Execution could not start at all (parse or setup failure). That
		// is a failed step with the reason kept as evidence.
		step.ExitCode = 1
		output = strings.TrimSpace(output + "\n" + err.Error())
	}
	if step.TimedOut {
		step.ExitCode = 1
		output = strings.TrimSpace(output + fmt.Sprintf("\nTimed out after %s", timeout))
	}

	step.OutputTail = tail(output, tailBytes)
	writeLog(bundle, step, output)
	return step
}

func verdictFor(steps []Step) string {
	if len(steps) == 0 {
		return VerdictNoCommands
	}
	for _, step := range steps {
		if step.ExitCode != 0 || step.TimedOut {
			return VerdictFail
		}
	}
	return VerdictPass
}

// writeLog persists a command's full output next to the JSON record. A
// log that cannot be written is not worth failing the run over: the
// output tail is in verify.json either way.
func writeLog(bundle string, step Step, output string) {
	var b strings.Builder
	fmt.Fprintf(&b, "$ %s\n", step.Command)
	fmt.Fprintf(&b, "# kind: %s\n", step.Kind)
	fmt.Fprintf(&b, "# exit: %d\n", step.ExitCode)
	fmt.Fprintf(&b, "# duration_ms: %d\n\n", step.DurationMS)
	b.WriteString(output)
	if !strings.HasSuffix(output, "\n") {
		b.WriteString("\n")
	}
	_ = os.WriteFile(filepath.Join(bundle, filepath.FromSlash(step.Log)), []byte(b.String()), 0o644)
}

// tail keeps the end of output, where failures live, and marks what was
// dropped.
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
		if step.TimedOut {
			outcome = "timed out"
		} else if step.ExitCode != 0 {
			outcome = fmt.Sprintf("fail (exit %d)", step.ExitCode)
		}
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", step.Kind, step.Command, outcome, time.Duration(step.DurationMS)*time.Millisecond)
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
		if step.Log != "" {
			fmt.Fprintf(&b, "- [%s](%s)\n", step.Log, step.Log)
		}
	}
	return b.String()
}
