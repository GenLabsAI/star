# Star Desktop

A desktop app for driving Star from a native window. It is the upstream
[acp-ui](https://github.com/formulahendry/acp-ui) project (an ACP client with a
Tauri shell) plus a tiny Star overlay applied at build time:

- a preconfigured **`Star` agent** (`star acp` over stdio, first in the list),
- **Star branding** (app name, window title, bundle identifier).

## How users get it

End users never build this. They download the bare `star` launcher for
their OS from a GitHub release (`star-<os>-<arch>[.exe]`, renamed to
`star`), put it on PATH, and run `star`. On first run the launcher
silently fetches the latest `star-core` and the Star Desktop app
(per-user, no admin, background install while the core boots).

## Layout

```
clients/
  acp-ui/          # git submodule, pinned to upstream tag v0.1.16 — NEVER edit here.
                   # Upstream changes belong upstream, not in a fork.
  star-desktop/
    build.sh       # init submodule -> apply patches -> build
    patches/       # the entire Star overlay, as numbered patch files
    README.md      # this file
```

`clients/acp-ui` is a submodule so upstream stays pristine: `go build ./...`,
linters, and every other part of this repo ignore it, and pulling upstream is a
one-line pin bump. The overlay is two small patches, so rebasing onto a newer
acp-ui is `git apply --check` away from a yes/no answer.

## Prerequisites

- Node.js (lts) + npm, git, python3.
- For full installer builds: the Rust toolchain plus
  [Tauri v2 prerequisites](https://v2.tauri.app/start/prerequisites/) for your OS.
- The **Star CLI on your PATH** (install it from the matching Star release).
  The desktop app spawns `star acp` through your login shell; if Star isn't
  installed, starting the agent shows "Failed to spawn agent" — install the CLI
  and retry.

## Local build

```bash
clients/star-desktop/build.sh --frontend-only  # patches + web bundle, no system deps
clients/star-desktop/build.sh                 # full installer (needs Tauri deps)
STAR_DESKTOP_VERSION=0.1.24 clients/star-desktop/build.sh  # stamp a version
```

Installers land in `clients/acp-ui/src-tauri/target/release/bundle/`
(`.msi`/`.exe` on Windows, `.dmg` on macOS, `.deb`/`.AppImage` on Linux).

## Releases

`.github/workflows/desktop.yml` runs automatically every time the CLI release
publishes (and on manual dispatch): it checks out the repo **with submodules**,
applies the patches, stamps the release version, builds installers for
Windows x64, macOS ARM64 + x64, and Linux x64 + ARM64, and uploads them to the
same GitHub release. The desktop job is independent of the CLI build — if it
fails, the CLI release is unaffected.

## Bumping the upstream pin

```bash
cd clients/acp-ui
git fetch --tags origin
git checkout vX.Y.Z          # the new upstream tag
git apply --check ../star-desktop/patches/*.patch   # still clean?
```

If both patches still apply, stage the new gitlink (`git add clients/acp-ui`
from the repo root) and note the new pin in your commit message. If a patch
fails, update the patch file (never edit the submodule content directly) and
re-run `build.sh --frontend-only` to verify.

## Customisation notes

- Icons are still upstream's. To rebrand fully, drop Star icon files into the
  overlay and extend `02-star-branding.patch` with the `bundle.icon` paths.
- Android/iOS targets are intentionally not built: the mobile shell only talks
  to remote agents over WebSocket, while Star runs locally over stdio.
