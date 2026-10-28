"""Render headless TUI screenshots.

Reads the per-cell JSON grids written by
internal/ui/model/tui_screenshot_test.go (TestTUIScreenshots /
CaptureTUIShots) and renders one PNG per grid with a monospace font.

Usage:
    python3 artifacts/render_tui.py [options] [input_dir]

    python3 artifacts/render_tui.py artifacts/
    python3 artifacts/render_tui.py --pattern 'yeehaw*' --font-size 16 artifacts/
    TUI_SHOT_FILTER=yeehaw go test ./internal/ui/model \\
        -run TestTUIScreenshots -count=1 -v
    python3 artifacts/render_tui.py --pattern 'tui-yeehaw*'

Options mirror the generic harness: the input directory defaults to the
artifacts/ folder next to this script, the output directory defaults to
the input directory, and --pattern filters which grids get rendered.

Requires: Pillow. The font defaults to DejaVuSansMono.ttf next to this
script and is downloaded automatically if missing.
"""

import argparse
import fnmatch
import json
import os
import sys
import urllib.request

DEFAULT_FONT_SIZE = 15
DEFAULT_PAD = 12
DEFAULT_FG = "#e5e5e5"
DEFAULT_BG = "#101014"
FONT_URL = (
    "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/"
    "DejaVuSansMono.ttf"
)


def parse_args(argv=None):
    here = os.path.dirname(os.path.abspath(__file__))
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument(
        "input_dir",
        nargs="?",
        default=here,
        help="Directory holding tui-*.json grids (default: %(default)s).",
    )
    p.add_argument(
        "--out-dir",
        default=None,
        help="Where to write PNGs (default: same as input dir).",
    )
    p.add_argument(
        "--font",
        default=None,
        help="TTF path (default: DejaVuSansMono.ttf next to this script).",
    )
    p.add_argument("--font-size", type=int, default=DEFAULT_FONT_SIZE)
    p.add_argument("--pad", type=int, default=DEFAULT_PAD)
    p.add_argument(
        "--pattern",
        default="tui-*.json",
        help="Glob for grid files (default: %(default)s).",
    )
    p.add_argument("--bg", default=DEFAULT_BG)
    p.add_argument("--fg", default=DEFAULT_FG)
    return p.parse_args(argv)


def ensure_font(path):
    if os.path.exists(path):
        return path
    try:
        print(f"downloading font to {os.path.basename(path)} ...")
        urllib.request.urlretrieve(FONT_URL, path)
        return path
    except Exception as exc:  # noqa: BLE001 - report and exit cleanly.
        sys.exit(f"missing font {path}: {exc}\n"
                 "install DejaVuSansMono.ttf next to this script.")


def main(argv=None) -> None:
    args = parse_args(argv)
    indir = os.path.abspath(args.input_dir)
    outdir = os.path.abspath(args.out_dir or indir)
    os.makedirs(outdir, exist_ok=True)
    font_path = args.font or os.path.join(
        os.path.dirname(os.path.abspath(__file__)), "DejaVuSansMono.ttf")
    ensure_font(font_path)

    from PIL import Image, ImageDraw, ImageFont

    font = ImageFont.truetype(font_path, args.font_size)
    # Monospace cell metrics from the font itself.
    bbox = font.getbbox("M")
    cell_w = bbox[2] - bbox[0]
    ascent, descent = font.getmetrics()
    cell_h = ascent + descent + 2

    names = sorted(n for n in os.listdir(indir)
                   if fnmatch.fnmatch(n, args.pattern))
    if not names:
        sys.exit(f"no files matching {args.pattern!r} in {indir}")
    for name in names:
        with open(os.path.join(indir, name), encoding="utf-8") as fh:
            rows = [r or [] for r in json.load(fh)]
        width = max((len(r) for r in rows), default=0)
        img = Image.new(
            "RGB",
            (width * cell_w + args.pad * 2,
             len(rows) * cell_h + args.pad * 2),
            args.bg,
        )
        draw = ImageDraw.Draw(img)
        for y, row in enumerate(rows):
            for x, cell in enumerate(row):
                ch = cell["ch"]
                fg = cell["fg"] or args.fg
                bg = cell["bg"] or args.bg
                x0, y0 = (args.pad + x * cell_w, args.pad + y * cell_h)
                if bg.lower() != args.bg.lower():
                    draw.rectangle(
                        [x0, y0, x0 + cell_w, y0 + cell_h], fill=bg)
                if ch != " ":
                    draw.text((x0, y0), ch, font=font, fill=fg)
        stem = name[: -len(".json")] if name.endswith(".json") else name
        png = os.path.join(outdir, stem + ".png")
        img.save(png)
        print(f"wrote {os.path.basename(png)} ({img.width}x{img.height})")


if __name__ == "__main__":
    main()
