#!/bin/bash
# Proof for item 14 (durable cross-session knowledge): write an entry in
# one process, read it back verbatim in a fresh process, and show where it
# lives on disk.
#
#   ./artifacts/orbs-proof/make-knowledge-proof.sh /path/to/star OUTDIR
set -u
STAR="${1:?path to the star binary}"
OUT="${2:?output directory}"
mkdir -p "$OUT"

PROJ=/tmp/orbs-knowledge-proj
rm -rf "$PROJ"
mkdir -p "$PROJ"
cd "$PROJ"

BODY='# CI quirk

CI runs go test with -count=1, so a cached result cannot hide a failure.
The database container listens on 5433, not 5432.'
printf '%s\n' "$BODY" > /tmp/knowledge-source.md

{
  echo '$ star knowledge list        # nothing stored yet'
  "$STAR" knowledge list
  echo "exit=$?"
  echo
  echo '$ star knowledge show ci-count   # missing entry is an error, not silence'
  "$STAR" knowledge show ci-count
  echo "exit=$?"
  echo
  echo '# process A: record the discovery (a separate OS process)'
  echo '$ star knowledge add ci-count --body "<the discovery>"'
  "$STAR" knowledge add ci-count --body "$BODY"
  echo "exit=$?"

  echo
  echo '# process B: a fresh process with no shared memory reads it back'
  echo '$ star knowledge show ci-count > /tmp/knowledge-readback.md'
  "$STAR" knowledge show ci-count > /tmp/knowledge-readback.md
  echo "exit=$?"
  echo
  echo '--- readback ---'
  cat /tmp/knowledge-readback.md
  echo '--- verbatim check ---'
  echo "sha256(source)   = $(sha256sum /tmp/knowledge-source.md | cut -d' ' -f1)"
  echo "sha256(readback) = $(sha256sum /tmp/knowledge-readback.md | cut -d' ' -f1)"
  if cmp -s /tmp/knowledge-source.md /tmp/knowledge-readback.md; then
    echo "identical: yes"
  else
    echo "identical: NO"
  fi
  echo
  echo '$ star knowledge list'
  "$STAR" knowledge list
  echo
  echo '$ star knowledge list --json'
  "$STAR" knowledge list --json
  echo
  echo '$ star knowledge path && ls -l'
  "$STAR" knowledge path
  ls -l "$("$STAR" knowledge path)"
  echo
  echo '$ star env | grep -i knowledge   # agents discover it from the manifest'
  "$STAR" env | grep -i knowledge
  echo
  echo '# a second entry, and a replacement of the first'
  echo '$ printf "5433, from the compose file" | star knowledge add db-port'
  printf '5433, from the compose file' | "$STAR" knowledge add db-port
  echo "exit=$?"
  echo '$ star knowledge add ci-count --body "corrected: -count=1 is set in Taskfile.yaml"'
  "$STAR" knowledge add ci-count --body "corrected: -count=1 is set in Taskfile.yaml"
  echo "exit=$?"
  echo '$ star knowledge show ci-count'
  "$STAR" knowledge show ci-count
  echo "exit=$?"
  echo
  echo '$ star knowledge remove ci-count && star knowledge list'
  "$STAR" knowledge remove ci-count
  echo "exit=$?"
  "$STAR" knowledge list
  echo "exit=$?"
} > "$OUT/06-knowledge.txt" 2>&1
