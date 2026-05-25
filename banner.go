package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Banner türleri
type BannerKind int

const (
	BannerCem   BannerKind = iota
	BannerCemi
	BannerCemir
)

var (
	colorAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	colorMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	colorURL     = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	colorTagline = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	colorGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	colorYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	colorRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

const asciiCem = `
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝`

func PrintBanner(kind BannerKind) {
	logo := colorAccent.Render(asciiCem)

	var subtitle, badge, tip string

	switch kind {
	case BannerCem:
		badge    = colorAccent.Render("  ⚡ Unified AI Orchestrator")
		subtitle = colorTagline.Render("  Birden fazla AI'ı tek komutla kullan")
		tip      = colorMuted.Render("  cem -p \"görev\"  →  pair modu")

	case BannerCemi:
		badge    = colorGreen.Render("  📦 AI Tool Installer")
		subtitle = colorTagline.Render("  AI CLI araçlarını yükle ve güncelle")
		tip      = colorMuted.Render("  cemi all  →  hepsini kur")

	case BannerCemir:
		badge    = colorRed.Render("  🗑  AI Tool Remover")
		subtitle = colorTagline.Render("  Kurulu AI araçlarını kaldır")
		tip      = colorMuted.Render("  cemir claude  →  claude'u kaldır")
	}

	url := colorURL.Render("  cem.pw")
	sep := colorMuted.Render("  " + repeat("─", 38))

	fmt.Println(logo)
	fmt.Println()
	fmt.Println(badge + colorMuted.Render("  ·  ") + url)
	fmt.Println(subtitle)
	fmt.Println(sep)
	fmt.Println(tip)
	fmt.Println()
}

// ShowConfigSource — hangi config kullanıldığını göster
func ShowConfigSource(rc *ResolvedConfig) {
	if rc.HasProjectConfig() {
		fmt.Println(colorYellow.Render("  📁 Proje config: ") +
			colorTagline.Render(".cem.yaml") +
			colorMuted.Render("  (global override)"))
	} else {
		fmt.Println(colorMuted.Render("  🌍 Global config: ~/.cem/config.yaml"))
	}
	fmt.Println()
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
