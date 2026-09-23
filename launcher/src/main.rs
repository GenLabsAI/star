mod handshake;
mod platform;
mod update;

use std::env;
use std::io::{self, Write};
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::thread;
use std::time::Duration;

fn find_core() -> PathBuf {
    if let Ok(exe_path) = env::current_exe() {
        if let Some(dir) = exe_path.parent() {
            let local_core = if cfg!(target_os = "windows") {
                dir.join("star-core.exe")
            } else {
                dir.join("star-core")
            };
            if local_core.exists() {
                return local_core;
            }
        }
    }

    let home = env::var_os("USERPROFILE")
        .or_else(|| env::var_os("HOME"))
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."));
    let bin_dir = home.join("bin");
    if cfg!(target_os = "windows") {
        bin_dir.join("star-core.exe")
    } else {
        bin_dir.join("star-core")
    }
}

fn run_core() -> i32 {
    let mut stdout = io::stdout();
    let tty = platform::is_tty();
    let stop = Arc::new(AtomicBool::new(false));
    let mut animation = None;

    if tty {
        // Set alt screen, hide cursor, set bg to black, clear screen
        let _ = stdout.write_all(b"\x1b[?1049h\x1b[?25l\x1b[48;2;0;0;0m\x1b[H\x1b[2J");
        let _ = stdout.flush();

        let stop_clone = Arc::clone(&stop);
        animation = Some(thread::spawn(move || {
            let mut out = io::BufWriter::with_capacity(64 * 1024, io::stdout());
            let label = [
                "╭──╮╶─┬─╴╭──╮ ╭──╮",
                "╰──╮  │  ├──┤ ├─┬╯",
                "╰──╯  ╵  ╵  ╵ ╵ ╰╴",
            ];
            let braille = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
            let star_symbol = '✦';
            let (cols, rows) = platform::console_size();
            let label_width = label
                .iter()
                .map(|line| line.chars().count())
                .max()
                .unwrap_or(0);
            // spinner(1) + gap(3) + label
            let spinner_gap = 3;
            let full_width = 1 + spinner_gap + label_width;
            let label_height = label.len();
            let top = ((rows as usize).saturating_sub(label_height)) / 2;
            let cx = ((cols as usize).saturating_sub(full_width)) / 2;
            let text_start = cx + 1 + spinner_gap;
            let spinner_col = cx;
            let mut tick: usize = 0;

            // Generate a sparse, random star field
            let num_stars = (cols as usize * rows as usize) / 130;
            let mut stars = Vec::with_capacity(num_stars);
            let mut seed = 123456789u32;
            for _ in 0..num_stars {
                seed = seed.wrapping_mul(1664525).wrapping_add(1013904223);
                let x = (seed % cols as u32) as usize;
                seed = seed.wrapping_mul(1664525).wrapping_add(1013904223);
                let y = (seed % rows as u32) as usize;
                // Don't place stars anywhere near the logo or spinner
                let is_near_logo = y >= top.saturating_sub(2)
                    && y <= top + label_height + 2
                    && x >= cx.saturating_sub(4)
                    && x <= cx + full_width + 4;
                if !is_near_logo {
                    stars.push((x, y));
                }
            }

            // Paint the whole screen black once up front. From then on we only
            // repaint cells that change, which eliminates the full-screen
            // clear that caused the STAR text to flicker each frame.
            let mut init = String::with_capacity(cols as usize * rows as usize + 256);
            init.push_str("\x1b[H\x1b[48;2;0;0;0m\x1b[2J");
            init.push_str("\x1b[38;2;255;255;255m"); // White text for logo
            for (i, line) in label.iter().enumerate() {
                init.push_str(&format!("\x1b[{};{}H{}", top + i + 1, text_start + 1, line));
            }
            let _ = out.write_all(init.as_bytes());
            let _ = out.flush();

            while !stop_clone.load(Ordering::Relaxed) {
                let mut buf = String::with_capacity(4096);

                // Draw twinkling stars with staggered animation phases. Each
                // star is followed by a space to erase the right-edge overhang
                // the glyph leaves in the next cell.
                for &(x, y) in &stars {
                    // Brightness follows a smooth sinusoid so stars breathe
                    // rather than hard-flicker.
                    let symbol = star_symbol;
                    let twinkle_t = tick as f64 * 0.075 + (x * 31 + y * 17) as f64 * 0.05;
                    let twinkle = (twinkle_t.sin() + 1.0) * 0.5;
                    let brightness = (65.0 + 125.0 * twinkle).round() as u8;
                    buf.push_str(&format!(
                        "\x1b[{};{}H\x1b[48;2;0;0;0m\x1b[38;2;{};{};{}m{}\x1b[{};{}H ",
                        y + 1,
                        x + 1,
                        brightness,
                        brightness,
                        brightness,
                        symbol,
                        y + 1,
                        x + 2,
                    ));
                }

                let spinner = braille[tick % braille.len()];
                let spinner_bg_r = 0;
                let spinner_bg_g = 0;
                let spinner_bg_b = 0;
                let spinner_fg_g = 180;
                let spinner_fg_b = 180;
                buf.push_str(&format!(
                    "\x1b[{};{}H\x1b[48;2;{};{};{}m\x1b[38;2;255;{};{}m{}",
                    top + label_height / 2 + 1,
                    spinner_col + 1,
                    spinner_bg_r,
                    spinner_bg_g,
                    spinner_bg_b,
                    spinner_fg_g,
                    spinner_fg_b,
                    spinner,
                ));

                for (i, line) in label.iter().enumerate() {
                    buf.push_str(&format!(
                        "\x1b[{};{}H\x1b[48;2;0;0;0m\x1b[38;2;255;255;255m{}",
                        top + i + 1,
                        text_start + 1,
                        line
                    ));
                }

                let _ = out.write_all(buf.as_bytes());
                let _ = out.flush();

                tick += 1;
                thread::sleep(Duration::from_millis(50));
            }
        }));
    }

    let core = find_core();
    let mut args: Vec<String> = env::args().skip(1).collect();

    if let Ok(session) = env::var("STAR_UPDATE_SESSION") {
        args.retain(|arg| arg != "--continue" && arg != "-C");
        if let Some(index) = args
            .iter()
            .position(|arg| arg == "--session" || arg == "-s")
        {
            args.drain(index..=(index + 1).min(args.len() - 1));
        }
        args.push("--session".into());
        args.push(session);
        unsafe {
            env::remove_var("STAR_UPDATE_SESSION");
        }
    }

    let handshake = handshake::Handshake::new(std::process::id());
    handshake.cleanup();

    let mut child = match Command::new(&core)
        .args(&args)
        .env("STAR_LAUNCHER_HANDSHAKE", "1")
        .env("STAR_LAUNCHER_PID", handshake.env_pid())
        .stdin(Stdio::inherit())
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
    {
        Ok(c) => c,
        Err(e) => {
            let _ = stdout.write_all(b"\x1b[?25h\x1b[?1049l");
            let _ = stdout.flush();
            eprintln!("failed to launch star-core: {e}");
            std::process::exit(1);
        }
    };

    handshake.wait_ready();

    stop.store(true, Ordering::Relaxed);
    if let Some(animation) = animation {
        let _ = animation.join();
    }

    // Clear our alt-screen buffer completely so nothing bleeds into Bubble Tea's
    // session, then restore a pristine terminal before handing off.
    let _ = stdout.write_all(b"\x1b[H\x1b[2J\x1b[0m\x1b[39;49m\x1b[r\x1b[?7h\x1b[?25h\x1b[?1049l");
    let _ = stdout.flush();

    handshake.signal_release();
    handshake.wait_rendered();
    handshake.cleanup();

    let status = child.wait();

    match status {
        Ok(s) => s.code().unwrap_or(0),
        Err(_) => 1,
    }
}

fn main() {
    let launcher_pid = std::process::id();
    let update_session_path = env::temp_dir().join(format!("star-update-session-{launcher_pid}"));
    let update_request_path = env::temp_dir().join("star-update-request");
    let update_lock_path = env::temp_dir().join("star-update-lock");

    loop {
        let launch_time = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();
        unsafe {
            env::set_var("STAR_LAUNCHER_TIME", launch_time.to_string());
        }

        let exit_code = run_core();

        if exit_code != 42 && exit_code != 43 {
            // On normal exit, clear any lingering update session files to avoid cross-talk
            let _ = std::fs::remove_file(&update_session_path);
            std::process::exit(exit_code);
        }

        let mut stdout = io::stdout();
        let _ = stdout.write_all(b"\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r\x1b[H\x1b[2J");
        let _ = stdout.flush();

        // The instance that initiates the update exits with 42.
        // Other instances following the broadcast signal exit with 43.
        let updater = exit_code == 42;

        if updater {
            // The updater creates the lock and does the work.
            let _lock = std::fs::OpenOptions::new()
                .write(true)
                .create(true)
                .open(&update_lock_path);
            thread::sleep(Duration::from_millis(500));
            let core = find_core();
            if let Err(error) = update::perform_update(&core) {
                let _ = std::fs::remove_file(&update_lock_path);
                let _ = std::fs::remove_file(&update_request_path);
                let _ = std::fs::remove_file(&update_session_path);
                eprintln!("Update failed: {error}");
                std::process::exit(1);
            }
            let _ = std::fs::remove_file(&update_request_path);
            let _ = std::fs::remove_file(&update_lock_path);
        } else {
            // Followers just wait until both signal files are gone.
            while update_request_path.exists() || update_lock_path.exists() {
                thread::sleep(Duration::from_millis(100));
            }
        }

        // As a safeguard against bootloops, always ensure signals are cleared before relaunch.
        let _ = std::fs::remove_file(&update_request_path);
        let _ = std::fs::remove_file(&update_lock_path);

        if let Ok(session_id) = std::fs::read_to_string(&update_session_path) {
            let session_id = session_id.trim();
            if !session_id.is_empty() {
                unsafe {
                    env::set_var("STAR_UPDATE_SESSION", session_id);
                }
            }
            let _ = std::fs::remove_file(&update_session_path);
        }
    }
}
