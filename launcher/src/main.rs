//! The `star` launcher: renders the splash banner, supervises the Go core, and
//! coordinates versioned self-updates across every running instance.

use std::io::{self, Write};
use std::path::Path;
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use star_launcher::coord::{self, Coord};
use star_launcher::desktop;
use star_launcher::handshake;
use star_launcher::splash;
use star_launcher::{platform, update};

/// Core exits closer together than this are treated as a bootloop rather than
/// as a user-initiated update.
const RAPID_EXIT_WINDOW: Duration = Duration::from_secs(10);
const MAX_RAPID_UPDATES: u32 = 3;

/// Leave the alternate screen without touching the main screen's contents.
///
/// A successful handoff deliberately leaves the alternate screen active. The
/// Go core takes over that same screen, so there is no ?1049 exit/re-entry for
/// a terminal to restore a stale buffer through. The launcher still owns the
/// final restore when the core goes away, even if the core was killed before
/// Bubble Tea could clean up.
fn leave_alt_screen(stdout: &Arc<Mutex<io::Stdout>>) {
    if let Ok(mut out) = stdout.lock() {
        let _ = out.write_all(b"\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r");
        let _ = out.flush();
    }
}

/// Restore the terminal after a failed launch.
fn restore_terminal(stdout: &Arc<Mutex<io::Stdout>>) {
    if let Ok(mut out) = stdout.lock() {
        let _ = out.write_all(b"\x1b[2J\x1b[H\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r");
        let _ = out.flush();
    }
}

fn run_core(coord: &Coord) -> i32 {
    let pid = std::process::id();
    let stdout = Arc::new(Mutex::new(io::stdout()));
    let tty = platform::is_tty();
    let stop = Arc::new(AtomicBool::new(false));
    let mut animation = None;
    let start = Instant::now();
    // Stamped into the core's environment so it can tell a request that
    // arrived after this launch from a leftover one from an earlier run.
    let launched_nanos = coord::now_nanos();

    if tty {
        if let Ok(mut out) = stdout.lock() {
            // Reset any scroll region inherited from whatever ran before us
            // (vim, a pager, a fancy prompt). With a region active, frames
            // written below its last row scroll the screen, and the diff then
            // paints against a moved grid, leaving ghost copies of the banner.
            let _ = out.write_all(b"\x1b[?1049h\x1b[?25l\x1b[48;2;0;0;0m\x1b[r\x1b[H\x1b[2J");
            let _ = out.flush();
        }

        let stop_clone = Arc::clone(&stop);
        let stdout_clone = Arc::clone(&stdout);
        animation = Some(thread::spawn(move || {
            let mut size = platform::console_size();
            let mut banner = splash::Splash::new(size.0, size.1);
            while !stop_clone.load(Ordering::Relaxed) {
                // A terminal restored from a saved session can resize during
                // startup. Rebuild on the new grid rather than painting the
                // old layout into a different-sized screen.
                let current = platform::console_size();
                if current != size {
                    size = current;
                    banner = splash::Splash::new(size.0, size.1);
                }
                let bytes = banner.frame(start.elapsed());
                if let Ok(mut out) = stdout_clone.lock() {
                    let _ = out.write_all(bytes.as_bytes());
                    let _ = out.flush();
                }
                thread::sleep(splash::FRAME_INTERVAL);
            }
        }));
    }

    let core = update::find_core();
    let mut args: Vec<String> = std::env::args().skip(1).collect();

    // Resume the exact session this instance was in before the launcher
    // restarted it. An explicit session wins over `--continue`, which would
    // otherwise resume whatever session happens to be newest.
    if let Some(session) = coord.take_session(pid) {
        args.retain(|arg| arg != "--continue" && arg != "-C");
        if let Some(index) = args
            .iter()
            .position(|arg| arg == "--session" || arg == "-s")
        {
            let end = (index + 1).min(args.len().saturating_sub(1));
            args.drain(index..=end);
        }
        args.push("--session".into());
        args.push(session);
    }

    let handshake = handshake::Handshake::new(pid);
    handshake.cleanup();

    let mut child = match Command::new(&core)
        .args(&args)
        .env("STAR_LAUNCHER_HANDSHAKE", "1")
        .env("STAR_LAUNCHER_PID", handshake.env_pid())
        .env("STAR_UPDATE_DIR", coord.root())
        .env("STAR_LAUNCHER_TIME", launched_nanos.to_string())
        .stdin(Stdio::inherit())
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
    {
        Ok(c) => c,
        Err(e) => {
            restore_terminal(&stdout);
            eprintln!("failed to launch star-core: {e}");
            std::process::exit(1);
        }
    };

    handshake.wait_ready();

    // Hold the banner for its full entrance. If the core booted slowly, the
    // banner simply keeps animating until the handoff.
    let remaining = splash::min_duration().saturating_sub(start.elapsed());
    if !remaining.is_zero() {
        thread::sleep(remaining);
    }

    stop.store(true, Ordering::Relaxed);
    if let Some(animation) = animation {
        let _ = animation.join();
    }

    // Do not leave the alternate screen here. The core starts on the same
    // buffer, clears it, and draws the UI. Exiting and re-entering ?1049 is
    // what gives a terminal the opportunity to restore the old splash buffer
    // underneath the new content.
    handshake.signal_release();
    handshake.wait_rendered();
    handshake.cleanup();

    let code = match child.wait() {
        Ok(status) => status.code().unwrap_or(0),
        Err(_) => 1,
    };

    // The launcher entered the alternate screen, so make sure it is the one
    // that hands the terminal back. This also covers a core killed before its
    // own renderer could run its shutdown path.
    if tty {
        leave_alt_screen(&stdout);
    }

    code
}

/// Install a new core version for the instance that asked to update.
///
/// The install writes a new versioned binary and flips the `current` pointer.
/// Other running instances keep the file they were started from, so they do
/// not need to pause or close. This instance reopens on the new version because
/// its core exited with `EXIT_INITIATOR`.
fn run_update(coord: &Coord, pid: u32) -> Result<(), String> {
    // If another instance already holds the lock it is doing this work
    // already; wait for its verdict rather than racing it onto the same file.
    if !coord.acquire_install_lock(pid) {
        coord.wait_for_result(coord::INSTALL_TIMEOUT);
        return Ok(());
    }

    coord.set_phase(coord::Phase::Installing);

    // Versioned installs never overwrite a running core, so no quiesce wait is
    // needed. Old windows continue on their old files; new windows read the
    // new pointer.
    let result = update::perform_update(&update::install_dir());

    match &result {
        Ok(_) => coord.set_phase(coord::Phase::Done),
        Err(error) => coord.set_phase(coord::Phase::Failed(error.clone())),
    }

    // Release only after the phase is published, so a follower that wakes up
    // never observes "nobody is installing" before the result is readable.
    coord.release_install_lock();
    coord.clear_request();
    clear_dir(&coord.root().join("acks"));

    result.map(|_| ())
}

fn clear_dir(dir: &Path) {
    if let Ok(entries) = std::fs::read_dir(dir) {
        for entry in entries.flatten() {
            let _ = std::fs::remove_file(entry.path());
        }
    }
}

/// Resolve the core to launch, downloading it first on a fresh machine.
///
/// Exits with a plain message (no splash) when there is nothing to run:
/// without network there is no core to fall back to, and starting the UI
/// would only fail more confusingly a moment later.
fn ensure_core(coord: &Coord, pid: u32) {
    if update::find_core().is_file() {
        return;
    }
    eprintln!("  First run — downloading the latest Star core (one-time setup)...");
    if coord.acquire_install_lock(pid) {
        let result = update::perform_update(&update::install_dir());
        coord.release_install_lock();
        if let Err(error) = result {
            eprintln!("  could not download star-core: {error}");
            eprintln!("  check your network connection and run `star` again");
            std::process::exit(1);
        }
    } else {
        // Another instance on this machine is already installing.
        coord.wait_for_result(coord::INSTALL_TIMEOUT);
    }
    if !update::find_core().is_file() {
        eprintln!("  star-core is still missing after setup; run `star` again");
        std::process::exit(1);
    }
}

fn main() {
    // Backups from a previous update could not be removed at the time because
    // this binary was still running. Nothing holds them now.
    let pid = std::process::id();
    let coord = Coord::new();
    // Best-effort cleanup of versions left by earlier launches. An old core
    // that is still running cannot be removed on Windows, so cleanup simply
    // leaves it for a later launch.
    update::cleanup_install(&update::install_dir());
    coord.register(pid);

    // Fresh machine, first run: no core on disk yet. Fetch it before the
    // splash so the user sees one setup pass instead of a launch error.
    ensure_core(&coord, pid);

    // Desktop installs in the background while the core boots. It never
    // blocks the terminal and never fails the launch.
    desktop::ensure_desktop_async();

    let mut rapid_updates = 0u32;
    let mut last_exit = Instant::now();

    loop {
        let exit_code = run_core(&coord);

        if exit_code != coord::EXIT_INITIATOR && exit_code != coord::EXIT_FOLLOWER {
            coord.unregister(pid);
            std::process::exit(exit_code);
        }

        if last_exit.elapsed() < RAPID_EXIT_WINDOW {
            rapid_updates += 1;
        } else {
            rapid_updates = 1;
        }
        last_exit = Instant::now();

        // Our core is closed, so this instance can install a new version
        // while every other instance keeps running its existing core.
        coord.ack(pid);

        if exit_code == coord::EXIT_INITIATOR {
            // A failed install must never take the session down with it: report
            // it and fall through to the relaunch below, exactly like a
            // follower, so the user keeps working on the current version.
            if let Err(error) = run_update(&coord, pid) {
                eprintln!("\r\n  star could not install the update: {error}");
                eprintln!("  continuing on the current version\r\n");
            }
        } else {
            // Someone else is updating. Bounded, so an updater that dies
            // mid-install cannot strand this instance with no terminal.
            coord.wait_for_result(coord::INSTALL_TIMEOUT);
        }

        if rapid_updates >= MAX_RAPID_UPDATES {
            coord.unregister(pid);
            coord.clear_request();
            eprintln!("\r\n  the update did not complete cleanly; stopping here\r\n");
            std::process::exit(1);
        }

        let mut out = io::stdout().lock();
        let _ = out.write_all(b"\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r\x1b[H\x1b[2J");
        let _ = out.flush();
    }
}
