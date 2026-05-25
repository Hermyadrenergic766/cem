// IntelliJ Platform plugin for cem — runs `cem`/`cemi`/`cemir` on editor selection.
//
// Build:    ./gradlew buildPlugin
// Run IDE:  ./gradlew runIde
// Publish:  ./gradlew publishPlugin  (requires PUBLISH_TOKEN env var)
//
// Output:   build/distributions/cem-intellij-*.zip — installable via
//           PyCharm → Settings → Plugins → ⚙ → Install Plugin from Disk.

import org.jetbrains.intellij.platform.gradle.TestFrameworkType

plugins {
    id("java")
    id("org.jetbrains.kotlin.jvm") version "2.0.21"
    id("org.jetbrains.intellij.platform") version "2.16.0"
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
        // 2023.2 = IntelliJ Platform 232.x — sinceBuild ile aynı baseline.
        // Daha eski IDE'lerle uyumlu compile etmek için bu sürümle derliyoruz.
        intellijIdeaCommunity("2023.2")
        bundledPlugin("com.intellij.platform.images")
        testFramework(TestFrameworkType.Platform)
    }
    testImplementation("junit:junit:4.13.2")
}

intellijPlatform {
    pluginConfiguration {
        ideaVersion {
            // PyCharm 2023.2+ (Ekim 2023). Daha geriye gitmek için kullandığımız
            // API'lerin uyumlu olduğundan emin olmak gerek.
            sinceBuild = "232"
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
