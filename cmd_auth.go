package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var authCodeFlag string

var authCmd = &cobra.Command{
	Use:   "auth <tool>",
	Short: "Run a tool's login flow (clipboard helper for stuck OAuth paste)",
	Long: `  cem auth claude            → run Claude Code's login flow
  cem auth agy               → run Antigravity login
  cem auth gpt               → run Codex login
  cem auth cursor            → run Cursor login

  --code <kod>               → copy the OAuth code to system clipboard
                              before launching the CLI, then paste with
                              right-click (bypasses PowerShell PSReadline
                              paste truncation/timeout issues).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolKey := strings.ToLower(args[0])
		if _, ok := KnownTools[toolKey]; !ok {
			if sugg := suggestTool(toolKey); sugg != "" && askYN(fmt.Sprintf(
				"  '%s' bilinmiyor — '%s' demek istedin mi?",
				toolKey, styleBold.Render(sugg))) {
				toolKey = sugg
			} else {
				fmt.Println(styleError.Render("✗ Unknown tool: " + toolKey))
				os.Exit(1)
			}
		}
		meta := KnownTools[toolKey]
		cfg, err := loadGlobalConfig()
		if err != nil {
			fmt.Println(styleError.Render("✗ " + err.Error()))
			os.Exit(1)
		}
		rc := &ResolvedConfig{Global: cfg}
		bin := resolveCommand(toolKey, rc)
		if _, err := exec.LookPath(bin); err != nil {
			fmt.Println(styleError.Render(fmt.Sprintf(
				"✗ %s bulunamadı — önce kur: cemi %s", bin, toolKey)))
			os.Exit(1)
		}

		// --code verilmişse panoya kopyala. Tool çalışırken sağ-tık yapıştır.
		if authCodeFlag != "" {
			if err := copyToClipboard(authCodeFlag); err != nil {
				fmt.Println(styleWarn.Render("  ⚠ panoya kopyalama başarısız: " + err.Error()))
				fmt.Println(styleDim.Render("    Kodu kendin yapıştır: " + authCodeFlag))
			} else {
				fmt.Println(styleSuccess.Render("  ✓ kod panoya kopyalandı"))
				fmt.Println(styleDim.Render("    CLI prompt'unda sağ-tık ile yapıştır (PowerShell paste sorunlarını bypass eder)"))
			}
			fmt.Println()
		}

		// Tool'un login subcommand'ini çağır (AuthCmd boşsa düz bin).
		invokeArgs := append([]string{}, meta.AuthCmd...)
		c := exec.Command(bin, invokeArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		fmt.Printf("  %s %s %s\n",
			styleDim.Render("→"),
			styleBold.Render(bin),
			styleDim.Render(strings.Join(invokeArgs, " ")))
		if err := c.Run(); err != nil {
			fmt.Println(styleError.Render("✗ auth akışı hata verdi: " + err.Error()))
			os.Exit(1)
		}
	},
}

func init() {
	authCmd.Flags().StringVar(&authCodeFlag, "code", "",
		"OAuth code to copy to clipboard before launching the CLI")
	rootCmd.AddCommand(authCmd)
}

// copyToClipboard — OS-specific clipboard helper.
func copyToClipboard(s string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("powershell", "-NoProfile", "-Command", "$input | Set-Clipboard")
	case "darwin":
		c = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			c = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			c = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			c = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("wl-copy / xclip / xsel kurulu değil")
		}
	default:
		return fmt.Errorf("clipboard desteklenmiyor: %s", runtime.GOOS)
	}
	c.Stdin = strings.NewReader(s)
	return c.Run()
}
