package cmd

import (
	"fmt"

	"github.com/charmbracelet/crush/internal/pairing"
	"github.com/spf13/cobra"
)

var pairOut string

var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Mint a pairing token for a Star server",
	Long: `Print a fresh pairing token. A server started with the same token
(--pair-token, or STAR_SERVER_TOKEN) answers 401 to any peer that does not
present it, so a workspace exposed over TCP is only reachable by peers you
authorized.

Give the token to the server and to each peer through --pair-token /
--pair-file, or the STAR_SERVER_TOKEN / STAR_SERVER_TOKEN_FILE variables.
The Star client sends it automatically when it is set.

With --out FILE the token is also written with owner-only permissions, so
a server can be started later with --pair-file FILE.`,
	Example: `
# Mint a token and hand it to both sides
star pair

# Persist a token for this machine
star pair --out ~/.config/crush/pairing-token
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())

		token, err := pairing.NewToken()
		if err != nil {
			return err
		}
		if pairOut != "" {
			if err := pairing.Save(pairOut, token); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Wrote %s (mode 0600)\nStart the server with --pair-file %s\n", pairOut, pairOut)
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Pass this token to the server (--pair-token) and to each peer (STAR_SERVER_TOKEN). It is not stored anywhere.")
		}
		cmd.Println(token)
		return nil
	},
}

func init() {
	pairCmd.Flags().StringVar(&pairOut, "out", "", "Also write the token to this file with mode 0600")
}
