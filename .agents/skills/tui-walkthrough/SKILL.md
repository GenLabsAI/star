---
name: tui-walkthrough
description: Use when a Star TUI change needs credible visual proof, including capturing real UI.View() screenshots, rendering screenshot grids to PNGs, or assembling an ordered captioned GIF/MP4 walkthrough for review.
---

# TUI Walkthroughs

Build visual evidence from the real Bubble Tea render path. Do not hand-draw
mockups or substitute a description for a rendered state.

## Read first

- `internal/ui/AGENTS.md` for rendering and layout invariants.
- `artifacts/README.md` for the harness and renderer details.
- `artifacts/flows/fork.json` for a complete manifest example.

## Workflow

Replace `<feature>` and `<Feature>` with the feature name in every command.

1. Add a screenshot test in `internal/ui/model`, modeled on
   `TestForkScreenshots` and `tui_screenshot_test.go`. Use one `TUIShot` per
   user-visible state; keep names stable
   (`tui-<feature>-<state>-<width>`). Capture at 140x45, and add 100x30 when
   responsive behavior matters.
2. Run only that test:
   `go test ./internal/ui/model -run Test<Feature>Screenshots -count=1 -v`
3. Render grids, never PNGs, with the pattern ending in `.json`:
   `python3 artifacts/render_tui.py --pattern 'tui-<feature>*.json' artifacts/`
4. Inspect the PNGs. Fix layout or state bugs before making a walkthrough.
5. Add `artifacts/flows/<feature>.json` with the ordered frames and a short
   caption per state. Copy the shape from `fork.json`. `output` resolves under
   `artifacts/`; `width` defaults to 1000 and `duration_ms` to 2500.
6. Render and inspect:
   `python3 artifacts/render_flow.py artifacts/flows/<feature>.json`
   Add `--mp4` when ffmpeg or the `imageio-ffmpeg` wheel is available.
7. Verify the GIF: frame order, captions, selection indicators, no clipped
   text, and the final state. Re-capture any wrong frame.
8. Attach the GIF (and MP4 if useful) to the response. They are git-ignored
   review artifacts; commit the manifest, PNGs, and TXT frames instead.
9. Run the feature's behavior tests and `go build ./...`. The walkthrough is
   evidence, not a substitute for tests.

## Manifest

```json
{
  "output": "tui-<feature>-flow",
  "width": 1000,
  "duration_ms": 2600,
  "frames": [
    {
      "image": "tui-<feature>-<state>-140.png",
      "caption": "Short description of this user-visible state."
    }
  ]
}
```

Image paths resolve from `artifacts/` first, then the manifest directory.
Set `"duration_ms"` per frame when needed, or override all frames with
`--duration`. Keep captions short; text that does not fit is ellipsized.

## Rules

- Use the production `UI.View()` through `NewScreenshotUI`; no mock renderer.
- One state per frame, in the order a user encounters them.
- Do not commit `tui-*.json`, `tui-*.gif`, or `tui-*.mp4`.
- Keep behavior tests separate from visual proof.
- The font auto-downloads when missing; `--font` overrides it.
- Without ffmpeg the GIF still renders and MP4 generation is skipped.
