# cem in Continue

[Continue](https://continue.dev) is an open-source AI assistant for VS Code and JetBrains IDEs. cem integrates as an MCP server.

---

## Install

1. Download `cem-mcp-<os>-<arch>` from the latest release.
2. Make it executable and remember its absolute path.
3. Edit `~/.continue/config.json` (or `%USERPROFILE%\.continue\config.json` on Windows):

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "/absolute/path/to/cem-mcp",
          "env": { "CEM_BIN": "cem" }
        }
      }
    ]
  }
}
```

4. Reload Continue (Command Palette → `Continue: Reload`).

The three tools (`think`, `write`, `pair`) appear in Continue's tool drawer.

---

## Usage

In Continue's chat, ask the agent to use cem:

> Use the cem `pair` tool to implement the function.

Continue will call cem-mcp, which spawns `cem -p "..."` and pipes the result back into Continue's context.

---

Same MCP protocol details as Claude Desktop — see [CLAUDE-DESKTOP.md](CLAUDE-DESKTOP.md).
