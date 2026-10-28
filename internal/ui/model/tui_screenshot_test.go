package model

// Reusable headless TUI screenshot harness.
//
// This file is a generic scaffold, not a yeehaw-specific test. Add a new
// visual state by appending one entry to the shots table in
// TestTUIScreenshots — no harness changes needed.
//
// Run with:
//
//	go test ./internal/ui/model -run TestTUIScreenshots -count=1 -v
//
// It renders the full UI View at fixed sizes and writes artifacts under
// <repo>/artifacts (overridable with TUI_SHOT_DIR):
//
//	artifacts/tui-<name>.txt   plain-text frame (ANSI stripped)
//	artifacts/tui-<name>.json   per-cell {ch, fg, bg} grid for PNG rendering
//
// A companion script (artifacts/render_tui.py) turns the JSON grids into
// PNGs with a real monospace font. Re-run both whenever a visual state
// changes. Filter to a subset while iterating:
//
//	TUI_SHOT_FILTER=yeehaw go test ./internal/ui/model -run TestTUIScreenshots -count=1 -v

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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

// TUIShotSetup mutates a fresh harness UI into the state to capture
// (modes, session content, focus, etc.).
type TUIShotSetup func(*UI)

// TUIShot describes one frame to capture. Warmup renders and discards N
// frames before capturing, which is how multi-frame transitions (fades,
// spinners) are sampled without sharing one UI instance across shots.
type TUIShot struct {
	Name   string
	Width  int
	Height int
	Setup  TUIShotSetup
	Warmup int
}

// TUIShotDefFg and TUIShotDefBg are the fallback colors used for cells
// that carry no explicit SGR color.
const (
	TUIShotDefFg = "#e5e5e5"
	TUIShotDefBg = "#101014"
)

// NewScreenshotUI builds a fresh harness UI with a session, prompt text,
// and layout applied. Each shot gets its own instance so shots stay
// independent and order-free.
func NewScreenshotUI(t *testing.T) *UI {
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

// newShotUI is kept as an alias for NewScreenshotUI.
func newShotUI(t *testing.T) *UI {
	return NewScreenshotUI(t)
}

func shotSize(u *UI, w, h int) {
	u.width, u.height = w, h
	u.updateLayoutAndSize()
}

// TUIShotDir resolves where frames are written. TUI_SHOT_DIR wins when
// set; otherwise it walks up from the caller to the module root (go.mod)
// and uses <root>/artifacts.
func TUIShotDir(t *testing.T) string {
	t.Helper()
	if dir := strings.TrimSpace(os.Getenv("TUI_SHOT_DIR")); dir != "" {
		require.NoError(t, os.MkdirAll(dir, 0o755))
		return dir
	}
	_, file, _, ok := runtime.Caller(0)
	if ok {
		d := filepath.Dir(file)
		for i := 0; i < 8; i++ {
			if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
				dir := filepath.Join(d, "artifacts")
				require.NoError(t, os.MkdirAll(dir, 0o755))
				return dir
			}
			d = filepath.Dir(d)
		}
	}
	// Fallback for unusual layouts: historical relative path.
	dir := filepath.Join("..", "..", "..", "artifacts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	return dir
}

// shotFilter returns the TUI_SHOT_FILTER substrings (comma-separated).
// When non-empty, only shots whose name contains one of them run.
func shotFilter() []string {
	raw := strings.TrimSpace(os.Getenv("TUI_SHOT_FILTER"))
	if raw == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(raw, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// CaptureTUIShots renders each shot on a fresh UI and writes its .txt
// (ANSI-stripped) and .json (cell grid) frames into dir.
func CaptureTUIShots(t *testing.T, dir string, shots []TUIShot) {
	t.Helper()
	filter := shotFilter()
	for _, s := range shots {
		if len(filter) > 0 {
			matched := false
			for _, f := range filter {
				if strings.Contains(s.Name, f) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		u := NewScreenshotUI(t)
		w, h := s.Width, s.Height
		if w <= 0 {
			w = 140
		}
		if h <= 0 {
			h = 45
		}
		shotSize(u, w, h)
		if s.Setup != nil {
			s.Setup(u)
		}
		for i := 0; i < s.Warmup; i++ {
			_ = u.View()
		}
		v := u.View()
		plain := ansi.Strip(v.Content)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "tui-"+s.Name+".txt"), []byte(plain), 0o644))
		grid := shotParseCells(v.Content, TUIShotDefFg, TUIShotDefBg)
		raw, err := json.Marshal(grid)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "tui-"+s.Name+".json"), raw, 0o644))
		t.Logf("wrote tui-%s (%d rows)", s.Name, len(grid))
	}
}

func TestTUIScreenshots(t *testing.T) {
	dir := TUIShotDir(t)

	yolo := TUIShotSetup(func(u *UI) {
		u.setPermissionMode(permission.ModeYolo)
	})
	yeehaw := TUIShotSetup(func(u *UI) {
		u.setYeehawMode(true)
		// Skip the fade so the snapshot shows the lane, and freeze the
		// horse mid-lane so it reads as galloping rather than riding in.
		u.yeehawFadeFrames = 0
		u.yeehawHorseTick = 20
	})
	yeehawEnter := TUIShotSetup(func(u *UI) {
		u.setYeehawMode(true)
		u.yeehawFadeFrames = 0
		u.yeehawHorseTick = 2
	})

	// To capture a new state, append one entry here. Keep names stable:
	// render_tui.py turns tui-<name>.json into tui-<name>.png.
	shots := []TUIShot{
		{Name: "normal-140", Width: 140, Height: 45},
		{Name: "normal-100", Width: 100, Height: 30},
		{Name: "yolo-140", Width: 140, Height: 45, Setup: yolo},
		{Name: "yeehaw-fade0-140", Width: 140, Height: 45, Setup: yeehaw},
		{Name: "yeehaw-fade5-140", Width: 140, Height: 45, Setup: yeehaw, Warmup: 5},
		{Name: "yeehaw-done-140", Width: 140, Height: 45, Setup: yeehaw, Warmup: 15},
		{Name: "yeehaw-done-100", Width: 100, Height: 30, Setup: yeehaw, Warmup: 15},
		{Name: "yeehaw-horse-enter-140", Width: 140, Height: 45, Setup: yeehawEnter},
	}

	CaptureTUIShots(t, dir, shots)
}
