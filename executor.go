package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Mode int

const (
	ModeThink Mode = iota
	ModeWrite
	ModePair
)

// ReadStdin — pipe ile gelen veriyi okur (interaktif tty ise boş döner)
func ReadStdin() string {
	info, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if (info.Mode() & os.ModeCharDevice) != 0 {
		return ""
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(data), "\n")
}

// Run — seçili moda göre AI(ları) çalıştırır
func Run(input string, mode Mode, rc *ResolvedConfig) error {
	roles := rc.ActiveRoles()

	switch mode {
	case ModeThink:
		if roles.Thinker == "" {
			return errMissingRole("thinker")
		}
		return runTool(roles.Thinker, rc, input, "🧠")

	case ModeWrite:
		if roles.Writer == "" {
			return errMissingRole("writer")
		}
		return runTool(roles.Writer, rc, input, "✍️")

	case ModePair:
		if roles.Thinker == "" {
			return errMissingRole("thinker")
		}
		if roles.Writer == "" {
			return errMissingRole("writer")
		}

		fmt.Println(styleDim.Render("  🧠 " + roles.Thinker + " düşünüyor..."))
		thought, err := captureTool(roles.Thinker, rc, input)
		if err != nil {
			return err
		}
		fmt.Println(thought)
		fmt.Println()
		fmt.Println(styleDim.Render("  ✍️  " + roles.Writer + " yazıyor..."))

		writerInput := input + "\n\n--- Thinker analizi ---\n" + thought
		return runTool(roles.Writer, rc, writerInput, "✍️")
	}
	return fmt.Errorf("bilinmeyen mod")
}

func errMissingRole(name string) error {
	msg := fmt.Sprintf("%s rolü atanmamış — cem roles ile ayarla", name)
	fmt.Println(styleError.Render("✗ " + msg))
	return fmt.Errorf("%s", msg)
}

// resolveCommand — config'de saklanan command tercih edilir, yoksa tool key
func resolveCommand(toolKey string, rc *ResolvedConfig) string {
	if t, ok := rc.Global.Tools[toolKey]; ok && t.Command != "" {
		return t.Command
	}
	return toolKey
}

// runTool — stdin'i pipe edip stdout/stderr'i kullanıcıya gösterir
func runTool(toolKey string, rc *ResolvedConfig, input, icon string) error {
	bin := resolveCommand(toolKey, rc)
	if _, err := exec.LookPath(bin); err != nil {
		fmt.Println(styleError.Render(
			fmt.Sprintf("✗ %s bulunamadı — kurmak için: cemi %s", bin, toolKey)))
		return err
	}

	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println(styleError.Render("✗ " + bin + " hata: " + err.Error()))
		return err
	}
	return nil
}

// captureTool — pair modu için: çıktıyı yakalar
func captureTool(toolKey string, rc *ResolvedConfig, input string) (string, error) {
	bin := resolveCommand(toolKey, rc)
	if _, err := exec.LookPath(bin); err != nil {
		fmt.Println(styleError.Render(
			fmt.Sprintf("✗ %s bulunamadı — kurmak için: cemi %s", bin, toolKey)))
		return "", err
	}

	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}
