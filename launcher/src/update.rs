//! Self-update: download the latest release from GitHub, extract the new
//! `star-core` binary, and replace the one sitting next to us.

use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};
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

/// The binary name we need to extract from the archive.
fn core_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "star-core.exe"
    } else {
        "star-core"
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

/// Extract the core binary from a .zip archive (Windows).
#[cfg(target_os = "windows")]
fn extract_core(archive_bytes: &[u8], dest: &Path) -> Result<(), String> {
    use std::io::Cursor;
    let reader = Cursor::new(archive_bytes);
    let mut archive =
        zip::ZipArchive::new(reader).map_err(|e| format!("could not open the downloaded archive: {e}"))?;

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
            atomic_replace(dest, &buf)?;
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
            atomic_replace(dest, &buf)?;
            // Make executable on Unix.
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(dest, fs::Permissions::from_mode(0o755))
                    .map_err(|e| format!("could not make star-core executable: {e}"))?;
            }
            return Ok(());
        }
    }
    Err(format!("{core_name} was not in the downloaded archive"))
}

/// Write bytes to a file atomically: write to a temp file next to `dest`, then
/// rename over the target. The previous binary is kept until the rename
/// succeeds so a failure here cannot leave the user with no `star-core` at all.
fn atomic_replace(dest: &Path, data: &[u8]) -> Result<(), String> {
    let dir = dest.parent().unwrap_or(Path::new("."));
    let tmp = dir.join(".star-core-update.tmp");
    let bak = dir.join(".star-core-old.bak");

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

    // Only now is the backup safe to drop.
    let _ = fs::remove_file(&bak);
    Ok(())
}

/// Run the full self-update flow: fetch the latest release, download the
/// platform archive, extract the core binary, and replace it on disk.
pub fn perform_update(core_path: &Path) -> Result<PathBuf, String> {
    eprintln!("Checking for updates...");
    let url = fetch_asset_url()?;
    eprintln!("Downloading update...");
    let archive = download(&url)?;
    eprintln!("Installing update...");
    extract_core(&archive, core_path)?;
    eprintln!("Update installed.");
    Ok(core_path.to_path_buf())
}
