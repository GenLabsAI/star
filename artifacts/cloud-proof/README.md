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
* The supplied Deca key 401s on every authed endpoint tested
  (`/terminal/session`, `/profile`, `/usage`) — it is in neither the
  `TrustKeys` nor `OpencodeApiKeys` tables, so no live session could
  be created with it. A working user key (`deca_oc_...` or TrustKey)
  is the one remaining input for the live run:
  `star cloud session` -> real GH Actions runner -> `star cloud attach`.
* Bonus fix found by probing: prod 400s bodyless GETs carrying
  `Content-Type`; the Go client only sends it with a body now.
