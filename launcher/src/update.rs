//! Self-update: download the latest release from GitHub, extract the new
//! `star-core` binary into a versioned directory, and atomically point the
//! launcher at it.
//!
//! Running cores keep the file they were started from. Installing a new
//! version never overwrites that file, so other Star windows can stay open
//! while new windows start on the new core.
//!
//! Layout:
//!
//! ```text
//! ~/bin/star-core.exe                 legacy fallback
//! ~/bin/current                       active version, e.g. "v0.1.95"
//! ~/bin/versions/v0.1.95/star-core.exe
//! ```
//!
//! Cleanup keeps the current version and one previous version. A version that
//! is still in use may not be removable on Windows, in which case the next
//! cleanup attempt tries again.

use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};
use std::time::{Duration, SystemTime};

pub(crate) const GITHUB_RELEASE_LATEST: &str =
    "https://api.github.com/repos/genlabsai/star/releases/latest";
const USER_AGENT: &str = "star-launcher";
const CONNECT_TIMEOUT: Duration = Duration::from_secs(10);
const CALL_TIMEOUT: Duration = Duration::from_secs(120);
/// Transient network failures are common enough that a single failed attempt
/// should not surface to the user as "update failed".
const ATTEMPTS: u32 = 3;
const VERSIONS_DIR: &str = "versions";
const CURRENT_FILE: &str = "current";
const STAGING_PREFIX: &str = ".staging-";
const STAGING_STALE: Duration = Duration::from_secs(60 * 60);
const KEEP_VERSIONS: usize = 2;

/// The user's home directory, best effort.
pub(crate) fn home_dir() -> PathBuf {
    std::env::var_os("USERPROFILE")
        .filter(|value| !value.is_empty())
        .or_else(|| std::env::var_os("HOME").filter(|value| !value.is_empty()))
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| ".".into())
}

/// The directory containing the launcher and its core versions.
pub fn install_dir() -> PathBuf {
    home_dir().join("bin")
}

fn core_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "star-core.exe"
    } else {
        "star-core"
    }
}

fn versions_dir(install_dir: &Path) -> PathBuf {
    install_dir.join(VERSIONS_DIR)
}

fn current_path(install_dir: &Path) -> PathBuf {
    install_dir.join(CURRENT_FILE)
}

/// Resolve the core the next launch should use.
///
/// The versioned path wins when the pointer is valid and the file exists. The
/// legacy root binary remains the fallback for installs created before the
/// versioned layout, or if the pointer is missing or corrupt.
pub fn find_core() -> PathBuf {
    find_core_in(&install_dir())
}

pub fn find_core_in(install_dir: &Path) -> PathBuf {
    if let Some(version) = read_current(install_dir) {
        let candidate = versions_dir(install_dir)
            .join(&version)
            .join(core_binary_name());
        if candidate.is_file() {
            return candidate;
        }
    }
    install_dir.join(core_binary_name())
}

fn read_current(install_dir: &Path) -> Option<String> {
    let raw = fs::read_to_string(current_path(install_dir)).ok()?;
    let version = raw.trim();
    if safe_version(version) {
        Some(version.to_string())
    } else {
        None
    }
}

pub(crate) fn safe_version(version: &str) -> bool {
    !version.is_empty()
        && version.len() <= 64
        && version != "."
        && version != ".."
        && version
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '.' | '-' | '_'))
}

fn write_current(install_dir: &Path, version: &str) -> Result<(), String> {
    fs::create_dir_all(install_dir)
        .map_err(|e| format!("could not create install directory: {e}"))?;
    let tmp = install_dir.join(".current.tmp");
    let mut file =
        fs::File::create(&tmp).map_err(|e| format!("could not create version pointer: {e}"))?;
    file.write_all(version.as_bytes())
        .map_err(|e| format!("could not write version pointer: {e}"))?;
    file.sync_all()
        .map_err(|e| format!("could not flush version pointer: {e}"))?;
    drop(file);

    // Unix replaces atomically. Windows' MoveFileEx-backed rename also replaces
    // an existing file; the fallback covers filesystems that refuse that.
    if fs::rename(&tmp, current_path(install_dir)).is_err() {
        let _ = fs::remove_file(current_path(install_dir));
        fs::rename(&tmp, current_path(install_dir))
            .map_err(|e| format!("could not activate the new version: {e}"))?;
    }
    Ok(())
}

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

pub(crate) fn request(url: &str) -> Result<ureq::http::Response<ureq::Body>, String> {
    agent(CALL_TIMEOUT)
        .get(url)
        .header("Accept", "application/vnd.github.v3+json")
        .call()
        .map_err(|e| format!("{e}"))
}

/// Retry `op` a few times, sleeping between attempts. The last error is
/// returned if every attempt fails.
pub(crate) fn retry<T, F>(mut op: F) -> Result<T, String>
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
                    eprintln!("  retrying ({attempt}/{ATTEMPTS})...");
                    std::thread::sleep(Duration::from_millis(500 * attempt as u64));
                }
            }
        }
    }
    Err(last)
}

struct Release {
    version: String,
    asset_url: String,
}

/// Find an asset URL in the release JSON without pulling in a full JSON parser.
///
/// GitHub writes `"name"` before `"browser_download_url"` in each asset
/// object. Splitting at the name key keeps the two fields together. The text
/// after the key starts with a colon, so consume that before reading the value.
fn asset_url_from_release(body: &str, wanted: &str) -> Option<String> {
    for block in body.split("\"name\"").skip(1) {
        let Some(rest) = block.trim_start().strip_prefix(':') else {
            continue;
        };
        let Some(rest) = rest.trim_start().strip_prefix('"') else {
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
        return Some(after[..end].to_string());
    }
    None
}

/// Fetch the latest release and its platform asset. The version comes from the
/// asset URL (`/releases/download/<tag>/<asset>`), so we do not need a JSON
/// parser just to learn the tag.
fn fetch_release() -> Result<Release, String> {
    let body: String = retry(|| {
        let resp = request(GITHUB_RELEASE_LATEST)?;
        resp.into_body()
            .read_to_string()
            .map_err(|e| format!("could not read release metadata: {e}"))
    })?;

    let wanted = asset_name();
    let asset_url = asset_url_from_release(&body, wanted)
        .ok_or_else(|| format!("release has no {wanted} asset for this platform"))?;
    let version = version_from_asset_url(&asset_url)
        .ok_or_else(|| format!("could not determine the release version from {asset_url}"))?;
    if !safe_version(&version) {
        return Err(format!(
            "release version is not safe to use as a directory: {version}"
        ));
    }
    Ok(Release { version, asset_url })
}

fn version_from_asset_url(url: &str) -> Option<String> {
    let rest = url.split_once("/download/")?.1;
    let version = rest.split('/').next()?;
    if safe_version(version) {
        Some(version.to_string())
    } else {
        None
    }
}

/// Download bytes from a URL.
pub(crate) fn download(url: &str) -> Result<Vec<u8>, String> {
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

/// Extract the core binary from a .zip archive (Windows).
#[cfg(target_os = "windows")]
fn extract_core(archive_bytes: &[u8], dest: &Path) -> Result<(), String> {
    use std::io::Cursor;
    let reader = Cursor::new(archive_bytes);
    let mut archive = zip::ZipArchive::new(reader)
        .map_err(|e| format!("could not open the downloaded archive: {e}"))?;

    let core_name = core_binary_name();
    for i in 0..archive.len() {
        let mut file = archive
            .by_index(i)
            .map_err(|e| format!("could not read archive entry: {e}"))?;
        let name = file.name().replace('\\', "/");
        if name.ends_with(core_name) {
            let mut buf = Vec::new();
            file.read_to_end(&mut buf)
                .map_err(|e| format!("could not read {core_name} from the archive: {e}"))?;
            write_core(dest, &buf)?;
            return Ok(());
        }
    }
    Err(format!("{core_name} was not in the downloaded archive"))
}

/// Extract the core binary from a .tar.gz archive (Unix).
#[cfg(not(target_os = "windows"))]
fn extract_core(archive_bytes: &[u8], dest: &Path) -> Result<(), String> {
    use std::io::Cursor;
    let gz = flate2::read::GzDecoder::new(Cursor::new(archive_bytes));
    let mut archive = tar::Archive::new(gz);
    let core_name = core_binary_name();

    for entry in archive
        .entries()
        .map_err(|e| format!("could not read the downloaded archive: {e}"))?
    {
        let mut entry = entry.map_err(|e| format!("could not read archive entry: {e}"))?;
        let path = entry
            .path()
            .map_err(|e| format!("bad path in archive: {e}"))?
            .to_path_buf();
        if path.file_name().map(|n| n == core_name).unwrap_or(false) {
            let mut buf = Vec::new();
            entry
                .read_to_end(&mut buf)
                .map_err(|e| format!("could not read {core_name} from the archive: {e}"))?;
            write_core(dest, &buf)?;
            return Ok(());
        }
    }
    Err(format!("{core_name} was not in the downloaded archive"))
}

fn write_core(dest: &Path, data: &[u8]) -> Result<(), String> {
    let mut file =
        fs::File::create(dest).map_err(|e| format!("could not create the new core: {e}"))?;
    file.write_all(data)
        .map_err(|e| format!("could not write the new binary: {e}"))?;
    file.sync_all()
        .map_err(|e| format!("could not flush the new binary: {e}"))?;
    drop(file);

    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(dest, fs::Permissions::from_mode(0o755))
            .map_err(|e| format!("could not make star-core executable: {e}"))?;
    }
    Ok(())
}

/// Install the latest core version and atomically make it the launch target.
///
/// The archive is extracted into a staging directory first, then the complete
/// directory is renamed into `versions/<tag>`. A crash leaves either the old
/// pointer or a staging directory, never a half-written active core.
pub fn perform_update(install_dir: &Path) -> Result<PathBuf, String> {
    eprintln!("Checking for updates...");
    let release = fetch_release()?;
    let dest_dir = versions_dir(install_dir).join(&release.version);
    let dest = dest_dir.join(core_binary_name());

    if dest.is_file() {
        // The version is already downloaded. A previous attempt may have
        // stopped after extraction but before publishing the pointer.
        write_current(install_dir, &release.version)?;
        cleanup_install(install_dir);
        eprintln!("Update installed.");
        return Ok(dest);
    }

    eprintln!("Downloading update...");
    let archive = download(&release.asset_url)?;
    install_archive(install_dir, &release.version, &archive)
}

/// Install an already-downloaded archive into a versioned directory.
///
/// Keeping this separate from the network call makes the promotion path
/// testable with a local archive and, importantly, keeps the current version
/// untouched until the complete new directory is ready.
fn install_archive(install_dir: &Path, version: &str, archive: &[u8]) -> Result<PathBuf, String> {
    if !safe_version(version) {
        return Err(format!(
            "release version is not safe to use as a directory: {version}"
        ));
    }
    eprintln!("Installing update...");

    let versions = versions_dir(install_dir);
    let dest_dir = versions.join(version);
    let dest = dest_dir.join(core_binary_name());

    fs::create_dir_all(&versions)
        .map_err(|e| format!("could not create versions directory: {e}"))?;
    let staging = versions.join(format!("{STAGING_PREFIX}{version}-{}", std::process::id()));
    let _ = fs::remove_dir_all(&staging);
    fs::create_dir_all(&staging).map_err(|e| format!("could not create staging directory: {e}"))?;

    let staging_core = staging.join(core_binary_name());
    if let Err(error) = extract_core(archive, &staging_core) {
        let _ = fs::remove_dir_all(&staging);
        return Err(error);
    }

    if dest_dir.exists() {
        // A partial directory from an interrupted attempt has no usable core.
        let _ = fs::remove_dir_all(&dest_dir);
    }

    let mut renamed = false;
    let mut last_error = String::new();
    for attempt in 0..10u64 {
        match fs::rename(&staging, &dest_dir) {
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
        let _ = fs::remove_dir_all(&staging);
        return Err(format!(
            "could not activate version {version}: {last_error}"
        ));
    }

    write_current(install_dir, version)?;
    cleanup_install(install_dir);
    eprintln!("Update installed.");
    Ok(dest)
}

/// Remove old versions and stale staging/backup files.
///
/// Deletion is deliberately best-effort: on Windows a core that is still
/// running cannot be deleted, and that is correct. It will be picked up by
/// cleanup on a later launch. Only the current version and one previous version
/// are retained, so inactive downloads cannot accumulate.
pub fn cleanup_install(install_dir: &Path) {
    let versions = versions_dir(install_dir);
    let current = read_current(install_dir);

    let mut candidates: Vec<(String, PathBuf, SystemTime)> = Vec::new();
    if let Ok(entries) = fs::read_dir(&versions) {
        for entry in entries.flatten() {
            let Ok(file_type) = entry.file_type() else {
                continue;
            };
            if !file_type.is_dir() {
                continue;
            }
            let name = entry.file_name().to_string_lossy().to_string();
            let modified = entry
                .metadata()
                .and_then(|m| m.modified())
                .unwrap_or(SystemTime::UNIX_EPOCH);

            if name.starts_with(STAGING_PREFIX) {
                if modified.elapsed().unwrap_or_default() > STAGING_STALE {
                    let _ = fs::remove_dir_all(entry.path());
                }
                continue;
            }

            if safe_version(&name) {
                candidates.push((name, entry.path(), modified));
            }
        }
    }

    // Newest first. Keep the active version even if its mtime is older, then
    // one previous version for rollback.
    candidates.sort_by_key(|a| std::cmp::Reverse(a.2));
    let mut keep: Vec<String> = Vec::new();
    if let Some(version) = current {
        keep.push(version);
    }
    for (name, _, _) in &candidates {
        if keep.len() >= KEEP_VERSIONS {
            break;
        }
        if !keep.contains(name) {
            keep.push(name.clone());
        }
    }

    for (name, path, _) in candidates {
        if keep.contains(&name) {
            continue;
        }
        let _ = fs::remove_dir_all(path);
    }

    // Files left by the old in-place updater.
    let _ = fs::remove_file(install_dir.join(".star-core-update.tmp"));
    let _ = fs::remove_file(install_dir.join(".star-core-old.bak"));
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    fn temp_dir(tag: &str) -> PathBuf {
        let path = std::env::temp_dir().join(format!(
            "star-version-test-{tag}-{}-{:?}",
            std::process::id(),
            std::thread::current().id()
        ));
        let _ = fs::remove_dir_all(&path);
        fs::create_dir_all(&path).expect("create temp dir");
        path
    }

    fn add_version(install: &Path, version: &str) {
        let dir = versions_dir(install).join(version);
        fs::create_dir_all(&dir).expect("create version dir");
        fs::write(dir.join(core_binary_name()), version.as_bytes()).expect("write core");
        std::thread::sleep(Duration::from_millis(20));
    }

    #[test]
    fn current_pointer_selects_versioned_core() {
        let install = temp_dir("current");
        add_version(&install, "v0.1.94");
        add_version(&install, "v0.1.95");
        write_current(&install, "v0.1.95").expect("write current");

        assert_eq!(
            find_core_in(&install),
            versions_dir(&install)
                .join("v0.1.95")
                .join(core_binary_name())
        );
        let _ = fs::remove_dir_all(install);
    }

    #[test]
    fn missing_or_unsafe_pointer_falls_back_to_legacy_core() {
        let install = temp_dir("fallback");
        fs::write(install.join(core_binary_name()), b"legacy").expect("write legacy");

        assert_eq!(find_core_in(&install), install.join(core_binary_name()));
        fs::write(current_path(&install), b"../../escape").expect("write unsafe pointer");
        assert_eq!(find_core_in(&install), install.join(core_binary_name()));
        let _ = fs::remove_dir_all(install);
    }

    #[test]
    fn cleanup_keeps_current_and_one_previous_version() {
        let install = temp_dir("cleanup");
        add_version(&install, "v0.1.91");
        add_version(&install, "v0.1.92");
        add_version(&install, "v0.1.93");
        write_current(&install, "v0.1.92").expect("write current");

        cleanup_install(&install);

        assert!(versions_dir(&install).join("v0.1.92").is_dir());
        assert!(versions_dir(&install).join("v0.1.93").is_dir());
        assert!(!versions_dir(&install).join("v0.1.91").is_dir());
        let _ = fs::remove_dir_all(install);
    }

    #[test]
    fn cleanup_never_deletes_the_current_version() {
        let install = temp_dir("keep-current");
        add_version(&install, "v0.1.90");
        add_version(&install, "v0.1.91");
        add_version(&install, "v0.1.92");
        write_current(&install, "v0.1.90").expect("write current");

        cleanup_install(&install);

        assert!(versions_dir(&install).join("v0.1.90").is_dir());
        assert!(versions_dir(&install).join("v0.1.92").is_dir());
        assert!(!versions_dir(&install).join("v0.1.91").is_dir());
        let _ = fs::remove_dir_all(install);
    }

    #[cfg(not(target_os = "windows"))]
    fn test_archive(contents: &[u8]) -> Vec<u8> {
        use flate2::Compression;
        use flate2::write::GzEncoder;
        use tar::{Builder, Header};

        let gz = GzEncoder::new(Vec::new(), Compression::default());
        let mut builder = Builder::new(gz);
        let mut header = Header::new_gnu();
        header
            .set_path("star-linux-amd64/star-core")
            .expect("archive path");
        header.set_size(contents.len() as u64);
        header.set_mode(0o755);
        header.set_cksum();
        builder
            .append(&header, contents)
            .expect("append archive core");
        builder
            .into_inner()
            .expect("finish archive")
            .finish()
            .expect("finish gzip")
    }

    #[cfg(not(target_os = "windows"))]
    #[test]
    fn archive_install_publishes_pointer_and_bounds_versions() {
        let install = temp_dir("archive");

        install_archive(&install, "v0.1.99", &test_archive(b"one")).expect("install first archive");
        assert_eq!(read_current(&install).as_deref(), Some("v0.1.99"));
        assert!(versions_dir(&install).join("v0.1.99/star-core").is_file());

        std::thread::sleep(Duration::from_millis(20));
        install_archive(&install, "v0.1.100", &test_archive(b"two"))
            .expect("install second archive");
        assert_eq!(read_current(&install).as_deref(), Some("v0.1.100"));
        assert!(versions_dir(&install).join("v0.1.99/star-core").is_file());

        std::thread::sleep(Duration::from_millis(20));
        install_archive(&install, "v0.1.101", &test_archive(b"three"))
            .expect("install third archive");
        assert_eq!(read_current(&install).as_deref(), Some("v0.1.101"));
        assert!(versions_dir(&install).join("v0.1.100/star-core").is_file());
        assert!(versions_dir(&install).join("v0.1.101/star-core").is_file());
        assert!(!versions_dir(&install).join("v0.1.99").exists());
        let _ = fs::remove_dir_all(install);
    }

    #[test]
    fn finds_the_platform_asset_after_its_name() {
        let body = r#"{
            "tag_name": "v0.1.100",
            "name": "v0.1.100",
            "assets": [
                {
                    "name": "checksums.txt",
                    "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.100/checksums.txt"
                },
                {
                    "name": "star-linux-amd64.tar.gz",
                    "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.100/star-linux-amd64.tar.gz"
                }
            ]
        }"#;

        assert_eq!(
            asset_url_from_release(body, "star-linux-amd64.tar.gz").as_deref(),
            Some(
                "https://github.com/GenLabsAI/star/releases/download/v0.1.100/star-linux-amd64.tar.gz"
            )
        );
        assert_eq!(asset_url_from_release(body, "star-windows-amd64.zip"), None);
    }

    #[test]
    fn version_comes_from_the_asset_url() {
        assert_eq!(
            version_from_asset_url(
                "https://github.com/genlabsai/star/releases/download/v0.1.95/star-windows-amd64.zip"
            )
            .as_deref(),
            Some("v0.1.95")
        );
        assert_eq!(
            version_from_asset_url("https://example.com/not-a-release"),
            None
        );
        assert_eq!(
            version_from_asset_url("https://x/download/../evil/file"),
            None
        );
    }
}
