# ⚡ CEM — Compose · Execute · Multiplex · [cem.pw](https://cem.pw)

```
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
```

**Tek komut, çok AI.** Claude düşünsün, Antigravity yazsın — ya da istediğin
kombinasyon. Her proje farklı bir yapı kullanabilir (`.cem.yaml`).

> English README: [README.md](README.md)

---

## Kur

```sh
# macOS & Linux
curl -fsSL cem.pw/install | sh

# Windows (PowerShell)
irm cem.pw/install.ps1 | iex
```

İlk `cem` komutunda wizard açılır.

---

## Kullan

```sh
cem "soru"           # thinker AI — varsayılan, "think" yazmaya gerek yok
cem -w "görev"       # writer AI
cem -p "görev"       # pair: düşün → yaz (writer thinker çıktısını alır)
cem -f dosya.py      # dosyayı thinker'a gönder
cem -wf dosya.py     # dosyayı writer'a gönder
cat kod.py | cem -p  # pipe ile pair

cem roles               # aktif rolleri göster
cem roles claude agy    # global rolleri değiştir
cem roles --here c agy  # sadece bu proje için (.cem.yaml)
cem init                # interaktif .cem.yaml oluştur
cem status              # aktif yapılandırma
cem doctor              # tanı raporu (sistem + roller + araçlar + PATH)
cem history             # son 20 komut
cem history -n 100      # son 100
cem setup               # kurulum sihirbazını yeniden çalıştır
cem uninstall           # cem/cemi/cemir'i sistemden kaldır
```

### Installer (`cemi`)

```sh
cemi                    # bilinen araçları listele (kurulu vs eksik)
cemi claude             # Claude Code kur
cemi agy                # Antigravity kur
cemi all                # hepsini kur (her birinde onay sorar)
cemi update             # hepsini güncelle
cemi update agy         # sadece Antigravity'i güncelle
```

### Remover (`cemir`)

```sh
cemir                   # kurulu araçları listele
cemir claude            # Claude Code'u kaldır
cemir all               # tüm kurulu araçları kaldır (onaylı)
```

---

## Desteklenen AI CLI'ları

| Key | Araç | Notlar |
|---|---|---|
| `claude` | Anthropic Claude Code | npm |
| `agy` | **Antigravity** (Google) | Eski Gemini CLI — otonom kodlama ajanı |
| `aider` | Aider | Açık kaynak eş programlama AI (pip) |
| `gemini` | Google Gemini CLI | ⚠ Kişisel kullanım **2026-06-16**'da sonlanıyor — `agy` tercih edin |
| `gpt` | OpenAI Codex CLI | Eski `gpt` CLI'nın yeni adı (npm) |
| `goose` | Block Goose | Açık kaynak otonom ajan (pip) |
| `cody` | Sourcegraph Cody | npm |
| `continue` | Continue.dev | VSCode/JetBrains otopilot (npm) |
| `openhands` | OpenHands | Eski OpenDevin — otonom yazılım mühendisi ajanı (pip) |
| `cursor` | Cursor | Cursor terminal ajanı (npm) |

---

## Yapılandırma

- `~/.cem/config.yaml` — global yapılandırma (varsayılan roller + kurulu araçlar)
- `.cem.yaml` — proje override; repo köküne konur

Proje değerleri global'i ezer; boş bırakılan alanlar global'e düşer. Örnek
`.cem.yaml`:

```yaml
roles:
  thinker: gemini
  writer:  aider
```

`cem roles` ile hangi yapılandırmanın aktif olduğunu ve nereden geldiğini
görebilirsin.

---

## `cem -p` nasıl çalışır

```
     sen yaz ──► thinker AI ──► analiz ──► writer AI ──► son kod
     (girdi)    (claude)        (metin)    (agy)
```

Thinker'ın tüm çıktısı writer'ın istemine `--- Thinker analizi ---` etiketiyle
eklenir; writer gerekçeyle birlikte kod üretir.

---

## Geçmiş

Her `cem` çalışmasında `~/.cem/history.log`'a bir satır eklenir:

```
2026-05-25T13:42:11Z    pair    claude+agy    0    middleware'i async refactorle …
```

Görmek için `cem history -n 50`, silmek için `cem history --clear`.

---

## Kaynaktan kurma

```sh
git clone https://github.com/muslu/cem.git
cd cem
make build            # ./build/{cem,cemi,cemir} üretir
make dev              # ~/.local/bin'e kur
make install          # /usr/local/bin'e kur (sudo)
make test             # go test ./...
```

Sürüm `git describe --tags --always --dirty` çıktısından enjekte edilir:

```sh
./build/cem --version
# cem version v1.0.0
```

---

## Lisans

MIT — bkz. [LICENSE](LICENSE).

## Bağlantılar

- Site: [cem.pw](https://cem.pw)
- Kaynak: [github.com/muslu/cem](https://github.com/muslu/cem)
- Issue: [github.com/muslu/cem/issues](https://github.com/muslu/cem/issues)
