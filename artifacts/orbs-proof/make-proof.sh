#!/bin/bash
# Reproducible demo of `star verify` and `star changes` against a scratch repo.
set -u
STAR="${1:?path to star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
export PATH="$HOME/sdk/go/bin:$PATH"
PROJ=/tmp/orbs-proof-proj
rm -rf "$PROJ"; mkdir -p "$PROJ"; cd "$PROJ"
git init -q -b main
git config user.email proof@example.com
git config user.name "Orbs Proof"
cat > go.mod <<'EOF'
module example.com/proof

go 1.27
EOF
cat > sum.go <<'EOF'
package proof

// Sum adds two integers.
func Sum(a, b int) int { return a + b }
EOF
cat > sum_test.go <<'EOF'
package proof

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(2, 2); got != 5 {
		t.Fatalf("Sum(2, 2) = %d, want 5", got)
	}
}
EOF
git add . && git commit -qm "feat: add Sum with a deliberately wrong test"

{
  echo "\$ star verify        # first run: the test is wrong on purpose"
  "$STAR" verify --out "$OUT/01-fail-bundle" 2>&1
  echo "exit=$?"
  echo
  echo "--- bundle contents ---"
  find "$OUT/01-fail-bundle" -type f | sed "s|$OUT/||" | sort
  echo
  echo "--- summary.md ---"
  cat "$OUT/01-fail-bundle/summary.md"
} > "$OUT/01-verify-fail.txt" 2>&1

# Fix the expectation and add a function, so the tree has a real diff.
cat > sum_test.go <<'EOF'
package proof

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(2, 2); got != 4 {
		t.Fatalf("Sum(2, 2) = %d, want 4", got)
	}
}
EOF
cat > sum.go <<'EOF'
package proof

// Sum adds two integers.
func Sum(a, b int) int { return a + b }

// Double returns twice n.
func Double(n int) int { return Sum(n, n) }
EOF

{
  echo "\$ star verify        # second run: expectation fixed, code changed"
  "$STAR" verify --out "$OUT/02-pass-bundle" 2>&1
  echo "exit=$?"
  echo
  echo "--- verify.json (verdict, steps, changes) ---"
  python3 - "$OUT/02-pass-bundle/verify.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
print("verdict:", d["verdict"])
print("star:", d["workspace"]["star"]["version"], d["workspace"]["star"]["commit"][:7] if d["workspace"]["star"].get("commit") else "-")
print("branch:", d["workspace"]["git"]["branch"])
for s in d["steps"]:
    print(f"  {s['kind']:<6} {s['command']:<18} exit={s['exit_code']} {s['duration_ms']}ms log={s['log']}")
print("changes:", [(f["status"].strip(), f["path"]) for f in d["changes"]["files"]])
PY
  echo
  echo "--- logs/02-test.log ---"
  cat "$OUT/02-pass-bundle/logs/02-test.log"
  echo
  echo "--- changes.patch ---"
  cat "$OUT/02-pass-bundle/changes.patch"
} > "$OUT/02-verify-pass.txt" 2>&1

{
  echo "\$ star changes"
  "$STAR" changes 2>&1
  echo "exit=$?"
  echo
  echo "\$ star changes --json | python3 -c 'summary'"
  "$STAR" changes --json | python3 -c 'import json,sys;d=json.load(sys.stdin);print("repository:",d["repository"]);print("branch:",d["branch"]);print("patch_bytes:",d["patch_bytes"]);print("files:",[(f["status"].strip(),f["path"]) for f in d["files"]])'
  echo "exit=$?"
  echo
  echo "\$ star changes --patch review.patch"
  "$STAR" changes --patch /tmp/orbs-review.patch 2>&1
  echo "exit=$?"
  echo
  echo "--- review.patch ---"
  cat /tmp/orbs-review.patch
} > "$OUT/03-changes.txt" 2>&1
