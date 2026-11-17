package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/crush/internal/agent"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/workspace"
	"github.com/spf13/cobra"
)

var wakeupCmd = &cobra.Command{
	Use:   "wakeup",
	Short: "Manage durable agent wakeups (timers, repeats, cron, events)",
	Long: `Schedules created by the schedule_wakeup tool survive restarts and fire
through the workspace that owns the session: locally, or on the cloud
machine for cloud schedules.

Pair with cron/systemd for machine-level scheduling, or trigger named
events from webhooks and CI:

  star wakeup schedule --session <id> --prompt "check CI" --repeat-seconds 1800
  star wakeup schedule --session <id> --prompt "morning triage" --daily-at 08:00 --cloud
  star wakeup schedule --session <id> --prompt "bisect the failure" --event ci-failed
  star wakeup trigger --event ci-failed
  star wakeup run --due   # what a cron entry calls every minute`,
}

var (
	wakeupSession string
	wakeupPrompt  string
	wakeupReason  string
	wakeupDelay   int
	wakeupRepeat  int
	wakeupCron    string
	wakeupDailyAt string
	wakeupEvent   string
	wakeupCloud   bool
	wakeupID      string
)

var wakeupScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Create a durable wakeup schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if wakeupSession == "" {
			return fmt.Errorf("Provide --session")
		}
		if wakeupPrompt == "" {
			return fmt.Errorf("Provide --prompt")
		}
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		if _, err := ws.GetSession(cmd.Context(), wakeupSession); err != nil {
			return fmt.Errorf("Unknown session %q: %w", wakeupSession, err)
		}
		disp, err := wakeupDispatcherFor(ws)
		if err != nil {
			return err
		}
		result, err := disp.ScheduleAdvanced(cmd.Context(), tools.ScheduleRequest{
			SessionID:     wakeupSession,
			Delay:         time.Duration(wakeupDelay) * time.Second,
			RepeatSeconds: int64(wakeupRepeat),
			Cron:          wakeupCron,
			DailyAt:       wakeupDailyAt,
			Event:         wakeupEvent,
			Reason:        wakeupReason,
			Prompt:        wakeupPrompt,
			Cloud:         wakeupCloud,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s\nschedule %s\n", result.Message, result.Schedule.ID)
		return nil
	},
}

var wakeupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List durable wakeup schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		disp, err := wakeupDispatcherFor(ws)
		if err != nil {
			return err
		}
		schedules, err := disp.List(cmd.Context())
		if err != nil {
			return err
		}
		for _, sched := range schedules {
			if wakeupSession != "" && sched.SessionID != wakeupSession {
				continue
			}
			next := "armed (event only)"
			if sched.Timed() {
				next = "next " + time.Unix(sched.NextFireUnix, 0).Format(time.RFC822)
			}
			where := "local"
			if sched.Cloud {
				where = "cloud"
			}
			fmt.Printf("%s  session=%s  %s  %s  %s  event=%q reason=%q\n",
				sched.ID, sched.SessionID, sched.Repeat(), next, where, sched.Event, sched.Reason)
		}
		return nil
	},
}

var wakeupCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel wakeup schedules (--session cancels all of a session, --id one schedule)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if wakeupSession == "" && wakeupID == "" {
			return fmt.Errorf("Provide --session or --id")
		}
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		disp, err := wakeupDispatcherFor(ws)
		if err != nil {
			return err
		}
		if wakeupID != "" {
			if err := disp.CancelSchedule(cmd.Context(), wakeupID); err != nil {
				return err
			}
			fmt.Printf("Cancelled schedule %s\n", wakeupID)
			return nil
		}
		disp.Cancel(wakeupSession)
		fmt.Printf("Cancelled wakeups for session %s\n", wakeupSession)
		return nil
	},
}

var wakeupTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Fire a named event (wakes every schedule armed for it)",
	Long: `Fire a named event: every schedule armed with --event <name> wakes now.
With no matching schedule, --session plus --prompt runs a one-off wakeup
directly, so webhooks and CI can wake the agent without pre-arming.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wakeupEvent == "" {
			return fmt.Errorf("Provide --event")
		}
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		disp, err := wakeupDispatcherFor(ws)
		if err != nil {
			return err
		}
		fired, err := disp.FireEvent(cmd.Context(), wakeupEvent, wakeupSession, wakeupPrompt)
		if err != nil {
			return err
		}
		if fired > 0 {
			fmt.Printf("Event %q woke %d schedule(s)\n", wakeupEvent, fired)
			return nil
		}
		if wakeupSession == "" || wakeupPrompt == "" {
			return fmt.Errorf("No schedule is armed for event %q (add --session and --prompt to wake directly)", wakeupEvent)
		}
		if err := runWakeupDirect(cmd.Context(), ws, wakeupSession, wakeupPrompt, wakeupCloud); err != nil {
			return err
		}
		fmt.Printf("Event %q woke session %s directly\n", wakeupEvent, wakeupSession)
		return nil
	},
}

var wakeupRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Fire every due wakeup now (for cron/systemd entries)",
	Long: `Reaps every due schedule and exits: seconds when nothing is due.
Run it every few minutes (cron/systemd, or the wakeup-sweep GitHub
workflow) and no process ever needs to outlive a job limit — a wakeup
24h out simply waits through the no-op sweeps and fires within one
sweep interval of its due time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		disp, err := wakeupDispatcherFor(ws)
		if err != nil {
			return err
		}
		fired, err := disp.RunDue(cmd.Context())
		if err != nil {
			return err
		}
		fmt.Printf("Fired %d due wakeup(s)\n", len(fired))
		return nil
	},
}

var wakeupServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a headless wakeup dispatcher until interrupted",
	Long: `Hosts the durable wakeup timers without a TUI: timed schedules fire
as their time comes and event schedules wait for 'star wakeup trigger'.
Run under systemd (or any supervisor) for machine-level wakeups.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, cleanup, err := setupLocalWorkspace(cmd)
		if err != nil {
			return err
		}
		defer cleanup()
		appWs, ok := ws.(*workspace.AppWorkspace)
		if !ok {
			return fmt.Errorf("Wakeups need a local workspace")
		}
		// This process hosts the timers for as long as it runs.
		if err := appWs.ArmWakeupTimers(cmd.Context()); err != nil {
			return err
		}
		fmt.Println("Wakeup dispatcher running; ctrl+c to stop.")
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		<-ctx.Done()
		return nil
	},
}

// wakeupDispatcherProvider is implemented by coordinators exposing
// their durable wakeup dispatcher.
type wakeupDispatcherProvider interface {
	WakeupDispatcher() *agent.WakeupScheduler
}

func wakeupDispatcherFor(ws workspace.Workspace) (*agent.WakeupScheduler, error) {
	appWs, ok := ws.(*workspace.AppWorkspace)
	if !ok {
		return nil, fmt.Errorf("Wakeups need a local workspace")
	}
	coord := appWs.App().AgentCoordinator
	if coord == nil {
		return nil, fmt.Errorf("Agent is not initialized")
	}
	provider, ok := coord.(wakeupDispatcherProvider)
	if !ok {
		return nil, fmt.Errorf("Agent does not support durable wakeups")
	}
	return provider.WakeupDispatcher(), nil
}

// runWakeupDirect runs a one-off wakeup without a stored schedule,
// with the same routing and busy rules as fired schedules.
func runWakeupDirect(ctx context.Context, ws workspace.Workspace, sessionID, prompt string, cloud bool) error {
	appWs, ok := ws.(*workspace.AppWorkspace)
	if !ok {
		return fmt.Errorf("Wakeups need a local workspace")
	}
	if _, err := ws.GetSession(ctx, sessionID); err != nil {
		return fmt.Errorf("Unknown session %q: %w", sessionID, err)
	}
	if appWs.AgentIsSessionBusy(sessionID) {
		return agent.ErrWakeupBusy
	}
	if cloud {
		return appWs.AgentRunCloud(ctx, sessionID, prompt)
	}
	return appWs.AgentRun(ctx, sessionID, prompt)
}

func init() {
	wakeupScheduleCmd.Flags().StringVar(&wakeupSession, "session", "", "Session to wake")
	wakeupScheduleCmd.Flags().StringVar(&wakeupPrompt, "prompt", "", "Prompt to inject on wake")
	wakeupScheduleCmd.Flags().StringVar(&wakeupReason, "reason", "", "Why the agent is waiting")
	wakeupScheduleCmd.Flags().IntVar(&wakeupDelay, "delay-seconds", 0, "First fire in N seconds")
	wakeupScheduleCmd.Flags().IntVar(&wakeupRepeat, "repeat-seconds", 0, "Repeat every N seconds")
	wakeupScheduleCmd.Flags().StringVar(&wakeupCron, "cron", "", "Cron schedule (five fields or @hourly/@daily/@weekly/@every)")
	wakeupScheduleCmd.Flags().StringVar(&wakeupDailyAt, "daily-at", "", "Wake daily at HH:MM local time")
	wakeupScheduleCmd.Flags().StringVar(&wakeupEvent, "event", "", "Arm for a named event instead of (or in addition to) a timer")
	wakeupScheduleCmd.Flags().BoolVar(&wakeupCloud, "cloud", false, "Run the wakeup on the cloud machine")
	wakeupListCmd.Flags().StringVar(&wakeupSession, "session", "", "Only show one session's schedules")
	wakeupCancelCmd.Flags().StringVar(&wakeupSession, "session", "", "Cancel all of a session's schedules")
	wakeupCancelCmd.Flags().StringVar(&wakeupID, "id", "", "Cancel one schedule by id")
	wakeupTriggerCmd.Flags().StringVar(&wakeupEvent, "event", "", "Event name to fire")
	wakeupTriggerCmd.Flags().StringVar(&wakeupSession, "session", "", "Restrict to (or directly wake) one session")
	wakeupTriggerCmd.Flags().StringVar(&wakeupPrompt, "prompt", "", "Override prompt (or direct-run prompt)")
	wakeupTriggerCmd.Flags().BoolVar(&wakeupCloud, "cloud", false, "Direct runs: use the cloud machine")
	wakeupCmd.AddCommand(wakeupScheduleCmd)
	wakeupCmd.AddCommand(wakeupListCmd)
	wakeupCmd.AddCommand(wakeupCancelCmd)
	wakeupCmd.AddCommand(wakeupTriggerCmd)
	wakeupCmd.AddCommand(wakeupRunCmd)
	wakeupCmd.AddCommand(wakeupServeCmd)
}
