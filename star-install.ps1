#requires -Version 5.1
param(
    [switch]$SkipUpgrade,
    [string]$Tag = ""
)
<#
  Install or upgrade Star to the latest GitHub release on Windows.

  Always installs into %USERPROFILE%\bin (creating it if needed), because
  that is the only folder the star launcher loads its core from. Installing
  anywhere else is silently ignored by the launcher, so this script does not
  offer another target. %USERPROFILE%\bin is added to the user PATH when
  missing, so `star` works in new terminals.

  The latest release is resolved from the GitHub Releases API at runtime so
  this script does not hardcode a version.

  Idempotent: rerun safely. If the release is unreachable or the download is
  corrupt it aborts without replacing the current install. After installing
  the binaries, the first `star` run uses the new core immediately, and Star
  Desktop installs itself in the background (Start Menu entry included).

  Optional switches:
    -SkipUpgrade
        Only install if star.exe is absent in %USERPROFILE%\bin.
    -Tag <tag>
        Install this release (e.g. -Tag v0.1.116) instead of asking the
        GitHub API for the latest. Useful when api.github.com is blocked
        on your network but github.com downloads still work.
#>
$ErrorActionPreference = "Stop"

# Windows PowerShell 5.1 defaults to TLS 1.0, which api.github.com rejects.
# Opt into TLS 1.2/1.3 by numeric value so this also works on .NET versions
# where the Tls13 enum member does not exist (12288 = Tls13, 3072 = Tls12).
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor 3072 -bor 12288
} catch { }

$Owner  = "GenLabsAI"
$Repo   = "star"
$Arch   = "windows-amd64"
$Asset  = "star-$Arch.zip"

function Test-Uri([string]$Url) {
    try {
        $req = [System.Net.WebRequest]::Create($Url)
        $req.Method = "HEAD"
        $resp = $req.GetResponse()
        $resp.Close()
        return $true
    } catch {
        return $false
    }
}

$BinHome = $env:USERPROFILE
if ([string]::IsNullOrEmpty($BinHome)) {
    $BinHome = [Environment]::GetFolderPath("Profile")
}
$InstallDir = Join-Path $BinHome "bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}
$InstallDir = (Resolve-Path $InstallDir).Path

Write-Host "[0/7] install target: $InstallDir"

if ($SkipUpgrade -and (Test-Path (Join-Path $InstallDir "star.exe"))) {
    Write-Host "[0/7] star.exe already present and -SkipUpgrade set; stopping."
    Write-Host "        To force an upgrade, rerun without -SkipUpgrade or delete the old binaries first."
    exit 0
}

# Fail early with a clear message instead of a cryptic "Access denied" halfway
# through: a running star.exe cannot be replaced on Windows.
$TargetStar = Join-Path $InstallDir "star.exe"
if (Test-Path $TargetStar) {
    $locked = $true
    for ($i = 0; $i -lt 3 -and $locked; $i++) {
        try {
            $probe = [System.IO.File]::Open($TargetStar, [System.IO.FileMode]::Open, [System.IO.FileAccess]::ReadWrite, [System.IO.FileShare]::None)
            $probe.Close()
            $locked = $false
        } catch {
            Start-Sleep -Milliseconds 500
        }
    }
    if ($locked) {
        throw "star.exe is currently running from $InstallDir. Close all Star windows and rerun this script."
    }
}

$LatestApiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
$LatestPageUrl = "https://github.com/$Owner/$Repo/releases/latest"

$ResolvedTag = $null
$ResolvedZipUrl = $null
$ResolvedSumUrl = $null

# An explicit -Tag skips all version discovery: the release assets live at
# fixed download URLs, so no GitHub API call is needed at all.
if (-not [string]::IsNullOrWhiteSpace($Tag)) {
    $ResolvedTag = $Tag.Trim()
    $ResolvedZipUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/$Asset"
    $ResolvedSumUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/checksums.txt"
    Write-Host "[1/7] using explicit tag: $ResolvedTag"
}

$apiOk = $false
$apiResp = $null
if ([string]::IsNullOrEmpty($ResolvedTag)) {
    try {
        $req = [System.Net.WebRequest]::Create($LatestApiUrl)
        $req.UserAgent = "star-update-script/1.0"
        $req.Headers.Add("Accept", "application/vnd.github+json, application/vnd.github.v3+json")
        $resp = $req.GetResponse()
        $stream = $resp.GetResponseStream()
        $reader = [System.IO.StreamReader]::new($stream)
        $body = $reader.ReadToEnd()
        $resp.Close()
        $apiResp = $body
        $apiOk = $true
    } catch {
        Write-Warning "GitHub Releases API unreachable via $LatestApiUrl : $($_.Exception.GetType().Name)"
    }
}

if ($apiOk) {
    $obj = $null
    try {
        $obj = $apiResp | ConvertFrom-Json -ErrorAction Stop
    } catch {
        throw "GitHub API returned unexpected content and could not parse JSON"
    }

    $resolvedTagCandidate = $null
    if ($obj.tag_name) {
        $resolvedTagCandidate = $obj.tag_name
    } elseif ($obj.name -and $obj.name -match '^(v?[0-9]+\.[0-9]+(\.[0-9]+)?)') {
        $resolvedTagCandidate = $Matches[1]
    }

    if ([string]::IsNullOrEmpty($resolvedTagCandidate)) {
        throw "GitHub API response had no usable version tag (tag_name or name). Response:`n$apiResp"
    }

    $isPrerelease = $false
    if ($obj.prerelease -eq $true -or $obj.prerelease -eq "true" -or $obj.prerelease -eq 1) {
        $isPrerelease = $true
    }

    $ResolvedTag = $resolvedTagCandidate
    $ResolvedZipUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/$Asset"
    $ResolvedSumUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/checksums.txt"

    if ($isPrerelease) {
        Write-Warning "Latest release $ResolvedTag is marked as a prerelease. Proceeding anyway."
    }
} elseif ([string]::IsNullOrEmpty($ResolvedTag)) {
    # The API call failed (often TLS or a blocked api.github.com). Fall back
    # to the /releases/latest page, which 302-redirects to /releases/tag/<tag>.
    Write-Warning "API lookup failed; trying $LatestPageUrl instead."
    try {
        $redirReq = [System.Net.WebRequest]::Create($LatestPageUrl)
        $redirReq.AllowAutoRedirect = $false
        $redirReq.UserAgent = "star-update-script/1.0"
        try {
            $redirResp = $redirReq.GetResponse()
            $redirResp.Close()
        } catch [System.Net.WebException] {
            $redirResp = $_.Exception.Response
        }
        $location = $null
        if ($null -ne $redirResp) {
            $location = $redirResp.Headers["Location"]
            try { $redirResp.Close() } catch { }
        }
        if ($location -match '/releases/tag/([^/?#]+)') {
            $ResolvedTag = $Matches[1]
            $ResolvedZipUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/$Asset"
            $ResolvedSumUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedTag/checksums.txt"
            Write-Host "[1/7] resolved latest release via redirect: $ResolvedTag"
        }
    } catch {
        Write-Warning "Redirect lookup failed: $($_.Exception.Message)"
    }
    if ([string]::IsNullOrEmpty($ResolvedTag)) {
        throw "Could not reach GitHub to determine the latest release (tried the Releases API and the /releases/latest redirect). Run this script again when GitHub is reachable, or open $LatestPageUrl in a browser and rerun with -Tag <tag> (e.g. .\star-install.ps1 -Tag v0.1.117)."
    }
}

if ([string]::IsNullOrWhiteSpace($Tag)) {
    Write-Host "[1/7] resolved latest release: $ResolvedTag"
}
Write-Host "        zip : $ResolvedZipUrl"
Write-Host "        sum : $ResolvedSumUrl"

if (-not (Test-Uri $ResolvedSumUrl)) {
    throw "checksums.txt not reachable: $ResolvedSumUrl"
}
if (-not (Test-Uri $ResolvedZipUrl)) {
    throw "zip not reachable: $ResolvedZipUrl"
}

$ZipPath = Join-Path $InstallDir $Asset
$SumPath  = Join-Path $InstallDir "checksums.txt"

Write-Host "[2/7] downloading checksums.txt"
Invoke-WebRequest -UseBasicParsing -Uri $ResolvedSumUrl -OutFile $SumPath
if (-not (Test-Path $SumPath)) {
    throw "checksums.txt download failed"
}

$SumLines = [System.IO.File]::ReadAllLines($SumPath)
$AssetLine = $null
foreach ($line in $SumLines) {
    if ($line -match "^\s*([A-Za-z0-9]{64})\s+.*$([regex]::Escape($Asset))\s*$") {
        $AssetLine = $line
        break
    }
}
if ($null -eq $AssetLine -or $null -eq $Matches[1]) {
    throw "checksums.txt does not contain an entry for $Asset.`nContents:`n$($SumLines -join "`n")"
}
$ExpectedSha = $Matches[1]
Write-Host "[2/7] expected sha256 = $ExpectedSha"

Write-Host "[3/7] downloading zip"
$Start = Get-Date
Invoke-WebRequest -UseBasicParsing -Uri $ResolvedZipUrl -OutFile $ZipPath
$Elapsed = (Get-Date) - $Start
$ActualSize = (Get-Item $ZipPath).Length
Write-Host "[3/7] zip written: $ZipPath  size=$ActualSize bytes  elapsed=$([math]::Round($Elapsed.TotalSeconds, 1))s"

Write-Host "[4/7] verifying SHA-256"
$ActualSha = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLower()
if ($ActualSha -ne $ExpectedSha.ToLower()) {
    Remove-Item $ZipPath -Force -ErrorAction SilentlyContinue
    throw "SHA-256 mismatch:`n  expected: $ExpectedSha`n  actual:   $ActualSha"
}
Write-Host "[4/7] SHA-256 OK"

Write-Host "[5/7] extracting"
$ExtractDir = Join-Path $InstallDir "star-extract"
if (Test-Path $ExtractDir) {
    Remove-Item $ExtractDir -Recurse -Force
}
Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

$Needed = @("star.exe", "star-core.exe")
$FoundNames = @()
$BinDir = $null
foreach ($f in (Get-ChildItem $ExtractDir -Recurse -File)) {
    if ($f.Name -in $Needed) {
        $FoundNames += $f.Name
        if ($null -eq $BinDir) { $BinDir = $f.DirectoryName }
    }
}
$missing = $Needed | Where-Object { $_ -notin $FoundNames }
if ($null -eq $BinDir -or $missing) {
    Remove-Item $ExtractDir -Recurse -Force -ErrorAction SilentlyContinue
    throw "zip did not contain the expected binaries.`nExpected: $($Needed -join ' ')`nFound: $($FoundNames -join ' ')`nContents:`n$(Get-ChildItem $ExtractDir -Recurse -File | Select-Object -ExpandProperty FullName -Unique | Out-String)"
}

Write-Host "[6/7] installing binaries into $InstallDir"

# Install the core into a versioned directory and point the launcher at it.
# Running cores keep their old file, so an in-app update never needs to close
# every Star window. The root star-core.exe is kept as a legacy fallback.
$VersionsRoot = Join-Path $InstallDir "versions"
$VersionDir = Join-Path $VersionsRoot $ResolvedTag
New-Item -ItemType Directory -Path $VersionDir -Force | Out-Null
Copy-Item -Path (Join-Path $BinDir "star-core.exe") -Destination (Join-Path $VersionDir "star-core.exe") -Force
Move-Item -Path (Join-Path $BinDir "star.exe") -Destination $InstallDir -Force
Move-Item -Path (Join-Path $BinDir "star-core.exe") -Destination $InstallDir -Force
[System.IO.File]::WriteAllText((Join-Path $InstallDir "current"), $ResolvedTag, [System.Text.Encoding]::ASCII)

# Keep the active version and one previous version. In-use files may not be
# removable yet; cleanup retries on the next launch.
Get-ChildItem -Path $VersionsRoot -Directory -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -ne $ResolvedTag -and -not $_.Name.StartsWith(".staging-") } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip 1 |
    ForEach-Object { Remove-Item $_.FullName -Recurse -Force -ErrorAction SilentlyContinue }
Get-ChildItem -Path $VersionsRoot -Directory -Filter ".staging-*" -ErrorAction SilentlyContinue |
    ForEach-Object { Remove-Item $_.FullName -Recurse -Force -ErrorAction SilentlyContinue }

Remove-Item $ExtractDir -Recurse -Force
Remove-Item $ZipPath -Force

Write-Host "[7/7] ensuring $InstallDir is on the user PATH"
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ([string]::IsNullOrEmpty($UserPath)) { $UserPath = "" }
$onPath = $false
foreach ($p in ($UserPath -split ';')) {
    if ($p -and ($p.TrimEnd('\', '/') -ieq $InstallDir.TrimEnd('\', '/'))) { $onPath = $true; break }
}
if (-not $onPath) {
    [Environment]::SetEnvironmentVariable("Path", ($UserPath.TrimEnd(';') + ';' + $InstallDir).TrimStart(';'), "User")
    Write-Host "[7/7] added to user PATH (takes effect in new terminals)"
} else {
    Write-Host "[7/7] already on user PATH"
}
if ($env:Path) {
    $sessionOnPath = $false
    foreach ($p in ($env:Path -split ';')) {
        if ($p -and ($p.TrimEnd('\', '/') -ieq $InstallDir.TrimEnd('\', '/'))) { $sessionOnPath = $true; break }
    }
    if (-not $sessionOnPath) { $env:Path += ";$InstallDir" }
}

Write-Host "`nInstalled/updated $ResolvedTag into:"
Write-Host "  $InstallDir"
Write-Host ""
Write-Host "Binaries:"
Get-ChildItem -Path $InstallDir -Filter "*.exe" | ForEach-Object {
    $h = (Get-FileHash $_.FullName -Algorithm SHA256).Hash
    Write-Host "  $($_.Name)  $($_.Length) bytes  SHA256=$h"
}
Write-Host ""
Write-Host "Next steps:"
Write-Host "  1. Open a NEW terminal (so it picks up PATH) and run: star"
Write-Host "  2. The first run uses the new core immediately; Star Desktop"
Write-Host "     installs itself in the background with a Start Menu entry."
