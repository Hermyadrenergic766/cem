package dev.cempw.intellij

import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.service
import com.intellij.openapi.options.Configurable
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBTextField
import com.intellij.util.ui.FormBuilder
import javax.swing.JComponent
import javax.swing.JPanel

@Service
@State(name = "CemSettings", storages = [Storage("cem.xml")])
class CemSettings : PersistentStateComponent<CemSettings.State> {
    data class State(
        var cemPath: String = "cem",
    )

    private var state = State()
    override fun getState() = state
    override fun loadState(s: State) {
        state = s
    }

    var cemPath: String
        get() = state.cemPath
        set(v) {
            state.cemPath = v
        }

    companion object {
        val instance: CemSettings get() = service()
    }
}

class CemSettingsConfigurable : Configurable {
    private var panel: JPanel? = null
    private var cemPathField: JBTextField? = null

    override fun getDisplayName() = "cem"

    override fun createComponent(): JComponent {
        cemPathField = JBTextField(CemSettings.instance.cemPath, 40)
        val form = FormBuilder.createFormBuilder()
            .addLabeledComponent(JBLabel("cem binary path:"), cemPathField!!, 1, false)
            .addComponent(JBLabel("Leave as 'cem' if it's on PATH. Otherwise full path."))
            .addComponentFillVertically(JPanel(), 0)
            .panel
        panel = form
        return form
    }

    override fun isModified(): Boolean =
        cemPathField?.text != CemSettings.instance.cemPath

    override fun apply() {
        CemSettings.instance.cemPath = cemPathField?.text ?: "cem"
    }

    override fun reset() {
        cemPathField?.text = CemSettings.instance.cemPath
    }

    override fun disposeUIResources() {
        panel = null
        cemPathField = null
    }
}
