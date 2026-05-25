#!/usr/bin/env sh
# ⚡ CEM — Compose · Execute · Multiplex
# curl -fsSL cem.pw/install.sh | sh

set -e

REPO="https://github.com/muslu/cem"
RELEASE_URL="https://github.com/muslu/cem/releases/latest/download"
INSTALL_DIR="/usr/local/bin"
FALLBACK_DIR="$HOME/.local/bin"

# ─── Renkler ─────────────────────────────────────────────────────────────────
ESC=$(printf '\033')
reset="${ESC}[0m"; bold="${ESC}[1m"
cyan="${ESC}[36m"; green="${ESC}[32m"; yellow="${ESC}[33m"; red="${ESC}[31m"; dim="${ESC}[2m"

info()    { printf "${cyan}  →${reset} %s\n" "$1"; }
ok()      { printf "${green}  ✓${reset} ${bold}%s${reset}\n" "$1"; }
warn()    { printf "${yellow}  ⚠${reset} %s\n" "$1"; }
die()     { printf "${red}  ✗${reset} %s\n" "$1" >&2; exit 1; }
step()    { printf "\n${bold}  %s${reset}\n" "$1"; }

# ─── ASCII Banner ─────────────────────────────────────────────────────────────
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
printf "\n${bold}  ⚡ Compose · Execute · Multiplex${reset}  ${cyan}·  cem.pw${reset}\n"
printf "${dim}  One command, many AIs.${reset}\n\n"
printf "  ${dim}────────────────────────────────────────${reset}\n\n"

# ─── Platform ────────────────────────────────────────────────────────────────
step "1/4  Platform tespit ediliyor"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  OS_NAME="linux"  ;;
  darwin) OS_NAME="darwin" ;;
  *)      die "Desteklenmeyen OS: $OS" ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  armv7l)        ARCH_NAME="arm"   ;;
  *)             die "Desteklenmeyen mimari: $ARCH" ;;
esac

info "OS   : $OS_NAME"
info "Arch : $ARCH_NAME"

BINARY="cem-${OS_NAME}-${ARCH_NAME}"
URL="$RELEASE_URL/$BINARY"

# ─── İndir ───────────────────────────────────────────────────────────────────
step "2/4  İndiriliyor"
info "URL: ${dim}$URL${reset}"

TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -fsSL --progress-bar "$URL" -o "$TMP" || die "İndirme başarısız"
elif command -v wget >/dev/null 2>&1; then
  wget -q --show-progress "$URL" -O "$TMP"  || die "İndirme başarısız"
else
  die "curl veya wget bulunamadı"
fi

chmod +x "$TMP"

# ─── Kur ─────────────────────────────────────────────────────────────────────
step "3/4  Kuruluyor"

do_install() {
  DIR="$1"
  mkdir -p "$DIR" 2>/dev/null || return 1
  cp "$TMP" "$DIR/cem"   2>/dev/null || return 1
  cp "$DIR/cem" "$DIR/cemi"
  cp "$DIR/cem" "$DIR/cemir"
  chmod +x "$DIR/cem" "$DIR/cemi" "$DIR/cemir"
  INSTALLED="$DIR"
  return 0
}

INSTALLED=""

if do_install "$INSTALL_DIR"; then
  ok "Kuruldu → $INSTALL_DIR"
elif command -v sudo >/dev/null 2>&1; then
  info "sudo ile deneniyor..."
  sudo sh -c "
    cp '$TMP' '$INSTALL_DIR/cem'
    cp '$INSTALL_DIR/cem' '$INSTALL_DIR/cemi'
    cp '$INSTALL_DIR/cem' '$INSTALL_DIR/cemir'
    chmod +x '$INSTALL_DIR/cem' '$INSTALL_DIR/cemi' '$INSTALL_DIR/cemir'
  " && INSTALLED="$INSTALL_DIR" && ok "Kuruldu → $INSTALL_DIR"
fi

if [ -z "$INSTALLED" ]; then
  warn "/usr/local/bin yazılamadı → ~/.local/bin kullanılıyor"
  do_install "$FALLBACK_DIR" || die "Kurulum başarısız"
fi

# ─── PATH ────────────────────────────────────────────────────────────────────
step "4/4  PATH"

if echo ":$PATH:" | grep -q ":$INSTALLED:"; then
  ok "PATH zaten doğru"
else
  warn "$INSTALLED PATH'de değil"
  case "$SHELL" in
    */zsh)  RC="$HOME/.zshrc"  ;;
    */fish) RC="$HOME/.config/fish/config.fish" ;;
    *)      RC="$HOME/.bashrc" ;;
  esac

  if echo "$SHELL" | grep -q fish; then
    LINE="fish_add_path $INSTALLED"
  else
    LINE="export PATH=\"$INSTALLED:\$PATH\""
  fi

  printf "\n  ${cyan}echo '%s' >> %s${reset}\n" "$LINE" "$RC"
  printf "  Otomatik eklensin mi? (y/N): "
  read -r R
  if [ "$R" = "y" ] || [ "$R" = "Y" ]; then
    echo "$LINE" >> "$RC"
    ok "PATH güncellendi — yeni terminal açın"
  fi
fi

# ─── Bitti ───────────────────────────────────────────────────────────────────
printf "\n${green}${bold}"
cat << 'DONE'
  ╭─────────────────────────────────────────────╮
  │                                             │
  │   ✓  CEM kuruldu!   ·   cem.pw             │
  │                                             │
  ╰─────────────────────────────────────────────╯
DONE
printf "${reset}\n"

printf "  ${bold}Kullanım:${reset}\n\n"
printf "  ${cyan}cem \"soru\"${reset}         → thinker AI (ilk çalıştırmada wizard)\n"
printf "  ${dim}cem -w \"görev\"${reset}     → writer AI\n"
printf "  ${dim}cem -p \"görev\"${reset}     → pair modu\n"
printf "  ${dim}cemi claude${reset}        → Claude kur\n"
printf "  ${dim}cemi agy${reset}           → Agy kur\n"
printf "  ${dim}cem roles${reset}          → kim ne yapıyor?\n"
printf "\n  ${dim}Doküman → https://github.com/muslu/cem${reset}\n\n"
