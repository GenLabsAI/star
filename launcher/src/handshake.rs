use std::path::PathBuf;
use std::time::Duration;
use std::{fs, thread};

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

    fn wait_for(path: &PathBuf) {
        while !path.exists() {
            thread::sleep(Duration::from_millis(10));
        }
        let _ = fs::remove_file(path);
    }

    pub fn wait_ready(&self) {
        Self::wait_for(&self.ready_path);
    }

    pub fn signal_release(&self) {
        let _ = fs::write(&self.release_path, b"1");
    }

    pub fn wait_rendered(&self) {
        Self::wait_for(&self.rendered_path);
    }

    pub fn cleanup(&self) {
        let _ = fs::remove_file(&self.ready_path);
        let _ = fs::remove_file(&self.release_path);
        let _ = fs::remove_file(&self.rendered_path);
    }
}
