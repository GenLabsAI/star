package envreport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
	return path
}

func TestDetectStacks(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		require.Empty(t, DetectStacks(t.TempDir()))
	})

	t.Run("go module", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, dir, "go.mod", "module example.com/x\n", 0o644)
		writeFile(t, dir, "go.sum", "", 0o644)

		stacks := DetectStacks(dir)
		require.Len(t, stacks, 1)
		require.Equal(t, "Go", stacks[0].Name)
		require.Equal(t, []string{"go"}, stacks[0].Runtimes)
		require.Equal(t, []string{
			filepath.Join(dir, "go.mod"),
			filepath.Join(dir, "go.sum"),
		}, stacks[0].Manifests)
		require.Equal(t, []Command{
			{Kind: "install", Command: "go mod download"},
			{Kind: "build", Command: "go build ./..."},
			{Kind: "test", Command: "go test ./..."},
		}, stacks[0].Commands)
	})

	t.Run("node with bun and scripts", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, dir, "package.json", `{"scripts":{"build":"vite build","test":"vitest","dev":"vite"}}`, 0o644)
		writeFile(t, dir, "bun.lock", "", 0o644)

		stacks := DetectStacks(dir)
		require.Len(t, stacks, 1)
		stack := stacks[0]
		require.Equal(t, "Node", stack.Name)
		require.Equal(t, []string{"node", "bun"}, stack.Runtimes)
		require.Equal(t, []Command{
			{Kind: "install", Command: "bun install"},
			{Kind: "build", Command: "bun run build"},
			{Kind: "test", Command: "bun test"},
			{Kind: "dev", Command: "bun run dev"},
		}, stack.Commands)
		require.Contains(t, stack.Manifests, filepath.Join(dir, "bun.lock"))
	})

	t.Run("node with npm lockfile runs ci", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, dir, "package.json", `{"scripts":{"start":"node server.js"}}`, 0o644)
		writeFile(t, dir, "package-lock.json", "{}", 0o644)

		stack := DetectStacks(dir)[0]
		require.Equal(t, []Command{
			{Kind: "install", Command: "npm ci"},
			{Kind: "dev", Command: "npm run start"},
		}, stack.Commands)
	})

	t.Run("python with uv", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, dir, "pyproject.toml", "[project]\nname = \"x\"\n", 0o644)
		writeFile(t, dir, "uv.lock", "", 0o644)

		stack := DetectStacks(dir)[0]
		require.Equal(t, "Python", stack.Name)
		require.Equal(t, []string{"python3", "uv"}, stack.Runtimes)
		require.Equal(t, "uv sync", stack.Commands[0].Command)
	})
}

func TestBuildManifest(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CRUSH_GLOBAL_CONFIG", filepath.Join(dir, "config"))
	t.Setenv("CRUSH_GLOBAL_DATA", filepath.Join(dir, "data"))

	writeFile(t, dir, "go.mod", "module example.com/x\n", 0o644)
	writeFile(t, dir, "AGENTS.md", "# Instructions\n", 0o644)
	setup := "#!/bin/sh\nmake deps\n"
	writeFile(t, dir, ".agents/setup", setup, 0o755)
	writeFile(t, dir, ".agents/resume", "#!/bin/sh\nmake ping\n", 0o644)
	writeFile(t, dir, ".agents/skills/deploy/SKILL.md", "---\nname: deploy\n---\n", 0o644)
	writeFile(t, dir, ".crush/logs/crush.log", "started\n", 0o644)

	m, err := Build(dir)
	require.NoError(t, err)
	require.Equal(t, dir, m.Workspace.Dir)
	require.Equal(t, filepath.Join(dir, "config"), m.Workspace.ConfigDir)
	require.Equal(t, filepath.Join(dir, "data"), m.Workspace.StateDir)
	require.NotEmpty(t, m.Star.OS)

	require.False(t, m.Git.Repository)
	require.Equal(t, []string{filepath.Join(dir, "AGENTS.md")}, m.Project.ContextFiles)
	require.Equal(t, []string{"deploy"}, m.Project.Skills)
	require.Len(t, m.Project.Stacks, 1)

	sum := sha256.Sum256([]byte(setup))
	require.Equal(t, filepath.Join(dir, ".agents", "setup"), m.Hooks.Setup)
	require.Equal(t, hex.EncodeToString(sum[:]), m.Hooks.SetupHash)
	require.Equal(t, filepath.Join(dir, ".agents", "resume"), m.Hooks.Resume)
	if runtime.GOOS != "windows" {
		require.True(t, m.Hooks.SetupExecutable)
	}

	require.Equal(t, filepath.Join(dir, ".crush", "logs"), m.Logs.Dir)
	require.Equal(t, []string{filepath.Join(dir, ".crush", "logs", "crush.log")}, m.Logs.Files)

	// The report is designed to be pasted into logs, so it must never
	// carry a credential value.
	data, err := json.Marshal(m)
	require.NoError(t, err)
	require.NotContains(t, string(data), "secret-value")
}

func TestBuildRejectsFile(t *testing.T) {
	t.Parallel()

	path := writeFile(t, t.TempDir(), "not-a-dir", "", 0o644)
	_, err := Build(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
}

func TestCloudPresenceNeverReportsValues(t *testing.T) {
	t.Setenv("DECA_AGENT_API_BASE_URL", "https://api.example.test")
	t.Setenv("DECA_API_KEY", "secret-value")
	t.Setenv("DECA_SESSION_ID", "")
	t.Setenv("DECA_AGENT_WORKER_TOKEN", "")
	t.Setenv("DECA_MODEL", "model")
	t.Setenv("DECA_MODEL_BASE_URL", "https://models.example.test")
	t.Setenv("DECA_MODEL_API_KEY", "secret-value")

	m, err := Build(t.TempDir())
	require.NoError(t, err)
	require.True(t, m.Cloud.APIBaseURLSet)
	require.True(t, m.Cloud.UserCredential)
	require.False(t, m.Cloud.WorkerToken)
	require.True(t, m.Cloud.ModelReady)

	data, err := json.Marshal(m)
	require.NoError(t, err)
	require.NotContains(t, string(data), "secret-value")
	require.NotContains(t, string(data), "api.example.test")
}

func TestMissingListsUnavailableRuntimes(t *testing.T) {
	t.Parallel()

	missing := Manifest{Runtimes: []Runtime{
		{Name: "git", Found: true},
		{Name: "bun", Found: false},
	}}.Missing()
	require.Equal(t, []string{"bun"}, missing)
}

func TestReadinessIncludesBaseTooling(t *testing.T) {
	t.Parallel()

	runtimes := readiness(nil)
	require.NotEmpty(t, runtimes)
	var names []string
	for _, rt := range runtimes {
		names = append(names, rt.Name)
	}
	require.Contains(t, names, "git")
	require.True(t, strings.Contains(strings.Join(names, ","), "sh"))
}
