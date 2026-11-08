package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/event"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/common"
	ui "github.com/charmbracelet/crush/internal/ui/model"
	"github.com/charmbracelet/crush/internal/workspace"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/spf13/cobra"
)

type splashReadyMsg struct {
	model   *ui.UI
	ws      workspace.Workspace
	cleanup func()
	err     error
}

type launcherReleasedMsg struct{}

type splashModel struct {
	cmd               *cobra.Command
	sessionID         string
	continueLast      bool
	width             int
	height            int
	model             *ui.UI
	ws                workspace.Workspace
	cleanup           func()
	initializationErr error
	program           *tea.Program
	handshake         *launcherHandshake
	rendered          bool
	pendingReady      *splashReadyMsg
	launcherReleased  bool
	// launcherOwnsAltScreen is true when the Rust launcher entered the
	// alternate screen and will leave it after the core exits. The core must
	// not request the same screen again: a second DECSET 1049 can restore the
	// saved splash buffer on top of the UI.
	launcherOwnsAltScreen bool
}

func newSplashModel(cmd *cobra.Command, sessionID string, continueLast bool) *splashModel {
	return &splashModel{cmd: cmd, sessionID: sessionID, continueLast: continueLast}
}

func (m *splashModel) Init() tea.Cmd {
	// The launcher has already released the terminal before the program is
	// started (see rootCmd), so this only needs to report our first frame
	// and let the launcher exit.
	return func() tea.Msg { return launcherReleasedMsg{} }
}

func (m *splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.model != nil {
		return m.model.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.launcherReleased && m.pendingReady != nil && m.width > 0 && m.height > 0 {
			return m, m.transitionToUI(m.pendingReady)
		}
	case splashReadyMsg:
		if msg.err != nil {
			m.initializationErr = msg.err
			return m, tea.Quit
		}
		if m.launcherReleased && m.width > 0 && m.height > 0 {
			return m, m.transitionToUI(&msg)
		}
		m.pendingReady = &msg
	case launcherReleasedMsg:
		m.launcherReleased = true
		if m.handshake != nil {
			m.handshake.notifyRendered()
		}
		// The launcher kept the alternate screen active through the handoff,
		// so this is the same buffer it painted the splash on. Clear it now
		// that this process is drawing, then let the UI take over.
		clear := func() tea.Msg { return tea.ClearScreen() }
		if m.pendingReady != nil && m.width > 0 && m.height > 0 {
			return m, tea.Batch(clear, m.transitionToUI(m.pendingReady))
		}
		return m, clear
	}
	return m, nil
}

func (m *splashModel) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	if m.model != nil {
		return m.model.Draw(scr, area)
	}
	return nil
}

func (m *splashModel) View() tea.View {
	if m.model != nil {
		return m.model.View()
	}

	var v tea.View
	// A directly launched core owns the terminal and needs the alternate
	// screen. A launcher-driven core must leave it alone: the launcher already
	// entered it and will restore it when the core exits.
	v.AltScreen = !m.launcherOwnsAltScreen
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "Star"
	v.Content = ""
	return v
}

func (m *splashModel) transitionToUI(ready *splashReadyMsg) tea.Cmd {
	m.pendingReady = nil
	m.model = ready.model
	m.model.SetAltScreen(!m.launcherOwnsAltScreen)
	m.ws = ready.ws
	m.cleanup = ready.cleanup
	if m.program != nil {
		go m.ws.Subscribe(m.program)
	}
	newModel, updateCmd := m.model.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.model = newModel.(*ui.UI)
	initCmd := m.model.Init()
	return tea.Batch(updateCmd, initCmd)
}

func (m *splashModel) initialize() splashReadyMsg {
	ws, cleanup, err := setupWorkspace(m.cmd)
	if err != nil {
		return splashReadyMsg{err: err}
	}

	sessionID := m.sessionID
	if sessionID != "" {
		sess, err := resolveWorkspaceSessionID(m.cmd.Context(), ws, sessionID)
		if err != nil {
			cleanup()
			return splashReadyMsg{err: err}
		}
		sessionID = sess.ID
	}

	event.AppInitialized()
	return splashReadyMsg{
		model:   ui.New(common.DefaultCommon(ws), sessionID, m.continueLast),
		ws:      ws,
		cleanup: cleanup,
	}
}

func (m *splashModel) ViewModel() *ui.UI {
	return m.model
}

func (m *splashModel) Subscribe(program *tea.Program) {
	m.program = program
	if m.ws != nil {
		go m.ws.Subscribe(program)
	}
}

func (m *splashModel) Shutdown() {
	if m.cleanup != nil {
		m.cleanup()
	}
}

func (m *splashModel) Err() error {
	return m.initializationErr
}

var _ tea.Model = (*splashModel)(nil)
var _ = fmt.Sprintf
var _ uv.Environ
var _ = session.HashID
