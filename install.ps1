# ⚡ CEM — Unified AI Orchestrator
# Kurulum: irm cem.pw/install.ps1 | iex

$ErrorActionPreference = "Stop"
$REPO    = "http://gitlab.makdos.biz/makdos/cem"
$RELEASE = "https://cem.pw/r"

function Write-Banner {
    $c = [char]27  # ESC
    Write-Host ""
    Write-Host "   ██████╗███████╗███╗   ███╗" -ForegroundColor Cyan
    Write-Host "  ██╔════╝██╔════╝████╗ ████║" -ForegroundColor Cyan
    Write-Host "  ██║     █████╗  ██╔████╔██║" -ForegroundColor Cyan
    Write-Host "  ██║     ██╔══╝  ██║╚██╔╝██║" -ForegroundColor Cyan
    Write-Host "  ╚██████╗███████╗██║  ╚═╝ ██║" -ForegroundColor Cyan
    Write-Host "   ╚═════╝╚══════╝╚═╝      ╚═╝" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  ⚡ Unified AI Orchestrator" -ForegroundColor White -NoNewline
    Write-Host "  ·  " -ForegroundColor DarkGray -NoNewline
    Write-Host "cem.pw" -ForegroundColor Cyan
    Write-Host "  Birden fazla AI CLI'ı tek komutla kullan" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  ────────────────────────────────────────" -ForegroundColor DarkGray
    Write-Host ""
}

function Write-Step { param($n,$t) Write-Host "`n  $n  $t" -ForegroundColor White }
function Write-Info { param($m) Write-Host "  → $m" -ForegroundColor Cyan }
function Write-Ok   { param($m) Write-Host "  ✓ $m" -ForegroundColor Green }
function Write-Warn { param($m) Write-Host "  ⚠ $m" -ForegroundColor Yellow }
function Write-Err  { param($m) Write-Host "  ✗ $m" -ForegroundColor Red; exit 1 }

Write-Banner

# ─── Platform ────────────────────────────────────────────────────────────────
Write-Step "1/4" "Platform tespit ediliyor"

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
Write-Info "OS   : windows"
Write-Info "Arch : $arch"

$binary = "cem-windows-$arch.exe"
$url    = "$RELEASE/$binary"

# ─── Kurulum dizini ──────────────────────────────────────────────────────────
Write-Step "2/4" "Dizin hazırlanıyor"
$installDir = "$env:LOCALAPPDATA\cem\bin"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Write-Info "Dizin: $installDir"

# ─── İndir ───────────────────────────────────────────────────────────────────
Write-Step "3/4" "İndiriliyor"
Write-Info "URL: $url"

$cemExe   = Join-Path $installDir "cem.exe"
$cemiExe  = Join-Path $installDir "cemi.exe"
$cemirExe = Join-Path $installDir "cemir.exe"

try {
    $ProgressPreference = "SilentlyContinue"
    Invoke-WebRequest -Uri $url -OutFile $cemExe -UseBasicParsing
    $ProgressPreference = "Continue"
} catch {
    Write-Err "İndirme başarısız: $_"
}

Copy-Item $cemExe $cemiExe  -Force
Copy-Item $cemExe $cemirExe -Force
Write-Ok "İndirildi"

# ─── PATH ────────────────────────────────────────────────────────────────────
Write-Step "4/4" "PATH ayarlanıyor"

$userPath = [System.Environment]::GetEnvironmentVariable("PATH","User")
if ($userPath -notlike "*$installDir*") {
    [System.Environment]::SetEnvironmentVariable("PATH","$installDir;$userPath","User")
    $env:PATH = "$installDir;$env:PATH"
    Write-Ok "PATH güncellendi"
} else {
    Write-Ok "PATH zaten doğru"
}

# ─── Bitti ───────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ╭─────────────────────────────────────────────╮" -ForegroundColor Green
Write-Host "  │                                             │" -ForegroundColor Green
Write-Host "  │   ✓  CEM kuruldu!   ·   cem.pw             │" -ForegroundColor Green
Write-Host "  │                                             │" -ForegroundColor Green
Write-Host "  ╰─────────────────────────────────────────────╯" -ForegroundColor Green
Write-Host ""
Write-Host "  Yeni terminal açın ve başlayın:" -ForegroundColor White
Write-Host ""
Write-Host '  cem "soru"' -ForegroundColor Cyan -NoNewline
Write-Host "         → thinker AI (wizard açılır)" -ForegroundColor DarkGray
Write-Host "  cem -w `"görev`"     → writer AI" -ForegroundColor DarkGray
Write-Host "  cem -p `"görev`"     → pair modu" -ForegroundColor DarkGray
Write-Host "  cemi claude        → Claude kur" -ForegroundColor DarkGray
Write-Host "  cemi agy           → Agy kur" -ForegroundColor DarkGray
Write-Host "  cem roles          → kim ne yapıyor?" -ForegroundColor DarkGray
Write-Host ""
Write-Host "  Döküman → https://cem.pw/docs" -ForegroundColor DarkGray
Write-Host ""
