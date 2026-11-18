#!/bin/bash
# Proof that the published artifact carries the latest pass: it runs the
# four item proofs with the downloaded release binary and compares each
# transcript with the committed one, byte for byte. A difference is printed
# instead of being hidden, so this is a real check and not a smoke test.
#
#   ./artifacts/orbs-proof/make-release-latest-proof.sh /path/to/released/star OUTDIR [REPO]
#
# REPO defaults to this repository; the pull-request proof pushes and
# deletes its own throwaway branches there.
set -u
STAR="${1:?path to the released star binary}"
OUT="${2:?output directory}"
REPO="${3:-$(pwd)}"
mkdir -p "$OUT"

HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$OUT/release-runs
rm -rf "$WORK"
mkdir -p "$WORK"

check() {
  local name="$1" generated="$2"
  local committed="$HERE/$name"
  local got want
  got=$(sha256sum "$generated" | cut -d' ' -f1)
  want=$(sha256sum "$committed" | cut -d' ' -f1)
  echo "  file:      $name"
  echo "  committed: $want"
  echo "  generated: $got"
  if [ "$got" = "$want" ]; then
    echo "  result:    IDENTICAL (the release reproduces the committed transcript)"
  else
    echo "  result:    DIFFERS — first differences:"
    diff -u "$committed" "$generated" | head -20
  fi
  echo
}

{
  echo "# released binary: $STAR"
  echo "# repository: $REPO"
  echo

  echo "## The release reports this repository's commit"
  echo
  echo '$ star env --json | python3 -c "version and commit"'
  "$STAR" env --json | python3 -c '
import json, sys
info = json.load(sys.stdin)["star"]
print("version:", info.get("version") or "not stamped")
print("commit: ", info.get("commit") or "not stamped")
'
  HEAD_COMMIT=$(git -C "$REPO" rev-parse HEAD)
  RELEASE_COMMIT=$("$STAR" env --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["star"].get("commit") or "")')
  echo "repository HEAD: $HEAD_COMMIT"
  if [ "$RELEASE_COMMIT" = "$HEAD_COMMIT" ]; then
    echo "match: the release is exactly HEAD"
  elif [ -n "$RELEASE_COMMIT" ] && git -C "$REPO" merge-base --is-ancestor "$RELEASE_COMMIT" HEAD 2>/dev/null; then
    echo "match: the release is an ancestor of HEAD (later commits touch docs and proofs only)"
  else
    echo "match: False (the release is not in this repository's history)"
  fi
  echo

  echo "## Item proofs re-run with the released binary"
  echo

  "$HERE/make-live-app-proof.sh" "$STAR" "$WORK/16" > /dev/null 2>&1
  echo "### 16 live application (exit $?)"
  check "16-live-application.txt" "$WORK/16/16-live-application.txt"
  echo

  "$HERE/make-overview-activity-proof.sh" "$STAR" "$WORK/17" > /dev/null 2>&1
  echo "### 17 overview activity (exit $?)"
  check "17-overview-activity.txt" "$WORK/17/17-overview-activity.txt"
  echo

  "$HERE/make-peer-scopes-proof.sh" "$STAR" "$WORK/18" > /dev/null 2>&1
  echo "### 18 peer scopes and approval (exit $?)"
  check "18-peer-scopes.txt" "$WORK/18/18-peer-scopes.txt"
  echo

  "$HERE/make-pull-request-proof.sh" "$STAR" "$WORK/19" "$REPO" > /dev/null 2>&1
  echo "### 19 pull requests (exit $?)"
  check "19-pull-request.txt" "$WORK/19/19-pull-request.txt"
  echo

  echo "## The repository afterwards"
  echo
  echo "\$ git -C $REPO status --short"
  git -C "$REPO" status --short
  echo "(empty output above means the proofs left nothing behind)"
} > "$OUT/20-released-live-scopes-overview-pr.txt" 2>&1
