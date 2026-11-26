"""Prove a cloud image reference becomes a clickable local file link.

Runs the real Star TUI against `mock_image_agent.py`, which answers one
cloud prompt with a text chunk, a canonical ACP image content block, and
the Markdown reference `![My work](/artifacts/image.png)`. The script
then reads the TUI's own message database and its raw terminal output to
check three things:

  * the image was saved under the session's cloud-images directory,
  * the Markdown reference was rewritten to a file:// link,
  * the rendered screen carries an OSC 8 hyperlink for that file.

Usage: python3 image_link_proof.py <star-binary> <out-dir> [hero.png]
"""

from __future__ import annotations

import fcntl
import hashlib
import json
import os
import pty
import re
import select
import shutil
import signal
import socket
import sqlite3
import struct
import subprocess
import sys
import termios
import time

import pyte

HERE = os.path.dirname(os.path.abspath(__file__))
FOOTER = re.compile(r"\bin \d+m?\d+s\b")
ANSI = re.compile(rb"\x1b\[[0-9;?]*[A-Za-z]")


def wait_port(port: int, deadline: float = 20.0) -> None:
    end = time.time() + deadline
    while time.time() < end:
        with socket.socket() as s:
            s.settimeout(0.5)
            if s.connect_ex(("127.0.0.1", port)) == 0:
                return
        time.sleep(0.2)
    raise RuntimeError(f"mock agent did not listen on {port}")


def main() -> int:
    star = sys.argv[1]
    out_dir = sys.argv[2]
    hero = sys.argv[3] if len(sys.argv) > 3 else os.path.join(HERE, "hero", "hero.png")
    os.makedirs(out_dir, exist_ok=True)

    port = 8951
    home = os.path.join(out_dir, "home")
    config_dir = os.path.join(home, ".config", "crush")
    data_dir = os.path.join(home, ".local", "share", "crush")
    work_dir = os.path.join(home, "work")
    for d in (config_dir, data_dir, work_dir):
        os.makedirs(d, exist_ok=True)

    mock_log = open(os.path.join(out_dir, "mock.log"), "w")
    mock = subprocess.Popen(
        [_system_python(), os.path.join(HERE, "mock_image_agent.py"), str(port)],
        env={**os.environ, "MOCK_IMAGE": hero, "MOCK_IMAGE_URI": "/artifacts/image.png"},
        stdout=mock_log, stderr=subprocess.STDOUT,
    )
    wait_port(port)

    cfg = {
        "options": {"disable_default_providers": True, "disable_notifications": True},
        "providers": {"local": {
            "name": "Cloud Demo", "type": "openai",
            "base_url": f"http://127.0.0.1:{port}/v1", "api_key": "mock-key",
            "models": [{"id": "cloud-demo", "name": "Cloud Demo", "context_window": 200000}],
        }},
        "models": {"large": {"provider": "local", "model": "cloud-demo"},
                   "small": {"provider": "local", "model": "cloud-demo"}},
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
        "DECA_AGENT_API_BASE_URL": f"http://127.0.0.1:{port}",
        "DECA_API_KEY": "mock-key",
    }

    started = time.time()
    pid, pfd = pty.fork()
    if pid == 0:
        os.chdir(work_dir)
        os.execve(star, [star, "--debug"], env)
        os._exit(127)
    fcntl.ioctl(pfd, termios.TIOCSWINSZ, struct.pack("HHHH", 45, 140, 0, 0))

    screen = pyte.Screen(140, 45)
    stream = pyte.ByteStream(screen)
    raw: list[bytes] = []

    def drain(seconds: float) -> bool:
        end = time.time() + seconds
        while time.time() < end:
            r, _, _ = select.select([pfd], [], [], 0.2)
            if not r:
                continue
            try:
                data = os.read(pfd, 65536)
            except OSError:
                return False
            if not data:
                return False
            raw.append(data)
            stream.feed(data)
        return True

    def render() -> str:
        return "\n".join(screen.display).rstrip()

    def send(text: str, pause: float = 1.0, per_char: float = 0.008):
        for ch in text:
            os.write(pfd, ch.encode())
            time.sleep(per_char)
        drain(pause)

    drain(10)
    send("/cloud", 1.5)
    os.write(pfd, b"\r")
    drain(2.0)
    os.write(pfd, b"\x1b")
    drain(1.0)
    send("show me the picture", 0.5)
    os.write(pfd, b"\r")

    deadline = time.time() + 60
    final = ""
    while time.time() < deadline:
        if not drain(2):
            break
        final = render()
        if FOOTER.search(final):
            drain(2)
            final = render()
            break
    else:
        final = render()

    os.kill(pid, signal.SIGKILL)
    time.sleep(0.5)

    db = os.path.join(work_dir, ".crush", "crush.db")
    conn = sqlite3.connect(f"file:{db}?mode=ro", uri=True)
    texts = []
    for (parts,) in conn.execute("select parts from messages where role='assistant' order by created_at"):
        for part in json.loads(parts):
            if part["type"] == "text":
                texts.append(part["data"]["text"])
    assistant = "\n".join(texts)

    link_re = re.compile(r"\[My work\]\((file://[^)]+)\)")
    match = link_re.search(assistant)
    local_path = ""
    if match:
        local_path = match.group(1).removeprefix("file://")

    source_sha = hashlib.sha256(open(hero, "rb").read()).hexdigest()
    saved_sha = hashlib.sha256(open(local_path, "rb").read()).hexdigest() if local_path and os.path.exists(local_path) else ""

    # Glamour emits OSC 8 with a per-link id parameter: \x1b]8;id=N;file://...
    hyperlink = re.search(rb"\x1b\]8;[^;]*;file://[^\x1b\x07]+", b"".join(raw)) is not None
    link_line = next((line.strip() for line in final.splitlines() if "My work" in line), "")

    checks = [
        ("image sent over ACP was saved under the session's data dir",
         bool(local_path) and os.path.exists(local_path) and f"{os.sep}cloud-images{os.sep}" in local_path),
        ("saved bytes are identical to the picture the agent sent", source_sha == saved_sha != ""),
        ("agent's Markdown reference links the saved file", bool(match)),
        ("rendered screen shows the reference as a link", bool(link_line)),
        ("rendered screen carries an OSC 8 hyperlink to the file", hyperlink),
    ]

    lines = [
        "=== CLOUD IMAGE END TO END (client half, real TUI) ===",
        f"star: {star}",
        f"mock agent: mock_image_agent.py on 127.0.0.1:{port}",
        f"picture the agent sent: {hero} (sha256 {source_sha[:16]}...)",
        f"elapsed: {time.time() - started:.1f}s",
        "",
        "=== assistant text as stored ===",
        assistant.strip(),
        "",
        "=== final screen ===",
        final,
        "",
        "=== checks ===",
    ]
    for name, ok in checks:
        lines.append(f"[{'PASS' if ok else 'FAIL'}] {name}")

    proof = "\n".join(lines) + "\n"
    with open(os.path.join(out_dir, "image-link-proof.txt"), "w") as fh:
        fh.write(proof)
    with open(os.path.join(out_dir, "session.raw"), "wb") as fh:
        fh.write(b"".join(raw))
    print(proof)

    mock.send_signal(signal.SIGTERM)
    return 0 if all(ok for _, ok in checks) else 1


def _system_python() -> str:
    for candidate in ("/usr/bin/python3", "/usr/local/bin/python3", shutil.which("python3") or ""):
        if candidate and os.path.exists(candidate):
            probe = subprocess.run([candidate, "-c", "import aiohttp"], capture_output=True)
            if probe.returncode == 0:
                return candidate
    raise RuntimeError("no python3 with aiohttp for the mock agent")


if __name__ == "__main__":
    sys.exit(main())
