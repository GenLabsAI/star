package model

import (
	"image"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// ghostTextDialog paints a sentinel string at a fixed screen position so a
// test can prove the canvas is reset when the dialog closes.
type ghostTextDialog struct {
	text string
	at   image.Point
}

func (ghostTextDialog) ID() string                      { return "ghost-text" }
func (ghostTextDialog) HandleMsg(tea.Msg) dialog.Action { return nil }
func (d ghostTextDialog) Draw(scr uv.Screen, _ uv.Rectangle) *tea.Cursor {
	uv.NewStyledString(d.text).Draw(scr, image.Rect(d.at.X, d.at.Y, d.at.X+len(d.text), d.at.Y+1))
	return nil
}

func TestDialogCloseClearsPaintedCells(t *testing.T) {
	t.Parallel()

	u := newFrameTestUI(t)
	// Keep the wide layout: the compact header is not part of this fixture.
	u.width, u.height = 140, 45
	u.updateLayoutAndSize()
	u.chat.ScrollToTop()

	// Paint over a blank part of a chat row. The cached chat buffer has no
	// cells there, so only an explicit clear can remove the dialog text.
	const text = "MODAL_GHOST"
	at := image.Pt(u.layout.main.Min.X+8, u.layout.main.Min.Y+4)
	u.dialog.OpenDialog(ghostTextDialog{text: text, at: at})

	require.Contains(t, u.View().Content, text)

	u.dialog.CloseFrontDialog()
	require.NotContains(t, u.View().Content, text,
		"closing a dialog must clear cells it painted over blank chat rows")
}
