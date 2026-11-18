package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/crush/internal/pairing"
	"github.com/spf13/cobra"
)

var (
	pairOut         string
	pairScope       string
	pairApproval    bool
	pairApprovalTTL string
)

var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Mint a pairing token, or an approval code, for a Star server",
	Long: `Print a fresh pairing token. A server started with the same token
(--pair-token, or STAR_SERVER_TOKEN) answers 401 to any peer that does not
present it, so a workspace exposed over TCP is only reachable by peers you
authorized.

With --scope the token carries only those capabilities (read, write, exec),
so a peer that only needs to look cannot also run commands. The scopes
travel in the token itself; a server reads them and answers 403 to anything
outside them. A token without a scope prefix is unrestricted, which is what
tokens minted before scopes existed are.

With --approval the command mints a short-lived approval code instead. A
server started with --pair-approval-file FILE accepts a request that runs a
command only while a valid code is present in the X-Star-Approval header,
so a person on the server host can let one session through without handing
out permanent permission.

Give the token to the server and to each peer through --pair-token /
--pair-file, or the STAR_SERVER_TOKEN / STAR_SERVER_TOKEN_FILE variables.
The Star client sends it automatically when it is set.`,
	Example: `
# Mint a token and hand it to both sides
star pair

# Mint a read-only token for a peer that only needs to look
star pair --scope read --out ~/.config/crush/pairing-token

# Let one exec through, for the next five minutes
star pair --approval --out /run/user/1000/star-approval.json

# Persist a token for this machine
star pair --out ~/.config/crush/pairing-token
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())

		if pairApproval {
			ttl := pairing.DefaultApprovalTTL
			if pairApprovalTTL != "" {
				parsed, err := time.ParseDuration(pairApprovalTTL)
				if err != nil {
					return fmt.Errorf("Invalid --approval-ttl: %w", err)
				}
				if parsed <= 0 {
					return fmt.Errorf("Invalid --approval-ttl: %s, want a positive duration", parsed)
				}
				ttl = parsed
			}
			if pairOut == "" {
				return fmt.Errorf("An approval code needs somewhere to live: pass --out FILE")
			}
			approval, err := pairing.SaveApproval(pairOut, ttl)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(),
				"Approval valid until %s. Start the server with --pair-approval-file %s and present this code in the %s header.\n",
				approval.ExpiresAt.Format(time.RFC3339), pairOut, pairing.ApprovalHeader)
			cmd.Println(approval.Code)
			return nil
		}

		scopes, err := pairing.ParseScopes(pairScope)
		if err != nil {
			return err
		}
		token, err := pairing.Mint(scopes)
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
		if !scopes.All() {
			fmt.Fprintf(cmd.ErrOrStderr(), "Scopes: %s. Requests outside them are answered 403 by the paired server.\n", scopes.String())
		}
		cmd.Println(token)
		return nil
	},
}

func init() {
	pairCmd.Flags().StringVar(&pairOut, "out", "", "Also write the token, or the approval, to this file with mode 0600")
	pairCmd.Flags().StringVar(&pairScope, "scope", "", "Limit the token to these capabilities: read, write, exec (default all)")
	pairCmd.Flags().BoolVar(&pairApproval, "approval", false, "Mint a short-lived approval code instead of a token")
	pairCmd.Flags().StringVar(&pairApprovalTTL, "approval-ttl", "", "How long an approval code stays valid, e.g. 2m (default 5m)")
}
