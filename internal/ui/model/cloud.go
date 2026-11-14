package model

import (
	"context"
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/util"
)

// Cloud mode runs a session's prompts on a cloud machine through the
// Agent Jobs API and streams the cloud agent's ACP events back into the
// local message store as native assistant/tool-call parts. Rendering,
// spinner, cancellation, and turn lifetime all flow through the normal
// pipeline, so a cloud turn looks exactly like a local one apart from
// the CLOUD badge in the status bar.
//
// cloudRunner is the optional workspace capability behind cloud mode.
// AppWorkspace implements it; remote workspaces do not, so the TUI
// probes with a type assertion instead of extending the Workspace
// interface.
type cloudRunner interface {
	AgentRunCloud(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) error
}

// cloudRunnerFor reports the workspace's cloud capability, or nil when
// cloud turns are unavailable (e.g. client/server mode).
func (m *UI) cloudRunnerFor() cloudRunner {
	if m.com == nil || m.com.Workspace == nil {
		return nil
	}
	runner, _ := m.com.Workspace.(cloudRunner)
	return runner
}

// setCloudMode writes the session's cloud flag, allocating the map on
// first use so a UI built without NewUI (tests, embedders) cannot
// panic on a nil map.
func (m *UI) setCloudMode(sessionID string, on bool) {
	if sessionID == "" {
		return
	}
	if m.cloudSessions == nil {
		m.cloudSessions = make(map[string]bool)
	}
	m.cloudSessions[sessionID] = on
}

// markCloudActive records that a cloud turn started for the session.
func (m *UI) markCloudActive(sessionID string) {
	if sessionID == "" {
		return
	}
	if m.cloudActive == nil {
		m.cloudActive = make(map[string]bool)
	}
	m.markCloudActive(sessionID)
}

// cloudModeFor reports whether cloud mode is enabled for the session.
func (m *UI) cloudModeFor(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	return m.cloudSessions[sessionID]
}

// cloudActiveFor reports whether a cloud turn is currently running for
// the session. It mirrors the workspace-side registry so esc and the
// busy indicator work even when the local coordinator is not ready.
func (m *UI) cloudActiveFor(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	return m.cloudActive[sessionID]
}

// clearCloudTurn retires the in-flight marker for a finished cloud
// turn. It runs on the RunComplete path every turn ends through.
func (m *UI) clearCloudTurn(sessionID string) {
	if sessionID == "" {
		return
	}
	delete(m.cloudActive, sessionID)
}

// handleCloudSubmit intercepts /cloud commands and routes cloud-mode
// prompts to the cloud runner. It reports whether it handled the
// submission; when false, the normal local path owns the content.
func (m *UI) handleCloudSubmit(content string, attachments []message.Attachment) (bool, tea.Cmd) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "/cloud" || strings.HasPrefix(trimmed, "/cloud ") {
		return true, m.toggleCloudMode(strings.TrimSpace(strings.TrimPrefix(trimmed, "/cloud")))
	}
	if m.cloudModeFor(m.currentSessionID()) {
		return true, m.sendCloudMessage(content, attachments)
	}
	return false, nil
}

// toggleCloudMode flips cloud mode for the current session. /cloud
// toggles, /cloud on|off sets explicitly; anything else explains usage.
func (m *UI) toggleCloudMode(arg string) tea.Cmd {
	loadCmd, err := m.ensureSession()
	if err != nil {
		return util.ReportError(err)
	}
	var cmds []tea.Cmd
	if loadCmd != nil {
		cmds = append(cmds, loadCmd)
	}
	if m.cloudRunnerFor() == nil {
		cmds = append(cmds, util.ReportError(errors.New("Cloud mode needs a local workspace (unavailable in client/server mode)")))
		return tea.Batch(cmds...)
	}
	sessionID := m.currentSessionID()
	switch arg {
	case "", "toggle":
		m.setCloudMode(sessionID, !m.cloudModeFor(sessionID))
	case "on":
		m.setCloudMode(sessionID, true)
	case "off":
		m.setCloudMode(sessionID, false)
	default:
		cmds = append(cmds, util.ReportInfo("Usage: /cloud [on|off]"))
		return tea.Batch(cmds...)
	}
	on := m.cloudModeFor(sessionID)
	m.status.SetCloudMode(on)
	if on {
		cmds = append(cmds, util.ReportInfo("Cloud mode on — prompts run on a cloud machine"))
	} else {
		cmds = append(cmds, util.ReportInfo("Cloud mode off — prompts run locally"))
	}
	return tea.Batch(cmds...)
}

// sendCloudMessage submits one prompt to the cloud runner. It mirrors
// sendMessageInternal's pending-turn and optimistic-busy sequence so
// the echo, spinner, esc, and RunComplete retirement behave
// identically; only the execution backend differs.
func (m *UI) sendCloudMessage(content string, attachments []message.Attachment) tea.Cmd {
	runner := m.cloudRunnerFor()
	if runner == nil {
		return util.ReportError(errors.New("Cloud mode needs a local workspace (unavailable in client/server mode)"))
	}

	// Any new prompt supersedes a pending, unconfirmed plan.
	m.setPlanReadyPending("")
	m.turnOutcome = ""

	var cmds []tea.Cmd
	loadCmd, err := m.ensureSession()
	if err != nil {
		return util.ReportError(err)
	}
	if loadCmd != nil {
		cmds = append(cmds, loadCmd)
	}

	sessionID := m.currentSessionID()
	if m.cloudActiveFor(sessionID) || m.turnInFlight(sessionID) {
		return util.ReportWarn("A cloud turn is already running — esc cancels it")
	}
	m.chat.SetAnimationsAllowed(true)
	common.StartTurn(sessionID)
	if m.showPendingTurn(sessionID, content, attachments) {
		m.chat.ScrollToBottom()
	}

	ctx := context.Background()
	cmds = append(cmds, func() tea.Msg {
		for _, path := range m.sessionFileReads {
			m.com.Workspace.FileTrackerRecordRead(ctx, m.session.ID, path)
			m.com.Workspace.LSPStart(ctx, path)
		}
		return nil
	})

	m.agentBusyCache.set(true)
	m.busyFetchGen++
	m.invalidatePromptQueue()
	m.cloudActive[sessionID] = true
	cmds = append(cmds, func() tea.Msg {
		// AgentRunCloud is fire-and-forget like AgentRun: it returns
		// once the cloud turn is accepted. Failures to start it surface
		// here; run failures and cancellation surface through the
		// message stream and the terminal RunComplete.
		err := runner.AgentRunCloud(context.Background(), sessionID, content, attachments...)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		return agentRunSubmittedMsg{sessionID: sessionID, err: err}
	})
	return tea.Batch(cmds...)
}
