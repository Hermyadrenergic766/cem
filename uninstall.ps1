# ⚡ CEM — Windows Kaldırma Scripti
# iwr cem.pw/uninstall.ps1 -UseB | iex

$ErrorActionPreference = "Stop"

function Write-Banner {
    Write-Host ""
    Write-Host "   ██████╗███████╗███╗   ███╗" -ForegroundColor Cyan
    Write-Host "  ██╔════╝██╔════╝████╗ ████║" -ForegroundColor Cyan
    Write-Host "  ██║     █████╗  ██╔████╔██║" -ForegroundColor Cyan
    Write-Host "  ██║     ██╔══╝  ██║╚██╔╝██║" -ForegroundColor Cyan
    Write-Host "  ╚██████╗███████╗██║  ╚═╝ ██║" -ForegroundColor Cyan
    Write-Host "   ╚═════╝╚══════╝╚═╝      ╚═╝" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  🗑  Kaldırma Scripti" -ForegroundColor White -NoNewline
    Write-Host "  ·  cem.pw" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  ────────────────────────────────────────" -ForegroundColor DarkGray
    Write-Host ""
}

function Write-Ok   { param($m) Write-Host "  ✓ $m" -ForegroundColor Green }
function Write-Skip { param($m) Write-Host "  ○ $m" -ForegroundColor DarkGray }
function Write-Warn { param($m) Write-Host "  ⚠ $m" -ForegroundColor Yellow }

Write-Banner

# ─── Onay ────────────────────────────────────────────────────────────────────
Write-Host "  cem, cemi ve cemir silinecek." -ForegroundColor Yellow
$confirm = Read-Host "  Devam edilsin mi? (y/N)"
if ($confirm -notin @("y","Y","yes","YES")) {
    Write-Host "`n  İptal." -ForegroundColor DarkGray
    Write-Host ""
    exit 0
}

Write-Host ""

# ─── Binary'leri sil ─────────────────────────────────────────────────────────
$removed = 0

function Remove-Bin {
    param([string]$Name)

    $path = (Get-Command $Name -ErrorAction SilentlyContinue)?.Source
    if (-not $path) {
        Write-Skip "$Name  bulunamadı"
        return
    }

    try {
        Remove-Item -Force $path
        Write-Ok "$Name  → $path"
        $script:removed++
    } catch {
        Write-Warn "$Name silinemedi: $path"
        Write-Host "    Manuel: Remove-Item -Force '$path'" -ForegroundColor DarkGray
    }
}

Remove-Bin "cem"
Remove-Bin "cemi"
Remove-Bin "cemir"

# ─── Kurulum dizini (%LOCALAPPDATA%\cem\bin) ──────────────────────────────────
$installDir = "$env:LOCALAPPDATA\cem"
if (Test-Path $installDir) {
    Write-Host ""
    Write-Host "  Kurulum dizini: $installDir" -ForegroundColor DarkGray
    $delDir = Read-Host "  Kurulum dizini de silinsin mi? (y/N)"
    if ($delDir -in @("y","Y","yes")) {
        try {
            Remove-Item -Recurse -Force $installDir
            Write-Ok "Kurulum dizini silindi"
            $removed++
        } catch {
            Write-Warn "Dizin silinemedi: $_"
        }
    } else {
        Write-Skip "Kurulum dizini korundu"
    }
}

# ─── Config klasörü (%USERPROFILE%\.cem) ─────────────────────────────────────
$cemDir = "$env:USERPROFILE\.cem"
if (Test-Path $cemDir) {
    Write-Host ""
    Write-Host "  Config klasörü: $cemDir" -ForegroundColor DarkGray
    $delCfg = Read-Host "  Config ve ayarlar da silinsin mi? (y/N)"
    if ($delCfg -in @("y","Y","yes")) {
        Remove-Item -Recurse -Force $cemDir
        Write-Ok "Config silindi"
    } else {
        Write-Skip "Config korundu → $cemDir"
    }
}

# ─── PATH temizliği ──────────────────────────────────────────────────────────
$userPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
$cleanPath = ($userPath -split ";") |
    Where-Object { $_ -notmatch "\\cem\\bin" -and $_ -ne "" } |
    Join-String -Separator ";"

if ($cleanPath -ne $userPath) {
    [System.Environment]::SetEnvironmentVariable("PATH", $cleanPath, "User")
    Write-Ok "PATH temizlendi"
}

# ─── Proje .cem.yaml ─────────────────────────────────────────────────────────
if (Test-Path ".cem.yaml") {
    Write-Host ""
    $delProj = Read-Host "  Bu dizindeki .cem.yaml silinsin mi? (y/N)"
    if ($delProj -in @("y","Y","yes")) {
        Remove-Item ".cem.yaml"
        Write-Ok ".cem.yaml silindi"
    } else {
        Write-Skip ".cem.yaml korundu"
    }
}

# ─── Sonuç ───────────────────────────────────────────────────────────────────
Write-Host ""

if ($removed -gt 0) {
    Write-Host "  ╭─────────────────────────────────────────────╮" -ForegroundColor Green
    Write-Host "  │   ✓  CEM kaldırıldı.                       │" -ForegroundColor Green
    Write-Host "  ╰─────────────────────────────────────────────╯" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Yeniden kurmak için:" -ForegroundColor White
    Write-Host "  iwr cem.pw/install.ps1 -UseB | iex" -ForegroundColor Cyan
} else {
    Write-Host "  ⚠  Binary silinemedi." -ForegroundColor Yellow
    Write-Host "  Manuel kaldır:" -ForegroundColor DarkGray
    Write-Host "  Remove-Item (Get-Command cem).Source" -ForegroundColor DarkGray
}

Write-Host ""
