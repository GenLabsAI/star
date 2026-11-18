#!/bin/bash
# Runs the item 21 (commit) and item 16 (flakes) proofs against a RELEASED
# star binary, so those commands are tied to a shipped artifact that
# reports its own version and commit.
#
#   ./artifacts/orbs-proof/make-release-commit-flake-proof.sh /path/to/released/star-core OUTDIR
set -u
STAR="${1:?path to the released star-core}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
HERE="$(cd "$(dirname "$0")" && pwd)"
RUN=$(mktemp -d)

{
  echo "# binary: $STAR"
  echo '$ star env --json        # the artifact reports its own version and commit'
  "$STAR" env --json | python3 -c '
import json, sys
star = json.load(sys.stdin)["star"]
print("version:", star["version"])
print("commit: ", star["commit"])
'
  echo
  for proof in commit flake; do
    echo "=== item proof: $proof, run with the released binary ==="
    if bash "$HERE/make-$proof-proof.sh" "$STAR" "$RUN"; then
      grep -hE "commits:|record hash|nothing was pushed|Verification:|runs:|verdict:|attempts:|flaky: no" "$RUN"/1*.txt
    else
      echo "FAILED: make-$proof-proof.sh $proof"
    fi
    rm -f "$RUN"/1*.txt
    echo
  done
} > "$OUT/12-released-commit-flake.txt" 2>&1
