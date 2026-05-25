# CLAUDE.md — cem

> Bu dosya bu projeye özgüdür ve **global `~/.claude/CLAUDE.md` kurallarının üstüne yazar**. Aşağıdaki kurallar mutlaktır; global kurallarla çelişen her madde için bu dosya geçerlidir.

---

## Proje Özeti

**CEM** — Birden fazla AI CLI aracını (Claude, Agy, Aider, Gemini, GPT) tek
komutla yöneten Go orchestrator'dır. "Thinker" düşünür, "Writer" yazar; pair
modunda thinker çıktısı writer'a beslenir.

- **Domain:** `cem.pw`
- **Git:** `http://gitlab.makdos.biz/makdos/cem.git`
- **Dil:** Go 1.25 (cobra + lipgloss + yaml.v3)
- **Binary:** 3 ad — `cem`, `cemi` (installer), `cemir` (remover). Hepsi
  aynı kaynaktan derlenir, `main.go` binary adına göre dispatch yapar.

---

## Global Kurallarla Farklılaşan Noktalar

Global `~/.claude/CLAUDE.md` Python + PostgreSQL + FastAPI + nginx merkezli.
Bu proje **Go CLI**'dır → aşağıdaki global maddeler **uygulanmaz**:

| Global madde | Sebep |
|---|---|
| Pydantic / `.env` / Fernet | CLI; secret yok, config = YAML |
| FastAPI middleware / Server-Timing | API değil |
| PostgreSQL / MongoDB / Valkey | Veri katmanı yok |
| Cache dekoratörü / rate limit | İlgisiz |
| Swagger UI / `src/swagger.py` | API değil |

**Uygulanır:** Docker yasağı (bare-metal binary), `nala` zorunluluğu, nginx
production standartları (HTTP/3, Server header, fail2ban) — `cem.pw` sunucusu için.

---

## Dosya Yapısı

```
cem/
├── main.go             — Binary adına göre dispatch + LDFLAGS version
├── config.go           — GlobalConfig + ProjectConfig + ResolvedConfig + KnownTools
├── config_test.go      — ActiveRoles override + KnownTools sanity (5 test)
├── executor.go         — ModeThink/Write/Pair + Run + ReadStdin
├── spinner.go          — TTY-aware tek satır spinner (pair modu için)
├── history.go          — AppendHistory → ~/.cem/history.log (TSV)
├── wizard.go           — RunSetupWizard, InstallTool, RemoveTool, ShowRoles, askYN
├── banner.go           — ASCII art + lipgloss styles + ShowConfigSource
├── cmd_cem.go          — Cobra: rootCmd, rolesCmd, setupCmd, initCmd, statusCmd
├── cmd_doctor.go       — cem doctor: tanı raporu (sistem/roller/araçlar/PATH)
├── cmd_history.go      — cem history: -n N, --clear
├── cmd_cemi.go         — cemi: tool kurulumu (claude/agy/aider/gemini/gpt + all + update)
├── cmd_cemir.go        — cemir: tool kaldırma (tek araç + all)
├── cmd_uninstall.go    — cem uninstall: kendini sil
├── .gitlab-ci.yml      — Canonical CI: test + 7 platform × 3 binary release
├── install.sh / .ps1   — cem.pw/install üzerinden kurulum
├── uninstall.sh / .ps1 — cem.pw/uninstall üzerinden kaldırma
├── Makefile            — build / dev / install / clean / tidy / test
├── go.mod
├── .github/workflows/release.yml  — 7 platform binary + SHA256SUMS
├── nginx/
│   ├── nginx.conf
│   ├── setup.sh
│   ├── snippets/        — ssl.conf, security-headers.conf, block-rules.conf
│   ├── sites-available/ — cem.pw.conf (rotalar: /install /uninstall /r/* /docs /health)
│   └── fail2ban/jail.d/ — cem-nginx.conf
├── OPERATIONS.md
├── README.md
└── todo.md             — Açık görev listesi
```

---

## Kod Kuralları

1. **`package main`** — tek paket. Alt paket açma (gerekmedikçe).
2. **Logger yok** — kullanıcıya `fmt.Println` + lipgloss style (`styleSuccess`,
   `styleError`, `styleDim`, `styleBold` — `wizard.go`'da tanımlı).
3. **Hata mesajları Türkçe** — `styleError.Render("✗ ...")`. Stack trace gösterme.
4. **`exec.Command` ile araç çalıştır** — stdin: input, stdout/stderr: passthrough.
   Pair modunda `cmd.Output()` ile çıktıyı yakala.
5. **Config IO:** `~/.cem/config.yaml` permission `0600`, dir `0755`.
6. **YAML marshal:** `gopkg.in/yaml.v3`. Tag'ler küçük harf snake_case.
7. **Yeni AI aracı eklemek:** `KnownTools` map'ine `ToolMeta` ekle, başka yere
   dokunma — wizard, installer, executor otomatik kullanır.
8. **Cobra subcommand eklemek:** yeni dosya `cmd_<name>.go`, `rootCmd.AddCommand`
   çağrısı `cmd_cem.go`'nun `init()` bloğuna.
9. **Banner sadece help / setup / status / cemi-no-args / cemir-no-args'da
   görünür** — normal komut çıktısında değil.
10. **Build:** her zaman 3 binary üret. Geliştirmede `make build` yeter;
    sistem-wide kurulum için `make install` (sudo).

---

## Yasaklar (bu proje)

- **Docker / docker-compose ekleme** — bare-metal binary.
- **Python köprüsü yazma** — saf Go. AI CLI'ları subprocess olarak çağrılır.
- **Versiyon dosyası ayrı tutma** — sürüm `cmd_cem.go`'da `Version: "1.0.0"`.
- **`go run main.go`** — main.go diğer dosyalara bağımlı. `go run .` kullan.
- **Binary adı değiştirme** — `cem` / `cemi` / `cemir` sabit. `main.go` dispatch
  bunlara bağlı; değişirse install.sh ve release.yml kırılır.
- **`fmt.Errorf` mesajları İngilizce** olabilir (debug için), ama kullanıcıya
  gösterilen `styleError.Render(...)` Türkçe.

---

## Git & Release

- **Canonical repo:** `http://gitlab.makdos.biz/makdos/cem.git`
- **GitHub mirror:** Sadece release.yml çalıştırmak için (Actions GitLab CI'a
  taşınana kadar). Tag `v*.*.*` push edildiğinde 7 platform binary üretir.
- **Binary download:** `cem.pw/r/*` → nginx proxy. Install script'leri bu URL'i
  kullanır, doğrudan GitHub/GitLab release URL'ine bağımlı değildir.
- **Tag formatı:** `v1.2.3` (semver). `make` ile `LDFLAGS` versiyon enjeksiyonu
  henüz yok — eklenecekse `-X main.version=...` ile.

---

## Test / Doğrulama Akışı

```sh
make clean && make build
./build/cem --help          # banner + komut listesi
./build/cemi                # araç listesi (banner ile)
./build/cemir               # kurulu araçlar (banner ile)
./build/cem roles           # config kaynak + aktif roller
```

İlk çalıştırmada wizard açılır; `~/.cem/config.yaml` oluşur. Test dizininde
`.cem.yaml` ile proje override edilir.

---

## Bilinen Eksikler

`todo.md` güncel listeyi tutar. Açık başlıklar:
- `.claude/agents/` ve `.claude/skills/` `autoinstalltrixie` kalıntısı —
  silme/değiştirme kararı bekliyor (`.claude/` gitignore'lı).
- Daha kapsamlı testler: `executor_test.go`, `history_test.go` yok.
- macOS/Linux/Windows entegrasyon testleri yok.

---

*Bu CLAUDE.md değişirse `todo.md`'ye "doc:CLAUDE update" satırı eklenir.*
