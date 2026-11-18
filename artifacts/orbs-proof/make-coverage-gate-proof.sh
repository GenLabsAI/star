#!/bin/bash
# Proof for item 16 (coverage gate): `star verify --min-coverage` refuses a
# passing-but-under-covered run, passes a covered one, and reports out loud
# when no coverage number can be measured instead of passing silently.
#
#   ./artifacts/orbs-proof/make-coverage-gate-proof.sh /path/to/star OUTDIR
#
# Deterministic by construction: every scenario uses a fixed --out bundle and
# reports selected JSON fields, never wall-clock durations or timestamps, so
# re-running it from a clean rebuild produces the same bytes.
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
if ! command -v go >/dev/null 2>&1; then
  echo "go is required on PATH for this proof (it builds the scratch projects)" >&2
  exit 1
fi

PROJ=/tmp/orbs-coverage-proj
NODE=/tmp/orbs-coverage-node
BUNDLES=/tmp/orbs-coverage-bundles
TMP=/tmp/orbs-coverage-tmp
rm -rf "$PROJ" "$NODE" "$BUNDLES" "$TMP"
mkdir -p "$PROJ" "$BUNDLES" "$TMP"
cd "$PROJ" || exit 1

# calc.go has three one-statement functions; the test only calls Add, so the
# real coverage of this project is 1 of 3 statements.
printf 'module example.com/coverageproof\n\ngo 1.21\n' > go.mod
cat > calc.go <<'EOF'
package coverageproof

// Add adds two integers.
func Add(a, b int) int { return a + b }

// Sub subtracts b from a. Nothing tests it yet.
func Sub(a, b int) int { return a - b }

// Mul multiplies two integers. Nothing tests it yet.
func Mul(a, b int) int { return a * b }
EOF
cat > calc_test.go <<'EOF'
package coverageproof

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add(2, 3) = %d, want 5", got)
	}
}
EOF

# report prints the parts of a `star verify --json` record this proof is
# about, in a stable order.
report() {
  python3 - "$1" <<'PY'
import json, sys
record = json.load(open(sys.argv[1]))
steps = ", ".join(f"{step['kind']}={step['exit_code']}" for step in record["steps"])
print("verdict:", record["verdict"], "| steps:", steps)
coverage = record.get("coverage")
if coverage is None:
    print("coverage: absent (no gate requested)")
elif not coverage["available"]:
    print("coverage: unavailable ->", coverage["reason"])
else:
    print("coverage: %.1f%% (%d of %d statements), required %.1f%%, passed=%s" % (
        coverage["percent"], coverage["covered"], coverage["statements"],
        coverage["threshold"], coverage["passed"]))
PY
}

{
  echo "# scratch Go project: $PROJ"
  echo "# calc.go has 3 one-statement functions; calc_test.go only calls Add"
  echo
  echo "# the project's own toolchain agrees on the number the gate must see"
  echo '$ go test -coverprofile=/tmp/orbs-coverage-tmp/cover.out ./... && go tool cover -func=/tmp/orbs-coverage-tmp/cover.out | tail -1'
  go test -coverprofile="$TMP/cover.out" ./... >/dev/null && go tool cover -func="$TMP/cover.out" | tail -1
  echo "exit=$?"
  echo
  echo '# a passing run below the threshold exits non-zero and names the number'
  echo '$ star verify --min-coverage 90 --json --out /tmp/orbs-coverage-bundles/under'
  "$STAR" verify --min-coverage 90 --json --out "$BUNDLES/under" > "$TMP/under.json" 2> "$TMP/under.err"
  echo "exit=$?"
  cat "$TMP/under.err"
  report "$TMP/under.json"
  echo
  echo '# the same measured number, in the bundle a reviewer keeps'
  echo '$ sed -n "/## Coverage/,/## Changes/p" /tmp/orbs-coverage-bundles/under/summary.md'
  sed -n '/## Coverage/,/## Changes/p' "$BUNDLES/under/summary.md"
  echo
  echo '# the same project, a threshold it does meet: the run passes'
  echo '$ star verify --min-coverage 20 --json --out /tmp/orbs-coverage-bundles/over'
  "$STAR" verify --min-coverage 20 --json --out "$BUNDLES/over" > "$TMP/over.json" 2> "$TMP/over.err"
  echo "exit=$?"
  report "$TMP/over.json"
  echo
  echo '# below a threshold the measurement still reports the real number, not the threshold'
  echo '$ star verify --min-coverage 100 --json --out /tmp/orbs-coverage-bundles/all'
  "$STAR" verify --min-coverage 100 --json --out "$BUNDLES/all" > "$TMP/all.json" 2> "$TMP/all.err"
  echo "exit=$?"
  report "$TMP/all.json"
  echo
  echo '# without the flag nothing changes: no coverage gate, same pass as before'
  echo '$ star verify --json --out /tmp/orbs-coverage-bundles/plain'
  "$STAR" verify --json --out "$BUNDLES/plain" > "$TMP/plain.json" 2> "$TMP/plain.err"
  echo "exit=$?"
  report "$TMP/plain.json"
  echo
  echo '# a failing test stays a failure: the gate never relabels it as coverage'
  printf 'package coverageproof\n\n// Add adds two integers. It is wrong on purpose here.\nfunc Add(a, b int) int { return a + b + 1 }\n\n// Sub subtracts b from a.\nfunc Sub(a, b int) int { return a - b }\n\n// Mul multiplies two integers.\nfunc Mul(a, b int) int { return a * b }\n' > calc.go
  echo '$ star verify --min-coverage 90 --json --out /tmp/orbs-coverage-bundles/broken'
  "$STAR" verify --min-coverage 90 --json --out "$BUNDLES/broken" > "$TMP/broken.json" 2> "$TMP/broken.err"
  echo "exit=$?"
  report "$TMP/broken.json"
  echo
  echo '# restore the correct code before the next scenario'
  cat > calc.go <<'EOF'
package coverageproof

// Add adds two integers.
func Add(a, b int) int { return a + b }

// Sub subtracts b from a. Nothing tests it yet.
func Sub(a, b int) int { return a - b }

// Mul multiplies two integers. Nothing tests it yet.
func Mul(a, b int) int { return a * b }
EOF
  echo '# a project with no coverage provider says so, instead of passing or guessing'
  mkdir -p "$NODE"
  printf '{\n  "name": "coverage-node-probe",\n  "version": "1.0.0",\n  "scripts": { "test": "node ok.js" }\n}\n' > "$NODE/package.json"
  printf 'console.log("the node test passes and measures nothing");\n' > "$NODE/ok.js"
  echo '$ cd /tmp/orbs-coverage-node && star verify --min-coverage 80 --json --out /tmp/orbs-coverage-bundles/node'
  ( cd "$NODE" && "$STAR" verify --min-coverage 80 --json --out "$BUNDLES/node" > "$TMP/node.json" 2> "$TMP/node.err" )
  echo "exit=$?"
  cat "$TMP/node.err"
  report "$TMP/node.json"
  echo
  echo '# a flaky test under a coverage gate is still FLAKY, not quietly re-labelled'
  FLAKE=/tmp/orbs-coverage-flake
  rm -rf "$FLAKE" && mkdir -p "$FLAKE"
  printf 'module example.com/coverageflake\n\ngo 1.21\n' > "$FLAKE/go.mod"
  cat > "$FLAKE/sum.go" <<'EOF'
package coverageflake

// Sum adds two integers.
func Sum(a, b int) int { return a + b }
EOF
  cat > "$FLAKE/sum_test.go" <<'EOF'
package coverageflake

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestSum fails on exactly its second execution, so one run cannot see it.
func TestSum(t *testing.T) {
	if counter := os.Getenv("FLAKE_COUNTER"); counter != "" {
		count := 0
		if data, err := os.ReadFile(counter); err == nil {
			count, _ = strconv.Atoi(strings.TrimSpace(string(data)))
		}
		count++
		_ = os.WriteFile(counter, []byte(strconv.Itoa(count)), 0o644)
		if count == 2 {
			t.Fatalf("second execution failed")
		}
	}
}
EOF
  echo '$ cd /tmp/orbs-coverage-flake && star verify --repeat 3 --min-coverage 10 --json --out /tmp/orbs-coverage-bundles/flake'
  ( cd "$FLAKE" && rm -f /tmp/orbs-coverage-tmp/counter && FLAKE_COUNTER=/tmp/orbs-coverage-tmp/counter "$STAR" verify --repeat 3 --min-coverage 10 --json --out "$BUNDLES/flake" > "$TMP/flake.json" 2> "$TMP/flake.err" )
  echo "exit=$?"
  report "$TMP/flake.json"
  python3 - "$TMP/flake.json" <<'PY'
import json, sys
record = json.load(open(sys.argv[1]))
test = [step for step in record["steps"] if step["kind"] == "test"][0]
print("attempt exit codes:", [attempt["exit_code"] for attempt in test["attempts"]], "| flaky:", test["flaky"])
PY
  echo
  echo '# an impossible threshold is rejected instead of being treated as 0'
  echo '$ star verify --min-coverage 120'
  "$STAR" verify --min-coverage 120 2>&1
  echo "exit=$?"
} > "$OUT/13-coverage-gate.txt" 2>&1
