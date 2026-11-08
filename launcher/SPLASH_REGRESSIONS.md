# Splash Regression Notes

Last updated: 2026-11-08

This file is the short memory for splash bugs that keep returning. Read it
before changing `launcher/src/splash.rs`, `launcher/src/main.rs`, or the
launcher/core handoff.

## 1. Ghost STAR / stale splash cells

### Symptoms

- Two STAR wordmarks are visible, usually with a dim older one behind or to
  the left of the new one.
- Old splash content remains at the start of the splash, or after the splash
  has handed off to the Go UI.
- The stale content can appear in the very first visible frame, not only
  after a late terminal restore.

### Cause

The terminal can restore its previous alternate-screen buffer at any time.
An incremental cell diff only writes changed cells, so cells that a new frame
considers blank are left showing the old buffer. A second request for the
alternate screen after the launcher has already entered it can also trigger a
restore, making the old splash reappear.

### Required invariants

1. `Splash::frame` paints every cell in the current offscreen buffer on every
   frame. Do not reintroduce a diff-only fast path for the splash.
2. Never write the bottom-right cell. That cell can scroll the terminal.
3. Reset the scroll region before painting, and keep all splash painting
   inside the alternate screen.
4. The launcher enters the alternate screen exactly once. A successful
   handoff deliberately leaves that screen active for the Go core.
5. The Go core must not request the alternate screen again when it detects
   the launcher's handshake. `splash.launcherOwnsAltScreen` gates that request;
   the core clears the handoff screen when it takes over.
6. A splash fix only reaches users when both binaries are updated. The
   launcher renders the splash; the Go core cannot make an old launcher paint
   differently.

Relevant fixes:

- `de4a9990` - keep the splash off the bottom-right cell and reset the
  scroll region.
- `a6895b1a` - repaint the first frame and keep Bubble Tea off the splash.
- `2a710f6f` - erase late alternate-screen restores.
- `4c147644` - repaint fully and keep the alternate screen through handoff.
- `4b40d0f7` - gate the core's alternate-screen request on the launcher
  handshake; clear closed dialog cells.

### Regression checks

```sh
cargo test --release --manifest-path launcher/Cargo.toml \
  first_frame_clears_a_stale_screen -- --nocapture
cargo test --release --manifest-path launcher/Cargo.toml \
  late_stale_restore_is_erased_at_any_time -- --nocapture
cargo test --release --manifest-path launcher/Cargo.toml \
  every_frame_is_a_full_repaint -- --nocapture
cargo test --release --manifest-path launcher/Cargo.toml \
  never_uses_the_bottom_right_cell -- --nocapture
```

For the handoff, run the single Go test:

```sh
CGO_ENABLED=0 go test ./internal/cmd \
  -run '^TestSplashModelAltScreenWhenLauncherOwnsIt$' -count=1
```

Do not run `go test ./...`; run one package/test at a time.

In a captured launcher session, every pure splash frame must contain exactly
one wordmark, and the stream must contain exactly one `?1049h` request.

## 2. Light sweep on the left instead of the wordmark centre

### Symptoms

- The travelling bloom appears near the left edge of the terminal while the
  wordmark is centred.
- At `IGNITE_MS / 2` the brightest bloom cell is not at the wordmark centre.

### Cause

`sweep` is a **wordmark-local** x coordinate. The bloom is painted into the
screen cell buffer, so using `sweep` directly places the bloom near the left
edge of the screen. This is the old v0.1.89 behaviour:

```rust
let x0 = ((sweep - 9.0) as isize)...;
let x1 = ((sweep + 9.0) as isize)...;
let d = x as f64 - sweep;
```

### Fix

Shift the bloom into screen coordinates once, and use that value for the
bounds and the distance calculation:

```rust
let sweep_x = l.left as f64 + sweep;
let x0 = ((sweep_x - 9.0) as isize)...;
let x1 = ((sweep_x + 9.0) as isize)...;
let d = x as f64 - sweep_x;
```

The wordmark's own illumination still uses local `sweep`. Do not shift the
wordmark coordinates a second time.

The bloom starts outside the wordmark and exits to the right: that movement
is intentional. The regression is that the bloom is off-centre at the
midpoint. If it never reaches the centred midpoint, check the launcher binary
before changing the animation.

### Regression check

```sh
cargo test --release --manifest-path launcher/Cargo.toml \
  light_sweep_is_centred_on_the_wordmark -- --nocapture
```

At 100 columns the test expects the bloom at x=50, the wordmark centre. The
old code leaves it on the left.

Relevant fix: `30c12f61` (`fix(splash): center the travelling bloom on the
wordmark`).

## 3. Stale text after closing a dialog

This one lives in the Go core, but it is part of the same family of repaint
bugs.

### Symptoms

- Closed dialogs or command-palette rows remain painted over the chat.
- The stale text survives until an unrelated repaint.

### Cause

The dialog overlay and the chat blit cache can disagree about which cells are
blank. Closing a dialog must invalidate the cached frame and repaint the
underlying cells, including cells that were covered only by the dialog.

### Fix

Keep dialog rendering derived from the current model; invalidate the chat
cache when a dialog closes; and never let the blit skip a frame that must
erase previously painted cells. See `internal/ui/model/ui.go`,
`internal/ui/model/chat.go`, and `internal/ui/dialog/dialog.go`.

### Regression checks

```sh
CGO_ENABLED=0 go test ./internal/ui/model \
  -run '^(TestDialogCloseClearsPaintedCells|TestChatBlitSkipsUnchangedFrame|TestView_DialogDisablesCache)$' \
  -count=1
CGO_ENABLED=0 go build ./...
```

## 4. Core updated, launcher left behind

The updater in `launcher/src/update.rs` installs `star-core` into a versioned
directory. It does not replace the running `star` launcher. As a result,
`star-core --version` can report the latest release while the splash is still
being drawn by an older launcher.

If a splash bug "comes back" on a new version:

1. Confirm the installed `star` launcher and `star-core` came from the same
   release archive, not just that the core reports the new version.
2. Replace both binaries if either is stale, then restart Star.
3. Only then investigate the renderer.

The release archive must be checked with both binaries present:

```sh
sha256sum star-linux-amd64.tar.gz
tar -tzf star-linux-amd64.tar.gz
./star-linux-amd64/star-core --version
```

Compare the archive hash with the published `checksums.txt`. Run the released
launcher, not only a locally built one, and capture the opening frames.

## When a splash report arrives

- Ask which binary is installed, not just which core version is printed.
- Reproduce from the released archive.
- Capture the first 1-2 seconds, not only the settled screen.
- Check: one wordmark per splash frame, no stale cells, exactly one `?1049h`,
  and the bloom centred at the midpoint.
- Run the focused test that matches the symptom before changing the renderer.
- Never paper over a renderer bug by hiding the splash or weakening a test.
