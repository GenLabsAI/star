//! Coordination for the in-place self-update.
//!
//! Updating means replacing the `star-core` binary that every running Star
//! instance is currently executing, so every instance has to close its core
//! first and reopen it afterwards. The old implementation coordinated that with
//! a couple of bare files in the temp directory and no notion of *which*
//! instance was which, which is why updates were flaky: a paused session had
//! no record of the session it was in and came back empty, a stale request
//! file left behind by a crashed update made every freshly started instance
//! immediately tear itself down again, and a follower that outlived the
//! updater had nothing to wake it up.
//!
//! This module replaces that with an explicit, self-describing protocol:
//!
//! ```text
//! <root>/request          set by the TUI that the user clicked "Update Now" on
//! <root>/state           idle | installing | done | failed
//! <root>/error           why the last install failed
//! <root>/run/<pid>       a live launcher instance
//! <root>/sessions/<pid>  the session that instance must resume with
//! <root>/acks/<pid>      that instance's core is closed and it is safe to
//!                         overwrite the binary
//! <root>/install.lock    held by whichever instance is downloading
//! ```
//!
//! Every instance registers itself on startup and records the session it is in
//! before it pauses, so a session that gets paused by *someone else's* update
//! resumes exactly where it was.

use std::fs;
use std::path::{Path, PathBuf};
use std::thread;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

/// Core exit code: this instance initiated the update.
pub const EXIT_INITIATOR: i32 = 42;
/// Core exit code: this instance was paused by another instance's update.
pub const EXIT_FOLLOWER: i32 = 43;

/// How long the updater gives other instances to close their cores before it
/// overwrites the binary anyway. Bounded so one wedged instance can never hold
/// an update hostage.
pub const QUIESCE_TIMEOUT: Duration = Duration::from_secs(5);
/// How long a paused instance waits for the install to finish before it gives
/// up and resumes on whatever binary is on disk.
pub const INSTALL_TIMEOUT: Duration = Duration::from_secs(10 * 60);
/// A lock older than this belongs to an updater that died mid-install.
const LOCK_STALE: Duration = Duration::from_secs(10 * 60);

const REQUEST: &str = "request";
const STATE: &str = "state";
const ERROR: &str = "error";
const LOCK: &str = "install.lock";
const RUN_DIR: &str = "run";
const ACK_DIR: &str = "acks";
const SESSION_DIR: &str = "sessions";

/// Nanoseconds since the Unix epoch. Used instead of a timestamp in seconds
/// because a same-second comparison cannot tell "requested after I started"
/// from "requested before I started", and that ambiguity is what let a stale
/// request file tear down instances that had just booted.
pub fn now_nanos() -> u128 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0)
}

/// The coordination directory for the current user.
pub fn default_root() -> PathBuf {
    let mut root = std::env::temp_dir();
    // Keep separate users (and separate machines sharing a temp dir) out of
    // each other's way.
    if let Some(user) = std::env::var_os("USER").or_else(|| std::env::var_os("USERNAME")) {
        root.push(format!("star-update-{}", user.to_string_lossy()));
    } else {
        root.push("star-update");
    }
    root
}

/// Where the update protocol currently stands.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Phase {
    Idle,
    Installing,
    Done,
    Failed(String),
}

impl Phase {
    fn encode(&self) -> &'static str {
        match self {
            Phase::Idle => "idle",
            Phase::Installing => "installing",
            Phase::Done => "done",
            Phase::Failed(_) => "failed",
        }
    }

    fn decode(raw: &str, error: &str) -> Self {
        match raw.trim() {
            "installing" => Phase::Installing,
            "done" => Phase::Done,
            "failed" => Phase::Failed(error.trim().to_string()),
            _ => Phase::Idle,
        }
    }
}

pub struct Coord {
    root: PathBuf,
}

impl Coord {
    /// Attach to (and if no update is in flight, reset) the coordination
    /// directory.
    pub fn new() -> Self {
        let coord = Self::at(default_root());
        coord.reset_if_idle();
        coord
    }

    /// Clear leftovers from a previous run, but only when nothing is installing.
    /// An instance that boots in the middle of somebody else's update must not
    /// wipe the state that update is depending on.
    pub fn reset_if_idle(&self) {
        self.ensure_dirs();
        if self.install_lock_is_held() {
            return;
        }
        self.set_phase(Phase::Idle);
        let _ = fs::remove_file(self.path(REQUEST));
        clear_dir(&self.path(ACK_DIR));
    }

    pub fn at(root: PathBuf) -> Self {
        Coord { root }
    }

    pub fn root(&self) -> &Path {
        &self.root
    }

    fn path(&self, rel: &str) -> PathBuf {
        self.root.join(rel)
    }

    fn pid_path(&self, dir: &str, pid: u32) -> PathBuf {
        self.path(dir).join(pid.to_string())
    }

    /// Create the coordination tree. Safe to call repeatedly.
    fn ensure_dirs_only(&self) {
        let _ = fs::create_dir_all(&self.root);
        for dir in [RUN_DIR, ACK_DIR, SESSION_DIR] {
            let _ = fs::create_dir_all(self.path(dir));
        }
    }

    /// Create the coordination tree and make sure the phase file exists.
    pub fn ensure_dirs(&self) {
        self.ensure_dirs_only();
        if !self.state_file().exists() {
            let _ = fs::write(self.state_file(), Phase::Idle.encode());
        }
    }

    fn state_file(&self) -> PathBuf {
        self.path(STATE)
    }

    // --- instance registration ------------------------------------------

    /// Announce that `pid` is a live instance. An instance that dies without
    /// cleaning up leaves a stale entry, which is why quiescing is bounded by
    /// a timeout rather than waiting for a definitive answer.
    pub fn register(&self, pid: u32) {
        self.ensure_dirs();
        let _ = fs::write(self.pid_path(RUN_DIR, pid), pid.to_string());
    }

    pub fn unregister(&self, pid: u32) {
        let _ = fs::remove_file(self.pid_path(RUN_DIR, pid));
        let _ = fs::remove_file(self.pid_path(ACK_DIR, pid));
        let _ = fs::remove_file(self.pid_path(SESSION_DIR, pid));
    }

    // --- request ---------------------------------------------------------

    /// The TUI calls this when the user asks for an update.
    pub fn publish_request(&self) {
        self.ensure_dirs_only();
        let _ = fs::write(self.path(REQUEST), now_nanos().to_string());
    }

    /// Whether an update requested at or after `launched_nanos` is still
    /// waiting to be serviced. Requests older than this instance's own launch
    /// are leftovers from a previous run and must be ignored.
    pub fn request_is_pending(&self, launched_nanos: u128) -> bool {
        match self.request_nanos() {
            Some(requested) => requested > launched_nanos,
            None => false,
        }
    }

    fn request_nanos(&self) -> Option<u128> {
        let meta = fs::metadata(self.path(REQUEST)).ok()?;
        let modified = meta
            .modified()
            .ok()?
            .duration_since(UNIX_EPOCH)
            .ok()?
            .as_nanos();
        Some(modified)
    }

    pub fn clear_request(&self) {
        let _ = fs::remove_file(self.path(REQUEST));
    }

    // --- phase -----------------------------------------------------------

    pub fn phase(&self) -> Phase {
        let raw = fs::read_to_string(self.state_file()).unwrap_or_default();
        let error = fs::read_to_string(self.path(ERROR)).unwrap_or_default();
        Phase::decode(&raw, &error)
    }

    pub fn set_phase(&self, phase: Phase) {
        self.ensure_dirs_only();
        let _ = fs::write(self.state_file(), phase.encode());
        if let Phase::Failed(ref message) = phase {
            let _ = fs::write(self.path(ERROR), message.as_bytes());
        } else {
            let _ = fs::remove_file(self.path(ERROR));
        }
    }

    /// Block until the install finishes. Returns `true` if it finished, and
    /// `false` if it timed out — in which case the caller should resume
    /// anyway rather than leave the user staring at a dead terminal.
    pub fn wait_for_result(&self, timeout: Duration) -> Phase {
        let deadline = Instant::now() + timeout;
        loop {
            let phase = self.phase();
            match phase {
                Phase::Done | Phase::Failed(_) => return phase,
                _ => {}
            }
            if Instant::now() >= deadline {
                return Phase::Idle;
            }
            thread::sleep(Duration::from_millis(100));
        }
    }

    // --- install lock ----------------------------------------------------

    fn install_lock_is_held(&self) -> bool {
        let lock = self.path(LOCK);
        let Ok(meta) = fs::metadata(&lock) else {
            return false;
        };
        // A lock whose owner died is not a lock.
        let stale = meta
            .modified()
            .map(|m| m.elapsed().unwrap_or_default() > LOCK_STALE)
            .unwrap_or(false);
        if stale {
            let _ = fs::remove_file(&lock);
            return false;
        }
        true
    }

    /// Try to become the instance that downloads the update. Returns `false`
    /// if somebody else already is, in which case this instance should just
    /// wait for their result rather than racing them onto the same file.
    pub fn acquire_install_lock(&self, pid: u32) -> bool {
        self.ensure_dirs();
        let lock = self.path(LOCK);
        if fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&lock)
            .is_ok()
        {
            let _ = fs::write(&lock, pid.to_string());
            return true;
        }
        if self.install_lock_is_held() {
            return false;
        }
        // Stale lock from a crashed updater: clear it and try once more.
        let _ = fs::remove_file(&lock);
        fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&lock)
            .is_ok()
            && {
                let _ = fs::write(&lock, pid.to_string());
                true
            }
    }

    pub fn release_install_lock(&self) {
        let _ = fs::remove_file(self.path(LOCK));
    }

    // --- quiesce ---------------------------------------------------------

    /// "My core is closed, the binary is yours to overwrite."
    pub fn ack(&self, pid: u32) {
        self.ensure_dirs_only();
        let _ = fs::write(self.pid_path(ACK_DIR, pid), b"1");
    }

    /// Wait for every other live instance to acknowledge that it has closed
    /// its core. Returns `false` on timeout, in which case the caller must
    /// proceed anyway.
    pub fn wait_for_quiesce(&self, self_pid: u32, timeout: Duration) -> bool {
        let deadline = Instant::now() + timeout;
        loop {
            if self.all_acked(self_pid) {
                return true;
            }
            if Instant::now() >= deadline {
                return false;
            }
            thread::sleep(Duration::from_millis(50));
        }
    }

    fn all_acked(&self, self_pid: u32) -> bool {
        let Ok(entries) = fs::read_dir(self.path(RUN_DIR)) else {
            return true;
        };
        for entry in entries.flatten() {
            let Some(name) = entry.file_name().to_str().map(str::to_owned) else {
                continue;
            };
            let Ok(other) = name.parse::<u32>() else {
                continue;
            };
            if other == self_pid {
                continue;
            }
            if !self.pid_path(ACK_DIR, other).exists() {
                return false;
            }
        }
        true
    }

    // --- session hand-off ------------------------------------------------

    /// Record the session this instance is in, so that pausing and reopening
    /// it lands back on the same conversation.
    pub fn record_session(&self, pid: u32, session: &str) {
        self.ensure_dirs_only();
        let _ = fs::write(self.pid_path(SESSION_DIR, pid), session.as_bytes());
    }

    /// Consume this instance's recorded session, if any.
    pub fn take_session(&self, pid: u32) -> Option<String> {
        let path = self.pid_path(SESSION_DIR, pid);
        let session = fs::read_to_string(&path).ok().map(|s| s.trim().to_string());
        let _ = fs::remove_file(&path);
        match session {
            Some(ref s) if !s.is_empty() => Some(s.clone()),
            _ => None,
        }
    }
}

fn clear_dir(dir: &Path) {
    if let Ok(entries) = fs::read_dir(dir) {
        for entry in entries.flatten() {
            let _ = fs::remove_file(entry.path());
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    struct Temp(PathBuf);

    impl Temp {
        fn new(tag: &str) -> Self {
            let path = std::env::temp_dir().join(format!(
                "star-update-test-{tag}-{}-{:?}",
                std::process::id(),
                std::thread::current().id()
            ));
            let _ = fs::remove_dir_all(&path);
            fs::create_dir_all(&path).expect("create temp dir");
            Temp(path)
        }
    }

    impl Drop for Temp {
        fn drop(&mut self) {
            let _ = fs::remove_dir_all(&self.0);
        }
    }

    #[test]
    fn request_before_launch_is_ignored() {
        let tmp = Temp::new("stale");
        let coord = Coord::at(tmp.0.clone());
        coord.ensure_dirs();

        // A request left over from an earlier run.
        coord.publish_request();
        let launched_after = now_nanos() + 1_000_000_000;

        assert!(!coord.request_is_pending(launched_after));
    }

    #[test]
    fn request_after_launch_is_pending() {
        let tmp = Temp::new("pending");
        let coord = Coord::at(tmp.0.clone());
        coord.ensure_dirs();

        let launched_before = now_nanos().saturating_sub(60_000_000_000);
        coord.publish_request();

        assert!(coord.request_is_pending(launched_before));
    }

    #[test]
    fn cleared_request_is_not_pending() {
        let tmp = Temp::new("cleared");
        let coord = Coord::at(tmp.0.clone());
        coord.ensure_dirs();
        coord.publish_request();
        coord.clear_request();

        assert!(!coord.request_is_pending(0));
    }

    #[test]
    fn session_round_trips_once() {
        let tmp = Temp::new("session");
        let coord = Coord::at(tmp.0.clone());
        coord.record_session(4242, "sess-abc");

        assert_eq!(coord.take_session(4242).as_deref(), Some("sess-abc"));
        assert_eq!(coord.take_session(4242), None);
    }

    #[test]
    fn empty_session_is_ignored() {
        let tmp = Temp::new("emptysession");
        let coord = Coord::at(tmp.0.clone());
        coord.record_session(7, "");

        assert_eq!(coord.take_session(7), None);
    }

    #[test]
    fn phase_round_trips_including_failure() {
        let tmp = Temp::new("phase");
        let coord = Coord::at(tmp.0.clone());

        coord.set_phase(Phase::Idle);
        assert_eq!(coord.phase(), Phase::Idle);

        coord.set_phase(Phase::Installing);
        assert_eq!(coord.phase(), Phase::Installing);

        coord.set_phase(Phase::Done);
        assert_eq!(coord.phase(), Phase::Done);

        coord.set_phase(Phase::Failed("no route to host".into()));
        assert_eq!(
            coord.phase(),
            Phase::Failed("no route to host".to_string())
        );
    }

    #[test]
    fn quiesce_waits_for_other_instances() {
        let tmp = Temp::new("quiesce");
        let coord = Coord::at(tmp.0.clone());
        coord.register(1);
        coord.register(2);
        coord.ack(1);

        // Instance 2 has not closed its core yet.
        assert!(!coord.all_acked(1));

        coord.ack(2);
        assert!(coord.all_acked(1));
    }

    #[test]
    fn quiesce_ignores_self() {
        let tmp = Temp::new("self");
        let coord = Coord::at(tmp.0.clone());
        coord.register(9);

        assert!(coord.all_acked(9));
    }

    #[test]
    fn install_lock_is_exclusive() {
        let tmp = Temp::new("lock");
        let a = Coord::at(tmp.0.clone());
        let b = Coord::at(tmp.0.clone());

        assert!(a.acquire_install_lock(1));
        assert!(!b.acquire_install_lock(2));

        a.release_install_lock();
        assert!(b.acquire_install_lock(2));
    }

    #[test]
    fn wait_for_result_times_out_instead_of_hanging() {
        let tmp = Temp::new("timeout");
        let coord = Coord::at(tmp.0.clone());
        coord.set_phase(Phase::Installing);

        let start = Instant::now();
        let phase = coord.wait_for_result(Duration::from_millis(200));
        assert_eq!(phase, Phase::Idle);
        assert!(start.elapsed() < Duration::from_secs(5));
    }

    #[test]
    fn wait_for_result_returns_immediately_once_done() {
        let tmp = Temp::new("done");
        let coord = Coord::at(tmp.0.clone());
        coord.set_phase(Phase::Installing);

        let waiter = Coord::at(tmp.0.clone());
        std::thread::spawn(move || {
            thread::sleep(Duration::from_millis(50));
            waiter.set_phase(Phase::Done);
        });

        assert_eq!(coord.wait_for_result(Duration::from_secs(5)), Phase::Done);
    }

    /// Walk the whole scenario the update flow has to get right: two sessions
    /// open, one of them asks to update, both cores close, the binary is
    /// replaced, and both sessions come back on the same conversation.
    #[test]
    fn two_sessions_pause_and_resume_on_the_same_sessions() {
        let tmp = Temp::new("multi");
        let root = tmp.0.clone();

        // Two instances start, each with its own session.
        let a = Coord::at(root.clone());
        let b = Coord::at(root.clone());
        a.register(100);
        b.register(200);
        a.record_session(100, "session-a");
        b.record_session(200, "session-b");

        // Both have been running for a while, so neither mistakes an old
        // request for a fresh one.
        let a_launched = now_nanos().saturating_sub(60_000_000_000);
        let b_launched = now_nanos().saturating_sub(60_000_000_000);
        assert!(!a.request_is_pending(a_launched));
        assert!(!b.request_is_pending(b_launched));

        // The user clicks "Update Now" in session A.
        a.publish_request();

        // Both cores notice and shut down, recording where they were.
        assert!(a.request_is_pending(a_launched));
        assert!(b.request_is_pending(b_launched));
        a.ack(100);
        b.ack(200);

        // A performs the install.
        assert!(a.acquire_install_lock(100));
        a.set_phase(Phase::Installing);
        assert!(a.wait_for_quiesce(100, Duration::from_secs(1)));
        a.set_phase(Phase::Done);
        a.release_install_lock();
        a.clear_request();

        // B was paused by somebody else's update and waits for the verdict.
        assert_eq!(b.wait_for_result(Duration::from_secs(5)), Phase::Done);

        // Both come back on exactly the conversation they left.
        assert_eq!(a.take_session(100).as_deref(), Some("session-a"));
        assert_eq!(b.take_session(200).as_deref(), Some("session-b"));
    }

    /// A failed install must not cost anyone their session.
    #[test]
    fn failed_install_still_resumes_everyone() {
        let tmp = Temp::new("failure");
        let root = tmp.0.clone();

        let a = Coord::at(root.clone());
        let b = Coord::at(root.clone());
        a.register(100);
        b.register(200);
        a.record_session(100, "session-a");
        b.record_session(200, "session-b");

        a.publish_request();
        a.ack(100);
        b.ack(200);

        assert!(a.acquire_install_lock(100));
        a.set_phase(Phase::Installing);
        a.set_phase(Phase::Failed("no route to host".into()));
        a.release_install_lock();

        assert_eq!(
            b.wait_for_result(Duration::from_secs(5)),
            Phase::Failed("no route to host".into())
        );
        assert_eq!(a.take_session(100).as_deref(), Some("session-a"));
        assert_eq!(b.take_session(200).as_deref(), Some("session-b"));
    }

    /// An instance that starts *during* an install must not wipe the state the
    /// install is relying on.
    #[test]
    fn starting_mid_install_leaves_the_install_alone() {
        let tmp = Temp::new("midinstall");
        let root = tmp.0.clone();

        let a = Coord::at(root.clone());
        a.register(100);
        a.publish_request();
        assert!(a.acquire_install_lock(100));
        a.set_phase(Phase::Installing);

        // A brand new instance boots here.
        let c = Coord::at(root.clone());
        c.reset_if_idle();

        assert_eq!(c.phase(), Phase::Installing, "a booting instance clobbered the install");
        assert!(c.request_is_pending(now_nanos().saturating_sub(60_000_000_000)));

        a.set_phase(Phase::Done);
        a.release_install_lock();
    }

    /// A leftover request from an update that crashed must not tear down
    /// sessions that start afterwards.
    #[test]
    fn leftover_request_does_not_pause_later_instances() {
        let tmp = Temp::new("leftover");
        let root = tmp.0.clone();

        // An update dies mid-flight, leaving the request behind.
        let dead = Coord::at(root.clone());
        dead.publish_request();
        drop(dead);

        // A new instance starts an hour later.
        let later = Coord::at(root.clone());
        later.reset_if_idle();
        assert!(!later.request_is_pending(now_nanos()));
    }
}
