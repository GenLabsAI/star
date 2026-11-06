// Star ACP web bridge.
//
// Star's `star acp` speaks newline-delimited JSON-RPC over stdio, which
// browsers cannot do. This server (built on existing libraries: express +
// ws) accepts browser WebSocket connections on /acp, spawns one `star acp`
// child per connection, and relays messages both ways:
//
//   WS text frame  ->  child stdin line  ("{...}" + "\n")
//   child stdout line  ->  WS text frame
//
// This framing is safe because both sides use one compact JSON object per
// message (the TS SDK sends one JSON text frame per message; the Go SDK
// reads \n-delimited lines with a 10MB buffer).
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

const app = express();
app.use(express.static(path.join(__dirname, "public")));
app.get("/config.json", (_req, res) => res.json({ cwd: STAR_CWD }));

const server = app.listen(PORT, "0.0.0.0", () => {
  console.log(`[star-acp-web] UI  http://localhost:${PORT}/`);
  console.log(`[star-acp-web] ACP bridge ws://localhost:${PORT}/acp -> ${STAR_BIN} acp (cwd ${STAR_CWD})`);
});

const wss = new WebSocketServer({ server, path: "/acp" });

wss.on("connection", (ws) => {
  console.log("[star-acp-web] browser connected, spawning star acp");
  const child = spawn(STAR_BIN, ["acp"], { cwd: STAR_CWD, stdio: ["pipe", "pipe", "pipe"] });
  child.on("error", (err) => {
    console.error("[star-acp-web] failed to spawn star:", err.message);
    try {
      ws.send(JSON.stringify({ jsonrpc: "2.0", id: null, error: { code: -32000, message: `Failed to start star acp: ${err.message}` } }));
    } catch { /* ignore */ }
    ws.close();
  });
  child.stderr.on("data", (d) => process.stderr.write(`[star-acp] ${d}`));

  const rl = readline.createInterface({ input: child.stdout });
  rl.on("line", (line) => {
    if (line.trim() === "") return;
    if (ws.readyState === ws.OPEN) ws.send(line);
  });

  ws.on("message", (data) => {
    const text = data.toString().trim();
    if (text === "" || child.stdin.destroyed) return;
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
