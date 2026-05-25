# cem — VS Code Extension

Run [cem](https://github.com/muslu/cem) — the multi-AI CLI orchestrator — from inside VS Code.

## Commands

| Command | Default shortcut | Effect |
|---|---|---|
| `cem: think on selection` | `Ctrl+Alt+I` (`⌥⌘I` on macOS) | Runs `cem "selection"` (thinker only) |
| `cem: write on selection` | `Ctrl+Alt+W` (`⌥⌘W`) | Runs `cem -w "selection"` (writer only) |
| `cem: pair on selection`  | `Ctrl+Alt+P` (`⌥⌘P`) | Runs `cem -p "selection"` (thinker → writer) |

If no text is selected, the entire active file is sent.

Output streams into the **Output** panel (drop-down: `cem`). Cancel via the progress notification.

## Configuration

| Setting | Default | Description |
|---|---|---|
| `cem.path` | `cem` | Path to the cem binary; absolute path if it's not on PATH. |

## Install

### From release (recommended)

Download `cem-vscode-<version>.vsix` from [GitHub Releases](https://github.com/muslu/cem/releases) and:

```sh
code --install-extension cem-vscode-<version>.vsix
```

### From source

```sh
cd plugin/vscode
npm install
npm run package
code --install-extension cem-0.1.0.vsix
```

## Requirements

`cem` must be installed and on PATH (or its location set via the `cem.path` setting). See the [main project](https://github.com/muslu/cem) for installation.
