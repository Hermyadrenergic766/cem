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
	Use:     "cemir [araç]",
	Short:   "AI araçlarını kaldır",
	Version: "1.0.0",
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
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

		if err := RemoveTool(target, cfg); err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
	},
}

func initCemirCmd() {}

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
