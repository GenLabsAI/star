package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/crush/internal/cloud"
	"github.com/spf13/cobra"
)

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Talk to the Deca Agent Jobs API (cloud proof)",
	Long: `Probe the Agent Jobs API that backs cloud agents.

Credentials come from the environment so secrets never land in files:
  DECA_AGENT_API_BASE_URL  API base, e.g. https://<genlabsapi-host>
  DECA_API_KEY             user key (Bearer) for user-side calls
  DECA_AGENT_WORKER_TOKEN  worker token (X-Worker-Token) for job calls
  DECA_AGENT_JOB_ID        default job id for job-scoped calls`,
}

var cloudPingCmd = &cobra.Command{
	Use:   "status",
	Short: "Check reachability and auth against the Agent Jobs API",
	RunE: func(cmd *cobra.Command, args []string) error {
		base := os.Getenv("DECA_AGENT_API_BASE_URL")
		if base == "" {
			return fmt.Errorf("DECA_AGENT_API_BASE_URL is required")
		}
		client := cloud.NewClient(base, os.Getenv("DECA_API_KEY"), os.Getenv("DECA_AGENT_WORKER_TOKEN"))
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		result, err := client.Ping(ctx)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cloudJobCmd = &cobra.Command{
	Use:   "job [job-id]",
	Short: "Fetch one job status as the harness worker would",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		base := os.Getenv("DECA_AGENT_API_BASE_URL")
		if base == "" {
			return fmt.Errorf("DECA_AGENT_API_BASE_URL is required")
		}
		jobID := ""
		if len(args) > 0 {
			jobID = args[0]
		} else {
			jobID = os.Getenv("DECA_AGENT_JOB_ID")
		}
		if jobID == "" {
			return fmt.Errorf("Provide a job id or set DECA_AGENT_JOB_ID")
		}
		client := cloud.NewClient(base, os.Getenv("DECA_API_KEY"), os.Getenv("DECA_AGENT_WORKER_TOKEN"))
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		result, err := client.JobStatus(ctx, jobID)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cloudSessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Create a cloud terminal session (prod relay protocol)",
	RunE: func(cmd *cobra.Command, args []string) error {
		base := os.Getenv("DECA_AGENT_API_BASE_URL")
		if base == "" {
			return fmt.Errorf("DECA_AGENT_API_BASE_URL is required")
		}
		client := cloud.NewClient(base, os.Getenv("DECA_API_KEY"), os.Getenv("DECA_AGENT_WORKER_TOKEN"))
		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()
		sessionID, relayURL, err := client.CreateTerminalSession(ctx)
		if err != nil {
			return err
		}
		fmt.Println(sessionID)
		fmt.Println(relayURL)
		return nil
	},
}

var cloudAttachCmd = &cobra.Command{
	Use:   "attach <relay-url>",
	Short: "Watch a cloud terminal session stream (TV screen)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		base := os.Getenv("DECA_AGENT_API_BASE_URL")
		if base == "" {
			return fmt.Errorf("DECA_AGENT_API_BASE_URL is required")
		}
		client := cloud.NewClient(base, os.Getenv("DECA_API_KEY"), os.Getenv("DECA_AGENT_WORKER_TOKEN"))
		return client.WatchTerminal(cmd.Context(), args[0], os.Stdout)
	},
}

var cloudACPModel = ""
var cloudACPBaseURL = ""
var cloudACPPrompt = ""
var cloudACPRelayURL = ""

var cloudACPCmd = &cobra.Command{
	Use:   "acp",
	Short: "Run a cloud ACP session: send model config, print the agent reply",
	RunE: func(cmd *cobra.Command, args []string) error {
		base := os.Getenv("DECA_AGENT_API_BASE_URL")
		if base == "" {
			return fmt.Errorf("DECA_AGENT_API_BASE_URL is required")
		}
		if cloudACPPrompt == "" {
			return fmt.Errorf("Provide --prompt")
		}
		apiKey := os.Getenv("NVIDIA_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("NVIDIA_API_KEY is required (env only, never committed)")
		}
		if cloudACPModel == "" {
			cloudACPModel = "mistralai/mistral-large-4-1t-a52b"
		}
		if cloudACPBaseURL == "" {
			cloudACPBaseURL = "https://integrate.api.nvidia.com/v1"
		}
		client := cloud.NewClient(base, os.Getenv("DECA_API_KEY"), os.Getenv("DECA_AGENT_WORKER_TOKEN"))
		relayURL := cloudACPRelayURL
		if relayURL == "" {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()
			sessionID, createdURL, err := client.CreateTerminalSession(ctx)
			if err != nil {
				return err
			}
			relayURL = createdURL
			fmt.Fprintf(os.Stderr, "session %s\nrelay %s\n", sessionID, relayURL)
		} else {
			fmt.Fprintf(os.Stderr, "relay %s\n", relayURL)
		}
		actx, acancel := context.WithTimeout(cmd.Context(), 180*time.Second)
		defer acancel()
		return client.RunACPSession(actx, relayURL, cloud.ACPModelConfig{
			Prompt:  cloudACPPrompt,
			Model:   cloudACPModel,
			BaseURL: cloudACPBaseURL,
			APIKey:  apiKey,
		}, os.Stdout)
	},
}

func init() {
	cloudACPCmd.Flags().StringVar(&cloudACPPrompt, "prompt", "", "Prompt to send the cloud agent")
	cloudACPCmd.Flags().StringVar(&cloudACPModel, "model", "", "Model id (default mistralai/mistral-large-4-1t-a52b)")
	cloudACPCmd.Flags().StringVar(&cloudACPBaseURL, "base-url", "", "Model base URL (default https://integrate.api.nvidia.com/v1)")
	cloudACPCmd.Flags().StringVar(&cloudACPRelayURL, "relay-url", "", "Reuse an existing relay browser URL instead of creating a session")
	cloudCmd.AddCommand(cloudPingCmd)
	cloudCmd.AddCommand(cloudJobCmd)
	cloudCmd.AddCommand(cloudSessionCmd)
	cloudCmd.AddCommand(cloudAttachCmd)
	cloudCmd.AddCommand(cloudACPCmd)
}
