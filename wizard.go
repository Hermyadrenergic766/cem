package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleBox     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			Padding(0, 2)
)

// RunSetupWizard — ilk kurulum sihirbazı
func RunSetupWizard(cfg *GlobalConfig) error {
	toolOrder := []string{"claude", "agy", "aider", "gemini", "gpt"}

	fmt.Println(styleBox.Render(
		styleTitle.Render("Hangi AI düşünür, hangi AI yazar?") + "\n" +
			styleDim.Render("Bir kez seç, istediğin zaman değiştir: cem roles")))
	fmt.Println()

	// Araç listesi
	for i, key := range toolOrder {
		meta := KnownTools[key]
		installed := ""
		if _, ok := cfg.Tools[key]; ok {
			installed = styleSuccess.Render(" ✓")
		}
		fmt.Printf("  %s  %-12s %s%s\n",
			colorMuted.Render(fmt.Sprintf("[%d]", i+1)),
			styleBold.Render(meta.Name),
			colorTagline.Render(meta.Description),
			installed,
		)
	}
	fmt.Println()

	// Thinker seç
	thinker := pickTool("  🧠 Düşünen AI", toolOrder, cfg)
	if thinker == "" {
		return fmt.Errorf("iptal edildi")
	}

	// Writer seç
	fmt.Println()
	writer := pickTool("  ✍️  Yazan AI  ", toolOrder, cfg)
	if writer == "" {
		return fmt.Errorf("iptal edildi")
	}

	// Kurulu değilse kur
	fmt.Println()
	for _, key := range []string{thinker, writer} {
		if _, ok := cfg.Tools[key]; !ok {
			meta := KnownTools[key]
			if askYN(fmt.Sprintf("  %s kurulsun mu?", styleBold.Render(meta.Name))) {
				if err := InstallTool(key, cfg); err != nil {
					fmt.Println(styleWarn.Render("  ⚠ Kurulum başarısız: " + err.Error()))
					fmt.Println(styleDim.Render("    Manuel kurabilir, devam edebilirsiniz."))
				}
			}
		}
	}

	cfg.Roles = Roles{Thinker: thinker, Writer: writer}
	cfg.Setup = true

	if err := saveGlobalConfig(cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(styleSuccess.Render("  ✓ Hazır!"))
	fmt.Println()
	fmt.Printf("  🧠 %s  →  %s\n",
		styleBold.Render(thinker),
		colorMuted.Render(`cem "soru"`))
	fmt.Printf("  ✍️  %s   →  %s\n",
		styleBold.Render(writer),
		colorMuted.Render("cem -w \"görev\""))
	fmt.Printf("  🤝 %s + %s  →  %s\n",
		styleBold.Render(thinker),
		styleBold.Render(writer),
		colorMuted.Render("cem -p \"görev\""))
	fmt.Println()
	fmt.Println(styleDim.Render("  Roller değiştirmek için: cem roles claude agy"))
	fmt.Println(styleDim.Render("  Proje bazlı config için: cem init"))
	fmt.Println()

	return nil
}

// InstallTool — bir AI CLI aracını kur
func InstallTool(toolKey string, cfg *GlobalConfig) error {
	meta, ok := KnownTools[toolKey]
	if !ok {
		return fmt.Errorf("bilinmeyen araç: %s", toolKey)
	}
	if meta.InstallCmd == nil {
		fmt.Printf("  %s manuel kurulum gerekiyor\n", meta.Name)
		return nil
	}

	fmt.Printf("  ⏳ %s kuruluyor...\n", styleBold.Render(meta.Name))

	cmd := exec.Command(meta.InstallCmd[0], meta.InstallCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	version := ""
	if meta.VersionFlag != "" {
		out, _ := exec.Command(meta.InstallCmd[0], meta.VersionFlag).Output()
		version = strings.TrimSpace(string(out))
	}

	cfg.Tools[toolKey] = InstalledTool{Command: toolKey, Version: version}
	fmt.Printf("  %s %s kuruldu\n", styleSuccess.Render("✓"), meta.Name)
	return nil
}

// RemoveTool — bir AI CLI aracını kaldır
func RemoveTool(toolKey string, cfg *GlobalConfig) error {
	meta, ok := KnownTools[toolKey]
	if !ok {
		return fmt.Errorf("bilinmeyen araç: %s", toolKey)
	}
	if _, installed := cfg.Tools[toolKey]; !installed {
		return fmt.Errorf("%s zaten kurulu değil", toolKey)
	}
	if !askYN(fmt.Sprintf("  %s kaldırılsın mı?", styleBold.Render(meta.Name))) {
		fmt.Println("  İptal.")
		return nil
	}

	fmt.Printf("  ⏳ %s kaldırılıyor...\n", meta.Name)

	ic := meta.InstallCmd
	var unCmd *exec.Cmd
	if len(ic) >= 2 && ic[0] == "npm" {
		unCmd = exec.Command("npm", "uninstall", "-g", ic[len(ic)-1])
	} else if len(ic) >= 2 && ic[0] == "pip" {
		unCmd = exec.Command("pip", "uninstall", "-y", ic[len(ic)-1])
	} else {
		return fmt.Errorf("otomatik kaldırma desteklenmiyor")
	}

	unCmd.Stdout = os.Stdout
	unCmd.Stderr = os.Stderr
	if err := unCmd.Run(); err != nil {
		return err
	}

	delete(cfg.Tools, toolKey)
	if cfg.Roles.Thinker == toolKey {
		cfg.Roles.Thinker = ""
	}
	if cfg.Roles.Writer == toolKey {
		cfg.Roles.Writer = ""
	}

	fmt.Printf("  %s %s kaldırıldı\n", styleSuccess.Render("✓"), meta.Name)
	return saveGlobalConfig(cfg)
}

// ShowRoles — aktif rolleri güzel şekilde göster
func ShowRoles(rc *ResolvedConfig) {
	roles := rc.ActiveRoles()

	src := colorMuted.Render("  kaynak: ~/.cem/config.yaml  ") +
		styleDim.Render("(global)")
	if rc.HasProjectConfig() {
		src = colorYellow.Render("  kaynak: .cem.yaml  ") +
			styleDim.Render("(proje — global override)")
	}

	fmt.Println(styleBox.Render(
		styleTitle.Render("Aktif Roller") + "\n\n" +
			fmt.Sprintf("  🧠 %-10s  %s\n",
				styleBold.Render(roles.Thinker),
				colorMuted.Render(`cem "soru"`)) +
			fmt.Sprintf("  ✍️  %-10s  %s\n",
				styleBold.Render(roles.Writer),
				colorMuted.Render("cem -w \"görev\"")) +
			fmt.Sprintf("  🤝 %s + %s  %s",
				styleBold.Render(roles.Thinker),
				styleBold.Render(roles.Writer),
				colorMuted.Render("cem -p \"görev\"")),
	))
	fmt.Println(src)
	fmt.Println()

	// Kurulu araçlar
	if len(rc.Global.Tools) > 0 {
		fmt.Println(styleBold.Render("  Kurulu araçlar:"))
		for key, tool := range rc.Global.Tools {
			v := tool.Version
			if v == "" {
				v = "kurulu"
			}
			fmt.Printf("    %-12s %s\n", key, styleDim.Render(v))
		}
		fmt.Println()
	}

	fmt.Println(styleDim.Render("  cem roles claude agy         → global değiştir"))
	fmt.Println(styleDim.Render("  cem roles --here claude agy   → sadece bu proje"))
	fmt.Println(styleDim.Render("  cem init                      → proje wizard"))
	fmt.Println()
}

// pickTool — wizard için araç seçtir
func pickTool(prompt string, toolOrder []string, cfg *GlobalConfig) string {
	fmt.Printf("%s [1-%d]: ", styleBold.Render(prompt), len(toolOrder))
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" || input == "" {
			return ""
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(toolOrder) {
				key := toolOrder[idx-1]
				meta := KnownTools[key]
				suffix := ""
				if _, ok := cfg.Tools[key]; ok {
					suffix = styleSuccess.Render(" (kurulu)")
				}
				fmt.Printf("  → %s%s\n", styleBold.Render(meta.Name), suffix)
				return key
			}
		}
		fmt.Printf("  %s [1-%d]: ", styleWarn.Render("Geçersiz, tekrar gir"), len(toolOrder))
	}
}

func askYN(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes" || resp == "e" || resp == "evet"
}
