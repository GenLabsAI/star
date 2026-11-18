package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/crush/internal/knowledge"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

var (
	knowledgeBody string
	knowledgeFile string
	knowledgeJSON bool
)

var knowledgeCmd = &cobra.Command{
	Use:   "knowledge",
	Short: "Durable project knowledge that outlives a session",
	Long: `Store discoveries, commands and quirks with the project so a later
session does not have to work them out again.

Entries are plain markdown files in the project's data directory
(.crush/knowledge by default), stored and returned verbatim. They survive
restarts, are readable by any editor, and can be committed with the
project if that is what you want.`,
	Example: `
# Record what you learned, then read it back in another process
star knowledge add ci-quirk --body "CI runs go test with -count=1 only."
star knowledge show ci-quirk

# From a file, or from stdin
star knowledge add flaky-test --file notes.txt
printf 'Local DB is 5433.' | star knowledge add db-port

# List everything, for agents
star knowledge list --json
  `,
}

var knowledgeAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Store an entry",
	Long: `Store an entry under a slug name. An existing entry with the same name is
replaced, so the newest answer wins.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())
		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		body, err := knowledgeInput(cmd)
		if err != nil {
			return err
		}
		store := knowledge.NewStore(dir)
		entry, err := store.Add(args[0], body)
		if err != nil {
			return err
		}
		cmd.Printf("Added %s (%d bytes)\n", entry.Name, entry.Size)
		return nil
	},
}

var knowledgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored entries",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())
		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		entries, err := knowledge.NewStore(dir).List()
		if err != nil {
			return err
		}
		if knowledgeJSON {
			if entries == nil {
				entries = []knowledge.Entry{}
			}
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
			return nil
		}
		if len(entries) == 0 {
			cmd.Println("No knowledge entries yet.")
			return nil
		}
		for _, entry := range entries {
			cmd.Printf("  %-24s %s\n", entry.Name, entry.Summary)
		}
		return nil
	},
}

var knowledgeShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Print an entry exactly as stored",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		_, body, err := knowledge.NewStore(dir).Get(args[0])
		if err != nil {
			return err
		}
		_, err = io.WriteString(cmd.OutOrStdout(), body)
		return err
	},
}

var knowledgeRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Delete an entry",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())
		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		if err := knowledge.NewStore(dir).Remove(args[0]); err != nil {
			return err
		}
		cmd.Printf("Removed %s\n", args[0])
		return nil
	},
}

var knowledgePathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the knowledge directory",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := envWorkspaceDir(cmd)
		if err != nil {
			return err
		}
		cmd.SetOut(cmd.OutOrStdout())
		cmd.Println(knowledge.NewStore(dir).Dir())
		return nil
	},
}

func init() {
	knowledgeAddCmd.Flags().StringVar(&knowledgeBody, "body", "", "Entry content")
	knowledgeAddCmd.Flags().StringVar(&knowledgeFile, "file", "", "Read content from a file (\"-\" for stdin)")
	knowledgeListCmd.Flags().BoolVar(&knowledgeJSON, "json", false, "Output the entries as JSON")
	knowledgeCmd.AddCommand(knowledgeAddCmd, knowledgeListCmd, knowledgeShowCmd, knowledgeRemoveCmd, knowledgePathCmd)
}

// knowledgeInput resolves the entry body from --body, --file or stdin.
func knowledgeInput(cmd *cobra.Command) (string, error) {
	switch {
	case strings.TrimSpace(knowledgeBody) != "":
		return knowledgeBody, nil
	case knowledgeFile == "-":
		data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), knowledge.MaxBody+1))
		if err != nil {
			return "", fmt.Errorf("Failed to read stdin: %w", err)
		}
		return string(data), nil
	case knowledgeFile != "":
		data, err := os.ReadFile(knowledgeFile)
		if err != nil {
			return "", fmt.Errorf("Failed to read %s: %w", knowledgeFile, err)
		}
		return string(data), nil
	}
	if term.IsTerminal(os.Stdin.Fd()) {
		return "", fmt.Errorf("Provide --body, --file or pipe the entry on stdin")
	}
	data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), knowledge.MaxBody+1))
	if err != nil {
		return "", fmt.Errorf("Failed to read stdin: %w", err)
	}
	return string(data), nil
}
