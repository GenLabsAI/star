"""ACP agent bridge: the cloud-runner side of a Star cloud session.

This is the real agent loop that runs on the cloud machine. It connects
to the relay's runner WebSocket, reads the session config frame Star
sends over the relay, speaks ACP as the agent, calls the configured
model with a tool schema, executes the tools it asks for on this
machine, and streams everything back as ACP session/update frames:

  agent_message_chunk   streamed assistant text
  agent_thought_chunk   streamed reasoning
  tool_call             a tool the cloud agent is about to run
  tool_call_update      that tool's result and terminal status

Star translates those frames into its own message-store writes, so the
TUI renders native tool calls with the same renderers a local run uses.
`_meta.starToolName` carries the Star tool name through ACP's extension
point so the renderer picks the right tool.

Run: python3 acp_bridge.py <runner-ws-url>
Env: none required (model key arrives in the config frame). Logs redact
     any api_key value before printing.
"""

from __future__ import annotations

import asyncio
import glob as globmod
import json
import os
import re
import subprocess
import sys

import websockets

# Seconds between keepalive frames while inference runs. The relay path
# drops an idle direction quickly, so keep traffic flowing faster than
# that. Overridable per run without editing the file.
KEEPALIVE_S = float(os.environ.get("ACP_KEEPALIVE_S", "0.5"))
# Hard cap on agent/tool iterations for one prompt.
MAX_STEPS = int(os.environ.get("ACP_MAX_STEPS", "24"))
# Per-tool timeout for shell commands.
BASH_TIMEOUT_S = int(os.environ.get("ACP_BASH_TIMEOUT_S", "120"))

try:
    import requests
except ImportError:
    requests = None


def redact(obj, secret: str | None):
    if not secret:
        return obj
    text = json.dumps(obj)
    return json.loads(text.replace(secret, "***REDACTED***"))


# The tool schema mirrors the Star tools the TUI already knows how to
# render, so a cloud tool call shows up as the same item a local call
# would.
TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "bash",
            "description": "Run a shell command on the cloud machine and return its output.",
            "parameters": {
                "type": "object",
                "properties": {
                    "command": {"type": "string", "description": "Shell command to run."},
                    "description": {"type": "string", "description": "Short description of what the command does."},
                },
                "required": ["command"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "view",
            "description": "Read a file from the cloud machine. Returns numbered lines.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string"},
                    "offset": {"type": "integer", "description": "First line to read (1-based)."},
                    "limit": {"type": "integer", "description": "Maximum lines to read."},
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "write",
            "description": "Write a file on the cloud machine, replacing its contents.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string"},
                    "content": {"type": "string"},
                },
                "required": ["path", "content"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "edit",
            "description": "Replace an exact string in a file on the cloud machine.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string"},
                    "old_string": {"type": "string"},
                    "new_string": {"type": "string"},
                },
                "required": ["path", "old_string", "new_string"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "ls",
            "description": "List the contents of a directory on the cloud machine.",
            "parameters": {
                "type": "object",
                "properties": {"path": {"type": "string"}},
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "grep",
            "description": "Search file contents on the cloud machine with a regular expression.",
            "parameters": {
                "type": "object",
                "properties": {
                    "pattern": {"type": "string"},
                    "path": {"type": "string", "description": "Directory or file to search."},
                },
                "required": ["pattern"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "glob",
            "description": "Find files on the cloud machine by glob pattern.",
            "parameters": {
                "type": "object",
                "properties": {"pattern": {"type": "string"}},
                "required": ["pattern"],
            },
        },
    },
]

# ACP tool kinds, so clients can pick icons without knowing the tool.
TOOL_KINDS = {
    "bash": "execute",
    "view": "read",
    "ls": "read",
    "write": "edit",
    "edit": "edit",
    "grep": "search",
    "glob": "search",
}

SYSTEM_PROMPT = (
    "You are Star, a coding agent running on a cloud machine. "
    "You are working in {cwd}. Use the provided tools to inspect and modify "
    "files and to run commands. Be concise. Show real results from the "
    "machine rather than guessing."
)


def call_model_sync(model: str, base_url: str, api_key: str, messages: list) -> dict:
    """Blocking OpenAI-compatible chat call with tools. Returns the raw
    message dict (content, tool_calls, reasoning_content)."""
    if not (requests and model and base_url and api_key):
        return {"content": "CLOUD-ACP-OK (canned: no model configured)"}
    resp = requests.post(
        base_url.rstrip("/") + "/chat/completions",
        headers={"Authorization": f"Bearer {api_key}"},
        json={
            "model": model,
            "messages": messages,
            "tools": TOOLS,
            "max_tokens": 2048,
            "temperature": 0,
        },
        timeout=300,
    )
    resp.raise_for_status()
    data = resp.json()
    return data["choices"][0]["message"]


async def call_model(model, base_url, api_key, messages, keepalive=None) -> dict:
    """Call the model in a thread, invoking keepalive() while it thinks.

    The relay path closes an idle browser connection within seconds, so
    inference must be accompanied by traffic.
    """
    task = asyncio.create_task(asyncio.to_thread(call_model_sync, model, base_url, api_key, messages))
    while not task.done():
        try:
            return await asyncio.wait_for(asyncio.shield(task), timeout=KEEPALIVE_S)
        except asyncio.TimeoutError:
            if keepalive is not None:
                await keepalive()
    return await task


def run_tool(name: str, args: dict, cwd: str) -> tuple[str, bool]:
    """Execute one tool call on the cloud machine. Returns (output, is_error)."""
    try:
        if name == "bash":
            proc = subprocess.run(
                args.get("command", ""),
                shell=True,
                cwd=cwd,
                capture_output=True,
                text=True,
                timeout=BASH_TIMEOUT_S,
            )
            out = (proc.stdout or "") + (proc.stderr or "")
            if proc.returncode != 0:
                out += f"\n[exit code {proc.returncode}]"
                return out, True
            return out or "(no output)", False

        if name == "view":
            path = args.get("path", "")
            offset = int(args.get("offset") or 1)
            limit = int(args.get("limit") or 2000)
            with open(path, "r", errors="replace") as fh:
                lines = fh.readlines()
            chunk = lines[max(0, offset - 1): max(0, offset - 1) + limit]
            return "".join(f"{offset + i:>6}  {line}" for i, line in enumerate(chunk)), False

        if name == "write":
            path = args.get("path", "")
            with open(path, "w") as fh:
                fh.write(args.get("content", ""))
            return f"Wrote {path}", False

        if name == "edit":
            path = args.get("path", "")
            old = args.get("old_string", "")
            new = args.get("new_string", "")
            with open(path, "r") as fh:
                data = fh.read()
            if old not in data:
                return f"old_string not found in {path}", True
            with open(path, "w") as fh:
                fh.write(data.replace(old, new, 1))
            return f"Edited {path}", False

        if name == "ls":
            path = args.get("path") or cwd
            return "\n".join(sorted(os.listdir(path))), False

        if name == "grep":
            pattern = args.get("pattern", "")
            root = args.get("path") or cwd
            hits = []
            targets = [root] if os.path.isfile(root) else [
                os.path.join(dp, f)
                for dp, _, fs in os.walk(root)
                for f in fs
            ]
            rx = re.compile(pattern)
            for target in targets:
                try:
                    with open(target, "r", errors="replace") as fh:
                        for lineno, line in enumerate(fh, 1):
                            if rx.search(line):
                                hits.append(f"{target}:{lineno}:{line.rstrip()}")
                                if len(hits) >= 200:
                                    return "\n".join(hits), False
                except (OSError, UnicodeDecodeError):
                    continue
            return "\n".join(hits) or "(no matches)", False

        if name == "glob":
            pattern = args.get("pattern", "*")
            matches = globmod.glob(pattern, recursive=True)
            return "\n".join(sorted(matches)[:300]) or "(no matches)", False

        return f"unknown tool {name}", True
    except subprocess.TimeoutExpired:
        return f"command timed out after {BASH_TIMEOUT_S}s", True
    except Exception as e:  # noqa: BLE001 - surface every tool failure to the model
        return f"{type(e).__name__}: {e}", True


MSG_DATA = 0x00
MSG_EXIT = 0x02


def to_json(raw) -> dict:
    """Parse a relay frame: raw JSON text, or PTY-framed bytes."""
    data = raw.encode() if isinstance(raw, str) else bytes(raw)
    if data and data[0] <= 0x03:
        data = data[1:]
    return json.loads(data.decode())


def prompt_text(params: dict) -> str:
    blocks = params.get("prompt") or []
    parts = []
    for block in blocks:
        if isinstance(block, dict) and block.get("type") == "text":
            parts.append(block.get("text", ""))
    return "\n".join(p for p in parts if p)


async def main(url: str):
    cwd = os.environ.get("ACP_CWD") or os.getcwd()
    async with websockets.connect(url, max_size=8 * 1024 * 1024) as ws:
        # Announce readiness before reading anything: the relay drops a
        # frame sent before this side attaches, so Star waits for this
        # handshake before sending its session config.
        await ws.send(bytes([0x00]) + json.dumps({"type": "acp_ready"}).encode())
        print("bridge: connected to relay, waiting for config", flush=True)
        config = None
        while config is None:
            raw = await asyncio.wait_for(ws.recv(), timeout=120)
            try:
                candidate = to_json(raw)
            except Exception:
                continue  # ping/heartbeat frame: ignore
            if isinstance(candidate, dict) and candidate.get("type") == "acp_init":
                config = candidate
        secret = config.get("api_key")
        print(f"bridge: config: {json.dumps(redact(config, secret))}", flush=True)
        model = config.get("model", "")
        base_url = config.get("base_url", "")

        agent_session = "sess-cloud-1"

        async def send(payload, quiet=False):
            # PTY-style framing (MSG_DATA prefix) like terminal_runner, so
            # any relay viewer handles the bytes uniformly.
            await ws.send(bytes([MSG_DATA]) + json.dumps(payload).encode())
            if not quiet:
                logged = redact(payload, secret)
                print(f"bridge: -> {json.dumps(logged)[:300]}", flush=True)

        async def update(body: dict):
            await send({"jsonrpc": "2.0", "method": "session/update",
                        "params": {"sessionId": agent_session, "update": body}})

        async def keepalive():
            await send({"jsonrpc": "2.0", "method": "session/update",
                        "params": {"sessionId": agent_session,
                                   "update": {"sessionUpdate": "agent_thought_chunk",
                                              "content": {"type": "text", "text": ""}}}},
                       quiet=True)

        async for raw in ws:
            try:
                msg = json.loads(raw if isinstance(raw, str) else raw.decode())
            except Exception:
                continue  # ping/heartbeat frame: ignore
            if "method" not in msg:
                continue
            method, mid = msg["method"], msg.get("id")
            print(f"bridge: <- {method} id={mid}", flush=True)

            if method == "initialize":
                await send({"jsonrpc": "2.0", "id": mid,
                            "result": {"protocolVersion": 1,
                                       "agentCapabilities": {},
                                       "agentInfo": {"name": "cloud-acp-bridge", "version": "0.2.0"},
                                       "authMethods": []}})
            elif method == "session/new":
                await send({"jsonrpc": "2.0", "id": mid, "result": {"sessionId": agent_session}})
            elif method == "session/prompt":
                prompt = prompt_text(msg.get("params") or {})
                messages = [
                    {"role": "system", "content": SYSTEM_PROMPT.format(cwd=cwd)},
                    {"role": "user", "content": prompt},
                ]
                try:
                    for step in range(MAX_STEPS):
                        reply = await call_model(model, base_url, secret, messages, keepalive=keepalive)
                        text = (reply.get("content") or "").strip()
                        if text:
                            for i in range(0, len(text), 400):
                                await update({"sessionUpdate": "agent_message_chunk",
                                              "content": {"type": "text", "text": text[i:i + 400]}})
                        calls = reply.get("tool_calls") or []
                        if not calls:
                            break

                        messages.append({"role": "assistant",
                                         "content": text,
                                         "tool_calls": calls})
                        for call in calls:
                            fn = call.get("function") or {}
                            name = fn.get("name", "")
                            call_id = call.get("id") or f"call_{step}_{len(messages)}"
                            try:
                                args = json.loads(fn.get("arguments") or "{}")
                            except json.JSONDecodeError:
                                args = {}
                            title = args.get("description") or args.get("command") or args.get("path") or name
                            # Start frame: the client adds a native tool-call part.
                            await update({
                                "sessionUpdate": "tool_call",
                                "toolCallId": call_id,
                                "title": title,
                                "kind": TOOL_KINDS.get(name, "other"),
                                "status": "in_progress",
                                "rawInput": args,
                                "_meta": {"starToolName": name},
                            })
                            output, is_error = await asyncio.to_thread(run_tool, name, args, cwd)
                            # Terminal frame: the client records the result.
                            await update({
                                "sessionUpdate": "tool_call_update",
                                "toolCallId": call_id,
                                "status": "failed" if is_error else "completed",
                                "content": [{"type": "content",
                                             "content": {"type": "text", "text": output}}],
                                "rawOutput": output,
                                "_meta": {"starToolName": name},
                            })
                            messages.append({"role": "tool",
                                             "tool_call_id": call_id,
                                             "name": name,
                                             "content": output})
                except Exception as e:  # noqa: BLE001 - report and end the turn
                    print(f"bridge: turn failed: {type(e).__name__}: {e}", flush=True)
                    await update({"sessionUpdate": "agent_message_chunk",
                                  "content": {"type": "text",
                                              "text": f"\n[cloud agent error: {type(e).__name__}: {e}]\n"}})

                await send({"jsonrpc": "2.0", "id": mid, "result": {"stopReason": "end_turn"}})
                # Grace period: the relay tears down the browser side when
                # the runner goes away, so linger to let Star drain every
                # frame before the session ends.
                print("bridge: prompt complete, lingering 5s for drain", flush=True)
                await asyncio.sleep(5)
                await ws.send(bytes([MSG_EXIT, 0]))
                print("bridge: exiting")
                return
            else:
                await send({"jsonrpc": "2.0", "id": mid,
                            "error": {"code": -32601, "message": f"unknown method {method}"}})


if __name__ == "__main__":
    asyncio.run(main(sys.argv[1]))
