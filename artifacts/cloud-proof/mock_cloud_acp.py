"""Local stand-in for the cloud ACP backend, for driving the real TUI.

Serves everything a Star cloud turn talks to, so the whole loop runs
without production credentials:

  GET  /v1/models                         OpenAI-compatible model list
  POST /v1/chat/completions               scripted model: one bash tool
                                          call, then a final answer
  POST /deca/v1/cloud/acp/session         creates a session, starts the
                                          real acp_bridge.py as the
                                          cloud runner, returns the
                                          browser relay URL
  WS   /relay/runner/<session>            runner side (bridge)
  WS   /relay/browser/<session>           viewer side (Star)

The relay pumps bytes both ways and, unlike the production _pump bug,
does not treat an idle receive timeout as end-of-stream.

Run: python3 mock_cloud_acp.py [port]
Env: MOCK_USER_KEY (default user-secret), ACP_BRIDGE (bridge path),
     ACP_CWD (working directory the bridge's tools run in).
"""

from __future__ import annotations

import asyncio
import json
import os
import sys
import uuid

import aiohttp.web as web

USER_KEY = os.environ.get("MOCK_USER_KEY", "user-secret")
HERE = os.path.dirname(os.path.abspath(__file__))
BRIDGE = os.environ.get("ACP_BRIDGE", os.path.join(HERE, "acp_bridge.py"))
ACP_CWD = os.environ.get("ACP_CWD", "/tmp/cloud-machine")

sessions: dict[str, dict] = {}


async def models(request: web.Request) -> web.Response:
    return web.json_response({"data": [{"id": "cloud-demo", "object": "model"}]})


async def chat_completions(request: web.Request) -> web.Response:
    """Scripted model: first turn asks for one bash call, second answers."""
    body = await request.json()
    messages = body.get("messages") or []
    saw_tool_result = any(m.get("role") == "tool" for m in messages)
    if not saw_tool_result:
        message = {
            "role": "assistant",
            "content": "",
            "tool_calls": [{
                "id": "call_cloud_bash_1",
                "type": "function",
                "function": {
                    "name": "bash",
                    "arguments": json.dumps({
                        "command": "uname -sr && pwd && ls -1",
                        "description": "Check the cloud machine",
                    }),
                },
            }],
        }
    else:
        tool_output = ""
        for m in messages:
            if m.get("role") == "tool":
                tool_output = m.get("content") or ""
        message = {
            "role": "assistant",
            "content": "This is a cloud machine, not your laptop. `uname` says "
                       f"{tool_output.splitlines()[0] if tool_output else 'unknown'}, "
                       "and the working directory is empty apart from the files below. "
                       "Nothing was run locally.",
        }
    return web.json_response({"choices": [{"message": message, "index": 0}]})


async def create_acp_session(request: web.Request) -> web.Response:
    auth = request.headers.get("Authorization", "")
    if auth != f"Bearer {USER_KEY}":
        return web.json_response({"error": "Not authenticated"}, status=401)
    session_id = str(uuid.uuid4())
    base = f"{request.scheme}://{request.host}"
    ws_base = base.replace("http", "ws", 1)
    runner_url = f"{ws_base}/relay/runner/{session_id}"
    browser_url = f"{ws_base}/relay/browser/{session_id}"
    sessions[session_id] = {"runner": None, "browser": None}
    os.makedirs(ACP_CWD, exist_ok=True)
    proc = await asyncio.create_subprocess_exec(
        sys.executable, BRIDGE, runner_url,
        env={**os.environ, "ACP_CWD": ACP_CWD},
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.STDOUT,
    )
    print(f"api: started bridge pid={proc.pid} for session {session_id}", flush=True)

    async def log_bridge():
        assert proc.stdout is not None
        async for line in proc.stdout:
            print("bridge| " + line.decode(errors="replace").rstrip(), flush=True)

    asyncio.create_task(log_bridge())
    return web.json_response({"session_id": session_id, "relay_url": browser_url})


async def relay_ws(request: web.Request) -> web.WebSocketResponse:
    role = request.match_info["role"]
    session_id = request.match_info["session"]
    session = sessions.setdefault(session_id, {"runner": None, "browser": None})
    ws = web.WebSocketResponse(max_msg_size=8 * 1024 * 1024)
    await ws.prepare(request)
    session[role] = ws
    print(f"relay: {role} attached to {session_id}", flush=True)

    async def pump():
        try:
            async for msg in ws:
                if msg.type == web.WSMsgType.BINARY:
                    other = session.get("browser" if role == "runner" else "runner")
                    if other is not None and not other.closed:
                        await other.send_bytes(msg.data)
                elif msg.type == web.WSMsgType.TEXT:
                    other = session.get("browser" if role == "runner" else "runner")
                    if other is not None and not other.closed:
                        await other.send_str(msg.data)
        except Exception:
            pass
        finally:
            other = session.get("browser" if role == "runner" else "runner")
            if other is not None and not other.closed:
                await other.close()

    await pump()
    return ws


def main():
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8944
    app = web.Application()
    app.router.add_get("/v1/models", models)
    app.router.add_post("/v1/chat/completions", chat_completions)
    app.router.add_post("/deca/v1/cloud/acp/session", create_acp_session)
    app.router.add_get("/relay/{role}/{session}", relay_ws)
    print(f"api: listening on http://127.0.0.1:{port}", flush=True)
    web.run_app(app, host="127.0.0.1", port=port, print=None)


if __name__ == "__main__":
    main()
