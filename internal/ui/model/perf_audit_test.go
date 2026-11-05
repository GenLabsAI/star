package model

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/session"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// auditUI returns a wide-layout chat UI carrying a realistic session, so the
// whole frame path is measured: layout, sidebar, chat list, editor, status bar
// and the screen-buffer render. newFrameTestUI already wires the stubs View
// needs (workspace, dialog overlay, key map, attachments).
func auditUI(tb testing.TB, n int) *UI {
	tb.Helper()
	u := newFrameTestUI(tb)
	u.com.Workspace = &testWorkspace{cfg: &config.Config{Options: &config.Options{}}}
	u.header = newHeader(u.com)
	u.width, u.height = 140, 45
	u.chat.SetMessages(buildSession(n)...)
	u.session = &session.Session{ID: "audit", Title: "Refactor the render pipeline"}
	u.textarea.Focus()
	u.updateLayoutAndSize()
	return u
}

// BenchmarkAuditFrame is the felt experience: one keypress or one streamed
// delta, followed by the full View the event loop composes after every
// message. view_uncached disables the frame cache so the raw composition cost
// is visible rather than a memoized hit.
func BenchmarkAuditFrame(b *testing.B) {
	for _, n := range []int{50, 200} {
		b.Run(fmt.Sprintf("msgs=%d", n*2), func(b *testing.B) {
			b.Run("view_uncached", func(b *testing.B) {
				u := auditUI(b, n)
				u.frames = nil
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					u.beginFrameUpdate()
					u.View()
				}
				reportHeap(b, u)
			})

			b.Run("keypress_type", func(b *testing.B) {
				u := auditUI(b, n)
				u.frames = nil
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; b.Loop(); i++ {
					// Alternate a rune and a backspace so the editor buffer
					// stays the same size and the measurement is steady state.
					var msg tea.KeyPressMsg
					if i%2 == 0 {
						msg = tea.KeyPressMsg{Code: 'a', Text: "a"}
					} else {
						msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
					}
					model, _ := u.Update(msg)
					u = model.(*UI)
					u.View()
				}
				reportHeap(b, u)
				b.ReportMetric(float64(u.chat.drawDecodes)/float64(b.N), "chat-decodes/frame")
			})

			b.Run("keypress_scroll", func(b *testing.B) {
				u := auditUI(b, n)
				u.frames = nil
				u.chat.ScrollToTop()
				u.View()
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; b.Loop(); i++ {
					if i%2 == 0 {
						u.chat.ScrollToTop()
					}
					model, _ := u.Update(tea.KeyPressMsg{Code: tea.KeyDown})
					u = model.(*UI)
					u.View()
				}
				reportHeap(b, u)
			})
		})
	}
}

// BenchmarkAuditFrameCached measures the same keypress with the frame cache
// enabled: if it is not dramatically cheaper than view_uncached, the cache is
// being invalidated on the keypress path.
func BenchmarkAuditFrameCached(b *testing.B) {
	u := auditUI(b, 200)
	u.chat.ScrollToTop()
	u.View()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if i%2 == 0 {
			u.chat.ScrollToTop()
		}
		model, _ := u.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		u = model.(*UI)
		u.View()
	}
	b.ReportMetric(float64(u.frames.hits), "cache-hits")
	b.ReportMetric(float64(u.frames.misses), "cache-misses")
}

func reportHeap(b *testing.B, u *UI) {
	b.Helper()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	b.ReportMetric(float64(m.HeapAlloc), "heap-B")
}

// BenchmarkAuditParts breaks a frame into its contributors so a regression
// points at the guilty stage instead of just a bigger total.
func BenchmarkAuditParts(b *testing.B) {
	b.Run("draw", func(b *testing.B) {
		u := auditUI(b, 200)
		u.beginFrameUpdate()
		canvas := u.frameCanvas()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = u.Draw(canvas, canvas.Bounds())
		}
	})

	b.Run("buffer_render_trim", func(b *testing.B) {
		u := auditUI(b, 200)
		u.beginFrameUpdate()
		canvas := u.frameCanvas()
		_ = u.Draw(canvas, canvas.Bounds())
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = frameContent(canvas.Render())
		}
	})

	b.Run("sidebar", func(b *testing.B) {
		u := auditUI(b, 200)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			u.updateSidebarScrollState()
		}
	})

	b.Run("pills", func(b *testing.B) {
		u := auditUI(b, 200)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			u.renderPills()
		}
	})

	b.Run("update_size", func(b *testing.B) {
		u := auditUI(b, 200)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			u.updateSize()
		}
	})
}

// TestFrameContentMatchesLegacy pins frameContent to the ReplaceAll + Split +
// TrimRight + Join pipeline it replaced, including CRLF and blank-line edges.
func TestFrameContentMatchesLegacy(t *testing.T) {
	t.Parallel()
	cases := []string{
		"", "a", "a  ", "a\r\nb", "a\r\r\nb  \n", "x \n y", "\n\n",
		"no trailing", "trailing  \r\n", "mixed\r\nlines\nhere  ", "lone\r", "a\rb",
	}
	for _, in := range cases {
		want := strings.ReplaceAll(in, "\r\n", "\n")
		lines := strings.Split(want, "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " ")
		}
		want = strings.Join(lines, "\n")
		require.Equal(t, want, frameContent(in), "input %q", in)
	}
}

// fullFrame composes the current state the way View used to: a fresh buffer,
// a full draw and a full render.
func fullFrame(u *UI) string {
	scr := uv.NewScreenBuffer(u.width, u.height)
	u.invalidateBlits()
	_ = u.Draw(scr, scr.Bounds())
	// The draw above painted a scratch buffer, not the View canvas, so the
	// blits must not claim the real canvas is up to date.
	u.invalidateBlits()
	return frameContent(scr.Render())
}

// TestFrameRowsMatchFullRender guards the per-row frame cache: the
// incrementally rendered View must equal a full composition of the same state
// through typing, scrolling and resizes.
func TestFrameRowsMatchFullRender(t *testing.T) {
	u := newFrameTestUI(t)
	u.com.Workspace = &testWorkspace{cfg: &config.Config{Options: &config.Options{}}}
	u.header = newHeader(u.com)
	u.width, u.height = 120, 40
	u.frames = nil // Exercise row-level reuse, not the whole-frame cache.
	u.chat.SetMessages(buildSession(6)...)
	u.updateLayoutAndSize()

	check := func(what string) {
		t.Helper()
		require.Equal(t, fullFrame(u), u.View().Content, "frame mismatch after %s", what)
	}
	check("initial")

	for i := range 3 {
		model, _ := u.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
		u = model.(*UI)
		check(fmt.Sprintf("typing %d", i))
	}
	for i := range 3 {
		model, _ := u.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		u = model.(*UI)
		check(fmt.Sprintf("scroll %d", i))
	}
	model, _ := u.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	u = model.(*UI)
	check("resize smaller")
	model, _ = u.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	u = model.(*UI)
	check("resize larger")
}

// BenchmarkAuditPartsBottomUp measures the pieces a steady-state frame spends
// its time in, with a session loaded so the sidebar is real.
func BenchmarkAuditPartsBottomUp(b *testing.B) {
	run := func(name string, fn func(u *UI, canvas uv.ScreenBuffer)) {
		b.Run(name, func(b *testing.B) {
			u := auditUI(b, 200)
			u.beginFrameUpdate()
			canvas := u.frameCanvas()
			_ = u.Draw(canvas, canvas.Bounds())
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				fn(u, canvas)
			}
		})
	}
	run("chat_draw", func(u *UI, c uv.ScreenBuffer) { u.chat.Draw(c, u.layout.main) })
	run("chat_list_render", func(u *UI, c uv.ScreenBuffer) { _ = u.chat.list.Render() })
	run("sidebar_scroll_state", func(u *UI, c uv.ScreenBuffer) { u.updateSidebarScrollState() })
	run("status_draw", func(u *UI, c uv.ScreenBuffer) { u.status.Draw(c, u.layout.status) })
	run("editor_render", func(u *UI, c uv.ScreenBuffer) { _ = u.renderEditorView(u.layout.editor.Dx()) })
	run("rows_render", func(u *UI, c uv.ScreenBuffer) { _ = u.rows.render(c) })
}

// BenchmarkAuditKeypressLatency is the felt cost: one keypress through Update
// and the whole View. The metrics report how much of the screen actually had
// to be re-rendered and whether the chat had to re-decode its ANSI output.
func BenchmarkAuditKeypressLatency(b *testing.B) {
	for _, n := range []int{100, 200} {
		b.Run(fmt.Sprintf("msgs=%d", n*2), func(b *testing.B) {
			u := auditUI(b, n)
			u.frames = nil
			u.focus = uiFocusEditor
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				var msg tea.KeyPressMsg
				if i%2 == 0 {
					msg = tea.KeyPressMsg{Code: 'a', Text: "a"}
				} else {
					msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
				}
				model, _ := u.Update(msg)
				u = model.(*UI)
				u.frames = nil // A keypress invalidates the frame cache.
				u.View()
			}
			b.ReportMetric(float64(u.chat.drawDecodes)/float64(b.N), "chat-decodes/frame")
			b.ReportMetric(float64(u.rows.rendered)/float64(b.N), "rows/frame")
			reportHeap(b, u)
		})
	}
}

// TestChatBlitSkipsUnchangedFrame proves the steady-state frame does not
// repaint the chat: once the decoded buffer is on screen, drawing the same
// content again must leave every cell untouched so the row cache can reuse it.
func TestChatBlitSkipsUnchangedFrame(t *testing.T) {
	u := auditUI(t, 20)
	canvas := u.frameCanvas()
	u.chat.Draw(canvas, u.layout.main) // paint the chat once

	clearTouched(canvas)
	u.chat.Draw(canvas, u.layout.main)
	require.Zero(t, dirtyRows(canvas), "an unchanged chat must not repaint any row")

	// Scrolling changes the content, so the chat has to repaint.
	u.chat.ScrollToTop()
	clearTouched(canvas)
	u.chat.Draw(canvas, u.layout.main)
	require.NotZero(t, dirtyRows(canvas), "scrolling must repaint the chat")
}

func clearTouched(canvas uv.ScreenBuffer) {
	for i := range canvas.Touched {
		canvas.Touched[i] = nil
	}
}

func dirtyRows(canvas uv.ScreenBuffer) int {
	dirty := 0
	for _, touched := range canvas.Touched {
		if touched != nil {
			dirty++
		}
	}
	return dirty
}

// TestSidebarMemoTracksInputs guards the sidebar content memo: identical
// inputs must reuse the rendered content, and a changed input must rebuild it.
func TestSidebarMemoTracksInputs(t *testing.T) {
	u := auditUI(t, 5)
	u.updateSidebarScrollState()
	first := u.sidebarContent
	require.NotEmpty(t, first)

	u.updateSidebarScrollState()
	require.Equal(t, first, u.sidebarContent, "unchanged inputs must reuse the content")

	u.session.Title = "A brand new session title"
	u.updateSidebarScrollState()
	require.NotEqual(t, first, u.sidebarContent, "the memo must notice a new title")
}
