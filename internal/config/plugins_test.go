package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/crush/internal/plugins"
	"github.com/stretchr/testify/require"
)

func writeManifest(t *testing.T, dir string, manifest plugins.Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, plugins.ManifestFile), data, 0o644))
}

func TestMergePlugins(t *testing.T) {
	dir := t.TempDir()

	source := filepath.Join(dir, "source")
	manifest := plugins.Manifest{
		Name:    "test-plugin",
		Version: "1.0.0",
		MCP: map[string]any{
			"my-mcp": map[string]any{
				"command": "relative/path/mcp",
				"args":    []string{"run"},
				"type":    "stdio",
			},
		},
		LSP: map[string]any{
			"my-lsp": map[string]any{
				"command": "relative/path/lsp",
				"args":    []string{"start"},
			},
		},
		Agents: map[string]any{
			"my-agent": map[string]any{
				"description": "A test agent",
				"model":       "large",
			},
		},
	}
	writeManifest(t, source, manifest)

	pluginsDir := filepath.Join(dir, "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0o755))

	installed, err := plugins.Install(source, pluginsDir)
	require.NoError(t, err)
	require.Equal(t, "test-plugin", installed.Name)

	cfg := &Config{}
	cfg.mergePlugins(pluginsDir)

	require.Contains(t, cfg.MCP, "my-mcp")
	mcp := cfg.MCP["my-mcp"]
	require.Equal(t, filepath.Join(installed.Path, "relative/path/mcp"), mcp.Command)
	require.Equal(t, []string{"run"}, mcp.Args)
	require.Equal(t, MCPStdio, mcp.Type)

	require.Contains(t, cfg.LSP, "my-lsp")
	lsp := cfg.LSP["my-lsp"]
	require.Equal(t, filepath.Join(installed.Path, "relative/path/lsp"), lsp.Command)
	require.Equal(t, []string{"start"}, lsp.Args)

	require.Contains(t, cfg.Agents, "my-agent")
	agent := cfg.Agents["my-agent"]
	require.Equal(t, "A test agent", agent.Description)
	require.Equal(t, SelectedModelTypeLarge, agent.Model)
}
