"""Dumb echo runner that stays alive: pairs with the prod relay, logs every
inbound frame, replies to initialize/session-new with canned results, and
never exits on its own. Usage: echo_runner.py <runner-ws-url>."""

from __future__ import annotations

import asyncio
import json
import sys

import websockets

MSG_DATA = 0x00


async def main(url: str):
    async with websockets.connect(url, max_size=4 * 1024 * 1024) as ws:
        print("echo-runner: paired, staying alive", flush=True)
        async for raw in ws:
            data = raw.encode() if isinstance(raw, str) else bytes(raw)
            body = data[1:] if data and data[0] <= 0x03 else data
            print(f"echo-runner: got {len(data)}b: {body[:200]!r}", flush=True)
            try:
                msg = json.loads(body.decode())
            except Exception:
                continue
            if msg.get("type") == "acp_init":
                continue
            mid = msg.get("id")
            if msg.get("method") == "initialize":
                await ws.send(bytes([MSG_DATA]) + json.dumps(
                    {"jsonrpc": "2.0", "id": mid, "result": {"protocolVersion": 1}}).encode())
            elif msg.get("method") == "session/new":
                await ws.send(bytes([MSG_DATA]) + json.dumps(
                    {"jsonrpc": "2.0", "id": mid, "result": {"sessionId": "echo-1"}}).encode())
            elif msg.get("method") == "session/prompt":
                await ws.send(bytes([MSG_DATA]) + json.dumps(
                    {"jsonrpc": "2.0", "method": "session/update",
                     "params": {"sessionId": "echo-1", "update": {"sessionUpdate": "agent_message_chunk",
                                                                 "content": {"type": "text",
                                                                             "text": "ECHO-OK"}}}}).encode())
                await ws.send(bytes([MSG_DATA]) + json.dumps(
                    {"jsonrpc": "2.0", "id": mid, "result": {"stopReason": "end_turn"}}).encode())
                print("echo-runner: answered prompt, staying alive", flush=True)


asyncio.run(main(sys.argv[1]))
