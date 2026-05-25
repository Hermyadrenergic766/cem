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

## 7. Açık (yapılmadı)
- [ ] `.claude/agents/` ve `.claude/skills/` — bunlar `autoinstalltrixie`
      projesinin kalıntısı (iso-builder, preseed-engineer, qemu-tester, build-iso,
      test-iso vs). CEM ile ilgisi yok. **Karar bekleniyor:** silinsin mi yoksa
      CEM-spesifik agents (örn. `tool-installer`, `role-switcher`) ile değiştirilsin
      mi? `.claude/hooks/` zaten generic güvenlik hookları — kalabilir.
- [ ] `*_test.go` — birim testi yazılmamış
- [ ] `cem doctor` komutu (kurulu araç + PATH kontrolü)
- [ ] `~/.cem/history.log` (komut geçmişi)
- [ ] `cem -p` için spinner (bubbletea entegrasyonu)
- [ ] `cemir all` (toplu kaldırma)
- [ ] GitLab CI çevrimi (release.yml → .gitlab-ci.yml)
- [ ] LDFLAGS ile versiyon enjeksiyonu (`-X main.version=$(git describe)`)
- [ ] doc:CLAUDE update — başlangıç sürümü yazıldı (2026-05-25)
