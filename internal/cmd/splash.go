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
}

func newSplashModel(cmd *cobra.Command, sessionID string, continueLast bool) *splashModel {
	return &splashModel{cmd: cmd, sessionID: sessionID, continueLast: continueLast}
}

func (m *splashModel) Init() tea.Cmd {
	return m.initialize()
}

func (m *splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.model != nil {
		return m.model.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case splashReadyMsg:
		if msg.err != nil {
			m.initializationErr = msg.err
			return m, tea.Quit
		}
		m.pendingReady = &msg
		return m, m.awaitReleaseCmd()
	case launcherReleasedMsg:
		if m.pendingReady == nil {
			return m, nil
		}
		ready := m.pendingReady
		m.pendingReady = nil
		m.model = ready.model
		m.ws = ready.ws
		m.cleanup = ready.cleanup
		if m.program != nil {
			go m.ws.Subscribe(m.program)
		}
		newModel, updateCmd := m.model.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.model = newModel.(*ui.UI)
		initCmd := m.model.Init()
		return m, tea.Batch(updateCmd, initCmd)
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
		view := m.model.View()
		if !m.rendered {
			m.rendered = true
			if m.handshake != nil {
				m.handshake.notifyRendered()
			}
		}
		return view
	}

	var v tea.View
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "Star"
	v.Content = ""
	return v
}

func (m *splashModel) awaitReleaseCmd() tea.Cmd {
	return func() tea.Msg {
		if m.handshake != nil {
			_ = m.handshake.awaitRelease()
		}
		return launcherReleasedMsg{}
	}
}

func (m *splashModel) initialize() tea.Cmd {
	return func() tea.Msg {
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
