# Cloud proof: Star -> Agent Jobs API -> harness

Goal: prove Star TUI talks to the Agent Jobs API in `genlabsapi` and
`genlabsapi` talks to the harness (`opencode-action-harness`) runner,
using the exact HTTP contract the GitHub Actions worker uses.

## What exists now

* `internal/cloud/client.go` — Go client for the Agent Jobs API.
  User calls use `Authorization: Bearer`, worker calls use
  `X-Worker-Token`, matching `../harness/runner.py`.
* `internal/cmd/cloud.go` — `star cloud status` (user-auth reachability
  check) and `star cloud job [id]` (worker-auth job fetch). Registered
  on the root command. Credentials only via env:
  `DECA_AGENT_API_BASE_URL`, `DECA_API_KEY`,
  `DECA_AGENT_WORKER_TOKEN`, `DECA_AGENT_JOB_ID`.
* `mock_agent_jobs.py` — local mock of the API implementing the
  `runner.py` contract (`POST .../jobs/<id>/events`,
  `PATCH .../jobs/<id>`, `GET .../jobs`, `GET .../jobs/<id>`).
* `probe_agent_jobs.py` — user-side + worker-side probe. Works against
  the mock or prod, driven entirely by env (no secrets in files).
* `../harness/` — read-only reference copy of
  `GenLabsAI/opencode-action-harness` (public) for contract checks.

## Local proof (passes)

```sh
cd artifacts/cloud-proof
python3 mock_agent_jobs.py 8931 &
DECA_AGENT_API_BASE_URL=http://127.0.0.1:8931 DECA_API_KEY=user-secret \
  DECA_AGENT_WORKER_TOKEN=worker-secret python3 probe_agent_jobs.py
```

Also verified: the real `runner.py` helpers (`post_json`,
`flush_events`, `patch_status`) against the mock, and
`go vet ./internal/cloud/ ./internal/cmd/` clean.

## Prod proof (needs you)

In a terminal with this repo checked out:

```sh
export PATH="$HOME/.koyeb/bin:$PATH"
koyeb login            # paste a token from https://app.koyeb.com/user/settings/api
koyeb apps list        # find the genlabsapi app -> prod API base URL

export DECA_AGENT_API_BASE_URL="https://<genlabsapi-host>"
export DECA_API_KEY="<your deca key>"            # never commit this
export DECA_AGENT_WORKER_TOKEN="<worker token>"  # only for worker-side check
export DECA_AGENT_JOB_ID="<job id>"              # only for worker-side check
python3 artifacts/cloud-proof/probe_agent_jobs.py
```

Still needed: `gh auth login` (to clone private `genlabsapi` and find
the real `/deca-agents/v1/jobs` handlers), then replace the guessed
user-side paths in `client.go:Pong` with the real ones and re-run the
probe against prod.
