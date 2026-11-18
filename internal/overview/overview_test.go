package overview

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/crush/internal/projects"
	"github.com/stretchr/testify/require"
)

func isolateProjects(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("CRUSH_GLOBAL_DATA", "")
}

func writeBundle(t *testing.T, dataDir, stamp, record string) {
	t.Helper()
	dir := filepath.Join(dataDir, "verify", stamp)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "verify.json"), []byte(record), 0o644))
}

func TestBuildJoinsTrackedWorkspace(t *testing.T) {
	isolateProjects(t)

	workspace := t.TempDir()
	dataDir := t.TempDir()
	require.NoError(t, projects.Register(workspace, dataDir))
	writeBundle(t, dataDir, "20260101T000000Z", `{
	  "verdict": "fail",
	  "started_at": "2026-01-01T00:00:00Z",
	  "duration_ms": 10,
	  "steps": [{"kind": "test", "command": "old", "exit_code": 1}],
	  "bundle": "/tmp/old"
	}`)
	writeBundle(t, dataDir, "20260202T000000Z", `{
	  "verdict": "pass",
	  "started_at": "2026-02-02T00:00:00Z",
	  "duration_ms": 20,
	  "steps": [
	    {"kind": "build", "command": "go build ./...", "exit_code": 0},
	    {"kind": "test", "command": "go test ./...", "exit_code": 0}
	  ],
	  "bundle": "/tmp/new"
	}`)

	report, err := Build()
	require.NoError(t, err)
	require.Len(t, report.Projects, 1)

	entry := report.Projects[0]
	require.Equal(t, workspace, entry.Path)
	require.Equal(t, dataDir, entry.DataDir)
	require.Empty(t, entry.Problem)
	require.Zero(t, entry.ChangedPaths)
	require.NotNil(t, entry.Verification, "expected the newest verification bundle")
	require.Equal(t, "pass", entry.Verification.Verdict)
	require.Equal(t, 2, entry.Verification.Steps)
	require.Zero(t, entry.Verification.FailedSteps)
	require.Equal(t, "/tmp/new", entry.Verification.Bundle)
}

func TestBuildReportsMissingWorkspace(t *testing.T) {
	isolateProjects(t)

	require.NoError(t, projects.Register(filepath.Join(t.TempDir(), "gone"), t.TempDir()))

	report, err := Build()
	require.NoError(t, err)
	require.Len(t, report.Projects, 1)
	require.Equal(t, "workspace directory is missing", report.Projects[0].Problem)
	require.Nil(t, report.Projects[0].Verification)
}

func TestBuildWithoutProjects(t *testing.T) {
	isolateProjects(t)

	report, err := Build()
	require.NoError(t, err)
	require.Empty(t, report.Projects)
	require.False(t, report.GeneratedAt.IsZero())
}
