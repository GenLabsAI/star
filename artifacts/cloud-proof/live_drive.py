"""Drive a LIVE prod relay session: watch the PTY, and send keystrokes.

Proves the chain is bidirectional: bytes from the cloud runner reach the
viewer, and viewer input reaches the runner's PTY and changes its screen.

Usage: python3 live_drive.py <browser-ws-url> <out.log> [budget_seconds]
"""

from __future__ import annotations

import asyncio
import sys
import time

import websockets

MSG_DATA = 0x00
MSG_EXIT = 0x02


async def drive(url: str, out: str, budget: int):
    deadline = time.time() + budget
    with open(out, "a") as logf:
        def note(line: str):
            logf.write(f"\n=== {line} ===\n")
            logf.flush()

        attempt = 0
        while time.time() < deadline:
            attempt += 1
            try:
                async with websockets.connect(url, open_timeout=20, max_size=4 * 1024 * 1024) as ws:
                    note(f"attempt {attempt} attached")
                    # Wait for first runner bytes before sending anything.
                    first = await asyncio.wait_for(ws.recv(), timeout=120)
                    data = first if isinstance(first, bytes) else first.encode()
                    if data and data[0] == MSG_EXIT:
                        note("runner exited before input")
                        return
                    logf.write(data[1:].decode("utf-8", errors="replace"))
                    logf.flush()

                    note("banner received; sending 2x DOWN then ENTER")
                    for _ in range(2):
                        await ws.send(bytes([MSG_DATA]) + b"\x1b[B")
                        await asyncio.sleep(0.6)
                    await ws.send(bytes([MSG_DATA]) + b"\r")
                    note("input sent; collecting response")
                    end = time.time() + 20
                    while time.time() < end:
                        try:
                            msg = await asyncio.wait_for(ws.recv(), timeout=5)
                        except asyncio.TimeoutError:
                            continue
                        d = msg if isinstance(msg, bytes) else msg.encode()
                        if d and d[0] == MSG_EXIT:
                            note(f"runner exit code={d[1] if len(d) > 1 else '?'}")
                            return
                        logf.write(d[1:].decode("utf-8", errors="replace"))
                        logf.flush()
                    note("response window closed")
                    return
            except Exception as e:
                note(f"attempt {attempt} ended: {type(e).__name__}: {e}")
                await asyncio.sleep(3)
        note("budget exhausted")


if __name__ == "__main__":
    asyncio.run(drive(sys.argv[1], sys.argv[2], int(sys.argv[3]) if len(sys.argv) > 3 else 300))
