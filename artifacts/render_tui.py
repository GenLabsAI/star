"""Render headless TUI screenshots.

Reads the per-cell JSON grids written by
internal/ui/model/tui_screenshot_test.go (TestTUIScreenshots) and renders
one PNG per grid with a real monospace font.

Usage:
    python3 artifacts/render_tui.py [artifacts_dir]

Re-run after any change to the editor fade, placeholder, or status badge.
Requires: Pillow + DejaVuSansMono.ttf next to this script.
"""

import json
import os
import sys

from PIL import Image, ImageDraw, ImageFont

FONT_SIZE = 15
PAD = 12


def main() -> None:
    outdir = sys.argv[1] if len(sys.argv) > 1 else os.path.dirname(os.path.abspath(__file__))
    font = ImageFont.truetype(os.path.join(outdir, "DejaVuSansMono.ttf"), FONT_SIZE)
    # Monospace cell metrics from the font itself.
    bbox = font.getbbox("M")
    cell_w = bbox[2] - bbox[0]
    ascent, descent = font.getmetrics()
    cell_h = ascent + descent + 2

    for name in sorted(os.listdir(outdir)):
        if not (name.startswith("tui-") and name.endswith(".json")):
            continue
        with open(os.path.join(outdir, name), encoding="utf-8") as fh:
            rows = [r or [] for r in json.load(fh)]
        width = max((len(r) for r in rows), default=0)
        img = Image.new("RGB", (width * cell_w + PAD * 2, len(rows) * cell_h + PAD * 2), "#101014")
        draw = ImageDraw.Draw(img)
        for y, row in enumerate(rows):
            for x, cell in enumerate(row):
                ch, fg, bg = cell["ch"], cell["fg"] or "#e5e5e5", cell["bg"] or "#101014"
                x0, y0 = PAD + x * cell_w, PAD + y * cell_h
                if bg.lower() != "#101014":
                    draw.rectangle([x0, y0, x0 + cell_w, y0 + cell_h], fill=bg)
                if ch != " ":
                    draw.text((x0, y0), ch, font=font, fill=fg)
        png = os.path.join(outdir, name[:-len(".json")] + ".png")
        img.save(png)
        print(f"wrote {os.path.basename(png)} ({img.width}x{img.height})")


if __name__ == "__main__":
    main()
