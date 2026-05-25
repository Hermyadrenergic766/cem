# ⚡ CEM — Compose · Execute · Multiplex · [cem.pw](https://cem.pw)

```
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
```

**One command, many AIs.** Make Claude think, let Antigravity write — or any
combination you prefer. Switch per project with a single YAML file.

> Türkçe README: [README.tr.md](README.tr.md)

---

## Install

```sh
# macOS & Linux
curl -fsSL cem.pw/install | sh

# Windows (PowerShell)
irm cem.pw/install.ps1 | iex
```

A setup wizard launches the first time you run `cem`.

---

## Use

```sh
cem "question"          # thinker AI — the default; you don't write "think"
cem -w "task"           # writer AI
cem -p "task"           # pair: think → write (writer receives thinker's analysis)
cem -f file.py          # send file contents to thinker
cem -wf file.py         # send file contents to writer
cat code.py | cem -p    # pipe into pair mode

cem roles               # show active roles
cem roles claude agy    # change roles globally
cem roles --here c agy  # change for this project only (.cem.yaml)
cem init                # create .cem.yaml interactively
cem status              # active configuration
cem doctor              # diagnostic report (system + roles + tools + PATH)
cem history             # last 20 commands
cem history -n 100      # last 100
cem setup               # rerun the install wizard
cem uninstall           # remove cem/cemi/cemir from your system
```

### Installer (`cemi`)

```sh
cemi                    # list known tools (installed vs missing)
cemi claude             # install Claude Code
cemi agy                # install Antigravity
cemi all                # install everything (with per-tool confirmation)
cemi update             # update everything
cemi update agy         # update only Antigravity
```

### Remover (`cemir`)

```sh
cemir                   # list installed tools
cemir claude            # remove Claude Code
cemir all               # remove every installed tool (with confirmation)
```

---

## Supported AI CLIs

| Key | Tool | Notes |
|---|---|---|
| `claude` | Anthropic Claude Code | npm |
| `agy` | **Antigravity** (Google) | Formerly Gemini CLI — autonomous coding agent |
| `aider` | Aider | Open-source pair-programming AI (pip) |
| `gemini` | Google Gemini CLI | ⚠ Personal use ends **2026-06-16** — prefer `agy` |
| `gpt` | OpenAI Codex CLI | Renamed from the `gpt` CLI (npm) |
| `goose` | Block Goose | Open-source autonomous agent (pip) |
| `cody` | Sourcegraph Cody | npm |
| `continue` | Continue.dev | Autopilot for VSCode/JetBrains (npm) |
| `openhands` | OpenHands | Formerly OpenDevin — autonomous SWE agent (pip) |
| `cursor` | Cursor | Cursor terminal agent (npm) |

---

## Configuration

- `~/.cem/config.yaml` — global config (default roles + installed tools)
- `.cem.yaml` — project override; lives at the repo root

Project values take precedence; any field left out falls back to the global
config. Example `.cem.yaml`:

```yaml
roles:
  thinker: gemini
  writer:  aider
```

Run `cem roles` to see which config is active and where it comes from.

---

## How `cem -p` works

```
       you type ──► thinker AI ──► analysis ──► writer AI ──► final code
       (input)      (claude)       (text)        (agy)
```

The thinker's full output is appended to the writer's prompt as
`--- Thinker analysis ---`, giving the writer reasoning context before it
produces code.

---

## History

Every `cem` invocation appends a line to `~/.cem/history.log`:

```
2026-05-25T13:42:11Z    pair    claude+agy    0    refactor middleware to async …
```

Inspect with `cem history -n 50` or wipe with `cem history --clear`.

---

## Build from source

```sh
git clone https://github.com/muslu/cem.git
cd cem
make build            # builds ./build/{cem,cemi,cemir}
make dev              # installs to ~/.local/bin
make install          # installs to /usr/local/bin (sudo)
make test             # go test ./...
```

Version is injected from `git describe --tags --always --dirty`:

```sh
./build/cem --version
# cem version v1.0.0
```

---

## License

MIT — see [LICENSE](LICENSE).

## Links

- Site: [cem.pw](https://cem.pw)
- Source: [github.com/muslu/cem](https://github.com/muslu/cem)
- Issues: [github.com/muslu/cem/issues](https://github.com/muslu/cem/issues)
