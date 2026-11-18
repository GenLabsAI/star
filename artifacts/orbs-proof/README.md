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

# and the same three against a published build:
gh release download v0.1.141 --repo GenLabsAI/star -p 'star-linux-amd64.tar.gz' -D /tmp/rel
tar -C /tmp/rel -xzf /tmp/rel/star-linux-amd64.tar.gz
./artifacts/orbs-proof/make-release-commands-proof.sh /tmp/rel/star-linux-amd64/star-core /tmp/release-proof
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
| `make-proof.sh` | The script that produced 01–04, so nothing is hand-edited. |
| `make-knowledge-proof.sh`, `make-ports-proof.sh`, `make-pairing-proof.sh` | The scripts that produced 06, 07 and 08. |
| `make-release-commands-proof.sh` | The script that produced 09: it runs the three scripts above with a released binary and records its version and commit. |
| `make-commit-proof.sh`, `make-flake-proof.sh` | The scripts that produced 10 and 11. |
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

## What this does NOT prove

These runs execute on the local machine with no cloud orb involved: they
prove the verification and change-management loop, not cloud isolation
(items 1–6 are covered by `artifacts/cloud-proof/`). They also do not
prove behaviour on a stack other than Go; the detection and command plan
for Node, Rust, Python and Docker Compose are covered by unit tests in
`internal/envreport` and `internal/verify`.
