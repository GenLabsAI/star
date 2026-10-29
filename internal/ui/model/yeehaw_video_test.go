package model

// Video-frame dumper for the yeehaw horse. Guarded by TUI_VIDEO=1 so
// normal test runs stay fast.
//
//	TUI_VIDEO=1 go test ./internal/ui/model -run TestYeehawHorseVideo -count=1 -v
//	python3 artifacts/render_video.py
//
// Writes one JSON cell grid per tick into artifacts/video/, which
// render_video.py assembles into artifacts/yeehaw-horse.gif at the real
// 120ms frame interval.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestYeehawHorseVideo(t *testing.T) {
	if os.Getenv("TUI_VIDEO") == "" {
		t.Skip("set TUI_VIDEO=1 to dump video frames")
	}
	dir := filepath.Join(TUIShotDir(t), "video")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	const (
		width  = 100
		height = 30
		// 40 ticks = one full left-to-right crossing at 3 cells/tick
		// and exactly five 8-frame gait cycles, so the loop is seamless.
		frames = 40
	)
	u := NewScreenshotUI(t)
	shotSize(u, width, height)
	u.setYeehawMode(true)
	u.yeehawFadeFrames = 0

	for i := 0; i < frames; i++ {
		u.yeehawHorseTick = i
		v := u.View()
		plain := ansi.Strip(v.Content)
		name := "frame-" + strconv.Itoa(i)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name+".txt"), []byte(plain), 0o644))
		grid := shotParseCells(v.Content, TUIShotDefFg, TUIShotDefBg)
		raw, err := json.Marshal(grid)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name+".json"), raw, 0o644))
	}
	t.Logf("wrote %d frames to %s", frames, dir)
}
