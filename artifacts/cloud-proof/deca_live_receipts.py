"""Turn one live cloud-turn transcript into a receipts report.

Reads the Star message database the driven TUI wrote, plus the facts this
run produced (session id, workflow run, commit), and prints the markdown
that `deca-live-hero.txt` holds.

Usage:
  python3 deca_live_receipts.py <crush.db> <facts.json>
"""

import json
import sqlite3
import sys
from datetime import datetime, timezone


def utc(ts):
    return datetime.fromtimestamp(ts, tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def main():
    db_path, facts_path = sys.argv[1], sys.argv[2]
    facts = json.load(open(facts_path))

    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    rows = list(conn.execute("select role, created_at, parts from messages order by created_at"))
    start = rows[0][1]
    end = None
    tools = 0
    first_tool = None
    steps = []
    for role, ts, parts in rows:
        kinds = [p["type"] for p in json.loads(parts)]
        if role == "tool":
            tools += 1
            if first_tool is None:
                first_tool = ts
        for p in json.loads(parts):
            if p["type"] == "finish" and p["data"].get("reason"):
                end = ts
        steps.append((ts - start, role, ",".join(kinds)))

    print(f"""# Live cloud run: UI edit + screenshot, end to end

One prompt typed into the normal Star composer with `/cloud` on. No
session ids on the command line, no harness checkout, no manual
dispatch: submit, watch it work, read the answer.

Everything below is copied from this run's own output.

## Headline

| | |
|---|---|
| Prompt submitted | {utc(start)} |
| First cloud tool call | +{first_tool - start}s |
| Turn finished (`end_turn`) | {utc(end)} |
| Submit -> done | {(end - start) // 60}m{(end - start) % 60:02d}s |
| Cloud tool calls | {tools} |
| Cloud agent session | `{facts['session_id']}` |
| GitHub Actions run | [{facts['run_id']}](https://github.com/GenLabsAI/opencode-action-harness/actions/runs/{facts['run_id']}) (`{facts['run_name']}`), job `{facts['job_id']}` |
| Runner | ubuntu-latest, `/tmp/agent-workspace` (empty, no git repo) |
| Model | `deca-2.6-pro` through the Deca gateway, key enrolled at run time |
| Result | branch `{facts['branch']}` in `GenLabsAI/star`, commit [`{facts['commit'][:7]}`](https://github.com/GenLabsAI/star/commit/{facts['commit']}) |
| Screenshot | `artifacts/cloud-proof/hero/hero.png` (1440x900, pulled from that branch) |

## What the agent did

The prompt asked for a premium dark hero page for "Star" with a slim
chat bar, a headless render, a commit and a push. Its own tool log, in
order:

1. `pwd; ls -a; git remote -v` -> empty `/tmp/agent-workspace`, no repo.
2. `which chromium node python3` -> chromium present, so it can render.
3. `git clone https://github.com/GenLabsAI/star` (the git token
   travelled in the session config over the relay; it never appears in
   a command or a log).
4. Read `README.md`, `demo/`, `launcher/src/splash.rs` to pull the real
   brand palette instead of inventing one.
5. Write `cloud-demo/hero.html`.
6. `chromium --headless ... --window-size=1440,900 --screenshot`.
7. Decode the PNG in Python to verify the render, not just the exit
   code: bar at x 418-1021 / y 570-626, 1px translucent border,
   placeholder glyph pixels present.
8. Repo-local git identity, branch `cloud-demo/hero`, commit, push,
   then `git ls-remote` to confirm the remote SHA matches.

Agent's closing line:

> A near-black canvas lit by a warm gold radial halo and a faint
> starfield, with an 80px gradient wordmark headline ("The terminal,
> with a mind of its own.") and a slim 604px glass chat bar - 1px
> translucent white border, tight 9/18px padding, gold soft glow, and
> the placeholder "Ask Star anything...".

## Per-message timeline (seconds after submit)

```
""")
    for delta, role, kinds in steps:
        print(f"{delta:+4d}s {role:9s} {kinds}")
    print("""```

## TUI frame at the end

```
  ✓ Ran 1 command (double-click to expand)

  ## Done
   Where I worked: /tmp/agent-workspace was empty and not a git repo,
   so I cloned github.com/GenLabsAI/star into it.

    Item          │ Value
   ───────────────┼──────────────────────────────────────────────
    Commit SHA    │ b3798aa55e352dbeb443dfbe50f932a29de0f4dc
    Branch        │ cloud-demo/hero (pushed, tracking origin)
    Screenshot    │ cloud-demo/hero.png (1440x900 PNG, 140 KB)

  ◇ Deca 2.6 Pro via Deca in 2m16s
  CLOUD   tab focus chat • shift+tab mode • / or ctrl+p commands
```

The `CLOUD` badge in the status bar is the only visual difference from a
local turn; the footer counts the turn like any other.

## Credentials this needed

| Variable | Value in this run |
|---|---|
| `DECA_AGENT_API_BASE_URL` | `https://api.genlabs.dev` |
| `DECA_SESSION_ID` | the web session id supplied in chat |
| `DECA_GIT_TOKEN` | the existing GitHub token, so the agent could push |

The model key was not configured anywhere: Star enrolls a short-lived
`deca_oc_` key from the session id and sends it to the runner inside the
session config frame, so the runner calls the same gateway the user's own
chat does.

## Friction found while doing this

* Toggling `/cloud` opens the command palette and the first automated
  keystrokes landed in its filter; a person just presses Enter and
  types, so this is a driver problem, not a product one. Fixed in the
  driver by dismissing the palette with Esc.
* The screenshot could not come back **inside** the conversation: the
  cloud agent's ACP frames carry text and tool calls only, so
  `cloudTurn.apply` drops image content blocks. The image above was
  recovered from the branch the agent pushed. Getting the agent to hand
  the picture straight into the chat is the next real step.
* The driver's stop heuristic waits for a quiet screen, so it reported
  5m20s of wall time for a 2m16s turn. The turn time above comes from
  the transcript and from the TUI footer, not the driver clock.

## Files

* `artifacts/cloud-proof/deca_live_run.py` - the driver that typed the
  prompt into the real TUI (PTY) and captured frames.
* `artifacts/cloud-proof/live-run/run3.txt` - captured screens and the
  live timing lines.
* `artifacts/cloud-proof/hero/hero.png`, `hero.html` - the agent's
  deliverables, as pushed.
* `artifacts/cloud-proof/deca_model_probe.py` - proves the Deca gateway
  answers tool calls for `deca-2.6-pro` / `deca-2.5-pro`.
""")


if __name__ == "__main__":
    main()
