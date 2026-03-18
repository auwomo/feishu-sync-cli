Param(
  [string]$Version = "latest",
  [switch]$Uninstall,
  [switch]$NoPath
)

# Default behavior: add install dir to user PATH unless -NoPath is specified.
$AddToPath = -not $NoPath

$ErrorActionPreference = "Stop"

$Repo = "auwomo/feishu-sync-cli"
$BinName = "feishu-sync.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "feishu-sync\bin"
$Target = Join-Path $InstallDir $BinName

function Write-Info($s) { Write-Host $s -ForegroundColor DarkGray }
function Write-Ok($s) { Write-Host $s -ForegroundColor Green }
function Write-Warn($s) { Write-Host $s -ForegroundColor Yellow }
function Write-Err($s) { Write-Host "error: $s" -ForegroundColor Red }

function Get-LatestTag {
  $url = "https://api.github.com/repos/$Repo/releases/latest"
  $json = Invoke-RestMethod -Uri $url -Headers @{"User-Agent"="feishu-sync-install"}
  return $json.tag_name
}

function Ensure-Path {
  if ($AddToPath) {
    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($current -notlike "*$InstallDir*") {
      [Environment]::SetEnvironmentVariable("Path", "$current;$InstallDir", "User")
      Write-Ok "Added to user PATH: $InstallDir"
    } else {
      Write-Info "Already on user PATH: $InstallDir"
    }
  } else {
    Write-Warn "$InstallDir is not on PATH (because -NoPath was used)"
    Write-Host "You can add it to your user PATH: $InstallDir"
    Write-Host "Or re-run without -NoPath to add it automatically."
  }
}

if ($Uninstall) {
  if (Test-Path $Target) {
    Remove-Item -Force $Target
    Write-Ok "Removed $Target"
  } else {
    Write-Info "Not installed: $Target"
  }

  # Best-effort PATH removal (user scope). Some terminals cache env; user may need to verify manually.
  $current = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($current -and ($current -like "*$InstallDir*")) {
    $parts = $current.Split(';') | Where-Object { $_ -and ($_ -ne $InstallDir) }
    $newPath = ($parts -join ';')
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Ok "Removed from user PATH: $InstallDir"
  } else {
    Write-Info "Install dir not found on user PATH: $InstallDir"
    Write-Host "If you added it manually, remove this entry from your PATH: $InstallDir"
  }

  Write-Host "Next steps:"
  Write-Host "  - Open a new terminal"
  return 0
}

if ($Version -eq "latest") {
  $Tag = Get-LatestTag
} else {
  $Tag = $Version
}

Write-Info "version: $Tag"
Write-Info "install dir: $InstallDir"

# Windows binaries may not be published yet.
# Placeholder expectations: a zip asset like feishu-sync_<ver>_windows_amd64.zip
# If not found, print friendly message.
$url = "https://api.github.com/repos/$Repo/releases/tags/$Tag"
$release = Invoke-RestMethod -Uri $url -Headers @{"User-Agent"="feishu-sync-install"}
# Prefer a windows zip matching this arch.
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $arch = "arm64" }

$versionNoV = $Tag
if ($versionNoV.StartsWith('v')) { $versionNoV = $versionNoV.Substring(1) }

# Match exact asset name for this version/arch, e.g.
# feishu-sync_0.1.3_windows_amd64.zip
$assetName = "feishu-sync_${versionNoV}_windows_${arch}.zip"
$asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1

# Fallback: any windows zip for this arch (useful if naming changes)
if (-not $asset) {
  $fallbackPattern = "windows_${arch}\\.zip$"
  $asset = $release.assets | Where-Object { $_.browser_download_url -match $fallbackPattern } | Select-Object -First 1
}

if (-not $asset) {
  Write-Warn "Windows binaries not published yet"
  Write-Host "See releases: https://github.com/$Repo/releases"
  return 3
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$tmp = New-Item -ItemType Directory -Force -Path ([IO.Path]::Combine([IO.Path]::GetTempPath(), "feishu-sync-install" + [guid]::NewGuid().ToString()))
$zip = Join-Path $tmp "asset.zip"

Write-Info "download: $($asset.browser_download_url)"
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zip

Expand-Archive -Path $zip -DestinationPath $tmp -Force
$exe = Get-ChildItem -Path $tmp -Recurse -Filter $BinName | Select-Object -First 1
if (-not $exe) {
  Write-Err "Downloaded asset did not contain $BinName"
  return 5
}

Copy-Item -Force $exe.FullName $Target
Write-Ok "Installed to $Target"

Ensure-Path

Remove-Item -Recurse -Force $tmp

Write-Host ""
Write-Ok "Done."
Write-Host "Next steps:"
Write-Host "  - Open a new terminal"
Write-Host "  - feishu-sync --help"

return 0
