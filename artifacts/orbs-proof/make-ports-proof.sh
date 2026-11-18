#!/bin/bash
# Proof for item 7 (local port/process inventory): a real listener started
# here is reported with its pid/port, a listener outside the workspace is
# filtered out, and the entry disappears once the process is killed.
#
#   ./artifacts/orbs-proof/make-ports-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

PROJ=/tmp/orbs-ports-proj
OUTSIDE=/tmp/orbs-ports-outside
PORT=18765
OUTSIDE_PORT=18766
rm -rf "$PROJ" "$OUTSIDE"
mkdir -p "$PROJ" "$OUTSIDE"

cleanup() {
  for pid in ${PROJ_PID:-} ${OUTSIDE_PID:-}; do
    kill "$pid" 2>/dev/null
  done
}
trap cleanup EXIT

wait_for_port() {
  for _ in $(seq 1 50); do
    if curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$1/"; then
      return 0
    fi
    if [ -n "${2:-}" ] && [ -s "$2" ] && grep -qi "Address already in use\|Traceback" "$2"; then
      return 1
    fi
    sleep 0.1
  done
  return 1
}

(
  cd "$PROJ" || exit 1
  exec python3 -m http.server "$PORT" --bind 127.0.0.1
) > /tmp/orbs-ports-server.log 2>&1 &
PROJ_PID=$!
if ! wait_for_port "$PORT" /tmp/orbs-ports-server.log; then
  echo "could not start the scratch server on port $PORT:" >&2
  cat /tmp/orbs-ports-server.log >&2
  exit 1
fi

(
  cd "$OUTSIDE" || exit 1
  exec python3 -m http.server "$OUTSIDE_PORT" --bind 127.0.0.1
) > /tmp/orbs-ports-outside.log 2>&1 &
OUTSIDE_PID=$!
if ! wait_for_port "$OUTSIDE_PORT" /tmp/orbs-ports-outside.log; then
  echo "could not start the outside server on port $OUTSIDE_PORT:" >&2
  cat /tmp/orbs-ports-outside.log >&2
  exit 1
fi

cd "$PROJ" || exit 1

{
  echo "# scratch workspace: $PROJ"
  echo "# a python http.server is listening on $PORT in this directory"
  echo "# another one is listening on $OUTSIDE_PORT in $OUTSIDE (outside the workspace)"
  echo
  echo '$ star ports        # only listeners whose process lives in this workspace'
  "$STAR" ports
  echo "exit=$?"
  echo
  echo '$ star ports --json'
  "$STAR" ports --json
  echo "exit=$?"
  echo
  echo '# machine-readable assertions, checked with python3 rather than by eye'
  "$STAR" ports --json | python3 -c '
import json, sys
data = json.load(sys.stdin)
listeners = data["listeners"]
expected = {
    "workspace": "'"$PROJ"'",
    "filtered": True,
}
for key, value in expected.items():
    assert data[key] == value, (key, data[key], value)
ports = [listener["port"] for listener in listeners]
assert ports == ['"$PORT"'], ports
listener = listeners[0]
assert listener["workspace"] is True, listener
assert listener["cwd"] == "'"$PROJ"'", listener
assert listener["process"] == "python3", listener
assert listener["pid"] > 0, listener
print("workspace filter: only port", listener["port"], "from pid", listener["pid"], "in", listener["cwd"])
print("outside listener on '"$OUTSIDE_PORT"' correctly hidden")
open("/tmp/orbs-ports-pid", "w").write(str(listener["pid"]))
'
  echo "exit=$?"
  echo
  echo '$ star ports --all | grep -E "PORT|<the two scratch ports>"'
  "$STAR" ports --all | grep -E "PORT|$PORT|$OUTSIDE_PORT"
  echo "exit=$?"
  echo
  PID=$(cat /tmp/orbs-ports-pid)
  echo "# kill the reported pid ($PID) and look again"
  echo "\$ kill $PID"
  kill "$PID"
  echo "exit=$?"
  sleep 0.5
  echo '$ star ports'
  "$STAR" ports
  echo "exit=$?"
  echo '$ star ports --json'
  "$STAR" ports --json
  echo "exit=$?"
  "$STAR" ports --json | python3 -c '
import json, sys
listeners = json.load(sys.stdin)["listeners"]
assert [listener["port"] for listener in listeners] == [], listeners
print("no listeners remain: the killed process is gone from the inventory")
'
  echo "exit=$?"
  echo
  echo '# process-lookup failure case: nothing in this workspace is listening'
  rm -rf "$PROJ"
  mkdir -p "$PROJ"
  cd "$PROJ" || exit 1
  echo '$ star ports'
  "$STAR" ports
  echo "exit=$?"
} > "$OUT/07-ports.txt" 2>&1
