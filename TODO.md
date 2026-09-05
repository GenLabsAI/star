# Star TODO

## Launcher

- [ ] **Cross-platform splash screen.** The Rust launcher (`launcher/`)
      currently uses Windows-only Win32 APIs (`kernel32`, named events) for
      the console handshake, so it only ships on Windows. Linux and macOS
      releases include just the Go binary with no splash animation. Rework
      the launcher to provide a splash screen on all platforms, either by
      replacing the Win32 FFI with platform-agnostic terminal handling or by
      gating OS-specific parts behind `#[cfg(target_os = "...")]`. Once done,
      wire the launcher into the Linux/macOS jobs in
      `.github/workflows/release.yml`.
