# cem — PyCharm / IntelliJ Plugin (B)

Native IntelliJ Platform plugin scaffold for cem. Adds **Tools → cem** menu with three actions (`think` / `write` / `pair`) and a dedicated tool window for streaming output.

Source: [`plugin/intellij/`](../plugin/intellij/)

Türkçe: [PYCHARM-PLUGIN.tr.md](PYCHARM-PLUGIN.tr.md)

---

## Status

**Scaffolded, not yet released.** The plugin is buildable but needs:

- [ ] Streaming output (currently line-by-line, but on `process.inputStream` which may buffer)
- [ ] Run configuration support (`cem` as a debug target)
- [ ] Project-aware working directory (currently uses `project.basePath`)
- [ ] Output as collapsible groups (per invocation)
- [ ] Marketplace publishing (`./gradlew publishPlugin`)

Pre-release builds will be attached to GitHub Releases starting with cem v0.2.0.

---

## Build

Requires:
- JDK 21
- Gradle (any 8.x)

```sh
cd plugin/intellij
./gradlew buildPlugin
```

Output: `build/distributions/cem-<version>.zip`.

Install via **PyCharm → Settings → Plugins → ⚙ → Install Plugin from Disk** and pick that zip.

## Run sandbox IDE

```sh
./gradlew runIde
```

Spawns a sandboxed PyCharm (or whichever IDE the IntelliJ Platform Gradle plugin pulls — see `build.gradle.kts`) with the plugin loaded. Edit code, watch hot reload.

---

## Architecture

```
plugin/intellij/
├── build.gradle.kts                 — Gradle config (IntelliJ Platform Gradle Plugin v2)
├── settings.gradle.kts
├── gradle.properties
└── src/main/
    ├── kotlin/dev/cempw/intellij/
    │   ├── CemAction.kt             — sealed class: Think / Write / Pair extends AnAction
    │   ├── CemToolWindow.kt         — JTextPane + tool window factory
    │   └── CemSettings.kt           — PersistentStateComponent + Configurable panel
    └── resources/META-INF/plugin.xml
```

### Action flow

1. User selects text in editor, presses `Ctrl+Alt+I`/`W`/`P`.
2. `CemAction.actionPerformed` reads `editor.selectionModel.selectedText` (falls back to the full document).
3. Tool window is shown with a header line.
4. Background thread runs `ProcessBuilder(cemPath, [-w|-p], prompt)` with project's basePath.
5. Each line of `process.inputStream` is appended to the tool window via `invokeLater`.

### Settings

`Settings → Tools → cem` exposes one field: **cem binary path** (default `"cem"`, resolved against `PATH`). Persisted in `cem.xml` via IntelliJ's standard `PersistentStateComponent`.

---

## Known constraints

- **Selection size:** `prompt` is passed as a positional argument. On Windows, command line is capped at ~32 KB. Avoid running cem on multi-megabyte selections.
- **Output buffering:** cem's spinner uses `\r` carriage returns; the plugin's line-based reader currently shows them as garbage. We'll switch to a `Reader.read(charBuffer)` polling model when adding streaming support.
- **Cancel:** No cancel button yet. Killing the process requires switching to `OSProcessHandler` (built into IntelliJ Platform) so we get Stop semantics for free.

---

## Roadmap

| Step | Effort | Notes |
|---|---|---|
| Initial scaffold | done | This commit |
| OSProcessHandler + Stop button | ½ day | Streaming + cancel |
| Inline diff insertion | 1 day | "Apply" button on writer output |
| Settings: model picker per project | ½ day | Read/write `.cem.yaml` from project root |
| Marketplace listing | ½ day | Screenshots, description, vendor verification |

Open PRs welcome.
