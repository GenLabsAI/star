"""Assemble artifacts/yeehaw-horse.gif from dumped video frames.

Reads the per-tick JSON cell grids written by
TestYeehawHorseVideo (artifacts/video/frame-*.json) and stitches them
into a looping GIF at the real horse tick interval.

Usage:
    TUI_VIDEO=1 go test ./internal/ui/model -run TestYeehawHorseVideo -count=1
    python3 artifacts/render_video.py [video_dir] [--out gif_path]

Requires: Pillow. Reuses the DejaVuSansMono.ttf next to this script,
downloading it if missing (same as render_tui.py).
"""

import argparse
import glob
import json
import os
import sys
import urllib.request

FONT_SIZE = 13
PAD = 10
FRAME_MS = 120
FONT_URL = (
    "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/"
    "DejaVuSansMono.ttf"
)


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("video_dir", nargs="?", default=os.path.join(here, "video"))
    ap.add_argument("--out", default=os.path.join(here, "yeehaw-horse.gif"))
    ap.add_argument("--font-size", type=int, default=FONT_SIZE)
    ap.add_argument("--duration", type=int, default=FRAME_MS)
    args = ap.parse_args()

    names = sorted(glob.glob(os.path.join(args.video_dir, "frame-*.json")))
    if not names:
        sys.exit(f"no frame-*.json in {args.video_dir} "
                 "(run TestYeehawHorseVideo with TUI_VIDEO=1 first)")
    font_path = os.path.join(here, "DejaVuSansMono.ttf")
    if not os.path.exists(font_path):
        print("downloading font ...")
        urllib.request.urlretrieve(FONT_URL, font_path)

    from PIL import Image, ImageDraw, ImageFont

    font = ImageFont.truetype(font_path, args.font_size)
    bbox = font.getbbox("M")
    cell_w = bbox[2] - bbox[0]
    ascent, descent = font.getmetrics()
    cell_h = ascent + descent + 2

    grids = []
    for name in names:
        with open(name, encoding="utf-8") as fh:
            grids.append([r or [] for r in json.load(fh)])
    width = max(max((len(r) for r in g), default=0) for g in grids)
    height = max(len(g) for g in grids)

    imgs = []
    for grid in grids:
        img = Image.new("RGB", (width * cell_w + PAD * 2,
                                height * cell_h + PAD * 2), "#101014")
        draw = ImageDraw.Draw(img)
        for y, row in enumerate(grid):
            for x, cell in enumerate(row):
                ch, fg, bg = cell["ch"], cell["fg"] or "#e5e5e5", cell["bg"] or "#101014"
                x0, y0 = PAD + x * cell_w, PAD + y * cell_h
                if bg.lower() != "#101014":
                    draw.rectangle([x0, y0, x0 + cell_w, y0 + cell_h], fill=bg)
                if ch != " ":
                    draw.text((x0, y0), ch, font=font, fill=fg)
        imgs.append(img.convert("P", palette=Image.ADAPTIVE))

    imgs[0].save(args.out, save_all=True, append_images=imgs[1:],
                 duration=args.duration, loop=0)
    size = os.path.getsize(args.out)
    print(f"wrote {args.out} ({len(imgs)} frames, {size // 1024} KiB)")


if __name__ == "__main__":
    main()
