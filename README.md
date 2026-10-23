# Star

Star is a terminal AI coding assistant — a rebranded fork of Crush (by Charm), built as a two-binary product:

- **star.exe** (Rust launcher, ~200KB) — the fast-start splash + splash banner + self-update.
- **star-core.exe** (Go core) — the agent, TUI, LSP, MCP, hooks, and all the real work.

The launcher prints a gold breathing-star splash on the alt screen, hands the terminal to the Go core once it's ready, and keeps itself in sync with every other running instance during in-place self-updates.

## Quick start

1. Get `star.exe` and `star-core.exe` for your platform (see Releases below).
2. Put them side by side in one folder.
3. Run `star.exe`. It shows the splash, then starts `star-core.exe` and hands over the terminal.

## Updating

Star updates itself. When a new release is available:

- **On Windows:** run `star.exe`. The launcher downloads the latest `star.exe` and `star-core.exe` from GitHub Releases, verifies them against `checksums.txt`, swaps them in, and launches the new version.
- **On macOS/Linux:** same thing — run `./star`; the launcher updates itself and the core.

If the launcher doesn't pick up the new version automatically, drop the matching `star` / `star.exe` and `star-core` / `star-core.exe` from the latest release assets into the same folder as your current binaries.

## Configuration

Star reads `crushrc` (a bash-powered config using builtins like `provider`, `model`, `mcp`, `lsp`, `permissions`, `hook`, `options`) and `crush.json`. Config and context files (`AGENTS.md`, `CRUSH.md`, `CLAUDE.md`, `GEMINI.md` and `.local` variants) are discovered from the working directory.

## Releases

Releases are triggered by pushes to `main` (and `workflow_dispatch`). Each release publishes `star` / `star.exe` (launcher) and `star-core` / `star-core.exe` (Go core) for Windows, Linux (amd64/arm64), and macOS (amd64/arm64), plus checksums.

## Development

- Go core: `go build -o star-core.exe ./main.go` (or `go build .` on the host platform).
- Rust launcher: `cd launcher && cargo build --release` (produces `launcher/target/release/star.exe` on Windows, `star` elsewhere).

## License

MIT
