"""Mimic Star's RunACPSession exactly (text frames, same sequence), but in
Python, to isolate whether prod relay -> reader failures are coder/websocket
specific. Usage: py_mimic.py <browser-ws-url> <api_key>."""

from __future__ import annotations

import asyncio
import json
import sys

import websockets

MSG_EXIT = 0x02


def strip(data: bytes) -> bytes:
    if data and data[0] <= 0x03:
        return data[1:]
    return data


async def main(url: str, api_key: str):
    async with websockets.connect(url, max_size=4 * 1024 * 1024) as ws:
        await ws.send(json.dumps({"type": "acp_init", "prompt": "Reply with exactly: PY-MIMIC-OK",
                                  "model": "deca-2.5-mini", "base_url": "https://api.genlabs.dev/deca/v1",
                                  "api_key": api_key}))
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize",
                                  "params": {"protocolVersion": 1}}))

        async def wait_id(want: int):
            while True:
                raw = await asyncio.wait_for(ws.recv(), timeout=120)
                data = raw.encode() if isinstance(raw, str) else bytes(raw)
                if data and data[0] == MSG_EXIT:
                    print("GOT-EXIT")
                    return None
                msg = json.loads(strip(data).decode())
                if msg.get("method"):
                    print("chunk:", json.dumps(msg)[:160], flush=True)
                    continue
                if str(msg.get("id")) == str(want):
                    print(f"id={want} ok:", json.dumps(msg)[:160], flush=True)
                    return msg

        await wait_id(1)
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "session/new",
                                  "params": {"cwd": "/tmp", "mcpServers": []}}))
        resp = await wait_id(2)
        sid = ((resp or {}).get("result") or {}).get("sessionId", "cloud-1")
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 3, "method": "session/prompt",
                                  "params": {"sessionId": sid, "prompt": [{"type": "text",
                                            "text": "Reply with exactly: PY-MIMIC-OK"}]}}))
        while True:
            raw = await asyncio.wait_for(ws.recv(), timeout=120)
            data = raw.encode() if isinstance(raw, str) else bytes(raw)
            if data and data[0] == MSG_EXIT:
                print("GOT-EXIT-CODE", data[1] if len(data) > 1 else "?", flush=True)
                return
            msg = json.loads(strip(data).decode())
            if msg.get("method"):
                print("chunk:", json.dumps(msg)[:160], flush=True)
                continue
            print("final:", json.dumps(msg)[:200], flush=True)
            return


asyncio.run(main(sys.argv[1], sys.argv[2]))
