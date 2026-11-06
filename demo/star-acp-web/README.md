# Star ACP Web UI

Drives Star from the browser over the **Agent Client Protocol** — no custom
protocol code. Uses the official [`@agentclientprotocol/sdk`](https://www.npmjs.com/package/@agentclientprotocol/sdk)
(existing library) in the browser, plus `express` + `ws` for hosting and the
WebSocket→stdio bridge.

## Architecture

```
browser (app.bundle.js) --WS /acp--> server.mjs --stdio--> `star acp`
   @agentclientprotocol/sdk            one child per       coder/acp-go-sdk
   createWebSocketStream               browser tab         newline JSON-RPC
```

WS text frames and stdio lines are both one compact JSON-RPC object per
message, so the bridge relays them 1:1.

## Run

```sh
# 1. Build star with ACP support (branch feat/acp-2450-port or newer)
go build -o /tmp/star-acp .

# 2. Start the web UI (defaults: STAR_BIN=/tmp/star-acp, cwd /tmp/star-work, port 3003)
cd demo/star-acp-web
npm install
npm start
# or: STAR_BIN=/path/to/star STAR_CWD=/tmp/my-project PORT=3003 npm start
```

Open `http://localhost:3003/`: **Connect** → **New session** → send a prompt.
Tool approvals render as permission cards (allowed/denied via the agent's own
options), plans/diffs/tool status stream live, slash commands from the agent
appear as clickable chips.

## Notes

- Each browser tab spawns its own `star acp` child; closing the tab kills it.
- Agent prompts need an LLM provider configured in Star, same as the TUI.
- `npm run build` (esbuild, via `prestart`) bundles `src/app.js` + the SDK +
  `marked` into `public/app.bundle.js`.
