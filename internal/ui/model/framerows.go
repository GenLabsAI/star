package model

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// frameRowCache memoizes the rendered text of every screen row, so a frame only
// pays for the rows whose cells changed since the previous one.
//
// ultraviolet's RenderBuffer marks a line as touched whenever a cell on it
// changes, and the draw paths that run every frame avoid clearing regions whose
// content is unchanged (see [styledBlit]), so the touched set is an accurate
// diff of the screen.
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

// render returns the frame string for canvas, re-rendering only dirty rows.
func (c *frameRowCache) render(canvas uv.ScreenBuffer) string {
	w, h := canvas.Width(), canvas.Height()
	resized := !c.valid || c.width != w || c.height != h
	if resized {
		c.text = make([]string, h)
		c.scratch = uv.NewScreenBuffer(max(w, 1), 1)
		c.width, c.height = w, h
		c.valid = true
	}

	dirty := resized
	for y := range h {
		if !resized && (y >= len(canvas.Touched) || canvas.Touched[y] == nil) {
			continue
		}
		c.text[y] = c.renderRow(canvas, y)
		dirty = true
	}
	for i := range canvas.Touched {
		canvas.Touched[i] = nil
	}
	if !dirty {
		return c.content
	}
	c.content = strings.Join(c.text, "\n")
	return c.content
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
	row := c.scratch.Render()
	if y < c.height-1 {
		// Every row but the last is followed by a LF, so its trailing CR
		// is a line ending and is trimmed, matching a whole-frame render.
		row = strings.TrimSuffix(row, "\r")
	}
	return strings.TrimRight(row, " ")
}
