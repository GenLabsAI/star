//! Self-update logic: download the latest release from GitHub, extract
//! the new star-core binary, and replace the one sitting next to us.

use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

const GITHUB_RELEASE_LATEST: &str = "https://api.github.com/repos/genlabsai/star/releases/latest";

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
        "star"
    }
}

/// Fetch the browser_download_url for our platform asset from the latest
/// GitHub release.
fn fetch_asset_url() -> Result<String, String> {
    let resp = ureq::get(GITHUB_RELEASE_LATEST)
        .header("Accept", "application/vnd.github.v3+json")
        .header("User-Agent", "star-launcher/1.0")
        .call()
        .map_err(|e| format!("GitHub API request failed: {e}"))?;
    let body: String = resp
        .into_body()
        .read_to_string()
        .map_err(|e| format!("Failed to read response body: {e}"))?;

    let wanted = asset_name();

    // Minimal JSON parsing: find the asset whose name matches and grab
    // its browser_download_url. We avoid pulling in serde_json to keep
    // the binary tiny.
    for asset_block in body.split("\"name\"") {
        if let Some(name_start) = asset_block.find('"') {
            let rest = &asset_block[name_start + 1..];
            if let Some(name_end) = rest.find('"') {
                let name = &rest[..name_end];
                if name == wanted {
                    // Find browser_download_url in this block
                    if let Some(url_key) = asset_block.find("\"browser_download_url\"") {
                        let after_key = &asset_block[url_key + "\"browser_download_url\"".len()..];
                        // Skip optional whitespace and colon
                        let after_colon = after_key
                            .trim_start()
                            .strip_prefix(':')
                            .unwrap_or(after_key)
                            .trim_start();
                        if let Some(url_start) = after_colon.find('"') {
                            let url_rest = &after_colon[url_start + 1..];
                            if let Some(url_end) = url_rest.find('"') {
                                return Ok(url_rest[..url_end].to_string());
                            }
                        }
                    }
                }
            }
        }
    }

    Err(format!("Asset {wanted} not found in latest release"))
}

/// Download bytes from a URL.
fn download(url: &str) -> Result<Vec<u8>, String> {
    let resp = ureq::get(url)
        .header("User-Agent", "star-launcher/1.0")
        .call()
        .map_err(|e| format!("Download failed: {e}"))?;
    let mut buf = Vec::new();
    resp.into_body()
        .into_reader()
        .read_to_end(&mut buf)
        .map_err(|e| format!("Failed to read download body: {e}"))?;
    Ok(buf)
}

/// Extract the core binary from a .zip archive (Windows).
#[cfg(target_os = "windows")]
fn extract_core(archive_bytes: &[u8], dest: &Path) -> Result<(), String> {
    use std::io::Cursor;
    let reader = Cursor::new(archive_bytes);
    let mut archive =
        zip::ZipArchive::new(reader).map_err(|e| format!("Failed to open zip: {e}"))?;

    let core_name = core_binary_name();
    for i in 0..archive.len() {
        let mut file = archive
            .by_index(i)
            .map_err(|e| format!("Failed to read zip entry: {e}"))?;
        let name = file.name().replace('\\', "/");
        if name.ends_with(core_name) {
            let mut buf = Vec::new();
            file.read_to_end(&mut buf)
                .map_err(|e| format!("Failed to read {core_name} from zip: {e}"))?;
            atomic_replace(dest, &buf)?;
            return Ok(());
        }
    }
    Err(format!("{core_name} not found in archive"))
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
        .map_err(|e| format!("Failed to read tar: {e}"))?
    {
        let mut entry = entry.map_err(|e| format!("Failed to read tar entry: {e}"))?;
        let path = entry
            .path()
            .map_err(|e| format!("Failed to get entry path: {e}"))?
            .to_path_buf();
        if path.file_name().map(|n| n == core_name).unwrap_or(false) {
            let mut buf = Vec::new();
            entry
                .read_to_end(&mut buf)
                .map_err(|e| format!("Failed to read {core_name} from tar: {e}"))?;
            atomic_replace(dest, &buf)?;
            // Make executable on Unix
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(dest, fs::Permissions::from_mode(0o755))
                    .map_err(|e| format!("Failed to set permissions: {e}"))?;
            }
            return Ok(());
        }
    }
    Err(format!("{core_name} not found in archive"))
}

/// Write bytes to a file atomically: write to a temp file next to dest,
/// then rename over the target.
fn atomic_replace(dest: &Path, data: &[u8]) -> Result<(), String> {
    let dir = dest.parent().unwrap_or(Path::new("."));
    let tmp = dir.join(".star-core-update.tmp");

    // On Windows the running binary is locked; the Go core has already
    // exited by the time we get here, so the file should be unlocked.
    // If it is still locked, try renaming the old binary out of the way
    // first.
    if dest.exists() {
        let bak = dir.join(".star-core-old.bak");
        let _ = fs::remove_file(&bak);
        if let Err(e) = fs::rename(dest, &bak) {
            // Non-fatal on Unix; fatal on Windows if the file is locked.
            eprintln!("warning: could not move old binary aside: {e}");
        }
    }

    let mut f = fs::File::create(&tmp).map_err(|e| format!("Failed to create temp file: {e}"))?;
    f.write_all(data)
        .map_err(|e| format!("Failed to write temp file: {e}"))?;
    f.sync_all()
        .map_err(|e| format!("Failed to sync temp file: {e}"))?;
    drop(f);

    let mut retries = 0;
    loop {
        match fs::rename(&tmp, dest) {
            Ok(_) => break,
            Err(e) => {
                if retries >= 10 {
                    return Err(format!("Failed to rename temp to {}: {e}", dest.display()));
                }
                std::thread::sleep(std::time::Duration::from_millis(500));
                retries += 1;
            }
        }
    }

    Ok(())
}

/// Run the full self-update flow: fetch the latest release, download the
/// platform archive, extract the core binary, and replace it on disk.
/// Returns the path to the updated core binary.
pub fn perform_update(core_path: &Path) -> Result<PathBuf, String> {
    eprintln!("Checking for updates...");
    let url = fetch_asset_url()?;
    eprintln!("Downloading update...");
    let archive = download(&url)?;
    eprintln!("Installing update...");
    extract_core(&archive, core_path)?;
    eprintln!("Update complete.");
    Ok(core_path.to_path_buf())
}
