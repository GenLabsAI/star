// Package overview joins the pieces of Star's local state into one view:
// which workspaces are tracked, what changed in them, whether they were
// verified, how many conversations and schedules each one holds, and which
// previews are listening, next to which cloud credentials are present.
package overview

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/crush/internal/changes"
	"github.com/charmbracelet/crush/internal/db"
	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/charmbracelet/crush/internal/gitutil"
	"github.com/charmbracelet/crush/internal/ports"
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

// Preview is one running server that belongs to a workspace, addressed the
// way a person or a browser would reach it.
type Preview struct {
	Port    int    `json:"port"`
	URL     string `json:"url"`
	Process string `json:"process,omitempty"`
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
	// Sessions is how many conversations the workspace has.
	Sessions int `json:"sessions"`
	// LastSessionAt is when the newest conversation was last active.
	LastSessionAt *time.Time `json:"last_session_at,omitempty"`
	// Schedules is how many durable wakeups are armed in the workspace.
	Schedules int `json:"schedules"`
	// NextScheduleAt is the next timed wakeup, when one is armed.
	NextScheduleAt *time.Time `json:"next_schedule_at,omitempty"`
	// Previews are the listeners whose process runs in the workspace.
	Previews []Preview `json:"previews,omitempty"`
}

// Report is the whole overview.
type Report struct {
	GeneratedAt time.Time       `json:"generated_at"`
	Projects    []Entry         `json:"projects"`
	Cloud       envreport.Cloud `json:"cloud"`
	// Sessions, Schedules and Previews are the totals across projects.
	Sessions  int `json:"sessions"`
	Schedules int `json:"schedules"`
	Previews  int `json:"previews"`
}

// Build collects the overview of every tracked project. A workspace that
// has been moved or deleted is reported as a problem rather than an
// error, so one stale entry cannot hide the rest.
func Build() (Report, error) {
	tracked, err := projects.List()
	if err != nil {
		return Report{}, fmt.Errorf("Failed to read tracked projects: %w", err)
	}

	// One machine-wide inventory answers every workspace's previews.
	listeners, _ := ports.List(context.Background())

	report := Report{
		GeneratedAt: time.Now().UTC(),
		Projects:    make([]Entry, 0, len(tracked)),
		Cloud:       envreport.CloudPresence(),
	}
	for _, project := range tracked {
		entry := describe(project, listeners)
		report.Sessions += entry.Sessions
		report.Schedules += entry.Schedules
		report.Previews += len(entry.Previews)
		report.Projects = append(report.Projects, entry)
	}
	// Most recently used first: that is the workspace a user is most
	// likely to care about.
	sort.SliceStable(report.Projects, func(i, j int) bool {
		return report.Projects[i].LastAccessed.After(report.Projects[j].LastAccessed)
	})
	return report, nil
}

func describe(project projects.Project, listeners []ports.Listener) Entry {
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
	entry.Sessions, entry.LastSessionAt = sessionStats(project.DataDir)
	entry.Schedules, entry.NextScheduleAt = scheduleStats(project.DataDir)
	entry.Previews = previewsOf(ports.Within(listeners, project.Path))
	return entry
}

// sessionStats counts a workspace's conversations by reading its database
// read-only. A workspace that was never opened has no database and zero
// conversations, which is not a problem worth reporting.
func sessionStats(dataDir string) (int, *time.Time) {
	path := filepath.Join(dataDir, "crush.db")
	if _, err := os.Stat(path); err != nil {
		return 0, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := db.ConnectReadOnly(ctx, path)
	if err != nil {
		return 0, nil
	}
	defer conn.Close()
	sessions, err := db.New(conn).ListSessions(ctx)
	if err != nil {
		return 0, nil
	}
	var newest *time.Time
	for _, session := range sessions {
		updated := time.Unix(session.UpdatedAt, 0).UTC()
		if newest == nil || updated.After(*newest) {
			newest = &updated
		}
	}
	return len(sessions), newest
}

// scheduleFile is the shape of the durable wakeup store on disk. Reading
// it directly keeps the overview free of the agent stack it only needs a
// count from.
type scheduleFile struct {
	Schedules []struct {
		NextFireUnix int64 `json:"next_fire_unix"`
	} `json:"schedules"`
}

// scheduleStats counts a workspace's armed wakeups and the next one due.
func scheduleStats(dataDir string) (int, *time.Time) {
	data, err := os.ReadFile(filepath.Join(dataDir, "wakeups", "schedules.json"))
	if err != nil {
		return 0, nil
	}
	var file scheduleFile
	if err := json.Unmarshal(data, &file); err != nil {
		return 0, nil
	}
	var next *time.Time
	for _, schedule := range file.Schedules {
		if schedule.NextFireUnix <= 0 {
			continue
		}
		at := time.Unix(schedule.NextFireUnix, 0).UTC()
		if next == nil || at.Before(*next) {
			next = &at
		}
	}
	return len(file.Schedules), next
}

// previewsOf turns the listeners inside a workspace into the addresses a
// browser would use. A server bound to every interface is still reached on
// loopback from this machine.
func previewsOf(listeners []ports.Listener) []Preview {
	previews := make([]Preview, 0, len(listeners))
	for _, listener := range listeners {
		host := listener.Address
		switch host {
		case "", "0.0.0.0", "::", "[::]", "*":
			host = "127.0.0.1"
		}
		previews = append(previews, Preview{
			Port:    listener.Port,
			URL:     fmt.Sprintf("http://%s:%d", host, listener.Port),
			Process: listener.Process,
		})
	}
	sort.SliceStable(previews, func(i, j int) bool {
		return previews[i].Port < previews[j].Port
	})
	return previews
}
