# VS Code'da cem

English: [VSCODE.md](VSCODE.md)

---

## Kurulum

1. `cem` PATH'de olmalı ([README](../README.tr.md)).
2. Son `.vsix`'i indir:
   ```
   https://github.com/muslu/cem/releases/latest/download/cem-vscode-<sürüm>.vsix
   ```
3. Kur:
   ```sh
   code --install-extension cem-vscode-<sürüm>.vsix
   ```
   Veya VS Code'da: **Extensions → ⋯ → Install from VSIX**.
4. Reload (Command Palette → `Developer: Reload Window`).

Palette'te 3 komut + editör sağ-tık menüsünde **cem** alt-menüsü.

## Kullanım

| Komut | Kısayol | Etkisi |
|---|---|---|
| `cem: think on selection` | `Ctrl+Alt+I` (`⌥⌘I`) | Thinker |
| `cem: write on selection` | `Ctrl+Alt+W` (`⌥⌘W`) | Writer |
| `cem: pair on selection`  | `Ctrl+Alt+P` (`⌥⌘P`) | Thinker → writer |

Seçim yoksa tüm aktif dosya gönderilir. Çıktı **Output** panel'inde (`cem` drop-down'ı). Sağ-alt köşedeki progress notification'dan iptal.

## Ayarlar

```jsonc
// settings.json
"cem.path": "cem"  // PATH'da değilse mutlak yol
```

GUI için: **Settings → Extensions → cem**.

## Cursor uyumluluğu

Cursor bir VS Code fork'u — aynı `.vsix` direkt kurulur:
```sh
cursor --install-extension cem-vscode-<sürüm>.vsix
```

(Cursor'ın yerleşik agent'ına ek olarak cem'i kullanırsın — farklı model ikinci görüş vs.)

## Sorun giderme

### Komutlar greyed out / "no active editor"

Editör alanında bir dosya açık ve focus'lu olmalı.

### "spawn cem ENOENT"

cem PATH'da değil veya `cem.path` yanlış. Terminal'de:
```sh
cem --version
```
Çalışıyorsa ama extension hata veriyorsa, VS Code'un kullandığı shell env farklı. `cem.path`'i mutlak yola ayarla.

## Kaynaktan derleme

```sh
cd plugin/vscode
npm install
npm run compile
npx vsce package --no-dependencies
```

Node.js 20+ gerekli.
