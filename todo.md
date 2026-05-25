# CEM — TODO

> Turkish version: [todo.tr.md](todo.tr.md)

## 1. Folder & File Layout
- [x] Move `uninstall/cmd_uninstall.go` → root (`cmd_uninstall.go`)
- [x] Move `uninstall/uninstall.sh` → root
- [x] Move `uninstall/uninstall.ps1` → root
- [x] Delete the now-empty `uninstall/` directory
- [x] Reorganize `nginx/` (snippets/, sites-available/, fail2ban/)

## 2. Missing Core Go Files
- [x] `main.go` — dispatch by binary name (cem / cemi / cemir)
- [x] `config.go` — GlobalConfig + ProjectConfig + ResolvedConfig + Roles
      + InstalledTool + KnownTools + LoadConfig / saveGlobalConfig
      / loadGlobalConfig / SaveProjectConfig
- [x] `executor.go` — ModeThink/ModeWrite/ModePair + Run + ReadStdin

## 3. Build & Dependencies
- [x] Populate `go.mod` (cobra + lipgloss + yaml.v3)
- [x] `go mod tidy`
- [x] `Makefile` (build / dev / install / clean / tidy / test)
- [x] `go build` smoke test — 3 binaries (3.9 MB each with ldflags)

## 4. CI/CD
- [x] `.github/workflows/release.yml` (7 platforms + SHA256SUMS)
- [x] `.gitlab-ci.yml` mirror (kept in tree; canonical is GitHub)

## 5. Verification
- [x] `./build/cem --help` works
- [x] `./build/cemi --version` → 1.0.0
- [x] `./build/cemir --version` → 1.0.0

## 6. Extras (initial session)
- [x] `.claude/` audit (agents/skill/hooks)
- [x] `CLAUDE.md` rewritten for the project (English) + `CLAUDE.tr.md`
- [x] `install.sh` + `install.ps1` URLs → GitHub canonical

## 7. New features (follow-up session)
- [x] `cem doctor` command (system + roles + tools + PATH)
- [x] `cemir all` bulk uninstall (with confirmation + failure summary)
- [x] LDFLAGS version injection from `git describe`
- [x] `config_test.go` — 8 unit tests including deprecation and order
- [x] `.gitlab-ci.yml` release stage — 21 binary asset links + SHA256SUMS
- [x] `.gitignore` rewritten for CEM
- [x] Git init + GitHub origin

## 8. Persistence & UX (follow-up session)
- [x] `~/.cem/history.log` + `cem history` (-n / --clear)
- [x] `cem -p` spinner — TTY-aware, no bubbletea dep
- [x] nginx `/r/` proxy → GitHub Releases
- [x] CLAUDE.md / README updates

## 9. Brand + tool catalogue (current session)
- [x] Slogan: **CEM — Compose · Execute · Multiplex**
      ("One command, many AIs.")
- [x] Canonical repo flipped to `https://github.com/muslu/cem.git`
- [x] `agy` description corrected → **Antigravity (Google)**
- [x] `gemini` deprecation note (personal use ends 2026-06-16)
- [x] 5 new AI CLIs added: goose, cody, continue, openhands, cursor
- [x] `orderedToolKeys` introduced; 4 duplicated `[]string{...}` lists
      reduced to one source of truth
- [x] All MD docs split English/Turkish (README/CLAUDE/OPERATIONS/todo)

## 10. Open (deferred / decision pending)
- [ ] `.claude/agents/` and `.claude/skills/` — leftovers from the
      `autoinstalltrixie` project. Decision pending: delete or replace
      with CEM-specific agents (e.g., `tool-installer`, `role-switcher`)?
      `.claude/` is gitignored.
- [ ] More test coverage: `executor_test.go`, `history_test.go`,
      OS-integration tests
- [ ] LICENSE file (README references MIT but no LICENSE in the tree)
- [ ] `cem.pw/docs` static site (referenced from install scripts)
- [ ] Confirm exact install commands for the 5 newly-added AI CLIs —
      some `InstallCmd` values are best-guess upstream names
