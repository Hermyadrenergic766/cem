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
        // 2023.3 = IntelliJ Platform 233.x — sinceBuild ile aynı baseline.
        intellijIdeaCommunity("2023.3")
        bundledPlugin("com.intellij.platform.images")
        testFramework(TestFrameworkType.Platform)
    }
    // ~/.cem/config.yaml okuma/yazma için
    implementation("org.snakeyaml:snakeyaml-engine:2.7")
    testImplementation("junit:junit:4.13.2")
}

intellijPlatform {
    pluginConfiguration {
        ideaVersion {
            // PyCharm/IDEA/GoLand 2023.3+ (Kasım 2023). IntelliJ Platform Gradle
            // Plugin 2.16.0 minimum 233 destekliyor — 232 'too low' diye reddediyor.
            sinceBuild = "233"
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
    // CI'da headless IDE bazen exit 255 atıyor; arama indeksini pre-build etmek
    // opsiyonel (Search Everywhere'de plugin settings'i bulmaya yarar). Bizim
    // settings tek bir 'cem.path' field'ı, indeksleme yok = pratik sorun yok.
    buildSearchableOptions {
        enabled = false
    }
}
