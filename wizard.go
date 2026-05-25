package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	toolOrder := orderedToolKeys

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
		if meta.Deprecated != "" {
			fmt.Printf("       %s %s\n",
				styleWarn.Render("⚠"),
				styleDim.Render(meta.Deprecated))
		}
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

// pickInstallShell — platforma uyan shell-install komutunu döndürür; yoksa "".
func pickInstallShell(meta ToolMeta) string {
	if runtime.GOOS == "windows" {
		return meta.InstallShellWin
	}
	return meta.InstallShellUnix
}

// ensureDep — bir bin'i PATH'da arar; yoksa veya çok eskise kullanıcıya öneri
// sunar ve (kullanıcı kabul ederse) OS paket yöneticisiyle kurmayı dener.
// true → bin artık PATH'da ve yeterince güncel.
func ensureDep(bin string) bool {
	if _, err := exec.LookPath(bin); err == nil {
		if depVersionOK(bin) {
			return true
		}
		fmt.Println(styleDim.Render(fmt.Sprintf("  ⚠ %s sürümü çok eski — modern sürüm kuruluyor", bin)))
	}
	install, label := depInstallCommand(bin)
	if install == nil {
		fmt.Println(styleError.Render(fmt.Sprintf("  ✗ %s PATH'de yok ve otomatik kurulum tanımlanmamış", bin)))
		return false
	}
	fmt.Println(styleDim.Render(fmt.Sprintf("  ⚠ %s eksik — kurmak için: %s", bin, label)))
	if !askYN("  Şimdi kurulsun mu?") {
		return false
	}
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		fmt.Println(styleError.Render("  ✗ kurulum başarısız: " + err.Error()))
		return false
	}
	_, err := exec.LookPath(bin)
	return err == nil
}

// depVersionOK — bin'in çıktısından major sürümü çekip minimum eşikle karşılaştırır.
// Claude Code/OpenAI Codex CLI gibi paketler Node 18+ ister; npm 3.x (Ubuntu 18.04)
// gibi antika sürümlerde install argv parse hatası verir.
func depVersionOK(bin string) bool {
	minMajor := map[string]int{
		"npm":     9, // Node 18 LTS ile gelir
		"node":    18,
		"python":  3,
		"python3": 3,
	}
	threshold, has := minMajor[bin]
	if !has {
		return true // bilinmeyen → kontrol etme
	}
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return false
	}
	s := strings.TrimSpace(string(out))
	s = strings.TrimPrefix(s, "v")
	major := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		major = major*10 + int(s[i]-'0')
	}
	return major >= threshold
}

// depInstallCommand — bin için OS-spesifik kurulum komutunu döndürür.
func depInstallCommand(bin string) (*exec.Cmd, string) {
	switch bin {
	case "npm", "node":
		switch runtime.GOOS {
		case "windows":
			return exec.Command("winget", "install", "-e", "--id", "OpenJS.NodeJS.LTS", "--accept-package-agreements", "--accept-source-agreements"),
				"winget install OpenJS.NodeJS.LTS"
		case "darwin":
			return exec.Command("brew", "install", "node"), "brew install node"
		case "linux":
			// Distro paketleri eski (Ubuntu 18.04 → Node 8). Claude Code Node 18+ ister.
			// NodeSource LTS setup script'iyle güncel Node kuruyoruz.
			if _, err := exec.LookPath("apt-get"); err == nil {
				return exec.Command("sh", "-c",
						"curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"),
					"NodeSource LTS + apt-get install nodejs"
			}
			if _, err := exec.LookPath("dnf"); err == nil {
				return exec.Command("sh", "-c",
						"curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo -E bash - && sudo dnf install -y nodejs"),
					"NodeSource LTS + dnf install nodejs"
			}
			if _, err := exec.LookPath("pacman"); err == nil {
				return exec.Command("sudo", "pacman", "-S", "--noconfirm", "nodejs", "npm"),
					"sudo pacman -S nodejs npm"
			}
		}
	case "python", "python3":
		switch runtime.GOOS {
		case "windows":
			return exec.Command("winget", "install", "-e", "--id", "Python.Python.3.12", "--accept-package-agreements", "--accept-source-agreements"),
				"winget install Python.Python.3.12"
		case "darwin":
			return exec.Command("brew", "install", "python"), "brew install python"
		case "linux":
			return exec.Command("sudo", "apt-get", "install", "-y", "python3", "python3-pip"),
				"sudo apt-get install -y python3 python3-pip"
		}
	}
	return nil, ""
}

// InstallTool — bir AI CLI aracını kur
func InstallTool(toolKey string, cfg *GlobalConfig) error {
	meta, ok := KnownTools[toolKey]
	if !ok {
		return fmt.Errorf("bilinmeyen araç: %s", toolKey)
	}

	shellCmd := pickInstallShell(meta)
	if shellCmd == "" && meta.InstallCmd == nil {
		fmt.Printf("  %s manuel kurulum gerekiyor\n", meta.Name)
		if meta.Description != "" {
			fmt.Printf("  → %s\n", meta.Description)
		}
		return nil
	}

	// Önkoşul: doğrudan komutsa ilk binary (npm, pip, python), shell-install ise sh/cmd zaten var.
	if shellCmd == "" && len(meta.InstallCmd) > 0 {
		if !ensureDep(meta.InstallCmd[0]) {
			return fmt.Errorf("%s eksik — %s kurulamadı", meta.InstallCmd[0], meta.Name)
		}
	}

	fmt.Printf("  ⏳ %s kuruluyor...\n", styleBold.Render(meta.Name))

	var cmd *exec.Cmd
	if shellCmd != "" {
		if runtime.GOOS == "windows" {
			// Doğrudan PowerShell'e ver. cmd /c üzerinden geçince çift-quote
			// nesting bazen child output'unu yutuyor.
			cmd = exec.Command("powershell", "-NoProfile",
				"-ExecutionPolicy", "Bypass", "-Command", shellCmd)
		} else {
			cmd = exec.Command("sh", "-c", shellCmd)
		}
	} else {
		cmd = exec.Command(meta.InstallCmd[0], meta.InstallCmd[1:]...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	version := ""
	if meta.VersionFlag != "" {
		// Versiyonu ararken tool adını kullan; shell install için meta.InstallCmd nil olabilir.
		out, _ := exec.Command(toolKey, meta.VersionFlag).Output()
		version = strings.TrimSpace(string(out))
	}

	// Shell-install (curl|bash, iwr|iex) sıklıkla PATH'i güncellemiyor;
	// önce PATH'da arıyoruz, yoksa bilinen kurulum konumlarını deniyoruz.
	command := toolKey
	if _, lookErr := exec.LookPath(toolKey); lookErr != nil {
		if fallback := fallbackInstallPath(toolKey); fallback != "" {
			command = fallback
			fmt.Println(styleDim.Render("    bulundu: " + fallback))
		} else {
			cfg.Tools[toolKey] = InstalledTool{Command: toolKey, Version: version}
			fmt.Printf("  %s %s kuruldu ama %s henüz PATH'de değil\n",
				styleWarn.Render("⚠"), meta.Name, toolKey)
			fmt.Println(styleDim.Render("    Yeni terminal aç (PATH bu oturumda yenilenmez)"))
			return nil
		}
	}
	cfg.Tools[toolKey] = InstalledTool{Command: command, Version: version}
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
	} else if pickInstallShell(meta) != "" {
		// Shell-installed tool: silinecek binary'yi config'den veya fallback'ten al.
		path := ""
		if t, ok := cfg.Tools[toolKey]; ok && t.Command != "" && filepath.IsAbs(t.Command) {
			path = t.Command
		} else {
			path = fallbackInstallPath(toolKey)
		}
		if path == "" {
			return fmt.Errorf("binary konumu bulunamadı — manuel sil")
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("%s silinemedi: %w", path, err)
		}
		fmt.Println(styleDim.Render("    silindi: " + path))
		// Boş kalan ana klasörü temizle (örn. %LOCALAPPDATA%\agy\bin → \agy)
		if dir := filepath.Dir(path); dir != "" {
			_ = os.Remove(dir)             // bin/
			_ = os.Remove(filepath.Dir(dir)) // agy/
		}
		delete(cfg.Tools, toolKey)
		fmt.Printf("  %s %s kaldırıldı\n", styleSuccess.Render("✓"), meta.Name)
		return saveGlobalConfig(cfg)
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
