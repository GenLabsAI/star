# Live end-to-end evidence (2026-11-12, api.genlabs.dev on Koyeb)

Every line below is copied from tool output this session. Secrets are
env-only and were never printed; sweeps for `deca_oc_`, `nvapi-`, and the
session cookie value returned 0 hits in every captured log.

## 1. GitHub credential: identity, scopes, and the exact write denial

`GET /user` and the permission headers:

```
{"id":96586789,"login":"ccocks","type":"User"}
HTTP/2 200
x-accepted-github-permissions: allows_permissionless_access=true
```
No `X-OAuth-Scopes` header at all -> this is a **fine-grained PAT**, not a
classic token, so scope information lives in the token's repository
grant, not a scope string.

Repo discovery from the key's own listings:

```
GET /user/repos?per_page=100&affiliation=owner,collaborator,organization_member -> 55 repos
GET /orgs/GenLabsAI/repos?per_page=100 -> 13 repos:
  GenLabsAI/Agentica, Deca, DecaEditor, GenLabs-Python, Genlabs-SDK-Python,
  OpenDeca-2-mini, OpenLogitsProcessor, ai-cli, deca-builder, docs,
  opencode, opencode-action-harness, star
GET /user/orgs -> (empty: no org membership rows for this token)
```

`GenlabsAPI` (the repo Koyeb actually builds from, per the deployment's
`definition.git.repository`) is **not in either listing**, and does not
resolve in any casing:

```
GET /repos/GenLabsAI/GenlabsAPI     -> 404
GET /repos/GenLabsAI/genlabsapi     -> 404
GET /repos/GenLabsAI/Genlabsapi     -> 404
GET /repos/GenLabsAI/GenLabsApi     -> 404
GET /search/code?q="deca-agents"+org:GenLabsAI -> total_count 0
git ls-remote .../GenlabsAPI.git    -> "Write access to repository not granted"
```

Write test on the one org repo the key *can* see, with headers captured:

```
POST /repos/GenLabsAI/opencode-action-harness/git/refs
HTTP/2 403
x-accepted-github-permissions: contents=write; contents=write,workflows=write
{"message":"Resource not accessible by personal access token","status":"403"}
```

So: `permissions.push: true` in the repo object describes the *user's*
role (admin), not the token's grant. GitHub's own
`X-Accepted-GitHub-Permissions` says the required permission is
`contents=write`, and the token does not have it. Forking is also denied:

```
gh repo fork GenLabsAI/opencode-action-harness -> HTTP 403: Resource not accessible by personal access token
```

Conclusion: no repo write is possible with this credential anywhere, and
`GenlabsAPI` is not even readable. The ACP workflow + bridge are prepared
in this repo (`acp-session.yml`, `acp_bridge.py`) and apply cleanly to a
harness checkout; only the push is blocked. Required fix: regenerate the
PAT with **Contents: Read and write** on `GenLabsAI/GenlabsAPI` and
`GenLabsAI/opencode-action-harness`.

## 2. Live chain A: Star -> Koyeb API -> GitHub Actions -> agent -> back to Star

```
Star binary: cloud session  (sessionid cookie auth)
  22:29:19  Incoming request to /deca/v1/terminal/session
  22:29:25  "POST /deca/v1/terminal/session HTTP/1.1" 200 144  "star-cloud-probe"
  session_id 6fe8e926-2598-430e-9141-6e1e30ba00d5
  relay      wss://api.genlabs.dev/terminal/browser/6fe8e926-...
```

genlabsapi dispatched the harness workflow:

```
GH run 43320352920  Deca Terminal Session  workflow_dispatch  success  1m0s  22:29:21Z
  Run PTY bridge | python terminal_runner.py --relay-url wss://api.genlabs.dev/terminal/runner/6fe8e926-...
  Run PTY bridge | INFO:__main__:terminal_runner: connected to relay
```

The runner then pushed scrollback back through the same API (proof the
cloud side is talking to genlabsapi, not just the socket):

```
22:30:14  Incoming request to /deca/v1/terminal/session/6fe8e926-.../scrollback
22:30:14  "POST /deca/v1/terminal/session/6fe8e926-.../scrollback HTTP/1.1" 200 16
          "Python/3.11 aiohttp/3.14.5"
```

Star's viewer side received the live agent UI, and Star's input reached
the cloud PTY (2x DOWN + ENTER). Rendered capture in
`live-crush-render.txt`:

```
  ╱╱╱╱╱╱  Charm™v0.101.0 ╱╱╱╱╱╱╱╱╱
 To start, let's choose a provider and model.
 > Find your fave
 Charm Hyper  ─────...
 Prism | Prism Core | Prism Flash | DeepSeek V4 Pro 0813 | Kimi K3 ...
 ↑/↓ choose • enter confirm
```

That is the deployed harness's agent (crush) running on the GitHub
runner, streamed to Star and reacting to Star's keystrokes.

## 3. Live chain B: Star -> Koyeb API -> model -> reply returned to Star

Run by the Star binary itself (`star cloud infer`), no curl:

```
enrolled device key (51 chars, not printed)
STAR-KOYEB-ROUNDTRIP-OK
```

Koyeb service log for the same call:

```
22:35:03  Incoming request to /deca/v1/opencode/enroll
22:35:04  "POST /deca/v1/opencode/enroll HTTP/1.1" 200 66  "star-cloud-probe"
```

Enrollment is the documented device-signal flow the opencode client uses;
it mints a `deca_oc_` key that the Star client keeps in memory, sends as
`Authorization: Bearer` to `POST /deca/v1/chat/completions`, and discards.
The completion returned `STAR-KOYEB-ROUNDTRIP-OK` from `deca-2.5-mini`.

## 4. Where the two chains do not yet join (single blocker)

Chain A's agent (crush) boots to first-run provider setup instead of
using the Deca provider, because `terminal-session.yml` writes
`~/.config/crush/crush.json` while crush reads `~/.crush/crush.json`
(the path `runner.py` uses). With the correct path, chain A's agent would
reach the same gateway chain B proves, and a single run would carry a
prompt to the model and the answer back through the relay. The fix is one
line in the harness workflow — blocked only by the missing
`contents=write` grant documented above.

## 5. Koyeb CLI

Installed: `~/.koyeb/bin/koyeb`, v5.12.0 (`koyeb --help` runs).
`koyeb login` is **interactive token-paste**, not a device-code flow, so
it cannot complete unattended. Its prompt, printed in the user's terminal:

```
✗ Enter your personal access token. You can create a new token here
  (https://app.koyeb.com/user/settings/api):
```

That URL is the only user action needed. Meanwhile the CLI is already
authenticated for every call via `KOYEB_TOKEN`:
`koyeb services get personal-jenda/genlabsapi` -> `HEALTHY`.
