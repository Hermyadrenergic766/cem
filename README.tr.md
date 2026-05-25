# ⚡ CEM — Compose · Execute · Multiplex

```
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
```

> **One command, many AIs.**
> CEM — birden fazla AI CLI aracını (Claude · Antigravity · Codex · Cursor) tek bir komutla yöneten Go orchestrator. Bir AI **düşünür**, bir AI **yazar**; `pair` modunda düşünenin analizi yazana beslenir.

- Domain: [cem.pw](https://cem.pw)
- Source: <https://github.com/muslu/cem>
- Türkçe: [README.tr.md](README.tr.md)

---

## Install

**macOS / Linux:**
```sh
curl -fsSL cem.pw/install | sh
```

**WSL (Windows Subsystem for Linux):**
```sh
curl -fsSL cem.pw/install | sh
```
> WSL bir Linux ortamı; yukarıdaki Linux komutu doğrudan çalışır. Native Windows kullanıyorsan PowerShell'e geç.

**Windows — PowerShell zorunlu, CMD/Git Bash değil:**
```powershell
irm cem.pw/install | iex
```
> ⚠ Komutu **PowerShell**'de çalıştır. `cmd.exe` veya Git Bash'te `irm` yoktur. Prompt'un `PS C:\` ile başlaması PowerShell'de olduğunu gösterir.

Installer OS/arch tespit eder, 3 binary (`cem`, `cemi`, `cemir`) indirir, Unix/WSL'de `/usr/local/bin`'e (veya yazılamazsa `~/.local/bin`'e), Windows'ta `%LOCALAPPDATA%\cem\bin`'e koyar. Sunucudaki UA tespiti otomatik: PowerShell → `.ps1`, curl/wget → `.sh`.

### Update / Uninstall

```sh
cem update      # cem.pw'den son sürümü çeker (GitHub API ile mevcut/yeni karşılaştırır)
cem uninstall   # 3 binary + config klasörü
```

---

## Quick Start

```sh
cem "fibonacci nedir?"            # thinker (düşünen AI tek başına)
cem -w "fibonacci.py yaz"         # writer (yazan AI tek başına)
cem -p "fibonacci.py yaz"         # pair: thinker → writer
```

İlk çalıştırma wizard'ı açar; rolleri sonradan değiştir:
```sh
cem roles claude agy              # global: thinker=claude, writer=agy
cem roles --here claude codex     # sadece bu dizin için (.cem.yaml)
cem init                          # proje-spesifik wizard
```

`pair` modu akıllı: **thinker == writer** ise writer atlanır; ne soruda ne thinker çıktısında kod istemi yoksa writer atlanır (boşa LLM çağrısı yok).

---

## Supported AI CLIs

| Anahtar | Araç | Kurulum kaynağı | Non-interactive |
|---|---|---|---|
| `claude` | **Claude Code** (Anthropic) | [native installer](https://code.claude.com/docs/en/quickstart) — auto-update | `claude -p` (stdin) |
| `agy` | **Antigravity** (Google) | [native installer](https://antigravity.google/docs/cli-getting-started) | `agy -p` (stdin) |
| `gpt` | **Codex** (OpenAI) | `npm i -g @openai/codex` | `codex exec "prompt"` |
| `cursor` | **Cursor agent** | [native installer](https://cursor.com/cli) | `cursor-agent -p "prompt"` |

```sh
cemi                              # mevcut & yüklenebilir
cemi claude                       # tek araç (önkoşulları algılar: npm/Node)
cemi all                          # 4'ünü birden
cemir agy                         # tek araç kaldır (shell-install da silinir)
cemir all                         # hepsini kaldır
```

`cemi <tool>` çağrısı **npm/Node** gerektiriyor ama bulunamıyorsa (veya çok eski — npm 3 gibi) → otomatik olarak `winget` / `brew` / NodeSource `apt-get` ile Node LTS kurar (kullanıcı onayıyla).

---

## Pair Mode

```sh
cem -p "binary search'ü TypeScript'te yaz"
```

1. 🧠 **Thinker** (örn. `claude`) sorunu çözer — algoritma, edge cases, tip seçimi
2. ✍️ **Writer** (örn. `agy`) thinker'ın analizini + asıl soruyu input alır, kodu yazar

Skip kuralları:

| Durum | Davranış |
|---|---|
| `thinker == writer` (ikisi de `claude`) | Writer atlanır (duplikasyon engeli) |
| Soru kod istemiyor ve thinker çıktısında \`\`\` yok | Writer atlanır |
| Diğer | Writer thinker analizini bağlam alarak çalışır |

---

## Diagnostics

```sh
cem doctor                        # sistem + roller + araçlar + PATH raporu
cem status                        # özet
cem roles                         # aktif roller
cem history                       # son komutlar (TSV)
cem history -n 50                 # son 50
cem history --clear
```

`~/.cem/history.log` — tab-separated log.

---

## API Key Management & Auto-Rotation

Büyük projeleri yarıda kestirmeden yürütmek için her provider'a birden fazla key ekleyebilirsin. Bir key rate-limit'e takıldığında **otomatik olarak sıradakine geçer** — manuel müdahale gerekmez.

```sh
cem keys add anthropic             # interaktif: key + opsiyonel etiket
cem keys add openai
cem keys list                      # mask'li görünüm
cem keys remove anthropic 1        # 1. anthropic key'i sil
```

Desteklenen provider'lar:
| Provider | Env var | Hangi tool |
|---|---|---|
| `anthropic` | `ANTHROPIC_API_KEY` | Claude (`claude`) |
| `openai` | `OPENAI_API_KEY` | Codex (`gpt`/`codex`) |

> `agy` (Antigravity) ve `cursor` Google/Cursor OAuth ile çalışıyor; CLI'lar resmi key env'i yayımlamıyor → rotasyon kapsamında değil.

Rotasyon algılaması: stderr'de `rate limit`, `429`, `quota`, `too many requests`, `overloaded` görülürse cem sıradaki key'i dener. Tüm key'ler bitmişse son hatayı yansıtır.

---

## Config

`~/.cem/config.yaml` (global) ve proje kökünde `.cem.yaml` (override):

```yaml
roles:
  thinker: claude
  writer:  agy

tools:
  claude:
    command: claude
    version: 2.1.143
  agy:
    # Native installer PATH'i güncellemezse, post-install yakaladığımız mutlak yol:
    command: C:\Users\Muslu\AppData\Local\agy\bin\agy.exe
    version: 1.2.0

api_keys:
  anthropic:
    - value: sk-ant-...
      label: personal
    - value: sk-ant-...
      label: company-backup
  openai:
    - value: sk-proj-...
```

---

## Build From Source

```sh
git clone https://github.com/muslu/cem.git
cd cem
make build                        # 3 binary → build/
make install                      # /usr/local/bin (sudo)
go test ./...
```

Sürüm `git describe --tags --always --dirty` çıktısından LDFLAGS ile enjekte edilir.

---

## IDE entegrasyonları

cem'i terminal yerine editör içinden çağır:

| Editör | Kurulum | Kısayollar |
|---|---|---|
| **PyCharm / IntelliJ IDEA / GoLand / WebStorm / vb.** | `cem-intellij-<sürüm>.zip` indir → [Releases](https://github.com/muslu/cem/releases/latest) → Settings → Plugins → ⚙ → Install Plugin from Disk | `Ctrl+Alt+I/W/P` |
| **VS Code** | `cem-vscode-<sürüm>.vsix` indir → `code --install-extension cem-vscode-<sürüm>.vsix` | `Ctrl+Alt+I/W/P` |
| **Claude Desktop / Cursor / Continue** (MCP) | `cem-mcp-<os>-<arch>` indir → host'unun MCP config'ine ekle | — (chat'ten tool call) |
| **Antigravity IDE** | Şimdilik dahili terminal; MCP desteği yol haritasında | — |

Adım adım kurulum için: [docs/IDE.tr.md](docs/IDE.tr.md), Antigravity için: [docs/ANTIGRAVITY.md](docs/ANTIGRAVITY.md).

---

## License

MIT — see [LICENSE](LICENSE).
