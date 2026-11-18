#!/bin/bash
# Proof for items 16 (coverage gate) and 21 (approval gate) against the
# artifact users actually download. It records the release's own version and
# commit, then runs both scripts with that binary.
#
#   ./artifacts/orbs-proof/make-release-coverage-approval-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the released star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
if ! command -v go >/dev/null 2>&1; then
  echo "go is required on PATH for this proof (it builds the scratch projects)" >&2
  exit 1
fi

{
  echo "# released binary: $STAR"
  echo '$ star env --json   (selected fields)'
  "$STAR" env --json | python3 -c 'import json,sys
star=json.load(sys.stdin)["star"]
print("version:", star["version"])
print("commit: ", star["commit"])' 
  echo "exit=$?"
  echo
  echo '# the coverage gate, from the released binary'
  ./artifacts/orbs-proof/make-coverage-gate-proof.sh "$STAR" /tmp/orbs-release-coverage
  echo "coverage script exit=$?"
  cat /tmp/orbs-release-coverage/13-coverage-gate.txt
  echo
  echo '# the approval gate, from the released binary'
  ./artifacts/orbs-proof/make-approval-gate-proof.sh "$STAR" /tmp/orbs-release-approval
  echo "approval script exit=$?"
  cat /tmp/orbs-release-approval/14-approval-gate.txt
} > "$OUT/15-released-coverage-approval.txt" 2>&1
