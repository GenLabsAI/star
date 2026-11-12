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

func init() {
	cloudCmd.AddCommand(cloudPingCmd)
	cloudCmd.AddCommand(cloudJobCmd)
}
