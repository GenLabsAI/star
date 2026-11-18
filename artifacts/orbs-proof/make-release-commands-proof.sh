#!/bin/bash
# Runs the item 14 / 7 / 20 proofs against a RELEASED star binary, so the
# commands are tied to a shipped artifact that reports its own version and
# commit.
#
#   ./artifacts/orbs-proof/make-release-commands-proof.sh /path/to/released/star-core OUTDIR
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
data = json.load(sys.stdin)
star = data["star"]
print("version:", star["version"])
print("commit: ", star["commit"])
'
  echo
  for proof in knowledge ports pairing; do
    echo "=== item proof: $proof, run with the released binary ==="
    if bash "$HERE/make-$proof-proof.sh" "$STAR" "$RUN"; then
      grep -hE "identical:|workspace filter|correctly hidden|no listeners remain|no file was written|unpaired peer ->|wrong token ->|paired peer ->|no entry for|token from the file" "$RUN"/0*.txt
    else
      echo "FAILED: make-$proof-proof.sh $proof"
    fi
    rm -f "$RUN"/0*.txt
    echo
  done
} > "$OUT/09-released-commands.txt" 2>&1
