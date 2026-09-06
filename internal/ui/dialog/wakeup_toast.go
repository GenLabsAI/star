package dialog

import (
	"fmt"
	"image"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/ui/common"
	uv "github.com/charmbracelet/ultraviolet"
)

// WakeupToastID is the identifier for the wakeup countdown toast.
const WakeupToastID = "wakeup-toast"

// ActionCancelWakeup is returned when the user cancels a pending wakeup.
type ActionCancelWakeup struct {
	SessionID string
}

// WakeupToast is a small toast dialog showing a countdown until the agent
// wakes up again, with a Cancel button.
type WakeupToast struct {
	com        *common.Common
	SessionID  string
	Reason     string
	EndTime    time.Time
	compositor *lipgloss.Compositor
	keyMap     struct {
		Cancel,
		Close key.Binding
	}
}

var _ Dialog = (*WakeupToast)(nil)

// NewWakeupToast creates a new wakeup countdown toast.
func NewWakeupToast(com *common.Common, sessionID, reason string, endTime time.Time) *WakeupToast {
	w := &WakeupToast{
		com:       com,
		SessionID: sessionID,
		Reason:    reason,
		EndTime:   endTime,
	}
	w.keyMap.Cancel = key.NewBinding(
		key.WithKeys("esc", "alt+esc", "c", "C"),
		key.WithHelp("esc", "cancel"),
	)
	w.keyMap.Close = CloseKey
	return w
}

// ID implements [Dialog].
func (*WakeupToast) ID() string {
	return WakeupToastID
}

// HandleMsg implements [Dialog].
func (w *WakeupToast) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, w.keyMap.Cancel, w.keyMap.Close):
			return ActionCancelWakeup{SessionID: w.SessionID}
		}
	case tea.MouseClickMsg:
		if common.HitButtonIndex(w.compositor, msg.X, msg.Y) == 0 {
			return ActionCancelWakeup{SessionID: w.SessionID}
		}
	}
	return nil
}

// Draw implements [Dialog].
func (w *WakeupToast) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	remaining := time.Until(w.EndTime)
	if remaining < 0 {
		remaining = 0
	}
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60

	countdown := fmt.Sprintf("Star will wake up in %02d:%02d", minutes, seconds)

	var (
		baseStyle = w.com.Styles.Dialog.Quit.Content
		hintStyle = w.com.Styles.Dialog.Quit.Hint
	)

	buttonOpts := []common.ButtonOpts{
		{Text: "Cancel", Selected: true, Padding: 3},
	}
	buttons := common.ButtonGroup(w.com.Styles, buttonOpts, " ")

	lines := []string{countdown}
	if w.Reason != "" {
		lines = append(lines, hintStyle.Render(w.Reason))
	}
	lines = append(lines, "", buttons)

	content := baseStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center, lines...),
	)

	frameStyle := w.com.Styles.Dialog.Quit.Frame
	view := frameStyle.Render(content)

	width, height := lipgloss.Size(view)
	width = min(width, area.Dx())
	height = min(height, area.Dy())
	bottomRight := image.Rect(area.Max.X-width, area.Max.Y-height, area.Max.X, area.Max.Y)

	frameTop := bottomRight.Min.Y + frameStyle.GetBorderTopSize() + frameStyle.GetPaddingTop()
	contentWidth := lipgloss.Width(content)
	contentLeft := bottomRight.Min.X + frameStyle.GetBorderLeftSize() + frameStyle.GetPaddingLeft()
	buttonWidth := lipgloss.Width(buttons)
	buttonX := contentLeft + max(0, (contentWidth-buttonWidth)/2)
	buttonY := frameTop + len(lines) - 1
	w.compositor = common.ButtonHitCompositor(w.com.Styles, buttonOpts, " ", buttonX, buttonY)

	uv.NewStyledString(view).Draw(scr, bottomRight)
	return nil
}

// ShortHelp implements [help.KeyMap].
func (w *WakeupToast) ShortHelp() []key.Binding {
	return []key.Binding{w.keyMap.Cancel}
}

// FullHelp implements [help.KeyMap].
func (w *WakeupToast) FullHelp() [][]key.Binding {
	return [][]key.Binding{{w.keyMap.Cancel}}
}
