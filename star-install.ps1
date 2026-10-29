#requires -Version 5.1
<#
  Install or upgrade Star to the latest GitHub release on Windows.

  By default, installs into the folder containing this script. If a suitable
  install is already present in ~/bin (both star.exe and star-core.exe),
  that folder is preferred instead, so existing PATH-based usage keeps working.

  The latest release is resolved from the GitHub Releases API at runtime so
  this script does not hardcode a version.

  Idempotent: rerun safely. If the release is unreachable or the download is
  corrupt it aborts without replacing the current install.

  Optional switches:
    -SkipUpgrade
        Only install if star.exe is absent in the chosen install folder.

  Environment:
    STAR_INSTALL_BIN
        If set and non-empty, overrides folder selection and installs into this
        path instead of the script folder or ~/bin.
#>
$ErrorActionPreference = "Stop"

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
$StarBinDir = Join-Path $BinHome "bin"

$Explicit = $env:STAR_INSTALL_BIN
if ([string]::IsNullOrWhiteSpace($Explicit)) {
    $Explicit = $null
}

$InstallDir = $null
$Reason = $null

if (-not [string]::IsNullOrWhiteSpace($Explicit)) {
    if (-not (Test-Path $Explicit)) {
        New-Item -ItemType Directory -Path $Explicit -Force | Out-Null
    }
    $InstallDir = (Resolve-Path $Explicit).Path
    $Reason = "STAR_INSTALL_BIN=$InstallDir"
} elseif ((Test-Path $StarBinDir) -and
          (Test-Path (Join-Path $StarBinDir "star.exe")) -and
          (Test-Path (Join-Path $StarBinDir "star-core.exe"))) {
    $InstallDir = (Resolve-Path $StarBinDir).Path
    $Reason = "~/bin already has star.exe and star-core.exe"
} else {
    $InstallDir = $PSScriptRoot
    if ([string]::IsNullOrWhiteSpace($InstallDir)) {
        $InstallDir = (Get-Location).Path
    }
    $InstallDir = (Resolve-Path $InstallDir).Path
    $Reason = "fallback: script folder"
}

Write-Host "[0/6] install target: $InstallDir  ($Reason)"

$LatestApiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"

$ResolvedTag = $null
$ResolvedZipUrl = $null
$ResolvedSumUrl = $null

$apiOk = $false
$apiResp = $null
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
} else {
    throw "Could not reach GitHub Releases API and cannot safely pick a default latest tag. Run this script again when GitHub is reachable, or open https://github.com/$Owner/$Repo/releases/latest and tell me the tag."
}

Write-Host "[1/6] resolved latest release: $ResolvedTag"
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

Write-Host "[2/6] downloading checksums.txt"
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
Write-Host "[2/6] expected sha256 = $ExpectedSha"

if ($SkipUpgrade -and (Test-Path (Join-Path $InstallDir "star.exe"))) {
    Write-Host "[3/6] star.exe already present and -SkipUpgrade set; stopping."
    Write-Host "        To force an upgrade, rerun without -SkipUpgrade or delete the old binaries first."
    exit 0
}

Write-Host "[3/6] downloading zip"
$Start = Get-Date
Invoke-WebRequest -UseBasicParsing -Uri $ResolvedZipUrl -OutFile $ZipPath
$Elapsed = (Get-Date) - $Start
$ActualSize = (Get-Item $ZipPath).Length
Write-Host "[3/6] zip written: $ZipPath  size=$ActualSize bytes  elapsed=$([math]::Round($Elapsed.TotalSeconds, 1))s"

Write-Host "[4/6] verifying SHA-256"
$ActualSha = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLower()
if ($ActualSha -ne $ExpectedSha.ToLower()) {
    Remove-Item $ZipPath -Force -ErrorAction SilentlyContinue
    throw "SHA-256 mismatch:`n  expected: $ExpectedSha`n  actual:   $ActualSha"
}
Write-Host "[4/6] SHA-256 OK"

Write-Host "[5/6] extracting"
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

Write-Host "[6/6] installing binaries into $InstallDir"
Move-Item -Path (Join-Path $BinDir "star.exe") -Destination $InstallDir -Force
Move-Item -Path (Join-Path $BinDir "star-core.exe") -Destination $InstallDir -Force
Remove-Item $ExtractDir -Recurse -Force
Remove-Item $ZipPath -Force

Write-Host "`nInstalled/updated $ResolvedTag into:"
Write-Host "  $InstallDir  ($Reason)"
Write-Host ""
Write-Host "Binaries:"
Get-ChildItem -Path $InstallDir -Filter "*.exe" | ForEach-Object {
    $h = (Get-FileHash $_.FullName -Algorithm SHA256).Hash
    Write-Host "  $($_.Name)  $($_.Length) bytes  SHA256=$h"
}
Write-Host ""
Write-Host "Note: this script installed binaries into $InstallDir."
Write-Host "      PATH or shortcuts are not modified automatically."
