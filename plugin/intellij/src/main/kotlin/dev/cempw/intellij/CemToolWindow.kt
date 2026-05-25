package dev.cempw.intellij

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.openapi.wm.ToolWindowManager
import com.intellij.ui.JBColor
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import java.awt.BorderLayout
import javax.swing.JPanel
import javax.swing.JTextPane
import javax.swing.text.SimpleAttributeSet
import javax.swing.text.StyleConstants

/**
 * Tool window that displays cem output. One instance per project; auto-created
 * the first time a cem action runs.
 */
class CemToolWindowFactory : ToolWindowFactory {

    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        val panel = CemToolWindow(project)
        val content = ContentFactory.getInstance().createContent(panel.component, "", false)
        toolWindow.contentManager.addContent(content)
        instances[project] = panel
    }

    companion object {
        private val instances = mutableMapOf<Project, CemToolWindow>()
        fun show(project: Project): CemToolWindow {
            val tw = ToolWindowManager.getInstance(project).getToolWindow("cem")
                ?: error("cem tool window not registered")
            tw.show()
            return instances[project]
                ?: error("CemToolWindow not yet created for $project")
        }
    }
}

class CemToolWindow(@Suppress("UNUSED_PARAMETER") project: Project) {
    private val textPane = JTextPane().apply {
        isEditable = false
        background = JBColor.background()
    }
    val component: JPanel = JPanel(BorderLayout()).apply {
        add(JBScrollPane(textPane), BorderLayout.CENTER)
    }

    init {
        appendStyled(
            """
            ⚡ cem — Compose · Execute · Multiplex
            One command, many AIs.

            Send code from the editor to your configured AIs:
              Ctrl+Alt+I   →  cem: think on selection   (thinker AI)
              Ctrl+Alt+W   →  cem: write on selection   (writer AI)
              Ctrl+Alt+P   →  cem: pair on selection    (thinker → writer)

            Tip: with no selection, the whole active file is sent.

            Configure thinker/writer/model:  Settings → Tools → cem
            See also:  https://github.com/muslu/cem

            ─── output will appear below ───

            """.trimIndent(),
            bold = false,
            color = JBColor.GRAY,
        )
    }

    fun appendHeader(mode: String, snippet: String) {
        appendStyled("\n─── cem $mode · ${snippet.replace("\n", " ")} ───\n",
            bold = true, color = JBColor.GRAY)
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
}
