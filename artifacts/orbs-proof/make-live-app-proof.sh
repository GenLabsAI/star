#!/bin/bash
# Proof for item 16 (verification against the real running application).
#
# A scratch Go service with a real bug answers 500. Its build and tests are
# green, so a test-suite-only verification would call it healthy; `star
# verify --app` starts the service, probes it, reports LIVE-FAIL with the
# status it answered and the panic from the service's own log, keeps that
# log in the evidence bundle, and leaves nothing running. After the bug is
# fixed, the same command passes with a 200 and the real response body.
#
#   ./artifacts/orbs-proof/make-live-app-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

export PATH="$HOME/go/bin:$HOME/sdk/go/bin:$PATH"
export GOCACHE="${GOCACHE:-/tmp/gocache}"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/^go//')

PROJ=/tmp/orbs-live-proj
PORT=18111
URL="http://127.0.0.1:$PORT/items"
BUNDLES=/tmp/orbs-live-bundles
RUN_JSON=/tmp/orbs-live.json
RUN_ERR=/tmp/orbs-live.err

rm -rf "$PROJ" "$BUNDLES"
mkdir -p "$PROJ"
cd "$PROJ" || exit 1

cat > go.mod <<EOF
module example.com/liveapp

go $GO_VERSION
EOF

cat > app.go <<'EOF'
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type app struct {
	items map[string]int
}

func (a *app) handleItems(w http.ResponseWriter, r *http.Request) {
	// BUG: a nil map has no storage, so the increment panics. The handler
	// recovers and answers 500, as a real service would, while every test
	// below still passes.
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("items handler failed: %v", recovered)
			http.Error(w, "items handler failed", http.StatusInternalServerError)
		}
	}()
	a.items[r.URL.Query().Get("id")]++
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"count\":%d}\n", len(a.items))
}

func (a *app) handleHealth(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "ok")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	a := &app{}
	mux := http.NewServeMux()
	mux.HandleFunc("/items", a.handleItems)
	mux.HandleFunc("/health", a.handleHealth)
	fmt.Println("listening on " + port)
	if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
EOF

cat > app_test.go <<'EOF'
package main

import "testing"

func TestServiceStarts(t *testing.T) {
	// A test that exercises none of the request path: it passes whether or
	// not the handler works, which is exactly the gap a live run closes.
	var a app
	if a.items != nil {
		t.Fatal("a zero app has no items yet")
	}
}
EOF

# run_live LABEL BUNDLE prints the evidence that matters, and nothing that
# changes between runs (durations, timestamps, token values).
run_live() {
  local label="$1" bundle="$2"
  echo "\$ star verify --app 'PORT=$PORT go run .' --app-url $URL --app-expect 200"
  "$STAR" verify --app "PORT=$PORT go run ." \
    --app-url "$URL" --app-expect 200 --app-ready-timeout 20s \
    --out "$BUNDLES/$bundle" --json > "$RUN_JSON" 2> "$RUN_ERR"
  echo "exit=$?"
  python3 - "$label" "$BUNDLES/$bundle" <<'PY'
import json, sys, os
label, bundle = sys.argv[1], sys.argv[2]
try:
    record = json.load(open("/tmp/orbs-live.json", encoding="utf-8"))
except Exception:
    print("no JSON record; stderr:")
    print(open("/tmp/orbs-live.err", encoding="utf-8").read().strip())
    raise SystemExit(1)
print("verdict:", record["verdict"])
for step in record.get("steps", []):
    print(f"  step {step['kind']}: exit {step['exit_code']}")
live = record.get("live") or {}
print("live.command:", live.get("command"))
print("live.url:", live.get("url"))
print("live.expected_status:", live.get("expected_status"))
print("live.ready:", live.get("ready"))
print("live.status_code:", live.get("status_code"))
print("live.passed:", live.get("passed"))
print("live.reason:", live.get("reason") or "(none)")
first = [l for l in (live.get("output_tail") or "").splitlines() if l.strip()][:1]
print("app log, first line:", first[0] if first else "(none)")
body = (live.get("body_tail") or "").splitlines()[:1]
print("response body:", body[0] if body else "(empty)")
print("bundle files:", ", ".join(sorted(os.listdir(bundle))))
PY
}

{
  echo "# scratch service: $PROJ (Go, one handler, one trivial test)"
  echo "# service: PORT=$PORT go run .   probe: GET $URL, expected 200"
  echo

  echo "## The bug: green tests, broken application"
  echo
  echo '$ go test ./...     # what a test-suite-only verification sees'
  go test ./... 2>&1 | tail -2 | sed -E 's/[0-9.]+s$/TIME/'
  echo "exit=${PIPESTATUS[0]}"
  echo
  run_live "buggy" "buggy"
  echo
  echo '# the service log is kept in the bundle, next to the record'
  echo "\$ head -5 $BUNDLES/buggy/logs/90-app.log"
  head -5 "$BUNDLES/buggy/logs/90-app.log"
  echo
  echo "# the service's own log lines (shown without log.Printf's timestamp)"
  grep -E 'listening|items handler failed' "$BUNDLES/buggy/logs/90-app.log" \
    | sed -E 's/^[0-9]{4}\/[0-9]{2}\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} //'
  echo
  echo '# summary.md carries the same failure for a reviewer who reads prose'
  grep -E '^\- (Result|Reason|Body):' "$BUNDLES/buggy/summary.md"
  echo
  echo '# the failure is real, and the process is gone: nothing answers that port'
  curl -s -o /dev/null --max-time 2 "$URL"
  echo "curl after the run -> exit $? (7 = connection refused)"
  echo "\$ star ports --all | grep $PORT"
  if "$STAR" ports --all | grep -q " $PORT "; then
    echo "port $PORT is still listed (unexpected)"
  else
    echo "no listener on port $PORT after the run"
  fi
  echo

  echo "## The fix"
  echo
  echo '$ sed -i "s/a := &app{}/a := \&app{items: map[string]int{}}/" app.go'
  sed -i 's/a := &app{}/a := \&app{items: map[string]int{}}/' app.go
  grep -n 'a := &app{' app.go
  echo
  run_live "fixed" "fixed"
  echo
  echo '# the recorded body is the real response, not a health-check echo'
  grep -E '^\- (Result|Reason|Body):' "$BUNDLES/fixed/summary.md"
  echo
  echo '# and again nothing is left listening'
  curl -s -o /dev/null --max-time 2 "$URL"
  echo "curl after the run -> exit $? (7 = connection refused)"
} > "$OUT/16-live-application.txt" 2>&1
