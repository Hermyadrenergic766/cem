#!/usr/bin/env bash
##############################################################################
# setup.sh — cem.pw Nginx Tam Kurulum Scripti
# Ubuntu 22.04 / Debian 12 için
#
# Kullanım:
#   chmod +x setup.sh
#   sudo ./setup.sh
##############################################################################

set -euo pipefail

DOMAIN="cem.pw"
WEBROOT="/var/www/${DOMAIN}"
NGINX_SITES="/etc/nginx/sites-available"
NGINX_ENABLED="/etc/nginx/sites-enabled"
EMAIL="admin@cem.pw"           # Let's Encrypt bildirimleri için

# ─── Renkler ─────────────────────────────────────────────────────────────────
bold="\033[1m"; cyan="\033[36m"; green="\033[32m"
yellow="\033[33m"; red="\033[31m"; reset="\033[0m"

step()  { echo -e "\n${cyan}${bold}▶ $1${reset}"; }
ok()    { echo -e "${green}  ✓ $1${reset}"; }
warn()  { echo -e "${yellow}  ⚠ $1${reset}"; }
die()   { echo -e "${red}  ✗ $1${reset}" >&2; exit 1; }

[[ $EUID -ne 0 ]] && die "Root yetkisi gerekli: sudo ./setup.sh"

##############################################################################
step "1/8  Paketleri yükle"
##############################################################################
apt-get update -qq
apt-get install -y -qq \
    nginx \
    certbot \
    python3-certbot-nginx \
    fail2ban \
    openssl \
    ufw

ok "Paketler kuruldu"

##############################################################################
step "2/8  UFW Güvenlik Duvarı"
##############################################################################
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

ok "UFW aktif: SSH(22), HTTP(80), HTTPS(443)"

##############################################################################
step "3/8  DH parametreleri üret (4096-bit)"
##############################################################################
if [[ ! -f /etc/nginx/dhparam.pem ]]; then
    echo "  Bu işlem 2-5 dakika sürebilir..."
    openssl dhparam -out /etc/nginx/dhparam.pem 4096
    chmod 640 /etc/nginx/dhparam.pem
    ok "dhparam.pem oluşturuldu"
else
    warn "dhparam.pem zaten var, atlandı"
fi

##############################################################################
step "4/8  Web dizinlerini oluştur"
##############################################################################
mkdir -p "${WEBROOT}/public"
mkdir -p "${WEBROOT}/scripts"
mkdir -p "${WEBROOT}/docs"
mkdir -p /var/www/certbot

# Hata sayfaları
cat > "${WEBROOT}/public/404.html" << 'HTML'
<!DOCTYPE html>
<html><head><title>404 — cem.pw</title>
<style>body{font-family:monospace;text-align:center;padding:4rem;background:#0d0d0d;color:#ccc}
h1{color:#e040fb;font-size:3rem}a{color:#82b1ff}</style></head>
<body><h1>404</h1><p>Sayfa bulunamadı.</p><a href="/">← cem.pw</a></body></html>
HTML

cat > "${WEBROOT}/public/429.html" << 'HTML'
<!DOCTYPE html>
<html><head><title>429 — cem.pw</title>
<style>body{font-family:monospace;text-align:center;padding:4rem;background:#0d0d0d;color:#ccc}
h1{color:#ffab40;font-size:3rem}a{color:#82b1ff}</style></head>
<body><h1>429</h1><p>Çok fazla istek. 60 saniye bekleyin.</p><a href="/">← cem.pw</a></body></html>
HTML

# Geçici index (gerçek site daha sonra deploy edilir)
cat > "${WEBROOT}/public/index.html" << 'HTML'
<!DOCTYPE html>
<html><head><title>cem.pw</title>
<style>body{font-family:monospace;text-align:center;padding:4rem;background:#0d0d0d;color:#ccc}
pre{color:#e040fb;font-size:0.9rem;text-align:left;display:inline-block}
code{background:#1a1a1a;padding:0.3rem 0.8rem;border-radius:4px;color:#82b1ff}</style></head>
<body>
<pre>
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
</pre>
<h2>⚡ Unified AI Orchestrator</h2>
<p><code>curl -fsSL cem.pw/install | sh</code></p>
<p><code>irm cem.pw/install.ps1 | iex</code></p>
</body></html>
HTML

chown -R www-data:www-data "${WEBROOT}"
ok "Web dizinleri hazır: ${WEBROOT}"

##############################################################################
step "5/8  Install scriptlerini kopyala"
##############################################################################
SCRIPT_DIR="$(dirname "$0")"

if [[ -f "${SCRIPT_DIR}/../cem/install.sh" ]]; then
    cp "${SCRIPT_DIR}/../cem/install.sh"  "${WEBROOT}/scripts/install.sh"
    cp "${SCRIPT_DIR}/../cem/install.ps1" "${WEBROOT}/scripts/install.ps1"
    chmod 644 "${WEBROOT}/scripts/install.sh" "${WEBROOT}/scripts/install.ps1"
    ok "Scriptler kopyalandı"
else
    warn "install.sh bulunamadı — manuel kopyala: ${WEBROOT}/scripts/"
fi

##############################################################################
step "6/8  Nginx yapılandırması"
##############################################################################

# Snippets kopyala
mkdir -p /etc/nginx/snippets
cp snippets/ssl.conf               /etc/nginx/snippets/
cp snippets/security-headers.conf  /etc/nginx/snippets/
cp snippets/block-rules.conf       /etc/nginx/snippets/

# Site config kopyala
cp "sites-available/cem.pw.conf" "${NGINX_SITES}/${DOMAIN}.conf"

# Default site'ı devre dışı bırak
rm -f "${NGINX_ENABLED}/default"

# Sembolik link oluştur
ln -sfn "${NGINX_SITES}/${DOMAIN}.conf" "${NGINX_ENABLED}/${DOMAIN}.conf"

# Test
nginx -t || die "Nginx syntax hatası! Kontrol et: nginx -T"
ok "Nginx yapılandırması geçerli"

##############################################################################
step "7/8  SSL Sertifikası (Let's Encrypt)"
##############################################################################

# Önce HTTP ile test et (certbot webroot)
systemctl start nginx

if certbot --nginx \
    -d "${DOMAIN}" \
    -d "www.${DOMAIN}" \
    --email "${EMAIL}" \
    --agree-tos \
    --no-eff-email \
    --redirect \
    --staple-ocsp; then
    ok "SSL sertifikası alındı"
else
    warn "Certbot başarısız — self-signed ile devam ediliyor"
    # Self-signed (development)
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/letsencrypt/live/${DOMAIN}/privkey.pem \
        -out    /etc/letsencrypt/live/${DOMAIN}/fullchain.pem \
        -subj   "/CN=${DOMAIN}" 2>/dev/null
    mkdir -p /etc/letsencrypt/live/${DOMAIN}
fi

# Otomatik yenileme cron
(crontab -l 2>/dev/null; echo "0 3 * * 1 certbot renew --quiet --post-hook 'systemctl reload nginx'") | crontab -
ok "Sertifika yenileme cron'u eklendi (her Pazartesi 03:00)"

##############################################################################
step "8/8  Fail2ban yapılandırması"
##############################################################################
mkdir -p /etc/fail2ban/jail.d
mkdir -p /etc/fail2ban/filter.d

cp fail2ban/jail.d/cem-nginx.conf /etc/fail2ban/jail.d/

# Filtreleri ayrı dosyalara yaz
cat > /etc/fail2ban/filter.d/nginx-limit-req.conf << 'F2B'
[Definition]
failregex = ^<HOST> .* "(GET|POST|HEAD) .+" 429
ignoreregex =
F2B

cat > /etc/fail2ban/filter.d/nginx-bad-request.conf << 'F2B'
[Definition]
failregex = ^<HOST> .* "(GET|POST|HEAD) .+" (400|444|405)
ignoreregex =
F2B

cat > /etc/fail2ban/filter.d/nginx-404-flood.conf << 'F2B'
[Definition]
failregex = ^<HOST> .* "(GET|POST) .+" 404
ignoreregex = \.(css|js|png|jpg|ico|svg|woff2?)
F2B

cat > /etc/fail2ban/filter.d/nginx-download-flood.conf << 'F2B'
[Definition]
failregex = ^\S+ <HOST> "(GET|HEAD) /install" (200|304)
ignoreregex =
F2B

systemctl enable fail2ban
systemctl restart fail2ban
ok "Fail2ban aktif"

##############################################################################
# Nginx'i yeniden başlat
##############################################################################
systemctl enable nginx
systemctl restart nginx

##############################################################################
echo -e "\n${green}${bold}"
cat << 'DONE'
  ╭──────────────────────────────────────────────────╮
  │                                                  │
  │   ✓  cem.pw kurulumu tamamlandı!                │
  │                                                  │
  ╰──────────────────────────────────────────────────╯
DONE
echo -e "${reset}"

echo "  Test:"
echo -e "  ${cyan}curl -I https://cem.pw/health${reset}"
echo -e "  ${cyan}curl -fsSL https://cem.pw/install | head -5${reset}"
echo ""
echo "  Log dosyaları:"
echo "  /var/log/nginx/cem.pw.access.log"
echo "  /var/log/nginx/cem.pw.downloads.log"
echo "  /var/log/nginx/cem.pw.error.log"
echo ""
echo "  Fail2ban durumu:"
echo -e "  ${cyan}fail2ban-client status${reset}"
echo ""
echo "  SSL test:"
echo -e "  ${cyan}https://www.ssllabs.com/ssltest/analyze.html?d=cem.pw${reset}"
echo ""
