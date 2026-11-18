# Orbs — the requirements list

An orb is the machine an agent works on: a cloud environment per chat
thread, with its own filesystem, terminal, processes, dependencies and
network. This is the living requirements list for that platform, as we
have stated it. Every item records what already exists, what is missing,
and whether it is waiting on something only the user can provide.

Status legend: DONE = delivered end to end (on `main`, proven). WIP =
code exists but not closed out. OPEN = not started. SKIPPED = explicitly
deferred. BLOCKED = cannot progress without the user (listed below).

## Blocked on you

These are the only things standing between the WIP items below and real
proof. Everything else is ours to finish.

1. **Cloud credentials for live proof.** The cloud paths (`star cloud
   session|attach|acp`, `/cloud` in the TUI, cloud wakeups) authenticate
   from the environment only. Set, in Settings → Secrets:
   `DECA_AGENT_API_BASE_URL`, `DECA_API_KEY` (or `DECA_SESSION_ID`),
   `DECA_AGENT_WORKER_TOKEN`, and for cloud inference `DECA_MODEL`,
   `DECA_MODEL_BASE_URL`, `DECA_MODEL_API_KEY`. Without them we can only
   prove cloud work against the local mocks in `artifacts/cloud-proof/`.
2. **Write access to `GenLabsAI/opencode-action-harness`.** The supplied
   token authenticates as `ccocks` and reports `permissions.push = true`,
   but every write is denied server-side (`403`); `GenLabsAI/GenlabsAPI`
   is not visible to it at all. Most remaining platform work (items 4, 7,
   8, 9, 22) lands in the harness image, the workflow, or that API — for
   item 7 specifically, the port→preview-URL mapping and crash
   supervision.
3. **A decision on item 8 (live previews).** Previously deferred by you;
   it is in this list again, so we need to know whether it is back in
   scope and, if so, what the preview networking should look like
   (authenticated vs shareable, one port vs many).
4. **Prod access policy for item 19.** Whether agents get read-only prod
   logs and monitoring, or may also deploy/rollback, and which changes
   require a human approval gate.
5. **Harness-only bug fix.** Prod's `terminal-session.yml` writes config
   to `~/.config/crush/crush.json` while crush reads `~/.crush/crush.json`,
   so a fresh orb boots to first-run setup. The fix is a one-line change
   in the harness repo, which item 2 currently blocks.

## The list

1. **Fully isolated cloud environments** created automatically per task,
   each with its own filesystem, terminal, processes, dependencies and
   resources; agents never execute code or consume resources on the
   user's machine. — **DONE**. Containerized harness orb, cloud runs
   proven end to end; `internal/cloud/`, `artifacts/cloud-proof/`.
2. **A dedicated environment per conversation**, a 1:1 chat↔workspace
   mapping that keeps files, terminal state, processes, installed
   dependencies and history for when the user returns. — **DONE**.
   Chat-affine cloud sessions (`CreateACPSession` keyed by Star session),
   `/cloud` is per session; `internal/cloud/client.go`,
   `internal/workspace/app_workspace_cloud.go`.
3. **Persistent, resumable environments**: suspend when idle, resume
   without losing work, restoring services, connections and processes. —
   **DONE**. Idle reuse, stash restore, snapshot cache, dead-runner
   self-heal, plus `.agents/resume` on every wake-up.
4. **Automatic provisioning** with no manual intervention: clone repos,
   detect project requirements, install dependencies, configure runtimes,
   prepare databases, run repository setup scripts. — **WIP**. The
   `.agents/setup` lifecycle hook and its content hash (used for snapshot
   reuse) are forwarded per turn; detection, readiness and documented
   commands are now reported locally by `star env` (see below). Missing:
   importing a repo automatically and applying that plan inside the orb,
   which is harness-side.
5. **Reusable environment snapshots** so new workspaces start warm without
   repeating expensive setup, while staying isolated per workspace. —
   **DONE**. Setup hash → snapshot cache in the harness.
6. **Complete development capabilities**: unrestricted shell, package
   installs, scripting, compilation, debugging, databases, dev servers,
   browsers, arbitrary tooling. — **DONE**. Harness `Dockerfile.agent`
   (Debian, git, `gh`, tmux, ffmpeg, ImageMagick, ripgrep, Node/Bun/pnpm,
   headless browsers) plus Star's local toolchain.
7. **Intelligent environment management**: ports, processes, server
   startup, crashes, dependencies, resource allocation, recovery, without
   the agent debugging infrastructure. — **WIP → local half done**. New
   `star ports [--json] [--all]` (`internal/ports`) answers "what is
   listening here, whose process is it, and where does it live": kernel
   socket tables (`/proc/net/tcp{,6}` + inode→pid + cwd) on Linux, `ss`
   and `lsof` fallbacks, `netstat` on Windows; by default only listeners
   whose process runs inside the workspace. Proof:
   `artifacts/orbs-proof/07-ports.txt`. Existing: background jobs and
   streaming via `internal/shell` (`job_output`, `job_kill`, `monitor`),
   runner self-heal. Missing: mapping a port to an orb-side preview URL and
   crash supervision with a restart policy; both need write access to the
   harness repository (blocker 2).
8. **Live application previews** exposed as URLs, interactive in the
   browser, shareable, several at once without access to the environment.
   — **SKIPPED** (explicitly deferred by you; see blocker 3).
9. **Full browser automation**: navigate, click, type, submit forms, upload
   files, test authenticated flows, inspect output, capture screenshots and
   record demos. — **WIP**. Headless browsers exist in the orb image and
   screenshot/video artifacts are produced today, but Star itself has no
   browser tool (`internal/agent/tools/` has no navigation/click/upload
   tool). Needs the harness bridge.
10. **First-class agent collaboration**: spawn agents in independent
    environments, delegate, exchange messages and files, coordinate
    parallel work, hand off with context. — **WIP**. Local sub-agents,
    the `team` tool (`internal/teams`) and ACP work; spawning a *cloud*
    orb from an agent is OPEN.
11. **Multiple independent implementations** explored in parallel, each
    previewable and comparable until one is chosen. — **OPEN**. Only
    session forks exist in the TUI (`/fork`, `internal/ui/model/fork.go`);
    there is no parallel-orb comparison or selection flow.
12. **Background and scheduled execution** for extended work without the
    user connected: future work, self-wake, event response, recurring
    tasks. — **DONE**. Durable wakeup store + cron engine + `star wakeup
    schedule|trigger|run|list` + `schedule_wakeup` tool + TUI toast, and
    `cloud: true` runs the wake-up on the orb. Release-binary verification
    of the CLI is still outstanding.
13. **Remote control and session management** from anywhere: create tasks,
    send instructions, inspect progress, open terminals, review files;
    closing the browser or shutting the laptop changes nothing. — **DONE**.
    `star server` (HTTP/TCP/socket) with client mode, cloud turns stream
    into the local store and continue server-side.
14. **Persistent project knowledge**: repository instructions, docs,
    architecture notes, skills, scripts and earlier discoveries available
    across sessions. — **DONE**. Context files (AGENTS.md/CLAUDE.md/CRUSH.md
    and `.local` variants), builtin and project skills, `search_history`,
    MCP resources, and `star env` lists which of these a project has. New
    `star knowledge add|list|show|remove|path` stores discoveries as plain
    markdown files in the project data directory (`.crush/knowledge/`),
    body returned byte-for-byte, names restricted to slugs so a path can
    never escape the directory, size-capped, listed with summaries, and
    surfaced in the `star env` manifest. Because entries are files, they
    are shared by every session on the machine and can be reviewed,
    grepped, edited or committed. Proof:
    `artifacts/orbs-proof/06-knowledge.txt` (write in one process, read
    back verbatim with matching sha256 in another).
15. **A complete agent-oriented development interface**: predictable
    filesystem locations, machine-readable environment information,
    documented commands, readiness checks, accessible logs, clear
    diagnostics. — **WIP → advanced here**. `star env [--json]` now emits
    the manifest (paths, git, detected stacks with install/build/test
    commands, lifecycle hooks and their hashes, runtime readiness,
    credential *presence* only), alongside the existing `star dirs`,
    `star logs`, `crush_info`, `crush_logs`, and now a port inventory
    (`star ports`, see item 7).
16. **Autonomous testing and verification** against the real running
    application: reproduce bugs, investigate failures, run regressions,
    iterate until it works. — **WIP → advanced here**. `star verify` runs
    the commands `star env` detected (build and test by default, install
    with `--install`), bounds each with `--timeout`, records exit code,
    duration and output, and exits non-zero when anything fails. Evidence
    in `artifacts/orbs-proof/`. `--repeat N` runs every command N times and
    reports a command that both passed and failed as **FLAKY** (non-zero
    exit, every attempt logged, the failing attempt's output kept) instead
    of letting a lucky retry pass for a stable result; repeated runs set
    `GOFLAGS=-count=1` so Go's test cache cannot replay a cached pass.
    Proof: `artifacts/orbs-proof/11-flaky-verification.txt`. `--min-coverage
    PERCENT` is the coverage gate: the run measures statement coverage
    (`go test -coverprofile` today), exits non-zero with a distinct
    **UNDER-COVERED** verdict when a passing run measures below the
    threshold, reports the measured number rather than the threshold, and
    prints **COVERAGE-UNAVAILABLE** with a reason (exit non-zero) when no
    number can be measured — a project with no coverage provider, or a
    toolchain that cannot produce a profile — instead of passing silently.
    A failing test stays FAIL and a flaky repeated run stays FLAKY: the gate
    only downgrades a run that would otherwise pass. Proof:
    `artifacts/orbs-proof/13-coverage-gate.txt`. Missing: driving a running
    application rather than its test suite.
17. **Evidence-based task completion**: test results, screenshots, videos,
    logs and a summary the user can check instead of trusting a claim. —
    **DONE**. Screenshot/video/gif + e2e log convention and the
    `artifacts/` tree; cloud transcripts via `star cloud attach`. `star
    verify` now emits the bundle in the shape this needs: `verify.json`,
    `summary.md`, per-command logs and `changes.patch` in one directory.
18. **External integrations and event triggers**: tasks originating from
    GitHub, Slack, Linear, webhooks; agents calling external APIs. —
    **WIP**. MCP servers, plugins, hooks, and event wakeups triggered by
    CI (`repository_dispatch` type `wakeup-event`); `gh` is available in
    the orb. First-class GitHub/Slack/Linear connectors are missing, and
    the 15-minute GitHub sweep lives harness-side.
19. **Production and infrastructure access**: logs, monitoring,
    deployments, health, incident investigation, prepared fixes, with
    permissions and human approval for sensitive steps. — **OPEN** (see
    blocker 4).
20. **Local machine connectivity** when a task needs local hardware, OS
    specifics or local resources. — **DONE** (authorization half). Star's
    client/server split lets a remote client drive a workspace on the
    machine that runs `star server`; that is now an explicit opt-in, not a
    default: `star pair` mints a 256-bit token (printed once, never
    stored), the server accepts it via `--pair-token`/`--pair-file` or
    `STAR_SERVER_TOKEN`/`STAR_SERVER_TOKEN_FILE`, and a middleware rejects
    every unpaired peer with 401 while the bundled client presents the
    token automatically. A server started without a token keeps its old
    local-only behaviour. Proof: `artifacts/orbs-proof/08-pairing.txt`
    (401 unpaired and wrong-token, 200 paired, token never echoed).
    Missing: per-workspace scopes and an approval prompt for sensitive
    operations — one token grants everything that server exposes.
21. **Complete change management**: track every modification, produce clean
    diffs, commits, branches and pull requests, isolated until review. —
    **WIP → advanced here**. Per-session git worktrees
    (`internal/worktree`, `enter_worktree`/`exit_worktree`), the diff tool,
    `internal/gitutil` branch detection, and now `star changes`: branch and
    upstream divergence, every touched path, diffstat and patch size, JSON
    for agents, and `--patch FILE` to hand a reviewer a real diff. `star
    verify` bundles carry the same patch, so a diff never travels without
    the run that produced it. `star commit [-m MESSAGE] [--branch NAME]
    [--include-untracked] [--json]` is the other half: it stages tracked
    changes, refuses to touch untracked files until they are approved, needs
    an explicit message, optionally creates the branch it commits on, and
    reports the resulting hash; it never pushes. An explicit approval gate
    comes from the project's own config: a `"commit"` block in `crush.json`
    or `.crush.json` can set `require_approval` and/or
    `protected_branches`, and a commit that arrives without `--approve` is
    refused before anything is staged, so the tree, the index and
    `git rev-list --count HEAD` are unchanged; the same commit with
    `--approve` lands and reports the hash `git rev-parse HEAD` confirms.
    `star commit --policy` prints the effective policy. Proof:
    `artifacts/orbs-proof/10-commit.txt` and
    `artifacts/orbs-proof/14-approval-gate.txt`. Missing: pull-request
    creation (needs a remote and credentials).
22. **Resource and cost management**: configurable CPU/memory/storage,
    scaling with workload, minimal idle cost, usage and cost visibility. —
    **WIP**. Sleeping/reused orbs, `DeleteSession`, and explicit session
    cleanup keep spend down; there is no usage or cost reporting surface,
    and no per-orb resource configuration in Star.
23. **Environment and access security**: strong isolation, protected
    credentials and secrets, configurable permissions, no access to
    unauthorized resources. — **WIP**. Permission engine, hooks, the
    cross-origin guard on the local server, env-only credentials (never in
    files or logs), and `safe_config` guards. Missing: an explicit
    per-orb permission model to match.
24. **A unified workspace management experience**: one overview of
    environments, conversations, running agents, previews, scheduled
    tasks, completed work and artifacts, with inspect/share/pause/resume/
    terminate. — **WIP → advanced here**. `star overview [--json]` joins
    every tracked workspace with its branch, changed paths, latest
    verification verdict and cloud credential state, next to `star
    projects`, the sessions dialog, `star wakeup list` and the TUI status
    badges. Missing: sessions and schedules in the same view,
    previews/demos, and pause/resume/terminate wired to the cloud
    lifecycle.

## What changed from the previous list

The earlier 12-item list coalesced into this one: old 1 → 1, old 2 → 2,
old 3 + 6 → 3, old 4 → 4, old 5 → 5 (continuous validation is now part of
3/4), old 7 → 12, old 8 → 6, old 9 → 17, old 10 → 8, old 11 → 10, old 12
→ 13. Items 7, 9, 11, 14, 15, 16, 18, 19, 20, 21, 22, 23 and 24 are new.

## Done in this pass (proof in `artifacts/orbs-proof/`)

* **Item 15 + 4** — `star env [--json]`: the machine-readable environment
  manifest (paths, git, stacks, commands, hooks, readiness, credential
  presence). New `internal/envreport` package with tests.
* **Item 16** — `star verify [--json]`: builds and tests the detected
  stack, bounds every command, writes an evidence bundle, fails loudly.
  New `internal/verify` package with tests, including the timeout path.
* **Item 21 (review half)** — `star changes [--json] [--patch FILE]`:
  what changed, on which branch, and a diff a reviewer can read. New
  `internal/changes` package with tests.
* **Item 24 (local half)** — `star overview [--json]`: workspaces,
  branches, changed paths, latest verification and cloud credential state
  in one view. New `internal/overview` package with tests.

## Done two passes ago (items 14, 7, 20)

* **Item 14** — `star knowledge add|list|show|remove|path`: durable,
  file-backed project knowledge that survives sessions, with slug-only
  names, size limits, summaries and `star env` discovery. New
  `internal/knowledge` + `internal/datadir`. Proof:
  `artifacts/orbs-proof/06-knowledge.txt`, regenerated by
  `make-knowledge-proof.sh`.
* **Item 7 (local half)** — `star ports [--json] [--all]`: kernel-level
  listener inventory with pid, process, address and working directory,
  filtered to the workspace by default. New `internal/ports` with parsers
  for `/proc/net/tcp{,6}`, `ss`, `lsof` and `netstat`. Proof:
  `artifacts/orbs-proof/07-ports.txt`, regenerated by
  `make-ports-proof.sh`.
* **Item 20 (authorization half)** — `star pair` mints a bearer token; the
  server enforces it on every request and the client presents it. New
  `internal/pairing` + `internal/server/pairing.go`. Proof:
  `artifacts/orbs-proof/08-pairing.txt`, regenerated by
  `make-pairing-proof.sh`.
* All three commands were also re-run against the published `v0.1.141`
  artifact (commit `1022fd03a`): `artifacts/orbs-proof/09-released-commands.txt`.

## Done in the previous pass (items 21, 16)

* **Item 21 (commit half)** — `star commit`: real commits and branches from
  the reviewed working tree, gated on an explicit message and on approval
  of untracked paths, never pushing. New `internal/commit`. Proof:
  `artifacts/orbs-proof/10-commit.txt` (`make-commit-proof.sh`).
* **Item 16 (flakes)** — `star verify --repeat N`: repeated runs, a
  FLAKY verdict with a non-zero exit, per-attempt logs, the failing
  attempt's output preserved, and `GOFLAGS=-count=1` so Go's test cache
  cannot hide the flake. Proof: `artifacts/orbs-proof/11-flaky-verification.txt`
  (`make-flake-proof.sh`).
* Both commands were also re-run against the published `v0.1.143` artifact
  (commit `d465f916d`): `artifacts/orbs-proof/12-released-commit-flake.txt`.

## Done in the latest pass (items 16 coverage, 21 approval)

* **Item 16 (coverage gate)** — `star verify --min-coverage PERCENT`:
  measured statement coverage from a real coverage profile, an
  UNDER-COVERED verdict (non-zero exit) for a passing run below the
  threshold, a COVERAGE-UNAVAILABLE verdict with a named reason when no
  number is obtainable, and no relabelling of failing or flaky runs.
  Proof: `artifacts/orbs-proof/13-coverage-gate.txt`
  (`make-coverage-gate-proof.sh`).
* **Item 21 (approval gate)** — project policy from `crush.json`
  (`require_approval`, `protected_branches`, `approver`), `star commit
  --policy` to inspect it and `--approve` to satisfy it; a refusal happens
  before staging and leaves the tree and history untouched. Proof:
  `artifacts/orbs-proof/14-approval-gate.txt`
  (`make-approval-gate-proof.sh`).
* Both are reproducible byte-for-byte: each script was re-run after a
  clean rebuild and produced identical transcripts.

## Next (unblocked, partially built)

* **Item 7** — port→preview mapping, crash supervision and restart policy;
  both need harness access (blocker 1).
* **Item 20** — per-workspace scopes and an approval prompt for sensitive
  operations, so one token is not all-or-nothing.
* **Item 21** — pull-request creation (needs a remote and credentials);
  the approval gate itself is done.
* **Items 22, 23** — per-orb resource and permission surfaces, once the
  harness side is reachable.
