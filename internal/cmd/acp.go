package cmd

import (
	"github.com/charmbracelet/crush/internal/acp"
	"github.com/charmbracelet/crush/internal/event"
	"github.com/charmbracelet/crush/internal/workspace"
	"github.com/spf13/cobra"
)

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Start Star as an ACP server",
	Long: `Start Star in Agent-Client Protocol mode.

This allows external ACP clients (web, desktop, mobile) to drive Star
over stdio using JSON-RPC. The client sends prompts and receives
streaming updates about agent activity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		appInstance := ws.(*workspace.AppWorkspace).App()

		if shouldEnableMetrics(appInstance.Config()) {
			event.Init()
		}

		event.AppInitialized()
		defer event.AppExited()

		server := acp.NewServer(cmd.Context())
		defer server.Shutdown()

		agent := acp.NewAgent(appInstance)
		return server.Run(agent)
	},
}
