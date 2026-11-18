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
   8, 9, 22) lands in the harness image, the workflow, or that API.
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
   the agent debugging infrastructure. — **WIP**. Background jobs and
   streaming via `internal/shell` (`job_output`, `job_kill`, `monitor`),
   runner self-heal. Missing: a port registry/preview mapping and crash
   supervision with restart policy; needs item 2's harness access.
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
    across sessions. — **WIP**. Context files (AGENTS.md/CLAUDE.md/CRUSH.md
    and `.local` variants), builtin and project skills, `search_history`,
    MCP resources, and `star env` now lists which of these a project has.
    Missing: a durable store of discoveries shared across sessions.
15. **A complete agent-oriented development interface**: predictable
    filesystem locations, machine-readable environment information,
    documented commands, readiness checks, accessible logs, clear
    diagnostics. — **WIP → advanced here**. `star env [--json]` now emits
    the manifest (paths, git, detected stacks with install/build/test
    commands, lifecycle hooks and their hashes, runtime readiness,
    credential *presence* only), alongside the existing `star dirs`,
    `star logs`, `crush_info` and `crush_logs`. Missing: port and process
    inventory, which needs item 7.
16. **Autonomous testing and verification** against the real running
    application: reproduce bugs, investigate failures, run regressions,
    iterate until it works. — **WIP → advanced here**. `star verify` runs
    the commands `star env` detected (build and test by default, install
    with `--install`), bounds each with `--timeout`, records exit code,
    duration and output, and exits non-zero when anything fails. Evidence
    in `artifacts/orbs-proof/`. Missing: coverage thresholds, flake
    detection, and driving a running application rather than its test
    suite.
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
    specifics or local resources. — **WIP**. Star's client/server split
    lets a remote client drive a workspace on the machine that runs
    `star server`; there is no pairing/authorization model or opt-in
    worker policy yet.
21. **Complete change management**: track every modification, produce clean
    diffs, commits, branches and pull requests, isolated until review. —
    **WIP → advanced here**. Per-session git worktrees
    (`internal/worktree`, `enter_worktree`/`exit_worktree`), the diff tool,
    `internal/gitutil` branch detection, and now `star changes`: branch and
    upstream divergence, every touched path, diffstat and patch size, JSON
    for agents, and `--patch FILE` to hand a reviewer a real diff. `star
    verify` bundles carry the same patch, so a diff never travels without
    the run that produced it. Missing: commit/branch/PR creation in Star
    and an approval gate before changes leave the orb.
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

## Next (unblocked, partially built)

* **Item 14** — a durable cross-session knowledge pointer, so discoveries
  survive sessions instead of being rediscovered.
* **Item 7** — a port and process inventory for background jobs (local
  half only; the orb-side mapping needs harness access).
* **Items 22, 23** — per-orb resource and permission surfaces, once the
  harness side is reachable.
