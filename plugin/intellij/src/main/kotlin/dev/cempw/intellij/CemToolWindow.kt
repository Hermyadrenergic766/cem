package dev.cempw.intellij

import com.intellij.openapi.Disposable
import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.JBColor
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.Content
import com.intellij.ui.content.ContentFactory
import java.awt.BorderLayout
import javax.swing.JPanel
import javax.swing.JTextPane
import javax.swing.text.SimpleAttributeSet
import javax.swing.text.StyleConstants

/**
 * cem output tool window. Each invocation (think/write/pair) opens a NEW
 * closable tab. A persistent "Welcome" tab gives usage tips on first open.
 */
class CemToolWindowFactory : ToolWindowFactory {
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        // Welcome tab (closable; user can dismiss permanently)
        val welcome = CemTab.welcome()
        val content = ContentFactory.getInstance()
            .createContent(welcome.component, "Welcome", true)
        content.isCloseable = true
        toolWindow.contentManager.addContent(content)
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
        /** Welcome tab: ilk açılışta kısayolları gösteren bilgi sekmesi. */
        fun welcome(): CemTab {
            val tab = CemTab()
            tab.appendStyled(
                """
                ⚡ cem — Compose · Execute · Multiplex
                One command, many AIs.

                Send code from the editor to your configured AIs:
                  Ctrl+Alt+I   →  cem: think on selection   (thinker AI)
                  Ctrl+Alt+W   →  cem: write on selection   (writer AI)
                  Ctrl+Alt+P   →  cem: pair on selection    (thinker → writer)

                Tip: with no selection, the whole active file is sent.

                Configure thinker/writer/model:  Settings → Tools → cem
                Repo:  https://github.com/muslu/cem

                Each invocation opens a new tab in this tool window.

                """.trimIndent(),
                bold = false,
                color = JBColor.GRAY,
            )
            return tab
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
            content.setDisposer(Disposable { tab.cancel() })
            toolWindow.contentManager.addContent(content)
            toolWindow.contentManager.setSelectedContent(content)
            toolWindow.show()
            return tab
        }
    }
}
