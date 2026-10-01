//! Self-update: download the latest release from GitHub and replace both
//! binaries it ships -- `star-core`, which the user actually spends their time
//! in, and the `star` launcher that draws the splash.
//!
//! Replacing the launcher is the interesting part, because unlike the core it
//! is still running while we do it. That turns out to be fine on both
//! platforms: on Unix a `rename` over a running executable is legal (the
//! process keeps the old inode), and on Windows a running image cannot be
//! deleted or overwritten but *can* be renamed, which is what
//! `atomic_replace` already does to make room for the new one. What Windows
//! will not allow is deleting that moved-aside copy while it is still mapped,
//! so it is left behind and swept on the next launch.

use std::fs;
use std::io::{Read, Write};
use std::path::Path;
use std::time::Duration;

const GITHUB_RELEASE_LATEST: &str = "https://api.github.com/repos/genlabsai/star/releases/latest";
const USER_AGENT: &str = "star-launcher";
const CONNECT_TIMEOUT: Duration = Duration::from_secs(10);
const CALL_TIMEOUT: Duration = Duration::from_secs(120);
/// Transient network failures are common enough that a single failed attempt
/// should not surface to the user as "update failed".
const ATTEMPTS: u32 = 3;

/// Asset name for the current platform.
fn asset_name() -> &'static str {
    if cfg!(target_os = "windows") && cfg!(target_arch = "x86_64") {
        "star-windows-amd64.zip"
    } else if cfg!(target_os = "linux") && cfg!(target_arch = "x86_64") {
        "star-linux-amd64.tar.gz"
    } else if cfg!(target_os = "linux") && cfg!(target_arch = "aarch64") {
        "star-linux-arm64.tar.gz"
    } else if cfg!(target_os = "macos") && cfg!(target_arch = "x86_64") {
        "star-darwin-amd64.tar.gz"
    } else if cfg!(target_os = "macos") && cfg!(target_arch = "aarch64") {
        "star-darwin-arm64.tar.gz"
    } else {
        "star-linux-amd64.tar.gz" // fallback
    }
}

/// The core binary name inside the release archive.
fn core_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "star-core.exe"
    } else {
        "star-core"
    }
}

/// Our own binary name, as shipped in the release archive.
fn launcher_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "star.exe"
    } else {
        "star"
    }
}

/// Both binaries a release archive carries.
///
/// `launcher` is optional so that an archive missing it -- an older release,
/// or a packaging mistake -- downgrades to "the splash will not change"
/// rather than failing an otherwise-good update.
struct Extracted {
    core: Vec<u8>,
    launcher: Option<Vec<u8>>,
}

impl std::fmt::Debug for Extracted {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Extracted")
            .field("core", &format_args!("{} bytes", self.core.len()))
            .field(
                "launcher",
                &self.launcher.as_ref().map(|l| format!("{} bytes", l.len())),
            )
            .finish()
    }
}

/// Build an agent with an explicit timeout. Without one, a stalled connection
/// can hang the update indefinitely with the user's terminal already torn down.
fn agent(global: Duration) -> ureq::Agent {
    ureq::Agent::config_builder()
        .user_agent(USER_AGENT)
        .timeout_connect(Some(CONNECT_TIMEOUT))
        .timeout_global(Some(global))
        .build()
        .into()
}

fn request(url: &str) -> Result<ureq::http::Response<ureq::Body>, String> {
    agent(CALL_TIMEOUT)
        .get(url)
        .header("Accept", "application/vnd.github.v3+json")
        .call()
        .map_err(|e| format!("{e}"))
}

/// Retry `op` a few times, sleeping between attempts. The last error is
/// returned if every attempt fails.
fn retry<T, F>(mut op: F) -> Result<T, String>
where
    F: FnMut() -> Result<T, String>,
{
    let mut last = String::from("update failed");
    for attempt in 1..=ATTEMPTS {
        match op() {
            Ok(value) => return Ok(value),
            Err(error) => {
                last = error;
                if attempt < ATTEMPTS {
                    eprintln!("  retrying ({}/{})...", attempt, ATTEMPTS);
                    std::thread::sleep(Duration::from_millis(500 * attempt as u64));
                }
            }
        }
    }
    Err(last)
}

/// Fetch the browser_download_url for our platform asset from the latest
/// GitHub release.
fn fetch_asset_url() -> Result<String, String> {
    let body: String = retry(|| {
        let resp = request(GITHUB_RELEASE_LATEST)?;
        resp.into_body()
            .read_to_string()
            .map_err(|e| format!("could not read release metadata: {e}"))
    })?;

    // Minimal JSON parsing to avoid pulling in serde_json and bloating the
    // binary. GitHub lists `name` before `browser_download_url` within each
    // asset object, so splitting on the name key keeps the two together.
    let wanted = asset_name();
    let mut best: Option<String> = None;
    for block in body.split("\"name\"").skip(1) {
        let Some(rest) = block.strip_prefix('"') else {
            continue;
        };
        let Some(end) = rest.find('"') else {
            continue;
        };
        if &rest[..end] != wanted {
            continue;
        }
        // The download URL appears after the name within the same asset object.
        let Some(key) = block.find("\"browser_download_url\"") else {
            continue;
        };
        let after = &block[key + "\"browser_download_url\"".len()..];
        let Some(after) = after.trim_start().strip_prefix(':') else {
            continue;
        };
        let Some(after) = after.trim_start().strip_prefix('"') else {
            continue;
        };
        let Some(end) = after.find('"') else {
            continue;
        };
        best = Some(after[..end].to_string());
        break;
    }

    best.ok_or_else(|| format!("release has no {wanted} asset for this platform"))
}

/// Download bytes from a URL.
fn download(url: &str) -> Result<Vec<u8>, String> {
    retry(|| {
        let resp = agent(CALL_TIMEOUT)
            .get(url)
            .call()
            .map_err(|e| format!("{e}"))?;
        let mut buf = Vec::new();
        resp.into_body()
            .into_reader()
            .read_to_end(&mut buf)
            .map_err(|e| format!("could not read download: {e}"))?;
        if buf.is_empty() {
            return Err("download returned no data".into());
        }
        Ok(buf)
    })
}

/// Pull both binaries out of a .zip archive (Windows).
#[cfg(target_os = "windows")]
fn extract_all(archive_bytes: &[u8]) -> Result<Extracted, String> {
    use std::io::Cursor;
    let reader = Cursor::new(archive_bytes);
    let mut archive = zip::ZipArchive::new(reader)
        .map_err(|e| format!("could not open the downloaded archive: {e}"))?;

    let core_name = core_binary_name();
    let launcher_name = launcher_binary_name();
    let mut core = None;
    let mut launcher = None;

    for i in 0..archive.len() {
        let mut file = archive
            .by_index(i)
            .map_err(|e| format!("could not read archive entry: {e}"))?;
        let name = file.name().replace('\\', "/");
        // Releases nest both binaries under a per-platform directory, so match
        // on the last path segment rather than the full entry name.
        let base = name.rsplit('/').next().unwrap_or(&name);
        let wanted = match base {
            n if n == core_name => Some(true),
            n if n == launcher_name => Some(false),
            _ => None,
        };
        let Some(is_core) = wanted else { continue };
        if is_core && core.is_some() || !is_core && launcher.is_some() {
            continue;
        }

        let mut buf = Vec::new();
        file.read_to_end(&mut buf)
            .map_err(|e| format!("could not read {base} from the archive: {e}"))?;
        if is_core {
            core = Some(buf);
        } else {
            launcher = Some(buf);
        }
    }

    Ok(Extracted {
        core: core.ok_or_else(|| format!("{core_name} was not in the downloaded archive"))?,
        launcher,
    })
}

/// Pull both binaries out of a .tar.gz archive (Unix).
#[cfg(not(target_os = "windows"))]
fn extract_all(archive_bytes: &[u8]) -> Result<Extracted, String> {
    use std::io::Cursor;
    let gz = flate2::read::GzDecoder::new(Cursor::new(archive_bytes));
    let mut archive = tar::Archive::new(gz);
    let core_name = core_binary_name();
    let launcher_name = launcher_binary_name();
    let mut core = None;
    let mut launcher = None;

    for entry in archive
        .entries()
        .map_err(|e| format!("could not read the downloaded archive: {e}"))?
    {
        let mut entry = entry.map_err(|e| format!("could not read archive entry: {e}"))?;
        let path = entry
            .path()
            .map_err(|e| format!("bad path in archive: {e}"))?
            .to_path_buf();
        // Match on the last path segment: releases nest both binaries under a
        // per-platform directory.
        let base = path.file_name().map(|n| n.to_string_lossy().into_owned());
        let Some(base) = base else { continue };
        let is_core = if base == core_name {
            Some(true)
        } else if base == launcher_name {
            Some(false)
        } else {
            None
        };
        let Some(is_core) = is_core else { continue };
        if is_core && core.is_some() || !is_core && launcher.is_some() {
            continue;
        }

        let mut buf = Vec::new();
        entry
            .read_to_end(&mut buf)
            .map_err(|e| format!("could not read {base} from the archive: {e}"))?;
        if is_core {
            core = Some(buf);
        } else {
            launcher = Some(buf);
        }
    }

    Ok(Extracted {
        core: core.ok_or_else(|| format!("{core_name} was not in the downloaded archive"))?,
        launcher,
    })
}

/// Replace `dest` with `data`, leaving it executable on Unix.
fn install(dest: &Path, data: &[u8], label: &str) -> Result<(), String> {
    atomic_replace(dest, data)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(dest, fs::Permissions::from_mode(0o755))
            .map_err(|e| format!("could not make {label} executable: {e}"))?;
    }
    Ok(())
}

/// Write bytes to a file atomically: write to a temp file next to `dest`, then
/// rename over the target. The previous binary is kept until the rename
/// succeeds so a failure here cannot leave the user with no binary at all.
///
/// Temp and backup names are derived from `dest`, so replacing the launcher and
/// the core in the same directory cannot have them tread on each other.
fn atomic_replace(dest: &Path, data: &[u8]) -> Result<(), String> {
    let dir = dest.parent().unwrap_or(Path::new("."));
    let name = dest
        .file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_else(|| "star".into());
    let tmp = dir.join(format!(".{name}.update.tmp"));
    let bak = dir.join(format!(".{name}.old.bak"));

    let _ = fs::remove_file(&tmp);

    let mut f = fs::File::create(&tmp).map_err(|e| format!("could not create temp file: {e}"))?;
    f.write_all(data)
        .map_err(|e| format!("could not write the new binary: {e}"))?;
    f.sync_all()
        .map_err(|e| format!("could not flush the new binary: {e}"))?;
    drop(f);

    // On Windows the binary is locked while it runs, so the previous one has
    // to be moved out of the way before the new one can take its name.
    let had_previous = dest.exists();
    if had_previous {
        let _ = fs::remove_file(&bak);
        fs::rename(dest, &bak).map_err(|e| {
            let _ = fs::remove_file(&tmp);
            format!("could not replace the running star-core: {e}")
        })?;
    }

    let mut renamed = false;
    let mut last_error = String::new();
    for attempt in 0..10u64 {
        match fs::rename(&tmp, dest) {
            Ok(()) => {
                renamed = true;
                break;
            }
            Err(e) => {
                last_error = e.to_string();
                std::thread::sleep(Duration::from_millis(100 * (attempt + 1)));
            }
        }
    }

    if !renamed {
        // Put the old binary back so the next launch still works.
        if had_previous {
            let _ = fs::rename(&bak, dest);
        }
        let _ = fs::remove_file(&tmp);
        return Err(format!("could not install the new star-core: {last_error}"));
    }

    // Only now is the backup safe to drop. This is best-effort: when `dest` is
    // the launcher we are running from, Windows still has the moved-aside copy
    // mapped and the delete fails. `sweep_stale_backups` collects it on the
    // next launch, once this process is gone.
    let _ = fs::remove_file(&bak);
    Ok(())
}

/// Delete backups left behind by an earlier update.
///
/// Windows cannot unlink a running image, so when we replace the launcher we
// rename it aside and leave the renamed copy in place. Once this process has
// exited that file is just litter, and the next launch is the first moment we
/// know it is safe to remove.
pub fn sweep_stale_backups() {
    let Ok(exe) = std::env::current_exe() else {
        return;
    };
    let Some(dir) = exe.parent() else {
        return;
    };
    sweep_backups_in(dir);
}

fn sweep_backups_in(dir: &Path) {
    for name in [launcher_binary_name(), core_binary_name()] {
        let _ = fs::remove_file(dir.join(format!(".{name}.old.bak")));
    }
}
/// Run the full self-update flow: fetch the latest release, download the
/// platform archive, and replace both binaries.
///
/// `launcher_path` is our own executable. Replacing it is best-effort -- an
/// install directory the user cannot write to should not fail an update that
/// has already landed -- but when it does fail they are told, because their
/// splash will otherwise stay stale with no explanation.
pub fn perform_update(core_path: &Path, launcher_path: Option<&Path>) -> Result<(), String> {
    eprintln!("Checking for updates...");
    let url = fetch_asset_url()?;
    eprintln!("Downloading update...");
    let archive = download(&url)?;
    eprintln!("Installing update...");

    let files = extract_all(&archive)?;

    // The core is the binary the user works in: if this fails, the update did
    // not happen and the caller needs to say so.
    install(core_path, &files.core, "star-core")?;

    match (&files.launcher, launcher_path) {
        (Some(bytes), Some(path)) => {
            if let Err(error) = install(path, bytes, "the launcher") {
                eprintln!("  warning: star-core updated, but the launcher did not:");
                eprintln!("    {error}");
                eprintln!("    the splash stays as it is until Star is reinstalled");
            }
        }
        (None, _) => {
            eprintln!(
                "  warning: this release ships no launcher binary; the splash will not change"
            );
        }
        (_, None) => {
            eprintln!(
                "  warning: could not work out where the launcher lives; the splash will not change"
            );
        }
    }

    // The swap only takes effect on the next run: this process is still the
    // old binary, and on Windows it has to be for the rename to work at all.
    eprintln!("Update installed. The new splash appears next time you start star.");
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    /// A throwaway directory that cleans itself up, so a failing assertion
    /// cannot leave binaries lying around in the temp dir.
    struct TempDir(PathBuf);

    impl TempDir {
        fn new(tag: &str) -> Self {
            let n = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_nanos();
            let path = std::env::temp_dir().join(format!("star-upd-{tag}-{n}"));
            fs::create_dir_all(&path).unwrap();
            TempDir(path)
        }
        fn join(&self, name: &str) -> PathBuf {
            self.0.join(name)
        }
        fn join_owned(&self, name: String) -> PathBuf {
            self.0.join(name)
        }
    }

    impl Drop for TempDir {
        fn drop(&mut self) {
            let _ = fs::remove_dir_all(&self.0);
        }
    }

    fn on_unix() -> bool {
        cfg!(unix)
    }

    #[test]
    fn install_writes_the_new_bytes_and_marks_them_executable() {
        let dir = TempDir::new("install");
        let dest = dir.join("star-core");
        fs::write(&dest, b"old").unwrap();

        install(&dest, b"new", "star-core").unwrap();

        assert_eq!(fs::read(&dest).unwrap(), b"new");
        if on_unix() {
            use std::os::unix::fs::PermissionsExt;
            let mode = fs::metadata(&dest).unwrap().permissions().mode();
            assert_eq!(mode & 0o111, 0o111, "installed binary must stay executable");
        }
    }

    #[test]
    fn install_leaves_no_temp_files_behind() {
        let dir = TempDir::new("clean");
        let dest = dir.join("star-core");
        fs::write(&dest, b"old").unwrap();

        install(&dest, b"new", "star-core").unwrap();

        let leftovers: Vec<String> = fs::read_dir(&dir.0)
            .unwrap()
            .flatten()
            .map(|e| e.file_name().to_string_lossy().into_owned())
            .filter(|n| n != "star-core")
            .collect();
        assert!(leftovers.is_empty(), "left litter behind: {leftovers:?}");
    }

    /// The two binaries live side by side, so their temp and backup names must
    /// not collide -- otherwise installing the launcher would clobber the core
    /// (or the other way round) part-way through.
    #[test]
    fn core_and_launcher_can_be_replaced_in_the_same_directory() {
        let dir = TempDir::new("both");
        let core = dir.join(core_binary_name());
        let launcher = dir.join(launcher_binary_name());
        fs::write(&core, b"old core").unwrap();
        fs::write(&launcher, b"old launcher").unwrap();

        install(&core, b"new core", "star-core").unwrap();
        install(&launcher, b"new launcher", "the launcher").unwrap();

        assert_eq!(fs::read(&core).unwrap(), b"new core");
        assert_eq!(fs::read(&launcher).unwrap(), b"new launcher");
    }

    /// The feature depends on being able to overwrite the launcher while the
    /// launcher is the thing doing the overwriting. On Unix that is legal --
    /// the running process keeps the old inode -- and this is the only place
    /// that assumption gets checked, so check it directly.
    #[cfg(unix)]
    #[test]
    fn the_launcher_can_be_replaced_while_it_is_still_running() {
        let Some(sleep) = ["/bin/sleep", "/usr/bin/sleep"]
            .into_iter()
            .find(|p| Path::new(p).exists())
        else {
            eprintln!("skipping: no sleep binary to stand in for the launcher");
            return;
        };

        let dir = TempDir::new("self");
        let dest = dir.join(launcher_binary_name());
        fs::copy(sleep, &dest).unwrap();
        let original = fs::read(&dest).unwrap();

        // A stand-in "launcher" that is definitely running while we replace it.
        let mut child = std::process::Command::new(&dest)
            .arg("30")
            .spawn()
            .expect("spawn the stand-in launcher");

        install(&dest, b"#!/bin/sh\nexit 0\n", "the launcher").unwrap();

        assert_ne!(
            fs::read(&dest).unwrap(),
            original,
            "the launcher on disk was not replaced"
        );
        assert!(
            matches!(child.try_wait(), Ok(None)),
            "the running process should survive its own binary being replaced"
        );

        let _ = child.kill();
        let _ = child.wait();
    }

    #[test]
    fn sweeping_removes_backups_from_both_binaries() {
        let dir = TempDir::new("sweep");
        let core_bak = dir.join_owned(format!(".{}.old.bak", core_binary_name()));
        let launcher_bak = dir.join_owned(format!(".{}.old.bak", launcher_binary_name()));
        fs::write(&core_bak, b"stale").unwrap();
        fs::write(&launcher_bak, b"stale").unwrap();
        let keep = dir.join("star-core");
        fs::write(&keep, b"real binary").unwrap();

        sweep_backups_in(&dir.0);

        assert!(!core_bak.exists(), "core backup survived the sweep");
        assert!(!launcher_bak.exists(), "launcher backup survived the sweep");
        assert!(keep.exists(), "the sweep must not touch real binaries");
    }

    #[cfg(not(target_os = "windows"))]
    mod archive {
        use super::*;
        use std::io::Write;

        /// Build a .tar.gz shaped like a real release archive.
        fn release_archive(entries: &[(&str, &[u8])]) -> Vec<u8> {
            let mut builder = tar::Builder::new(Vec::new());
            for (path, body) in entries {
                let mut header = tar::Header::new_gnu();
                header.set_size(body.len() as u64);
                header.set_mode(0o755);
                header.set_cksum();
                builder
                    .append_data(&mut header, path, *body)
                    .expect("append entry");
            }
            let tar = builder.into_inner().expect("finish tar");
            let mut gz = flate2::write::GzEncoder::new(Vec::new(), flate2::Compression::fast());
            gz.write_all(&tar).expect("gzip");
            gz.finish().expect("finish gzip")
        }

        #[test]
        fn both_binaries_are_pulled_out_of_the_release_archive() {
            // This is the regression that matters: the old code returned on the
            // first match, so `star-core` was extracted and `star` was never
            // even looked at -- which is why the splash never updated.
            let archive = release_archive(&[
                ("star-linux-amd64/star-core", b"CORE"),
                ("star-linux-amd64/star", b"LAUNCHER"),
            ]);

            let files = extract_all(&archive).unwrap();

            assert_eq!(files.core, b"CORE");
            assert_eq!(files.launcher.as_deref(), Some(&b"LAUNCHER"[..]));
        }

        #[test]
        fn an_archive_without_the_launcher_is_not_an_error() {
            // An older or mispackaged release should still update the core
            // rather than failing outright.
            let archive = release_archive(&[("star-linux-amd64/star-core", b"CORE")]);

            let files = extract_all(&archive).unwrap();

            assert_eq!(files.core, b"CORE");
            assert_eq!(files.launcher, None);
        }

        #[test]
        fn an_archive_without_the_core_is_an_error() {
            let archive = release_archive(&[("star-linux-amd64/star", b"LAUNCHER")]);

            let error = extract_all(&archive).unwrap_err();

            assert!(
                error.contains(core_binary_name()),
                "error should name the missing core, got: {error}"
            );
        }

        #[test]
        fn a_binary_nested_at_the_archive_root_is_still_found() {
            // Releases nest under a directory today; do not silently depend on
            // that if someone ever repacks the archive flat.
            let archive = release_archive(&[(core_binary_name(), b"CORE")]);

            assert_eq!(extract_all(&archive).unwrap().core, b"CORE");
        }

        #[test]
        fn a_file_whose_name_merely_starts_with_core_is_not_mistaken_for_it() {
            let archive = release_archive(&[
                ("star-linux-amd64/star-core.sig", b"SIGNATURE"),
                ("star-linux-amd64/star-core", b"CORE"),
            ]);

            assert_eq!(extract_all(&archive).unwrap().core, b"CORE");
        }
    }
}
