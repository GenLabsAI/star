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
    let stdout = Arc::new(std::sync::Mutex::new(io::stdout()));
    let tty = platform::is_tty();
    let stop = Arc::new(AtomicBool::new(false));
    let pulse_finished = Arc::new(AtomicBool::new(false));
    let mut animation = None;

    if tty {
        // Set alt screen, hide cursor, set bg to black, clear screen
        if let Ok(mut out) = stdout.lock() {
            let _ = out.write_all(b"\x1b[?1049h\x1b[?25l\x1b[48;2;0;0;0m\x1b[H\x1b[2J");
            let _ = out.flush();
        }

        let stop_clone = Arc::clone(&stop);
        let pulse_finished_clone = Arc::clone(&pulse_finished);
        let stdout_clone = Arc::clone(&stdout);
        animation = Some(thread::spawn(move || {
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

            let pulse_frames: usize = 72;
            let glow_width: isize = 11;
            let glow_height: isize = 5;

            // Generate a sparse, random star field
            let num_stars = (cols as usize * rows as usize) / 130;
            let mut stars = Vec::with_capacity(num_stars);
            let mut seed = 123456789u32;
            for _ in 0..num_stars {
                seed = seed.wrapping_mul(1664525).wrapping_add(1013904223);
                let x = (seed % cols as u32) as usize;
                seed = seed.wrapping_mul(1664525).wrapping_add(1013904223);
                let y = (seed % rows as u32) as usize;
                // Don't place stars behind the logo/glow area
                let is_near_logo = y >= top.saturating_sub(glow_height as usize)
                    && y <= top + label_height + glow_height as usize
                    && x >= cx.saturating_sub(glow_width as usize)
                    && x <= cx + full_width + glow_width as usize;
                if !is_near_logo {
                    stars.push((x, y));
                }
            }

            // Paint the whole screen black once up front. From then on we only
            // repaint cells that change, which eliminates the full-screen
            // clear that caused the STAR text to flicker each frame.
            let mut init = String::with_capacity(cols as usize * rows as usize + 32);
            init.push_str("\x1b[H\x1b[48;2;0;0;0m\x1b[2J");
            if let Ok(mut out) = stdout_clone.lock() {
                let _ = out.write_all(init.as_bytes());
                let _ = out.flush();
            }

            while !stop_clone.load(Ordering::Relaxed) {
                let mut buf = String::with_capacity(4096);
                let pulse_active = tick < pulse_frames;
                let t = if pulse_active {
                    tick as f64 / (pulse_frames - 1) as f64
                } else {
                    1.0
                };
                // Smoothstep eases the beam's travel. A sine envelope makes
                // the light itself fade in at the left and fade out at the
                // right instead of abruptly appearing or disappearing.
                let eased_t = t * t * (3.0 - 2.0 * t);
                let pulse_envelope = (std::f64::consts::PI * t).sin().powf(0.65);
                let sweep_col = spinner_col as f64 - 6.0 + eased_t * (full_width as f64 + 12.0);

                // Composite each row of the logo band in a single pass: glow
                // background and text glyph are computed per cell and written
                // together, so a cell is drawn exactly once per frame. No
                // clear-then-redraw layering means no flicker as the light
                // passes over the letters.
                if pulse_active || tick == pulse_frames {
                    let center_y = top as f64 + (label_height as f64 - 1.0) / 2.0;
                    for dy in -glow_height..=(label_height as isize + glow_height) {
                        let row = top as isize + dy;
                        if row < 0 || row >= rows as isize {
                            continue;
                        }

                        let vertical = (-0.5 * ((row as f64 - center_y) / 3.1).powi(2)).exp();

                        // The text glyphs for this row, if any.
                        let text_line: Option<&&str> =
                            if row >= top as isize && (row as usize) < top + label_height {
                                label.get(row as usize - top)
                            } else {
                                None
                            };
                        buf.push_str(&format!("\x1b[{};1H", row + 1));

                        let mut last_bg = (0u8, 0u8, 0u8);
                        buf.push_str("\x1b[48;2;0;0;0m");
                        for col in 0..cols as usize {
                            // Background glow for this cell.
                            let (mut bg_r, mut bg_g, mut bg_b) = (0u8, 0u8, 0u8);
                            if pulse_active {
                                let horizontal =
                                    (-0.5 * ((col as f64 - sweep_col) / 4.25).powi(2)).exp();
                                let intensity = horizontal * vertical * pulse_envelope;
                                if intensity >= 0.015 {
                                    bg_r = (76.0 * intensity).round() as u8;
                                    bg_g = (52.0 * intensity).round() as u8;
                                    bg_b = (6.0 * intensity).round() as u8;
                                }
                            }
                            if (bg_r, bg_g, bg_b) != last_bg {
                                buf.push_str(&format!("\x1b[48;2;{};{};{}m", bg_r, bg_g, bg_b));
                                last_bg = (bg_r, bg_g, bg_b);
                            }

                            // Text glyph, if this cell is inside the label.
                            let glyph = text_line.and_then(|line| {
                                if col >= text_start {
                                    line.chars().nth(col - text_start)
                                } else {
                                    None
                                }
                            });

                            match glyph {
                                Some(c) if c != ' ' => {
                                    let char_col = col as f64;
                                    let intensity = if pulse_active {
                                        (-0.5 * ((char_col - sweep_col) / 2.15).powi(2)).exp()
                                            * pulse_envelope
                                    } else {
                                        0.0
                                    };
                                    let fg_g = 255 - (40.0 * intensity) as u8;
                                    let fg_b = 255 - (255.0 * intensity) as u8;

                                    buf.push_str(&format!("\x1b[38;2;255;{};{}m{}", fg_g, fg_b, c));
                                }
                                _ => buf.push(' '),
                            }
                        }
                    }
                }

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
                let spinner_intensity = if pulse_active {
                    (-0.5 * ((spinner_col as f64 - sweep_col) / 2.15).powi(2)).exp()
                        * pulse_envelope
                } else {
                    0.0
                };
                let spinner_bg_r = (76.0 * spinner_intensity).round() as u8;
                let spinner_bg_g = (52.0 * spinner_intensity).round() as u8;
                let spinner_bg_b = (6.0 * spinner_intensity).round() as u8;
                let spinner_fg_g = 180 + (75.0 * spinner_intensity).round() as u8;
                let spinner_fg_b = 180 - (180.0 * spinner_intensity).round() as u8;
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

                // Park the cursor off-screen instead of resetting SGR.
                // A full \x1b[0m reset between frames causes the default
                // background to flash through for one refresh cycle.
                buf.push_str(&format!("\x1b[{};1H", rows + 1));
                if let Ok(mut out) = stdout_clone.lock() {
                    let _ = out.write_all(buf.as_bytes());
                    let _ = out.flush();
                }

                tick += 1;
                if tick > pulse_frames {
                    pulse_finished_clone.store(true, Ordering::Release);
                }
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
            if let Ok(mut out) = stdout.lock() {
                let _ = out.write_all(b"\x1b[?25h\x1b[?1049l");
                let _ = out.flush();
            }
            eprintln!("failed to launch star-core: {e}");
            std::process::exit(1);
        }
    };

    handshake.wait_ready();

    while !pulse_finished.load(Ordering::Acquire) {
        thread::sleep(Duration::from_millis(10));
    }

    thread::sleep(Duration::from_millis(1000));

    stop.store(true, Ordering::Relaxed);
    if let Some(animation) = animation {
        let _ = animation.join();
    }

    // Keep the shared alternate screen active during handoff. Bubble Tea
    // takes ownership of the existing buffer, avoiding a visible switch back
    // through the primary screen and a second alternate-screen entry.
    if let Ok(mut out) = stdout.lock() {
        let _ = out.write_all(b"\x1b[0m\x1b[H\x1b[2J\x1b[?25l");
        let _ = out.flush();
    }

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

        if exit_code != 42 {
            std::process::exit(exit_code);
        }

        let mut stdout = io::stdout();
        let _ = stdout.write_all(b"\x1b[?1049l\x1b[?25h\x1b[0m\x1b[r\x1b[H\x1b[2J");
        let _ = stdout.flush();

        let updater = std::fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&update_lock_path)
            .is_ok();

        if updater {
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
