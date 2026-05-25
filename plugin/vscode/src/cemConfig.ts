/**
 * VS Code settings ↔ ~/.cem/config.yaml senkronu.
 *
 * Kullanıcı VS Code Settings → cem altında thinker/writer/model değiştirince
 * bu modül ~/.cem/config.yaml dosyasını günceller. cem CLI'nın da kullandığı
 * tek source-of-truth.
 */
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import * as yaml from "yaml";

export function configPath(): string {
  return path.join(os.homedir(), ".cem", "config.yaml");
}

interface CemYaml {
  roles?: { thinker?: string; writer?: string };
  tools?: Record<string, { command?: string; version?: string; model?: string }>;
  [key: string]: unknown;
}

export function loadCemYaml(): CemYaml {
  const p = configPath();
  if (!fs.existsSync(p)) return {};
  try {
    return (yaml.parse(fs.readFileSync(p, "utf8")) as CemYaml) ?? {};
  } catch {
    return {};
  }
}

export interface UiState {
  thinker: string;
  writer: string;
  thinkerModel: string;
  writerModel: string;
}

/**
 * UI state'inden ~/.cem/config.yaml'a yaz. Tek alanı güncelleyince diğerleri
 * (api_keys, command path'leri vs.) dokunulmaz; sadece roles + tools.<k>.model
 * patch'lenir.
 */
export function saveToCemYaml(state: UiState): void {
  const cfg = loadCemYaml();
  cfg.roles ??= {};
  if (state.thinker) cfg.roles.thinker = state.thinker;
  if (state.writer)  cfg.roles.writer  = state.writer;

  cfg.tools ??= {};
  const applyModel = (toolKey: string, model: string): void => {
    if (!toolKey) return;
    cfg.tools![toolKey] ??= {};
    if (model) cfg.tools![toolKey].model = model;
    else       delete cfg.tools![toolKey].model;
  };
  applyModel(state.thinker, state.thinkerModel);
  applyModel(state.writer, state.writerModel);

  const p = configPath();
  fs.mkdirSync(path.dirname(p), { recursive: true });
  fs.writeFileSync(p, yaml.stringify(cfg), { mode: 0o600 });
}
