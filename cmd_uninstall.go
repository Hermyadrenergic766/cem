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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "CEM'i sistemden kaldır",
	Long: `  cem uninstall   → cem, cemi, cemir binary'lerini sil
                    config klasörünü silmek ister misin diye sorar`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(BannerCem)
		runUninstall()
	},
}

func init_uninstall() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall() {
	fmt.Println(styleBold.Render("  CEM kaldırılacak."))
	fmt.Println(styleDim.Render("  Bu işlem cem, cemi ve cemir komutlarını siler."))
	fmt.Println()

	if !askYN("  Devam edilsin mi?") {
		fmt.Println(styleDim.Render("  İptal."))
		return
	}

	// ── Binary'leri bul ve sil ───────────────────────────────────────────────
	fmt.Println()
	fmt.Println(styleBold.Render("  Binary'ler aranıyor..."))

	names := []string{"cem", "cemi", "cemir"}
	if runtime.GOOS == "windows" {
		names = []string{"cem.exe", "cemi.exe", "cemir.exe"}
	}

	removed := 0
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err != nil {
			fmt.Printf("  %s %-8s bulunamadı, atlandı\n", styleDim.Render("○"), name)
			continue
		}
		if err := os.Remove(path); err != nil {
			// sudo gerekebilir
			if sudoRemove(path) {
				fmt.Printf("  %s %-8s %s\n",
					styleSuccess.Render("✓"), name, styleDim.Render(path))
				removed++
			} else {
				fmt.Printf("  %s %-8s silinemedi: %v\n",
					styleError.Render("✗"), name, err)
				fmt.Printf("    %s\n", styleDim.Render("Manuel: sudo rm "+path))
			}
		} else {
			fmt.Printf("  %s %-8s %s\n",
				styleSuccess.Render("✓"), name, styleDim.Render(path))
			removed++
		}
	}

	// ── Config klasörü ────────────────────────────────────────────────────────
	fmt.Println()
	home, _ := os.UserHomeDir()
	cemDir := filepath.Join(home, ".cem")

	if _, err := os.Stat(cemDir); err == nil {
		fmt.Printf("  Config klasörü: %s\n", styleDim.Render(cemDir))
		if askYN("  Config ve ayarlar da silinsin mi?") {
			if err := os.RemoveAll(cemDir); err != nil {
				fmt.Printf("  %s Config silinemedi: %v\n", styleError.Render("✗"), err)
			} else {
				fmt.Printf("  %s Config silindi\n", styleSuccess.Render("✓"))
			}
		} else {
			fmt.Printf("  %s Config korundu → %s\n",
				styleDim.Render("○"), cemDir)
		}
	}

	// ── Proje .cem.yaml ────────────────────────────────────────────────────────
	if _, err := os.Stat(".cem.yaml"); err == nil {
		fmt.Println()
		if askYN("  Bu dizindeki .cem.yaml da silinsin mi?") {
			os.Remove(".cem.yaml")
			fmt.Printf("  %s .cem.yaml silindi\n", styleSuccess.Render("✓"))
		}
	}

	// ── Sonuç ─────────────────────────────────────────────────────────────────
	fmt.Println()
	if removed > 0 {
		fmt.Println(styleSuccess.Render("  ✓ CEM kaldırıldı."))
		fmt.Println()
		fmt.Println(styleDim.Render("  Yeniden kurmak için:"))
		if runtime.GOOS == "windows" {
			fmt.Println(styleDim.Render("  irm cem.pw/install.ps1 | iex"))
		} else {
			fmt.Println(styleDim.Render("  curl -fsSL cem.pw/install | sh"))
		}
	} else {
		fmt.Println(styleWarn.Render("  ⚠ Hiçbir binary silinemedi."))
		fmt.Println(styleDim.Render("  sudo ile dene veya manuel sil."))
	}
	fmt.Println()
}

// sudoRemove — sudo ile silmeyi dene
func sudoRemove(path string) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	cmd := exec.Command("sudo", "rm", "-f", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// stdin'i bağla (sudo şifre sorabilir)
	cmd.Stdin = os.Stdin
	return cmd.Run() == nil
}

// PATH'daki tüm cem binary konumlarını bul (birden fazla olabilir)
func findAllBinaries() []string {
	var found []string
	names := []string{"cem", "cemi", "cemir"}
	if runtime.GOOS == "windows" {
		names = []string{"cem.exe", "cemi.exe", "cemir.exe"}
	}

	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		for _, name := range names {
			full := filepath.Join(dir, name)
			if _, err := os.Stat(full); err == nil {
				found = append(found, full)
			}
		}
	}

	// Tekrarları temizle
	seen := map[string]bool{}
	unique := found[:0]
	for _, f := range found {
		key := strings.ToLower(f)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, f)
		}
	}
	return unique
}
