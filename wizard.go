package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

	// Model seçimi — her rol için ayrı, varsayılan: CLI default'u (boş kalır)
	if !autoYes {
		fmt.Println()
		askModel(thinker, "🧠 thinker", cfg)
		if writer != thinker {
			askModel(writer, "✍️  writer", cfg)
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

// printTail — metnin son n satırını dim renkle yazdırır (hata bağlamı için).
func printTail(s string, n int) {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
		fmt.Println(styleDim.Render("  ... (çıktı kısaltıldı)"))
	}
	for _, l := range lines {
		fmt.Println(styleDim.Render("  " + l))
	}
}

// askModel — kullanıcıya bir tool için model seçtirir; seçimi cfg.Tools[key].Model'a
// kaydeder. Tool için ModelFlag tanımlı değilse sessizce döner.
func askModel(toolKey, label string, cfg *GlobalConfig) {
	meta, ok := KnownTools[toolKey]
	if !ok || meta.ModelFlag == "" || len(meta.Models) == 0 {
		return
	}
	fmt.Printf("  %s · %s için model:\n", styleBold.Render(label), styleBold.Render(meta.Name))
	for i, m := range meta.Models {
		marker := " "
		if t, ok := cfg.Tools[toolKey]; ok && t.Model == m {
			marker = styleSuccess.Render("✓")
		}
		fmt.Printf("    %s [%d] %s\n", marker, i+1, m)
	}
	fmt.Printf("      [%d] custom (kendi adını gir)\n", len(meta.Models)+1)
	fmt.Printf("      [0] default (CLI kendi seçer)\n")
	fmt.Print("  Seçim: ")
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(resp)

	t := cfg.Tools[toolKey] // zero-value ok
	switch resp {
	case "", "0":
		t.Model = "" // default
	default:
		idx, err := strconv.Atoi(resp)
		if err != nil || idx < 1 || idx > len(meta.Models)+1 {
			fmt.Println(styleDim.Render("  geçersiz, default kullanılacak"))
			t.Model = ""
		} else if idx == len(meta.Models)+1 {
			fmt.Print("  Model adı: ")
			line, _ := reader.ReadString('\n')
			t.Model = strings.TrimSpace(line)
		} else {
			t.Model = meta.Models[idx-1]
		}
	}
	if cfg.Tools == nil {
		cfg.Tools = map[string]InstalledTool{}
	}
	cfg.Tools[toolKey] = t
	if t.Model != "" {
		fmt.Printf("  %s model: %s\n", styleSuccess.Render("✓"), styleBold.Render(t.Model))
	}
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
	// Output'u yakala, başarısız olunca son satırları göster.
	var depBuf strings.Builder
	install.Stdout = &depBuf
	install.Stderr = &depBuf
	if err := install.Run(); err != nil {
		printTail(depBuf.String(), 12)
		fmt.Println(styleError.Render("  ✗ kurulum başarısız: " + err.Error()))
		return false
	}
	// Linux'ta nvm Node'u ~/.nvm/versions/node/<v>/bin/'e koyar; çalışan
	// cem süreci için PATH'i o dizinle güncelle ki LookPath bulsun.
	if runtime.GOOS == "linux" && (bin == "npm" || bin == "node") {
		home, _ := os.UserHomeDir()
		matches, _ := filepath.Glob(filepath.Join(home, ".nvm", "versions", "node", "*", "bin"))
		if len(matches) > 0 {
			// En son sürümü al (lexicographic son)
			nvmBin := matches[len(matches)-1]
			os.Setenv("PATH", nvmBin+string(os.PathListSeparator)+os.Getenv("PATH"))
		}
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
			// nvm — prebuilt Node binary. NodeSource artık glibc 2.28+ istiyor;
			// nvm Ubuntu 18.04 (libc 2.27) dahil her distro'da çalışıyor.
			return exec.Command("sh", "-c", `
set -e
if [ ! -s "$HOME/.nvm/nvm.sh" ]; then
  curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
fi
export NVM_DIR="$HOME/.nvm"
. "$NVM_DIR/nvm.sh"
nvm install --lts
nvm use --lts
`), "nvm + Node LTS (prebuilt, glibc bağımsız)"
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
	// Sessiz kurulum: çıktıyı buffer'a, hata durumunda son satırlar gösterilir.
	var instBuf strings.Builder
	cmd.Stdout = &instBuf
	cmd.Stderr = &instBuf
	if err := cmd.Run(); err != nil {
		printTail(instBuf.String(), 15)
		return err
	}

	// ToolMeta.Binary set ise PATH'da o adla aranır (örn. cursor → cursor-agent).
	binName := toolKey
	if meta.Binary != "" {
		binName = meta.Binary
	}

	version := ""
	if meta.VersionFlag != "" {
		out, _ := exec.Command(binName, meta.VersionFlag).Output()
		version = strings.TrimSpace(string(out))
	}

	// Shell-install (curl|bash, iwr|iex) sıklıkla PATH'i güncellemiyor;
	// önce PATH'da arıyoruz, yoksa bilinen kurulum konumlarını deniyoruz.
	command := binName
	if _, lookErr := exec.LookPath(binName); lookErr != nil {
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

	// Post-install auth setup: provider varsa kullanıcıya API key girme şansı ver,
	// yoksa CLI'ın kendi login akışını işaret et.
	postInstallAuthSetup(toolKey, meta, command, cfg)

	// Her kurulumdan sonra .gitignore güvenliği — proje config'i (.cem.yaml)
	// yanlışlıkla repo'ya gitmesin.
	ensureGitignoreSafe()

	return nil
}

// postInstallAuthSetup — provider tanımlı ise kullanıcıyı API key / login
// arasında seçim yapmaya yönlendirir. autoYes modunda atlanır (kullanıcı
// non-interactive istemiş, sessiz bırak).
func postInstallAuthSetup(toolKey string, meta ToolMeta, binPath string, cfg *GlobalConfig) {
	if meta.Provider == "" || meta.APIKeyEnv == "" {
		// OAuth-only araç (agy, cursor) — CLI kendi login'ini yönetir
		return
	}
	if autoYes {
		fmt.Println(styleDim.Render(fmt.Sprintf(
			"  ⓘ Auth: 'cem keys add %s' ile key gir veya '%s' çalıştırıp login ol",
			meta.Provider, filepath.Base(binPath))))
		return
	}
	fmt.Println()
	fmt.Printf("  %s için auth:\n", styleBold.Render(meta.Name))
	fmt.Println(styleDim.Render("    [1] API key kaydet (çoklu key + rate-limit rotasyonu)"))
	fmt.Println(styleDim.Render("    [2] Subscription / OAuth login (sonra: '" + filepath.Base(binPath) + "' çalıştır)"))
	fmt.Println(styleDim.Render("    [3] Şimdilik atla"))
	fmt.Print("  Seçim [1-3]: ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	switch choice {
	case "1":
		for {
			fmt.Print("  API key: ")
			val, _ := reader.ReadString('\n')
			val = strings.TrimSpace(val)
			if val == "" {
				fmt.Println(styleDim.Render("  iptal"))
				return
			}
			fmt.Print("  Etiket (opsiyonel, örn. 'personal'): ")
			label, _ := reader.ReadString('\n')
			label = strings.TrimSpace(label)
			if cfg.APIKeys == nil {
				cfg.APIKeys = map[string][]APIKey{}
			}
			cfg.APIKeys[meta.Provider] = append(cfg.APIKeys[meta.Provider],
				APIKey{Value: val, Label: label})
			fmt.Printf("  %s %s key #%d kaydedildi\n",
				styleSuccess.Render("✓"), meta.Provider, len(cfg.APIKeys[meta.Provider]))
			if !askYN("  Başka key eklemek ister misin?") {
				return
			}
		}
	case "2":
		fmt.Println(styleDim.Render("  → " + filepath.Base(binPath) + "  (interaktif login akışı açılır)"))
	default:
		fmt.Println(styleDim.Render("  Atlandı. Sonra: cem keys add " + meta.Provider))
	}
}

// ensureGitignoreSafe — CWD bir git repo ise .gitignore'da .cem.yaml var mı
// kontrol eder; yoksa ekler ve kullanıcıya bildirir. Repo değilse sessizce
// döner. cem'in proje config'i yanlışlıkla repo'ya commit edilmesin diye.
func ensureGitignoreSafe() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	if _, err := os.Stat(filepath.Join(wd, ".git")); err != nil {
		return // git repo değil
	}
	gi := filepath.Join(wd, ".gitignore")
	entry := ".cem.yaml"
	data, err := os.ReadFile(gi)
	if err != nil {
		// .gitignore yok — uyar
		fmt.Println(styleWarn.Render(
			"  ⚠ Git repo'da .gitignore yok — .cem.yaml repo'ya gidebilir"))
		return
	}
	if strings.Contains("\n"+string(data)+"\n", "\n"+entry+"\n") {
		return // zaten var
	}
	f, err := os.OpenFile(gi, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	suffix := ""
	if len(data) > 0 && data[len(data)-1] != '\n' {
		suffix = "\n"
	}
	if _, err := f.WriteString(suffix + "\n# cem proje config\n" + entry + "\n"); err != nil {
		return
	}
	fmt.Println(styleSuccess.Render("  ✓ .gitignore'a '" + entry + "' eklendi"))
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

// autoYes — cemi -y veya benzeri etkileşimsiz mod aktifse tüm askYN
// çağrıları "y" döner. Komut başlangıcında set edilir, çıkışta sıfırlanmaz
// (kısa ömürlü süreç).
var autoYes bool

func askYN(prompt string) bool {
	if autoYes {
		fmt.Printf("%s (y/N): y  (auto)\n", prompt)
		return true
	}
	fmt.Printf("%s (y/N): ", prompt)
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes" || resp == "e" || resp == "evet"
}
