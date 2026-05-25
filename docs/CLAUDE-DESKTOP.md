# cem in Claude Desktop (MCP)

Use cem as a tool inside Claude Desktop conversations — `pair` mode results inline in chat, etc.

Turkish: [CLAUDE-DESKTOP.tr.md](CLAUDE-DESKTOP.tr.md)

---

## How it works

Claude Desktop supports [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) servers. We ship `cem-mcp` — a small Go binary that exposes cem's three modes (`think`, `write`, `pair`) as MCP tools. Claude can then invoke cem inside its own reasoning chain.

Why useful: Claude gets a second-opinion from a different model (whatever cem's writer is — Codex, Antigravity, etc.) without you switching context.

## Install

1. Make sure `cem` is on your PATH ([README](../README.md)).
2. Download `cem-mcp-<os>-<arch>` from the latest release and put it somewhere stable:
   - macOS / Linux: `~/.local/bin/cem-mcp` then `chmod +x ~/.local/bin/cem-mcp`
   - Windows: `C:\Users\<you>\AppData\Local\cem\bin\cem-mcp.exe`

   ```
   https://github.com/muslu/cem/releases/latest/download/cem-mcp-<os>-<arch>
   ```

3. Edit Claude Desktop's MCP config:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

   ```json
   {
     "mcpServers": {
       "cem": {
         "command": "/Users/<you>/.local/bin/cem-mcp",
         "env": {
           "CEM_BIN": "cem"
         }
       }
     }
   }
   ```

   (Use absolute paths; Claude Desktop won't resolve PATH itself.)

4. Restart Claude Desktop.

## Usage

In any conversation, mention the tool by name:

> Use the cem `pair` tool to write a Rust SSE server for me.

Claude will call `cem-mcp` → `cem -p "..."` runs → Claude receives the output → integrates into its response.

The three tools:

| Tool | Equivalent shell command |
|---|---|
| `think(prompt)` | `cem "<prompt>"` |
| `write(prompt)` | `cem -w "<prompt>"` |
| `pair(prompt)`  | `cem -p "<prompt>"` |

## Troubleshooting

### Tool not visible in Claude Desktop

1. Did you fully quit and relaunch Claude Desktop? Hot-reload of `claude_desktop_config.json` is hit-and-miss.
2. Check Claude Desktop's MCP logs (search for "MCP" in its log directory).
3. Confirm the `command` path resolves: open Terminal and run it directly. `cem-mcp` should sit silently waiting for JSON-RPC on stdin.

### "cem-mcp: command not found" or "cem: command not found"

Both binaries must exist and be executable. cem-mcp runs `cem` as a subprocess, so cem must be on PATH **as seen by cem-mcp** (which Claude Desktop spawns). On macOS, GUI apps don't get Homebrew's PATH automatically — use absolute paths or set `CEM_BIN` to the full path in the MCP config.

### Tool returns errors

Run cem from your terminal with the same prompt:
```sh
cem -p "your prompt"
```
If that fails, the issue is upstream (auth, missing AI tool, etc.). Fix that first and the MCP integration will follow.

## Same MCP server works in

- **Cursor** — `~/.cursor/mcp.json` (same schema)
- **Continue** — `.continue/config.json` `mcp` section
- **Antigravity** — pending (track [docs/ANTIGRAVITY.md](ANTIGRAVITY.md))
- Any other MCP-compatible host
