package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Kurulum sağlığını kontrol et",
	Long: `  cem doctor   → Kurulu araçlar, PATH, config, binary konumları
                  ve roller tutarlılığı için tanı raporu.`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(BannerCem)
		runDoctor()
	},
}

func runDoctor() {
	rc, err := LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render("✗ config yüklenemedi: " + err.Error()))
		os.Exit(1)
	}

	ok, warn, fail := 0, 0, 0
	tick := func(state string, line string) {
		switch state {
		case "ok":
			fmt.Println("  " + styleSuccess.Render("✓") + " " + line)
			ok++
		case "warn":
			fmt.Println("  " + styleWarn.Render("⚠") + " " + line)
			warn++
		case "fail":
			fmt.Println("  " + styleError.Render("✗") + " " + line)
			fail++
		}
	}

	fmt.Println(styleBold.Render("  Sistem"))
	tick("ok", fmt.Sprintf("%s/%s · Go runtime: %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version()))

	home, _ := os.UserHomeDir()
	cemDir := filepath.Join(home, ".cem")
	if _, err := os.Stat(cemDir); err == nil {
		tick("ok", "~/.cem dizini → "+styleDim.Render(cemDir))
	} else {
		tick("warn", "~/.cem dizini yok (ilk çalıştırmada oluşur)")
	}

	gp, _ := globalConfigPath()
	if _, err := os.Stat(gp); err == nil {
		tick("ok", "global config → "+styleDim.Render(gp))
	} else {
		tick("warn", "global config yok → "+styleDim.Render("cem setup ile oluştur"))
	}

	if rc.HasProjectConfig() {
		tick("ok", "proje config → "+styleDim.Render(".cem.yaml (global override aktif)"))
	} else {
		tick("ok", "proje config yok (global geçerli)")
	}

	fmt.Println()
	fmt.Println(styleBold.Render("  Roller"))
	roles := rc.ActiveRoles()
	if roles.Thinker == "" {
		tick("fail", "thinker atanmamış → "+styleDim.Render("cem roles claude"))
	} else if _, found := rc.Global.Tools[roles.Thinker]; !found {
		tick("warn", "thinker '"+roles.Thinker+"' config'de kayıtlı değil → "+
			styleDim.Render("cemi "+roles.Thinker))
	} else {
		tick("ok", "thinker → "+styleBold.Render(roles.Thinker))
	}
	if roles.Writer == "" {
		tick("fail", "writer atanmamış → "+styleDim.Render("cem roles - agy"))
	} else if _, found := rc.Global.Tools[roles.Writer]; !found {
		tick("warn", "writer '"+roles.Writer+"' config'de kayıtlı değil → "+
			styleDim.Render("cemi "+roles.Writer))
	} else {
		tick("ok", "writer → "+styleBold.Render(roles.Writer))
	}

	fmt.Println()
	fmt.Println(styleBold.Render("  Araçlar (PATH kontrolü)"))
	order := []string{"claude", "agy", "aider", "gemini", "gpt"}
	for _, key := range order {
		meta := KnownTools[key]
		cmd := resolveCommand(key, rc)
		path, lerr := exec.LookPath(cmd)
		_, registered := rc.Global.Tools[key]

		switch {
		case lerr == nil && registered:
			tick("ok", fmt.Sprintf("%-8s %s", styleBold.Render(meta.Name), styleDim.Render(path)))
		case lerr == nil && !registered:
			tick("warn", fmt.Sprintf("%-8s PATH'da var ama config'e kayıtlı değil → cemi %s",
				styleBold.Render(meta.Name), key))
		case lerr != nil && registered:
			tick("fail", fmt.Sprintf("%-8s config'de kayıtlı ama PATH'da yok",
				styleBold.Render(meta.Name)))
		default:
			tick("ok", fmt.Sprintf("%-8s %s", meta.Name, styleDim.Render("kurulu değil")))
		}
	}

	fmt.Println()
	fmt.Println(styleBold.Render("  Binary'ler"))
	for _, name := range []string{"cem", "cemi", "cemir"} {
		path, err := exec.LookPath(name)
		if err != nil {
			tick("warn", name+" → "+styleDim.Render("PATH'da bulunamadı"))
			continue
		}
		tick("ok", fmt.Sprintf("%-6s %s", styleBold.Render(name), styleDim.Render(path)))
	}

	fmt.Println()
	pathParts := filepath.SplitList(os.Getenv("PATH"))
	tick("ok", fmt.Sprintf("PATH girişi: %d dizin", len(pathParts)))
	hasLocal := false
	for _, p := range pathParts {
		if strings.Contains(p, ".local/bin") || strings.Contains(p, "/usr/local/bin") {
			hasLocal = true
			break
		}
	}
	if !hasLocal {
		tick("warn", "~/.local/bin veya /usr/local/bin PATH'da değil")
	}

	fmt.Println()
	summary := fmt.Sprintf("ok:%d  uyarı:%d  hata:%d", ok, warn, fail)
	switch {
	case fail > 0:
		fmt.Println("  " + styleError.Render("● Sistem sorunlu — ") + summary)
	case warn > 0:
		fmt.Println("  " + styleWarn.Render("● Sistem çalışıyor, eksikler var — ") + summary)
	default:
		fmt.Println("  " + styleSuccess.Render("● Sistem sağlıklı — ") + summary)
	}
	fmt.Println()
}
