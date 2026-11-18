// Package overview joins the pieces of Star's local state into one view:
// which workspaces are tracked, what changed in them, whether they were
// verified, and which cloud credentials are present.
package overview

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/crush/internal/changes"
	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/gitutil"
	"github.com/charmbracelet/crush/internal/projects"
	"github.com/charmbracelet/crush/internal/verify"
)

// Verification is the condensed result of a workspace's latest run.
type Verification struct {
	Verdict     string    `json:"verdict"`
	FinishedAt  time.Time `json:"finished_at"`
	DurationMS  int64     `json:"duration_ms"`
	Steps       int       `json:"steps"`
	FailedSteps int       `json:"failed_steps"`
	Bundle      string    `json:"bundle"`
}

// Entry is one tracked workspace.
type Entry struct {
	Path         string        `json:"path"`
	DataDir      string        `json:"data_dir"`
	LastAccessed time.Time     `json:"last_accessed"`
	Branch       string        `json:"branch,omitempty"`
	ChangedPaths int           `json:"changed_paths"`
	Verification *Verification `json:"verification,omitempty"`
	Problem      string        `json:"problem,omitempty"`
}

// Report is the whole overview.
type Report struct {
	GeneratedAt time.Time       `json:"generated_at"`
	Projects    []Entry         `json:"projects"`
	Cloud       envreport.Cloud `json:"cloud"`
}

// Build collects the overview of every tracked project. A workspace that
// has been moved or deleted is reported as a problem rather than an
// error, so one stale entry cannot hide the rest.
func Build() (Report, error) {
	tracked, err := projects.List()
	if err != nil {
		return Report{}, fmt.Errorf("Failed to read tracked projects: %w", err)
	}

	report := Report{
		GeneratedAt: time.Now().UTC(),
		Projects:    make([]Entry, 0, len(tracked)),
		Cloud:       envreport.CloudPresence(),
	}
	for _, project := range tracked {
		report.Projects = append(report.Projects, describe(project))
	}
	// Most recently used first: that is the workspace a user is most
	// likely to care about.
	sort.SliceStable(report.Projects, func(i, j int) bool {
		return report.Projects[i].LastAccessed.After(report.Projects[j].LastAccessed)
	})
	return report, nil
}

func describe(project projects.Project) Entry {
	entry := Entry{
		Path:         project.Path,
		DataDir:      project.DataDir,
		LastAccessed: project.LastAccessed,
	}
	info, err := os.Stat(project.Path)
	if err != nil || !info.IsDir() {
		entry.Problem = "workspace directory is missing"
		return entry
	}

	entry.Branch = gitutil.CurrentBranch(project.Path)
	if summary, err := changes.Collect(project.Path); err == nil {
		entry.ChangedPaths = len(summary.Files)
	}
	if result, err := verify.Latest(project.DataDir); err == nil && result != nil {
		entry.Verification = &Verification{
			Verdict:     result.Verdict,
			FinishedAt:  result.StartedAt.Add(time.Duration(result.DurationMS) * time.Millisecond),
			DurationMS:  result.DurationMS,
			Steps:       len(result.Steps),
			FailedSteps: result.FailedSteps(),
			Bundle:      result.Bundle,
		}
	}
	return entry
}
