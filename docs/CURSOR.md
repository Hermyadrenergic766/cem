# cem in Cursor

Cursor is a VS Code fork; cem integrates two ways at once.

---

## Path A — VS Code extension (commands + shortcuts)

The same `.vsix` from VS Code works in Cursor:

```sh
cursor --install-extension cem-vscode-<version>.vsix
```

Reload Cursor (Command Palette → `Developer: Reload Window`). Commands appear:

| Command | Shortcut | Effect |
|---|---|---|
| `cem: think on selection` | `Ctrl+Alt+I` | Thinker |
| `cem: write on selection` | `Ctrl+Alt+W` | Writer |
| `cem: pair on selection`  | `Ctrl+Alt+P` | Thinker → writer |

Full details: [VSCODE.md](VSCODE.md).

---

## Path B — MCP server (Cursor's own agent calls cem)

Lets Cursor's built-in agent invoke cem as a tool. Useful when you want a **second model opinion** inside a Cursor agent conversation.

1. Download `cem-mcp-<os>-<arch>` from the latest release.
2. Place it on disk and make executable.
3. Add to `~/.cursor/mcp.json`:
   ```json
   {
     "mcpServers": {
       "cem": {
         "command": "/absolute/path/to/cem-mcp",
         "env": { "CEM_BIN": "cem" }
       }
     }
   }
   ```
4. Restart Cursor.

In a Cursor agent chat:
> Use the cem `pair` tool to write a Tornado SSE server.

Cursor's agent calls cem-mcp → cem spawns the configured thinker → writer chain → output flows back.

See [CLAUDE-DESKTOP.md](CLAUDE-DESKTOP.md) for the MCP protocol details (identical schema).

---

## Both paths together

You can install **both** — the VS Code extension for keyboard-driven invocation AND the MCP server for agent delegation. They don't conflict; they target different workflows.
