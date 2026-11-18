#!/bin/bash
# Proof for item 21 (commit/branch creation): `star commit` really creates
# a git commit, refuses one without approval, and puts the work on a branch
# of its own without touching any remote.
#
#   ./artifacts/orbs-proof/make-commit-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

PROJ=/tmp/orbs-commit-proj
rm -rf "$PROJ"
mkdir -p "$PROJ"
cd "$PROJ" || exit 1

git init -q -b main
git config user.email proof@example.com
git config user.name "Orbs Commit Proof"
printf 'hello\n' > README.md
printf 'module example.com/commitproof\n' > go.mod
git add -A
git commit -q -m "chore: initial import"

# One tracked edit and one brand new file, the situation `star commit` has
# to handle carefully.
printf 'hello world\n' > README.md
printf 'serve --port 8080\n' > RUNBOOK.md

{
  echo "# scratch repository: $PROJ (branch main, one commit)"
  echo "# working tree: README.md edited, RUNBOOK.md is new and untracked"
  echo
  echo '$ star changes        # what would be committed'
  "$STAR" changes
  echo "exit=$?"
  echo
  echo '# the untracked file is a gate: it blocks the commit until approved'
  echo '$ star commit --message "fix: say hello world"'
  "$STAR" commit --message "fix: say hello world" 2>&1
  echo "exit=$?"
  echo '$ git log --oneline    # no commit was created by the refusal'
  git log --oneline
  echo "commits: $(git rev-list --count HEAD)"
  echo '$ git status --porcelain   # the work is still there, uncommitted'
  git status --porcelain
  echo
  echo '# a message is also required, even when the gate is opened'
  echo '$ star commit --include-untracked'
  "$STAR" commit --include-untracked 2>&1
  echo "exit=$?"
  echo "commits: $(git rev-list --count HEAD)"
  echo
  echo '# approved: commit both paths on a branch of their own'
  echo '$ star commit -m "fix: say hello world" --include-untracked --branch feat/greeting'
  "$STAR" commit -m "fix: say hello world" --include-untracked --branch feat/greeting
  echo "exit=$?"
  echo
  echo '# the commit is real, and it is the one the command reported'
  echo '$ git log --oneline --stat -1'
  git log --oneline --stat -1
  echo "\$ git rev-parse --short HEAD -> $(git rev-parse --short HEAD)"
  echo "\$ git branch --show-current -> $(git branch --show-current)"
  echo "\$ git show HEAD:RUNBOOK.md -> $(git show HEAD:RUNBOOK.md)"
  echo "\$ git status --porcelain -> '$(git status --porcelain)' (empty: nothing left over)"
  echo "\$ git rev-list --count main -> $(git rev-list --count main) (main still has only the import)"
  echo
  echo '# machine-readable record, second commit made on the same branch'
  printf 'hello again\n' > README.md
  echo '$ star commit -m "fix: greet again" --branch feat/greeting --json'
  "$STAR" commit -m "fix: greet again" --branch feat/greeting --json | tee /tmp/orbs-commit-record.json
  echo "exit=$?"
  "$STAR" commit -m "fix: greet again" --branch feat/greeting --json >/dev/null 2>&1
  echo "second run exit=$? (1: nothing left to commit)"
  python3 - <<'PY'
import json, subprocess
record = json.load(open("/tmp/orbs-commit-record.json"))
head = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
short = subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], text=True).strip()
assert record["hash"] == head, (record["hash"], head)
assert record["short"] == short, (record["short"], short)
assert record["branch"] == "feat/greeting", record["branch"]
assert not record.get("branch_created"), record  # the branch already existed
print("record hash matches git HEAD:", record["short"])
print("branch:", record["branch"], "| paths:", [f["path"] for f in record["files"]])
PY
  echo "exit=$?"
  echo
  echo '# nothing was pushed: this repository has no remote at all'
  echo "\$ git remote -v -> '$(git remote -v)' (empty)"
  echo
  echo '# a clean tree cannot be committed twice'
  echo '$ star commit -m "chore: nothing changed"'
  "$STAR" commit -m "chore: nothing changed" 2>&1
  echo "exit=$?"
  echo "commits on $(git branch --show-current): $(git rev-list --count HEAD)"
} > "$OUT/10-commit.txt" 2>&1
