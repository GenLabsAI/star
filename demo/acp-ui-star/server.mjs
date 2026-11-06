// Serves the acp-ui web build and bridges each browser tab to `star acp`.
//
// acp-ui (web) only talks to remote agents over ws(s)://, one JSON-RPC
// frame per message. `star acp` speaks newline-delimited JSON-RPC over
// stdio. This server (express + ws) hosts both on one origin so there is
// no mixed-content problem and no CORS:
//
//   GET /            -> acp-ui bundle (../acp-ui/dist-web)
//   WS  /acp         -> one `star acp` child per connection
//     WS text frame  -> child stdin line  ("{...}" + "\n")
//     child stdout line -> WS text frame
//
// Env:
//   STAR_BIN  path to the star binary (default /tmp/star-acp)
//   STAR_CWD  working dir the agent runs in (default /tmp/star-work)
//   PORT      HTTP port (default 3003)
import express from "express";
import { WebSocketServer } from "ws";
import { spawn } from "node:child_process";
import readline from "node:readline";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const STAR_BIN = process.env.STAR_BIN ?? "/tmp/star-acp";
const STAR_CWD = process.env.STAR_CWD ?? "/tmp/star-work";
const PORT = Number(process.env.PORT ?? 3003);
const DIST = path.join(__dirname, "..", "acp-ui", "dist-web");

const app = express();
app.use(express.static(DIST));
// SPA fallback: acp-ui uses client-side routing.
app.use((_req, res) => res.sendFile(path.join(DIST, "index.html")));

const server = app.listen(PORT, "0.0.0.0", () => {
  console.log(`[acp-ui-star] UI  http://localhost:${PORT}/`);
  console.log(`[acp-ui-star] ACP bridge ws://localhost:${PORT}/acp -> ${STAR_BIN} acp (cwd ${STAR_CWD})`);
});

const wss = new WebSocketServer({ server, path: "/acp" });

wss.on("connection", (ws) => {
  console.log("[acp-ui-star] browser connected, spawning star acp");
  const child = spawn(STAR_BIN, ["acp"], { cwd: STAR_CWD, stdio: ["pipe", "pipe", "pipe"] });
  child.on("error", (err) => {
    console.error("[acp-ui-star] failed to spawn star:", err.message);
    try {
      ws.send(JSON.stringify({ jsonrpc: "2.0", id: null, error: { code: -32000, message: `Failed to start star acp: ${err.message}` } }));
    } catch { /* ignore */ }
    ws.close();
  });
  child.stderr.on("data", (d) => process.stderr.write(`[star-acp] ${d}`));

  const rl = readline.createInterface({ input: child.stdout });
  rl.on("line", (line) => {
    if (line.trim() === "") return;
    if (process.env.DEBUG_FRAMES) process.stderr.write(`[frame star->ws] ${line}\n`);
    if (ws.readyState === ws.OPEN) ws.send(line);
  });

  ws.on("message", (data) => {
    const text = data.toString().trim();
    if (text === "" || child.stdin.destroyed) return;
    if (process.env.DEBUG_FRAMES) process.stderr.write(`[frame ws->star] ${text}\n`);
    child.stdin.write(text + "\n");
  });

  const cleanup = () => {
    try {
      child.kill();
    } catch { /* ignore */ }
  };
  ws.on("close", cleanup);
  ws.on("error", cleanup);
  child.on("exit", () => {
    try {
      ws.close();
    } catch { /* ignore */ }
  });
});
