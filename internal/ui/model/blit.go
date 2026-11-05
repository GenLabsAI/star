package model

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// styledBlit caches the decoded cells of a styled string and paints them into
// the screen. Unlike [uv.StyledString.Draw], which clears its whole area on
// every call, a blit only clears when its text changes. A region that renders
// the same content as the previous frame therefore leaves its cells untouched,
// which is what lets [frameRowCache] tell that the row did not change.
type styledBlit struct {
	text   string
	method ansi.Method
	buf    uv.ScreenBuffer
	valid  bool
}

// invalidate forces the next draw to decode and clear even if the text is
// unchanged. Use it when something else painted over the blit's area.
func (b *styledBlit) invalidate() {
	b.valid = false
}

// draw paints text into area, decoding it only when the text or the screen's
// width method changed.
func (b *styledBlit) draw(scr uv.Screen, area uv.Rectangle, text string) {
	method, _ := scr.WidthMethod().(ansi.Method)
	w, h := renderedBounds(text, method)
	w = max(w, 1)
	h = max(h, 1)
	resized := b.buf.RenderBuffer == nil || b.method != method || b.buf.Width() != w || b.buf.Height() != h
	if !b.valid || resized || b.text != text {
		clearScreenArea(scr, area)
		if resized {
			// Only the shape forces a new buffer. The editor's text changes
			// on every keystroke, and reallocating a full cell grid (tens of
			// kilobytes) for each of those is exactly the churn that makes
			// the UI feel sluggish on a loaded machine.
			b.buf = uv.NewScreenBuffer(w, h)
			b.method = method
		} else {
			// Drop the previous text so shorter content leaves no cells
			// behind: StyledString.Draw writes over, it does not clear.
			clearScreenArea(b.buf, b.buf.Bounds())
		}
		b.buf.Method = method
		uv.NewStyledString(text).Draw(b.buf, b.buf.Bounds())
		b.text = text
		b.valid = true
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
