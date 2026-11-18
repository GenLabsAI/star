#!/bin/bash
# Same scenarios as make-proof.sh, run against a RELEASED star binary, so
# the transcript ties a published version to the behaviour it claims.
#
#   ./artifacts/orbs-proof/make-release-proof.sh /path/to/released/star-core OUTDIR
set -u
STAR="${1:?path to the released star-core}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
export PATH="$HOME/sdk/go/bin:$PATH"

PROJ=/tmp/orbs-release-proj
STATE=/tmp/orbs-release-state
BUNDLES=/tmp/orbs-release-bundles
rm -rf "$PROJ" "$STATE" "$BUNDLES"
mkdir -p "$PROJ" "$STATE/crush" "$BUNDLES"
cd "$PROJ"
git init -q -b main
git config user.email proof@example.com
git config user.name "Orbs Release Proof"

printf '.crush/\n' > .gitignore
printf 'module example.com/releaseproof\n\ngo 1.27\n' > go.mod
cat > sum.go <<'EOF'
package releaseproof

// Sum adds two integers.
func Sum(a, b int) int { return a + b }
EOF
cat > sum_test.go <<'EOF'
package releaseproof

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(2, 2); got != 5 {
		t.Fatalf("Sum(2, 2) = %d, want 5", got)
	}
}
EOF
git add . && git commit -qm "feat: add Sum with a deliberately wrong test"

{
  echo "\$ $STAR env --json    # which version is being verified"
  "$STAR" env --json | python3 -c 'import json,sys;d=json.load(sys.stdin);print("version:",d["star"]["version"]);print("commit:",d["star"].get("commit",""))'
  echo "exit=$?"
  echo
  echo "\$ $STAR verify        # broken expectation, must fail"
  "$STAR" verify --out "$BUNDLES/fail" 2>&1
  echo "exit=$?"
  echo
  cat "$BUNDLES/fail/summary.md"
} > "$OUT/05-released-binary.txt" 2>&1

cat > sum_test.go <<'EOF'
package releaseproof

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(2, 2); got != 4 {
		t.Fatalf("Sum(2, 2) = %d, want 4", got)
	}
}
EOF
cat > sum.go <<'EOF'
package releaseproof

// Sum adds two integers.
func Sum(a, b int) int { return a + b }

// Double returns twice n.
func Double(n int) int { return Sum(n, n) }
EOF

{
  echo "\$ $STAR verify        # fixed expectation, must pass"
  "$STAR" verify --out "$BUNDLES/pass" 2>&1
  echo "exit=$?"
  echo
  echo "\$ $STAR verify --json | python3 -c 'summary'"
  "$STAR" verify --json --out "$BUNDLES/json" | python3 -c 'import json,sys;d=json.load(sys.stdin);print("verdict:",d["verdict"]);print("star:",d["workspace"]["star"]["version"]);print("steps:",[(s["kind"],s["exit_code"],s["log"]) for s in d["steps"]]);print("changes:",[(f["status"].strip(),f["path"]) for f in d["changes"]["files"]])'
  echo "exit=$?"
  echo
  echo "\$ $STAR changes --json | python3 -c 'summary'"
  "$STAR" changes --json | python3 -c 'import json,sys;d=json.load(sys.stdin);print("branch:",d["branch"]);print("changed:",[(f["status"].strip(),f["path"]) for f in d["files"]]);print("patch_bytes:",d["patch_bytes"])'
  echo "exit=$?"
  echo
  cat "$BUNDLES/pass/changes.patch"
} >> "$OUT/05-released-binary.txt" 2>&1

cat > "$STATE/crush/projects.json" <<EOF
{
  "projects": [
    {"path": "$PROJ", "data_dir": "$PROJ/.crush", "last_accessed": "2026-11-18T00:00:00Z"}
  ]
}
EOF

{
  echo
  echo "\$ $STAR verify        # default bundle, so the overview can see it"
  "$STAR" verify 2>&1
  echo "exit=$?"
  echo
  echo "\$ XDG_DATA_HOME=$STATE $STAR overview"
  XDG_DATA_HOME="$STATE" "$STAR" overview 2>&1
  echo "exit=$?"
  echo
  echo "\$ XDG_DATA_HOME=$STATE $STAR overview --json | python3 -c 'summary'"
  XDG_DATA_HOME="$STATE" "$STAR" overview --json | python3 -c 'import json,sys;d=json.load(sys.stdin);e=d["projects"][0];print("workspace:",e["path"]);print("branch:",e["branch"]);print("changed_paths:",e["changed_paths"]);print("verification:",e["verification"]["verdict"],"steps=",e["verification"]["steps"],"failed=",e["verification"]["failed_steps"])'
  echo "exit=$?"
} >> "$OUT/05-released-binary.txt" 2>&1
