#!/usr/bin/env python3
"""Turn the launcher's rendered splash frames into an animated GIF.

Reads the raw stream produced by `cargo run --release --example
splash_filmstrip`, rasterizes each cell with a real monospace font so box
drawing and symbols look like they do in a terminal, and writes a looping GIF.

    cargo run --release --example splash_filmstrip -- /tmp/splash.raw 100 30 3000
    python3 scripts/splash_preview.py /tmp/splash.raw /tmp/splash.gif

Only used to eyeball the banner during development; it is not part of any
shipped binary.
"""

from __future__ import annotations

import os
import re
import sys
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

# A terminal cell is wider than it is tall. 9x18 approximates the usual
# ~0.5 advance-width-to-line-height ratio of a monospace terminal font.
CELL_W, CELL_H = 9, 18
BG = (0, 0, 0)

FONT_CANDIDATES = [
    "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
    "/usr/share/fonts/TTF/DejaVuSansMono.ttf",
    "/usr/share/fonts/dejavu/DejaVuSansMono.ttf",
    "/tmp/fonts/DejaVuSansMono.ttf",
    "/Library/Fonts/Menlo.ttc",
    "/System/Library/Fonts/Menlo.ttc",
]


def load_font(size: int) -> ImageFont.FreeTypeFont:
    # `STAR_FONT` lets a machine without the usual system font paths point at
    # any monospace TTF with box-drawing and symbol coverage.
    candidates = list(FONT_CANDIDATES)
    if override := os.environ.get("STAR_FONT"):
        candidates.insert(0, override)
    for candidate in candidates:
        if Path(candidate).exists():
            return ImageFont.truetype(candidate, size)
    raise SystemExit("no monospace font found; install DejaVu Sans Mono")


CELL_RE = re.compile(r"(\d+)\|(\d+)\|(\d+)\|(\d+)\|(\d+)\|(\d+)\|(.)")


def read_frames(path: Path):
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "STARRSPL1":
        raise SystemExit("not a splash frame stream")
    cols, rows = int(lines[1]), int(lines[2])

    # Each frame is a single line holding cols*rows cells. Scan rather than
    # split: the trailing character field may itself be a space.
    body = lines[3:]
    if not body:
        raise SystemExit("frame stream is empty")

    for number, block in enumerate(body, start=1):
        frame = []
        for match in CELL_RE.finditer(block):
            br, bg, bb, fr, fg, fb, ch = match.groups()
            frame.append(
                ((int(br), int(bg), int(bb)), (int(fr), int(fg), int(fb)), ch)
            )
        if len(frame) != cols * rows:
            raise SystemExit(
                f"frame {number}: got {len(frame)} cells, expected {cols * rows}"
            )
        yield frame


def render_frame(frame, cols: int, rows: int, font) -> Image.Image:
    image = Image.new("RGB", (cols * CELL_W, rows * CELL_H), BG)
    draw = ImageDraw.Draw(image)

    for index, ((br, bg_, bb), (fr, fg, fb), ch) in enumerate(frame):
        if ch == " " and (br, bg_, bb) == BG:
            continue
        x = (index % cols) * CELL_W
        y = (index // cols) * CELL_H
        if (br, bg_, bb) != BG:
            draw.rectangle([x, y, x + CELL_W - 1, y + CELL_H - 1], fill=(br, bg_, bb))
        if ch.strip():
            draw.text((x, y), ch, fill=(fr, fg, fb), font=font)

    return image


def main() -> None:
    if len(sys.argv) < 3:
        raise SystemExit("usage: splash_preview.py <frames.raw> <out.gif>")

    src = Path(sys.argv[1])
    dst = Path(sys.argv[2])

    raw = src.read_text(encoding="utf-8").splitlines()
    cols, rows = int(raw[1]), int(raw[2])
    font = load_font(CELL_H - 4)

    images = [
        render_frame(frame, cols, rows, font)
        for frame in read_frames(src)
    ]
    if not images:
        raise SystemExit("no frames in stream")

    frame_ms = 1000 // 30
    images[0].save(
        dst,
        save_all=True,
        append_images=images[1:],
        duration=frame_ms,
        loop=0,
        optimize=True,
    )

    # Also emit a single representative frame (the settled banner) as a PNG so
    # the final look can be checked without scrubbing the animation.
    settled = images[-1]
    settled.save(dst.with_suffix(".png"))

    print(
        f"{len(images)} frames at {cols}x{rows} -> {dst} "
        f"({dst.stat().st_size // 1024} KiB), settled frame -> {dst.with_suffix('.png')}"
    )


if __name__ == "__main__":
    main()
