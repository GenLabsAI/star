# ACP UI for Star

Drives Star from any browser using the upstream
[acp-ui](https://github.com/formulahendry/acp-ui) web client (Vue 3 build)
talking to `star acp` over the Agent Client Protocol.

## How it fits together

- `demo/acp-ui/` — pristine shallow clone of upstream acp-ui (git-ignored,
  not committed). Its `dist-web/` bundle is built with `npm run build:web`.
- `demo/acp-ui-star/server.mjs` — this glue: serves the `dist-web` bundle
  and bridges each browser tab's WebSocket (`/acp`) to one `star acp`
  child process (WS text frame <-> stdio line, 1:1). One origin, so `wss://`
  pages can talk to the agent with no mixed-content issues.
- `internal/acp/` — Star's ACP server (`star acp`), ported from Crush
  PR #2450 and rebranded to star/Star.

## Reproduce from scratch

```sh
# 1. Star binary with the ACP server
go build -o /tmp/star-acp .

# 2. Upstream web UI (pinned commit recorded below)
git clone --depth 1 https://github.com/formulahendry/acp-ui.git demo/acp-ui
cd demo/acp-ui && npm install && npm run build:web && cd -

# 3. Bridge deps and start (port 3003)
cd demo/acp-ui-star && npm install
STAR_BIN=/tmp/star-acp STAR_CWD=/tmp/star-work PORT=3003 node server.mjs
```

Upstream commit verified against: acp-ui v0.1.16 (`@agentclientprotocol/sdk`
0.13.1, protocol version 1 — matches Star's Go SDK).

## Connect a browser

1. Open the served page.
2. Settings (gear) -> Add Agent: name `Star`, transport websocket,
   URL `wss://<this-host>/acp` (or `ws://` for plain-http origins).
3. Working Directory: absolute path **on the machine running `star acp`**
   (e.g. `/tmp/star-work`), then New Session.
4. Pick the `Local` auth method when prompted.

`DEBUG_FRAMES=1` in the env logs every raw JSON-RPC frame both ways.

## Verified working (browser-tested 2026-11-06)

- WSS connect through the preview proxy, `initialize` (star/Star),
  `authenticate` (Local), `session/new`, session resume via `session/load`.
- `session/prompt` round-trip: `/toggle_yolo` streams back an
  `agent_message_chunk` and renders in chat (see `verified-chat.png`).
- Slash commands advertised via `available_commands_update`; Ask/Code
  modes advertised; clean `-32602` (no crash) when no LLM provider is
  configured; per-frame log in acp-ui's Traffic Monitor.

Live agent turns need an LLM provider configured where `star acp` runs;
without one, plain prompts return `agent is not configured: no LLM provider
is set up` while commands, sessions, modes, and auth all work.
