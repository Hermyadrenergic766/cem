# cem-mcp — MCP server for cem

Exposes [cem](https://github.com/muslu/cem) as an MCP (Model Context Protocol) server. Any MCP-compatible host — **Claude Desktop**, **Cursor**, **Continue**, and (when supported) **Antigravity IDE** — can use cem as a tool inside its own conversations.

## Tools

| Tool | Effect |
|---|---|
| `think(prompt)` | Runs `cem "<prompt>"` — single AI thinker |
| `write(prompt)` | Runs `cem -w "<prompt>"` — single AI writer |
| `pair(prompt)`  | Runs `cem -p "<prompt>"` — thinker → writer chain |

## Install

Download `cem-mcp-<os>-<arch>` from [GitHub Releases](https://github.com/muslu/cem/releases) and place it on PATH (or anywhere referenced by your MCP host config).

## Register with Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "cem": {
      "command": "/absolute/path/to/cem-mcp",
      "env": {
        "CEM_BIN": "cem"
      }
    }
  }
}
```

Restart Claude Desktop. The three tools (`think`, `write`, `pair`) now appear in the tool list. Ask Claude:

> Use the cem `pair` tool to write a Rust SSE example.

Claude will invoke `cem-mcp`, which in turn spawns `cem -p "..."`, which pipes the prompt to your thinker → writer chain. Claude gets the output and integrates it into its response.

## Register with Cursor / Continue / others

Most MCP-compatible hosts use the same `command` + `env` schema. Consult the host's docs for the file location:

- Cursor: `~/.cursor/mcp.json`
- Continue: `.continue/config.json`
- Antigravity: not yet supported (research in progress, see [`docs/ANTIGRAVITY.md`](../../docs/ANTIGRAVITY.md))

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `CEM_BIN` | `cem` | Path to the cem binary the server should spawn. |

## Protocol

JSON-RPC 2.0 over stdio per the [Model Context Protocol spec](https://modelcontextprotocol.io/). Methods implemented: `initialize`, `tools/list`, `tools/call`, `ping`.

## Build from source

```sh
cd plugin/mcp
go build -o cem-mcp .
```

No external dependencies.
