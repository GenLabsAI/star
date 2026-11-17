"""Mock of the prod terminal relay (genlabsapi terminal_relay/flask_relay.py).

Implements the same surface the harness terminal_runner.py and a
browser/Star viewer use in production:
  POST /deca/v1/terminal/session                      -> {session_id, relay_url}
  POST /deca/v1/terminal/session/<id>/scrollback       -> {received: n}
  WS   /terminal/runner/<session_id>                   (runner side, PTY bytes)
  WS   /terminal/browser/<session_id>                  (viewer side)

Bytes are pumped bidirectionally between the paired sockets, exactly
like prod's _pump. Like prod, WS endpoints do not check tokens; the
session_id (uuid) is the capability.

Run: python3 mock_terminal_relay.py [port]  (default 8934)
Env: MOCK_USER_KEY (default user-secret)
"""

from __future__ import annotations

import asyncio
import json
import uuid

import aiohttp
import aiohttp.web as web

USER_KEY = __import__("os").environ.get("MOCK_USER_KEY", "user-secret")
SESSION_ID = __import__("os").environ.get("MOCK_SESSION_ID", "test-session-id")

sessions: dict[str, dict] = {}
lock = asyncio.Lock()


async def create_session(request: web.Request) -> web.Response:
    # Mirror prod: sessionid cookie first, then Bearer API key.
    if request.cookies.get("sessionid") != SESSION_ID:
        auth = request.headers.get("Authorization", "")
        if auth != f"Bearer {USER_KEY}":
            return web.json_response({"error": "Not authenticated"}, status=401)
    session_id = str(uuid.uuid4())
    async with lock:
        sessions[session_id] = {"browser": None, "runner": None, "scrollback": ""}
    base = f"{request.scheme}://{request.host}"
    ws_base = base.replace("http", "ws", 1)
    return web.json_response({
        "session_id": session_id,
        "relay_url": f"{ws_base}/terminal/browser/{session_id}",
        "runner_url": f"{ws_base}/terminal/runner/{session_id}",
    })


async def scrollback(request: web.Request) -> web.Response:
    session_id = request.match_info["session_id"]
    async with lock:
        sess = sessions.get(session_id)
        if sess is None:
            return web.json_response({"error": "unknown session"}, status=404)
        body = await request.json()
        sess["scrollback"] = body.get("scrollback", "")
        n = len(sess["scrollback"])
    return web.json_response({"received": n})


async def _pump_logged(src, dst, session_id: str, direction: str):
    n = 0
    try:
        async for msg in src:
            if msg.type in (aiohttp.WSMsgType.CLOSE, aiohttp.WSMsgType.CLOSED,
                            aiohttp.WSMsgType.CLOSING, aiohttp.WSMsgType.ERROR):
                break
            data = msg.data
            if isinstance(data, str):
                data = data.encode()
            if isinstance(data, bytes):
                n += 1
                await dst.send_bytes(data)
    except Exception as e:
        print(f"relay: pump {direction} {session_id} raised: {e!r}", flush=True)
    print(f"relay: pump {direction} {session_id} forwarded {n} msgs", flush=True)


async def _side(request: web.Request, side: str) -> web.WebSocketResponse:
    session_id = request.match_info["session_id"]
    async with lock:
        sess = sessions.get(session_id)
    ws = web.WebSocketResponse(max_msg_size=4 * 1024 * 1024)
    await ws.prepare(request)
    if sess is None:
        await ws.close(code=1008, message=b"Session not found")
        return ws
    print(f"relay: {side} connected {session_id} known={sess is not None}", flush=True)
    try:
        async with lock:
            sess[side] = ws
            peer = sess["runner" if side == "browser" else "browser"]
    except Exception as e:
        print(f"relay: {side} store failed: {e!r}", flush=True)
        raise
    # Wait for peer (prod waits up to 120s; mock uses 60s).
    for _ in range(600):
        if peer is not None and not peer.closed:
            break
        await asyncio.sleep(0.1)
        async with lock:
            peer = sess["runner" if side == "browser" else "browser"]
    else:
        await ws.close(code=1000, message=b"peer_timeout")
        return ws
    print(f"relay: {side} peer ready {session_id}, pumping", flush=True)
    try:
        await _pump_logged(ws, peer, session_id, side)
    except Exception as e:
        print(f"relay: {side} pump raised: {e!r}", flush=True)
    print(f"relay: {side} handler done {session_id}", flush=True)
    return ws


async def runner_ws(request: web.Request) -> web.WebSocketResponse:
    return await _side(request, "runner")


async def browser_ws(request: web.Request) -> web.WebSocketResponse:
    return await _side(request, "browser")


def main(port: int):
    app = web.Application()
    app.router.add_post("/deca/v1/terminal/session", create_session)
    app.router.add_post("/deca/v1/terminal/session/{session_id}/scrollback", scrollback)
    app.router.add_get("/terminal/runner/{session_id}", runner_ws)
    app.router.add_get("/terminal/browser/{session_id}", browser_ws)
    print(f"mock terminal relay on http://127.0.0.1:{port}", flush=True)
    web.run_app(app, host="127.0.0.1", port=port, print=None)


if __name__ == "__main__":
    import sys
    main(int(sys.argv[1]) if len(sys.argv) > 1 else 8934)
