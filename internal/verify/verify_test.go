package verify

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/stretchr/testify/require"
)

func stacksFixture() []envreport.Stack {
	return []envreport.Stack{
		{
			Name: "Go",
			Commands: []envreport.Command{
				{Kind: "install", Command: "go mod download"},
				{Kind: "build", Command: "go build ./..."},
				{Kind: "test", Command: "go test ./..."},
			},
		},
		{
			Name: "Docker Compose",
			Commands: []envreport.Command{
				{Kind: "services", Command: "docker compose up -d"},
			},
		},
	}
}

func TestPlanSkipsInstallByDefault(t *testing.T) {
	t.Parallel()

	require.Equal(t, []Spec{
		{Kind: "build", Command: "go build ./..."},
		{Kind: "test", Command: "go test ./..."},
	}, Plan(stacksFixture(), false))
}

func TestPlanIncludesInstallOnRequest(t *testing.T) {
	t.Parallel()

	require.Equal(t, []Spec{
		{Kind: "install", Command: "go mod download"},
		{Kind: "build", Command: "go build ./..."},
		{Kind: "test", Command: "go test ./..."},
	}, Plan(stacksFixture(), true))
}

func TestPlanSkipsDevelopmentAndServices(t *testing.T) {
	t.Parallel()

	stacks := []envreport.Stack{{
		Name: "Node",
		Commands: []envreport.Command{
			{Kind: "dev", Command: "npm run dev"},
			{Kind: "services", Command: "docker compose up -d"},
			{Kind: "test", Command: "npm test"},
		},
	}}
	require.Equal(t, []Spec{{Kind: "test", Command: "npm test"}}, Plan(stacks, true))
}

func TestRunWithoutCommandsFails(t *testing.T) {
	t.Parallel()

	_, err := Run(context.Background(), t.TempDir(), nil, Options{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Nothing to verify")
}

func TestRunRecordsPassingBundle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bundle := filepath.Join(t.TempDir(), "bundle")
	specs := []Spec{
		{Kind: "build", Command: "echo building"},
		{Kind: "test", Command: "echo testing"},
	}

	result, err := Run(context.Background(), dir, specs, Options{OutDir: bundle})
	require.NoError(t, err)
	require.Equal(t, VerdictPass, result.Verdict)
	require.False(t, result.Failed())
	require.Len(t, result.Steps, 2)
	require.Equal(t, 0, result.Steps[0].ExitCode)
	require.Contains(t, result.Steps[0].OutputTail, "building")
	require.Equal(t, bundle, result.Bundle)

	// The bundle is the evidence: the JSON record, the human summary and
	// one log per command.
	record, err := os.ReadFile(filepath.Join(bundle, "verify.json"))
	require.NoError(t, err)
	var decoded Result
	require.NoError(t, json.Unmarshal(record, &decoded))
	require.Equal(t, VerdictPass, decoded.Verdict)
	require.Len(t, decoded.Steps, 2)

	summary, err := os.ReadFile(filepath.Join(bundle, "summary.md"))
	require.NoError(t, err)
	require.Contains(t, string(summary), "# Verification summary")
	require.Contains(t, string(summary), "| build | `echo building` | pass |")

	for _, step := range result.Steps {
		log, err := os.ReadFile(filepath.Join(bundle, filepath.FromSlash(step.Log)))
		require.NoError(t, err)
		require.Contains(t, string(log), "$ "+step.Command)
		require.Contains(t, string(log), "exit: 0")
	}
}

func TestRunRecordsFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bundle := filepath.Join(t.TempDir(), "bundle")
	specs := []Spec{
		{Kind: "build", Command: "echo fine"},
		{Kind: "test", Command: "echo boom >&2; exit 3"},
	}

	result, err := Run(context.Background(), dir, specs, Options{OutDir: bundle})
	require.NoError(t, err)
	require.Equal(t, VerdictFail, result.Verdict)
	require.True(t, result.Failed())
	require.Equal(t, 1, result.FailedSteps())
	require.Equal(t, 3, result.Steps[1].ExitCode)
	require.Contains(t, result.Steps[1].OutputTail, "boom")

	summary, err := os.ReadFile(filepath.Join(bundle, "summary.md"))
	require.NoError(t, err)
	require.Contains(t, string(summary), "fail (exit 3)")
}

func TestRunRecordsTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	specs := []Spec{{Kind: "test", Command: "while true; do :; done"}}

	result, err := Run(context.Background(), dir, specs, Options{
		OutDir:  filepath.Join(t.TempDir(), "bundle"),
		Timeout: 150 * time.Millisecond,
	})
	require.NoError(t, err)
	require.Equal(t, VerdictFail, result.Verdict)
	require.True(t, result.Steps[0].TimedOut)
	require.Contains(t, result.Steps[0].OutputTail, "Timed out after")
}
