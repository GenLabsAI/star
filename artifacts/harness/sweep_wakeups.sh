#!/usr/bin/env bash
# Reap due Star wakeups for one orb session.
#
# Long-horizon story: no job ever sleeps. A short cron/trigger job runs
# this script every few minutes; each run restores the orb snapshot,
# fires whatever is due via `star wakeup run` (seconds when nothing is
# due), and saves the snapshot back. A wakeup 24h out simply waits
# through the no-op sweeps and fires within one sweep interval of its
# due time. Individual runs never approach the 6h GitHub Actions limit.
#
# Required env:
#   DECA_AGENT_STATE_PASSWORD  password for the encrypted snapshot
#   DECA_AGENT_CHAT_SESSION_ID orb chat session id (selects the snapshot)
#
# Optional env:
#   SNAPSHOT_REPO        repo holding the snapshots (default: GITHUB_REPOSITORY)
#   STAR_BIN             wakeup-capable Star binary (default: star)
#   WAKEUP_EVENT         when set, also fire this named event
#   WAKEUP_SESSION       optional session scope for the event
#   WAKEUP_PROMPT        optional prompt override/direct-run prompt
set -euo pipefail

: "${DECA_AGENT_STATE_PASSWORD:?DECA_AGENT_STATE_PASSWORD is required}"
: "${DECA_AGENT_CHAT_SESSION_ID:?DECA_AGENT_CHAT_SESSION_ID is required}"

HASH_INPUT="${DECA_AGENT_CHAT_SESSION_ID}"
SNAPSHOT_ID="$(printf '%s' "$HASH_INPUT" | sha256sum | cut -c1-32)"
IMAGE="deca-agent-base"
STATE_IMAGE="deca-agent-state:${SNAPSHOT_ID}"
CONTAINER="deca-wakeup-${SNAPSHOT_ID:0:12}"
SNAPSHOT_REPO="${SNAPSHOT_REPO:-${GITHUB_REPOSITORY}}"
SNAPSHOT_TAG="agent-state-${SNAPSHOT_ID}"
SNAPSHOT_FILE="agent-state.tar.zst.enc"
SNAPSHOT_DIR="$(mktemp -d)"
SNAPSHOT_PATH="$SNAPSHOT_DIR/$SNAPSHOT_FILE"
STAR_BIN="${STAR_BIN:-star}"

cleanup() {
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
  rm -rf "$SNAPSHOT_DIR"
}
trap cleanup EXIT

if gh release download "$SNAPSHOT_TAG" --repo "$SNAPSHOT_REPO" --pattern "$SNAPSHOT_FILE" --dir "$SNAPSHOT_DIR" --clobber 2>/dev/null; then
  openssl enc -d -aes-256-cbc -pbkdf2 -salt -in "$SNAPSHOT_PATH" -out "$SNAPSHOT_DIR/agent-state.tar.zst" -pass env:DECA_AGENT_STATE_PASSWORD
  zstd -d -q -f "$SNAPSHOT_DIR/agent-state.tar.zst" -o "$SNAPSHOT_DIR/agent-state.tar"
  docker load -i "$SNAPSHOT_DIR/agent-state.tar"
  IMAGE="$STATE_IMAGE"
else
  echo "No snapshot for session; nothing to sweep."
  exit 0
fi

docker rm -f "$CONTAINER" >/dev/null 2>&1 || true

set +e
docker run --name "$CONTAINER" \
  -e WAKEUP_EVENT="${WAKEUP_EVENT:-}" \
  -e WAKEUP_SESSION="${WAKEUP_SESSION:-}" \
  -e WAKEUP_PROMPT="${WAKEUP_PROMPT:-}" \
  -w /agent/workspace \
  "$IMAGE" \
  bash -lc '
    set -e
    bin="'"$STAR_BIN"'"
    command -v "$bin" >/dev/null || { echo "No $bin in orb image; skipping sweep."; exit 0; }
    "$bin" wakeup run
    if [ -n "${WAKEUP_EVENT:-}" ]; then
      args=(wakeup trigger --event "$WAKEUP_EVENT")
      [ -n "${WAKEUP_SESSION:-}" ] && args+=(--session "$WAKEUP_SESSION")
      [ -n "${WAKEUP_PROMPT:-}" ] && args+=(--prompt "$WAKEUP_PROMPT")
      "$bin" "${args[@]}" || true
    fi
  '
STATUS=$?
set -e

# Claims advance repeating schedules, so always write the state back.
docker commit "$CONTAINER" "$STATE_IMAGE"
docker save "$STATE_IMAGE" -o "$SNAPSHOT_DIR/agent-state.tar"
zstd -q -f "$SNAPSHOT_DIR/agent-state.tar" -o "$SNAPSHOT_DIR/agent-state.tar.zst"
openssl enc -aes-256-cbc -pbkdf2 -salt -in "$SNAPSHOT_DIR/agent-state.tar.zst" -out "$SNAPSHOT_PATH" -pass env:DECA_AGENT_STATE_PASSWORD

gh release delete "$SNAPSHOT_TAG" --repo "$SNAPSHOT_REPO" --yes 2>/dev/null || true
gh release create "$SNAPSHOT_TAG" "$SNAPSHOT_PATH" --repo "$SNAPSHOT_REPO" --title "$SNAPSHOT_TAG" --notes "Encrypted Deca agent state snapshot" --latest=false

exit "$STATUS"
