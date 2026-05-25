package dev.cempw.intellij

import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
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
        val editor = event.getData(CommonDataKeys.EDITOR)
        val text = editor?.selectionModel?.selectedText
            ?: editor?.document?.text
            ?: run {
                Messages.showWarningDialog(project, "No selection or open editor.", "cem")
                return
            }

        val window = CemToolWindowFactory.show(project)
        window.appendHeader(mode.name.lowercase(), text.take(80))

        // Run cem in a background thread to avoid blocking EDT
        ApplicationManager.getApplication().executeOnPooledThread {
            try {
                runCem(project, mode, text, window)
            } catch (e: Exception) {
                LOG.warn("cem invocation failed", e)
                ApplicationManager.getApplication().invokeLater {
                    window.appendError("Failed to invoke cem: ${e.message}")
                }
            }
        }
    }

    private fun runCem(project: Project, mode: Mode, prompt: String, window: CemToolWindow) {
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
                    window.appendLine(line)
                }
            }
        }
        val exitCode = process.waitFor()
        ApplicationManager.getApplication().invokeLater {
            if (exitCode != 0) {
                window.appendError("cem exited with code $exitCode")
            } else {
                window.appendDim("─── done ───")
            }
        }
    }

    companion object {
        private val LOG = Logger.getInstance(CemAction::class.java)
    }
}
