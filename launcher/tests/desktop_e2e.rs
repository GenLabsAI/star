//! End-to-end test for the Linux desktop bootstrap, against the real
//! latest GitHub release. Ignored by default: it hits the network and
//! downloads the full AppImage (~70 MB).
//!
//! ```sh
//! cargo test --test desktop_e2e -- --ignored --nocapture
//! ```
//!
//! Uses a scratch HOME, so the real user environment is untouched.

use std::os::unix::fs::PermissionsExt;
use std::path::PathBuf;
use std::time::{Duration, Instant};

/// Fresh, isolated HOME for the test. This is its own test binary (its own
/// process), so mutating process env here cannot leak into other tests.
fn scratch_home() -> PathBuf {
    let dir = std::env::temp_dir().join(format!("star-desktop-e2e-{}", std::process::id()));
    // Edition 2024: env mutation is unsafe.
    // SAFETY: single-threaded at this point; nothing else reads HOME yet.
    unsafe {
        std::env::set_var("HOME", &dir);
        std::env::remove_var("USERPROFILE");
    }
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(&dir).unwrap();
    dir
}

fn wait_for(path: &PathBuf, what: &str) {
    let deadline = Instant::now() + Duration::from_secs(480);
    while !path.is_file() {
        if Instant::now() > deadline {
            panic!("timed out waiting for {what} at {}", path.display());
        }
        std::thread::sleep(Duration::from_secs(5));
    }
}

#[test]
#[ignore]
fn linux_desktop_install_e2e() {
    if cfg!(not(target_os = "linux")) {
        eprintln!("linux-only test, skipping on this platform");
        return;
    }
    let home = scratch_home();
    let bin = home.join("bin");
    let marker = bin.join("desktop-version");
    let appimage = bin.join("star-desktop.AppImage");
    let desktop_entry = home.join(".local/share/applications/star-desktop.desktop");

    // Kick off the same background ensure the real launcher runs at startup.
    star_launcher::desktop::ensure_desktop_async();
    wait_for(&marker, "the desktop version marker");
    // Let the background thread finish writing everything else.
    std::thread::sleep(Duration::from_secs(2));

    let tag = std::fs::read_to_string(&marker).unwrap();
    let tag = tag.trim().to_string();
    assert!(
        tag.starts_with('v'),
        "marker should hold a release tag, got {tag:?}"
    );
    println!("installed desktop tag: {tag}");

    assert!(appimage.is_file(), "AppImage should be installed");
    let meta = std::fs::metadata(&appimage).unwrap();
    println!("appimage size: {} bytes", meta.len());
    assert!(
        meta.len() > 1_000_000,
        "AppImage suspiciously small: {}",
        meta.len()
    );
    assert_ne!(
        meta.permissions().mode() & 0o111,
        0,
        "AppImage should be executable"
    );
    let mut magic = [0u8; 4];
    use std::io::Read as _;
    std::fs::File::open(&appimage)
        .unwrap()
        .read_exact(&mut magic)
        .unwrap();
    assert_eq!(&magic, b"\x7fELF", "AppImage should be an ELF binary");

    let entry = std::fs::read_to_string(&desktop_entry).expect("menu entry should be registered");
    assert!(entry.contains("[Desktop Entry]"), "menu entry malformed");
    assert!(
        entry.contains(appimage.to_string_lossy().as_ref()),
        "menu entry Exec should point at the installed AppImage"
    );

    // Idempotency: a second ensure must not re-download.
    let mtime = meta.modified().unwrap();
    star_launcher::desktop::ensure_desktop_async();
    std::thread::sleep(Duration::from_secs(15));
    let meta2 = std::fs::metadata(&appimage).unwrap();
    assert_eq!(
        mtime,
        meta2.modified().unwrap(),
        "second ensure must not re-download the AppImage"
    );

    // Strongest check: the downloaded artifact is a working AppImage.
    // `--appimage-extract` needs no FUSE. Stage it under target/ (exec-safe,
    // gitignored) in case the temp dir is mounted noexec.
    let stage = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("target")
        .join(format!("e2e-extract-{}", std::process::id()));
    let _ = std::fs::remove_dir_all(&stage);
    std::fs::create_dir_all(&stage).unwrap();
    let staged = stage.join("star-desktop.AppImage");
    std::fs::copy(&appimage, &staged).unwrap();
    std::fs::set_permissions(&staged, std::fs::Permissions::from_mode(0o755)).unwrap();
    let status = std::process::Command::new(&staged)
        .arg("--appimage-extract")
        .current_dir(&stage)
        .status()
        .expect("could not execute the downloaded AppImage");
    assert!(status.success(), "AppImage self-extract failed");
    assert!(
        stage.join("squashfs-root").is_dir(),
        "extracted squashfs-root should exist"
    );
    let _ = std::fs::remove_dir_all(&stage);
    let _ = std::fs::remove_dir_all(&home);
}
