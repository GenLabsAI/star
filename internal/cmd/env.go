package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/tone"
	"github.com/charmbracelet/crush/internal/envreport"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show the agent environment manifest",
	Long: `Report what an agent needs to know before working in a project: where
Star keeps config and state, the repository, the detected project stacks
with their install, build and test commands, the orb lifecycle hooks,
runtime readiness, and which cloud credentials are present.

Credential values are never printed, only whether they are set.`,
	Example: `
# Show the environment manifest
star env

# Machine-readable manifest for agents and scripts
star env --json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra's Print/Println default to stderr. The manifest is the
		// command's product, so it belongs on stdout where pipes and
		// agents look for it.
		cmd.SetOut(cmd.OutOrStdout())

		jsonOutput, _ := cmd.Flags().GetBool("json")

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		manifest, err := envreport.Build(dir)
		if err != nil {
			return err
		}

		if jsonOutput {
			data, err := json.MarshalIndent(manifest, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}

		printEnvManifest(cmd, manifest)
		return nil
	},
}

// envWorkspaceDir resolves the directory to report on. --cwd is honored
// without changing into it, so the command stays read-only.
func envWorkspaceDir(cmd *cobra.Command) (string, error) {
	if cwd, _ := cmd.Flags().GetString("cwd"); cwd != "" {
		abs, err := filepath.Abs(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to resolve directory: %v", err)
		}
		return abs, nil
	}
	return os.Getwd()
}

func init() {
	envCmd.Flags().Bool("json", false, "Output the manifest as JSON")
}

func printEnvManifest(cmd *cobra.Command, m envreport.Manifest) {
	title := lipgloss.NewStyle().Bold(true).Foreground(tone.Charple)
	hint := lipgloss.NewStyle().Foreground(tone.Squid)

	section := func(name string) {
		cmd.Println(title.Render(name))
	}

	section("Workspace")
	envLine(cmd, "dir", m.Workspace.Dir)
	envLine(cmd, "config", m.Workspace.ConfigDir)
	envLine(cmd, "state", m.Workspace.StateDir)
	versionLine := m.Star.Version
	if m.Star.Commit != "" && m.Star.Commit != "unknown" {
		versionLine += " (" + m.Star.Commit + ")"
	}
	envLine(cmd, "star", fmt.Sprintf("%s %s/%s", versionLine, m.Star.OS, m.Star.Arch))

	section("Repository")
	if m.Git.Repository {
		envLine(cmd, "root", m.Git.Root)
		branch := m.Git.Branch
		if branch == "" {
			branch = "detached HEAD"
		}
		envLine(cmd, "branch", branch)
	} else {
		cmd.Println("  " + hint.Render("not a git repository"))
	}

	section("Project")
	if len(m.Project.Stacks) == 0 {
		cmd.Println("  " + hint.Render("no known project stack detected"))
	}
	for _, stack := range m.Project.Stacks {
		// Pad before styling: escape sequences would otherwise count
		// against the label width.
		label := title.Render(fmt.Sprintf("%-17s", stack.Name))
		cmd.Println("  " + label + " " + hint.Render(strings.Join(shortPaths(m.Workspace.Dir, stack.Manifests), ", ")))
		for _, command := range stack.Commands {
			envLine(cmd, command.Kind, command.Command)
		}
	}
	if len(m.Project.ContextFiles) > 0 {
		envLine(cmd, "context", strings.Join(shortPaths(m.Workspace.Dir, m.Project.ContextFiles), ", "))
	}
	if len(m.Project.Skills) > 0 {
		envLine(cmd, "skills", strings.Join(m.Project.Skills, ", "))
	}
	if len(m.Project.ConfigFiles) > 0 {
		envLine(cmd, "config", strings.Join(shortPaths(m.Workspace.Dir, m.Project.ConfigFiles), ", "))
	}

	section("Hooks")
	if m.Hooks.Setup == "" && m.Hooks.Resume == "" {
		cmd.Println("  " + hint.Render("no .agents/setup or .agents/resume hooks"))
	}
	if m.Hooks.Setup != "" {
		detail := "sha256 " + shortHash(m.Hooks.SetupHash)
		if !m.Hooks.SetupExecutable {
			detail += ", not executable"
		}
		envLine(cmd, "setup", m.Hooks.Setup+" ("+detail+")")
	}
	if m.Hooks.Resume != "" {
		envLine(cmd, "resume", m.Hooks.Resume)
	}

	section("Runtimes")
	for _, rt := range m.Runtimes {
		mark := "ok"
		value := rt.Path
		if !rt.Found {
			mark = "missing"
			value = "base tooling"
			if len(rt.RequiredBy) > 0 {
				value = "required by " + strings.Join(rt.RequiredBy, ", ")
			}
		}
		envLine(cmd, rt.Name, mark+" "+hint.Render(value))
	}

	if missing := m.Missing(); len(missing) > 0 {
		cmd.Println("  " + hint.Render("missing: "+strings.Join(missing, ", ")))
	}

	section("Cloud")
	envLine(cmd, "api base url", presence(m.Cloud.APIBaseURLSet))
	envLine(cmd, "user credential", presence(m.Cloud.UserCredential))
	envLine(cmd, "worker token", presence(m.Cloud.WorkerToken))
	envLine(cmd, "model", presence(m.Cloud.ModelReady))

	if m.Logs.Dir != "" {
		section("Logs")
		envLine(cmd, "dir", m.Logs.Dir)
		for _, file := range m.Logs.Files {
			envLine(cmd, "file", file)
		}
	}

	cmd.Println()
	cmd.Println(hint.Render("Credential values are never shown. Run `star env --json` for the machine-readable manifest."))
}

// envLine prints one aligned "label value" row inside a section.
func envLine(cmd *cobra.Command, label, value string) {
	cmd.Printf("  %-17s %s\n", label, value)
}

func presence(ok bool) string {
	if ok {
		return "set"
	}
	return "not set"
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}

// shortPaths renders manifest paths relative to the workspace when they
// live inside it, so the report stays readable.
func shortPaths(root string, paths []string) []string {
	short := make([]string, 0, len(paths))
	for _, path := range paths {
		if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
			short = append(short, rel)
			continue
		}
		short = append(short, path)
	}
	return short
}
