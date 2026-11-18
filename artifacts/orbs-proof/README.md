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
```

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
| `make-proof.sh` | The script that produced 01–04, so nothing is hand-edited. |
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

## What this does NOT prove

These runs execute on the local machine with no cloud orb involved: they
prove the verification and change-management loop, not cloud isolation
(items 1–6 are covered by `artifacts/cloud-proof/`). They also do not
prove behaviour on a stack other than Go; the detection and command plan
for Node, Rust, Python and Docker Compose are covered by unit tests in
`internal/envreport` and `internal/verify`.
