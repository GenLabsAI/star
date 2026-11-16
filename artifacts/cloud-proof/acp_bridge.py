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

Lifecycle (idle, stash, resume): after each turn the bridge persists its
conversation to ACP_STATE_DIR (keyed by the acp_init resume_key, never
containing secrets) and goes idle instead of exiting. A follow-up
session/prompt on the same relay connection reuses the live process
with full history; ACP_IDLE_S without one snapshots and exits, and a
later job with the same resume_key restores from disk as if nothing
happened. Env knobs: ACP_IDLE_S (default 1200), ACP_IDLE_KEEPALIVE_S
(default 30), ACP_STATE_DIR (default ~/.cloud-agent).
"""

from __future__ import annotations

import asyncio
import glob as globmod
import hashlib
import json
import os
import re
import shutil
import socket
import subprocess
import sys
import tempfile
import time

import websockets

# Seconds between keepalive frames while inference runs. The relay path
# drops an idle direction quickly, so keep traffic flowing faster than
# that. Overridable per run without editing the file.
KEEPALIVE_S = float(os.environ.get("ACP_KEEPALIVE_S", "0.5"))
# Hard cap on agent/tool iterations for one prompt.
MAX_STEPS = int(os.environ.get("ACP_MAX_STEPS", "24"))
# Per-tool timeout for shell commands.
BASH_TIMEOUT_S = int(os.environ.get("ACP_BASH_TIMEOUT_S", "120"))
# Idle hold after going idle: keep the relay session open this long
# waiting for a follow-up prompt before snapshotting state and exiting.
# The 20-minute default matches the product lifecycle (idle, stash,
# resume); a new full window starts after every completed turn.
ACP_IDLE_S = float(os.environ.get("ACP_IDLE_S", "1200"))
# Ticks between idle keepalives: the relay path kills a direction after
# ~30s of silence (seen in prod), so keep traffic flowing well under that.
ACP_IDLE_KEEPALIVE_S = float(os.environ.get("ACP_IDLE_KEEPALIVE_S", "15"))
# Seconds between HTTPS heartbeats proving this runner alive. The server
# only routes follow-ups to runners heard from recently, so a dead
# socket can never attract a new turn: the next request dispatches fresh
# and restores from the stash instead.
ACP_HEARTBEAT_S = float(os.environ.get("ACP_HEARTBEAT_S", "60"))
# Freshness window the server applies to heartbeats (seconds).
HEARTBEAT_FRESH_S = 240
# Directory for the persisted conversation. It never holds secrets: the
# model key and git token arrive fresh in every acp_init frame.
ACP_STATE_DIR = os.environ.get("ACP_STATE_DIR") or os.path.join(
    os.path.expanduser("~"), ".cloud-agent")
# State file version, bumped when the schema changes.
STATE_VERSION = 1

try:
    import requests
except ImportError:
    requests = None


def redact(obj, secrets):
    """Replace every secret value with ***REDACTED***. `secrets` is a
    single value or a list; falsy entries are ignored."""
    if isinstance(secrets, str):
        secrets = [secrets]
    live = [s for s in secrets if s]
    if not live:
        return obj
    text = json.dumps(obj)
    for s in live:
        text = text.replace(s, "***REDACTED***")
    return json.loads(text)


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
    "machine rather than guessing. "
    "If you need to push to a git remote, plain `git push` already "
    "authenticates via GIT_ASKPASS in your environment: never print, echo, "
    "or embed credentials in commands, files, or messages."
)


def call_model_sync(model: str, base_url: str, api_key: str, messages: list) -> dict:
    """Blocking OpenAI-compatible chat call with tools. Returns the raw
    message dict (content, tool_calls, reasoning_content)."""
    if not (requests and model and base_url and api_key):
        return {"content": "CLOUD-ACP-OK (canned: no model configured)"}
    # Retry transient gateway errors: a long cloud task should not die
    # on a brief gateway outage. Six attempts with growing backoff ride
    # out ~1 minute of consecutive 502/503/504s.
    last_exc: Exception | None = None
    for attempt in range(6):
        try:
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
        except Exception as exc:  # noqa: BLE001 - retry transient failures
            last_exc = exc
            status = getattr(getattr(exc, "response", None), "status_code", None)
            transient = status is None or status in (408, 425, 429, 500, 502, 503, 504)
            if not transient or attempt == 5:
                raise
            time.sleep(min(2 ** attempt, 30))
    raise last_exc  # unreachable: loop either returns or raises


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


def run_tool(name: str, args: dict, cwd: str, extra_env: dict | None = None) -> tuple[str, bool]:
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
                env={**os.environ, **(extra_env or {})},
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


def state_file(state_dir: str, resume_key: str) -> str:
    """State path for a resume key. The filename carries only a hash, so
    one session's conversation can never be picked up by another."""
    digest = hashlib.sha256(resume_key.encode()).hexdigest()[:32]
    return os.path.join(state_dir, f"agent-state-{digest}.json")


def load_conversation(state_dir: str, resume_key: str) -> dict | None:
    """Return the saved conversation for resume_key, or None to start
    fresh. Any corruption or mismatch fails closed to a new session."""
    if not resume_key:
        return None
    try:
        with open(state_file(state_dir, resume_key), "r") as fh:
            saved = json.load(fh)
    except (OSError, ValueError):
        return None
    if (not isinstance(saved, dict) or saved.get("version") != STATE_VERSION
            or saved.get("resume_key") != resume_key):
        return None
    if not isinstance(saved.get("messages"), list):
        return None
    return saved


def save_conversation(state_dir: str, resume_key: str, agent_session: str,
                      model: str, base_url: str, cwd: str, messages: list) -> None:
    """Persist the conversation so a later job resumes with full history.
    Secrets are never written: api_key/git_token arrive fresh in each
    acp_init frame. The write is atomic (tmp + replace)."""
    if not resume_key:
        return
    try:
        os.makedirs(state_dir, exist_ok=True)
        path = state_file(state_dir, resume_key)
        with open(path + ".tmp", "w") as fh:
            json.dump({"version": STATE_VERSION, "resume_key": resume_key,
                       "agent_session": agent_session, "model": model,
                       "base_url": base_url, "cwd": cwd,
                       "messages": messages, "updated_at": time.time()}, fh)
        os.replace(path + ".tmp", path)
    except OSError as exc:
        print(f"bridge: state save failed: {exc}", flush=True)


def api_base_of(relay_url: str) -> str:
    """HTTPS API base derived from the relay URL (same host by
    construction: the server mints both)."""
    base = relay_url.replace("wss://", "https://").replace("ws://", "http://")
    return base.split("/terminal/")[0]


def relay_sid_of(relay_url: str) -> str:
    return relay_url.rstrip("/").rsplit("/", 1)[-1]


def upload_stash(state_dir: str, resume_key: str) -> None:
    """Encrypt the saved state and upsert the chat-keyed release asset, so
    a later job restores even if this runner dies before job end. Needs
    GH_TOKEN + DECA_AGENT_STATE_PASSWORD (the workflow provides both to
    the bridge step); a silent no-op without them — the job-end Save step
    remains the fallback. Never logs secrets."""
    if not resume_key or not shutil.which("gh"):
        return
    if not os.environ.get("GH_TOKEN") or not os.environ.get("DECA_AGENT_STATE_PASSWORD"):
        return
    path = state_file(state_dir, resume_key)
    if not os.path.isfile(path):
        return
    tag = os.path.basename(path)[: -len(".json")]
    enc = path + ".upload.enc"
    try:
        subprocess.run(["openssl", "enc", "-aes-256-cbc", "-pbkdf2", "-salt",
                        "-in", path, "-out", enc, "-pass", "env:DECA_AGENT_STATE_PASSWORD"],
                       check=True, capture_output=True, timeout=60)
        subprocess.run(["gh", "release", "delete", tag, "--yes"],
                       capture_output=True, timeout=60)
        subprocess.run(["gh", "release", "create", tag, enc, "--title", tag,
                        "--notes", "Encrypted Deca ACP agent state", "--latest=false"],
                       check=True, capture_output=True, timeout=120)
        print(f"bridge: stashed state as release {tag}", flush=True)
    except Exception as e:
        print(f"bridge: stash upload skipped: {type(e).__name__}", flush=True)
    finally:
        try:
            os.remove(enc)
        except OSError:
            pass


async def main(url: str):
    cwd = os.environ.get("ACP_CWD") or os.getcwd()
    # ping_interval=None: the relay path does not answer protocol pings,
    # and the library would otherwise self-close a healthy session after
    # ~40s of receive silence with 1011. Liveness is carried by the
    # app-level thought-chunk keepalives instead, which also flow during
    # tool execution (see below), so neither direction ever idles.
    async with websockets.connect(url, max_size=8 * 1024 * 1024, ping_interval=None) as ws:
        # Announce readiness before reading anything: the relay drops a
        # frame sent before this side attaches, so Star waits for this
        # handshake before sending its session config.
        await ws.send(bytes([0x00]) + json.dumps({"type": "acp_ready"}).encode())
        print("bridge: connected to relay, waiting for config", flush=True)
        try:
            sock = ws.transport.get_extra_info("socket")
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)
            if hasattr(socket, "TCP_KEEPIDLE"):
                sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_KEEPIDLE, 60)
                sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_KEEPINTVL, 10)
                sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_KEEPCNT, 6)
        except Exception:
            pass  # keepalive is best-effort; app frames carry liveness
        config = None
        while config is None:
            try:
                raw = await asyncio.wait_for(ws.recv(), timeout=120)
            except asyncio.TimeoutError:
                continue  # no Star yet; the job timeout bounds the wait
            try:
                candidate = to_json(raw)
            except Exception:
                continue  # ping/heartbeat frame: ignore
            if isinstance(candidate, dict) and candidate.get("type") == "acp_init":
                config = candidate
        secret = config.get("api_key")
        git_token = config.get("git_token") or ""
        secrets = [secret, git_token]
        print(f"bridge: config: {json.dumps(redact(config, secrets))}", flush=True)
        model = config.get("model", "")
        base_url = config.get("base_url", "")
        # Stable key linking this runner to later jobs on the same chat
        # session. The dispatcher (and Star, eventually) passes the same
        # key every turn; without one each job starts fresh.
        resume_key = (config.get("resume_key") or "").strip()
        # Push credential for the agent's bash tool: a GIT_ASKPASS helper
        # backed by a 0600 file, so the token never appears in commands,
        # remotes, or logs.
        tool_env: dict[str, str] = {}
        if git_token:
            cred_dir = tempfile.mkdtemp(prefix="star-git-")
            token_path = os.path.join(cred_dir, "token")
            with open(token_path, "w") as fh:
                fh.write(git_token)
            os.chmod(token_path, 0o600)
            askpass_path = os.path.join(cred_dir, "askpass.sh")
            with open(askpass_path, "w") as fh:
                fh.write(
                    "#!/bin/sh\n"
                    f'TOKEN_FILE="{token_path}"\n'
                    'case "$1" in\n'
                    '  *Username*) echo "x-access-token" ;;\n'
                    '  *) cat "$TOKEN_FILE" ;;\n'
                    "esac\n"
                )
            os.chmod(askpass_path, 0o700)
            tool_env = {
                "GIT_ASKPASS": askpass_path,
                "GIT_ASKPASS_REQUIRE": "force",
                "GIT_TERMINAL_PROMPT": "0",
            }
            print("bridge: git push credential configured via GIT_ASKPASS", flush=True)

        agent_session = "sess-cloud-1"
        conversation: list = []
        api_base = api_base_of(url)
        relay_sid = relay_sid_of(url)
        last_hb = 0.0
        if resume_key:
            saved = load_conversation(ACP_STATE_DIR, resume_key)
            if saved is not None:
                conversation = saved.get("messages") or []
                agent_session = saved.get("agent_session") or agent_session
                print(f"bridge: resumed {len(conversation)} messages "
                      f"(session {agent_session})", flush=True)
            else:
                print("bridge: no saved state, starting fresh", flush=True)

        async def send(payload, quiet=False):
            # PTY-style framing (MSG_DATA prefix) like terminal_runner, so
            # any relay viewer handles the bytes uniformly.
            await ws.send(bytes([MSG_DATA]) + json.dumps(payload).encode())
            if not quiet:
                logged = redact(payload, secrets)
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

        def save():
            save_conversation(ACP_STATE_DIR, resume_key, agent_session,
                              model, base_url, cwd, conversation)

        # One read loop with an idle deadline: ACP_IDLE_S without a prompt
        # snapshots and exits; idle keepalives hold the relay direction
        # open while waiting. A fresh full window starts after every
        # completed turn (see the session/prompt branch).
        idle_deadline = time.monotonic() + ACP_IDLE_S
        while True:
            # Tick down to the deadline so expiry fires on time instead of
            # up to one keepalive interval late.
            remaining = idle_deadline - time.monotonic()
            if remaining <= 0:
                print("bridge: idle expired, saving state and exiting", flush=True)
                save()
                upload_stash(ACP_STATE_DIR, resume_key)
                try:
                    await ws.send(bytes([MSG_EXIT, 0]))
                except Exception:
                    pass
                return
            now = time.monotonic()
            if now - last_hb >= ACP_HEARTBEAT_S:
                last_hb = now
                try:
                    if requests is not None and api_base and relay_sid:
                        requests.post(api_base + f"/deca/v1/terminal/session/{relay_sid}/heartbeat",
                                      json={"run_id": os.environ.get("GITHUB_RUN_ID")},
                                      timeout=10)
                except Exception as e:
                    print(f"bridge: heartbeat failed: {type(e).__name__}", flush=True)
            try:
                raw = await asyncio.wait_for(ws.recv(), timeout=min(ACP_IDLE_KEEPALIVE_S, remaining))
            except asyncio.TimeoutError:
                await keepalive()
                continue
            except websockets.exceptions.ConnectionClosed:
                print("bridge: relay closed, saving state and exiting", flush=True)
                save()
                upload_stash(ACP_STATE_DIR, resume_key)
                return
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
                params = msg.get("params")
                requested = params.get("sessionId") if isinstance(params, dict) else None
                if requested and requested != agent_session:
                    agent_session = requested
                    conversation = []
                    save()
                    print(f"bridge: new agent session {agent_session}, conversation reset", flush=True)
                await send({"jsonrpc": "2.0", "id": mid, "result": {"sessionId": agent_session}})
            elif method == "session/prompt":
                prompt = prompt_text(msg.get("params") or {})
                # Conversation persists across prompts on this connection
                # (and across jobs via save/load): the system prompt leads
                # once, then every turn appends user/assistant/tool entries.
                if not conversation:
                    conversation.append(
                        {"role": "system",
                         "content": SYSTEM_PROMPT.format(cwd=cwd)})
                conversation.append({"role": "user", "content": prompt})
                try:
                    for step in range(MAX_STEPS):
                        reply = await call_model(model, base_url, secret, conversation, keepalive=keepalive)
                        text = (reply.get("content") or "").strip()
                        if text:
                            for i in range(0, len(text), 400):
                                await update({"sessionUpdate": "agent_message_chunk",
                                              "content": {"type": "text", "text": text[i:i + 400]}})
                        calls = reply.get("tool_calls") or []
                        if not calls:
                            break

                        conversation.append({"role": "assistant",
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
                            async def _tool_keepalive():
                                try:
                                    while True:
                                        await keepalive()
                                        await asyncio.sleep(0.5)
                                except asyncio.CancelledError:
                                    pass
                            _ka = asyncio.create_task(_tool_keepalive())
                            try:
                                output, is_error = await asyncio.to_thread(run_tool, name, args, cwd, tool_env)
                            finally:
                                _ka.cancel()
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
                            conversation.append({"role": "tool",
                                                 "tool_call_id": call_id,
                                                 "name": name,
                                                 "content": output})
                except Exception as e:  # noqa: BLE001 - report and end the turn
                    print(f"bridge: turn failed: {type(e).__name__}: {e}", flush=True)
                    await update({"sessionUpdate": "agent_message_chunk",
                                  "content": {"type": "text",
                                              "text": f"\n[cloud agent error: {type(e).__name__}: {e}]\n"}})

                await send({"jsonrpc": "2.0", "id": mid, "result": {"stopReason": "end_turn"}})
                # Turn done: persist so a later job resumes with full
                # history, then go idle. The next session/prompt reuses
                # this connection; ACP_IDLE_S without one snapshots and
                # exits (see the read loop). No MSG_EXIT here: leaving
                # would tear down the relay session a follow-up needs.
                save()
                upload_stash(ACP_STATE_DIR, resume_key)
                idle_deadline = time.monotonic() + ACP_IDLE_S
                print(f"bridge: prompt complete, idle up to {ACP_IDLE_S:g}s for follow-up", flush=True)
            else:
                await send({"jsonrpc": "2.0", "id": mid,
                            "error": {"code": -32601, "message": f"unknown method {method}"}})


if __name__ == "__main__":
    asyncio.run(main(sys.argv[1]))
