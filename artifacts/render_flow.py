#!/usr/bin/env python3
"""Render a captioned feature walkthrough from TUI screenshots.

A walkthrough manifest orders the still images and gives each one a short
caption:

    {
      "output": "tui-feature-flow",
      "width": 1000,
      "duration_ms": 2600,
      "frames": [
        {"image": "tui-feature-state-140.png",
         "caption": "Short description of this user-visible state."},
        ...
      ]
    }

Image names resolve from this script's directory first, then the manifest's
directory. The renderer writes a looping GIF next to the images, and can also
write an MP4 when ffmpeg is available. See artifacts/flows/fork.json for a
complete worked manifest.

Usage:
    python3 artifacts/render_flow.py artifacts/flows/fork.json
    python3 artifacts/render_flow.py artifacts/flows/fork.json --mp4
    python3 artifacts/render_flow.py artifacts/flows/fork.json --out /tmp/demo.gif

Requires: Pillow. The shared tui_font helper downloads a monospace TTF on
first use.
"""

import argparse
import json
import os
import shutil
import subprocess
import sys

from tui_font import resolve_font

HERE = os.path.dirname(os.path.abspath(__file__))
DEFAULT_WIDTH = 1000
DEFAULT_DURATION_MS = 2500
DEFAULT_FONT_SIZE = 21
DEFAULT_FG = "#f2f2f5"
DEFAULT_BG = "#101014"
DEFAULT_MUTED = "#8d8d99"
ACCENT = "#7d5cff"
PAD = 24
ACCENT_H = 4


def parse_args(argv=None):
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("manifest", help="JSON walkthrough manifest.")
    p.add_argument(
        "--out",
        default=None,
        help="Output .gif path or stem (default: manifest output/name).",
    )
    p.add_argument(
        "--width",
        type=int,
        default=None,
        help="Frame width in pixels (default: manifest width or 1000).",
    )
    p.add_argument(
        "--duration",
        type=int,
        default=None,
        help="Override every frame duration in milliseconds.",
    )
    p.add_argument(
        "--font",
        default=None,
        help="Monospace TTF for captions (default: auto-downloaded).",
    )
    p.add_argument("--font-size", type=int, default=DEFAULT_FONT_SIZE)
    p.add_argument(
        "--mp4",
        action="store_true",
        help="Also write an MP4 beside the GIF when ffmpeg is available.",
    )
    return p.parse_args(argv)


def resolve_image(name, manifest_dir):
    if os.path.isabs(name):
        candidates = [name]
    else:
        candidates = [
            os.path.join(HERE, name),
            os.path.join(manifest_dir, name),
            os.path.abspath(name),
        ]
    for path in candidates:
        if os.path.isfile(path):
            return path
    sys.exit(
        f"image not found: {name!r}\nlooked in:\n  "
        + "\n  ".join(candidates)
    )


def resolve_output(args, spec, manifest_path):
    raw = args.out or spec.get("output")
    if not raw:
        stem = os.path.splitext(os.path.basename(manifest_path))[0]
        raw = f"tui-{stem}-flow"
    if not os.path.isabs(raw):
        # Manifest outputs are artifact-relative; CLI outputs are cwd-relative.
        base = os.getcwd() if args.out else HERE
        raw = os.path.join(base, raw)
    stem, ext = os.path.splitext(raw)
    if ext.lower() == ".gif":
        return raw
    if ext:
        sys.exit("--out must be a .gif path or a filename stem")
    return raw + ".gif"


def fit_text(draw, text, font, max_width):
    if draw.textlength(text, font=font) <= max_width:
        return text
    ellipsis = "..."
    lo, hi = 0, len(text)
    while lo < hi:
        mid = (lo + hi + 1) // 2
        candidate = text[:mid] + ellipsis
        if draw.textlength(candidate, font=font) <= max_width:
            lo = mid
        else:
            hi = mid - 1
    return text[:lo].rstrip() + ellipsis


def find_ffmpeg():
    ffmpeg = shutil.which("ffmpeg")
    if ffmpeg:
        return ffmpeg
    try:
        import imageio_ffmpeg

        return imageio_ffmpeg.get_ffmpeg_exe()
    except Exception:  # noqa: BLE001 - optional dependency.
        return None


def write_mp4(frames, durations, path):
    ffmpeg = find_ffmpeg()
    if not ffmpeg:
        print("ffmpeg not found; skipped MP4")
        return False

    # Raw-video input has one constant frame rate. Duplicate each still to
    # approximate per-frame durations when a manifest mixes them.
    base_ms = max(1, min(durations))
    expanded = []
    for frame, ms in zip(frames, durations):
        expanded.extend([frame] * max(1, round(ms / base_ms)))
    fps = 1000.0 / base_ms
    width, height = frames[0].size
    cmd = [
        ffmpeg,
        "-y",
        "-f",
        "rawvideo",
        "-pix_fmt",
        "rgb24",
        "-s",
        f"{width}x{height}",
        "-framerate",
        f"{fps:.6f}",
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
        path,
    ]
    proc = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE,
    )
    try:
        for frame in expanded:
            proc.stdin.write(frame.tobytes())
        proc.stdin.close()
    except BrokenPipeError:
        pass
    if proc.wait() != 0:
        err = proc.stderr.read().decode(errors="replace").strip()
        print(f"ffmpeg failed: {err}")
        return False
    print(f"wrote {os.path.basename(path)} ({os.path.getsize(path) // 1024} KiB)")
    return True


def main(argv=None) -> None:
    args = parse_args(argv)
    manifest_path = os.path.abspath(args.manifest)
    manifest_dir = os.path.dirname(manifest_path)
    try:
        with open(manifest_path, encoding="utf-8") as fh:
            spec = json.load(fh)
    except (OSError, json.JSONDecodeError) as exc:
        sys.exit(f"could not read manifest {args.manifest!r}: {exc}")

    raw_frames = spec.get("frames")
    if not isinstance(raw_frames, list) or not raw_frames:
        sys.exit("manifest must contain a non-empty 'frames' list")

    width = args.width or int(spec.get("width", DEFAULT_WIDTH))
    if width <= 0:
        sys.exit("frame width must be positive")
    default_duration = int(spec.get("duration_ms", DEFAULT_DURATION_MS))
    if args.duration is not None:
        default_duration = args.duration
    if default_duration <= 0:
        sys.exit("frame duration must be positive")

    from PIL import Image, ImageDraw, ImageFont

    font_path = resolve_font(args.font, HERE)
    font = ImageFont.truetype(font_path, args.font_size)
    small = ImageFont.truetype(font_path, max(11, args.font_size - 8))

    images = []
    durations = []
    for index, item in enumerate(raw_frames, 1):
        if not isinstance(item, dict):
            sys.exit(f"frame {index} must be an object")
        name = item.get("image")
        caption = (item.get("caption") or "").strip()
        if not name:
            sys.exit(f"frame {index} is missing 'image'")
        if not caption:
            sys.exit(f"frame {index} is missing 'caption'")
        image_path = resolve_image(name, manifest_dir)
        with Image.open(image_path) as src:
            image = src.convert("RGB")
        if image.width != width:
            height = max(1, round(image.height * width / image.width))
            image = image.resize((width, height), Image.Resampling.LANCZOS)
        duration = int(
            item.get("duration_ms")
            or (args.duration if args.duration is not None else default_duration)
        )
        if duration <= 0:
            sys.exit(f"frame {index} has a non-positive duration")
        images.append((image, caption, index))
        durations.append(duration)

    content_height = max(image.height for image, _, _ in images)
    bar_height = max(76, args.font_size + 54)
    frames = []
    for image, caption, _ in images:
        frame = Image.new("RGB", (width, bar_height + content_height), DEFAULT_BG)
        frame.paste(image, (0, bar_height + (content_height - image.height) // 2))
        draw = ImageDraw.Draw(frame)
        draw.rectangle((0, 0, width, ACCENT_H), fill=ACCENT)

        caption_y = max(14, (bar_height - args.font_size) // 2 - 3)
        caption = fit_text(
            draw, caption, font, width - 2 * PAD - 80
        )
        draw.text((PAD, caption_y), caption, font=font, fill=DEFAULT_FG)
        counter = f"{len(frames) + 1}/{len(images)}"
        counter_width = draw.textlength(counter, font=small)
        draw.text(
            (width - PAD - counter_width, caption_y + 9),
            counter,
            font=small,
            fill=DEFAULT_MUTED,
        )
        frames.append(frame)

    gif_path = resolve_output(args, spec, manifest_path)
    os.makedirs(os.path.dirname(gif_path) or ".", exist_ok=True)
    frames[0].save(
        gif_path,
        save_all=True,
        append_images=frames[1:],
        duration=durations,
        loop=0,
        optimize=True,
        disposal=2,
    )
    print(
        f"wrote {os.path.basename(gif_path)} "
        f"({len(frames)} frames, {os.path.getsize(gif_path) // 1024} KiB)"
    )

    if args.mp4:
        mp4_path = os.path.splitext(gif_path)[0] + ".mp4"
        write_mp4(frames, durations, mp4_path)


if __name__ == "__main__":
    main()
