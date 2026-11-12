# Cloud proof: Star -> API -> harness runner -> back

Goal: prove a cloud agent loop where Star creates a session on the API,
the harness runner executes in the cloud, and Star watches the stream
like a TV screen.

## What exists now

* `internal/cloud/client.go` — API client. User calls use
  `Authorization: Bearer`; worker calls use `X-Worker-Token` (same
  headers as the harness). `CreateTerminalSession` matches prod's
  `POST /deca/v1/terminal/session`; `WatchTerminal` streams a relay
  browser URL, stripping the harness PTY frame byte.
* `internal/cmd/cloud.go` — `star cloud status|job|session|attach`,
  registered on the root command. Credentials only via env:
  `DECA_AGENT_API_BASE_URL`, `DECA_API_KEY`,
  `DECA_AGENT_WORKER_TOKEN`, `DECA_AGENT_JOB_ID`.
* `mock_agent_jobs.py` — mock of the harness `runner.py` jobs contract
  (`/deca-agents/v1/jobs/...`). Used to verify Star + `runner.py`
  helpers speak the same protocol.
* `mock_terminal_relay.py` — mock of prod's
  `terminal_relay/flask_relay.py` + session create: pairs
  `/terminal/runner/<id>` with `/terminal/browser/<id>` and pumps
  bytes both ways, like prod.
* `probe_agent_jobs.py` — env-driven probe for mock or prod.
* `../harness/` — read-only reference copy of public
  `GenLabsAI/opencode-action-harness` (not committed).

## ACP leg (passes): Star cloud session carrying model config

`star cloud acp --prompt ... [--model ...] [--base-url ...]` reads the
inference key from `NVIDIA_API_KEY` (env only), creates the session
(sessionid cookie or Bearer, like prod), sends an `acp_init` config
frame `{prompt, model, base_url, api_key}` over the relay, then runs
`initialize` / `session/new` / `session/prompt` and prints the
streamed `agent_message_chunk` text. The reference runner
(`acp_bridge.py`, the GH-actions side) reads the config, calls the
model, and answers with real ACP frames. Verified live against the
mock relay with NVIDIA `mistralai/mistral-large-4-1t-a52b`: prompt in,
`CLOUD-ACP-OK` out. Full exchange in `acp-transcript.txt` (key
redacted; repo swept clean). Deca provider/config untouched — the
default model flags and all pre-existing `star cloud` commands behave
as before.

## Local end-to-end (passes)

Uses the REAL star binary and the REAL harness `terminal_runner.py`:

```sh
python3 mock_terminal_relay.py 8939 &
export DECA_AGENT_API_BASE_URL=http://127.0.0.1:8939 DECA_API_KEY=user-secret
OUT=$(star cloud session)   # -> session_id + ws://.../terminal/browser/<id>
star cloud attach <relay-url> &   # the TV screen
python3 ../harness/terminal_runner.py \
  --relay-url ws://127.0.0.1:8939/terminal/runner/<id> \
  --command "printf 'CLOUD-AGENT-READY\n'; echo hello-from-cloud" \
  --session-token t --api-base-url http://127.0.0.1:8939 \
  --deca-api-key user-secret --session-id <id>
```

Viewer output:

```
CLOUD-AGENT-READY
hello-from-cloud

[agent exited: code 0]
```

`go vet ./internal/cloud/ ./internal/cmd/` and `gofmt` clean.
`go.mod` promotes `github.com/coder/websocket` (already a dependency)
to direct use.

## Prod status (api.genlabs.dev, verified live)

* Reachable; `/deca/v1/models` 200; service `genlabsapi` HEALTHY.
* `POST /deca/v1/terminal/session` route EXISTS (returns 401 JSON, not
  404) and dispatches `terminal-session.yml` in
  `GenLabsAI/opencode-action-harness` on success — this is the live
  Star->API->opencode path.
* Relay WS routes `/terminal/runner/<id>` + `/terminal/browser/<id>`
  are registered in the deployed app.
* `/deca-agents/v1/jobs/*` does NOT exist in prod (clean 404s).
* Live relay socket proof (2026-11-12): a real WS handshake to
  `wss://api.genlabs.dev/terminal/runner/<unknown-id>` completes the
  HTTP upgrade and the server closes with `1008 Session not found` —
  exactly the `handle_runner_ws` path in the deployed app. Sockets
  live, pairing enforced.
* Live dispatch proof: `terminal-session.yml` on main in
  `GenLabsAI/opencode-action-harness` shows prior `workflow_dispatch`
  runs, including a `success` run (2026-09-10, 2m6s) — the only
  dispatcher of that workflow is the API's `_dispatch_terminal_workflow`.
  Genlabsapi->GitHub-Actions->runner has executed in production.
* Auth correction: the supplied value is a web **sessionid**, not an
  API key — it authenticates via the `sessionid` cookie (supported by
  the Star client through `DECA_SESSION_ID`), which is how the live
  sessions below were created. `Authorization: Bearer` with it 401s.
* LIVE proof (2026-11-12 ~15:02 UTC, session `2d9146d7`): `star cloud
  session` (cookie auth) returned a real session; genlabsapi dispatched
  `terminal-session.yml` (GH run `43243464888`, success); the cloud
  runner installed crush, wrote its config, and connected to the prod
  relay (`terminal_runner: connected to relay` in the run logs); the
  watcher captured crush's real PTY bytes back (`live-transcript.txt`:
  crush v0.100.0 banner + provider picker). Full loop done: Star ->
  API -> GitHub Actions -> crush -> relay -> watcher. All three live
  sessions deleted afterwards (HTTP 200) to stop spend.
* Known prod breakage found live: crush boots to first-run setup
  ("choose a provider and model") because `terminal-session.yml`
  writes config to `~/.config/crush/crush.json` while crush reads
  `~/.crush/crush.json` (the path `runner.py` uses). Fix belongs in
  the harness workflow, which this token cannot push to.
* Bonus fix found by probing: prod 400s bodyless GETs carrying
  `Content-Type`; the Go client only sends it with a body now.
