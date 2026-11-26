"""Drive the real Star TUI into cloud mode against the live Deca API.

This is the "just ask it" path: one env var for the API base, one for the
web session id, and a normal prompt. Everything else (session create,
harness dispatch, relay, model) happens behind the TUI.

It writes a receipts file with wall-clock timings and screen frames, and
leaves the raw output beside it.

Usage:
  python3 deca_live_run.py <star-binary> <out.txt> [budget_seconds]
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
import urllib.request

BASE = os.environ.get("DECA_AGENT_API_BASE_URL", "https://api.genlabs.dev")
SESSION_ID = os.environ["DECA_SESSION_ID"]
GIT_TOKEN = os.environ.get("DECA_GIT_TOKEN", "")
MODEL = os.environ.get("CLOUD_DEMO_MODEL", "deca-2.6-pro")

# CLOUD_DEMO_PROMPT_FILE lets a run use its own prompt (e.g. a deliberately
# terse one) without editing this script.
PROMPT_FILE = os.environ.get("CLOUD_DEMO_PROMPT_FILE")

DEFAULT_PROMPT = (
    "Cloud demo design task, do it end to end: (1) run `pwd; ls -a | head; "
    "git remote -v; git log --oneline -3` and say where you are working; "
    "(2) build a premium dark hero page for a product called Star at "
    "cloud-demo/hero.html, single file, with a large headline, a short "
    "subline and a slim chat input bar (1px translucent white border, tight "
    "padding, soft glow, placeholder 'Ask Star anything...'); (3) render it "
    "headless at 1440x900 and save the screenshot to cloud-demo/hero.png; "
    "(4) commit both files as 'feat: cloud demo hero chat bar' on a new "
    "branch cloud-demo/hero and push that branch to "
    "https://github.com/GenLabsAI/star (the git token is already in your "
    "environment, never print it); (5) reply with the commit SHA, the "
    "branch, the screenshot path and one line describing the design."
)

PROMPT = open(PROMPT_FILE).read().strip() if PROMPT_FILE else DEFAULT_PROMPT


def enroll() -> str:
    signals = {
        "machine_id": "star-cloud-demo",
        "host_id": "star-cloud-demo",
        "hostname": "star-cloud-demo",
        "cpu": "star-cloud-demo",
        "os": "linux",
    }
    data = json.dumps({"signals": signals}).encode()
    req = urllib.request.Request(BASE + "/deca/v1/opencode/enroll", data=data, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("User-Agent", "star-cloud-probe")
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read())["api_key"]


def main():
    binary = sys.argv[1]
    out_path = sys.argv[2]
    budget = int(sys.argv[3]) if len(sys.argv) > 3 else 900

    home = os.environ.get("DRIVE_HOME", "/tmp/deca-live-home")
    config_dir = os.path.join(home, ".config", "crush")
    data_dir = os.path.join(home, ".local", "share", "crush")
    work_dir = os.path.join(home, "work")
    for d in (config_dir, data_dir, work_dir):
        os.makedirs(d, exist_ok=True)

    key = enroll()
    cfg = {
        "options": {"disable_default_providers": True, "disable_notifications": True},
        "providers": {
            "deca": {
                "name": "Deca",
                "type": "openai",
                "base_url": BASE + "/deca/v1",
                "api_key": key,
                "models": [{"id": MODEL, "name": "Deca 2.6 Pro", "context_window": 200000}],
            },
        },
        "models": {
            "large": {"provider": "deca", "model": MODEL},
            "small": {"provider": "deca", "model": MODEL},
        },
    }
    cfg_path = os.path.join(config_dir, "crush.json")
    fd = os.open(cfg_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, "w") as fh:
        json.dump(cfg, fh)
    del key

    env = {
        **os.environ,
        "HOME": home,
        "TERM": "xterm-256color",
        "COLORTERM": "truecolor",
        "CRUSH_GLOBAL_CONFIG": config_dir,
        "CRUSH_GLOBAL_DATA": data_dir,
        "DECA_AGENT_API_BASE_URL": BASE,
        "DECA_SESSION_ID": SESSION_ID,
    }
    if GIT_TOKEN:
        env["DECA_GIT_TOKEN"] = GIT_TOKEN

    started = time.time()

    def since() -> str:
        return f"{time.time() - started:6.1f}s"

    log = open(out_path, "w")

    def note(line: str) -> None:
        log.write(line + "\n")
        log.flush()

    def frame(name: str, view: str) -> None:
        log.write(f"\n=== screen: {name} (t={since()}) ===\n{view}\n")
        log.flush()

    note(f"t=0.0s exec {binary} (cwd {work_dir})")

    pid, pfd = pty.fork()
    if pid == 0:
        os.chdir(work_dir)
        os.execve(binary, [binary, "--debug"], env)
        os._exit(127)

    fcntl.ioctl(pfd, termios.TIOCSWINSZ, struct.pack("HHHH", 45, 140, 0, 0))

    import pyte

    screen = pyte.Screen(140, 45)
    stream = pyte.ByteStream(screen)
    raw: list[bytes] = []

    def render() -> str:
        return "\n".join(screen.display).rstrip()

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

    def send(text: str, pause: float = 1.0, per_char: float = 0.008):
        for ch in text:
            os.write(pfd, ch.encode())
            time.sleep(per_char)
        drain(pause)

    drain(12)
    note(f"t={since()} first paint")
    frame("startup", render())

    send("/cloud", 1.5)
    os.write(pfd, b"\r")
    drain(2.5)
    # The command palette can linger after the command runs; dismiss it so
    # the prompt goes into the composer, not the filter.
    os.write(pfd, b"\x1b")
    drain(1.5)
    note(f"t={since()} cloud mode toggled")
    frame("cloud mode on", render())

    send(PROMPT, 0.5)
    os.write(pfd, b"\r")
    submitted = time.time()
    note(f"t={since()} prompt submitted ({len(PROMPT)} chars)")
    frame("submitted", render())

    deadline = started + budget
    seen_tool = False
    last_change = time.time()
    last_view = ""
    while time.time() < deadline:
        if not drain(3):
            notes.append(f"t={since()} process exited")
            break
        view = render()
        if view != last_view:
            last_change = time.time()
            last_view = view
        # The turn footer ("◇ Deca 2.6 Pro via Deca in 2m16s") is the one
        # reliable end-of-turn marker in the rendered screen.
        if re.search(r"\bin \d+m?\d+s\b", view):
            note(f"t={since()} turn footer seen (+{time.time() - submitted:.1f}s after submit)")
            break
        if not seen_tool and ("tool" in view.lower()):
            frame("first tool activity", view)
            note(f"t={since()} first tool activity (+{time.time() - submitted:.1f}s after submit)")
            seen_tool = True
        # A finished turn stops repainting; give it a quiet window.
        if seen_tool and time.time() - last_change > 20:
            note(f"t={since()} screen quiet for 20s; capturing final")
            break
        if not seen_tool and time.time() - last_change > 180:
            note(f"t={since()} no activity for 180s; giving up")
            frame("stalled", view)
            break

    os.write(pfd, b"\t")
    drain(1.0)
    os.write(pfd, b"\t")
    drain(1.5)
    frame("final", render())
    note(f"t={since()} done (turn wall {(time.time() - submitted):.1f}s)")
    log.close()

    with open(out_path + ".raw", "wb") as fh:
        fh.write(b"".join(raw))
    print(f"wrote {out_path}")
    try:
        os.kill(pid, 9)
    except ProcessLookupError:
        pass


if __name__ == "__main__":
    main()
