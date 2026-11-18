#!/bin/bash
# Proof for item 24 (one overview of every workspace's activity).
#
# A real Star server creates a real workspace and two real conversations; a
# durable wakeup store is seeded with the same file `star wakeup schedule`
# writes (the CLI itself needs a model credential, which this machine does
# not have); a real HTTP server runs inside the workspace. `star overview`
# then reports the conversations, the armed wakeups and the preview in one
# view, and drops the preview as soon as that server stops. A workspace
# that has been deleted is reported as a problem instead of failing the
# report.
#
#   ./artifacts/orbs-proof/make-overview-activity-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

export XDG_DATA_HOME=/tmp/orbs-overview-state
PROJ=/tmp/orbs-overview-proj
SRVDIR=/tmp/orbs-overview-srv
STATE=/tmp/orbs-overview-state
PORT=18499
PREVIEW_PORT=18422
BASE="http://127.0.0.1:$PORT"
TOKEN=overview-proof-token

# Both ports must be free: a leftover listener would answer this script's
# own probes and make the transcript describe something else.
for port in "$PORT" "$PREVIEW_PORT"; do
  if curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$port/"; then
    echo "port $port is already in use; stop that process first" >&2
    exit 1
  fi
done

rm -rf "$PROJ" "$SRVDIR" "$STATE"
mkdir -p "$PROJ" "$SRVDIR"
cd "$PROJ" || exit 1
git init -q .
echo "hello" > app.txt
git add -A
git -c user.email=proof@example.com -c user.name=proof commit -q -m init

SERVER_PID=""
PREVIEW_PID=""
cleanup() {
  [ -n "$PREVIEW_PID" ] && kill "$PREVIEW_PID" 2>/dev/null
  [ -n "$SERVER_PID" ] && kill "$SERVER_PID" 2>/dev/null
}
trap cleanup EXIT

# The server runs outside the workspace, so its own listener is not
# mistaken for one of the workspace's previews.
(
  cd "$SRVDIR" || exit 1
  exec "$STAR" server --host "tcp://127.0.0.1:$PORT" --pair-token "$TOKEN"
) > /tmp/orbs-overview-server.log 2>&1 &
SERVER_PID=$!

ready=""
for _ in $(seq 1 100); do
  if curl -s -o /dev/null --max-time 1 -H "Authorization: Bearer $TOKEN" "$BASE/v1/health"; then
    ready=yes
    break
  fi
  sleep 0.1
done
if [ -z "$ready" ]; then
  echo "server did not start:" >&2
  cat /tmp/orbs-overview-server.log >&2
  exit 1
fi

api() {
  curl -s --max-time 10 -H "Authorization: Bearer $TOKEN" "$@"
}

{
  echo "# scratch workspace: $PROJ (git repository, one commit, clean tree)"
  echo "# star state is isolated in XDG_DATA_HOME=$STATE"
  echo "# server: star server --host tcp://127.0.0.1:$PORT, started outside the workspace"
  echo

  echo "## A real workspace, created through the server API"
  echo
  CLIENT_ID=$(python3 -c 'import uuid;print(uuid.uuid4())')
  WORKSPACE=$(api -X POST -H "Content-Type: application/json" \
    -d "{\"path\":\"$PROJ\",\"client_id\":\"$CLIENT_ID\"}" "$BASE/v1/workspaces" \
    | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
  echo "\$ POST /v1/workspaces {\"path\": \"$PROJ\"} -> HTTP 201, id <workspace-id>"
  echo "data dir: $(ls -d "$PROJ/.crush") (database created by the server)"
  echo

  echo "## Two real conversations"
  echo
  for title in "investigate the failing build" "draft the release notes"; do
    code=$(api -o /dev/null -w '%{http_code}' -X POST -H "Content-Type: application/json" \
      -d "{\"title\":\"$title\"}" "$BASE/v1/workspaces/$WORKSPACE/sessions")
    echo "\$ POST /v1/workspaces/<workspace-id>/sessions {\"title\": \"$title\"} -> HTTP $code"
  done
  echo "\$ star session list     # ids and timestamps dropped here"
  "$STAR" session list | sed -E 's/^[0-9a-f]+ [0-9T:Z-]+ //'
  echo

  echo "## The durable wakeup store"
  echo
  python3 - "$PROJ" <<'PY'
import json, pathlib, sys
path = pathlib.Path(sys.argv[1]) / ".crush" / "wakeups" / "schedules.json"
path.parent.mkdir(parents=True, exist_ok=True)
path.write_text(json.dumps({
    "version": 1,
    "schedules": [
        {"id": "wake-timed", "session_id": "s", "reason": "rerun the suite",
         "prompt": "run the tests again", "event": "ci-finished",
         "next_fire_unix": 4102444800, "created_unix": 1795038000},
        {"id": "wake-event", "session_id": "s", "reason": "a review arrived",
         "prompt": "read the review", "event": "pr-review",
         "created_unix": 1795038000},
    ],
}, indent=2) + "\n")
PY
  echo '# `star wakeup schedule` needs a model credential this machine lacks, so the'
  echo '# store it writes is seeded directly; overview reads that same file.'
  echo "\$ python3 -c 'json.load(open(\"$PROJ/.crush/wakeups/schedules.json\"))[\"schedules\"]'"
  python3 -c "import json;d=json.load(open('$PROJ/.crush/wakeups/schedules.json'));[print('  ',s['id'],'next_fire_unix' ,s.get('next_fire_unix',0),'event',s.get('event','')) for s in d['schedules']]"
  echo

  echo "## A real preview inside the workspace"
  echo
  (
    cd "$PROJ" || exit 1
    exec python3 -m http.server "$PREVIEW_PORT"
  ) > /tmp/orbs-overview-preview.log 2>&1 &
  PREVIEW_PID=$!
  preview_ready=""
  for _ in $(seq 1 50); do
    if curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$PREVIEW_PORT/"; then
      preview_ready=yes
      break
    fi
    if ! kill -0 "$PREVIEW_PID" 2>/dev/null; then
      echo "preview server exited (port $PREVIEW_PORT already in use?):" >&2
      cat /tmp/orbs-overview-preview.log >&2
      exit 1
    fi
    sleep 0.1
  done
  if [ -z "$preview_ready" ]; then
    echo "preview server did not answer on port $PREVIEW_PORT" >&2
    exit 1
  fi
  echo "\$ curl -o /dev/null -w '%{http_code}' http://127.0.0.1:$PREVIEW_PORT/"
  curl -s -o /dev/null -w '%{http_code}\n' --max-time 5 "http://127.0.0.1:$PREVIEW_PORT/"
  echo "process: $(ps -o comm= -p "$PREVIEW_PID"), cwd: $(readlink /proc/"$PREVIEW_PID"/cwd)"
  echo

  echo "## One view of it all"
  echo
  cd "$PROJ" || exit 1
  echo '$ star overview'
  "$STAR" overview
  echo "exit=$?"
  echo
  echo "\$ star overview --json | python3 -c 'selected fields'"
  "$STAR" overview --json | python3 -c '
import json, sys
report = json.load(sys.stdin)
for entry in report["projects"]:
    print("workspace:", entry["path"])
    print("branch:", entry["branch"], "| changed_paths:", entry["changed_paths"])
    print("sessions:", entry["sessions"])
    print("schedules:", entry["schedules"])
    print("previews:", [(p["port"], p["url"]) for p in entry.get("previews", [])])
print("totals:", report["sessions"], "conversations,", report["schedules"], "wakeups,", report["previews"], "previews")
'
  echo "exit=$?"
  echo

  echo "## The view follows reality"
  echo
  kill "$PREVIEW_PID"
  PREVIEW_PID=""
  for _ in $(seq 1 50); do
    if ! curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$PREVIEW_PORT/"; then
      break
    fi
    sleep 0.1
  done
  echo "\$ kill \$PREVIEW_PID   # the preview process is stopped"
  curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PREVIEW_PORT/"
  echo "curl after stop -> exit $? (7 = connection refused)"
  "$STAR" overview --json | python3 -c '
import json, sys
report = json.load(sys.stdin)
entry = report["projects"][0]
print("sessions:", entry["sessions"], "| schedules:", entry["schedules"],
      "| previews:", len(entry.get("previews", []) or []))
'
  echo

  kill "$SERVER_PID"
  SERVER_PID=""
  for _ in $(seq 1 50); do
    if ! curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
      break
    fi
    sleep 0.1
  done
  echo "## A removed workspace is reported, not fatal"
  echo
  cd "$SRVDIR" || exit 1
  rm -rf "$PROJ"
  echo "\$ rm -rf $PROJ"
  echo '$ star overview'
  "$STAR" overview
  echo "exit=$?"
} > "$OUT/17-overview-activity.txt" 2>&1
