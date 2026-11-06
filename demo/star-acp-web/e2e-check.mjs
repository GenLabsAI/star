// Throwaway end-to-end check: SDK client -> WS bridge -> `star acp`.
// Run: node e2e-check.mjs "say PONG"
import * as acp from "@agentclientprotocol/sdk";
import { createWebSocketStream } from "@agentclientprotocol/sdk/experimental/ws-client";
import WebSocket from "ws";

globalThis.WebSocket = WebSocket;

const stream = createWebSocketStream("ws://127.0.0.1:3003/acp", { WebSocket });
const app = acp
  .client({ name: "e2e-check" })
  .onRequest(acp.methods.client.session.requestPermission, async (req) => {
    console.log("PERMISSION (auto-deny):", JSON.stringify(req.params.toolCall).slice(0, 200));
    return { outcome: { outcome: "cancelled" } };
  });

await app.connectWith(stream, async (ctx) => {
  const init = await ctx.request(acp.methods.agent.initialize, {
    protocolVersion: acp.PROTOCOL_VERSION,
    clientCapabilities: {},
    clientInfo: { name: "e2e-check", version: "0.0.1" },
  });
  console.log("agentInfo:", JSON.stringify(init.agentInfo));
  await ctx.request(acp.methods.agent.authenticate, { methodId: "local" }).catch((e) => {
    console.log("auth note:", e.message);
  });
  const session = await ctx.buildSession({ cwd: "/tmp/star-work", mcpServers: [] }).start();
  console.log("session:", session.sessionId);

  const promptText = process.argv[2] ?? "Reply with exactly: PONG";
  const done = (async () => {
    for (;;) {
      const msg = await session.nextUpdate();
      if (msg.kind === "stop") {
        console.log("STOP:", msg.stopReason);
        return;
      }
      const u = msg.update;
      if (u.sessionUpdate === "agent_message_chunk") {
        const t = (Array.isArray(u.content) ? u.content : [u.content])
          .map((b) => (typeof b === "string" ? b : b.text ?? ""))
          .join("");
        process.stdout.write(t);
      } else {
        console.log(`\n[${u.sessionUpdate}]`, JSON.stringify(u).slice(0, 300));
      }
    }
  })();
  const resp = await session.prompt(promptText);
  console.log("\nPROMPT RESPONSE:", JSON.stringify(resp).slice(0, 500));
  await done;
  process.exit(0);
});
