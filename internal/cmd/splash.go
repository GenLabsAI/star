package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/event"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/common"
	ui "github.com/charmbracelet/crush/internal/ui/model"
	"github.com/charmbracelet/crush/internal/workspace"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
)

// launcherHandshake coordinates with the Rust launcher (star.exe) so that
// Bubble Tea never touches the terminal until the launcher's splash
// animation has fully released the alt-screen buffer. Without this gate,
// bubbletea's renderer and the launcher's animation thread both write to
// the same alt-screen buffer concurrently, causing visible flicker.
type launcherHandshake struct {
	active        bool
	readyEvent    windows.Handle
	releaseEvent  windows.Handle
	renderedEvent windows.Handle
}

func newLauncherHandshake() *launcherHandshake {
	if os.Getenv("STAR_LAUNCHER_HANDSHAKE") != "1" {
		return &launcherHandshake{}
	}
	pid := os.Getenv("STAR_LAUNCHER_PID")
	readyName, _ := windows.UTF16PtrFromString("Local\\star-ready-" + pid)
	releaseName, _ := windows.UTF16PtrFromString("Local\\star-release-" + pid)
	renderedName, _ := windows.UTF16PtrFromString("Local\\star-rendered-" + pid)
	readyEvent, err1 := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, readyName)
	releaseEvent, err2 := windows.OpenEvent(windows.SYNCHRONIZE, false, releaseName)
	renderedEvent, err3 := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, renderedName)
	if err1 != nil || err2 != nil || err3 != nil {
		return &launcherHandshake{}
	}
	return &launcherHandshake{
		active:        true,
		readyEvent:    readyEvent,
		releaseEvent:  releaseEvent,
		renderedEvent: renderedEvent,
	}
}

// awaitRelease notifies the launcher that Go has finished initializing and
// blocks until the launcher has stopped animating and given up ownership of
// the terminal. It must be called before tea.NewProgram/Run so bubbletea
// never writes to the terminal while the launcher still owns it.
func (h *launcherHandshake) awaitRelease() error {
	if !h.active {
		return nil
	}
	if err := windows.SetEvent(h.readyEvent); err != nil {
		return fmt.Errorf("notify launcher: %w", err)
	}
	if _, err := windows.WaitForSingleObject(h.releaseEvent, windows.INFINITE); err != nil {
		return fmt.Errorf("wait for launcher: %w", err)
	}
	_ = windows.CloseHandle(h.readyEvent)
	_ = windows.CloseHandle(h.releaseEvent)
	return nil
}

// notifyRendered tells the launcher that bubbletea has produced its first
// real frame, so the launcher process can finally exit. Safe to call
// multiple times; only the first call has any effect.
func (h *launcherHandshake) notifyRendered() {
	if !h.active {
		return
	}
	_ = windows.SetEvent(h.renderedEvent)
	_ = windows.CloseHandle(h.renderedEvent)
	h.active = false
}

type splashReadyMsg struct {
	model   *ui.UI
	ws      workspace.Workspace
	cleanup func()
	err     error
}

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
		// Release the launcher only once the workspace is ready, then let
		// the UI take over. The launcher has already stopped drawing by the
		// time awaitRelease returns, so there is no contention.
		if m.handshake != nil {
			_ = m.handshake.awaitRelease()
		}
		m.model = msg.model
		m.ws = msg.ws
		m.cleanup = msg.cleanup
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
