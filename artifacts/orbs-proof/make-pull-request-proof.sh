#!/bin/bash
# Proof for item 21 (pull requests): `star pr` refuses a branch that is not
# on the remote, publishes one when asked, and then calls the GitHub CLI
# with exactly the arguments a reviewer would recognise.
#
#   ./artifacts/orbs-proof/make-pull-request-proof.sh /path/to/star OUTDIR [REPO]
#
# REPO defaults to the current directory. The script pushes throwaway
# branches to that repository's remote and deletes them again; it creates a
# pull request only when GitHub allows it, and closes it immediately.
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
REPO="${3:-$PWD}"
mkdir -p "$OUT"

STUB_DIR=$OUT/pr-gh-stub
STUB_LOG=$OUT/pr-gh-stub-args.txt
BASE=$(git -C "$REPO" rev-parse --abbrev-ref HEAD)

# GitHub's CLI reports uncommitted files in its output, so a dirty tree
# would make this transcript describe something else every run.
if [ -n "$(git -C "$REPO" status --porcelain)" ]; then
  echo "$REPO has uncommitted changes; commit or stash them first (this proof needs a clean tree)" >&2
  exit 1
fi

rm -rf "$STUB_DIR"
mkdir -p "$STUB_DIR"

remote_branches() {
  git -C "$REPO" ls-remote --heads origin 'refs/heads/proof/*' 2>/dev/null | awk '{print $2}'
}

cleanup() {
  git -C "$REPO" checkout -q "$BASE" 2>/dev/null
  for branch in proof/pr-unsent proof/pr-stub proof/pr-creation; do
    git -C "$REPO" branch -D "$branch" > /dev/null 2>&1
    git -C "$REPO" push -q origin --delete "$branch" > /dev/null 2>&1
  done
}
trap cleanup EXIT

# An interrupted earlier run may have left throwaway branches behind; a
# leftover branch on the remote would reject this run's push.
for branch in proof/pr-unsent proof/pr-stub proof/pr-creation; do
  git -C "$REPO" push -q origin --delete "$branch" > /dev/null 2>&1 || true
  git -C "$REPO" branch -D "$branch" > /dev/null 2>&1 || true
done

# commit_marker BRANCH -- a real change on a throwaway branch, so the pull
# request would describe a real diff.
commit_marker() {
  local branch="$1"
  git -C "$REPO" checkout -q -B "$branch" "$BASE"
  printf 'temporary marker for %s\n' "$branch" > "$REPO/.pr-proof-marker"
  git -C "$REPO" add .pr-proof-marker
  git -C "$REPO" -c user.email=proof@example.com -c user.name=proof \
    commit -q -m "chore: temporary commit for the pull-request proof"
  rm -f "$REPO/.pr-proof-marker"
}

{
  echo "# repository: $REPO"
  echo "# base branch: $BASE (the branch this script returns to and leaves clean)"
  echo "# remote: $(git -C "$REPO" remote get-url origin)"
  echo

  echo "## A branch that is not on the remote"
  echo
  commit_marker proof/pr-unsent
  echo '$ star pr --title "chore: unsent branch"'
  "$STAR" pr --title "chore: unsent branch" 2>&1 | tail -4
  echo "exit=${PIPESTATUS[0]}"
  echo "branches on the remote after the refusal: $(remote_branches | wc -l)"
  git -C "$REPO" checkout -q "$BASE"
  git -C "$REPO" branch -D proof/pr-unsent > /dev/null 2>&1
  echo

  echo "## What the command asks GitHub"
  echo
  cat > "$STUB_DIR/gh" <<EOF
#!/bin/bash
# Records the arguments star passes to the GitHub CLI, then answers with a
# pull-request URL, so the invocation can be inspected without a network.
printf '%s\n' "\$*" >> "$STUB_LOG"
echo "https://github.com/GenLabsAI/star/pull/4242"
EOF
  chmod +x "$STUB_DIR/gh"
  rm -f "$STUB_LOG"
  commit_marker proof/pr-stub
  echo '$ PATH=<recording gh>:$PATH star pr --push --title "chore: stubbed pull request" --body "Recorded, not sent."'
  PATH="$STUB_DIR:$PATH" "$STAR" pr --push \
    --title "chore: stubbed pull request" --body "Recorded, not sent." 2>&1
  echo "exit=$?"
  echo
  echo "\$ cat <outdir>/pr-gh-stub-args.txt      # the exact gh invocation, one line of arguments"
  cat "$STUB_LOG"
  echo
  echo "the branch really is on the remote now:"
  remote_branches
  echo

  echo "## Against the real GitHub CLI"
  echo
  commit_marker proof/pr-creation
  echo '$ star pr --push --title "chore: exercise the pull-request command"'
  OUT_TEXT=$("$STAR" pr --push \
    --title "chore: exercise the pull-request command" \
    --body "Temporary pull request created by the orbs proof; it is closed by the same script." 2>&1)
  STATUS=$?
  echo "$OUT_TEXT" | sed -E 's#/pull/[0-9]+#/pull/<number>#g'
  echo "exit=$STATUS"
  NUMBER=$(printf '%s' "$OUT_TEXT" | grep -oE '/pull/[0-9]+' | head -1 | tr -dc '0-9')
  if [ -n "$NUMBER" ]; then
    echo
    echo "\$ gh pr view $NUMBER --json state,isDraft,headRefName,baseRefName"
    gh pr view "$NUMBER" --repo GenLabsAI/star --json state,isDraft,headRefName,baseRefName 2>&1
    echo
    echo "\$ gh pr close $NUMBER --delete-branch    # nothing is left open"
    gh pr close "$NUMBER" --repo GenLabsAI/star --delete-branch 2>&1 \
      | sed -E 's#/pull/[0-9]+#/pull/<number>#g'
  else
    echo
    echo "# GitHub refused the create call with the credential this machine has;"
    echo "# the attempt above is the real API answer, not a local check."
  fi

  echo
  echo "## The tree afterwards"
  echo
  git -C "$REPO" checkout -q "$BASE" 2>/dev/null || true
  for branch in proof/pr-unsent proof/pr-stub proof/pr-creation; do
    git -C "$REPO" branch -D "$branch" > /dev/null 2>&1 || true
    git -C "$REPO" push -q origin --delete "$branch" > /dev/null 2>&1 || true
  done
  echo "\$ git -C $REPO branch --list 'proof/*'     # throwaway branches are gone"
  git -C "$REPO" branch --list 'proof/*'
  echo "\$ git -C $REPO ls-remote --heads origin 'refs/heads/proof/*'"
  remote_branches
  echo "\$ git -C $REPO status --short"
  git -C "$REPO" status --short
  echo "(empty output above means no leftover branches and a clean tree)"
  echo "branch: $(git -C "$REPO" rev-parse --abbrev-ref HEAD)"
} > "$OUT/19-pull-request.txt" 2>&1
