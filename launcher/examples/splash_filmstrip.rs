//! Renders the real splash animation to a raw frame stream so it can be turned
//! into a video (or inspected frame by frame) without needing a terminal.
//!
//! This drives the exact same `Splash` renderer the launcher uses, so what you
//! see is what the banner actually draws — not a re-implementation.
//!
//! Usage:
//!   cargo run --release --example splash_filmstrip -- <out.raw> [cols] [rows] [ms]

use std::fs::File;
use std::io::{BufWriter, Write};
use std::time::Duration;

use star_launcher::splash::{self, Cell, Splash};

fn main() {
    let mut args = std::env::args().skip(1);
    let out_path = args.next().unwrap_or_else(|| "splash.raw".to_string());
    let cols: u16 = args.next().map(|s| s.parse().unwrap_or(100)).unwrap_or(100);
    let rows: u16 = args.next().map(|s| s.parse().unwrap_or(30)).unwrap_or(30);
    let total_ms: u64 = args.next().map(|s| s.parse().unwrap_or(3000)).unwrap_or(3000);

    let file = File::create(&out_path).expect("create output");
    let mut out = BufWriter::new(file);

    let mut banner = Splash::new(cols, rows);
    let step = splash::FRAME_INTERVAL.as_millis() as u64;
    let steps = total_ms / step;

    for i in 0..=steps {
        let elapsed = Duration::from_millis(i * step);
        let _ = banner.frame(elapsed);

        // Header: magic, frame count placeholder slot, cols, rows.
        if i == 0 {
            write!(out, "STARRSPL1\n{}\n{}\n", cols, rows).unwrap();
        }

        for cell in banner.cells() {
            write!(
                out,
                "{}|{}|{}|{}|{}|{}|{}|",
                cell.bg.0, cell.bg.1, cell.bg.2, cell.fg.0, cell.fg.1, cell.fg.2, cell.ch
            )
            .unwrap();
        }
        out.write_all(b"\n").unwrap();
    }

    out.flush().unwrap();
    eprintln!(
        "wrote {} frames ({cols}x{rows}, {total_ms}ms) to {out_path}",
        steps + 1
    );
    // Keep the type in scope so the example documents the cell shape.
    let _: Cell = Cell {
        ch: ' ',
        fg: (0, 0, 0),
        bg: (0, 0, 0),
    };
}
