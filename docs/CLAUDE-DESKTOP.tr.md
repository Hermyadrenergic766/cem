# Claude Desktop'ta cem (MCP)

Claude Desktop konuşmalarında cem'i tool olarak kullan — `pair` modu chat içinde sonuç döndürür.

English: [CLAUDE-DESKTOP.md](CLAUDE-DESKTOP.md)

---

## Nasıl çalışır

Claude Desktop [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) sunucularını destekler. `cem-mcp` adlı küçük Go binary'si cem'in 3 modunu (`think`, `write`, `pair`) MCP tool olarak sunar. Claude bunları kendi akıl yürütmesinde çağırabilir.

Faydası: Claude'u terketmeden farklı bir modelden (cem'in writer'ı — Codex, Antigravity vs.) ikinci görüş alırsın.

## Kurulum

1. `cem` PATH'de olmalı ([README](../README.tr.md)).
2. Son sürümden `cem-mcp-<os>-<arch>` indir, kalıcı yere koy:
   - macOS / Linux: `~/.local/bin/cem-mcp` + `chmod +x ~/.local/bin/cem-mcp`
   - Windows: `C:\Users\<sen>\AppData\Local\cem\bin\cem-mcp.exe`

3. Claude Desktop MCP config'ini düzenle:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

   ```json
   {
     "mcpServers": {
       "cem": {
         "command": "/Users/<sen>/.local/bin/cem-mcp",
         "env": { "CEM_BIN": "cem" }
       }
     }
   }
   ```

   Mutlak yollar şart — Claude Desktop PATH'i kendisi çözmez.

4. Claude Desktop'ı tamamen kapat + yeniden aç.

## Kullanım

Bir konuşmada tool'u adıyla iste:

> cem `pair` tool'unu kullan, Rust ile SSE server yaz.

Claude `cem-mcp`'yi çağırır → `cem -p "..."` çalışır → çıktı Claude'a döner → cevabına entegre eder.

3 tool:

| Tool | Eşdeğer shell |
|---|---|
| `think(prompt)` | `cem "<prompt>"` |
| `write(prompt)` | `cem -w "<prompt>"` |
| `pair(prompt)`  | `cem -p "<prompt>"` |

## Sorun giderme

### Tool Claude Desktop'ta görünmüyor

1. Claude Desktop tamamen kapatıp yeniden açtın mı? `claude_desktop_config.json` hot-reload kararsız.
2. Claude Desktop MCP log'larına bak.
3. `command` yolunun doğru olduğunu doğrula: terminalde direkt çalıştır. `cem-mcp` stdin'den JSON-RPC bekleyerek sessiz duracak.

### "cem-mcp: command not found" veya "cem: command not found"

İki binary de mevcut + executable olmalı. cem-mcp `cem`'i subprocess olarak çağırır → cem, cem-mcp'nin gördüğü PATH'te olmalı (Claude Desktop'ın spawn ettiği process). macOS'ta GUI uygulamaları Homebrew PATH'ini otomatik almaz — mutlak yol veya MCP config'inde `CEM_BIN`'i full path olarak ver.

### Tool hata dönüyor

Terminalde aynı prompt'u dene:
```sh
cem -p "prompt'un"
```
Burada başarısızsa sorun upstream'de (auth, eksik AI tool vs). Önce orayı çöz.

## Aynı MCP sunucusu şurada da çalışır

- **Cursor** — `~/.cursor/mcp.json` (aynı şema)
- **Continue** — `.continue/config.json` `mcp` bölümü
- **Antigravity** — yol haritasında ([docs/ANTIGRAVITY.md](ANTIGRAVITY.md))
- Diğer her MCP host'u
