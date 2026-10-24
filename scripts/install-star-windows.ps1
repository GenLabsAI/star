# Star - Windows installer

Downloads the latest Star release for Windows and copies the launcher + core
binaries into your personal bin folder so `star` is on your PATH.

## Prerequisites

- A folder named `bin` inside your user profile, i.e. `%USERPROFILE%\bin`
  (for example `C:\Users\you\bin`).
- That folder is on your PATH. Quick check in PowerShell:

  ```powershell
  $env:PATH -split ';' | Select-String 'bin'
  ```

  If nothing shows up, add it once with:

  ```powershell
  [Environment]::SetEnvironmentVariable('PATH',
      "$([Environment]::GetEnvironmentVariable('PATH','User'));$([Environment]::GetFolderPath('Profile'))\bin",
      'User')
  ```

  Then restart your shell for it to take effect.

## Install

Run this in PowerShell. It fetches the pinned v0.1.75 release, verifies the
download against the release checksums, extracts the two Windows binaries, and
copies them into `%USERPROFILE%\bin`.

```powershell
$ErrorActionPreference = 'Stop'
$version    = 'v0.1.75'
$archive    = "star-windows-amd64.zip"
$checksumFile = "checksums.txt"
$releaseBase = "https://github.com/GenLabsAI/star/releases/download/$version"
$binDir      = Join-Path ([Environment]::GetFolderPath('Profile')) 'bin'

New-Item -ItemType Directory -Force -Path $binDir | Out-Null

# Download the release zip and the checksum file.
$zipPath    = Join-Path $env:TEMP $archive
$checkPath  = Join-Path $env:TEMP $checksumFile

Invoke-WebRequest -Uri "$releaseBase/$archive" -OutFile $zipPath
Invoke-WebRequest -Uri "$releaseBase/$checksumFile" -OutFile $checkPath

# Verify the archive against the published checksum.
$expected = (Select-String -Path $checkPath -Pattern $archive).Line.Trim().Split(' ')[0].ToLower()
$actual   = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
if ($actual -ne $expected) {
    throw "Checksum mismatch for $archive (expected $expected, got $actual)"
}

# Extract and install the two binaries.
$extractDir = Join-Path $env:TEMP "star-extract-$version"
New-Item -ItemType Directory -Force -Path $extractDir | Out-Null
Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

Copy-Item -Path (Join-Path $extractDir 'star.exe')    -Destination $binDir -Force
Copy-Item -Path (Join-Path $extractDir 'star-core.exe') -Destination $binDir -Force

'@("Installed star.exe","Installed star-core.exe")' | ForEach-Object { Write-Host $_ }

Remove-Item -Path $zipPath, $checkPath, $extractDir -Force -ErrorAction SilentlyContinue
```

After installing, run `star` from any directory. The first run will also pull
`star-core.exe` alongside it if you ever end up with just the launcher on its
own, because the launcher expects the core next to it.

## Update later

To update, just run the same script again with a newer `$version`, or run your
existing `star.exe` — the launcher updates itself and the core from GitHub
Releases automatically.

## Manual fallback

If the script can't run for some reason, grab
`https://github.com/GenLabsAI/star/releases/download/v0.1.75/star-windows-amd64.zip`,
extract `star.exe` and `star-core.exe`, and drop both into `%USERPROFILE%\bin`.
