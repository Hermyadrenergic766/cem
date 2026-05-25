#!/usr/bin/env sh
# ⚡ CEM — Kaldırma Scripti
# curl -fsSL cem.pw/uninstall.sh | sh

set -e

ESC=$(printf '\033')
reset="${ESC}[0m"; bold="${ESC}[1m"
cyan="${ESC}[36m"; green="${ESC}[32m"; yellow="${ESC}[33m"; red="${ESC}[31m"; dim="${ESC}[2m"

ok()   { printf "${green}  ✓${reset} %s\n" "$1"; }
skip() { printf "${dim}  ○${reset} %s\n" "$1"; }
warn() { printf "${yellow}  ⚠${reset} %s\n" "$1"; }
die()  { printf "${red}  ✗${reset} %s\n" "$1" >&2; exit 1; }

# ─── Banner ──────────────────────────────────────────────────────────────────
printf "\n${cyan}${bold}"
cat << 'BANNER'
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
BANNER
printf "${reset}"
printf "\n${bold}  🗑  Kaldırma Scripti${reset}  ${dim}·  cem.pw${reset}\n\n"
printf "  ${dim}────────────────────────────────────────${reset}\n\n"

# ─── Onay ────────────────────────────────────────────────────────────────────
printf "  ${yellow}cem, cemi ve cemir silinecek.${reset}\n"
printf "  Devam edilsin mi? (y/N): "
read -r CONFIRM
case "$CONFIRM" in
  y|Y|yes|YES) ;;
  *) printf "\n  ${dim}İptal.${reset}\n\n"; exit 0 ;;
esac

printf "\n"

# ─── Binary'leri sil ─────────────────────────────────────────────────────────
REMOVED=0

remove_bin() {
    NAME="$1"
    PATH_BIN=$(command -v "$NAME" 2>/dev/null || true)

    if [ -z "$PATH_BIN" ]; then
        skip "$NAME  bulunamadı"
        return
    fi

    # Önce izinsiz dene, olmadı sudo ile
    if rm -f "$PATH_BIN" 2>/dev/null; then
        ok "$NAME  → $PATH_BIN"
        REMOVED=$((REMOVED + 1))
    elif command -v sudo >/dev/null 2>&1; then
        if sudo rm -f "$PATH_BIN" 2>/dev/null; then
            ok "$NAME  → $PATH_BIN  ${dim}(sudo)${reset}"
            REMOVED=$((REMOVED + 1))
        else
            warn "$NAME silinemedi: $PATH_BIN"
            printf "    ${dim}Manuel: sudo rm %s${reset}\n" "$PATH_BIN"
        fi
    else
        warn "$NAME silinemedi (yetki yok): $PATH_BIN"
    fi
}

remove_bin cem
remove_bin cemi
remove_bin cemir

# ─── Config klasörü ──────────────────────────────────────────────────────────
CEM_DIR="$HOME/.cem"

if [ -d "$CEM_DIR" ]; then
    printf "\n  Config klasörü: ${dim}%s${reset}\n" "$CEM_DIR"
    printf "  Config ve ayarlar da silinsin mi? (y/N): "
    read -r DEL_CFG
    case "$DEL_CFG" in
      y|Y|yes)
        rm -rf "$CEM_DIR"
        ok "Config silindi"
        ;;
      *)
        skip "Config korundu → $CEM_DIR"
        ;;
    esac
fi

# ─── Proje .cem.yaml ─────────────────────────────────────────────────────────
if [ -f ".cem.yaml" ]; then
    printf "\n  Bu dizinde .cem.yaml var.\n"
    printf "  Silinsin mi? (y/N): "
    read -r DEL_PROJ
    case "$DEL_PROJ" in
      y|Y|yes)
        rm -f ".cem.yaml"
        ok ".cem.yaml silindi"
        ;;
      *)
        skip ".cem.yaml korundu"
        ;;
    esac
fi

# ─── Sonuç ───────────────────────────────────────────────────────────────────
printf "\n"

if [ "$REMOVED" -gt 0 ]; then
    printf "${green}${bold}"
    cat << 'DONE'
  ╭─────────────────────────────────────────────╮
  │   ✓  CEM kaldırıldı.                       │
  ╰─────────────────────────────────────────────╯
DONE
    printf "${reset}\n"
    printf "  Yeniden kurmak için:\n\n"
    printf "  ${cyan}curl -fsSL cem.pw/install.sh | sh${reset}\n\n"
else
    printf "${yellow}  ⚠  Binary silinemedi.${reset}\n"
    printf "  ${dim}Manuel kaldır:${reset}\n"
    printf "  ${dim}sudo rm \$(which cem) \$(which cemi) \$(which cemir)${reset}\n\n"
fi
