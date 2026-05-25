package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "API key'leri yönet (anthropic, openai) — rate limit'te otomatik rotasyon",
}

var keysAddCmd = &cobra.Command{
	Use:   "add <provider>",
	Short: "Bir provider için yeni API key ekle (interaktif)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := normalizeProvider(args[0])
		if provider == "" {
			fmt.Println(styleError.Render("✗ bilinmeyen provider — kullanılabilir: anthropic, openai"))
			os.Exit(1)
		}
		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		fmt.Printf("  %s için API key girin: ", styleBold.Render(provider))
		reader := bufio.NewReader(os.Stdin)
		val, _ := reader.ReadString('\n')
		val = strings.TrimSpace(val)
		if val == "" {
			fmt.Println(styleError.Render("✗ boş key"))
			os.Exit(1)
		}
		fmt.Print("  Etiket (opsiyonel, örn. 'personal'): ")
		label, _ := reader.ReadString('\n')
		label = strings.TrimSpace(label)

		if cfg.APIKeys == nil {
			cfg.APIKeys = map[string][]APIKey{}
		}
		cfg.APIKeys[provider] = append(cfg.APIKeys[provider], APIKey{Value: val, Label: label})
		if err := saveGlobalConfig(cfg); err != nil {
			fmt.Println(styleError.Render("✗ kaydedilemedi: " + err.Error()))
			os.Exit(1)
		}
		fmt.Printf("  %s %s key eklendi (#%d)\n",
			styleSuccess.Render("✓"), provider, len(cfg.APIKeys[provider]))
	},
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "Saklı key'leri listele (mask'lenmiş)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		if len(cfg.APIKeys) == 0 {
			fmt.Println(styleDim.Render("  Hiç key yok. Ekle: cem keys add anthropic"))
			return
		}
		for _, provider := range []string{"anthropic", "openai"} {
			keys := cfg.APIKeys[provider]
			if len(keys) == 0 {
				continue
			}
			fmt.Println(styleBold.Render("  " + provider))
			for i, k := range keys {
				label := k.Label
				if label == "" {
					label = styleDim.Render("(etiket yok)")
				}
				fmt.Printf("    [%d] %s  %s\n", i+1, maskKey(k.Value), label)
			}
			fmt.Println()
		}
	},
}

var keysRemoveCmd = &cobra.Command{
	Use:   "remove <provider> <index>",
	Short: "Belirli bir key'i sil (index list'teki sırası)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		provider := normalizeProvider(args[0])
		idx, err := strconv.Atoi(args[1])
		if err != nil || idx < 1 {
			fmt.Println(styleError.Render("✗ geçersiz index"))
			os.Exit(1)
		}
		cfg, _ := loadGlobalConfig()
		keys := cfg.APIKeys[provider]
		if idx > len(keys) {
			fmt.Println(styleError.Render("✗ böyle bir key yok"))
			os.Exit(1)
		}
		cfg.APIKeys[provider] = append(keys[:idx-1], keys[idx:]...)
		_ = saveGlobalConfig(cfg)
		fmt.Printf("  %s %s #%d silindi\n", styleSuccess.Render("✓"), provider, idx)
	},
}

func init() {
	keysCmd.AddCommand(keysAddCmd, keysListCmd, keysRemoveCmd)
	rootCmd.AddCommand(keysCmd)
}

// maskKey — sk-ant-xxxxxxxxYYYY → "sk-ant-...YYYY"
func maskKey(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	prefix := s[:7] // "sk-ant-", "sk-proj-", "AIza..."
	suffix := s[len(s)-4:]
	return prefix + "..." + suffix
}

// normalizeProvider — kullanıcı girdisini bilinen provider adlarına eşler.
func normalizeProvider(s string) string {
	switch strings.ToLower(s) {
	case "anthropic", "claude":
		return "anthropic"
	case "openai", "codex", "gpt":
		return "openai"
	}
	return ""
}
