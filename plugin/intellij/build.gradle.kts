// IntelliJ Platform plugin for cem — runs `cem`/`cemi`/`cemir` on editor selection.
//
// Build:    ./gradlew buildPlugin
// Run IDE:  ./gradlew runIde
// Publish:  ./gradlew publishPlugin  (requires PUBLISH_TOKEN env var)
//
// Output:   build/distributions/cem-intellij-*.zip — installable via
//           PyCharm → Settings → Plugins → ⚙ → Install Plugin from Disk.

import org.jetbrains.changelog.Changelog
import org.jetbrains.intellij.platform.gradle.TestFrameworkType

plugins {
    id("java")
    id("org.jetbrains.kotlin.jvm") version "2.0.21"
    id("org.jetbrains.intellij.platform") version "2.1.0"
}

group = "dev.cempw"
version = providers.gradleProperty("pluginVersion").get()

kotlin {
    jvmToolchain(21)
}

repositories {
    mavenCentral()
    intellijPlatform { defaultRepositories() }
}

dependencies {
    intellijPlatform {
        // 2024.3 = IntelliJ Platform 243.x, PyCharm 2024.3, GoLand 2024.3 vb.
        intellijIdeaCommunity("2024.3")
        bundledPlugin("com.intellij.platform.images")
        testFramework(TestFrameworkType.Platform)
    }
    testImplementation("junit:junit:4.13.2")
}

intellijPlatform {
    pluginConfiguration {
        ideaVersion {
            sinceBuild = "243"
            untilBuild = provider { null }
        }
        changeNotes = providers.gradleProperty("pluginVersion").map { v ->
            "Plugin version $v. See <a href=\"https://github.com/muslu/cem/blob/main/CHANGELOG.md\">CHANGELOG</a>."
        }
    }
    publishing {
        token = providers.environmentVariable("PUBLISH_TOKEN")
    }
}

tasks {
    test {
        useJUnit()
    }
}
