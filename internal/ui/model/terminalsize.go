package model

import (
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"
)

// terminalSizePollMsg carries the size the terminal answers with when asked
// directly, rather than one carried by a resize event.
type terminalSizePollMsg struct {
	width  int
	height int
}

// terminalSizePollInterval is short enough that a resized window settles
// almost immediately and long enough to stay invisible in a profile.
const terminalSizePollInterval = 400 * time.Millisecond

// terminalSizeTicker schedules the next terminal size check.
//
// Bubble Tea's resize events are not reliable everywhere: a Windows console
// can drop them, and a dropped event leaves every frame laid out for the
// previous width. Rows wider than the new window then wrap, which reads as
// overwrapping with text stranded at the left edge. Asking the terminal
// directly on a timer converges on the real size without depending on the
// event at all.
func (m *UI) terminalSizeTicker() tea.Cmd {
	return tea.Tick(terminalSizePollInterval, func(time.Time) tea.Msg {
		width, height, err := term.GetSize(os.Stdout.Fd())
		if err != nil || width <= 0 || height <= 0 {
			return terminalSizePollMsg{}
		}
		return terminalSizePollMsg{width: width, height: height}
	})
}

// terminalSizePollCmd turns a poll result into the resize the UI still needs
// to apply, or nil when the reported size is already the current one.
func (m *UI) terminalSizePollCmd(msg terminalSizePollMsg) tea.Cmd {
	if msg.width <= 0 || msg.height <= 0 {
		return nil
	}
	if msg.width == m.width && msg.height == m.height {
		return nil
	}
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: msg.width, Height: msg.height}
	}
}
