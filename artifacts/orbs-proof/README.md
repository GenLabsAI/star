# Orbs proof pack

Evidence that the requirements work, captured from a real run rather than
described. Everything here was produced by `star` binaries on a scratch
git repository, and every claim below can be re-run.

Reproduce it yourself:

```sh
go build -o /tmp/star .
./artifacts/orbs-proof/make-proof.sh /tmp/star /tmp/proof-out
```

## Files

| File | What it proves |
| --- | --- |
| `01-verify-fail.txt` | `star verify` runs the project's real build and test commands, catches a deliberately wrong test expectation, prints the failing assertion, exits **1**, and leaves a bundle containing `verify.json`, `summary.md` and one log per command. |
| `01-fail-bundle/` | The failure bundle itself: the verdict, the exact commands, the exit code, and the command output that produced it. |
| `02-verify-pass.txt` | After the code is fixed, the same command exits **0**, every step passes, and the bundle records the two changed files and the tracked diff. Includes `verify.json`, `logs/02-test.log` and `changes.patch` verbatim. |
| `02-pass-bundle/` | The passing bundle, including a patch that matches the working tree at the time of the run. |
| `03-changes.txt` | `star changes` summarizes the tree (branch, per-file status, diffstat, patch size), emits JSON on stdout for agents, and writes a reviewable patch with `--patch`. |
| `make-proof.sh` | The script that produced all of the above, so nothing is hand-edited. |

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
5. Delete the scratch directory and re-run the script: the files regenerate
   with different timestamps and durations but the same verdicts.

## What this does NOT prove

These runs execute on the local machine with no cloud orb involved: they
prove the verification and change-management loop, not cloud isolation
(items 1–6 are covered by `artifacts/cloud-proof/`). They also do not
prove behaviour on a stack other than Go; the detection and command plan
for Node, Rust, Python and Docker Compose are covered by unit tests in
`internal/envreport` and `internal/verify`.
