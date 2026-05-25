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
> CEM is a Go orchestrator that drives multiple AI CLIs (Claude · Antigravity · Codex · Cursor) from a single command. One AI **thinks**, another **writes**; in `pair` mode the thinker's analysis feeds the writer.

- Domain: [cem.pw](https://cem.pw)
- Source: <https://github.com/muslu/cem>
- Türkçe sürüm: [README.tr.md](README.tr.md)

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
> WSL is a Linux environment; the Linux command above runs as-is. Use the PowerShell command below only on **native** Windows.

**Windows — PowerShell required (not CMD, not Git Bash):**
```powershell
irm cem.pw/install | iex
```
> ⚠ Run this in **PowerShell**. `irm` does not exist in `cmd.exe` or Git Bash. If your prompt shows `PS C:\` you are in PowerShell.

The installer detects OS/arch, downloads the three binaries (`cem`, `cemi`, `cemir`) from `cem.pw/r/` (which proxies to GitHub Releases) and drops them in `/usr/local/bin` (or `~/.local/bin` if not writable) on Unix/WSL, or `%LOCALAPPDATA%\cem\bin` on Windows. The server's User-Agent detection picks the right script automatically: PowerShell → `.ps1`, curl/wget → `.sh`.

### Update / Uninstall

```sh
cem update      # Pulls the latest release (prints "current vs latest" first)
cem uninstall   # Removes the three binaries + the config directory
```

---

## Quick Start

```sh
cem "what is fibonacci?"           # thinker only
cem -w "write fibonacci.py"        # writer only
cem -p "write fibonacci.py"        # pair: thinker → writer
```

The first run opens a setup wizard that asks **which AI is the thinker, which is the writer, and which model each one uses** (Claude opus/sonnet/haiku, Codex gpt-5.5/5-mini, Antigravity gemini-3-pro/flash, Cursor claude-4.6/gpt-5.2 — or a custom string). The model is saved to `~/.cem/config.yaml` under `tools.<key>.model` and appended to every call as `--model <model>`. Change roles later with:

```sh
cem roles claude agy               # global: thinker=claude, writer=agy
cem roles --here claude codex      # this directory only (.cem.yaml)
cem init                           # project-specific wizard
```

`pair` mode is smart: if **thinker == writer**, the second call is skipped (no duplicate output); if neither the prompt nor the thinker's analysis hints at code, the writer is skipped (no wasted LLM call).

---

## Supported AI CLIs

| Key | Tool | Install source | Non-interactive call |
|---|---|---|---|
| `claude` | **Claude Code** (Anthropic) | [native installer](https://code.claude.com/docs/en/quickstart) — auto-updates | `claude -p` (stdin) |
| `agy` | **Antigravity** (Google) | [native installer](https://antigravity.google/docs/cli-getting-started) | `agy -p "prompt"` |
| `gpt` | **Codex** (OpenAI) | `npm i -g @openai/codex` | `codex exec "prompt"` |
| `cursor` | **Cursor agent** | [native installer](https://cursor.com/cli) | `cursor-agent -p "prompt"` |

```sh
cemi                               # available & installed
cemi claude                        # install one (handles missing prerequisites: npm/Node)
cemi all                           # install all four
cemi all -y                        # all four, no prompts
cemir agy                          # uninstall one (shell-installed binaries are also removed)
cemir all -y                       # uninstall everything, no prompts
```

If `cemi <tool>` needs **npm/Node** and it is missing (or too old — npm 3 etc.), cem offers to auto-install Node LTS via `winget` (Windows), `brew` (macOS), or `nvm` (Linux — picked over NodeSource because NodeSource now requires glibc 2.28+, which kills Ubuntu 18.04).

---

## Pair Mode

```sh
cem -p "write a binary search in TypeScript"
```

1. 🧠 **Thinker** (e.g. `claude`) reasons about the problem — algorithm choice, edge cases, type signatures.
2. ✍️ **Writer** (e.g. `agy`) receives both the original prompt and the thinker's analysis, then writes the code.

Skip rules:

| Situation | Behaviour |
|---|---|
| `thinker == writer` (both `claude`) | Writer is skipped (no duplication) |
| Prompt has no code intent and thinker output has no \`\`\` block | Writer is skipped |
| Otherwise | Writer runs with the thinker's analysis as context |

---

## API Key Management & Auto-Rotation

For long projects that you don't want interrupted by a single key's rate limit, store **multiple keys per provider** and cem will rotate automatically.

```sh
cem keys add anthropic             # interactive: key + optional label
cem keys add openai
cem keys list                      # masked view
cem keys remove anthropic 1        # remove the first anthropic key
```

Supported providers:

| Provider | Env var | Tool |
|---|---|---|
| `anthropic` | `ANTHROPIC_API_KEY` | Claude (`claude`) |
| `openai` | `OPENAI_API_KEY` | Codex (`gpt`/`codex`) |

> `agy` (Antigravity) and `cursor` use Google/Cursor OAuth; the CLIs do not publish an official API-key env yet, so they are not in the rotation set.

Rotation trigger: if stderr contains `rate limit`, `429`, `quota`, `too many requests`, or `overloaded`, cem advances to the next key. If all keys are exhausted, the last error is returned.

### Post-install auth prompt

After a successful `cemi <tool>` install (when the tool supports API keys), cem asks how you want to authenticate:

```
  Claude için auth:
    [1] API key kaydet (çoklu key + rate-limit rotasyonu)
    [2] Subscription / OAuth login (sonra: 'claude' çalıştır)
    [3] Şimdilik atla
  Seçim [1-3]:
```

Picking **1** records your key (and lets you add more in a loop). Picking **2** prints the binary you should run to start the OAuth/subscription flow. With `cemi -y` the prompt is skipped silently — a hint reminds you of `cem keys add <provider>` and the binary name.

### Safety: .gitignore safeguard

After every successful install, if the current directory is a git repo cem appends `.cem.yaml` to `.gitignore` (and tells you it did). If there is no `.gitignore` at all, a warning is printed. This prevents accidentally committing a project-local cem config.

---

## Diagnostics

```sh
cem doctor                         # system + roles + tools + PATH report
cem status                         # short summary
cem roles                          # active roles
cem history                        # recent commands (TSV)
cem history -n 50                  # last 50
cem history --clear
```

`~/.cem/history.log` — tab-separated log of every cem invocation.

---

## Config

`~/.cem/config.yaml` (global) and `.cem.yaml` at a project root (override):

```yaml
roles:
  thinker: claude
  writer:  agy

tools:
  claude:
    command: claude
    version: 2.1.143
    model: opus              # passed as `claude --model opus`; empty → CLI default
  agy:
    # When a native installer doesn't refresh PATH for the current process,
    # cem records the absolute path it discovered post-install:
    command: C:\Users\Muslu\AppData\Local\agy\bin\agy.exe
    version: 1.2.0
    model: gemini-3-pro

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
make build                         # 3 binaries → build/
make install                       # /usr/local/bin (sudo)
go test ./...
```

The version string is injected from `git describe --tags --always --dirty` via LDFLAGS.

---

## License

MIT — see [LICENSE](LICENSE).
