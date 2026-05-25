package dev.cempw.intellij

import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import com.intellij.openapi.wm.ToolWindowManager
import java.io.BufferedReader
import java.io.InputStreamReader

/**
 * cem actions — one entry per cem mode (think / write / pair).
 *
 * The selected text (or current line if no selection) is piped to
 * `cem [-w | -p]` and the output is appended to the cem tool window.
 */
sealed class CemAction(private val mode: Mode) : AnAction() {

    enum class Mode(val flag: String?) {
        THINK(null),     // cem "prompt"
        WRITE("-w"),     // cem -w "prompt"
        PAIR("-p"),      // cem -p "prompt"
    }

    class Think : CemAction(Mode.THINK)
    class Write : CemAction(Mode.WRITE)
    class Pair : CemAction(Mode.PAIR)

    override fun actionPerformed(event: AnActionEvent) {
        val project = event.project ?: return
        // Önce action context'inden editor — editor focus'tayken bu doludur.
        // Tool window'dan kısayol bastıysa context boş; FileEditorManager'dan
        // son seçili editörü çek (kullanıcı tool window'a tıklamış olsa bile
        // arka plandaki kod editörü kalır).
        val editor: Editor? = event.getData(CommonDataKeys.EDITOR)
            ?: FileEditorManager.getInstance(project).selectedTextEditor
        if (editor == null) {
            Messages.showWarningDialog(
                project,
                "Açık bir editör yok. Önce bir dosya aç ve içeriğine tıkla, sonra kısayolu kullan.",
                "cem",
            )
            return
        }
        val text = editor.selectionModel.selectedText?.takeIf { it.isNotBlank() }
            ?: editor.document.text.takeIf { it.isNotBlank() }
            ?: run {
                Messages.showWarningDialog(project, "Dosya boş veya seçim yok.", "cem")
                return
            }

        // Her invocation YENİ tab açar; eski sekmeler kalır, kullanıcı
        // istediğinde kapatır (X tuşu).
        val toolWindow = ToolWindowManager.getInstance(project).getToolWindow("cem")
            ?: error("cem tool window not registered")
        val tab = CemTab.newRun(toolWindow, mode.name.lowercase(), text.take(80))

        // Run cem in a background thread to avoid blocking EDT
        ApplicationManager.getApplication().executeOnPooledThread {
            try {
                runCem(project, mode, text, tab)
            } catch (e: Exception) {
                LOG.warn("cem invocation failed", e)
                ApplicationManager.getApplication().invokeLater {
                    tab.appendError("Failed to invoke cem: ${e.message}")
                }
            }
        }
    }

    private fun runCem(project: Project, mode: Mode, prompt: String, tab: CemTab) {
        val cemPath = CemSettings.instance.cemPath.ifBlank { "cem" }
        val workDir = project.basePath

        val cmd = mutableListOf(cemPath)
        mode.flag?.let { cmd.add(it) }
        cmd.add(prompt)

        val pb = ProcessBuilder(cmd)
            .redirectErrorStream(true)
        if (workDir != null) pb.directory(java.io.File(workDir))

        val process = pb.start()
        BufferedReader(InputStreamReader(process.inputStream)).use { reader ->
            reader.lineSequence().forEach { line ->
                ApplicationManager.getApplication().invokeLater {
                    tab.appendLine(line)
                }
            }
        }
        val exitCode = process.waitFor()
        ApplicationManager.getApplication().invokeLater {
            if (exitCode != 0) {
                tab.appendError("cem exited with code $exitCode")
            } else {
                tab.appendDim("─── done ───")
            }
        }
    }

    companion object {
        private val LOG = Logger.getInstance(CemAction::class.java)
    }
}
