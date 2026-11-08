//! Star Desktop bootstrap: keep the desktop app installed without clicks.
//!
//! Runs on a background thread at startup. It never blocks the core and never
//! fails the launch: any error is reported as a single line and forgotten.
//! Installs are per-user (no admin needed) and version-tracked with a marker
//! file, so a present, current install costs one fast GitHub API call and
//! nothing else.
//!
//! Per-OS strategy (silent, no elevation):
//!
//! - Windows: NSIS `-setup.exe` with `/S /D=<LocalAppData>\Programs\…`,
//!   falling back to `msiexec /qn` per-user for `.msi`. The launcher also
//!   owns the Start Menu shortcut itself, since silent installs do not
//!   reliably leave one behind.
//! - macOS: `.app.tar.gz` extracted straight into `~/Applications`
//!   (no installer UI at all), falling back to a quiet `hdiutil` `.dmg`
//!   mount-and-copy.
//! - Linux: `.AppImage` saved executable into `~/bin` plus a best-effort
//!   `.desktop` entry. (A `.deb` needs root, so it is never attempted.)

use std::fs;
use std::path::{Path, PathBuf};

use crate::update;

/// Marker file recording the installed desktop version (a release tag).
const MARKER_FILE: &str = "desktop-version";
const APP_NAME: &str = "Star Desktop";

/// Kick off the desktop ensure on a background thread. Never blocks.
pub fn ensure_desktop_async() {
    std::thread::spawn(|| {
        if let Err(error) = ensure_desktop() {
            eprintln!("  star desktop: {error} (continuing without it)");
        }
    });
}

fn ensure_desktop() -> Result<(), String> {
    let install_dir = update::install_dir();
    let (tag, asset_name, asset_url) = latest_desktop_asset()?;
    if marker_version(&install_dir).as_deref() != Some(tag.as_str())
        || !install_present(&install_dir)
    {
        eprintln!("  Setting up Star Desktop {tag} (one-time setup)...");
        install_asset(&install_dir, &asset_name, &asset_url)?;
        write_marker(&install_dir, &tag)?;
        eprintln!("  Star Desktop {tag} is ready.");
    }
    // The silent NSIS install does not reliably leave a Start Menu entry,
    // so the launcher owns the shortcut itself. Cheap once it exists: a
    // single file-exists check on the background thread.
    #[cfg(target_os = "windows")]
    ensure_windows_shortcut()?;
    Ok(())
}

/// Find the best desktop installer asset in the latest release.
/// Returns `(tag, asset name, download url)`.
fn latest_desktop_asset() -> Result<(String, String, String), String> {
    let body: String = update::retry(|| {
        let resp = update::request(update::GITHUB_RELEASE_LATEST)?;
        resp.into_body()
            .read_to_string()
            .map_err(|e| format!("could not read release metadata: {e}"))
    })?;
    pick_asset(&body, asset_suffixes()).ok_or_else(|| {
        "latest release has no desktop installer asset for this platform".to_string()
    })
}

/// Preferred asset filename suffixes, best first, for this platform.
#[cfg(target_os = "windows")]
fn asset_suffixes() -> &'static [&'static str] {
    &["-setup.exe", ".msi"]
}

/// `.app.tar.gz` extracts with no installer UI at all, so it wins over `.dmg`.
#[cfg(target_os = "macos")]
fn asset_suffixes() -> &'static [&'static str] {
    &[".app.tar.gz", ".dmg"]
}

/// `.AppImage` runs from anywhere with no privileges; `.deb` needs root.
#[cfg(target_os = "linux")]
fn asset_suffixes() -> &'static [&'static str] {
    &[".appimage", ".deb"]
}

#[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
fn asset_suffixes() -> &'static [&'static str] {
    &[".appimage"]
}

/// Pick the first asset whose name ends with one of `suffixes` (case
/// insensitive), honouring preference order. Returns
/// `(tag, asset name, download url)`; the tag is parsed from the download URL
/// (`/releases/download/<tag>/<asset>`), the same way the core updater does.
/// Only assets built for this machine's architecture are considered: a
/// wrong-arch install would write the marker and never retry.
fn pick_asset(body: &str, suffixes: &[&str]) -> Option<(String, String, String)> {
    pick_asset_for_arch(body, suffixes, arch_tokens())
}

/// Architecture tokens identifying builds for this machine.
fn arch_tokens() -> &'static [&'static str] {
    #[cfg(target_arch = "x86_64")]
    {
        &["x86_64", "x64", "amd64"]
    }
    #[cfg(target_arch = "aarch64")]
    {
        &["aarch64", "arm64"]
    }
    #[cfg(not(any(target_arch = "x86_64", target_arch = "aarch64")))]
    {
        &[]
    }
}

/// Every architecture token we recognise, for any machine.
const ANY_ARCH_TOKENS: &[&str] = &["x86_64", "x64", "amd64", "aarch64", "arm64"];

/// Does this asset name target this machine? A name carrying our arch token
/// is ours; a name with no arch token at all is treated as universal; a
/// name carrying another arch's token is for someone else.
fn name_matches_arch(name: &str, tokens: &[&str]) -> bool {
    let lowered = name.to_lowercase();
    if tokens.iter().any(|token| lowered.contains(token)) {
        return true;
    }
    !ANY_ARCH_TOKENS.iter().any(|token| lowered.contains(token))
}

fn pick_asset_for_arch(
    body: &str,
    suffixes: &[&str],
    arch: &[&str],
) -> Option<(String, String, String)> {
    let mut found: Vec<(String, String)> = Vec::new();
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
        let name = rest[..end].to_string();
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
        found.push((name, after[..end].to_string()));
    }
    for suffix in suffixes {
        if let Some((name, url)) = found.iter().find(|(name, _)| {
            name.to_lowercase().ends_with(suffix) && name_matches_arch(name, arch)
        }) {
            let rest = url.split_once("/download/")?.1;
            let tag = rest.split('/').next()?;
            if update::safe_version(tag) {
                return Some((tag.to_string(), name.clone(), url.clone()));
            }
        }
    }
    None
}

fn marker_path(install_dir: &Path) -> PathBuf {
    install_dir.join(MARKER_FILE)
}

fn marker_version(install_dir: &Path) -> Option<String> {
    let raw = fs::read_to_string(marker_path(install_dir)).ok()?;
    let version = raw.trim();
    if update::safe_version(version) {
        Some(version.to_string())
    } else {
        None
    }
}

fn write_marker(install_dir: &Path, version: &str) -> Result<(), String> {
    fs::create_dir_all(install_dir)
        .map_err(|e| format!("could not create install directory: {e}"))?;
    fs::write(marker_path(install_dir), version)
        .map_err(|e| format!("could not record desktop version: {e}"))?;
    Ok(())
}

/// Is the desktop app actually on disk where we put it?
fn install_present(install_dir: &Path) -> bool {
    #[cfg(target_os = "windows")]
    {
        let _ = install_dir;
        windows_install_dir()
            .map(|dir| dir_contains_exe(&dir))
            .unwrap_or(false)
    }
    #[cfg(target_os = "macos")]
    {
        let _ = install_dir;
        macos_app_dir().is_dir() || any_star_app()
    }
    #[cfg(target_os = "linux")]
    {
        install_dir.join("star-desktop.AppImage").is_file()
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    {
        let _ = install_dir;
        false
    }
}

#[cfg(target_os = "windows")]
fn local_app_data() -> Option<PathBuf> {
    std::env::var_os("LOCALAPPDATA")
        .filter(|value| !value.is_empty())
        .map(PathBuf::from)
}

#[cfg(target_os = "windows")]
fn windows_install_dir() -> Option<PathBuf> {
    local_app_data().map(|base| base.join("Programs").join(APP_NAME))
}

#[cfg(target_os = "windows")]
fn dir_contains_exe(dir: &Path) -> bool {
    let Ok(entries) = fs::read_dir(dir) else {
        return false;
    };
    entries.flatten().any(|entry| {
        entry
            .path()
            .extension()
            .is_some_and(|ext| ext.eq_ignore_ascii_case("exe"))
    })
}

#[cfg(target_os = "macos")]
fn applications_dir() -> PathBuf {
    update::home_dir().join("Applications")
}

#[cfg(target_os = "macos")]
fn macos_app_dir() -> PathBuf {
    applications_dir().join(format!("{APP_NAME}.app"))
}

/// A manually installed copy still counts: any `*Star*.app` in the usual
/// places means the user already has the desktop app.
#[cfg(target_os = "macos")]
fn any_star_app() -> bool {
    for dir in [applications_dir(), PathBuf::from("/Applications")] {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for entry in entries.flatten() {
            let name = entry.file_name().to_string_lossy().to_lowercase();
            if name.contains("star") && name.ends_with(".app") && entry.path().is_dir() {
                return true;
            }
        }
    }
    false
}

/// Download the asset and install it with no prompts and no elevation.
fn install_asset(install_dir: &Path, asset_name: &str, asset_url: &str) -> Result<(), String> {
    let bytes = update::download(asset_url)?;
    if bytes.is_empty() {
        return Err("desktop download returned no data".to_string());
    }
    let lowered = asset_name.to_lowercase();
    #[cfg(target_os = "windows")]
    {
        let _ = install_dir;
        if lowered.ends_with("-setup.exe") || lowered.ends_with(".exe") {
            return install_nsis(&bytes);
        }
        if lowered.ends_with(".msi") {
            return install_msi(&bytes);
        }
    }
    #[cfg(target_os = "macos")]
    {
        let _ = install_dir;
        if lowered.ends_with(".app.tar.gz") || lowered.ends_with(".tgz") {
            return install_mac_app(&bytes);
        }
        if lowered.ends_with(".dmg") {
            return install_dmg(&bytes);
        }
    }
    #[cfg(target_os = "linux")]
    {
        if lowered.ends_with(".appimage") {
            return install_appimage(install_dir, &bytes);
        }
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    {
        let _ = (install_dir, asset_name);
    }
    Err(format!("no silent install path for {asset_name}"))
}

/// NSIS install, silent and per-user. `/D` must be the last argument.
#[cfg(target_os = "windows")]
fn install_nsis(installer_bytes: &[u8]) -> Result<(), String> {
    let dir = windows_install_dir()
        .ok_or_else(|| "could not locate the local app-data directory".to_string())?;
    let staging =
        std::env::temp_dir().join(format!("star-desktop-setup-{}.exe", std::process::id()));
    fs::write(&staging, installer_bytes)
        .map_err(|e| format!("could not stage the desktop installer: {e}"))?;
    let dir_arg = format!("/D={}", dir.to_string_lossy());
    let status = std::process::Command::new(&staging)
        .arg("/S")
        .arg(&dir_arg)
        .status()
        .map_err(|e| format!("could not run the desktop installer: {e}"))?;
    let _ = fs::remove_file(&staging);
    if !status.success() {
        return Err("desktop installer reported failure".to_string());
    }
    if !dir_contains_exe(&dir) {
        return Err("desktop installer finished but the app is not where expected".to_string());
    }
    Ok(())
}

/// MSI install, silent and per-user (`MSIINSTALLPERUSER` needs no elevation).
#[cfg(target_os = "windows")]
fn install_msi(msi_bytes: &[u8]) -> Result<(), String> {
    let staging = std::env::temp_dir().join(format!("star-desktop-{}.msi", std::process::id()));
    fs::write(&staging, msi_bytes)
        .map_err(|e| format!("could not stage the desktop installer: {e}"))?;
    let status = std::process::Command::new("msiexec")
        .args([
            "/qn",
            "/norestart",
            "/i",
            &staging.to_string_lossy(),
            "MSIINSTALLPERUSER=1",
            "ALLUSERS=2",
        ])
        .status()
        .map_err(|e| format!("could not run msiexec: {e}"))?;
    let _ = fs::remove_file(&staging);
    if !status.success() {
        return Err("desktop installer reported failure".to_string());
    }
    Ok(())
}

/// Make sure Star Desktop is in the Start Menu.
///
/// The silent NSIS install does not reliably create a Start Menu entry, so
/// the launcher owns the shortcut itself:
/// `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Star Desktop\Star Desktop.lnk`.
/// Creating it is idempotent and also repairs installs from before this
/// step existed. An MSI install that already left its own shortcut behind
/// is left alone.
#[cfg(target_os = "windows")]
fn ensure_windows_shortcut() -> Result<(), String> {
    let appdata = std::env::var_os("APPDATA")
        .filter(|value| !value.is_empty())
        .map(PathBuf::from)
        .ok_or_else(|| "could not locate the app-data directory".to_string())?;
    let link = start_menu_link_in(&appdata);
    if link.is_file() {
        return Ok(());
    }
    let dir = windows_install_dir()
        .ok_or_else(|| "could not locate the local app-data directory".to_string())?;
    let target = pick_main_exe(&dir, APP_NAME)
        .ok_or_else(|| "desktop app folder has no executable to link".to_string())?;
    if let Some(parent) = link.parent() {
        fs::create_dir_all(parent)
            .map_err(|e| format!("could not create the Start Menu folder: {e}"))?;
    }
    create_shortcut(&link, &target, &dir)
}

/// Join the Start Menu link path under an app-data dir. Pure for tests;
/// the only caller fills in `%APPDATA%`.
/// Compiled out of non-Windows builds except under test.
#[cfg(any(target_os = "windows", test))]
fn start_menu_link_in(appdata: &Path) -> PathBuf {
    appdata
        .join("Microsoft")
        .join("Windows")
        .join("Start Menu")
        .join("Programs")
        .join(APP_NAME)
        .join(format!("{APP_NAME}.lnk"))
}

/// Pick the executable a shortcut should point at: `<App>.exe` when it is
/// there, otherwise the first non-uninstaller `.exe` in the folder.
/// Compiled out of non-Windows builds except under test.
#[cfg(any(target_os = "windows", test))]
fn pick_main_exe(dir: &Path, app_name: &str) -> Option<PathBuf> {
    let wanted = format!("{app_name}.exe").to_lowercase();
    let mut fallback: Option<PathBuf> = None;
    for entry in fs::read_dir(dir).ok()?.flatten() {
        let path = entry.path();
        if !path.is_file() {
            continue;
        }
        if !path
            .extension()
            .is_some_and(|ext| ext.eq_ignore_ascii_case("exe"))
        {
            continue;
        }
        let name = path
            .file_name()
            .map(|name| name.to_string_lossy().to_lowercase())
            .unwrap_or_default();
        if name == wanted {
            return Some(path);
        }
        if fallback.is_none() && !name.starts_with("uninstall") && !name.starts_with("uninst")
        {
            fallback = Some(path);
        }
    }
    fallback
}

/// Create a `.lnk` through the built-in Windows shell object. PowerShell
/// ships with Windows, so this needs nothing extra installed.
#[cfg(target_os = "windows")]
fn create_shortcut(link: &Path, target: &Path, workdir: &Path) -> Result<(), String> {
    let script = format!(
        "$ws = New-Object -ComObject WScript.Shell; \
         $sc = $ws.CreateShortcut('{}'); \
         $sc.TargetPath = '{}'; $sc.WorkingDirectory = '{}'; $sc.Save()",
        ps_quote(&link.to_string_lossy()),
        ps_quote(&target.to_string_lossy()),
        ps_quote(&workdir.to_string_lossy()),
    );
    let status = std::process::Command::new("powershell")
        .args(["-NoProfile", "-NonInteractive", "-Command", &script])
        .status()
        .map_err(|e| format!("could not create the Start Menu shortcut: {e}"))?;
    if !status.success() {
        return Err("could not create the Start Menu shortcut".to_string());
    }
    Ok(())
}

/// Quote a value for embedding in a single-quoted PowerShell literal.
/// Compiled out of non-Windows builds except under test.
#[cfg(any(target_os = "windows", test))]
fn ps_quote(value: &str) -> String {
    value.replace('\'', "''")
}

/// Extract a `.app.tar.gz` straight into `~/Applications`.
/// No installer UI, no elevation.
#[cfg(target_os = "macos")]
fn install_mac_app(archive_bytes: &[u8]) -> Result<(), String> {
    use std::io::{Cursor, Read};
    let gz = flate2::read::GzDecoder::new(Cursor::new(archive_bytes));
    let mut archive = tar::Archive::new(gz);
    let entries = archive
        .entries()
        .map_err(|e| format!("could not read the desktop archive: {e}"))?;
    // The bundle root is the first `<Name>.app` directory in the archive.
    let mut root: Option<std::ffi::OsString> = None;
    let mut pending: Vec<(PathBuf, Vec<u8>, u32)> = Vec::new();
    for entry in entries {
        let mut entry = entry.map_err(|e| format!("could not read archive entry: {e}"))?;
        let path = entry
            .path()
            .map_err(|e| format!("bad path in archive: {e}"))?
            .to_path_buf();
        let mut parts = path.components();
        let Some(first) = parts.next() else {
            continue;
        };
        let first = first.as_os_str();
        if !first.to_string_lossy().ends_with(".app") {
            continue;
        }
        if root.is_none() {
            root = Some(first.to_os_string());
        }
        if root.as_deref() != Some(first) {
            continue;
        }
        let mut buf = Vec::new();
        entry
            .read_to_end(&mut buf)
            .map_err(|e| format!("could not read archive entry: {e}"))?;
        #[cfg(unix)]
        let mode = entry
            .header()
            .mode()
            .map_err(|e| format!("could not read entry permissions: {e}"))?;
        #[cfg(not(unix))]
        let mode = 0o644;
        pending.push((path, buf, mode));
    }
    let Some(root) = root else {
        return Err("desktop archive contains no .app bundle".to_string());
    };
    let dest_root = applications_dir().join(&root);
    if dest_root.exists() {
        fs::remove_dir_all(&dest_root)
            .map_err(|e| format!("could not replace the existing desktop app: {e}"))?;
    }
    for (path, buf, mode) in pending {
        let dest = applications_dir().join(&path);
        if let Some(parent) = dest.parent() {
            fs::create_dir_all(parent)
                .map_err(|e| format!("could not create app directory: {e}"))?;
        }
        // Tar entries for directories have no content; create them.
        if buf.is_empty() && path.to_string_lossy().ends_with('/') {
            fs::create_dir_all(&dest)
                .map_err(|e| format!("could not create app directory: {e}"))?;
            continue;
        }
        fs::write(&dest, &buf).map_err(|e| format!("could not write app file: {e}"))?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            let _ = fs::set_permissions(&dest, fs::Permissions::from_mode(mode));
        }
    }
    if !dest_root.is_dir() {
        return Err("desktop install finished but the app bundle is missing".to_string());
    }
    Ok(())
}

/// Quiet `.dmg` fallback: mount, copy the bundle, detach.
#[cfg(target_os = "macos")]
fn install_dmg(dmg_bytes: &[u8]) -> Result<(), String> {
    let staging = std::env::temp_dir().join(format!("star-desktop-{}.dmg", std::process::id()));
    fs::write(&staging, dmg_bytes)
        .map_err(|e| format!("could not stage the desktop installer: {e}"))?;
    let mount = PathBuf::from(format!("/tmp/star-desktop-mnt-{}", std::process::id()));
    let _ = fs::remove_dir_all(&mount);
    fs::create_dir_all(&mount).map_err(|e| format!("could not create mount point: {e}"))?;
    let result = (|| -> Result<(), String> {
        let status = std::process::Command::new("hdiutil")
            .args([
                "attach",
                "-nobrowse",
                "-mountpoint",
                &mount.to_string_lossy(),
                &staging.to_string_lossy(),
            ])
            .status()
            .map_err(|e| format!("could not mount the desktop image: {e}"))?;
        if !status.success() {
            return Err("could not mount the desktop image".to_string());
        }
        let mut bundle: Option<PathBuf> = None;
        let entries =
            fs::read_dir(&mount).map_err(|e| format!("could not read the mounted image: {e}"))?;
        for entry in entries.flatten() {
            let path = entry.path();
            if path.extension().is_some_and(|ext| ext == "app") && path.is_dir() {
                bundle = Some(path);
                break;
            }
        }
        let Some(bundle) = bundle else {
            return Err("desktop image contains no .app bundle".to_string());
        };
        let file_name = bundle
            .file_name()
            .ok_or_else(|| "desktop bundle has no name".to_string())?;
        let dest = applications_dir().join(file_name);
        if dest.exists() {
            fs::remove_dir_all(&dest)
                .map_err(|e| format!("could not replace the existing desktop app: {e}"))?;
        }
        fs::create_dir_all(applications_dir())
            .map_err(|e| format!("could not create ~/Applications: {e}"))?;
        let status = std::process::Command::new("cp")
            .args(["-R", &bundle.to_string_lossy(), &dest.to_string_lossy()])
            .status()
            .map_err(|e| format!("could not copy the desktop app: {e}"))?;
        if !status.success() {
            return Err("could not copy the desktop app".to_string());
        }
        Ok(())
    })();
    let _ = std::process::Command::new("hdiutil")
        .args(["detach", &mount.to_string_lossy(), "-quiet"])
        .status();
    let _ = fs::remove_file(&staging);
    let _ = fs::remove_dir_all(&mount);
    result
}

/// AppImage: save it executable into the install dir and register a menu
/// entry. Nothing here needs privileges.
#[cfg(target_os = "linux")]
fn install_appimage(install_dir: &Path, bytes: &[u8]) -> Result<(), String> {
    use std::os::unix::fs::PermissionsExt;
    fs::create_dir_all(install_dir)
        .map_err(|e| format!("could not create install directory: {e}"))?;
    let dest = install_dir.join("star-desktop.AppImage");
    fs::write(&dest, bytes).map_err(|e| format!("could not save the desktop app: {e}"))?;
    fs::set_permissions(&dest, fs::Permissions::from_mode(0o755))
        .map_err(|e| format!("could not make the desktop app executable: {e}"))?;
    // Best effort: show up in app menus. A missing entry never fails setup.
    let apps_dir = update::home_dir().join(".local/share/applications");
    if fs::create_dir_all(&apps_dir).is_ok() {
        let entry = format!(
            "[Desktop Entry]\nType=Application\nName={APP_NAME}\nExec={}\nTerminal=false\nCategories=Development;\n",
            dest.to_string_lossy()
        );
        let _ = fs::write(apps_dir.join("star-desktop.desktop"), entry);
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    const RELEASE_JSON: &str = r#"{
        "tag_name": "v0.1.115",
        "assets": [
            {"name": "star-windows-amd64.zip",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/star-windows-amd64.zip"},
            {"name": "Star Desktop_0.1.115_x64-setup.exe",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/Star%20Desktop_0.1.115_x64-setup.exe"},
            {"name": "Star Desktop_0.1.115_x64_en-US.msi",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/Star%20Desktop_0.1.115_x64_en-US.msi"},
            {"name": "Star Desktop_0.1.115_aarch64.dmg",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/Star%20Desktop_0.1.115_aarch64.dmg"},
            {"name": "Star Desktop_0.1.115_aarch64.app.tar.gz",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/Star%20Desktop_0.1.115_aarch64.app.tar.gz"},
            {"name": "star-desktop_0.1.115_amd64.AppImage",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/star-desktop_0.1.115_amd64.AppImage"},
            {"name": "star-desktop_0.1.115_amd64.deb",
             "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/star-desktop_0.1.115_amd64.deb"}
        ]
    }"#;

    #[test]
    fn prefers_nsis_over_msi() {
        let picked = pick_asset_for_arch(RELEASE_JSON, &["-setup.exe", ".msi"], &["x64"])
            .expect("pick windows asset");
        assert_eq!(picked.0, "v0.1.115");
        assert!(picked.1.ends_with("-setup.exe"));
        assert!(picked.2.contains("/download/v0.1.115/"));
    }

    #[test]
    fn prefers_app_tarball_over_dmg() {
        let picked = pick_asset_for_arch(RELEASE_JSON, &[".app.tar.gz", ".dmg"], &["aarch64"])
            .expect("pick macos asset");
        assert_eq!(picked.0, "v0.1.115");
        assert!(picked.1.ends_with(".app.tar.gz"));
    }

    #[test]
    fn prefers_appimage_over_deb_case_insensitively() {
        let picked = pick_asset_for_arch(RELEASE_JSON, &[".appimage", ".deb"], &["amd64"])
            .expect("pick linux asset");
        assert_eq!(picked.0, "v0.1.115");
        assert!(picked.1.to_lowercase().ends_with(".appimage"));
    }

    #[test]
    fn falls_back_down_the_preference_list() {
        // No NSIS uploader ran: the MSI is still a valid pick.
        let body = RELEASE_JSON.replace("-setup.exe", "-setup.ex_");
        let picked = pick_asset_for_arch(&body, &["-setup.exe", ".msi"], &["x64"])
            .expect("fall back to msi");
        assert!(picked.1.ends_with(".msi"));
    }

    #[test]
    fn no_matching_asset_is_none() {
        assert!(pick_asset_for_arch(RELEASE_JSON, &[".rpm"], &["x64"]).is_none());
        assert!(pick_asset_for_arch("{}", &["-setup.exe"], &["x64"]).is_none());
    }

    #[test]
    fn wrong_arch_assets_are_skipped() {
        // Mirrors the real v0.1.115 release: the aarch64 AppImage is listed
        // before the amd64 one, but each machine must pick its own build.
        const BODY: &str = r#"{
            "tag_name": "v0.1.115",
            "assets": [
                {"name": "star-desktop_0.1.115_aarch64.AppImage",
                 "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/star-desktop_0.1.115_aarch64.AppImage"},
                {"name": "star-desktop_0.1.115_amd64.AppImage",
                 "browser_download_url": "https://github.com/GenLabsAI/star/releases/download/v0.1.115/star-desktop_0.1.115_amd64.AppImage"}
            ]
        }"#;
        let picked = pick_asset_for_arch(BODY, &[".appimage"], &["x86_64", "amd64"])
            .expect("pick linux asset for x86_64");
        assert!(picked.1.contains("amd64"), "picked {}", picked.1);
        let picked = pick_asset_for_arch(BODY, &[".appimage"], &["aarch64", "arm64"])
            .expect("pick linux asset for aarch64");
        assert!(picked.1.contains("aarch64"), "picked {}", picked.1);
        // With only foreign-arch assets available there is no pick at all,
        // rather than a broken install (RELEASE_JSON has amd64-only Linux
        // assets).
        assert!(
            pick_asset_for_arch(RELEASE_JSON, &[".appimage", ".deb"], &["aarch64", "arm64"])
                .is_none()
        );
    }

    #[test]
    fn archless_names_are_universal() {
        assert!(name_matches_arch("Star Desktop.AppImage", &["amd64"]));
        assert!(name_matches_arch("Star Desktop_x64-setup.exe", &["x64"]));
        assert!(!name_matches_arch("Star Desktop_aarch64.dmg", &["x64"]));
    }

    #[test]
    fn marker_roundtrip() {
        let dir = std::env::temp_dir().join(format!(
            "star-desktop-test-{}-{:?}",
            std::process::id(),
            std::thread::current().id()
        ));
        let _ = fs::remove_dir_all(&dir);
        assert!(marker_version(&dir).is_none());
        write_marker(&dir, "v0.1.115").expect("write marker");
        assert_eq!(marker_version(&dir).as_deref(), Some("v0.1.115"));
        let _ = fs::remove_dir_all(dir);
    }

    #[test]
    fn shortcut_exe_prefers_product_over_uninstaller() {
        let dir = std::env::temp_dir().join(format!(
            "star-desktop-exe-{}-{:?}",
            std::process::id(),
            std::thread::current().id()
        ));
        let _ = fs::remove_dir_all(&dir);
        fs::create_dir_all(&dir).expect("create dir");
        fs::write(dir.join("Uninstall Star Desktop.exe"), b"x").expect("write");
        fs::write(dir.join("readme.txt"), b"x").expect("write");
        assert!(pick_main_exe(&dir, "Star Desktop").is_none());
        fs::write(dir.join("Star Desktop.exe"), b"x").expect("write");
        assert_eq!(
            pick_main_exe(&dir, "Star Desktop"),
            Some(dir.join("Star Desktop.exe"))
        );
        let _ = fs::remove_dir_all(&dir);
    }

    #[test]
    fn shortcut_path_matches_tauri_layout() {
        // Built with joins so the assertion holds on any host separator;
        // on Windows this is the `Programs\Star Desktop\Star Desktop.lnk`
        // layout the Start Menu indexer picks up.
        let base = Path::new("AppData");
        let link = start_menu_link_in(base);
        assert_eq!(
            link,
            base.join("Microsoft")
                .join("Windows")
                .join("Start Menu")
                .join("Programs")
                .join("Star Desktop")
                .join("Star Desktop.lnk")
        );
    }

    #[test]
    fn ps_single_quotes_are_escaped() {
        assert_eq!(ps_quote("o'clock"), "o''clock");
        assert_eq!(ps_quote("plain"), "plain");
    }
}
