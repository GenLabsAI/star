"""Attach to a LIVE prod terminal relay browser URL and log everything.

Holds the session alive (prod kills the runner when the browser goes
away) and survives the 120s peer timeout by re-attaching: the GH
Actions runner often needs minutes to boot.

Usage: python3 live_watch.py <relay_url> <out.log> [max_seconds]
Stops early if the runner sends an exit frame.
"""

from __future__ import annotations

import asyncio
import sys
import time

import websockets

MSG_EXIT = 0x02


async def one_attempt(url: str, logf, deadline: float) -> str:
    """Return 'exit' if runner exited, 'timeout' if peer never came, 'closed' otherwise."""
    try:
        async with websockets.connect(url, open_timeout=20, max_size=4 * 1024 * 1024) as ws:
            print(f"ATTACHED {url}", flush=True)
            logf.write(f"[{time.strftime('%H:%M:%S')}] ATTACHED\n")
            logf.flush()
            while time.time() < deadline:
                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=15)
                except asyncio.TimeoutError:
                    continue
                data = msg if isinstance(msg, bytes) else msg.encode()
                if not data:
                    continue
                if data[0] == MSG_EXIT:
                    code = data[1] if len(data) > 1 else -1
                    logf.write(f"[{time.strftime('%H:%M:%S')}] RUNNER EXIT code={code}\n")
                    logf.flush()
                    return "exit"
                payload = data[1:] if data[0] <= 0x03 else data
                try:
                    text = payload.decode("utf-8", errors="replace")
                except Exception:
                    text = repr(payload[:200])
                logf.write(text)
                logf.flush()
    except websockets.exceptions.ConnectionClosed as e:
        logf.write(f"[{time.strftime('%H:%M:%S')}] WS closed {e.code} {e.reason}\n")
        logf.flush()
        if e.code == 1008 or "peer" in (e.reason or "").lower():
            return "timeout"
        return "closed"
    except Exception as e:
        logf.write(f"[{time.strftime('%H:%M:%S')}] attach error: {e!r}\n")
        logf.flush()
        return "closed"
    return "timeout"


async def main(url: str, out: str, budget: int):
    deadline = time.time() + budget
    attempt = 0
    with open(out, "a") as logf:
        while time.time() < deadline:
            attempt += 1
            print(f"attempt {attempt}", flush=True)
            res = await one_attempt(url, logf, deadline)
            if res == "exit":
                print("RUNNER EXITED", flush=True)
                return
            await asyncio.sleep(3)
    print("BUDGET EXHAUSTED", flush=True)


if __name__ == "__main__":
    asyncio.run(main(sys.argv[1], sys.argv[2], int(sys.argv[3]) if len(sys.argv) > 3 else 600))
