package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// cemir [araç]   → direkt kaldır
// cemir          → kurulu araçları listele

var cemirRootCmd = &cobra.Command{
	Use:              "cemir [tool]",
	Short:            "Uninstall AI CLI tools",
	Version:          version,
	Args:             cobra.ArbitraryArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		OpenSourceNotice()
		checkUpdateNotice()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if yes, _ := cmd.Flags().GetBool("yes"); yes {
			autoYes = true
		}
		PrintBanner(BannerCemir)

		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}

		// Argüman yoksa kurulu araçları listele
		if len(args) == 0 {
			printInstalledTools(cfg)
			return
		}

		target := strings.ToLower(args[0])

		if target == "all" {
			removeAll(cfg)
			return
		}

		// Tool key tanınmıyorsa yazım hatası kontrolü
		if _, ok := KnownTools[target]; !ok {
			suggestion := suggestTool(target)
			if suggestion != "" && askYN(fmt.Sprintf("  '%s' bilinmiyor — '%s' demek istedin mi?",
				target, styleBold.Render(suggestion))) {
				target = suggestion
			} else {
				fmt.Println(styleError.Render("✗ Unknown tool: " + target))
				os.Exit(1)
			}
		}

		if err := RemoveTool(target, cfg); err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
	},
}

func initCemirCmd() {
	cemirRootCmd.Flags().BoolP("yes", "y", false, "auto-accept all prompts")
}

// removeAll — kurulu tüm araçları onay sorarak sırayla kaldırır
func removeAll(cfg *GlobalConfig) {
	if len(cfg.Tools) == 0 {
		fmt.Println(styleDim.Render("  Kurulu araç yok."))
		return
	}

	fmt.Println(styleBold.Render("  Tüm kurulu araçlar kaldırılacak:"))
	for key := range cfg.Tools {
		fmt.Printf("  · %s\n", styleBold.Render(key))
	}
	fmt.Println()
	if !askYN("  Devam edilsin mi?") {
		fmt.Println(styleDim.Render("  İptal."))
		return
	}

	order := orderedToolKeys
	failed := []string{}
	for _, key := range order {
		if _, ok := cfg.Tools[key]; !ok {
			continue
		}
		if err := RemoveTool(key, cfg); err != nil {
			fmt.Println(styleWarn.Render("  ⚠ " + err.Error()))
			failed = append(failed, key)
		}
	}

	// Orphan girdiler: cfg.Tools'da var ama artık KnownTools'da yok (eski sürüm
	// kalıntısı, örn. v0.1.13'ten önce 'gemini'). Config'den siliyoruz; binary
	// kaldırma denemiyoruz (paket adı belirsiz).
	for key := range cfg.Tools {
		if _, known := KnownTools[key]; known {
			continue
		}
		delete(cfg.Tools, key)
		fmt.Printf("  %s %s  %s\n",
			styleSuccess.Render("✓"), styleBold.Render(key),
			styleDim.Render("(orphan — config'den temizlendi, binary varsa manuel sil)"))
	}
	if err := saveGlobalConfig(cfg); err != nil {
		fmt.Println(styleWarn.Render("  ⚠ config kaydedilemedi: " + err.Error()))
	}

	fmt.Println()
	if len(failed) > 0 {
		fmt.Printf("  %s %d araç kaldırılamadı: %s\n",
			styleWarn.Render("⚠"), len(failed), strings.Join(failed, ", "))
	} else {
		fmt.Println(styleSuccess.Render("  ✓ Tüm araçlar kaldırıldı."))
	}
}

func printInstalledTools(cfg *GlobalConfig) {
	if len(cfg.Tools) == 0 {
		fmt.Println(styleDim.Render("  Kurulu araç yok."))
		fmt.Println()
		return
	}

	fmt.Println(styleBold.Render("  Kurulu araçlar:"))
	fmt.Println()

	for key, tool := range cfg.Tools {
		v := tool.Version
		if v == "" {
			v = "?"
		}
		fmt.Printf("  %s %-12s %s\n",
			styleSuccess.Render("✓"),
			styleBold.Render(key),
			styleDim.Render("v"+v))
	}

	fmt.Println()
	fmt.Println(styleDim.Render("  cemir claude  →  Claude'u kaldır"))
	fmt.Println(styleDim.Render("  cemir agy     →  Agy'i kaldır"))
	fmt.Println()
}
