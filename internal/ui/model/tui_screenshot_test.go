package model

// Reusable headless TUI screenshot harness.
//
// Run with:
//
//	go test ./internal/ui/model -run TestTUIScreenshots -count=1 -v
//
// It renders the full UI View at fixed sizes in normal, yolo, and yeehaw
// states (including yeehaw fade frames), and writes artifacts under
// artifacts/ at the repo root:
//
//	artifacts/tui-<name>.txt   plain-text frame (ANSI stripped)
//	artifacts/tui-<name>.json   per-cell {char, fg, bg} grid for PNG rendering
//
// A companion Python script (artifacts/render_tui.py) turns the JSON grids
// into PNGs with a real monospace font. Re-run both whenever the editor
// fade, placeholder, or status badge changes.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

type shotCell struct {
	Ch string `json:"ch"`
	Fg string `json:"fg"`
	Bg string `json:"bg"`
}

var (
	shotSGRRe = regexp.MustCompile("\x1b\\[([0-9;]*)m")
	// Any other CSI sequence (cursor visibility, etc.) carries no cell
	// content, so drop it before segmenting the line.
	shotCSIRe = regexp.MustCompile("\x1b\\[[0-9;?]*[A-Za-z]")
	shotOSCRe = regexp.MustCompile("\x1b\\][^\x07]*\x07")
)

// shotANSI16 maps the 16 ANSI colors to hex.
var shotANSI16 = []string{
	"#000000", "#cd0000", "#00cd00", "#cdcd00",
	"#0000ee", "#cd00cd", "#00cdcd", "#e5e5e5",
	"#7f7f7f", "#ff0000", "#00ff00", "#ffff00",
	"#5c5cff", "#ff00ff", "#00ffff", "#ffffff",
}

func shotANSI256(n int) string {
	if n < 8 {
		return shotANSI16[n]
	}
	if n < 16 {
		return shotANSI16[n]
	}
	if n < 232 {
		n -= 16
		r := (n / 36) * 51
		g := ((n % 36) / 6) * 51
		b := (n % 6) * 51
		return "#" + hex2(r) + hex2(g) + hex2(b)
	}
	g := 8 + (n-232)*10
	return "#" + hex2(g) + hex2(g) + hex2(g)
}

func hex2(v int) string {
	const d = "0123456789abcdef"
	return string([]byte{d[(v>>4)&15], d[v&15]})
}

type shotPen struct {
	fg    string
	bg    string
	fgSet bool
	bgSet bool
}

func (p *shotPen) apply(params []int) {
	if len(params) == 0 {
		params = []int{0}
	}
	for i := 0; i < len(params); i++ {
		c := params[i]
		switch {
		case c == 0:
			*p = shotPen{}
		case c == 1:
			// Bold: brighten a plain 16-color fg instead of tracking weight.
			if p.fgSet && len(p.fg) == 7 {
				for idx, h := range shotANSI16[:8] {
					if h == p.fg {
						p.fg = shotANSI16[idx+8]
						break
					}
				}
			}
		case c == 22:
			// Normal intensity: no-op for the grid.
		case c >= 30 && c <= 37:
			p.fg, p.fgSet = shotANSI16[c-30], true
		case c >= 90 && c <= 97:
			p.fg, p.fgSet = shotANSI16[c-90+8], true
		case c == 39:
			p.fg, p.fgSet = "", false
		case c >= 40 && c <= 47:
			p.bg, p.bgSet = shotANSI16[c-40], true
		case c >= 100 && c <= 107:
			p.bg, p.bgSet = shotANSI16[c-100+8], true
		case c == 49:
			p.bg, p.bgSet = "", false
		case c == 38 || c == 48:
			isFg := c == 38
			if i+1 < len(params) && params[i+1] == 5 && i+2 < len(params) {
				if isFg {
					p.fg, p.fgSet = shotANSI256(params[i+2]), true
				} else {
					p.bg, p.bgSet = shotANSI256(params[i+2]), true
				}
				i += 2
			} else if i+1 < len(params) && params[i+1] == 2 && i+4 < len(params) {
				h := "#" + hex2(params[i+2]) + hex2(params[i+3]) + hex2(params[i+4])
				if isFg {
					p.fg, p.fgSet = h, true
				} else {
					p.bg, p.bgSet = h, true
				}
				i += 4
			}
		}
	}
}

func shotParseCells(content, defFg, defBg string) [][]shotCell {
	content = shotOSCRe.ReplaceAllString(content, "")
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	rows := make([][]shotCell, 0, len(lines))
	for _, ln := range lines {
		// Hide everything except SGR, then walk text/SGR segments.
		locs := shotSGRRe.FindAllStringSubmatchIndex(ln, -1)
		var row []shotCell
		pen := shotPen{}
		pos := 0
		flush := func(text string) {
			fg := defFg
			if pen.fgSet {
				fg = pen.fg
			}
			bg := defBg
			if pen.bgSet {
				bg = pen.bg
			}
			for _, r := range text {
				row = append(row, shotCell{Ch: string(r), Fg: fg, Bg: bg})
			}
		}
		for _, loc := range locs {
			flush(shotCSIRe.ReplaceAllString(ln[pos:loc[0]], ""))
			params := []int{}
			for _, s := range strings.Split(ln[loc[2]:loc[3]], ";") {
				if s == "" {
					params = append(params, 0)
					continue
				}
				n, err := strconv.Atoi(s)
				if err == nil {
					params = append(params, n)
				}
			}
			pen.apply(params)
			pos = loc[1]
		}
		flush(shotCSIRe.ReplaceAllString(ln[pos:], ""))
		rows = append(rows, row)
	}
	return rows
}

func newShotUI(t *testing.T) *UI {
	t.Helper()
	u := newFrameTestUI(t)
	u.header = newHeader(u.com)
	// The sidebar skills section dereferences Config().Options once a
	// session is set, so give the stub workspace a non-nil Options.
	if ws, ok := u.com.Workspace.(*testWorkspace); ok && ws.cfg != nil {
		ws.cfg.Options = &config.Options{}
	}
	u.session = &session.Session{ID: "sess-1"}
	u.textarea.SetValue("Refactor the auth module")
	u.updateLayoutAndSize()
	return u
}

func shotSize(u *UI, w, h int) {
	u.width, u.height = w, h
	u.updateLayoutAndSize()
}

func TestTUIScreenshots(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "artifacts")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	const defFg, defBg = "#e5e5e5", "#101014"

	save := func(name string, u *UI) {
		t.Helper()
		v := u.View()
		plain := ansi.Strip(v.Content)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "tui-"+name+".txt"), []byte(plain), 0o644))
		grid := shotParseCells(v.Content, defFg, defBg)
		raw, err := json.Marshal(grid)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "tui-"+name+".json"), raw, 0o644))
		t.Logf("wrote tui-%s (%d rows)", name, len(grid))
	}

	// Normal coding mode at two widths.
	u := newShotUI(t)
	shotSize(u, 140, 45)
	save("normal-140", u)
	shotSize(u, 100, 30)
	save("normal-100", u)

	// YOLO mode for comparison.
	shotSize(u, 140, 45)
	u.setPermissionMode(permission.ModeYolo)
	save("yolo-140", u)
	u.setPermissionMode(permission.ModeNormal)

	// Yeehaw: activation frame, mid-fade, and settled placeholder.
	u.setYeehawMode(true)
	save("yeehaw-fade0-140", u)
	for i := 0; i < 5; i++ {
		_ = u.View()
	}
	save("yeehaw-fade5-140", u)
	for i := 0; i < 10; i++ {
		_ = u.View()
	}
	save("yeehaw-done-140", u)

	// Settled yeehaw placeholder at narrow width.
	shotSize(u, 100, 30)
	save("yeehaw-done-100", u)
}
