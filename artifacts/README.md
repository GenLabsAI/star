# TUI screenshot and walkthrough scaffold

Generic, reusable harness for headless screenshots of the Bubble Tea UI.
It captures feature states without a live terminal and turns an ordered set
of stills into a captioned, looping walkthrough. Walkthrough media is review
output: generate it, inspect it, share it, and leave it untracked.

## Pieces

- `internal/ui/model/tui_screenshot_test.go` — generic Go harness.
  - `TUIShot{Name, Width, Height, Setup, Warmup}` describes one frame.
  - `NewScreenshotUI(t)` builds a fresh harness UI per shot (session,
    prompt text, layout applied) so shots stay independent and order-free.
  - `CaptureTUIShots(t, dir, shots)` renders each shot, advances
    `Warmup` frames first for transitions, and writes
    `tui-<name>.txt` (ANSI-stripped) plus `tui-<name>.json` (cell grid).
  - `TUIShotDir(t)` resolves the output dir: `TUI_SHOT_DIR` wins,
    otherwise `<repo>/artifacts`.
  - `TestTUIScreenshots` is one table of shots; feature tests such as
    `TestForkScreenshots` keep their own table.
- `artifacts/render_tui.py` — generic JSON-grid → PNG renderer
  (argparse: input/output dirs, `--pattern`, `--font`, `--font-size`,
  `--pad`, `--fg`, `--bg`; font auto-downloads if missing).
- `artifacts/render_flow.py` — manifest-driven walkthrough renderer.
  It bakes short captions into a header strip and writes a looping GIF,
  plus an MP4 with `--mp4` when ffmpeg is available.
- `artifacts/flows/` — one JSON manifest per walkthrough. `fork.json` is a
  worked example to copy.
- `artifacts/tui_font.py` — shared font resolver for all renderers.

## Agent entry point

Agents should load `.agents/skills/tui-walkthrough/SKILL.md` before doing
visual proof. The skill is project-local, versioned with the repository, and
links back here for the full harness and renderer details.

## Typical screenshot flow

```sh
go test ./internal/ui/model -run TestTUIScreenshots -count=1 -v
python3 artifacts/render_tui.py --pattern 'tui-*.json' artifacts/
```

Iterate on one state without re-rendering everything:

```sh
TUI_SHOT_FILTER=yeehaw go test ./internal/ui/model \
  -run TestTUIScreenshots -count=1 -v
python3 artifacts/render_tui.py --pattern 'tui-yeehaw*.json' artifacts/
```

`--pattern` filters grid files only; always include `.json` so a previously
rendered PNG is never read as a grid.

## Feature walkthroughs

Capture one still per state, in the order a user encounters them, then list
the images and captions in a manifest:

```json
{
  "output": "tui-<feature>-flow",
  "width": 1000,
  "duration_ms": 2600,
  "frames": [
    {
      "image": "tui-<feature>-<state>-140.png",
      "caption": "Short description of this user-visible state."
    },
    {
      "image": "tui-<feature>-<next-state>-140.png",
      "caption": "What changed after the interaction."
    }
  ]
}
```

A complete worked manifest is `artifacts/flows/fork.json`. Render it:

```sh
go test ./internal/ui/model -run TestForkScreenshots -count=1 -v
python3 artifacts/render_tui.py --pattern 'tui-fork*.json' artifacts/
python3 artifacts/render_flow.py artifacts/flows/fork.json
# Optional actual video:
python3 artifacts/render_flow.py artifacts/flows/fork.json --mp4
```

The manifest's `output` resolves against `artifacts/`; image paths resolve
from `artifacts/` first, then the manifest's directory. Per-frame
`duration_ms` overrides the top-level value. `--width`, `--duration`,
`--font`, and `--out` are available for one-off overrides. Captions over
the available width are truncated with `...`.

The GIF is the review artifact: it embeds and animates in chat. The optional
MP4 uses ffmpeg from `PATH` or the `imageio-ffmpeg` wheel. Both are generated
proof, not source: they are git-ignored and should be attached or inspected
rather than committed. Use `render_video.py` instead for frame-by-frame
animation loops (fades, spinners, the horse), which need many captured ticks
rather than captions.

## Adding a new walkthrough

1. Add a `TUIShot` table and setup functions for each state, and keep the
   names stable (`tui-<feature>-<state>-<width>`). Prefer 140×45 and add a
   narrower shot when responsive behaviour matters.
2. Run the screenshot test and `render_tui.py` to produce the PNGs.
3. Add `artifacts/flows/<feature>.json` with the ordered frames and short
   captions.
4. Run `render_flow.py` and inspect the GIF before sharing it. Track the
   manifest, PNGs, and TXT frames; leave the GIF/MP4 untracked.
5. Treat the walkthrough as evidence, not as a behavior test. Keep the
   real integration/UI tests alongside it.

## What's tracked

- Tracked: this README, the renderer scripts, the project skill,
  `tui_screenshot_test.go`, flow manifests, and the review-friendly
  `tui-*.png` / `tui-*.txt` snapshots.
- Ignored (regenerable): `tui-*.json` grids, `tui-*.gif` / `tui-*.mp4`
  walkthrough media, and the auto-downloaded fonts.

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
