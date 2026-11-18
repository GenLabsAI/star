package model

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// frameRowCache memoizes the rendered text of every screen row, so a frame only
// pays for the rows whose cells changed since the previous one.
//
// View composes every frame from scratch onto a blank canvas, so the canvas is
// the whole truth for the current state: no cell can survive from an earlier
// frame and nothing painted over another region can leave anything behind. That
// is what makes a plain comparison against the previous frame's cells a
// complete and safe diff, with no invalidation bookkeeping to forget.
type frameRowCache struct {
	text    []string
	scratch uv.ScreenBuffer
	width   int
	height  int
	content string
	valid   bool

	// rendered counts rows re-rendered since the last reset. Benchmarks
	// read it to check that steady-state frames really are row-diffed.
	rendered int
}

// render returns the frame string for canvas, reusing the text of every row
// whose cells match prev. A frame with a different shape (or the first frame)
// renders every row; pass the previous frame's canvas as prev.
func (c *frameRowCache) render(canvas, prev uv.ScreenBuffer) string {
	w, h := canvas.Width(), canvas.Height()
	resized := !c.valid || c.width != w || c.height != h || prev.Width() != w || prev.Height() != h
	if resized {
		c.text = make([]string, h)
		c.scratch = uv.NewScreenBuffer(max(w, 1), 1)
		c.width, c.height = w, h
		c.valid = true
	}

	dirty := resized
	for y := range h {
		if !resized && lineEqual(canvas.Lines[y], prev.Lines[y]) {
			continue
		}
		c.text[y] = c.renderRow(canvas, y)
		dirty = true
	}
	if !dirty {
		return c.content
	}
	c.content = strings.Join(c.text, "\n")
	return c.content
}

// lineEqual reports whether two canvas rows paint the same cells.
func lineEqual(a, b uv.Line) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Content != b[i].Content || a[i].Width != b[i].Width ||
			a[i].Link != b[i].Link || !a[i].Style.Equal(&b[i].Style) {
			return false
		}
	}
	return true
}

// renderRow renders a single buffer line in isolation. Rendering line by line
// is what makes the cache possible: [uv.Buffer.Render] renders the whole grid,
// changed or not.
func (c *frameRowCache) renderRow(canvas uv.ScreenBuffer, y int) string {
	if y < 0 || y >= len(canvas.Lines) || len(c.scratch.Lines) == 0 {
		return ""
	}
	c.rendered++
	copy(c.scratch.Lines[0], canvas.Lines[y])
	return strings.TrimRight(c.scratch.Render(), " ")
}
