//! Star's splash banner.
//!
//! Every frame is composed into an offscreen cell buffer and then diffed
//! against the previous frame, so a frame only emits the cells that actually
//! changed. Repainting whole rows (or clearing and redrawing the screen) on
//! every tick is what makes terminal splash screens look glitchy: it produces
//! tens of kilobytes of escape sequences per frame, which slower terminals
//! cannot keep up with, and the parts of the banner that did not change still
//! blink while the new content is being transmitted.

use std::fmt::Write as _;
use std::time::Duration;

use crate::platform;

/// An RGB colour.
pub type Rgb = (u8, u8, u8);

/// The STAR wordmark. Kept character-for-character identical to
/// `renderStarWordmark` in `internal/ui/logo`, so the splash and the in-app
/// logo stay the same mark.
pub const WORDMARK: [&str; 3] = [
    "╭──╮╶─┬─╴╭──╮ ╭──╮",
    "╰──╮  │  ├──┤ ├─┬╯",
    "╰──╯  ╵  ╵  ╵ ╵ ╰╴",
];

const MARK: char = '✦';
const HAIRLINE: char = '─';
const DUST: char = '·';

/// How long the light takes to sweep across the wordmark.
const IGNITE_MS: f64 = 1_400.0;
/// How long the settled banner is held before handing off to the core.
const HOLD_MS: u64 = 320;
/// Target frame interval (~30fps).
pub const FRAME_INTERVAL: Duration = Duration::from_millis(33);

/// The shortest the banner should ever be shown: its entrance, plus a beat to
/// land on before the terminal changes hands.
pub fn min_duration() -> Duration {
    Duration::from_millis(IGNITE_MS as u64 + HOLD_MS)
}

const TAU: f64 = std::f64::consts::PI * 2.0;

const BLACK: Rgb = (0, 0, 0);
/// Wordmark before the light reaches it: a dim, unlit bronze.
const DIM: Rgb = (74, 56, 22);
/// Left/right ends of the settled wordmark gradient.
const GOLD_A: Rgb = (255, 168, 56);
const GOLD_B: Rgb = (255, 226, 138);
const WHITE: Rgb = (255, 250, 235);
/// Static warm halo behind the banner, so it sits in space instead of on a void.
const HALO: Rgb = (34, 24, 7);
/// The travelling light.
const BLOOM: Rgb = (104, 70, 14);
/// The hairline track, before the shimmer passes over it.
const RULE: Rgb = (74, 54, 18);

/// One terminal cell.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub struct Cell {
    pub ch: char,
    pub fg: Rgb,
    pub bg: Rgb,
}

impl Cell {
    const fn blank() -> Self {
        Cell {
            ch: ' ',
            fg: BLACK,
            bg: BLACK,
        }
    }
}

struct Star {
    x: usize,
    y: usize,
    phase: f64,
    rate: f64,
    big: bool,
}

#[derive(Clone, Copy)]
struct Layout {
    left: usize,
    width: usize,
    mark_y: usize,
    wm_y: usize,
    rule_y: usize,
}

/// Tracks the SGR state the terminal is actually in, so we only emit a colour
/// change when it differs from what was last sent.
#[derive(Clone, Copy)]
struct Style {
    fg: Rgb,
    bg: Rgb,
}

pub struct Splash {
    cols: usize,
    rows: usize,
    cur: Vec<Cell>,
    prev: Vec<Cell>,
    stars: Vec<Star>,
    layout: Option<Layout>,
    style: Style,
}

impl Splash {
    pub fn new(cols: u16, rows: u16) -> Self {
        let (cols, rows) = (cols as usize, rows as usize);
        let width = wordmark_width();
        let mut splash = Splash {
            cols,
            rows,
            cur: Vec::new(),
            prev: Vec::new(),
            stars: Vec::new(),
            layout: None,
            style: Style { fg: BLACK, bg: BLACK },
        };
        splash.prev = vec![Cell::blank(); cols * rows];

        // Six rows of banner (mark, wordmark, gap, hairline) plus breathing
        // room. Anything smaller just gets a blank screen.
        if cols >= width + 4 && rows >= 9 {
            let top = (rows - 6) / 2;
            let left = (cols - width) / 2;
            splash.layout = Some(Layout {
                left,
                width,
                mark_y: top,
                wm_y: top + 1,
                rule_y: top + 5,
            });
            splash.stars = starfield(&splash.layout, cols, rows);
        }

        splash
    }

    /// True once the banner has finished its entrance and is just idling.
    pub fn settled(&self, elapsed: Duration) -> bool {
        elapsed >= min_duration()
    }

    /// Compose the frame for `elapsed` and return the minimal escape sequence
    /// stream that brings the terminal from the previous frame to this one.
    pub fn frame(&mut self, elapsed: Duration) -> String {
        let cells = self.cols * self.rows;
        self.cur.clear();
        self.cur.resize(cells, Cell::blank());
        self.build(elapsed.as_secs_f64() * 1_000.0);
        let out = self.diff();
        std::mem::swap(&mut self.cur, &mut self.prev);
        out
    }

    /// The cells of the most recently rendered frame.
    pub fn cells(&self) -> &[Cell] {
        &self.prev
    }

    pub fn size(&self) -> (u16, u16) {
        (self.cols as u16, self.rows as u16)
    }

    fn build(&mut self, t: f64) {
        let Some(l) = self.layout else { return };
        let (cols, rows) = (self.cols, self.rows);
        let w = l.width;

        // Sweep position in wordmark-local columns. Runs from just off the left
        // edge to just past the right edge, eased so it accelerates in and
        // decelerates out instead of sliding at a constant, mechanical rate.
        let p = (t / IGNITE_MS).clamp(0.0, 1.0);
        let eased = p * p * (3.0 - 2.0 * p);
        let sweep = -9.0 + eased * (w as f64 + 18.0);
        // Fade the light itself in and out so it never pops into or out of
        // existence at the edges of the banner.
        let env = (std::f64::consts::PI * p).sin().powf(0.7);

        // --- static halo -------------------------------------------------
        let hx = l.left as f64 + w as f64 / 2.0;
        let hy = l.wm_y as f64 + 1.0;
        let rx = w as f64 * 0.62 + 9.0;
        let ry = 4.6;
        for y in 0..rows {
            let dy = (y as f64 - hy) / ry;
            if dy.abs() > 3.0 {
                continue;
            }
            for x in 0..cols {
                let dx = (x as f64 - hx) / rx;
                let f = (-1.7 * (dx * dx + dy * dy)).exp();
                if f > 0.015 {
                    self.cur[y * cols + x].bg = scale(HALO, f);
                }
            }
        }

        // --- travelling bloom --------------------------------------------
        if env > 0.001 {
            let cy = l.wm_y as f64 + 1.0;
            let top = l.mark_y as isize - 1;
            let bot = l.rule_y as isize + 1;
            let x0 = ((sweep - 9.0) as isize).clamp(0, cols as isize - 1) as usize;
            let x1 = ((sweep + 9.0) as isize).clamp(0, cols as isize - 1) as usize;
            for row in top..=bot {
                if row < 0 || row >= rows as isize {
                    continue;
                }
                let vertical = (-0.5 * ((row as f64 - cy) / 2.9).powi(2)).exp();
                if vertical < 0.02 {
                    continue;
                }
                for x in x0..=x1 {
                    let d = x as f64 - sweep;
                    let b = (-0.5 * (d / 3.6).powi(2)).exp() * vertical * env;
                    if b > 0.01 {
                        let cell = &mut self.cur[row as usize * cols + x];
                        cell.bg = mix(cell.bg, BLOOM, b);
                    }
                }
            }
        }

        // --- stars (behind the banner, so the mark overwrites any overlap) ---
        for s in &self.stars {
            let tw = (0.5 + 0.5 * (s.phase + t / 1_000.0 * s.rate).sin()).powf(2.4);
            let v = 0.10 + 0.90 * tw;
            let i = s.y * cols + s.x;
            let cell = &mut self.cur[i];
            cell.ch = if s.big { MARK } else { DUST };
            cell.fg = if s.big {
                // A few brighter stars, tinted warm so they belong to the same
                // light source as the wordmark.
                let warm = scale(GOLD_B, 0.25 + 0.75 * v);
                mix(scale(BLACK, v), warm, 1.0)
            } else {
                // Cool white dust, like a real night sky.
                let l = (30.0 + 205.0 * v).round() as u8;
                let cool = (l as f64 * 0.94) as u8;
                let cool2 = l.saturating_sub(6);
                (l, cool, cool2)
            };
        }

        // --- the mark ------------------------------------------------------
        {
            let breathe = 0.5 + 0.5 * (t / 2_600.0 * TAU).sin();
            let x = l.left + w / 2;
            let y = l.mark_y;
            if y < rows && x < cols {
                let cell = &mut self.cur[y * cols + x];
                cell.ch = MARK;
                cell.fg = mix(GOLD_A, WHITE, breathe * 0.55);
            }
        }

        // --- the wordmark --------------------------------------------------
        for (r, line) in WORDMARK.iter().enumerate() {
            let y = l.wm_y + r;
            if y >= rows {
                break;
            }
            for (i, ch) in line.chars().enumerate() {
                if ch == ' ' {
                    continue;
                }
                let x = l.left + i;
                if x >= cols {
                    break;
                }
                let d = i as f64 - sweep;
                let hot = (-0.5 * (d / 1.9).powi(2)).exp();
                let lit = smoothstep(1.6, -1.6, d);
                let frac = i as f64 / (w.saturating_sub(1)) as f64;
                let grad = mix(GOLD_A, GOLD_B, frac);
                let cell = &mut self.cur[y * cols + x];
                cell.ch = ch;
                cell.fg = mix(mix(DIM, grad, lit), WHITE, hot * 0.9);
            }
        }

        // --- the hairline ----------------------------------------------------
        {
            let period = 2_400.0;
            let sp = (t % period) / period * (w as f64 + 10.0) - 5.0;
            let y = l.rule_y;
            if y < rows {
                for i in 0..w {
                    let x = l.left + i;
                    if x >= cols {
                        break;
                    }
                    let d = i as f64 - sp;
                    let g = (-0.5 * (d / 1.6).powi(2)).exp();
                    let cell = &mut self.cur[y * cols + x];
                    cell.ch = HAIRLINE;
                    cell.fg = mix(RULE, GOLD_B, g);
                }
            }
        }
    }

    /// Emit only the cells that differ from the previous frame, coalescing
    /// contiguous changes into a single run so we pay one cursor-address per
    /// run rather than per cell.
    fn diff(&mut self) -> String {
        let mut out = String::with_capacity(2048);
        let mut st = self.style;
        for y in 0..self.rows {
            let base = y * self.cols;
            let mut x = 0;
            while x < self.cols {
                if self.cur[base + x] == self.prev[base + x] {
                    x += 1;
                    continue;
                }
                let _ = write!(out, "\x1b[{};{}H", y + 1, x + 1);
                while x < self.cols && self.cur[base + x] != self.prev[base + x] {
                    let c = self.cur[base + x];
                    if c.bg != st.bg {
                        let _ = write!(out, "\x1b[48;2;{};{};{}m", c.bg.0, c.bg.1, c.bg.2);
                        st.bg = c.bg;
                    }
                    if c.fg != st.fg {
                        let _ = write!(out, "\x1b[38;2;{};{};{}m", c.fg.0, c.fg.1, c.fg.2);
                        st.fg = c.fg;
                    }
                    out.push(c.ch);
                    x += 1;
                }
            }
        }
        self.style = st;
        out
    }
}

fn wordmark_width() -> usize {
    WORDMARK.iter().map(|l| l.chars().count()).max().unwrap_or(0)
}

fn lcg(state: &mut u32) -> u32 {
    *state = state.wrapping_mul(1_664_525).wrapping_add(1_013_904_223);
    *state
}

fn starfield(layout: &Option<Layout>, cols: usize, rows: usize) -> Vec<Star> {
    let count = (cols * rows / 220).min(64);
    let mut seed: u32 = 0x5EED_1234;
    let mut stars = Vec::with_capacity(count);
    for _ in 0..count {
        let x = lcg(&mut seed) as usize % cols.max(1);
        let y = lcg(&mut seed) as usize % rows.max(1);
        let phase = lcg(&mut seed) as f64 / u32::MAX as f64 * TAU;
        let rate = 1.0 + 1.5 * (lcg(&mut seed) as f64 / u32::MAX as f64);
        let big = lcg(&mut seed) % 7 == 0;

        // Keep stars out of the banner's bounding box so nothing sparkles
        // through the wordmark.
        if let Some(l) = layout {
            if y + 2 >= l.mark_y && y <= l.rule_y + 2 && x + 3 >= l.left && x <= l.left + l.width + 3 {
                continue;
            }
        }
        stars.push(Star { x, y, phase, rate, big });
    }
    stars
}

fn mix(a: Rgb, b: Rgb, t: f64) -> Rgb {
    let t = t.clamp(0.0, 1.0);
    let f = |x: u8, y: u8| (x as f64 + (y as f64 - x as f64) * t).round() as u8;
    (f(a.0, b.0), f(a.1, b.1), f(a.2, b.2))
}

fn scale(c: Rgb, t: f64) -> Rgb {
    (
        (c.0 as f64 * t).round() as u8,
        (c.1 as f64 * t).round() as u8,
        (c.2 as f64 * t).round() as u8,
    )
}

/// Hermite interpolation that also works with `edge0 > edge1`, which is how it
/// is used above to map "distance from the light" onto "already lit".
fn smoothstep(edge0: f64, edge1: f64, x: f64) -> f64 {
    if (edge1 - edge0).abs() < f64::EPSILON {
        return if x < edge0 { 0.0 } else { 1.0 };
    }
    let t = ((x - edge0) / (edge1 - edge0)).clamp(0.0, 1.0);
    t * t * (3.0 - 2.0 * t)
}

/// Read the terminal size, falling back to a sane default.
pub fn terminal_size() -> (u16, u16) {
    platform::console_size()
}

#[cfg(test)]
mod tests {
    use super::*;

    /// A deliberately tiny terminal emulator: just enough of CSI to replay what
    /// `diff` emits (cursor position and truecolor SGR) so a test can prove the
    /// incremental frames add up to exactly the frame the renderer intended.
    /// If diffing ever painted a cell the renderer did not mean to touch, or
    /// lost track of the terminal's colour state, this is where it shows up.
    struct Screen {
        cols: usize,
        rows: usize,
        cells: Vec<Cell>,
        fg: Rgb,
        bg: Rgb,
        row: usize,
        col: usize,
    }

    impl Screen {
        fn new(cols: usize, rows: usize) -> Self {
            Screen {
                cols,
                rows,
                cells: vec![Cell::blank(); cols * rows],
                fg: BLACK,
                bg: BLACK,
                row: 1,
                col: 1,
            }
        }

        fn put(&mut self, ch: char) {
            if self.row >= 1 && self.row <= self.rows && self.col >= 1 && self.col <= self.cols {
                let i = (self.row - 1) * self.cols + (self.col - 1);
                self.cells[i] = Cell {
                    ch,
                    fg: self.fg,
                    bg: self.bg,
                };
            }
            self.col += 1;
        }

        fn feed(&mut self, input: &str) {
            let mut chars = input.chars().peekable();
            while let Some(c) = chars.next() {
                if c != '\x1b' {
                    self.put(c);
                    continue;
                }
                if chars.next() != Some('[') {
                    continue;
                }
                let mut params = String::new();
                let final_byte = loop {
                    match chars.next() {
                        Some(c) if c.is_ascii_alphabetic() => break c,
                        Some(c) => params.push(c),
                        None => break 'm',
                    }
                };
                let parts: Vec<u16> = params
                    .split(';')
                    .map(|p| p.parse::<u16>().unwrap_or(0))
                    .collect();
                match final_byte {
                    'H' => {
                        self.row = parts.first().copied().unwrap_or(1).max(1) as usize;
                        self.col = parts.get(1).copied().unwrap_or(1).max(1) as usize;
                    }
                    'm' => {
                        let mut i = 0;
                        while i < parts.len() {
                            match parts[i] {
                                38 | 48 => {
                                    // 38;2;r;g;b -- the `2` is the colour
                                    // space, so the components start at i+2.
                                    let color = (
                                        parts.get(i + 2).copied().unwrap_or(0) as u8,
                                        parts.get(i + 3).copied().unwrap_or(0) as u8,
                                        parts.get(i + 4).copied().unwrap_or(0) as u8,
                                    );
                                    if parts[i] == 38 {
                                        self.fg = color;
                                    } else {
                                        self.bg = color;
                                    }
                                    i += 5;
                                }
                                _ => i += 1,
                            }
                        }
                    }
                    _ => {}
                }
            }
        }
    }

    const SIZE: (u16, u16) = (100, 30);

    fn run_animation(total_ms: u64) -> (Vec<usize>, usize) {
        let mut banner = Splash::new(SIZE.0, SIZE.1);
        let mut screen = Screen::new(SIZE.0 as usize, SIZE.1 as usize);
        let mut sizes = Vec::new();
        let step = FRAME_INTERVAL.as_millis() as u64;

        let mut t = 0;
        while t <= total_ms {
            let bytes = banner.frame(Duration::from_millis(t));
            screen.feed(&bytes);
            sizes.push(bytes.len());

            for (i, expected) in banner.cells().iter().enumerate() {
                assert_eq!(
                    screen.cells[i], *expected,
                    "cell {i} diverged at t={t}ms: terminal has {:?}, renderer intended {:?}",
                    screen.cells[i], expected
                );
            }
            t += step;
        }
        (sizes, screen.cells.len())
    }

    #[test]
    fn incremental_frames_reconstruct_every_frame() {
        run_animation(2_500);
    }

    #[test]
    fn frames_stay_small() {
        // The old splash repainted whole rows every tick. If this ever creeps
        // back up, the terminal will start dropping frames and the banner will
        // flicker again.
        let (sizes, _) = run_animation(2_500);
        let total: usize = sizes.iter().sum();
        let mean = total / sizes.len();
        assert!(
            mean < 3_000,
            "mean frame is {mean} bytes, which is too much to stream at 30fps"
        );

        let worst = sizes.iter().copied().max().unwrap_or(0);
        // Only the first frame may paint the whole screen.
        assert!(
            worst < 30_000,
            "largest frame is {worst} bytes; a full-screen repaint crept back in"
        );
    }

    #[test]
    fn no_frame_after_the_first_is_full_screen() {
        let (sizes, cells) = run_animation(2_500);
        let budget = cells * 2;
        for (i, size) in sizes.iter().enumerate().skip(1) {
            assert!(
                *size < budget,
                "frame {i} emitted {size} bytes, more than painting the whole \
                 screen twice; the diff is not doing its job"
            );
        }
    }

    /// Whether a character belongs to the wordmark rather than the starfield.
    fn is_banner(ch: char) -> bool {
        ch != ' ' && ch != DUST && ch != MARK
    }

    #[test]
    fn settled_wordmark_stops_changing() {
        // Once the entrance is over the wordmark must be static, otherwise it
        // shimmers at the user while they are trying to read it. (Stars in the
        // surrounding rows are meant to keep twinkling.)
        let mut banner = Splash::new(SIZE.0, SIZE.1);
        let after = min_duration() + Duration::from_millis(50);
        let cols = SIZE.0 as usize;

        banner.frame(after);
        let layout = banner.layout.expect("banner laid out");
        let wordmark = |banner: &Splash| -> Vec<Cell> {
            (0..WORDMARK.len())
                .flat_map(|r| {
                    let y = layout.wm_y + r;
                    (layout.left..layout.left + layout.width)
                        .map(move |x| banner.cells()[y * cols + x])
                })
                .collect()
        };
        let first = wordmark(&banner);

        banner.frame(after + Duration::from_millis(500));
        let later = wordmark(&banner);

        assert_eq!(first, later, "the settled wordmark should not animate");
    }

    #[test]
    fn wordmark_is_centred() {
        for cols in [40u16, 80, 100, 137, 200] {
            let mut banner = Splash::new(cols, SIZE.1);
            banner.frame(min_duration());
            let layout = banner.layout.expect("banner laid out");
            let cells = banner.cells();
            let cols = cols as usize;

            for row in 0..WORDMARK.len() {
                let y = layout.wm_y + row;
                let first = (0..cols).find(|x| is_banner(cells[y * cols + x].ch));
                let last = (0..cols).rev().find(|x| is_banner(cells[y * cols + x].ch));
                let (first, last) = (first.unwrap(), last.unwrap());
                let left_gap = first;
                let right_gap = cols - 1 - last;
                assert!(
                    left_gap.abs_diff(right_gap) <= 1,
                    "wordmark row {row} is off-centre at {cols} columns: \
                     {left_gap} left, {right_gap} right"
                );
            }
        }
    }

    #[test]
    fn tiny_terminal_renders_nothing_rather_than_panicking() {
        for (cols, rows) in [(1u16, 1u16), (10, 3), (19, 8)] {
            let mut banner = Splash::new(cols, rows);
            banner.frame(min_duration());
            assert!(
                banner.cells().iter().all(|c| c.ch == ' '),
                "{cols}x{rows} should fall back to a blank banner"
            );
        }
    }

    #[test]
    fn wordmark_matches_the_app_logo() {
        // The splash and `internal/ui/logo.renderStarWordmark` must stay in step;
        // if one changes, both should.
        assert_eq!(WORDMARK[0], "╭──╮╶─┬─╴╭──╮ ╭──╮");
        assert_eq!(WORDMARK[1], "╰──╮  │  ├──┤ ├─┬╯");
        assert_eq!(WORDMARK[2], "╰──╯  ╵  ╵  ╵ ╵ ╰╴");
        assert_eq!(wordmark_width(), 18);
    }
}
