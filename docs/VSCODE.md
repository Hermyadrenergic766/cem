# cem in VS Code

Turkish: [VSCODE.tr.md](VSCODE.tr.md)

---

## Install

1. Make sure `cem` is on your PATH ([README](../README.md)).
2. Download the latest `.vsix` from:
   ```
   https://github.com/muslu/cem/releases/latest/download/cem-vscode-<version>.vsix
   ```
3. Install:
   ```sh
   code --install-extension cem-vscode-<version>.vsix
   ```
   Or in VS Code: **Extensions → ⋯ → Install from VSIX**.
4. Reload VS Code (Command Palette → `Developer: Reload Window`).

You'll see three commands in the palette and a **cem** submenu in the editor right-click menu.

## Usage

| Command | Default shortcut | Effect |
|---|---|---|
| `cem: think on selection` | `Ctrl+Alt+I` (`⌥⌘I` macOS) | Thinker only |
| `cem: write on selection` | `Ctrl+Alt+W` (`⌥⌘W`) | Writer only |
| `cem: pair on selection`  | `Ctrl+Alt+P` (`⌥⌘P`) | Thinker → writer |

If no text is selected, the entire active file is sent.

Output streams to the **Output** panel (drop-down: `cem`). Cancel via the progress notification on the bottom-right.

## Settings

```jsonc
// settings.json
"cem.path": "cem"  // or absolute path if not on PATH
```

**Settings → Extensions → cem** for the GUI.

## Cursor compatibility

Cursor is a VS Code fork — the same `.vsix` installs cleanly:
```sh
cursor --install-extension cem-vscode-<version>.vsix
```

(Note: this lets you use cem from within Cursor as a *complement* to Cursor's built-in agent — different model second opinions, etc.)

## Troubleshooting

### Commands greyed out / "no active editor"

Make sure a file is open and focused in the editor area.

### "spawn cem ENOENT"

cem isn't on PATH or `cem.path` setting points somewhere wrong. Test in VS Code terminal:
```sh
cem --version
```
If that works but the extension errors, your VS Code is using a different shell env. Set `cem.path` to the absolute path (find with `which cem` / `where cem`).

### Extension installs but does nothing

VS Code shouldn't normally suppress activation, but check **Output → Log (Extension Host)** for cem activation errors.

## Build from source

```sh
cd plugin/vscode
npm install
npm run compile
npx vsce package --no-dependencies
code --install-extension cem-0.1.0.vsix
```

Requires Node.js 20+.
