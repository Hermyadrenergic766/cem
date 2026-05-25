package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// cemi [araç]    → direkt kur
// cemi all       → hepsini kur (onay sorarak)
// cemi           → listele + bilgi
// cemi update    → güncelle
// cemi update agy → sadece agy güncelle

var cemiRootCmd = &cobra.Command{
	Use:     "cemi [araç]",
	Short:   "AI araçlarını yükle",
	Version: version,
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(BannerCemi)

		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}

		// Argüman yoksa listele
		if len(args) == 0 {
			printToolList(cfg)
			return
		}

		target := strings.ToLower(args[0])

		if target == "all" {
			installAll(cfg)
			return
		}

		if target == "update" {
			// cemi update [araç]
			if len(args) >= 2 {
				updateTool(args[1], cfg)
			} else {
				updateAll(cfg)
			}
			return
		}

		// cemi claude / cemi agy / cemi aider ...
		if _, ok := KnownTools[target]; !ok {
			fmt.Println(styleError.Render("✗ Bilinmeyen araç: " + target))
			fmt.Println()
			printToolList(cfg)
			os.Exit(1)
		}

		if err := InstallTool(target, cfg); err != nil {
			fmt.Println(styleError.Render("✗ Kurulum başarısız: " + err.Error()))
			os.Exit(1)
		}

		if err := saveGlobalConfig(cfg); err != nil {
			fmt.Println(styleError.Render("✗ Config kaydedilemedi: " + err.Error()))
			os.Exit(1)
		}
	},
}

func initCemiCmd() {
	// Alt komut yok — her şey root Run'da handle edilir
}

// ─── Yardımcı fonksiyonlar ────────────────────────────────────────────────────

func installAll(cfg *GlobalConfig) {
	// Sıralı liste (map rastgele sıralı)
	order := orderedToolKeys

	for _, key := range order {
		meta, ok := KnownTools[key]
		if !ok || meta.InstallCmd == nil {
			continue
		}
		if _, installed := cfg.Tools[key]; installed {
			fmt.Printf("  %s %-10s zaten kurulu, atlandı\n",
				styleSuccess.Render("✓"), styleBold.Render(key))
			continue
		}
		if !askYN(fmt.Sprintf("\n  %s kurulsun mu?", styleBold.Render(meta.Name))) {
			fmt.Println(styleDim.Render("  atlandı"))
			continue
		}
		if err := InstallTool(key, cfg); err != nil {
			fmt.Println(styleWarn.Render("  ⚠ " + err.Error()))
		}
	}
	saveGlobalConfig(cfg)
	fmt.Println()
	fmt.Println(styleSuccess.Render("✓ Tamamlandı"))
	fmt.Println(styleDim.Render("  cem roles  →  kimlerin aktif olduğunu gör"))
}

func updateTool(name string, cfg *GlobalConfig) {
	if _, ok := cfg.Tools[name]; !ok {
		fmt.Printf("  %s kurulu değil — önce: cemi %s\n", name, name)
		return
	}
	fmt.Printf("  🔄 %s güncelleniyor...\n", styleBold.Render(name))
	if err := InstallTool(name, cfg); err != nil {
		fmt.Println(styleWarn.Render("  ⚠ " + err.Error()))
		return
	}
	saveGlobalConfig(cfg)
}

func updateAll(cfg *GlobalConfig) {
	if len(cfg.Tools) == 0 {
		fmt.Println(styleDim.Render("  Kurulu araç yok."))
		return
	}
	for key := range cfg.Tools {
		updateTool(key, cfg)
	}
}

func printToolList(cfg *GlobalConfig) {
	installed := len(cfg.Tools)
	available := len(KnownTools)

	fmt.Printf("  Kurulu: %s / %d araç\n\n",
		styleBold.Render(fmt.Sprintf("%d", installed)), available)

	order := orderedToolKeys
	for _, key := range order {
		meta := KnownTools[key]
		if t, ok := cfg.Tools[key]; ok {
			v := t.Version
			if v == "" {
				v = "kurulu"
			}
			fmt.Printf("  %s %-10s %s\n",
				styleSuccess.Render("✓"),
				styleBold.Render(key),
				styleDim.Render(v))
		} else {
			fmt.Printf("  %s %-10s %s\n",
				styleDim.Render("○"),
				key,
				styleDim.Render(meta.Description))
		}
		if meta.Deprecated != "" {
			fmt.Printf("    %s %s\n",
				styleWarn.Render("⚠ deprecated:"),
				styleDim.Render(meta.Deprecated))
		}
	}

	fmt.Println()
	fmt.Println(styleDim.Render("  cemi claude     → Claude kur"))
	fmt.Println(styleDim.Render("  cemi agy        → Agy kur"))
	fmt.Println(styleDim.Render("  cemi all        → hepsini kur"))
	fmt.Println(styleDim.Render("  cemi update     → hepsini güncelle"))
	fmt.Println(styleDim.Render("  cemi update agy → sadece agy güncelle"))
	fmt.Println()
}
