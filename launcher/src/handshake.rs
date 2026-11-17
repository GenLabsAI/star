use std::path::PathBuf;
use std::process::Child;
use std::time::Duration;
use std::{fs, thread};

/// Outcome of a handshake wait.
pub enum Waited {
    /// The sentinel appeared: the core reached that stage of the launch.
    Signalled,
    /// The core exited first, so the sentinel will never appear.
    CoreExited(i32),
}

pub struct Handshake {
    pid: u32,
    ready_path: PathBuf,
    release_path: PathBuf,
    rendered_path: PathBuf,
}

fn temp_dir() -> PathBuf {
    std::env::temp_dir()
}

impl Handshake {
    pub fn new(pid: u32) -> Self {
        let dir = temp_dir();
        Self {
            pid,
            ready_path: dir.join(format!("star-ready-{pid}")),
            release_path: dir.join(format!("star-release-{pid}")),
            rendered_path: dir.join(format!("star-rendered-{pid}")),
        }
    }

    pub fn env_pid(&self) -> String {
        self.pid.to_string()
    }

    fn wait_for(path: &PathBuf, child: &mut Child) -> Waited {
        // No timeout here - if the core is still working we must keep waiting:
        // releasing early would race it while it is still setting up the DB.
        // A core that has *exited* is different: it will never signal, so
        // waiting on it would leave the splash spinning forever.
        loop {
            if path.exists() {
                let _ = fs::remove_file(path);
                return Waited::Signalled;
            }
            match child.try_wait() {
                Ok(Some(status)) => return Waited::CoreExited(status.code().unwrap_or(1)),
                Ok(None) => thread::sleep(Duration::from_millis(10)),
                Err(_) => return Waited::CoreExited(1),
            }
        }
    }

    pub fn wait_ready(&self, child: &mut Child) -> Waited {
        Self::wait_for(&self.ready_path, child)
    }

    pub fn signal_release(&self) {
        let _ = fs::write(&self.release_path, b"1");
    }

    pub fn wait_rendered(&self, child: &mut Child) -> Waited {
        Self::wait_for(&self.rendered_path, child)
    }

    pub fn cleanup(&self) {
        let _ = fs::remove_file(&self.ready_path);
        let _ = fs::remove_file(&self.release_path);
        let _ = fs::remove_file(&self.rendered_path);
    }
}
