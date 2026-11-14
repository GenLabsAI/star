package workspace

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/charmbracelet/crush/internal/cloud"
	"github.com/charmbracelet/crush/internal/message"
)

// AgentRunCloud starts one cloud turn for the session and returns once
// it is accepted, mirroring AgentRun's fire-and-forget contract. The
// turn streams into the local message store and ends with a
// RunComplete, so cancellation (AgentCancel), busy probes, and the
// TUI's pending-turn machinery behave exactly like a local run.
func (w *AppWorkspace) AgentRunCloud(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) error {
	if w.app == nil || w.app.Messages == nil {
		return fmt.Errorf("Cloud mode needs a local workspace")
	}
	model := w.AgentSessionModel(sessionID)
	opts, err := cloud.RunOptionsFromEnv(model.ModelCfg.Model, string(model.ModelCfg.Provider))
	if err != nil {
		return err
	}
	client := cloud.NewClient(opts.BaseURL, opts.UserToken, os.Getenv("DECA_AGENT_WORKER_TOKEN"))

	runCtx, cancel := context.WithCancel(context.Background())
	w.trackCloudRun(sessionID, cancel)
	go func() {
		defer w.untrackCloudRun(sessionID, cancel)
		_ = client.RunCloudSession(runCtx, w.app.Messages, w.app.RunCompletions(), sessionID, prompt, attachments, opts)
	}()
	return nil
}

// AgentIsCloudSessionBusy reports whether a cloud turn is running for
// the session.
func (w *AppWorkspace) AgentIsCloudSessionBusy(sessionID string) bool {
	w.cloudMu.Lock()
	defer w.cloudMu.Unlock()
	_, ok := w.cloudRuns[sessionID]
	return ok
}

// AgentIsCloudBusy reports whether any cloud turn is running.
func (w *AppWorkspace) AgentIsCloudBusy() bool {
	w.cloudMu.Lock()
	defer w.cloudMu.Unlock()
	return len(w.cloudRuns) > 0
}

func (w *AppWorkspace) trackCloudRun(sessionID string, cancel context.CancelFunc) {
	w.cloudMu.Lock()
	defer w.cloudMu.Unlock()
	if w.cloudRuns == nil {
		w.cloudRuns = make(map[string]context.CancelFunc)
	}
	if prev, ok := w.cloudRuns[sessionID]; ok {
		prev()
	}
	w.cloudRuns[sessionID] = cancel
}

func (w *AppWorkspace) untrackCloudRun(sessionID string, cancel context.CancelFunc) {
	w.cloudMu.Lock()
	defer w.cloudMu.Unlock()
	if cur, ok := w.cloudRuns[sessionID]; ok && sameCancel(cur, cancel) {
		delete(w.cloudRuns, sessionID)
	}
}

// sameCancel compares cancel funcs by identity so a stale deferred
// untrack cannot drop a newer run's entry. Funcs are not comparable
// with ==, so compare their reflect pointers.
func sameCancel(a, b context.CancelFunc) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}
