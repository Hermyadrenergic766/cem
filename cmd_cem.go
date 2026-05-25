package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	flagWrite bool
	flagPair  bool
	flagFile  string
)

var rootCmd = &cobra.Command{
	Use:     "cem [input]",
	Short:   "⚡ Unified AI orchestrator",
	Version: version,
	Args:    cobra.ArbitraryArgs,
	// Banner her çalıştırmada değil sadece help'te görünsün
	// Kullanım sırasında kısa prefix yeterli
	Run: func(cmd *cobra.Command, args []string) {
		rc, err := loadAndCheckSetup()
		if err != nil {
			os.Exit(1)
		}

		// Config kaynağını göster
		ShowConfigSource(rc)

		// Input: pipe > -f > args
		input := ReadStdin()

		if flagFile != "" {
			data, err := os.ReadFile(flagFile)
			if err != nil {
				fmt.Println(styleError.Render("✗ Dosya okunamadı: " + err.Error()))
				os.Exit(1)
			}
			if len(args) > 0 {
				input = strings.Join(args, " ") + "\n\n" + string(data)
			} else {
				input = string(data)
			}
		} else if len(args) > 0 {
			if input != "" {
				input = strings.Join(args, " ") + "\n\n" + input
			} else {
				input = strings.Join(args, " ")
			}
		}

		if input == "" {
			PrintBanner(BannerCem)
			cmd.Help()
			return
		}

		mode := ModeThink
		if flagPair {
			mode = ModePair
		} else if flagWrite {
			mode = ModeWrite
		}

		runErr := Run(input, mode, rc)

		// History (rol: pair → thinker+writer, write → writer, think → thinker)
		roles := rc.ActiveRoles()
		var role string
		switch mode {
		case ModeWrite:
			role = roles.Writer
		case ModePair:
			role = roles.Thinker + "+" + roles.Writer
		default:
			role = roles.Thinker
		}
		exit := 0
		if runErr != nil {
			exit = 1
		}
		AppendHistory(mode, role, input, exit)

		if runErr != nil {
			os.Exit(1)
		}
	},
}

// ─── cem roles ───────────────────────────────────────────────────────────────

var rolesCmdHere bool

var rolesCmd = &cobra.Command{
	Use:   "roles [thinker] [writer]",
	Short: "Rolleri göster veya değiştir",
	Long: `  cem roles                    → mevcut rolleri göster
  cem roles claude agy         → global değiştir
  cem roles gemini             → sadece thinker
  cem roles --here claude agy  → sadece bu proje için`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		rc, err := LoadConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}

		if len(args) == 0 {
			PrintBanner(BannerCem)
			ShowRoles(rc)
			return
		}

		current := rc.ActiveRoles()
		if len(args) >= 1 {
			current.Thinker = args[0]
		}
		if len(args) >= 2 {
			current.Writer = args[1]
		}

		if rolesCmdHere {
			pc := &ProjectConfig{Roles: &current}
			if err := SaveProjectConfig(pc); err != nil {
				fmt.Println(styleError.Render("✗ " + err.Error()))
				os.Exit(1)
			}
			fmt.Println(styleSuccess.Render("✓ Proje rolleri güncellendi → .cem.yaml"))
		} else {
			rc.Global.Roles = current
			if err := saveGlobalConfig(rc.Global); err != nil {
				fmt.Println(styleError.Render("✗ " + err.Error()))
				os.Exit(1)
			}
			fmt.Println(styleSuccess.Render("✓ Global roller güncellendi → ~/.cem/config.yaml"))
		}

		fmt.Printf("  🧠 Thinker → %s\n", styleBold.Render(current.Thinker))
		fmt.Printf("  ✍️  Writer  → %s\n", styleBold.Render(current.Writer))
		fmt.Println()
		fmt.Println(styleDim.Render("  cem roles  →  kontrol et"))
	},
}

// ─── cem setup ───────────────────────────────────────────────────────────────

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Kurulum sihirbazını yeniden çalıştır",
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(BannerCem)
		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		if err := RunSetupWizard(cfg); err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
	},
}

// ─── cem init ────────────────────────────────────────────────────────────────

var initCmd = &cobra.Command{
	Use:   "init [thinker] [writer]",
	Short: "Bu proje için .cem.yaml oluştur",
	Long: `  cem init                 → interaktif wizard
  cem init claude agy      → direkt oluştur`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		rc, err := LoadConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}

		global := rc.Global.Roles
		var t, w string

		if len(args) >= 2 {
			t, w = args[0], args[1]
		} else {
			fmt.Println()
			fmt.Println(styleBold.Render("  Proje rolleri") +
				styleDim.Render("  (Enter = global değeri kullanılır)"))
			fmt.Println()
			fmt.Printf("  🧠 Thinker [%s]: ", styleBold.Render(global.Thinker))
			t = readLine()
			if t == "" {
				t = global.Thinker
			}
			fmt.Printf("  ✍️  Writer  [%s]: ", styleBold.Render(global.Writer))
			w = readLine()
			if w == "" {
				w = global.Writer
			}
		}

		pc := &ProjectConfig{Roles: &Roles{Thinker: t, Writer: w}}
		if err := SaveProjectConfig(pc); err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println(styleSuccess.Render("✓ .cem.yaml oluşturuldu"))
		fmt.Printf("  🧠 Thinker → %s\n", styleBold.Render(t))
		fmt.Printf("  ✍️  Writer  → %s\n", styleBold.Render(w))
		fmt.Println()
		fmt.Println(styleDim.Render("  Bu dizinden çalışırken proje config geçerli olur."))
		fmt.Println(styleDim.Render("  cem roles  →  aktif rolleri gör"))
	},
}

// ─── cem status ──────────────────────────────────────────────────────────────

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Kurulum durumunu göster",
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(BannerCem)
		rc, err := LoadConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		ShowRoles(rc)
	},
}

// ─── init & execute ──────────────────────────────────────────────────────────

func init() {
	rootCmd.Flags().BoolVarP(&flagWrite, "write", "w", false, "Writer AI kullan")
	rootCmd.Flags().BoolVarP(&flagPair, "pair", "p", false, "Pair: thinker → writer")
	rootCmd.Flags().StringVarP(&flagFile, "file", "f", "", "Dosya içeriğini gönder")

	rolesCmd.Flags().BoolVar(&rolesCmdHere, "here", false, "Sadece bu proje için (.cem.yaml)")

	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
}

func loadAndCheckSetup() (*ResolvedConfig, error) {
	rc, err := LoadConfig()
	if err != nil {
		fmt.Println(styleError.Render("✗ Config yüklenemedi: " + err.Error()))
		return nil, err
	}
	if !rc.Global.Setup || rc.Global.Roles.Thinker == "" {
		PrintBanner(BannerCem)
		fmt.Println(styleWarn.Render("  ⚡ İlk çalıştırma — sihirbaz başlatılıyor...\n"))
		if err := RunSetupWizard(rc.Global); err != nil {
			return nil, err
		}
		rc, err = LoadConfig()
		if err != nil {
			return nil, err
		}
	}
	return rc, nil
}

func readLine() string {
	var s string
	fmt.Scanln(&s)
	return strings.TrimSpace(s)
}
