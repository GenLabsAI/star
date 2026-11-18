#!/bin/bash
# Proof for item 16 (autonomous verification): `star verify --repeat N`
# catches a test that passes and fails between runs, which a single run
# either hides or reports as a plain failure.
#
#   ./artifacts/orbs-proof/make-flake-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"
if ! command -v go >/dev/null 2>&1; then
  echo "go is required on PATH for this proof (it builds the scratch project)" >&2
  exit 1
fi

PROJ=/tmp/orbs-flake-proj
BUNDLES=/tmp/orbs-flake-bundles
COUNTER=/tmp/orbs-flake-counter
rm -rf "$PROJ" "$BUNDLES"
mkdir -p "$PROJ" "$BUNDLES"
cd "$PROJ" || exit 1

printf 'module example.com/flakeproof\n\ngo 1.21\n' > go.mod
cat > sum.go <<'EOF'
package flakeproof

// Sum adds two integers.
func Sum(a, b int) int { return a + b }
EOF
cat > sum_test.go <<'EOF'
package flakeproof

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestSum fails on exactly its second execution in a counter file. One run
// sees a pass, three runs see pass, fail, pass: a flaky test.
func TestSum(t *testing.T) {
	if counter := os.Getenv("FLAKE_COUNTER"); counter != "" {
		count := 0
		if data, err := os.ReadFile(counter); err == nil {
			count, _ = strconv.Atoi(strings.TrimSpace(string(data)))
		}
		count++
		_ = os.WriteFile(counter, []byte(strconv.Itoa(count)), 0o644)
		if count == 2 {
			t.Fatalf("second execution failed: Sum(2, 2) = %d", Sum(2, 2))
		}
	}
	if got := Sum(2, 2); got != 4 {
		t.Fatalf("Sum(2, 2) = %d, want 4", got)
	}
}
EOF

DETECTED=$(cd "$PROJ" && "$STAR" env --json | python3 -c "import json,sys
d = json.load(sys.stdin)
print('; '.join(s['name'] + ': ' + c['kind'] + '=' + c['command'] for s in d['project']['stacks'] for c in s['commands']) or 'nothing')")

{
  echo "# scratch Go project: $PROJ"
  echo "# its test fails on exactly its second execution, so one run cannot see it"
  echo "# commands star env detects here: $DETECTED"
  echo "# --repeat also sets GOFLAGS=-count=1, or Go's test cache would replay the pass"
  echo
  echo '# a single run passes: the flake is invisible'
  rm -f "$COUNTER"
  echo '$ star verify'
  FLAKE_COUNTER="$COUNTER" "$STAR" verify 2>&1
  echo "exit=$?"
  echo
  echo '# three runs catch it, and the run exits non-zero'
  rm -f "$COUNTER"
  echo '$ star verify --repeat 3 --out /tmp/orbs-flake-bundles/flaky'
  FLAKE_COUNTER="$COUNTER" "$STAR" verify --repeat 3 --out "$BUNDLES/flaky" 2>&1
  echo "exit=$?"
  echo
  echo '# the bundle keeps every attempt, not just the last one'
  echo '$ cat /tmp/orbs-flake-bundles/flaky/summary.md'
  cat "$BUNDLES/flaky/summary.md"
  echo '$ ls /tmp/orbs-flake-bundles/flaky/logs'
  ls "$BUNDLES/flaky/logs"
  echo '$ head -4 /tmp/orbs-flake-bundles/flaky/logs/02-test-run2.log   # the failing attempt'
  head -4 "$BUNDLES/flaky/logs/02-test-run2.log"
  echo
  echo '# machine-readable assertions on the record'
  rm -f "$COUNTER"
  FLAKE_COUNTER="$COUNTER" "$STAR" verify --repeat 3 --json --out "$BUNDLES/flaky-json" > /tmp/orbs-flake-record.json 2>/dev/null
  echo "json run exit=$?"
  python3 - <<'PY'
import json, os
record = json.load(open("/tmp/orbs-flake-record.json"))
assert record["verdict"] == "flaky", record["verdict"]
test = [step for step in record["steps"] if step["kind"] == "test"][0]
assert [attempt["exit_code"] for attempt in test["attempts"]] == [0, 1, 0], test["attempts"]
assert test["flaky"] is True, test
log = os.path.join("/tmp/orbs-flake-bundles/flaky-json", test["attempts"][1]["log"])
assert os.path.exists(log), log
print("verdict:", record["verdict"])
print("attempts:", [attempt["exit_code"] for attempt in test["attempts"]], "| failing log exists:", os.path.exists(log))
PY
  echo "exit=$?"
  echo
  echo '# fix the flake and the repeated run is a clean pass'
  rm -f "$COUNTER"
  cat > sum_test.go <<'EOF'
package flakeproof

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(2, 2); got != 4 {
		t.Fatalf("Sum(2, 2) = %d, want 4", got)
	}
}
EOF
  echo '$ star verify --repeat 3 --out /tmp/orbs-flake-bundles/stable'
  "$STAR" verify --repeat 3 --out "$BUNDLES/stable" 2>&1
  echo "exit=$?"
  echo
  echo '# a test that fails every time stays a failure, not a flake'
  printf 'package flakeproof\n\n// Sum adds two integers.\nfunc Sum(a, b int) int { return a + b + 1 }\n' > sum.go
  echo '$ star verify --repeat 2 --out /tmp/orbs-flake-bundles/broken'
  "$STAR" verify --repeat 2 --out "$BUNDLES/broken" 2>&1
  echo "exit=$?"
  "$STAR" verify --repeat 2 --json --out "$BUNDLES/broken-json" > /tmp/orbs-flake-broken.json 2>/dev/null
  python3 - <<'PY'
import json
record = json.load(open("/tmp/orbs-flake-broken.json"))
test = [step for step in record["steps"] if step["kind"] == "test"][0]
assert record["verdict"] == "fail", record["verdict"]
assert test.get("flaky") is not True, test
assert [attempt["exit_code"] for attempt in test["attempts"]] == [1, 1], test["attempts"]
print("verdict:", record["verdict"], "| attempts:", [attempt["exit_code"] for attempt in test["attempts"]], "| flaky: no")
PY
  echo "exit=$?"
} > "$OUT/11-flaky-verification.txt" 2>&1
