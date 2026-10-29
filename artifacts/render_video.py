"""Assemble artifacts/yeehaw-horse.mp4 and artifacts/yeehaw-horse.gif.

Reads the per-tick JSON cell grids written by
TestYeehawHorseVideo (artifacts/video/frame-*.json) and stitches them
into a looping MP4 (plus a GIF and per-frame PNGs) at the real horse
tick interval. All frames share one fixed canvas so players never see
the image jump between frames.

Usage:
    TUI_VIDEO=1 go test ./internal/ui/model -run TestYeehawHorseVideo -count=1
    python3 artifacts/render_video.py [video_dir] [--out mp4_path]

Requires: Pillow. ffmpeg is used when found (PATH or the imageio-ffmpeg
wheel); without it only the GIF/PNGs are written.
The DejaVuSansMono.ttf next to this script is reused and downloaded if missing
(same as render_tui.py).
"""

import argparse
import glob
import json
import os
import re
import shutil
import subprocess
import sys
import urllib.request

FONT_SIZE = 13
PAD = 10
FRAME_MS = 120
FONT_URL = "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSansMono.ttf"


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("video_dir", nargs="?", default=os.path.join(here, "video"))
    ap.add_argument("--out", default=os.path.join(here, "yeehaw-horse.mp4"))
    ap.add_argument("--font-size", type=int, default=FONT_SIZE)
    ap.add_argument("--duration", type=int, default=FRAME_MS)
    args = ap.parse_args()

    # Numeric order: lexicographic sorting would play frame-10 right after
    # frame-1, teleporting the horse back to the left edge mid-crossing.
    def frame_index(path):
        m = re.search(r"frame-(\d+)\.json$", path)
        return int(m.group(1)) if m else 0

    names = sorted(glob.glob(os.path.join(args.video_dir, "frame-*.json")), key=frame_index)
    if not names:
        sys.exit(
            f"no frame-*.json in {args.video_dir!r} "
            "(run TestYeehawHorseVideo with TUI_VIDEO=1 first)"
        )

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
        img = Image.new(
            "RGB",
            (width * cell_w + PAD * 2, height * cell_h + PAD * 2),
            "#101014",
        )
        draw = ImageDraw.Draw(img)
        for y, row in enumerate(grid):
            for x, cell in enumerate(row):
                ch = cell["ch"]
                fg = cell["fg"] or "#e5e5e5"
                bg = cell["bg"] or "#101014"
                x0, y0 = PAD + x * cell_w, PAD + y * cell_h
                if bg.lower() != "#101014":
                    draw.rectangle([x0, y0, x0 + cell_w, y0 + cell_h], fill=bg)
                if ch != " ":
                    draw.text((x0, y0), ch, font=font, fill=fg)
        imgs.append(img)

    # Uniform per-frame PNGs (same fixed canvas) beside the JSON sources.
    for idx, img in zip([frame_index(n) for n in names], imgs):
        img.save(os.path.join(args.video_dir, f"frame-{idx}.png"))

    gif_out = args.out.rsplit(".", 1)[0] + ".gif"
    imgs[0].save(
        gif_out, save_all=True, append_images=imgs[1:], duration=args.duration, loop=0
    )
    print(f"wrote {gif_out} ({len(imgs)} frames, {os.path.getsize(gif_out) // 1024} KiB)")

    ffmpeg = shutil.which("ffmpeg")
    if ffmpeg is None:
        try:
            import imageio_ffmpeg

            ffmpeg = imageio_ffmpeg.get_ffmpeg_exe()
        except Exception:
            ffmpeg = None
    if ffmpeg is None:
        print("ffmpeg not found; skipped MP4")
        return

    fps = 1000 // (args.duration or 1)
    w, h = imgs[0].size
    cmd = [
        ffmpeg,
        "-y",
        "-f",
        "rawvideo",
        "-pix_fmt",
        "rgb24",
        "-s",
        f"{w}x{h}",
        "-framerate",
        str(fps),
        "-i",
        "pipe:0",
        "-vf",
        "scale=trunc(iw/2)*2:trunc(ih/2)*2",
        "-c:v",
        "libx264",
        "-pix_fmt",
        "yuv420p",
        "-movflags",
        "+faststart",
        args.out,
    ]
    proc = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE,
        text=False,
    )
    for img in imgs:
        proc.stdin.write(img.tobytes())
    proc.stdin.close()
    if proc.wait() != 0:
        err = proc.stderr.read().decode(errors="replace").strip()
        sys.exit(f"ffmpeg failed: {err}")

    size = os.path.getsize(args.out)
    print(f"wrote {args.out} ({len(imgs)} frames, {size // 1024} KiB)")


if __name__ == "__main__":
    main()
