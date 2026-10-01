//! The `star` launcher: renders the splash banner, supervises the Go core, and
//! coordinates in-place self-updates across every running instance.

use std::io::{self, Write};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use star_launcher::coord::{self, Coord};
use star_launcher::handshake;
use star_launcher::splash;
use star_launcher::{platform, update};

/// Core exits closer together than this are treated as a bootloop rather than
/// as a user-initiated update.
const RAPID_EXIT_WINDOW: Duration = Duration::from_secs(10);
const MAX_RAPID_UPDATES: u32 = 3;

fn find_core() -> PathBuf {
    let home = std::path::PathBuf::from(
        std::env::var_os("USERPROFILE")
            .or_else(|| std::env::var_os("HOME"))
            .unwrap_or_else(|| ".".into()),
    );
    let bin_dir = home.join("bin");
    if cfg!(target_os = "windows") {
        bin_dir.join("star-core.exe")
    } else {
        bin_dir.join("star-core")
    }
}

/// Leave the terminal in the exact state Bubble Tea expects to inherit: out of
/// the alt screen, cursor visible, all SGR reset, scroll region cleared.
fn restore_terminal(stdout: &Arc<Mutex<io::Stdout>>) {
    if let Ok(mut out) = stdout.lock() {
        let _ = out.write_all(b"\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r");
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
            let _ = out.write_all(b"\x1b[?1049h\x1b[?25l\x1b[48;2;0;0;0m\x1b[H\x1b[2J");
            let _ = out.flush();
        }

        let stop_clone = Arc::clone(&stop);
        let stdout_clone = Arc::clone(&stdout);
        animation = Some(thread::spawn(move || {
            let (cols, rows) = platform::console_size();
            let mut banner = splash::Splash::new(cols, rows);
            while !stop_clone.load(Ordering::Relaxed) {
                let bytes = banner.frame(start.elapsed());
                if let Ok(mut out) = stdout_clone.lock() {
                    let _ = out.write_all(bytes.as_bytes());
                    let _ = out.flush();
                }
                thread::sleep(splash::FRAME_INTERVAL);
            }
        }));
    }

    let core = find_core();
    let mut args: Vec<String> = std::env::args().skip(1).collect();

    // Resume the exact session this instance was in before it paused. An
    // explicit session wins over `--continue`, which would otherwise resume
    // whatever session happens to be newest.
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

    restore_terminal(&stdout);

    handshake.signal_release();
    handshake.wait_rendered();
    handshake.cleanup();

    match child.wait() {
        Ok(status) => status.code().unwrap_or(0),
        Err(_) => 1,
    }
}

/// Perform the install on behalf of every paused instance.
///
/// The updater is whichever instance the user clicked "Update Now" in. It waits
/// for every other live instance to confirm its core is closed, then
/// overwrites the binary exactly once, then tells everybody to come back.
fn run_update(coord: &Coord, pid: u32) -> Result<(), String> {
    // If another instance already holds the lock it is doing this work
    // already; wait for its verdict rather than racing it onto the same file.
    if !coord.acquire_install_lock(pid) {
        coord.wait_for_result(coord::INSTALL_TIMEOUT);
        return Ok(());
    }

    coord.set_phase(coord::Phase::Installing);

    // On Windows the binary is locked for as long as it is running, so this
    // wait is what makes replacing it reliable. It is bounded, because one
    // wedged instance must not be able to block updates forever.
    if !coord.wait_for_quiesce(pid, coord::QUIESCE_TIMEOUT) {
        eprintln!("  another star instance is still shutting down; continuing anyway");
    }

    // Update both binaries: the core we closed, and the launcher we are still
    // running from.
    let launcher = std::env::current_exe().ok();
    let result = update::perform_update(&find_core(), launcher.as_deref());

    match result {
        Ok(()) => coord.set_phase(coord::Phase::Done),
        Err(ref error) => coord.set_phase(coord::Phase::Failed(error.clone())),
    }

    // Release only after the phase is published, so a follower that wakes up
    // never observes "nobody is installing" before the result is readable.
    coord.release_install_lock();
    coord.clear_request();
    clear_dir(&coord.root().join("acks"));

    result
}

fn clear_dir(dir: &Path) {
    if let Ok(entries) = std::fs::read_dir(dir) {
        for entry in entries.flatten() {
            let _ = std::fs::remove_file(entry.path());
        }
    }
}

fn main() {
    // Backups from a previous update could not be removed at the time because
    // this binary was still running. Nothing holds them now.
    update::sweep_stale_backups();

    let pid = std::process::id();
    let coord = Coord::new();
    coord.register(pid);

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

        // Our core is closed, so the binary is now safe to overwrite.
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
