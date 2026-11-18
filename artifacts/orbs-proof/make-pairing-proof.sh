#!/bin/bash
# Proof for item 20 (pairing/authorization): a real `star server` rejects
# unpaired peers with 401 and accepts the one holding the token. The token
# itself is never printed; only its length and a hash prefix are shown.
#
#   ./artifacts/orbs-proof/make-pairing-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

PROJ=/tmp/orbs-pair-proj
PORT=18999
BASE="http://127.0.0.1:$PORT"
rm -rf "$PROJ"
mkdir -p "$PROJ"
cd "$PROJ" || exit 1

SERVER_PID=""
cleanup() {
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null
  fi
}
trap cleanup EXIT

TOKEN=$("$STAR" pair 2>/tmp/orbs-pair-stderr)
PAIR_EXIT=$?
PAIR_STDERR="$(cat /tmp/orbs-pair-stderr)"
if [ -z "$TOKEN" ]; then
  echo "star pair printed no token" >&2
  exit 1
fi
WRONG="not-the-token"

(
  cd "$PROJ" || exit 1
  STAR_SERVER_TOKEN="$TOKEN" exec "$STAR" server --host "tcp://127.0.0.1:$PORT"
) > /tmp/orbs-pair-server.log 2>&1 &
SERVER_PID=$!

ready=""
for _ in $(seq 1 100); do
  if curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
    ready=yes
    break
  fi
  if [ -s /tmp/orbs-pair-server.log ] && grep -qi "address already in use\|panic" /tmp/orbs-pair-server.log; then
    break
  fi
  sleep 0.1
done
if [ -z "$ready" ]; then
  echo "paired server did not start on port $PORT:" >&2
  cat /tmp/orbs-pair-server.log >&2
  exit 1
fi

status() {
  # status <label> <curl args...>: prints the label and the HTTP status
  local label="$1"
  shift
  local code
  code=$(curl -s -o /tmp/orbs-pair-body -w '%{http_code}' --max-time 5 "$@" "$BASE/v1/health")
  echo "$label -> HTTP $code"
  echo "body: $(cat /tmp/orbs-pair-body)"
}

{
  echo "# scratch project: $PROJ"
  echo "# server: star server --host tcp://127.0.0.1:$PORT with STAR_SERVER_TOKEN set"
  echo "# token length: ${#TOKEN} (value deliberately not printed)"
  echo "# token sha256 prefix: $(printf %s "$TOKEN" | sha256sum | cut -c1-16)"
  echo
  echo '$ star pair        # generates a token, prints it once, stores nothing'
  echo "stdout length: ${#TOKEN} (43 = 32 random bytes, base64url), sha256 prefix: $(printf %s "$TOKEN" | sha256sum | cut -c1-16)"
  echo "exit=$PAIR_EXIT"
  echo "stderr: $PAIR_STDERR"
  echo "no file was written: $(find "$PROJ" -mindepth 1 \( -name '*token*' -o -name '*pair*' \) | wc -l) token or pairing files in the project"
  echo
  echo '# the server is now listening; three real HTTP requests follow'
  echo '$ curl http://127.0.0.1:'"$PORT"'/v1/health          # no credential at all'
  status "unpaired peer"
  echo
  echo '$ curl -H "Authorization: Bearer <wrong>" ...     # wrong credential'
  status "wrong token" -H "Authorization: Bearer $WRONG"
  echo
  echo '$ curl -H "Authorization: Bearer <the paired token>" ...   # authorized peer'
  status "paired peer" -H "Authorization: Bearer $TOKEN"
  echo
  echo '# the paired peer can also use the X-Star-Token header'
  status "paired peer (X-Star-Token)" -H "X-Star-Token: $TOKEN"
  echo
  echo '# 401 responses are real protocol rejections, not connection failures:'
  curl -s -D - -o /dev/null --max-time 5 "$BASE/v1/health" | head -4
  echo
  echo '# who is listening? the paired server shows up in the port inventory'
  echo '$ star ports --all | grep -E "PORT|'"$PORT"'"'
  "$STAR" ports --all | grep -E "PORT|$PORT"
  echo "exit=$?"
  echo
  echo '# with the token in the environment, star env reports pairing without leaking it'
  echo '$ STAR_SERVER_TOKEN=<token> star env | grep -i pair'
  STAR_SERVER_TOKEN="$TOKEN" "$STAR" env | grep -i pair
  echo "exit=$?"
  echo
  echo '$ STAR_SERVER_TOKEN_FILE=<file> star env | grep -i pair'
  (umask 077; printf '%s\n' "$TOKEN" > /tmp/orbs-pair-token)
  STAR_SERVER_TOKEN_FILE=/tmp/orbs-pair-token "$STAR" env | grep -i pair
  echo "exit=$?"
  echo
  echo '# the token file is 0600 and contains only the token'
  ls -l /tmp/orbs-pair-token
  echo "file length: $(wc -c < /tmp/orbs-pair-token) bytes"
  echo
  echo '# stop the server and confirm the port is released'
  kill "$SERVER_PID"
  echo "kill exit=$?"
  SERVER_PID=""
  sleep 0.5
  if curl -s -o /dev/null --max-time 2 "$BASE/v1/health"; then
    echo "request after shutdown -> still answering (unexpected)"
  else
    echo "request after shutdown -> connection refused (curl exit $?)"
  fi
  leftover=$("$STAR" ports --all | grep " $PORT ")
  if [ -z "$leftover" ]; then
    echo "port inventory: no entry for $PORT any more"
  else
    echo "port inventory still lists $PORT: $leftover"
  fi
  echo
  echo '# an unpaired server is unchanged: local sockets keep working without a token'
  (
    cd "$PROJ" || exit 1
    exec "$STAR" server --host "tcp://127.0.0.1:$PORT"
  ) > /tmp/orbs-pair-open.log 2>&1 &
  OPEN_PID=$!
  for _ in $(seq 1 100); do
    if curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
      break
    fi
    sleep 0.1
  done
  status "unpaired server, no credential"
  kill "$OPEN_PID"
  echo "kill exit=$?"
  sleep 0.3
  echo
  echo '# the same token file can guard the server instead of the environment'
  (
    cd "$PROJ" || exit 1
    exec "$STAR" server --host "tcp://127.0.0.1:$PORT" --pair-file /tmp/orbs-pair-token
  ) > /tmp/orbs-pair-file.log 2>&1 &
  FILE_PID=$!
  for _ in $(seq 1 100); do
    if curl -s -o /dev/null --max-time 1 "$BASE/v1/health"; then
      break
    fi
    sleep 0.1
  done
  echo '$ star server --pair-file /tmp/orbs-pair-token'
  status "server with --pair-file, no credential"
  status "server with --pair-file, token from the file (trailing newline trimmed)" -H "Authorization: Bearer $(tr -d '\n' < /tmp/orbs-pair-token)"
  kill "$FILE_PID"
  echo "kill exit=$?"
} > "$OUT/08-pairing.txt" 2>&1
