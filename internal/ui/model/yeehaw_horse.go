package model

// Galloping horse animation for yeehaw mode.
//
// The horse crosses the retired editor lane from left to right, wraps
// around off screen, and comes back — 8 gait frames at ~120ms per tick.
// Frames adapted from the llizard aka ejm ASCII gallop cycle (animated
// in JavaScript by Shimrod):
// https://llizard.etherwork.net/ascii-animation/horsegaitsinclpr.html
// Used here as a motion reference, rendered stationary-in-place per
// frame with horizontal travel instead of the original textarea scroll.

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// yeehawHorseTickInterval is the delay between horse frames. Slightly
// slower than the original 100ms browser timing so it reads calm in a
// terminal without losing any of the 8 gait frames.
const yeehawHorseTickInterval = 120 * time.Millisecond

// yeehawHorseStepCells is how many cells the horse travels per frame.
const yeehawHorseStepCells = 3

// yeehawLaneHeight is the editor lane reserved while the horse runs: the
// six visible gallop rows plus a margin row. The yeehaw badge lives
// inside the track (covered by the horse as it passes) so it costs no
// extra line.
const yeehawLaneHeight = 7

// yeehawHorseTickMsg advances the horse one frame. gen retires stale
// clocks when the mode is toggled or the session switches.
type yeehawHorseTickMsg struct {
	gen uint64
}

// yeehawGallopFrames holds the 8-frame gallop cycle. Every frame is 7
// lines; all art is plain ASCII so byte length equals cell width.
var yeehawGallopFrames = [][]string{
	{
		"   ",
		"               ,~",
		"            ~~/(\\",
		"  ~~ _____~~/ / `",
		"~~~~( )___( / ",
		"   /|/_  _`/| ",
		"     \\    / ",
	},
	{
		"   ",
		"               ,~",
		"            ~~/(\\",
		"  ~~ _____~~/ / `",
		"~~~~( )___( / ",
		"    |\\_  _`/|",
		"     `    '  ",
	},
	{
		"   ",
		"               ,~",
		"            ~~/(\\",
		"   ~ _____~~/ / `",
		"~~~~( )___( / ",
		"    |\\_   `/|",
		"     `    // ",
	},
	{
		"   ",
		"              ,~",
		"            ~/(\\",
		"   ~ _____~/ / `",
		" ~~~( )___( |",
		"    |\\     ` \\",
		"     `\\    '/",
	},
	{
		"   ",
		"              ,~",
		"            ~/(\\",
		"   ~ _____~/ / `",
		"~~~~( )___(_|_",
		"     \\     _\\/",
		"     `\\  ",
	},
	{
		"   ",
		"              ,~",
		"            ~/(\\",
		"   ~ _____~/ / `",
		" ~~~( )___(_|_",
		"   ///     _\\/",
		"   ` \\",
	},
	{
		"   ",
		"              ,~",
		"            ~/(\\",
		"   ~ _____~/ / `",
		" ~~~( )___(_|_",
		"   / /     _\\ \\",
		"  //            ",
	},
	{
		"   ",
		"              ,~",
		"            ~/(\\",
		"   ~ _____~/ / `",
		" ~~~( )___(_|_",
		"   /|/_      \\ \\",
		"   `          \\",
	},
}

// yeehawHorseWidth is the widest gallop line, used for travel math and
// the narrow-terminal fallback.
var yeehawHorseWidth = func() int {
	w := 0
	for _, f := range yeehawGallopFrames {
		for _, ln := range f {
			if len(ln) > w {
				w = len(ln)
			}
		}
	}
	return w
}()

// armYeehawHorse starts (or continues) the horse clock. Each tick
// carries the current generation so a toggle or session switch retires
// the old clock.
func (m *UI) armYeehawHorse() tea.Cmd {
	m.yeehawHorseGen++
	gen := m.yeehawHorseGen
	return tea.Tick(yeehawHorseTickInterval, func(time.Time) tea.Msg {
		return yeehawHorseTickMsg{gen: gen}
	})
}

// yeehawBadgeText is the stand-in label for the retired editor. It sits
// at the center of the horse track, behind the horse: the horse covers it
// as it passes and it shows through the gaps in between.
const yeehawBadgeText = "\u25b6 yeehaw"

// yeehawHorseView renders the horse lane: the current gait frame at its
// travel position, clipped to width, with the badge composited into the
// middle track row. Narrow terminals fall back to the static
// placeholder.
func (m *UI) yeehawHorseView(width int) string {
	if width < yeehawHorseWidth+4 {
		return m.yeehawPlaceholderView(width)
	}
	// Skip the leading blank padding line so the lane stays compact.
	frame := yeehawGallopFrames[m.yeehawHorseTick%len(yeehawGallopFrames)][1:]
	travel := width + yeehawHorseWidth
	x := (m.yeehawHorseTick*yeehawHorseStepCells)%travel - yeehawHorseWidth
	lines := make([]string, 0, len(frame)+1)
	for i, ln := range frame {
		start := 0
		pad := x
		if x < 0 {
			start = -x
			pad = 0
		}
		visible := ""
		if start < len(ln) {
			visible = ln[start:]
		}
		row := strings.Repeat(" ", pad) + visible
		if len(row) > width {
			row = row[:width]
		} else {
			row += strings.Repeat(" ", width-len(row))
		}
		if i == len(frame)/2 {
			row = m.yeehawBadgeRow(row, width)
		}
		lines = append(lines, row)
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

// yeehawBadgeRow lays the centered badge behind the horse row: cells
// where the horse drew something keep the horse, gaps show the badge in
// the placeholder style. A badge rune overlapping the horse on only part
// of a wide cell is skipped so the row keeps its exact cell width.
func (m *UI) yeehawBadgeRow(row string, width int) string {
	cells := []rune(row)
	badged := make([]bool, len(cells))
	col := (width - lipgloss.Width(yeehawBadgeText)) / 2
	if col < 0 {
		col = 0
	}
	for _, r := range yeehawBadgeText {
		w := lipgloss.Width(string(r))
		if w < 1 {
			w = 1
		}
		if col+w > len(cells) {
			break
		}
		clear := true
		for c := col; c < col+w; c++ {
			if cells[c] != ' ' {
				clear = false
				break
			}
		}
		if clear {
			cells[col] = r
			badged[col] = true
			for c := col + 1; c < col+w; c++ {
				cells[c] = 0
				badged[c] = true
			}
		}
		col += w
	}
	// The placeholder style carries a set " ! " icon value and a right
	// margin for textarea prompts; strip both so badge runs render
	// exactly the cells placed above.
	badge := m.com.Styles.Editor.PromptYoloIconBlurred.UnsetString().UnsetMargins()
	var out strings.Builder
	var run []rune
	flush := func(styled bool) {
		if len(run) == 0 {
			return
		}
		if styled {
			out.WriteString(badge.Render(string(run)))
		} else {
			out.WriteString(string(run))
		}
		run = run[:0]
	}
	cur := false
	started := false
	for i, r := range cells {
		if r == 0 {
			continue
		}
		if !started {
			cur, started = badged[i], true
		} else if badged[i] != cur {
			flush(cur)
			cur = badged[i]
		}
		run = append(run, r)
	}
	flush(cur)
	return out.String()
}
