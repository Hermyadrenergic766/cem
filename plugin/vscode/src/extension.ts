/**
 * cem — VS Code extension entry point.
 *
 * Registers three commands (think / write / pair) that pipe the editor
 * selection to the `cem` CLI and stream the response into a dedicated
 * output channel.
 *
 * Plus: Settings → cem thinker/writer/model değişikliklerini ~/.cem/config.yaml'a
 * senkronlar (CLI ve plugin aynı config'i paylaşır).
 */
import * as vscode from "vscode";
import { spawn } from "node:child_process";
import { saveToCemYaml } from "./cemConfig";

type Mode = "think" | "write" | "pair";

const FLAGS: Record<Mode, string[]> = {
  think: [],
  write: ["-w"],
  pair:  ["-p"],
};

let channel: vscode.OutputChannel | undefined;

function output(): vscode.OutputChannel {
  if (!channel) {
    channel = vscode.window.createOutputChannel("cem");
  }
  return channel;
}

function getPrompt(): { text: string; from: string } | undefined {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    vscode.window.showWarningMessage("cem: no active editor.");
    return undefined;
  }
  const sel = editor.selection;
  if (!sel.isEmpty) {
    return { text: editor.document.getText(sel), from: "selection" };
  }
  return { text: editor.document.getText(), from: "file" };
}

function cemPath(): string {
  return vscode.workspace.getConfiguration("cem").get<string>("path", "cem");
}

async function runMode(mode: Mode) {
  const prompt = getPrompt();
  if (!prompt) return;

  const out = output();
  out.show(true);
  out.appendLine("");
  out.appendLine(`─── cem ${mode} · ${prompt.from} (${prompt.text.length} chars) ───`);

  const folder = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  const args = [...FLAGS[mode], prompt.text];

  await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: `cem ${mode}…`, cancellable: true },
    (_progress, token) => new Promise<void>((resolve) => {
      const proc = spawn(cemPath(), args, { cwd: folder ?? process.cwd() });
      token.onCancellationRequested(() => proc.kill());
      proc.stdout.on("data", (b) => out.append(b.toString()));
      proc.stderr.on("data", (b) => out.append(b.toString()));
      proc.on("close", (code) => {
        out.appendLine("");
        out.appendLine(code === 0 ? "─── done ───" : `─── exit ${code} ───`);
        resolve();
      });
      proc.on("error", (err) => {
        out.appendLine(`ERROR: ${err.message}`);
        vscode.window.showErrorMessage(
          `cem: failed to spawn '${cemPath()}'. Set cem.path in settings.`,
        );
        resolve();
      });
    }),
  );
}

export function activate(context: vscode.ExtensionContext) {
  for (const mode of ["think", "write", "pair"] as Mode[]) {
    context.subscriptions.push(
      vscode.commands.registerCommand(`cem.${mode}`, () => runMode(mode)),
    );
  }
  // Settings → cem değişimleri ~/.cem/config.yaml'a yansısın.
  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration((e) => {
      if (!e.affectsConfiguration("cem")) return;
      const cfg = vscode.workspace.getConfiguration("cem");
      const thinker = cfg.get<string>("thinker", "").trim();
      const writer  = cfg.get<string>("writer", "").trim();
      const thinkerModel = cfg.get<string>("thinkerModel", "").trim();
      const writerModel  = cfg.get<string>("writerModel", "").trim();
      // Hiçbiri set edilmemişse senkron atla (path-only değişimde YAML'ı bozma).
      if (!thinker && !writer && !thinkerModel && !writerModel) return;
      try {
        saveToCemYaml({ thinker, writer, thinkerModel, writerModel });
        vscode.window.setStatusBarMessage("$(check) cem settings → ~/.cem/config.yaml", 3000);
      } catch (err: unknown) {
        vscode.window.showWarningMessage(
          `cem: ~/.cem/config.yaml yazılamadı — ${(err as Error).message}`,
        );
      }
    }),
  );
}

export function deactivate() {
  channel?.dispose();
}
