package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
		printAIHeader("🧠 thinker", roles.Thinker)
		return runTool(roles.Thinker, rc, input, "🧠")

	case ModeWrite:
		if roles.Writer == "" {
			return errMissingRole("writer")
		}
		printAIHeader("✍️  writer", roles.Writer)
		return runTool(roles.Writer, rc, input, "✍️")

	case ModePair:
		if roles.Thinker == "" {
			return errMissingRole("thinker")
		}
		if roles.Writer == "" {
			return errMissingRole("writer")
		}

		sp := StartSpinner("🧠 " + roles.Thinker + " düşünüyor...")
		thought, err := captureTool(roles.Thinker, rc, input)
		sp.Stop()
		if err != nil {
			return err
		}
		printAIHeader("🧠 thinker", roles.Thinker)
		fmt.Println(thought)

		// Writer kararı:
		//   - Aynı AI ise (thinker == writer) tekrar çağırmak çıktıyı duplike eder.
		//   - Soru kod istemiyorsa ve thinker zaten kod üretmediyse writer atlanır.
		if roles.Thinker == roles.Writer {
			fmt.Println(styleDim.Render("\n  (writer = thinker, ikinci çağrı atlandı)"))
			return nil
		}
		if !hasCodeBlock(thought) && !looksLikeCodeRequest(input) {
			fmt.Println(styleDim.Render("\n  (yazılacak kod yok, writer atlandı)"))
			return nil
		}

		fmt.Println()
		printAIHeader("✍️  writer", roles.Writer)
		writerInput := input + "\n\n--- Thinker analizi ---\n" + thought
		return runTool(roles.Writer, rc, writerInput, "✍️")
	}
	return fmt.Errorf("bilinmeyen mod")
}

// printAIHeader — her AI çıktısının üstüne kim olduğunu belirten net bir başlık
// basar. Örnek: "─── 🧠 thinker · claude ───"
func printAIHeader(role, name string) {
	bar := strings.Repeat("─", 3)
	fmt.Println()
	fmt.Println(styleBold.Render(fmt.Sprintf("  %s %s · %s %s", bar, role, name, bar)))
}

func errMissingRole(name string) error {
	msg := fmt.Sprintf("%s rolü atanmamış — cem roles ile ayarla", name)
	fmt.Println(styleError.Render("✗ " + msg))
	return fmt.Errorf("%s", msg)
}

// resolveCommand — config'de saklanan command tercih edilir, yoksa tool key
func resolveCommand(toolKey string, rc *ResolvedConfig) string {
	binName := toolKey
	if meta, ok := KnownTools[toolKey]; ok && meta.Binary != "" {
		binName = meta.Binary
	}
	if t, ok := rc.Global.Tools[toolKey]; ok && t.Command != "" {
		// Config'deki yol hâlâ geçerli mi?
		if _, err := exec.LookPath(t.Command); err == nil {
			return t.Command
		}
	}
	// PATH'da düz isimle var mı?
	if _, err := exec.LookPath(binName); err == nil {
		return binName
	}
	// Bilinen kurulum konumlarını dene (bazı installer'lar PATH'i güncellemiyor).
	if p := fallbackInstallPath(toolKey); p != "" {
		return p
	}
	return binName
}

// rateLimitRe — stderr'de rate-limit / quota imzaları (provider'lar arası).
var rateLimitRe = regexp.MustCompile(`(?i)(rate.?limit|quota|429|too many requests|usage limit|overloaded)`)

// authFailRe — stderr'de yetkilendirme hatası imzaları (401, eksik token vb.).
// Rate-limit'ten farklıdır: rotasyonla çözülmez, kullanıcı login/key müdahalesi
// gerekir.
var authFailRe = regexp.MustCompile(`(?i)(401|unauthorized|missing bearer|invalid api key|not.?logged.?in|please run /login|please log in|authentication failed)`)

// errRateLimit — withKeyRotation iç sinyali. Dışarı sızmaz; tüm key'ler bittiğinde
// gerçek alt-process hatasına dönüşür.
var errRateLimit = errors.New("rate limit / quota")

func looksLikeRateLimit(stderr string) bool {
	return rateLimitRe.MatchString(stderr)
}

func looksLikeAuthFailure(stderr string) bool {
	return authFailRe.MatchString(stderr)
}

// hintAuth — auth hatası tespit edildiğinde kullanıcıya net düzeltme yolu sun.
func hintAuth(bin string, meta ToolMeta, cfg *GlobalConfig) {
	fmt.Println()
	fmt.Println(styleWarn.Render("  ⚠ " + bin + " yetkilendirilmemiş — auth eksik veya geçersiz"))
	if meta.Provider == "" {
		fmt.Println(styleDim.Render("    Login: " + bin + " (CLI'nin kendi akışı)"))
		return
	}
	if len(cfg.APIKeys[meta.Provider]) > 0 {
		fmt.Println(styleDim.Render(fmt.Sprintf(
			"    Kayıtlı %d %s key var ama biri/hepsi geçersiz olabilir:",
			len(cfg.APIKeys[meta.Provider]), meta.Provider)))
		fmt.Println(styleDim.Render("      cem keys list"))
		fmt.Println(styleDim.Render(fmt.Sprintf("      cem keys remove %s <index>", meta.Provider)))
		fmt.Println(styleDim.Render(fmt.Sprintf("      cem keys add %s", meta.Provider)))
	} else {
		fmt.Println(styleDim.Render(fmt.Sprintf(
			"    API key: cem keys add %s", meta.Provider)))
		fmt.Println(styleDim.Render("    veya login: " + bin + "  (CLI'nin kendi akışı)"))
	}
}

// withKeyRotation — meta.Provider varsa cfg.APIKeys[provider] içinden sırayla
// her key'i env değişkeni olarak set edip fn'i çağırır. fn errRateLimit
// dönerse sonraki key denenir. Provider tanımlı değilse fn bir kez OS env'iyle
// çalıştırılır (CLI'ın kendi auth'u devrede).
func withKeyRotation(meta ToolMeta, cfg *GlobalConfig, fn func(env []string) error) error {
	baseEnv := os.Environ()
	if meta.Provider == "" || meta.APIKeyEnv == "" {
		return fn(baseEnv)
	}
	keys := cfg.APIKeys[meta.Provider]
	if len(keys) == 0 {
		// Key tanımlanmamış → CLI'ın mevcut auth'unu kullan
		return fn(baseEnv)
	}
	var lastErr error
	for i, k := range keys {
		env := append(append([]string{}, baseEnv...), meta.APIKeyEnv+"="+k.Value)
		// Aynı env değişkeni baseEnv'de varsa Go'nun exec son tanımı kullanır,
		// yani append yeterli.
		err := fn(env)
		if err == nil {
			return nil
		}
		if !errors.Is(err, errRateLimit) {
			return err // başka bir hata: tekrar denemenin anlamı yok
		}
		lastErr = err
		label := k.Label
		if label == "" {
			label = fmt.Sprintf("#%d", i+1)
		}
		if i+1 < len(keys) {
			fmt.Println(styleWarn.Render(fmt.Sprintf("  ⚠ %s rate limit — sonraki key'e geçiliyor", label)))
		} else {
			fmt.Println(styleError.Render(fmt.Sprintf("  ✗ tüm %s key'leri rate limit", meta.Provider)))
		}
	}
	return lastErr
}

// codeRequestRe — input'ta kod yazma niyetini gösteren kelimeler (TR + EN).
var codeRequestRe = regexp.MustCompile(`(?i)\b(yaz|kod|script|fonksiyon|class|method|implement|kodla|oluştur|üret|döndür|export|function|code|write|build|generate|refactor|debug|fix)\b`)

// hasCodeBlock — metin markdown kod bloğu içeriyor mu (``` veya satır başı 4-boşluk değil).
func hasCodeBlock(s string) bool {
	return strings.Contains(s, "```")
}

// looksLikeCodeRequest — input metni kod yazılması/üretilmesi gerektiğini ima ediyor mu.
func looksLikeCodeRequest(s string) bool {
	return codeRequestRe.MatchString(s)
}

// fallbackInstallPath — araç PATH'da yoksa standart konumlarda arar.
func fallbackInstallPath(toolKey string) string {
	home, _ := os.UserHomeDir()
	candidates := []string{}
	switch toolKey {
	case "agy":
		if runtime.GOOS == "windows" {
			if lad := os.Getenv("LOCALAPPDATA"); lad != "" {
				candidates = append(candidates, filepath.Join(lad, "agy", "bin", "agy.exe"))
			}
		} else {
			candidates = append(candidates, filepath.Join(home, ".local", "bin", "agy"))
		}
	case "claude":
		// Native installer (claude.ai/install.sh) önce ~/.claude/local/bin/'e koyup
		// shell rc'lerine PATH ekler. Mevcut süreçte hâlâ yoksa direkt yolu deneyelim.
		if runtime.GOOS == "windows" {
			if lap := os.Getenv("LOCALAPPDATA"); lap != "" {
				candidates = append(candidates, filepath.Join(lap, "Claude", "claude.exe"))
			}
			candidates = append(candidates, filepath.Join(home, ".claude", "local", "claude.exe"))
		} else {
			candidates = append(candidates,
				filepath.Join(home, ".claude", "local", "claude"),
				filepath.Join(home, ".local", "bin", "claude"),
			)
		}
	case "cursor":
		if runtime.GOOS == "windows" {
			if lad := os.Getenv("LOCALAPPDATA"); lad != "" {
				candidates = append(candidates,
					filepath.Join(lad, "cursor-agent", "cursor-agent.exe"),
					filepath.Join(lad, "cursor-agent", "agent.exe"))
			}
		} else {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "cursor-agent"),
				filepath.Join(home, ".local", "bin", "agent"))
		}
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// runTool — stdin'i pipe edip stdout/stderr'i kullanıcıya gösterir
func runTool(toolKey string, rc *ResolvedConfig, input, icon string) error {
	bin := resolveCommand(toolKey, rc)
	if _, err := exec.LookPath(bin); err != nil {
		fmt.Println(styleError.Render(
			fmt.Sprintf("✗ %s bulunamadı — kurmak için: cemi %s", bin, toolKey)))
		return err
	}

	meta := KnownTools[toolKey]
	args := append([]string{}, meta.RunFlags...)
	if meta.ModelFlag != "" {
		if t, ok := rc.Global.Tools[toolKey]; ok && t.Model != "" {
			args = append(args, meta.ModelFlag, t.Model)
		}
	}
	if meta.PromptAsArg {
		args = append(args, input)
	}

	return withKeyRotation(meta, rc.Global, func(env []string) error {
		cmd := exec.Command(bin, args...)
		if !meta.PromptAsArg {
			cmd.Stdin = strings.NewReader(input)
		}
		cmd.Stdout = os.Stdout
		// stderr'i hem konsola yansıt hem buffer'a yaz (rate-limit / auth imzasını yakalamak için)
		var errBuf bytes.Buffer
		cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)
		cmd.Env = env
		err := cmd.Run()
		if err != nil && looksLikeRateLimit(errBuf.String()) {
			return errRateLimit
		}
		if err != nil {
			if looksLikeAuthFailure(errBuf.String()) {
				hintAuth(bin, meta, rc.Global)
			} else {
				fmt.Println(styleError.Render("✗ " + bin + " hata: " + err.Error()))
			}
		}
		return err
	})
}

// captureTool — pair modu için: çıktıyı yakalar
func captureTool(toolKey string, rc *ResolvedConfig, input string) (string, error) {
	bin := resolveCommand(toolKey, rc)
	if _, err := exec.LookPath(bin); err != nil {
		fmt.Println(styleError.Render(
			fmt.Sprintf("✗ %s bulunamadı — kurmak için: cemi %s", bin, toolKey)))
		return "", err
	}

	meta := KnownTools[toolKey]
	args := append([]string{}, meta.RunFlags...)
	if meta.ModelFlag != "" {
		if t, ok := rc.Global.Tools[toolKey]; ok && t.Model != "" {
			args = append(args, meta.ModelFlag, t.Model)
		}
	}
	if meta.PromptAsArg {
		args = append(args, input)
	}
	var out []byte
	err := withKeyRotation(meta, rc.Global, func(env []string) error {
		cmd := exec.Command(bin, args...)
		if !meta.PromptAsArg {
			cmd.Stdin = strings.NewReader(input)
		}
		var errBuf bytes.Buffer
		cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)
		cmd.Env = env
		var runErr error
		out, runErr = cmd.Output()
		if runErr != nil && looksLikeRateLimit(errBuf.String()) {
			return errRateLimit
		}
		return runErr
	})
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}
