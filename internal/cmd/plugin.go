package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/crush/internal/plugins"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Star plugins",
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a plugin from a local directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := pluginDirectory()
		if err != nil {
			return err
		}
		installed, err := plugins.Install(args[0], directory)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed %s %s\n", installed.Name, installed.Version)
		return err
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		directory, err := pluginDirectory()
		if err != nil {
			return err
		}
		installed, err := plugins.List(directory)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(installed)
	},
}

func pluginMutationCommand(use, short string, enabled *bool) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			directory, err := pluginDirectory()
			if err != nil {
				return err
			}
			if enabled == nil {
				return plugins.Remove(args[0], directory)
			}
			return plugins.SetEnabled(args[0], directory, *enabled)
		},
	}
}

func pluginDirectory() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve configuration directory: %w", err)
	}
	return filepath.Join(configDir, "star", "plugins"), nil
}

func init() {
	enabled := true
	disabled := false
	pluginCmd.AddCommand(
		pluginInstallCmd,
		pluginListCmd,
		pluginMutationCommand("remove", "Remove an installed plugin", nil),
		pluginMutationCommand("enable", "Enable an installed plugin", &enabled),
		pluginMutationCommand("disable", "Disable an installed plugin", &disabled),
	)
}
