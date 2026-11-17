# Orbs — "We want..." list

Source: Amp's deeper notes ("Orbs, Explained", "Putting an Agent in an Orb",
"What I Want to Tell You About Orbs"). Orbs are more than basic remote VMs:
specialized lifecycle hooks, codebases optimized for agent discovery,
debugging portals, and proof-generation loops.

Status legend: DONE = delivered end to end (on `main`, proven). WIP = code
exists but not closed out. OPEN = not started. SKIPPED = explicitly deferred.

## The list

1. **Fully isolated cloud sandboxes** that spin up automatically on demand,
   ensuring the agent never executes code or consumes resources on the
   user's local machine. — DONE (containerized harness + cloud runs proven
   e2e; see `artifacts/harness/`, `artifacts/cloud-proof/`).
2. **A 1:1 mapping between chat threads and execution environments**, so
   each agent session has its own dedicated workspace, terminal, and
   filesystem. — DONE (chat-affine cloud runs on `main`).
3. **Persistent state and project volumes** so files, code changes, and
   installed packages survive restarts, tool calls, and prompt wake-ups. —
   DONE (stash restore + snapshot cache on `main`).
4. **Pre-configured setup via `.agents/setup`** installing runtimes,
   databases (Postgres/Redis), and system packages, caching the snapshot
   for instant future launches. — DONE (orb lifecycle hooks + snapshot
   cache on `main`).
5. **Continuous validation via `.agents/resume`** on every wake-up:
   re-authenticate tokens, refresh connections, verify service health. —
   DONE (same lifecycle-hooks pass as 4).
6. **Intelligent sleep and wake states**: pause when idle (zero cost while
   sleeping), restore full context on new prompt/webhook. — DONE (idle
   reuse + stash restore + dead-runner self-heal on `main`).
7. **Scheduled and event-driven wakeups** (cron / webhooks like GitHub CI
   failures), integrated with Star's Schedule Wakeup tool, working on the
   cloud too. — DONE on `main` (durable store + cron engine + `star wakeup`
   CLI + 15-min GH sweep + webhook trigger + explicit timer arming +
   TUI toast; unit tests green, headless CLI proof: schedule →
   backdate → `wakeup run` fired exactly once). Release-binary
   verification still to do.
8. **Pre-loaded agent-friendly toolchain** (Debian + git, `gh`, tmux,
   ffmpeg, ImageMagick, ripgrep, Node/Bun/pnpm, headless browsers). — DONE
   (`artifacts/harness/Dockerfile.agent`, proven by e2e orb runs).
9. **End-to-end proof generation**: run apps, test matrices, screenshots,
   demo videos — irrefutable proof before bothering the human. — DONE
   (screenshot/video/gif + e2e log artifacts, TUI walkthrough skill).
10. **Live HTTP port previews and portals** to view/annotate/chat about
    live web apps. — SKIPPED (explicitly deferred by user).
11. **Multi-agent nesting and spawning**: spin up other orbs, pass files,
    coordinate handoffs, manage parallel sub-tasks. — OPEN.
12. **Cross-device continuity via native apps & TUI**: start at desktop,
    monitor/answer/reply from any authorized Star Desktop or CLI. — DONE
    (cloud-mode TUI with live turn streaming; `tui-cloud-live.txt` /
    `tui-cloud-prod.txt` proofs).
