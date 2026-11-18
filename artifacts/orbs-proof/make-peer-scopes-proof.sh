#!/bin/bash
# Proof for item 20 (per-peer scopes and an approval prompt).
#
# Three real servers, each started with a real token, answer real HTTP:
#   1. a read-only token reads but is refused every write and every exec;
#   2. a server that runs commands only with an approval code refuses exec
#      until a current code is present, and refuses an expired one;
#   3. the same server is limited to one workspace directory and refuses to
#      open any other.
# Token and approval values are never printed: only their length, their
# scope prefix, and a hash prefix.
#
#   ./artifacts/orbs-proof/make-peer-scopes-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

WORK=/tmp/orbs-scope-proof
PORT=18599
BASE="http://127.0.0.1:$PORT"
FAKE_ID=00000000-0000-0000-0000-000000000000
CLIENT_ID=11111111-1111-1111-1111-111111111111
SHELL_PATH="/v1/workspaces/$FAKE_ID/agent/sessions/$FAKE_ID/shell"

rm -rf "$WORK"
mkdir -p "$WORK/allowed" "$WORK/other" "$WORK/srv"

SERVER_PID=""
stop_server() {
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null
    for _ in $(seq 1 50); do
      curl -s -o /dev/null --max-time 1 "$BASE/v1/health" || break
      sleep 0.1
    done
    SERVER_PID=""
  fi
}
trap stop_server EXIT

if curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
  echo "port $PORT is already in use; stop that process first" >&2
  exit 1
fi

# start_server TOKEN [server args...]
start_server() {
  local token="$1"
  shift
  (
    cd "$WORK/srv" || exit 1
    exec "$STAR" server --host "tcp://127.0.0.1:$PORT" --pair-token "$token" "$@"
  ) > "$WORK/server.log" 2>&1 &
  SERVER_PID=$!
  for _ in $(seq 1 100); do
    if curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
      return 0
    fi
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      echo "server exited:" >&2
      cat "$WORK/server.log" >&2
      return 1
    fi
    sleep 0.1
  done
  echo "server did not answer on $BASE" >&2
  return 1
}

# request LABEL METHOD PATH [header...] -- prints the status and the start
# of the body, so a refusal can be read, not just counted.
request() {
  local label="$1" method="$2" path="$3"
  shift 3
  local code
  code=$(curl -s -o "$WORK/body" -w '%{http_code}' --max-time 10 \
    -X "$method" -H "Content-Type: application/json" "$@" "$BASE$path")
  echo "$label"
  echo "  -> HTTP $code"
  echo "  body: $(head -c 160 "$WORK/body" | tr -d '\n')"
}

# request_quiet LABEL METHOD PATH [header...] -- prints the status only,
# for a request whose body carries an id that changes between runs.
request_quiet() {
  local label="$1" method="$2" path="$3"
  shift 3
  local code
  code=$(curl -s -o "$WORK/body" -w '%{http_code}' --max-time 10 \
    -X "$method" -H "Content-Type: application/json" "$@" "$BASE$path")
  echo "$label"
  echo "  -> HTTP $code"
}

{
  echo "# scratch state: $WORK"
  echo "# server: star server --host tcp://127.0.0.1:$PORT (real process, real HTTP)"
  echo

  echo "## 1. A read-only token"
  echo
  READ_TOKEN=$("$STAR" pair --scope read 2> "$WORK/pair.stderr")
  echo '$ star pair --scope read'
  echo "stderr: $(cat "$WORK/pair.stderr")"
  echo "token: length ${#READ_TOKEN}, scope prefix '$(printf '%s' "$READ_TOKEN" | cut -d. -f1)' (value not printed)"
  OTHER_TOKEN=$("$STAR" pair --scope read 2> /dev/null)
  if [ "$READ_TOKEN" = "$OTHER_TOKEN" ]; then
    echo "two mints, same token (unexpected)"
  else
    echo "two mints: different secrets, same length ${#OTHER_TOKEN}"
  fi
  echo
  start_server "$READ_TOKEN" || exit 1
  AUTH=(-H "Authorization: Bearer $READ_TOKEN")
  request 'GET /v1/health                    # reading is allowed' GET /v1/health "${AUTH[@]}"
  request 'GET /v1/workspaces                # reading is allowed' GET /v1/workspaces "${AUTH[@]}"
  request 'POST /v1/workspaces               # writing is not' POST /v1/workspaces "${AUTH[@]}" \
    -d "{\"path\":\"$WORK/allowed\"}"
  request 'POST .../shell                    # running is not' POST "$SHELL_PATH" "${AUTH[@]}" -d '{}'
  echo
  echo '# the refusal is a real HTTP answer from the policy gate, not a failed connection'
  curl -s -o "$WORK/body" -X POST -H "Authorization: Bearer $READ_TOKEN" \
    -H "Content-Type: application/json" -d '{}' "$BASE/v1/workspaces"
  echo "curl exit=$? (0 = the server answered; the answer itself is the refusal above)"
  echo
  echo '# an unpaired peer is still 401, not 403'
  request 'GET /v1/health  (no token)        # who, before what' GET /v1/health
  stop_server
  echo

  echo "## 2. Running commands needs an approval code"
  echo
  FULL_TOKEN=$("$STAR" pair 2> /dev/null)
  echo '$ star pair                        # no scope: the whole surface'
  echo "token: length ${#FULL_TOKEN}, no scope prefix (compatible with older tokens)"
  echo
  APPROVAL_FILE="$WORK/approval.json"
  start_server "$FULL_TOKEN" --pair-approval-file "$APPROVAL_FILE" || exit 1
  AUTH=(-H "Authorization: Bearer $FULL_TOKEN")
  request 'POST .../shell  (no approval yet)' POST "$SHELL_PATH" "${AUTH[@]}" -d '{}'
  echo
  echo '$ star pair --approval --out /tmp/orbs-scope-proof/approval.json'
  APPROVAL=$("$STAR" pair --approval --out "$APPROVAL_FILE" --approval-ttl 2m 2> "$WORK/approval.stderr")
  echo "stderr: $(sed -E 's/valid until [^.]*\./valid until <expiry>./' "$WORK/approval.stderr")"
  echo "code: length ${#APPROVAL} (value not printed), file mode $(stat -c '%a' "$APPROVAL_FILE")"
  echo
  request 'POST .../shell  (wrong code)' POST "$SHELL_PATH" "${AUTH[@]}" \
    -H "X-Star-Approval: WRONGCODE" -d '{}'
  request 'POST .../shell  (the current code)' POST "$SHELL_PATH" "${AUTH[@]}" \
    -H "X-Star-Approval: $APPROVAL" -d '{}'
  echo
  echo '$ star pair --approval --out ... --approval-ttl 1s   # then wait for it to lapse'
  "$STAR" pair --approval --out "$APPROVAL_FILE" --approval-ttl 1s > /dev/null 2>&1
  sleep 2
  request 'POST .../shell  (expired code, correct value)' POST "$SHELL_PATH" "${AUTH[@]}" \
    -H "X-Star-Approval: $(python3 -c "import json;print(json.load(open('$APPROVAL_FILE'))['code'])")" -d '{}'
  echo
  echo '# reading is untouched by the approval gate'
  request 'GET /v1/health  (no approval)' GET /v1/health "${AUTH[@]}"
  stop_server
  echo

  echo "## 3. One pairing, one workspace"
  echo
  start_server "$FULL_TOKEN" --pair-workspace "$WORK/allowed" || exit 1
  AUTH=(-H "Authorization: Bearer $FULL_TOKEN")
  request 'POST /v1/workspaces  (outside the grant)' POST /v1/workspaces "${AUTH[@]}" \
    -d "{\"path\":\"$WORK/other\",\"client_id\":\"$CLIENT_ID\"}"
  request_quiet 'POST /v1/workspaces  (inside the grant)' POST /v1/workspaces "${AUTH[@]}" \
    -d "{\"path\":\"$WORK/allowed\",\"client_id\":\"$CLIENT_ID\"}"
  echo
  echo '# the created workspace is usable by id, because it is the granted one'
  WORKSPACE=$(python3 -c 'import json;print(json.load(open("'"$WORK"'/body"))["id"])')
  CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 \
    -H "Authorization: Bearer $FULL_TOKEN" "$BASE/v1/workspaces/$WORKSPACE")
  echo "GET /v1/workspaces/<granted-id>       -> HTTP $CODE"
  echo
  echo '# a token with no scope prefix keeps the old, unrestricted behaviour'
  stop_server
  start_server "$FULL_TOKEN" || exit 1
  request_quiet 'POST /v1/workspaces  (no restriction)' POST /v1/workspaces "${AUTH[@]}" \
    -d "{\"path\":\"$WORK/other\",\"client_id\":\"$CLIENT_ID\"}"
  stop_server
} > "$OUT/18-peer-scopes.txt" 2>&1
