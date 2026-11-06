// Star ACP web client.
//
// Drives Star through the Agent Client Protocol using the official
// @agentclientprotocol/sdk (existing library, not custom protocol code).
// Transport is createWebSocketStream to this demo server's /acp bridge,
// which relays each WS text frame to `star acp` stdio and back.
import * as acp from "@agentclientprotocol/sdk";
import { createWebSocketStream } from "@agentclientprotocol/sdk/experimental/ws-client";
import { marked } from "marked";

const $ = (id) => document.getElementById(id);
const esc = (s) =>
  String(s ?? "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

const state = {
  ctx: null,
  session: null,
  sessionId: null,
  pumping: false,
  turnText: "",
  turnEl: null,
  tools: new Map(),
  pendingPermission: null,
};

function setStatus(text, ok) {
  const el = $("status");
  el.textContent = text;
  el.className = "pill " + (ok ? "ok" : ok === false ? "bad" : "");
}

function logLine(text) {
  const el = $("events");
  el.textContent += text + "\n";
  el.scrollTop = el.scrollHeight;
}

// ---------- transcript rendering ----------

function addMsg(role, html) {
  const wrap = document.createElement("div");
  wrap.className = "msg " + role;
  wrap.innerHTML = html;
  $("msgs").appendChild(wrap);
  wrap.scrollIntoView({ block: "end" });
  return wrap;
}

function textContentOf(block) {
  if (!block) return "";
  if (typeof block === "string") return block;
  if (block.type === "text" && typeof block.text === "string") return block.text;
  return "";
}

function updateTextOf(update) {
  const c = update.content;
  if (Array.isArray(c)) return c.map(textContentOf).join("");
  return textContentOf(c);
}

function endStreamingTurn() {
  if (state.turnEl && state.turnText.trim() !== "") {
    state.turnEl.querySelector(".body").innerHTML = marked.parse(state.turnText);
  }
  state.turnEl = null;
  state.turnText = "";
}

function renderToolContent(container, contentArray) {
  for (const item of contentArray ?? []) {
    if (item.type === "content") {
      const t = textContentOf(item.content);
      if (t) {
        const pre = document.createElement("pre");
        pre.textContent = t.slice(0, 4000);
        container.appendChild(pre);
      }
    } else if (item.type === "diff") {
      const details = document.createElement("details");
      details.open = true;
      const path = item.path ?? item.filePath ?? "diff";
      details.innerHTML = `<summary>diff: ${esc(path)}</summary>`;
      const pre = document.createElement("pre");
      pre.textContent = `--- before\n${item.oldText ?? ""}\n+++ after\n${item.newText ?? item.new_text ?? ""}`.slice(0, 6000);
      details.appendChild(pre);
      container.appendChild(details);
    } else if (item.type === "terminal") {
      const div = document.createElement("div");
      div.className = "meta";
      div.textContent = `terminal: ${item.terminalId ?? ""}`;
      container.appendChild(div);
    } else {
      const pre = document.createElement("pre");
      pre.textContent = JSON.stringify(item).slice(0, 2000);
      container.appendChild(pre);
    }
  }
}

function upsertToolCard(tc) {
  let card = state.tools.get(tc.toolCallId);
  if (!card) {
    card = document.createElement("div");
    card.className = "msg tool";
    card.innerHTML = `<div class="meta"></div><div class="title"></div><div class="content"></div>`;
    $("msgs").appendChild(card);
    card.scrollIntoView({ block: "end" });
    state.tools.set(tc.toolCallId, card);
  }
  if (tc.title != null) card.querySelector(".title").textContent = tc.title;
  card.querySelector(".meta").textContent =
    `🔧 ${tc.name ?? "tool"} · ${tc.kind ?? "?"} · ${tc.status ?? "pending"}`;
  card.dataset.status = tc.status ?? "pending";
  if (tc.content) {
    const box = card.querySelector(".content");
    box.innerHTML = "";
    renderToolContent(box, tc.content);
  }
}

function renderUpdate(update) {
  switch (update.sessionUpdate) {
    case "agent_message_chunk": {
      const t = updateTextOf(update);
      if (!state.turnEl) {
        state.turnEl = addMsg("agent", `<div class="meta">star</div><div class="body"></div>`);
      }
      state.turnText += t;
      state.turnEl.querySelector(".body").textContent = state.turnText;
      break;
    }
    case "agent_thought_chunk": {
      const t = updateTextOf(update);
      if (t.trim()) logLine(`thought: ${t.slice(0, 200)}`);
      break;
    }
    case "tool_call":
    case "tool_call_update":
      upsertToolCard(update);
      break;
    case "plan": {
      const entries = update.plan?.entries ?? update.entries ?? [];
      $("plan").innerHTML =
        entries.map((e) => `<li><b>${esc(e.status ?? "")}</b> ${esc(e.content ?? e.title ?? "")}</li>`).join("") ||
        "<li class='dim'>empty plan</li>";
      break;
    }
    case "available_commands_update": {
      const cmds = update.availableCommands ?? [];
      $("cmds").innerHTML = "";
      for (const c of cmds) {
        const b = document.createElement("button");
        b.className = "chip";
        b.textContent = "/" + c.name;
        b.title = c.description ?? "";
        b.onclick = () => {
          $("prompt").value = `/${c.name} `;
          $("prompt").focus();
        };
        $("cmds").appendChild(b);
      }
      break;
    }
    case "current_mode_update":
      if (update.currentModeId) $("modeLabel").textContent = `mode: ${update.currentModeId}`;
      break;
    default:
      logLine(`update ${update.sessionUpdate}: ` + JSON.stringify(update).slice(0, 300));
  }
}

// ---------- permissions ----------

function askPermission(toolCall, options) {
  return new Promise((resolve) => {
    const box = $("perms");
    box.querySelector(".empty")?.remove();
    const div = document.createElement("div");
    div.className = "card perm";
    div.innerHTML =
      `<div><b>Permission:</b> ${esc(toolCall.title ?? toolCall.name ?? toolCall.toolCallId)}</div>` +
      `<div class="dim small">${esc(toolCall.kind ?? "")} · ${esc(toolCall.status ?? "")}</div>`;
    const row = document.createElement("div");
    row.className = "row";
    for (const opt of options) {
      const b = document.createElement("button");
      b.textContent = opt.name ?? opt.optionId;
      b.title = opt.optionId;
      b.onclick = () => {
        div.remove();
        resolve(opt.optionId);
      };
      row.appendChild(b);
    }
    const cancel = document.createElement("button");
    cancel.textContent = "cancel turn";
    cancel.onclick = () => {
      div.remove();
      resolve(null);
    };
    row.appendChild(cancel);
    div.appendChild(row);
    box.appendChild(div);
  });
}

// ---------- ACP wiring (official SDK) ----------

async function connect() {
  if (state.ctx) return;
  setStatus("connecting…");
  const stream = createWebSocketStream(`ws://${location.host}/acp`);
  const app = acp
    .client({ name: "star-acp-web", title: "Star ACP Web UI" })
    .onRequest(acp.methods.client.session.requestPermission, async (req) => {
      const { toolCall, options } = req.params;
      logLine(`permission: ${toolCall.title ?? toolCall.name} (${options.map((o) => o.optionId).join("/")})`);
      const optionId = await askPermission(toolCall, options);
      if (!optionId) return { outcome: { outcome: "cancelled" } };
      return { outcome: { outcome: "selected", optionId } };
    });

  await app.connectWith(stream, async (ctx) => {
    state.ctx = ctx;
    const initialized = await ctx.request(acp.methods.agent.initialize, {
      protocolVersion: acp.PROTOCOL_VERSION,
      clientCapabilities: {},
      clientInfo: { name: "star-acp-web", version: "0.1.0" },
    });
    const info = initialized.agentInfo ?? {};
    $("agent").textContent = `${info.title ?? info.name ?? "agent"} ${info.version ?? ""}`.trim();
    setStatus(`connected · ${info.name ?? ""}`, true);
    logLine(`initialize: protocol ${initialized.protocolVersion}, loadSession=${initialized.agentCapabilities?.loadSession}`);

    try {
      await ctx.request(acp.methods.agent.authenticate, { methodId: "local" });
      logLine("authenticated (local)");
    } catch (e) {
      logLine(`authenticate skipped/failed: ${e?.message ?? e}`);
    }

    await newSession();
    await new Promise((resolve) => ctx.signal.addEventListener("abort", resolve));
    state.ctx = null;
    state.session = null;
    setStatus("disconnected", false);
  }).catch((e) => {
    state.ctx = null;
    setStatus("connection failed", false);
    alert(`ACP connection failed: ${e?.message ?? e}`);
  });
}

async function newSession() {
  if (!state.ctx) {
    alert("Connect first.");
    return;
  }
  if (state.session) {
    try {
      state.session.dispose();
    } catch { /* ignore */ }
    state.tools.clear();
  }
  endStreamingTurn();
  const cwd = $("cwd").value.trim();
  if (!cwd.startsWith("/")) {
    alert("cwd must be an absolute path.");
    return;
  }
  state.session = await state.ctx.buildSession({ cwd, mcpServers: [] }).start();
  state.sessionId = state.session.sessionId;
  $("sess").textContent = state.sessionId;
  logLine(`session/new → ${state.sessionId}`);
  const modes = state.session.modes;
  renderModes(modes);
  if (!state.pumping) {
    state.pumping = true;
    pump();
  }
}

function renderModes(modes) {
  const sel = $("modeSel");
  sel.innerHTML = "";
  const list = modes?.availableModes ?? [];
  if (!list.length) {
    $("modeLabel").textContent = modes?.currentModeId ? `mode: ${modes.currentModeId}` : "mode: n/a";
    sel.style.display = "none";
    return;
  }
  sel.style.display = "";
  for (const m of list) {
    const o = document.createElement("option");
    o.value = m.id;
    o.textContent = m.name ?? m.id;
    o.title = m.description ?? "";
    sel.appendChild(o);
  }
  if (modes.currentModeId) sel.value = modes.currentModeId;
  $("modeLabel").textContent = `mode: ${sel.value}`;
}

async function pump() {
  while (state.session) {
    let msg;
    try {
      msg = await state.session.nextUpdate();
    } catch {
      break;
    }
    if (msg.kind === "stop") {
      endStreamingTurn();
      logLine(`turn stopped: ${msg.stopReason}`);
      $("send").disabled = false;
    } else {
      renderUpdate(msg.update);
    }
  }
  state.pumping = false;
}

async function sendPrompt() {
  if (!state.session) {
    alert("Create a session first.");
    return;
  }
  const text = $("prompt").value.trim();
  if (!text) return;
  $("prompt").value = "";
  addMsg("user", `<div class="meta">you</div><div class="body">${marked.parse(text)}</div>`);
  $("send").disabled = true;
  try {
    await state.session.prompt(text);
  } catch (e) {
    endStreamingTurn();
    addMsg("sys", `<div class="meta">error</div><div class="body">${esc(e?.message ?? e)}</div>`);
    $("send").disabled = false;
  }
}

$("connect").onclick = () => connect().catch((e) => alert(e.message));
$("newSession").onclick = () => newSession().catch((e) => alert(e.message));
$("send").onclick = () => sendPrompt().catch((e) => alert(e.message));
$("prompt").addEventListener("keydown", (e) => {
  if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) sendPrompt();
});
$("cancel").onclick = async () => {
  if (!state.ctx || !state.sessionId) return;
  try {
    await state.ctx.request(acp.methods.agent.session.cancel, { sessionId: state.sessionId });
    logLine("cancel requested");
  } catch (e) {
    alert(e.message);
  }
};
$("modeSel").onchange = async (e) => {
  if (!state.ctx || !state.sessionId) return;
  await state.ctx.request(acp.methods.agent.session.setMode, {
    sessionId: state.sessionId,
    modeId: e.target.value,
  });
  $("modeLabel").textContent = `mode: ${e.target.value}`;
};

fetch("/config.json")
  .then((r) => r.json())
  .then((c) => {
    if (c.cwd) $("cwd").value = c.cwd;
  })
  .catch(() => {});
