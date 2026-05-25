# CEM — Yapılacaklar

## 1. Klasör & Dosya Düzeni
- [x] `uninstall/cmd_uninstall.go` → root'a taşı (`cmd_uninstall.go`)
- [x] `uninstall/uninstall.sh` → root'a taşı
- [x] `uninstall/uninstall.ps1` → root'a taşı
- [x] `uninstall/` klasörünü sil
- [x] `nginx/` altını düzenle (snippets/, sites-available/, fail2ban/ alt klasörleri)

## 2. Eksik Çekirdek Go Dosyaları
- [x] `main.go` → binary adına göre dispatch (cem / cemi / cemir)
- [x] `config.go` → GlobalConfig + ProjectConfig + ResolvedConfig + Roles + InstalledTool + KnownTools + LoadConfig / saveGlobalConfig / loadGlobalConfig / SaveProjectConfig
- [x] `executor.go` → ModeThink/ModeWrite/ModePair + Run + ReadStdin

## 3. Build & Bağımlılıklar
- [x] `go.mod` doldur (cobra, lipgloss, yaml.v3)
- [x] `go mod tidy`
- [x] `Makefile` (build / dev / install / clean)
- [x] `go build` testi — 3 binary üretildi (5.6 MB → 3.9 MB ldflags ile)

## 4. CI/CD
- [x] `.github/workflows/release.yml` (7 platform binary + SHA256SUMS)

## 5. Doğrulama
- [x] `./build/cem --help` çalışıyor
- [x] `./build/cemi --version` → cemi version 1.0.0
- [x] `./build/cemir --version` → cemir version 1.0.0

## 6. Ek Görevler
- [x] `.claude/` agents / skill / hooks kontrolü
- [x] `CLAUDE.md` güncelle (proje-spesifik kılavuz)
- [x] `install.sh` + `install.ps1` → git URL `gitlab.makdos.biz/makdos/cem`,
      binary download `cem.pw/r`

## 7. Yeni Tamamlananlar (2026-05-25 oturumu)
- [x] `cem doctor` komutu (sistem + roller + araçlar PATH + binary'ler)
- [x] `cemir all` toplu kaldırma (onaylı + hatalı özet)
- [x] LDFLAGS versiyon enjeksiyonu (`-X main.version=$(git describe)`)
- [x] `config_test.go` — 5 test (ActiveRoles override + KnownTools sanity)
- [x] `.gitlab-ci.yml` — 7 platform binary + Release tag
- [x] `.gitignore` CEM'e özel yeniden yazıldı
- [x] Git init + gitlab.makdos.biz/makdos/cem origin
- [x] doc:CLAUDE update — proje-spesifik kılavuz yazıldı (2026-05-25)

## 8. Açık (sonraya bırakıldı)
- [ ] `.claude/agents/` ve `.claude/skills/` — `autoinstalltrixie` kalıntısı.
      Sil/değiştir kararı bekleniyor. `.claude/` artık gitignore'lı.
- [ ] `~/.cem/history.log` (komut geçmişi)
- [ ] `cem -p` için spinner (bubbletea entegrasyonu)
- [ ] `install.sh` ve `install.ps1`'da `cem.pw/r/` proxy henüz sunucuda yok;
      ilk release sonrası nginx site config'inde yol mapping eklenecek
