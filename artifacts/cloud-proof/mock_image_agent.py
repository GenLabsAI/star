"""Minimal cloud ACP agent that sends a picture the canonical ACP way.

Stands in for the harness bridge when the bridge's own behaviour is not
what is under test: it speaks the same relay protocol and ACP methods,
then answers one prompt with a text chunk, an image content block
(base64 data + mimeType + uri), and the Markdown reference the agent
would write for it.

Run: python3 mock_image_agent.py [port]
Env: MOCK_IMAGE (file whose bytes travel as the image block),
     MOCK_IMAGE_URI (the path the prose refers to).
"""

from __future__ import annotations

import base64
import json
import os
import sys
import uuid

import aiohttp.web as web

IMAGE = os.environ.get("MOCK_IMAGE", "")
IMAGE_URI = os.environ.get("MOCK_IMAGE_URI", "/artifacts/image.png")


async def healthz(request: web.Request) -> web.Response:
    return web.json_response({"ok": True})


async def create_session(request: web.Request) -> web.Response:
    session_id = str(uuid.uuid4())
    relay_url = f"ws://{request.host}/relay/browser/{session_id}"
    print(f"api: session {session_id} -> {relay_url}", flush=True)
    return web.json_response({"session_id": session_id, "relay_url": relay_url})


def frame(payload: dict) -> bytes:
    """PTY-style framing the real bridge uses: 0x00 then JSON."""
    return b"\x00" + json.dumps(payload).encode()


def update(session_id: str, content: dict) -> dict:
    return {
        "jsonrpc": "2.0",
        "method": "session/update",
        "params": {
            "sessionId": session_id,
            "update": {"sessionUpdate": "agent_message_chunk", "content": content},
        },
    }


async def relay(request: web.Request) -> web.WebSocketResponse:
    ws = web.WebSocketResponse(max_msg_size=32 * 1024 * 1024)
    await ws.prepare(request)
    async for msg in ws:
        if msg.type != web.WSMsgType.TEXT:
            continue
        try:
            req = json.loads(msg.data)
        except json.JSONDecodeError:
            continue
        method, rid = req.get("method"), req.get("id")
        if method == "initialize":
            await ws.send_bytes(frame({
                "jsonrpc": "2.0", "id": rid,
                "result": {"protocolVersion": 1, "agentCapabilities": {},
                           "agentInfo": {"name": "mock-image-agent", "version": "1"}},
            }))
        elif method == "session/new":
            await ws.send_bytes(frame({
                "jsonrpc": "2.0", "id": rid, "result": {"sessionId": "cloud-1"},
            }))
        elif method == "session/prompt":
            session_id = req["params"]["sessionId"]
            data = base64.b64encode(open(IMAGE, "rb").read()).decode()
            print(f"api: sending image ({len(data)} base64 chars) then prose", flush=True)
            await ws.send_bytes(frame(update(session_id, {
                "type": "text", "text": "Here's the picture of my work:\n\n",
            })))
            await ws.send_bytes(frame(update(session_id, {
                "type": "image", "data": data, "mimeType": "image/png", "uri": IMAGE_URI,
            })))
            await ws.send_bytes(frame(update(session_id, {
                "type": "text", "text": f"![My work]({IMAGE_URI})",
            })))
            await ws.send_bytes(frame({
                "jsonrpc": "2.0", "id": rid, "result": {"stopReason": "end_turn"},
            }))
            print("api: prompt answered", flush=True)
    return ws


def main():
    if not IMAGE:
        sys.exit("MOCK_IMAGE must point at a picture")
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8951
    app = web.Application()
    app.router.add_get("/healthz", healthz)
    app.router.add_post("/deca/v1/cloud/acp/session", create_session)
    app.router.add_get("/relay/browser/{session}", relay)
    print(f"api: listening on http://127.0.0.1:{port}", flush=True)
    web.run_app(app, host="127.0.0.1", port=port, print=None)


if __name__ == "__main__":
    main()
