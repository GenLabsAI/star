package cmd

import (
	"encoding/json"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/crush/internal/ports"
	"github.com/spf13/cobra"
)

var (
	portsJSON bool
	portsAll  bool
)

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Inventory listening TCP sockets and their processes",
	Long: `List the TCP servers running on this machine with the process and working
directory behind each one. By default only sockets whose process runs
inside the current workspace are shown, which is what an agent needs to
find the server it just started.

Use --all for the whole machine. Where the platform cannot report a
process directory (Windows), every listener is shown.`,
	Example: `
# Servers started from this project
star ports

# Everything listening on the machine
star ports --all

# Machine-readable inventory
star ports --json
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra prints to stderr by default; the inventory is the product.
		cmd.SetOut(cmd.OutOrStdout())

		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		listeners, err := ports.List(cmd.Context())
		if err != nil {
			return err
		}
		sort.Slice(listeners, func(i, j int) bool {
			if listeners[i].Port != listeners[j].Port {
				return listeners[i].Port < listeners[j].Port
			}
			return listeners[i].PID < listeners[j].PID
		})

		filtered := false
		note := ""
		if !portsAll {
			if ports.SupportsCWD() {
				listeners = ports.Within(listeners, dir)
				filtered = true
			} else {
				note = "This platform does not report process directories, so every listener is shown."
			}
		}

		if portsJSON {
			payload := struct {
				Workspace string           `json:"workspace"`
				Filtered  bool             `json:"filtered"`
				Listeners []ports.Listener `json:"listeners"`
			}{Workspace: dir, Filtered: filtered, Listeners: listeners}
			if payload.Listeners == nil {
				payload.Listeners = []ports.Listener{}
			}
			data, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}

		if len(listeners) == 0 {
			if filtered {
				cmd.Printf("No listening sockets from processes inside %s.\n", dir)
			} else {
				cmd.Println("No listening sockets found.")
			}
			if note != "" {
				cmd.Println(note)
			}
			return nil
		}

		cmd.Printf("%-7s %-8s %-16s %-18s %s\n", "PORT", "PID", "PROCESS", "ADDRESS", "CWD")
		for _, listener := range listeners {
			cwd := listener.CWD
			if cwd == "" {
				cwd = "-"
			}
			process := listener.Process
			if process == "" {
				process = "-"
			}
			cmd.Printf("%-7d %-8d %-16s %-18s %s\n", listener.Port, listener.PID, process, listener.Address, shortenCWD(dir, cwd))
		}
		if note != "" {
			cmd.Println()
			cmd.Println(note)
		}
		return nil
	},
}

func init() {
	portsCmd.Flags().BoolVar(&portsJSON, "json", false, "Output the inventory as JSON")
	portsCmd.Flags().BoolVar(&portsAll, "all", false, "Show listeners from every process, not just this workspace")
}

// shortenCWD renders a process directory relative to the workspace.
func shortenCWD(workspace, path string) string {
	if path == "" || path == "-" {
		return path
	}
	if rel, err := filepath.Rel(workspace, path); err == nil {
		if rel == "." {
			return "."
		}
		return rel
	}
	return path
}
