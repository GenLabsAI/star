# Orbs proof pack

Evidence that the requirements work, captured from a real run rather than
described. Everything here was produced by `star` binaries on a scratch
git repository, and every claim below can be re-run.

Reproduce it yourself:

```sh
go build -o /tmp/star .
./artifacts/orbs-proof/make-proof.sh /tmp/star /tmp/proof-out

# and against a published build:
gh release download v0.1.139 --repo GenLabsAI/star -p 'star-linux-amd64.tar.gz' -D /tmp/rel
tar -C /tmp/rel -xzf /tmp/rel/star-linux-amd64.tar.gz
./artifacts/orbs-proof/make-release-proof.sh /tmp/rel/star-linux-amd64/star-core /tmp/release-proof

# the newest commands, each self-contained (needs python3 and curl):
./artifacts/orbs-proof/make-knowledge-proof.sh /tmp/star /tmp/proof-out
./artifacts/orbs-proof/make-ports-proof.sh /tmp/star /tmp/proof-out
./artifacts/orbs-proof/make-pairing-proof.sh /tmp/star /tmp/proof-out
./artifacts/orbs-proof/make-commit-proof.sh /tmp/star /tmp/proof-out
./artifacts/orbs-proof/make-flake-proof.sh /tmp/star /tmp/proof-out   # needs go on PATH
./artifacts/orbs-proof/make-coverage-gate-proof.sh /tmp/star /tmp/proof-out   # needs go on PATH
./artifacts/orbs-proof/make-approval-gate-proof.sh /tmp/star /tmp/proof-out
./artifacts/orbs-proof/make-live-app-proof.sh /tmp/star /tmp/proof-out       # needs go on PATH
./artifacts/orbs-proof/make-overview-activity-proof.sh /tmp/star /tmp/proof-out   # needs python3, curl
./artifacts/orbs-proof/make-peer-scopes-proof.sh /tmp/star /tmp/proof-out    # needs curl
./artifacts/orbs-proof/make-pull-request-proof.sh /tmp/star /tmp/proof-out .  # pushes+deletes throwaway branches

# and both against a published build:
gh release download v0.1.146 --repo GenLabsAI/star -p 'star-linux-amd64.tar.gz' -D /tmp/rel4
tar -C /tmp/rel4 -xzf /tmp/rel4/star-linux-amd64.tar.gz
./artifacts/orbs-proof/make-release-coverage-approval-proof.sh /tmp/rel4/star-linux-amd64/star-core /tmp/release-proof

# and the same three against a published build:
gh release download v0.1.141 --repo GenLabsAI/star -p 'star-linux-amd64.tar.gz' -D /tmp/rel
tar -C /tmp/rel -xzf /tmp/rel/star-linux-amd64.tar.gz
./artifacts/orbs-proof/make-release-commands-proof.sh /tmp/rel/star-linux-amd64/star-core /tmp/release-proof

# and the commit/flake pair against the newest build:
gh release download v0.1.143 --repo GenLabsAI/star -p 'star-linux-amd64.tar.gz' -D /tmp/rel3
tar -C /tmp/rel3 -xzf /tmp/rel3/star-linux-amd64.tar.gz
./artifacts/orbs-proof/make-release-commit-flake-proof.sh /tmp/rel3/star-linux-amd64/star-core /tmp/release-proof
```

Each script rebuilds its scratch project from scratch, starts and stops its
own processes, and can be re-run in any order.

## Files

| File | What it proves |
| --- | --- |
| `01-verify-fail.txt` | `star verify` runs the project's real build and test commands, catches a deliberately wrong test expectation, prints the failing assertion, exits **1**, and leaves a bundle containing `verify.json`, `summary.md` and one log per command. |
| `01-fail-bundle/` | The failure bundle itself: the verdict, the exact commands, the exit code, and the command output that produced it. |
| `02-verify-pass.txt` | After the code is fixed, the same command exits **0**, every step passes, and the bundle records the two changed files and the tracked diff. Includes `verify.json`, `logs/02-test.log` and `changes.patch` verbatim. |
| `02-pass-bundle/` | The passing bundle, including a patch that matches the working tree at the time of the run. |
| `03-changes.txt` | `star changes` summarizes the tree (branch, per-file status, diffstat, patch size), emits JSON on stdout for agents, and writes a reviewable patch with `--patch`. |
| `04-overview.txt` | `star overview` joins tracked workspaces with their branch, changed paths, latest verification verdict and cloud credential state, in text and JSON. Run against a `projects.json` that points at the scratch repo. |
| `05-released-binary.txt` | The same fail → pass → changes → overview loop run against the **published** `v0.1.139` artifact (`star-linux-amd64.tar.gz` from the GitHub release), which stamps its own version and commit (`bc4e684e8`) into every record. This ties the claims to a shipped binary, not a local build. |
| `06-knowledge.txt` | `star knowledge` keeps a discovery across **processes**: `list` is empty, `show` on a missing name exits **1**, `add` in one process, then a fresh process reads it back with a byte-identical sha256, prints the JSON listing and the `.md` file it lives in, `star env` reports it, a second entry and a replacement work, and `remove` empties the store again. |
| `07-ports.txt` | `star ports` reports a real `python3 -m http.server` started by the script: port 18765, its actual pid, `python3`, and the workspace as cwd. A second listener started *outside* the workspace is hidden by default and visible with `--all`. Killing the reported pid removes the entry (`listeners: []`), and an idle workspace prints "No listening sockets". |
| `08-pairing.txt` | A real `star server` with a pairing token: request with no credential → **401**, wrong token → **401** (real `401 Unauthorized` headers, not a refused connection), correct token → **200**, and the same via `X-Star-Token`. The token is never printed (length + sha256 prefix only), no file is written by `star pair`, `star env` reports the token source without the value, the token file is `0600`, the server appears in `star ports` and its port is gone after shutdown, and a server started *without* a token still answers 200 (unchanged local behaviour). |
| `09-released-commands.txt` | Items 14, 7 and 20 re-run with the **published** `v0.1.141` artifact. `star env --json` reports version `v0.1.141` and commit `1022fd03a` (this repository's commit), then the three proofs pass unchanged: knowledge read back identical, the real listener found and its pid killed, 401/401/200 pairing. |
| `10-commit.txt` | `star commit` creates a real commit: a tracked edit plus an untracked file; it refuses while the new file is unapproved (`exit=1`, `git log` still shows one commit), refuses without a message, then with `--include-untracked --branch feat/greeting` creates the commit and the branch (`git show HEAD:RUNBOOK.md` prints the new file, `git status` is empty, `main` still has one commit). The JSON record's hash is compared against `git rev-parse HEAD`, and the repository has no remote, so nothing was pushed. |
| `11-flaky-verification.txt` | `star verify --repeat` catches a flake a single run cannot: one run of the same project prints PASS (exit 0), three runs print FLAKY (exit 1) with `runs: pass, fail, pass`, the bundle keeps one log per attempt including the failing `02-test-run2.log`, and the JSON record shows attempts `[0, 1, 0]` with `flaky: true`. After the test is fixed the same repeated run passes, and a test that fails every time is reported as FAIL, not flaky. |
| `13-coverage-gate.txt` | `star verify --min-coverage` enforces a coverage floor. The scratch Go project's own `go tool cover` reports 33.3% (1 of 3 statements); `--min-coverage 90` exits **1** with `Coverage gate failed: 33.3% measured, 90.0% required (1 of 3 statements)`, the bundle's `summary.md` records the same number, `--min-coverage 20` exits **0**, `--min-coverage 100` still reports 33.3%, and no flag leaves the record without a coverage block. A Node project with no coverage provider prints `coverage-unavailable` / `no coverage provider is known for the detected stack(s): Node` and exits **1** rather than passing. A failing test stays `fail` and a flaky `--repeat 3` run stays `flaky` with attempts `[0, 1, 0]`. `--min-coverage 120` is rejected. |
| `14-approval-gate.txt` | `star commit` honours a project policy from `crush.json`. With `"require_approval": true` the commit is refused (`exit=1`, `Refusing to commit: ... requires approval ... for every commit`), and `git rev-list --count HEAD`, `git status --porcelain` and the empty `git diff --cached` are identical before and after. With `--approve` the same commit lands: the record's hash equals `git rev-parse HEAD`, the count goes 1 → 2, the tree is clean, and the policy block records `approval_required: true, approved: true`. A `"protected_branches": ["main"]` policy refuses `main` but allows `--branch feat/gate-out`, and a repository with no policy commits exactly as before. `star commit --policy` prints the effective policy; neither repository has a remote. |
| `15-released-coverage-approval.txt` | Both gates re-run with the **published** `v0.1.146` artifact, which reports commit `7632a0e62` (this repository's commit): the same UNDER-COVERED, COVERAGE-UNAVAILABLE, flaky, policy-refusal and approved-commit results as the source-build transcripts, printed by the binary users download. |
| `12-released-commit-flake.txt` | Items 21 and 16 re-run with the **published** `v0.1.143` artifact, which reports commit `d465f916d` (this repository's commit): the commit refusals and the approved commit, the hidden flake under one run, the FLAKY verdict under three, the fixed pass, and the consistent failure. |
| `make-proof.sh` | The script that produced 01–04, so nothing is hand-edited. |
| `make-knowledge-proof.sh`, `make-ports-proof.sh`, `make-pairing-proof.sh` | The scripts that produced 06, 07 and 08. |
| `make-release-commands-proof.sh` | The script that produced 09: it runs the three scripts above with a released binary and records its version and commit. |
| `make-commit-proof.sh`, `make-flake-proof.sh` | The scripts that produced 10 and 11. |
| `make-release-coverage-approval-proof.sh` | The script that produced 15: it runs the two scripts above with a released binary and records its version and commit. |
| `make-coverage-gate-proof.sh`, `make-approval-gate-proof.sh` | The scripts that produced 13 and 14. They report selected JSON fields instead of raw output, so a clean rebuild reproduces the transcripts byte-for-byte (commit hashes are checked against `git rev-parse HEAD` rather than printed). |
| `make-release-commit-flake-proof.sh` | The script that produced 12: it runs those two with a released binary and records its version and commit. |
| `16-live-application.txt` | `star verify --app` verifies the **running** application, not just its tests. The scratch Go service has a real nil-map bug: `go test ./...` passes (`ok`), and `star verify --app 'PORT=18111 go run .' --app-url .../items` exits **1** with `verdict: live-fail`, `status_code: 500`, `reason: the application answered 500, expected 200`, the service's own log line `items handler failed: assignment to entry in nil map` kept in the bundle, and nothing listening on the port afterwards (`curl` exit 7, no `star ports` entry). After the one-line fix the same command exits **0** with `status_code: 200` and the real body `{"count":1}`. |
| `17-overview-activity.txt` | `star overview` shows the activity inside each workspace. A real `star server` creates a real workspace and two real conversations (HTTP 200 each, `star session list` shows both); a durable wakeup store with two schedules is seeded (the CLI needs a model credential this machine lacks); a real `python3 -m http.server` runs inside the workspace and answers 200. The overview then reports `2 conversation(s) · 2 armed wakeup(s) · preview http://127.0.0.1:18422` plus totals, the JSON record carries the same counts, killing the preview drops previews to 0 while sessions and schedules stay, and a deleted workspace is reported as `workspace directory is missing` with `exit=0` instead of failing the report. |
| `18-peer-scopes.txt` | Three real servers with real tokens. A `--scope read` token: `GET /v1/health` → **200**, `GET /v1/workspaces` → **200**, `POST /v1/workspaces` → **403** (`this pairing token grants \"read\" ... requires \"write\"`), the shell route → **403** (`requires \"exec\"`), while a peer with no token still gets **401** (who, before what). With `--pair-approval-file`: the shell route is **428** with no approval, **428** with a wrong code, then reaches the handler (**404 workspace not found**, not 428) with the current code, and **428** with an expired one; reads need no approval. With `--pair-workspace`: creating a workspace outside the grant is **403**, inside is **200**, and that workspace is readable by id. A token without a scope prefix still creates a workspace anywhere, so older tokens are unchanged. |
| `19-pull-request.txt` | `star pr` refuses a branch that is not on the remote (`exit=1`, `is not pushed to origin`, and the remote gains no branch). With a recording `gh` on `PATH` it publishes the branch (`--push`), prints `Pull request #4242`, and the recorded invocation is exactly `pr create --repo GenLabsAI/star --head proof/pr-stub --base main --title ... --body ...`. Against the real GitHub CLI the same command pushes for real and then reports GitHub's own answer: `pull request create failed: GraphQL: Resource not accessible by personal access token (createPullRequest)` — this machine's credential may not create pull requests (blocker 5 in `docs/orbs.md`). Afterwards no throwaway branch is left on the remote or locally and `git status --short` is empty. |
| `make-live-app-proof.sh`, `make-overview-activity-proof.sh`, `make-peer-scopes-proof.sh`, `make-pull-request-proof.sh` | The scripts that produced 16–19. Each was run twice and reproduced its transcript byte-for-byte; the pull-request one pushes and deletes its own throwaway branches and leaves the repository clean. |
| `make-release-proof.sh` | The script that produced 05 against a released binary. |

## What a skeptic should check

1. The exit codes: `exit=1` on the failing run, `exit=0` on the passing run.
2. The failure text in `01-verify-fail.txt` (`Sum(2, 2) = 4, want 5`) is the
   real assertion from the scratch project's `sum_test.go`, and it also
   appears inside `01-fail-bundle/logs/02-test.log`.
3. `02-pass-bundle/changes.patch` is a real `git diff`: it changes the
   expectation and adds `Double`, exactly what the script edited.
4. `02-pass-bundle/verify.json` contains the same commands and exit codes
   the terminal showed, plus the environment manifest (branch, stacks,
   runtime readiness).
5. In `04-overview.txt`, the overview reports the verification the previous
   command just ran (`pass`, 2 steps, 0 failed) and 2 changed paths, which
   matches `02-pass-bundle/verify.json`.
6. In `05-released-binary.txt`, `star env --json` reports `v0.1.139` and
   commit `bc4e684e8` — the release artifact is running the code in this
   repository, and the verdicts match the source-build transcript.
7. Delete the scratch directory and re-run either script: the files
   regenerate with different timestamps and durations but the same
   verdicts.
8. In `06-knowledge.txt`, the two sha256 lines under "verbatim check" are
   equal and the file under `.crush/knowledge/` holds exactly those bytes;
   the `star env` line lists `ci-count`, so a new session finds it without
   being told.
9. In `07-ports.txt`, the pid in the table is the pid the script later
   kills, and the JSON assertions (`workspace`, `filtered`, `cwd`,
   `process`, port list) are evaluated by python3, so a mismatch would
   print a traceback instead of "workspace filter: ...". Re-run the script
   yourself and compare: the pid differs every time, the port and the
   verdicts do not.
10. In `09-released-commands.txt`, the version and commit printed first are
   the release tag's, and the verdict lines below them are the same ones
   the source-build transcripts show — the commands behave the same in the
   binary users download.
11. In `10-commit.txt`, the two refusals exit 1 with `commits: 1` printed after
   them, and the approved commit's short hash equals both the one `star
   commit` printed and the one `git rev-parse --short HEAD` reports; `git
   rev-list --count main` stays at 1, so the work is isolated on the branch.
12. In `11-flaky-verification.txt`, the first run says PASS and the repeated
   run of the same untouched project says FLAKY — that pair is the proof
   that repetition found something one run missed. `--repeat` sets
   `GOFLAGS=-count=1`; without it Go's test cache would replay the first
   run's result and the flake would stay invisible.
13. In `08-pairing.txt`, the three status lines are 401, 401, 200 in that
   order, the 401 carries an `{"message":"unauthorized: ..."}` body, and
   the paired request returns the API response. The token's sha256 prefix
   proves the transcript shows a token was generated without revealing it;
   regenerate the transcript and the prefix changes, the verdicts do not.
14. In `13-coverage-gate.txt`, the `go tool cover` line says 33.3% before
   `star verify` ever runs, and every coverage number afterwards is that
   same 33.3% with the same "1 of 3 statements": the gate reads a real
   profile, it does not echo the threshold. The `COVERAGE-UNAVAILABLE` line
   is a non-zero exit with a reason, not a zero and not a pass.
15. In `14-approval-gate.txt`, the refusal block is bracketed by
   `git rev-list --count HEAD` = `1` and the same `git status --porcelain`
   output on both sides, with an empty `git diff --cached` (nothing staged)
   and `git log --format=%s` still showing only the initial import. The
   approved record then reports `reported hash matches git rev-parse HEAD:
   True`, and the count becomes 2 with a clean tree. `git remote -v` prints
   nothing for either repository, so nothing was pushed.

16. In `15-released-coverage-approval.txt`, the version and commit printed
   first are the release's own (`v0.1.146`, `7632a0e62`), and the verdict
   lines under them match the source-build transcripts exactly — the gates
   work in the binary users download, not just in a local build.
17. In `16-live-application.txt`, the `go test` line says `ok` on the same
   tree where `star verify --app` says `live-fail`, and the 500 is
   explained by the service's own log line (`assignment to entry in nil
   map`). The port check afterwards (`curl` exit 7, no `star ports` entry)
   shows the run cleaned up after itself, and the second run's `status_code:
   200` with body `{"count":1}` shows the verdict follows the code.
18. In `17-overview-activity.txt`, the two conversations were created
   through the server's own HTTP API in the same run, `star session list`
   shows them, and the preview URL the overview prints answers 200 from
   `curl` before the overview runs. The counts go to 0 for previews — and
   only previews — after the preview process is killed, so the view is read
   live rather than cached.
19. In `18-peer-scopes.txt`, the statuses tell the whole story: 200/200/403/403
   for read-only, 401 without a token, 428/428/404/428 for the approval
   gate, 403/200 for the workspace grant. The 404 on the shell route with a
   valid approval is the proof the gate **released** the request: it
   reached the handler, which then failed on the fake workspace id.
20. In `19-pull-request.txt`, the recorded `gh` line is the command
   `star pr` really built (the stub only replaces the binary, not the
   arguments), the pushed branch is listed by `git ls-remote` right after,
   and GitHub's refusal names the missing permission. Re-run it after the
   token is granted `Pull requests: write` and the same script will create,
   inspect and close a real pull request instead.

## What this does NOT prove

These runs execute on the local machine with no cloud orb involved: they
prove the verification and change-management loop, not cloud isolation
(items 1–6 are covered by `artifacts/cloud-proof/`). They also do not
prove behaviour on a stack other than Go; the detection and command plan
for Node, Rust, Python and Docker Compose are covered by unit tests in
`internal/envreport` and `internal/verify`.

The pull-request proof records GitHub's refusal because this machine's `gh`
credential is not allowed to create pull requests; everything up to that
API call (the branch check, the push, the exact `gh` invocation) is real.
Grant the token `Pull requests: write` and re-run `make-pull-request-proof.sh`
to see the same script open and close a real pull request.
