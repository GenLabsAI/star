#!/bin/bash
# Proof for item 21 (approval gate): a project policy in crush.json can
# refuse a commit before it touches the tree or the history, and the same
# commit succeeds once it is approved with --approve.
#
#   ./artifacts/orbs-proof/make-approval-gate-proof.sh /path/to/star OUTDIR
#
# Deterministic by construction: commit hashes vary with the commit time, so
# the transcript verifies the reported hash against `git rev-parse HEAD`
# instead of printing it, and never prints a wall clock.
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

PROJ=/tmp/orbs-approval-proj
FREE=/tmp/orbs-approval-free
TMP=/tmp/orbs-approval-tmp
rm -rf "$PROJ" "$FREE" "$TMP"
mkdir -p "$PROJ" "$FREE" "$TMP"

new_repo() {
  local dir=$1
  git -C "$dir" init -q -b main
  git -C "$dir" config user.email proof@example.com
  git -C "$dir" config user.name "Orbs Approval Proof"
}

# The policy lives in the project config file the rest of the CLI already
# reads: crush.json.
new_repo "$PROJ"
cat > "$PROJ/crush.json" <<'EOF'
{
  "commit": {
    "require_approval": true,
    "approver": "the repo owner"
  }
}
EOF
printf 'version 1\n' > "$PROJ/app.txt"
git -C "$PROJ" add -A
git -C "$PROJ" commit -q -m "chore: initial import"

# show_record prints the parts of a commit record this proof is about, and
# checks the reported hash against the repository itself.
show_record() {
  python3 - "$1" "$2" <<'PY'
import json, subprocess, sys
record = json.load(open(sys.argv[1]))
root = sys.argv[2]
head = subprocess.run(["git", "-C", root, "rev-parse", "HEAD"],
                      capture_output=True, text=True, check=True).stdout.strip()
print("reported hash matches git rev-parse HEAD:", record["hash"] == head)
print("reported hash is a full 40-hex object id:", len(record["hash"]) == 40)
print("reported short matches the hash prefix:", head.startswith(record["short"]))
print("branch:", record["branch"], "| message:", record["message"],
      "| committed paths:", len(record["files"]))
policy = record.get("policy")
if policy is None:
    print("policy: none recorded")
else:
    print("policy:", policy["path"], "| approval_required:", policy["approval_required"],
          "| approved:", policy["approved"], "| require_approval:", policy["require_approval"])
PY
}

{
  echo "# scratch repository: $PROJ"
  echo
  echo '# the policy is a project file, discovered the same way the CLI finds its config'
  echo '$ cat /tmp/orbs-approval-proj/crush.json'
  cat "$PROJ/crush.json"
  echo
  echo '$ star commit --policy'
  ( cd "$PROJ" && "$STAR" commit --policy )
  echo "exit=$?"
  echo
  echo '# one tracked edit, nothing staged, one commit in history'
  printf 'version 2\n' > "$PROJ/app.txt"
  echo '$ cd /tmp/orbs-approval-proj && git rev-list --count HEAD'
  git -C "$PROJ" rev-list --count HEAD
  echo '$ git status --porcelain'
  git -C "$PROJ" status --porcelain
  echo '$ git diff --cached --name-only   # nothing staged'
  git -C "$PROJ" diff --cached --name-only
  echo
  echo '# without approval the policy refuses, and nothing moves'
  echo '$ star commit -m "feat: bump version"'
  ( cd "$PROJ" && "$STAR" commit -m "feat: bump version" ) 2> "$TMP/refused.err"
  echo "exit=$?"
  cat "$TMP/refused.err"
  echo '$ git rev-list --count HEAD'
  git -C "$PROJ" rev-list --count HEAD
  echo '$ git status --porcelain   # identical to before: still one modified file'
  git -C "$PROJ" status --porcelain
  echo '$ git diff --cached --name-only   # the refusal staged nothing'
  git -C "$PROJ" diff --cached --name-only
  echo '$ git log --format=%s'
  git -C "$PROJ" log --format=%s
  echo
  echo '# the refusal is also an error in machine-readable mode'
  echo '$ star commit -m "feat: bump version" --json'
  ( cd "$PROJ" && "$STAR" commit -m "feat: bump version" --json ) > "$TMP/refused.json" 2> "$TMP/refused-json.err"
  echo "exit=$?"
  cat "$TMP/refused-json.err"
  echo "record bytes: $(wc -c < "$TMP/refused.json")"
  echo
  echo '# approved once, the same commit lands and reports the real object id'
  echo '$ star commit -m "feat: bump version" --approve --json'
  ( cd "$PROJ" && "$STAR" commit -m "feat: bump version" --approve --json ) > "$TMP/approved.json" 2> "$TMP/approved.err"
  echo "exit=$?"
  cat "$TMP/approved.err"
  show_record "$TMP/approved.json" "$PROJ"
  echo
  echo '# the repository confirms it: one more commit, a clean tree, the new message'
  echo '$ git rev-list --count HEAD'
  git -C "$PROJ" rev-list --count HEAD
  echo '$ git status --porcelain'
  git -C "$PROJ" status --porcelain
  echo '$ git log --format=%s'
  git -C "$PROJ" log --format=%s
  echo
  echo '# a policy can protect only some branches: main is gated, a feature branch is not'
  cat > "$PROJ/crush.json" <<'EOF'
{
  "commit": {
    "protected_branches": ["main"]
  }
}
EOF
  ( cd "$PROJ" && "$STAR" commit -m "chore: protect main only" --approve --json ) > "$TMP/policy.json" 2>&1
  echo "policy change exit=$?"
  printf 'version 3\n' > "$PROJ/app.txt"
  echo '$ star commit -m "feat: land on main"'
  ( cd "$PROJ" && "$STAR" commit -m "feat: land on main" ) 2> "$TMP/protected.err"
  echo "exit=$?"
  cat "$TMP/protected.err"
  echo '$ git rev-list --count HEAD'
  git -C "$PROJ" rev-list --count HEAD
  echo '$ star commit -m "feat: land on a feature branch" --branch feat/gate-out --json'
  ( cd "$PROJ" && "$STAR" commit -m "feat: land on a feature branch" --branch feat/gate-out --json ) > "$TMP/feature.json" 2> "$TMP/feature.err"
  echo "exit=$?"
  cat "$TMP/feature.err"
  show_record "$TMP/feature.json" "$PROJ"
  echo
  echo '# a project without a policy is not gated: no new gate for everyone else'
  new_repo "$FREE"
  printf 'hello\n' > "$FREE/README.md"
  git -C "$FREE" add -A
  git -C "$FREE" commit -q -m "chore: initial import"
  printf 'hello again\n' > "$FREE/README.md"
  echo '$ cd /tmp/orbs-approval-free && star commit --policy'
  ( cd "$FREE" && "$STAR" commit --policy )
  echo "exit=$?"
  echo '$ star commit -m "docs: say hello again" --json'
  ( cd "$FREE" && "$STAR" commit -m "docs: say hello again" --json ) > "$TMP/free.json" 2> "$TMP/free.err"
  echo "exit=$?"
  cat "$TMP/free.err"
  show_record "$TMP/free.json" "$FREE"
  echo '$ git rev-list --count HEAD'
  git -C "$FREE" rev-list --count HEAD
  echo
  echo '# nothing left the machine: neither repository has a remote'
  echo '$ git remote -v   (in both repositories)'
  echo "[$PROJ]"; git -C "$PROJ" remote -v
  echo "[$FREE]"; git -C "$FREE" remote -v
  echo "(empty output above means no remote is configured)"
} > "$OUT/14-approval-gate.txt" 2>&1
