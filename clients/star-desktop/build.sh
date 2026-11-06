#!/usr/bin/env bash
# Build the Star Desktop app (acp-ui + Star overlay) or verify the frontend only.
#
# Usage:
#   ./build.sh                 # full local build: patches + frontend + tauri build
#   ./build.sh --frontend-only # patches + `npm run build` (no system webkit deps needed)
#   STAR_DESKTOP_VERSION=0.1.24 ./build.sh   # stamp a release version before building
#
# Requirements: git, node (lts), npm. A full `tauri build` additionally needs the
# Rust toolchain plus Tauri's system dependencies for your OS
# (https://v2.tauri.app/start/prerequisites/).
#
# The desktop app spawns `star acp` through your login shell, so the Star CLI
# must be on PATH (install it from the matching Star release first).
set -euo pipefail

FRONTEND_ONLY=0
if [[ "${1:-}" == "--frontend-only" ]]; then
  FRONTEND_ONLY=1
fi

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ACPUI="$ROOT/clients/acp-ui"
PATCHES="$ROOT/clients/star-desktop/patches"

# 1. Make sure the pinned upstream checkout is present and pristine.
git -C "$ROOT" submodule update --init --depth 1 clients/acp-ui >/dev/null
if [[ -n "$(git -C "$ACPUI" status --porcelain)" ]]; then
  echo "error: clients/acp-ui has local changes; commit or stash them first." >&2
  echo "Never commit inside the submodule — upstream changes belong upstream." >&2
  exit 1
fi
PIN="$(git -C "$ROOT" ls-files -s clients/acp-ui | awk '{print $2}')"
HEAD="$(git -C "$ACPUI" rev-parse HEAD)"
if [[ "$PIN" != "$HEAD" ]]; then
  echo "error: submodule is at $HEAD but the repo pins $PIN." >&2
  echo "Run: git -C clients/acp-ui checkout $PIN" >&2
  exit 1
fi

# 2. Apply the Star overlay patches (checked first so a failed patch is fatal).
for p in "$PATCHES"/*.patch; do
  git -C "$ACPUI" apply --check "$p"
done
for p in "$PATCHES"/*.patch; do
  echo "applying $(basename "$p")"
  git -C "$ACPUI" apply "$p"
done

# 3. Optional release version stamp (CI sets STAR_DESKTOP_VERSION from the tag).
if [[ -n "${STAR_DESKTOP_VERSION:-}" ]]; then
  echo "stamping version $STAR_DESKTOP_VERSION"
  python3 - "$ACPUI" "$STAR_DESKTOP_VERSION" <<'PYEOF'
import json, re, sys
acpui, version = sys.argv[1], sys.argv[2]
conf = f"{acpui}/src-tauri/tauri.conf.json"
s = open(conf).read()
s, n = re.subn(r'"version": "[^"]+"', f'"version": "{version}"', s, count=1)
assert n == 1
open(conf, "w").write(s)
cargo = f"{acpui}/src-tauri/Cargo.toml"
s = open(cargo).read()
s, n = re.subn(r'^version = "[^"]+"', f'version = "{version}"', s, count=1, flags=re.M)
assert n == 1
open(cargo, "w").write(s)
pkg = f"{acpui}/package.json"
d = json.load(open(pkg))
d["version"] = version
json.dump(d, open(pkg, "w"), indent=2)
open(pkg, "a").write("\n")
print(f"version -> {version}")
PYEOF
fi

# 4. Frontend: typecheck + vite build (this is also Tauri's beforeBuildCommand).
cd "$ACPUI"
npm ci
npm run build

if [[ "$FRONTEND_ONLY" == "1" ]]; then
  echo "frontend build OK (patches applied, Tauri bundling skipped)."
  exit 0
fi

# 5. Full desktop installer build.
npx tauri build
echo "done — installers are under clients/acp-ui/src-tauri/target/release/bundle/"
