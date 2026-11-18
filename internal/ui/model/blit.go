package model

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// styledBlit caches the decoded cells of a styled string and paints them into
// the screen. Parsing ANSI into cells is the expensive half of drawing a
// region, so a region whose text did not change reuses its decoded buffer and
// only copies cells.
//
// A blit never needs to clear the screen: View composes every frame onto a
// blank canvas, so cells the text does not cover are already empty. The cached
// buffer itself is cleared before it is re-decoded, because StyledString.Draw
// writes over its area instead of clearing it.
type styledBlit struct {
	text   string
	method ansi.Method
	buf    uv.ScreenBuffer
}

// draw paints text into area, decoding it only when the text, the width method
// or the text's shape changed.
func (b *styledBlit) draw(scr uv.Screen, area uv.Rectangle, text string) {
	method, _ := scr.WidthMethod().(ansi.Method)
	w, h := renderedBounds(text, method)
	w = max(w, 1)
	h = max(h, 1)
	resized := b.buf.RenderBuffer == nil || b.method != method || b.buf.Width() != w || b.buf.Height() != h
	retext := b.text != text

	if resized {
		// Only the shape forces a new buffer. The editor's text changes on
		// every keystroke, and reallocating a full cell grid (tens of
		// kilobytes) for each of those is exactly the churn that makes the
		// UI feel sluggish on a loaded machine.
		b.buf = uv.NewScreenBuffer(w, h)
		b.method = method
	} else if retext {
		clearScreenArea(b.buf, b.buf.Bounds())
	}

	if resized || retext {
		b.buf.Method = method
		uv.NewStyledString(text).Draw(b.buf, b.buf.Bounds())
		b.text = text
	}

	b.buf.Draw(scr, area)
}

// clearScreenArea blanks every cell in area.
func clearScreenArea(scr uv.Screen, area uv.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			scr.SetCell(x, y, nil)
		}
	}
}
