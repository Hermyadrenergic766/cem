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
> Drive Claude · Antigravity · Codex · Cursor from a single CLI.

- Site: [cem.pw](https://cem.pw)
- Türkçe sürüm: [README.tr.md](README.tr.md)
- Power-user docs: [ADVANCED.md](ADVANCED.md) / [ADVANCED.tr.md](ADVANCED.tr.md)

---

## 1. Install

**macOS / Linux / WSL:**
```sh
curl -fsSL cem.pw/install | sh
```

**Windows (PowerShell — not CMD, not Git Bash):**
```powershell
irm cem.pw/install | iex
```

That's it. You now have three commands on your PATH: `cem`, `cemi`, `cemir`.

---

## 2. First run

```sh
cem "what is fibonacci?"
```

A short wizard opens the first time. Pick one AI to **think** and one to **write**. Pick a model for each (or hit Enter for the AI's default). Done.

If a tool isn't installed yet, cem offers to install it for you. If Node.js or another prerequisite is missing, cem offers to install that too — via `winget` (Windows), `brew` (macOS), or `nvm` (Linux).

---

## 3. Three commands you'll actually use

```sh
cem "explain this codebase"         # ask the THINKER (one AI)
cem -w "write a quicksort in Go"    # ask the WRITER (one AI)
cem -p "build me a CLI tool"        # PAIR: thinker plans, writer codes
```

That's the whole product. The pair mode is the interesting one — the thinker analyses the task and hands its plan to the writer, so the code you get is informed by reasoning from a different model.

---

## 4. Login

After installing an AI tool, cem asks how you want to authenticate:

```
  Claude için auth:
    [1] API key kaydet  (multiple keys + auto-rotation on rate limit)
    [2] Subscription / OAuth login  (run 'claude' to start the browser flow)
    [3] Skip
```

Pick **1** if you have an API key (Anthropic Console / OpenAI Platform). Pick **2** if you have a subscription (Claude Pro, ChatGPT Plus, Antigravity, Cursor).

Add or check keys any time:
```sh
cem keys add anthropic     # paste your sk-ant-... key
cem keys list              # masked view
cem keys remove openai 2   # delete the 2nd OpenAI key
```

Multiple keys are tried in order; when one hits a rate limit cem rotates to the next automatically. That's how a 3-hour project doesn't get interrupted.

---

## 5. Update / Uninstall

```sh
cem update         # fetch the latest release
cem uninstall      # remove cem
cemir all          # remove the installed AI tools
```

cem checks for new versions in the background and shows a one-line notice if there's an update — never blocks you.

> **Version format:** cem uses calendar versioning `YYYYMMDD.MINOR` since 2026-05-25 (example: `20260525.05`). Older `v0.1.x` semver tags keep working — `cem update` understands both formats and only suggests an update when the remote tag is genuinely newer.

---

## IDE integrations

Run cem from inside your editor instead of the terminal:

| Editor | How to install | Shortcuts |
|---|---|---|
| **PyCharm / IntelliJ IDEA / GoLand / WebStorm / etc.** | Download `cem-intellij-<version>.zip` from [Releases](https://github.com/muslu/cem/releases/latest) → Settings → Plugins → ⚙ → Install Plugin from Disk | `Ctrl+Alt+I/W/P` |
| **VS Code** | Download `cem-vscode-<version>.vsix` → `code --install-extension cem-vscode-<version>.vsix` | `Ctrl+Alt+I/W/P` |
| **Claude Desktop / Cursor / Continue** (MCP) | Download `cem-mcp-<os>-<arch>` → register in your host's MCP config | — (tool calls from chat) |
| **Antigravity IDE** | Use the built-in terminal for now; MCP support is on the roadmap | — |

See [docs/IDE.md](docs/IDE.md) for screen-by-screen setup and [docs/ANTIGRAVITY.md](docs/ANTIGRAVITY.md) for Antigravity-specific notes.

---

## Want more?

That covers maybe 80% of daily use. For everything else — project-specific configs, OAuth code helpers, custom models, deeper pair-mode rules, build-from-source — see [ADVANCED.md](ADVANCED.md).

---

## License

MIT — see [LICENSE](LICENSE).
