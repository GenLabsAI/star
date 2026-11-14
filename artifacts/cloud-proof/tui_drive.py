"""Drive the real Star TUI in a PTY: type /cloud, send a prompt, capture.

This runs the actual `star` binary (the same Bubble Tea program users
run) against the mock cloud backend, so the captured screen is the
production render path, not a mockup.

Usage:
  python3 tui_drive.py <star-binary> <out.txt> [budget_seconds]

Environment passed to the TUI is built here: a temp HOME with a global
config pointing at the mock OpenAI-compatible provider (so onboarding is
skipped) plus DECA_* variables for the cloud path.
"""

from __future__ import annotations

import fcntl
import json
import os
import pty
import re
import select
import struct
import sys
import termios
import time

ANSI_RE = re.compile(rb"\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07]*\x07|\x1b[()][A-Z0-9]")


def strip_ansi(data: bytes) -> str:
    return ANSI_RE.sub(b"", data).decode("utf-8", errors="replace")


def main():
    binary = sys.argv[1]
    out_path = sys.argv[2]
    budget = int(sys.argv[3]) if len(sys.argv) > 3 else 90
    api = os.environ.get("MOCK_API_BASE", "http://127.0.0.1:8944")

    home = os.environ.get("DRIVE_HOME", "/tmp/star-drive-home")
    config_dir = os.path.join(home, ".config", "crush")
    data_dir = os.path.join(home, ".local", "share", "crush")
    work_dir = os.path.join(home, "work")
    for d in (config_dir, data_dir, work_dir):
        os.makedirs(d, exist_ok=True)

    # A local provider that never serves the turn: cloud inference goes
    # through DECA_MODEL*, and this only makes the local agent ready so
    # the TUI skips onboarding.
    cfg = {
        "options": {"disable_default_providers": True, "disable_notifications": True},
        "providers": {
            "local": {
                "name": "Local Mock",
                "type": "openai",
                "base_url": api + "/v1",
                "api_key": "local-key",
                "models": [{"id": "cloud-demo", "name": "Cloud Demo", "context_window": 200000}],
            },
        },
        "models": {
            "large": {"provider": "local", "model": "cloud-demo"},
            "small": {"provider": "local", "model": "cloud-demo"},
        },
    }
    with open(os.path.join(config_dir, "crush.json"), "w") as fh:
        json.dump(cfg, fh)

    env = {
        **os.environ,
        "HOME": home,
        "TERM": "xterm-256color",
        "COLORTERM": "truecolor",
        "CRUSH_GLOBAL_CONFIG": config_dir,
        "CRUSH_GLOBAL_DATA": data_dir,
        "DECA_AGENT_API_BASE_URL": api,
        "DECA_API_KEY": os.environ.get("MOCK_USER_KEY", "user-secret"),
        "DECA_MODEL": "cloud-demo",
        "DECA_MODEL_BASE_URL": api + "/v1",
        "DECA_MODEL_API_KEY": "model-key",
    }

    pid, fd = pty.fork()
    if pid == 0:
        os.chdir(work_dir)
        os.execve(binary, [binary, "--debug"], env)
        os._exit(127)

    # 140x45 like the screenshot frames, so both proofs look alike.
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", 45, 140, 0, 0))

    chunks: list[bytes] = []

    def drain(seconds: float):
        end = time.time() + seconds
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.2)
            if not r:
                continue
            try:
                data = os.read(fd, 65536)
            except OSError:
                return False
            if not data:
                return False
            chunks.append(data)
        return True

    def send(text: str, pause: float = 1.2):
        for ch in text:
            os.write(fd, ch.encode())
            time.sleep(0.05)
        drain(pause)

    import pyte

    screen = pyte.Screen(140, 45)
    stream = pyte.ByteStream(screen)
    frames: list[tuple[str, str]] = []

    def render() -> str:
        return "\n".join(screen.display).rstrip()

    def drain(seconds: float) -> bool:
        end = time.time() + seconds
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.2)
            if not r:
                continue
            try:
                data = os.read(fd, 65536)
            except OSError:
                return False
            if not data:
                return False
            chunks.append(data)
            stream.feed(data)
        return True

    notes = []
    drain(8)  # splash + first paint
    notes.append(("after start", len(chunks)))
    frames.append(("startup", render()))

    send("/cloud", 1.0)
    os.write(fd, b"\r")
    drain(2.0)
    notes.append(("after /cloud", len(chunks)))
    frames.append(("cloud mode toggled", render()))

    send("Run a quick check on the cloud machine and tell me what you see.", 0.4)
    os.write(fd, b"\r")
    notes.append(("prompt sent", len(chunks)))

    deadline = time.time() + budget
    seen_tool = False
    while time.time() < deadline:
        if not drain(2):
            break
        view = render()
        if not seen_tool and "Running 1 tool" in view:
            frames.append(("tool running", view))
            seen_tool = True
        if "Nothing was run locally." in view:
            frames.append(("turn complete", view))
            drain(2)
            frames.append(("final", render()))
            break
    # Force one repaint (focus changes) so the captured frame is the live
    # UI state, not a stale screen left by a synchronized update.
    os.write(fd, b"\t")
    drain(1.0)
    os.write(fd, b"\t")
    drain(1.5)
    frames.append(("last screen", render()))

    with open(out_path, "w") as fh:
        fh.write("=== capture notes ===\n")
        for name, n in notes:
            fh.write(f"{name}: {n} chunks\n")
        for name, view in frames:
            fh.write(f"\n=== screen: {name} ===\n{view}\n")
    print(f"wrote {out_path} ({len(frames)} frames)")
    try:
        os.kill(pid, 9)
    except ProcessLookupError:
        pass


if __name__ == "__main__":
    main()
