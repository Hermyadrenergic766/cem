package dev.cempw.intellij

import com.intellij.openapi.Disposable
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.JBColor
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.components.JBTextField
import com.intellij.ui.content.Content
import com.intellij.ui.content.ContentFactory
import java.awt.BorderLayout
import java.awt.event.ActionEvent
import javax.swing.AbstractAction
import javax.swing.JLabel
import javax.swing.JPanel
import javax.swing.JTextPane
import javax.swing.KeyStroke
import javax.swing.text.SimpleAttributeSet
import javax.swing.text.StyleConstants

/**
 * cem output tool window. Each invocation (think/write/pair) opens a NEW
 * closable tab. A persistent "Welcome" tab gives usage tips on first open.
 */
class CemToolWindowFactory : ToolWindowFactory {
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        // İlk açılış: terminal benzeri etkileşimli sekme. Alttaki input'a yazıp
        // Enter'a basınca cem (think mode) çalışır, çıktı üstte birikir.
        val interactive = CemTab.interactive(project, toolWindow)
        val content = ContentFactory.getInstance()
            .createContent(interactive.component, "Interactive", false)
        content.isCloseable = false   // kapanmaz — kullanıcı her seferinde input'a yazar
        content.isPinned = true
        toolWindow.contentManager.addContent(content)
        toolWindow.contentManager.setSelectedContent(content)
    }
}

/**
 * Tek bir cem çıktısının container'ı. Action'lar her seferinde yeni tab
 * ister: CemTab.newRun(project, mode, snippet) → ContentFactory ile content
 * oluşturup ToolWindow'a addContent.
 */
class CemTab {
    val component: JPanel
    private val textPane = JTextPane().apply {
        isEditable = false
        background = JBColor.background()
    }
    /** Çalışan subprocess. Tab kapanırsa öldürülür. */
    @Volatile var process: Process? = null
    /** Tab kapatıldı mı (idempotency için). */
    @Volatile var cancelled = false

    init {
        component = JPanel(BorderLayout()).apply {
            add(JBScrollPane(textPane), BorderLayout.CENTER)
        }
    }

    /** Tab kapatıldığında çağırılır: subprocess'i öldür. */
    fun cancel() {
        if (cancelled) return
        cancelled = true
        process?.let { p ->
            if (p.isAlive) {
                p.destroy()
                // Grace period — sonra zorla
                Thread {
                    try { Thread.sleep(1500) } catch (_: InterruptedException) {}
                    if (p.isAlive) p.destroyForcibly()
                }.start()
            }
        }
    }

    fun appendHeader(mode: String, snippet: String) {
        val short = snippet.replace("\n", " ").take(80)
        appendStyled("─── cem $mode · $short ───\n", bold = true, color = JBColor.GRAY)
    }

    fun appendLine(s: String) {
        appendStyled(s + "\n", bold = false, color = JBColor.foreground())
    }

    fun appendDim(s: String) {
        appendStyled(s + "\n", bold = false, color = JBColor.GRAY)
    }

    fun appendError(s: String) {
        appendStyled(s + "\n", bold = true, color = JBColor.RED)
    }

    fun appendRaw(s: String) {
        appendStyled(s, bold = false, color = JBColor.foreground())
    }

    private fun appendStyled(text: String, bold: Boolean, color: java.awt.Color) {
        val doc = textPane.styledDocument
        val attrs = SimpleAttributeSet().apply {
            StyleConstants.setBold(this, bold)
            StyleConstants.setForeground(this, color)
            StyleConstants.setFontFamily(this, "Monospaced")
        }
        doc.insertString(doc.length, text, attrs)
        textPane.caretPosition = doc.length
    }

    companion object {
        /** Welcome tab (legacy — şimdi interactive ile değiştirildi). */
        fun welcome(): CemTab {
            val tab = CemTab()
            tab.appendStyled(welcomeText(), bold = false, color = JBColor.GRAY)
            return tab
        }

        private fun welcomeText(): String = """
            ⚡ cem — Compose · Execute · Multiplex
            One command, many AIs.

            Aşağıdaki input'a sorunu yaz, Enter'a bas → cem "..." (thinker)

            Editör shortcut'ları:
              Ctrl+Alt+I  →  cem: think on selection
              Ctrl+Alt+W  →  cem: write on selection
              Ctrl+Alt+P  →  cem: pair on selection (thinker → writer)
              Ctrl+Alt+A  →  cem: ask freely (custom prompt)

            Tip: editör seçimi olmadan kısayol → input dialog açılır.
            Settings → Tools → cem ile thinker/writer/model değiştirilir.

            ─── geçmiş ───

        """.trimIndent()

        /**
         * Etkileşimli "Interactive" tab — REPL benzeri:
         *  ┌──────────────────────────────┐
         *  │ Welcome + history (scroll)   │
         *  │                              │
         *  │                              │
         *  ├──────────────────────────────┤
         *  │ cem "...": [user types]   ⏎ │
         *  └──────────────────────────────┘
         * Enter → cem "<prompt>" (think mode) → çıktı history'ye eklenir.
         */
        fun interactive(project: Project, toolWindow: ToolWindow): CemTab {
            val tab = CemTab()
            tab.appendStyled(welcomeText(), bold = false, color = JBColor.GRAY)

            val input = JBTextField()
            input.toolTipText = "Sorunu yaz, Enter'a bas → cem \"...\" çalışır"
            val south = JPanel(BorderLayout()).apply {
                add(JLabel(" cem \"...\" › "), BorderLayout.WEST)
                add(input, BorderLayout.CENTER)
            }
            tab.component.add(south, BorderLayout.SOUTH)

            // Enter → submit
            input.actionMap.put("submit", object : AbstractAction() {
                override fun actionPerformed(e: ActionEvent) {
                    val prompt = input.text.trim()
                    if (prompt.isEmpty()) return
                    input.text = ""
                    tab.appendStyled("\n> $prompt\n", bold = true, color = JBColor.foreground())
                    // Aynı tab'a stream et — yeni tab açma. CemAction.launchCem
                    // yeni tab açar; biz inline çalıştıracağız.
                    launchInline(project, prompt, tab)
                }
            })
            input.inputMap.put(KeyStroke.getKeyStroke("ENTER"), "submit")
            return tab
        }

        /**
         * Interactive tab için — yeni tab AÇMADAN, mevcut tab'a stream et.
         */
        private fun launchInline(project: Project, prompt: String, tab: CemTab) {
            com.intellij.openapi.application.ApplicationManager.getApplication()
                .executeOnPooledThread {
                try {
                    val cemPath = CemAction.resolveCemBinary()
                    val pb = ProcessBuilder(cemPath, prompt).redirectErrorStream(true)
                    project.basePath?.let { pb.directory(java.io.File(it)) }
                    val process = pb.start()
                    tab.process = process
                    val reader = java.io.InputStreamReader(process.inputStream, Charsets.UTF_8)
                    val buf = CharArray(2048)
                    while (!tab.cancelled) {
                        val n = reader.read(buf)
                        if (n < 0) break
                        val chunk = String(buf, 0, n)
                        com.intellij.openapi.application.ApplicationManager.getApplication()
                            .invokeLater { tab.appendRaw(chunk) }
                    }
                    val exit = process.waitFor()
                    com.intellij.openapi.application.ApplicationManager.getApplication()
                        .invokeLater {
                            tab.appendDim(if (exit == 0) "─── done ───\n" else "─── exit $exit ───\n")
                        }
                } catch (e: Exception) {
                    com.intellij.openapi.application.ApplicationManager.getApplication()
                        .invokeLater { tab.appendError("cem hata: ${e.message}\n") }
                }
            }
        }

        /**
         * Yeni bir run-tab ekler, ToolWindow'u açıp seçili yapar.
         * Çağıran sonra tab.appendXxx ile çıktı yazar.
         *
         * Content disposer'ı tab.cancel()'a bağlanır → kullanıcı sekmeyi
         * X ile kapattığında subprocess öldürülür.
         */
        fun newRun(toolWindow: ToolWindow, mode: String, snippet: String): CemTab {
            val tab = CemTab()
            tab.appendHeader(mode, snippet)
            val title = "$mode · ${snippet.replace("\n", " ").take(30)}"
            val content: Content = ContentFactory.getInstance()
                .createContent(tab.component, title, true)
            content.isCloseable = true
            content.isPinned = false
            content.setDisposer(Disposable { tab.cancel() })
            toolWindow.contentManager.addContent(content)
            toolWindow.contentManager.setSelectedContent(content)
            toolWindow.show()
            return tab
        }
    }
}
