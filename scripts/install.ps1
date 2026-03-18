Param(
  [string]$Version = "latest",
  [switch]$Uninstall,
  [switch]$AddToPath
)

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
    Write-Warn "$InstallDir is not guaranteed to be on PATH"
    Write-Host "Add it to PATH (User): $InstallDir"
    Write-Host "Or re-run with: -AddToPath"
  }
}

if ($Uninstall) {
  if (Test-Path $Target) {
    Remove-Item -Force $Target
    Write-Ok "Removed $Target"
  } else {
    Write-Info "Not installed: $Target"
  }
  exit 0
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
$asset = $release.assets | Where-Object { $_.browser_download_url -match "windows" } | Select-Object -First 1

if (-not $asset) {
  Write-Warn "Windows binaries not published yet"
  Write-Host "See releases: https://github.com/$Repo/releases"
  exit 3
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
  exit 5
}

Copy-Item -Force $exe.FullName $Target
Write-Ok "Installed to $Target"

Ensure-Path

Remove-Item -Recurse -Force $tmp
exit 0
