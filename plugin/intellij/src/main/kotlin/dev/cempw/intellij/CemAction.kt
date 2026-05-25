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
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.wm.ToolWindowManager
import java.io.BufferedReader
import java.io.InputStreamReader

/**
 * cem editor actions (think / write / pair).
 *
 * Selection geçerliyse onu gönderir; seçim yok + dosya açık ve dolu ise tüm
 * dosyayı gönderir; ikisi de boşsa kullanıcıdan input alır.
 */
sealed class CemAction(val mode: Mode) : AnAction() {

    enum class Mode(val flag: String?) {
        THINK(null),
        WRITE("-w"),
        PAIR("-p"),
    }

    class Think : CemAction(Mode.THINK)
    class Write : CemAction(Mode.WRITE)
    class Pair : CemAction(Mode.PAIR)

    override fun actionPerformed(event: AnActionEvent) {
        val project = event.project ?: return
        val editor: Editor? = event.getData(CommonDataKeys.EDITOR)
            ?: FileEditorManager.getInstance(project).selectedTextEditor
        val text = pickPrompt(project, editor) ?: return
        launchCem(project, mode, text)
    }

    private fun pickPrompt(project: Project, editor: Editor?): String? {
        val sel = editor?.selectionModel?.selectedText?.takeIf { it.isNotBlank() }
        if (sel != null) return sel
        val all = editor?.document?.text?.takeIf { it.isNotBlank() }
        if (all != null) return all
        // Editor yok ya da dosya boş → kullanıcı prompt yazsın.
        return promptUser(project, "What do you want to ask cem?")
    }

    companion object {
        internal val LOG = Logger.getInstance(CemAction::class.java)

        /** Çok-satırlı input dialog. Kullanıcı iptal ederse null. */
        fun promptUser(project: Project, message: String, default: String = ""): String? {
            val resp = Messages.showMultilineInputDialog(
                project, message, "cem", default, Messages.getQuestionIcon(), null,
            )
            return resp?.takeIf { it.isNotBlank() }
        }

        /** Background thread'de cem'i çalıştır + sonucu yeni tab'a stream et. */
        fun launchCem(project: Project, mode: Mode, prompt: String) {
            val toolWindow = ToolWindowManager.getInstance(project).getToolWindow("cem")
                ?: return
            val tab = CemTab.newRun(toolWindow, mode.name.lowercase(), prompt.take(80))
            ApplicationManager.getApplication().executeOnPooledThread {
                try {
                    runCem(project, mode, prompt, tab)
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
            val pb = ProcessBuilder(cmd).redirectErrorStream(true)
            if (workDir != null) pb.directory(java.io.File(workDir))
            val process = pb.start()
            BufferedReader(InputStreamReader(process.inputStream)).use { reader ->
                reader.lineSequence().forEach { line ->
                    ApplicationManager.getApplication().invokeLater { tab.appendLine(line) }
                }
            }
            val exit = process.waitFor()
            ApplicationManager.getApplication().invokeLater {
                if (exit != 0) tab.appendError("cem exited with code $exit")
                else tab.appendDim("─── done ───")
            }
        }
    }
}

/**
 * cem "ask freely" — editor state'inden bağımsız, her zaman input dialog
 * açıp serbest prompt alır. Hiçbir dosya açık değilken bile çalışır.
 */
class CemAskAction : AnAction() {
    override fun actionPerformed(event: AnActionEvent) {
        val project = event.project ?: return
        val prompt = CemAction.promptUser(project, "What do you want to ask cem?") ?: return
        CemAction.launchCem(project, CemAction.Mode.PAIR, prompt)
    }
}

/**
 * Project tree'de dosyaya sağ-tık ile çağrılan file-level aksiyonlar.
 *
 * Her aksiyon dosya içeriğini okuyup önceden tanımlı bir talimat prefix'iyle
 * cem -p'ye yollar (pair modu, çünkü hem analiz hem üretim faydalı).
 */
sealed class CemFileAction(private val instruction: String) : AnAction() {

    class Review : CemFileAction(
        "Aşağıdaki dosyayı kod kalitesi, olası bug'lar, eksikler ve iyileştirme " +
                "önerileri için incele. Kısa ve eyleme dönüştürülebilir maddelerle yaz.",
    )
    class FixErrors : CemFileAction(
        "Aşağıdaki dosyada hata, bug veya yanlış kullanım varsa tespit et ve " +
                "düzeltilmiş kodu üret. Sadece değişen kısmı belirt, açıklamayı kısa tut.",
    )
    class Explain : CemFileAction(
        "Aşağıdaki dosyanın ne yaptığını, ana akışını ve önemli detaylarını sadece " +
                "düz prozayla açıkla. Kod tekrar etme.",
    )
    class Ask : CemFileAction("__ASK__") // kullanıcı talimatı interaktif girer

    override fun actionPerformed(event: AnActionEvent) {
        val project = event.project ?: return
        val files = event.getData(CommonDataKeys.VIRTUAL_FILE_ARRAY)
            ?: event.getData(CommonDataKeys.VIRTUAL_FILE)?.let { arrayOf(it) }
            ?: run {
                Messages.showWarningDialog(project, "Hiç dosya seçilmedi.", "cem")
                return
            }
        // Çoklu seçim → tek prompt'ta birleştir (dosya başına başlık ile).
        val sb = StringBuilder()
        for (f in files) {
            if (f.isDirectory) continue
            val text = readFile(f) ?: continue
            sb.append("===== ").append(f.path).append(" =====\n")
            sb.append(text).append("\n\n")
        }
        if (sb.isEmpty()) {
            Messages.showWarningDialog(project, "Seçili dosyaların hiçbiri okunamadı.", "cem")
            return
        }
        val actualInstruction = if (instruction == "__ASK__") {
            CemAction.promptUser(
                project,
                "Bu dosya(lar) için cem'e ne sormak istiyorsun?",
            ) ?: return
        } else {
            instruction
        }
        val prompt = "$actualInstruction\n\n${sb.toString().trim()}"
        CemAction.launchCem(project, CemAction.Mode.PAIR, prompt)
    }

    private fun readFile(file: VirtualFile): String? = try {
        file.inputStream.bufferedReader(Charsets.UTF_8).use { it.readText() }
    } catch (_: Exception) {
        null
    }
}
