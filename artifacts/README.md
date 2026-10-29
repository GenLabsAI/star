# TUI screenshot scaffold

Generic, reusable harness for headless screenshots of the Bubble Tea UI.
It exists so any future TUI state (modes, dialogs, animations) can be
captured and reviewed without a live terminal.

## Pieces

- `internal/ui/model/tui_screenshot_test.go` — generic Go harness.
  - `TUIShot{Name, Width, Height, Setup, Warmup}` describes one frame.
  - `NewScreenshotUI(t)` builds a fresh harness UI per shot (session,
    prompt text, layout applied) so shots stay independent and
    order-free.
  - `CaptureTUIShots(t, dir, shots)` renders each shot, advances
    `Warmup` frames first for transitions, and writes
    `tui-<name>.txt` (ANSI-stripped) plus `tui-<name>.json` (cell grid).
  - `TUIShotDir(t)` resolves the output dir: `TUI_SHOT_DIR` wins,
    otherwise `<repo>/artifacts`.
  - `TestTUIScreenshots` is just one table of shots — copy the pattern
    for new states.
- `artifacts/render_tui.py` — generic JSON-grid → PNG renderer
  (argparse: input/output dirs, `--pattern`, `--font`, `--font-size`,
  `--pad`, `--fg`, `--bg`; font auto-downloads if missing).

## Typical flow

```sh
go test ./internal/ui/model -run TestTUIScreenshots -count=1 -v
python3 artifacts/render_tui.py artifacts/
```

Iterate on one state without re-rendering everything:

```sh
TUI_SHOT_FILTER=yeehaw go test ./internal/ui/model \
  -run TestTUIScreenshots -count=1 -v
python3 artifacts/render_tui.py --pattern 'tui-yeehaw*' artifacts/
```

## Adding a new state

Append one entry to the `shots` table in `TestTUIScreenshots`:

```go
{Name: "mydialog-140", Width: 140, Height: 45, Setup: func(u *UI) {
    // Put the UI into the state to capture.
}},
```

Use `Warmup: N` to sample N frames into a transition (fades, spinners).
Keep `Name` stable: the renderer turns `tui-<name>.json` into
`tui-<name>.png`.

## What's tracked

- Tracked: this README, `render_tui.py`, `tui_screenshot_test.go`, and
  the review-friendly `tui-*.png` / `tui-*.txt` snapshots.
- Ignored (regenerable): `tui-*.json` grids and the downloaded
  `DejaVuSansMono.ttf` font.

## Horse video

The yeehaw horse loop is captured the same way, one JSON grid per tick:

```sh
TUI_VIDEO=1 go test ./internal/ui/model -run TestYeehawHorseVideo -count=1
python3 artifacts/render_video.py --out artifacts/yeehaw-horse.mp4
```

`internal/ui/model/yeehaw_video_test.go` dumps `video/frame-<tick>.json`
(`TUI_VIDEO=1` keeps normal test runs fast); `render_video.py` renders
all frames on one fixed canvas — ordered by numeric tick, never by
filename — and writes `yeehaw-horse.mp4`, `yeehaw-horse.gif`, plus
uniform per-frame PNGs. The tracked `.mp4`/`.gif` are the reviewable
outputs; everything under `video/` is ignored and regenerable.
