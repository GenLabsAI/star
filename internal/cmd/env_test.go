package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newEnvTestCommand(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.Flags().Bool("json", false, "")
	return cmd, &out, &errOut
}

func TestEnvJSONOutput(t *testing.T) {
	cmd, buf, errOut := newEnvTestCommand(t)
	require.NoError(t, cmd.Flags().Set("json", "true"))

	require.NoError(t, envCmd.RunE(cmd, nil))

	// The manifest is what a pipe reads, so nothing may land on stderr.
	require.Empty(t, errOut.String())

	var manifest envreport.Manifest
	require.NoError(t, json.Unmarshal(buf.Bytes(), &manifest))
	require.NotEmpty(t, manifest.Workspace.Dir)
	require.NotEmpty(t, manifest.Workspace.ConfigDir)
	require.NotEmpty(t, manifest.Star.OS)
	require.NotEmpty(t, manifest.Runtimes)
}

func TestEnvHumanOutput(t *testing.T) {
	cmd, buf, _ := newEnvTestCommand(t)

	printEnvManifest(cmd, envreport.Manifest{
		Workspace: envreport.Workspace{Dir: "/work", ConfigDir: "/work/cfg", StateDir: "/work/state"},
		Star:      envreport.Star{Version: "v0.1.0", OS: "linux", Arch: "amd64"},
		Git:       envreport.Git{Repository: true, Root: "/work", Branch: "main"},
		Project: envreport.Project{
			Stacks: []envreport.Stack{{
				Name:      "Go",
				Manifests: []string{"/work/go.mod"},
				Commands:  []envreport.Command{{Kind: "test", Command: "go test ./..."}},
			}},
			ContextFiles: []string{"/work/AGENTS.md"},
		},
		Runtimes: []envreport.Runtime{{Name: "git", Found: true, Path: "/usr/bin/git"}},
		Cloud:    envreport.Cloud{APIBaseURLSet: true},
	})

	out := buf.String()
	require.Contains(t, out, "Workspace")
	require.Contains(t, out, "go test ./...")
	require.Contains(t, out, "main")
	require.Contains(t, out, "AGENTS.md")
	require.Contains(t, out, "set")
	require.Contains(t, out, "not set")
}

func TestEnvHumanOutputWithoutRepository(t *testing.T) {
	cmd, buf, _ := newEnvTestCommand(t)

	printEnvManifest(cmd, envreport.Manifest{Workspace: envreport.Workspace{Dir: "/work"}})

	out := buf.String()
	require.Contains(t, out, "not a git repository")
	require.Contains(t, out, "no known project stack detected")
	require.Contains(t, out, "no .agents/setup or .agents/resume hooks")
}
