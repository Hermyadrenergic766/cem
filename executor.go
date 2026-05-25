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
		printAIHeader("🧠 thinker", roles.Thinker, rc)
		return runTool(roles.Thinker, rc, input, "🧠")

	case ModeWrite:
		if roles.Writer == "" {
			return errMissingRole("writer")
		}
		printAIHeader("✍️  writer", roles.Writer, rc)
		return runTool(roles.Writer, rc, input, "✍️")

	case ModePair:
		if roles.Thinker == "" {
			return errMissingRole("thinker")
		}
		if roles.Writer == "" {
			return errMissingRole("writer")
		}

		thinkerLabel := roles.Thinker
		if m := resolveModel(roles.Thinker, rc); m != "" {
			thinkerLabel += " (" + m + ")"
		} else {
			thinkerLabel += " (default)"
		}
		// Header'ı ÖNCE bas — çıktı streaming geldiği için kullanıcı kimin
		// konuştuğunu hemen bilsin.
		printAIHeader("🧠 thinker", roles.Thinker, rc)
		sp := StartSpinner("🧠 " + thinkerLabel + " düşünüyor...")
		thought, err := captureToolWithSpinner(roles.Thinker, rc, input, sp)
		sp.Stop() // captureTool içinde de stopWriter durdurabilir; idempotent
		if err != nil {
			return err
		}
		// thought zaten captureTool tarafından stream edildi; tekrar basmıyoruz.

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
		printAIHeader("✍️  writer", roles.Writer, rc)
		writerInput := buildWriterPrompt(input, thought)
		return runTool(roles.Writer, rc, writerInput, "✍️")
	}
	return fmt.Errorf("bilinmeyen mod")
}

// buildWriterPrompt — writer'a düşünenin çıktısını + net "tekrar analiz
// yapma" talimatı ile besler. Amaç: writer prompt'u yorumlamak yerine
// doğrudan implementasyona geçsin, thinker'ın işini tekrarlamasın.
func buildWriterPrompt(originalTask, thinkerOutput string) string {
	return strings.Join([]string{
		"Görev için bir başka AI tarafından analiz/plan yapıldı. Senin işin: aşağıdaki",
		"analizi uygulayan KODU üretmek. Tekrar analiz yapma, plan açıklaması ekleme,",
		"trade-off tartışması yapma. Sadece çalışan, eksiksiz kodu yaz (gerekirse kısa",
		"inline yorum). Birden fazla dosya varsa açıkça belirt.",
		"",
		"=== ORİJİNAL GÖREV ===",
		originalTask,
		"",
		"=== THINKER ANALİZİ (TAKİP ET) ===",
		thinkerOutput,
		"",
		"=== ŞİMDİ KODU YAZ ===",
	}, "\n")
}

// buildArgs — bir AI CLI çağrısının komut argümanlarını oluşturur. Sıra:
//
//	ModelBeforeRun=true  → [--model X, RunFlags..., input?]
//	ModelBeforeRun=false → [RunFlags..., --model X, input?]
//
// agy/cursor -p "PROMPT" alır; --model -p ile prompt arasına girerse -p'nin
// değeri "--model" oluyor. Bunu önlemek için ModelBeforeRun=true.
func buildArgs(meta ToolMeta, toolKey string, rc *ResolvedConfig, input string) []string {
	model := resolveModel(toolKey, rc)
	includeModel := model != "" && meta.ModelFlag != ""
	args := []string{}
	if includeModel && meta.ModelBeforeRun {
		args = append(args, meta.ModelFlag, model)
	}
	args = append(args, meta.RunFlags...)
	if includeModel && !meta.ModelBeforeRun {
		args = append(args, meta.ModelFlag, model)
	}
	if meta.PromptAsArg {
		args = append(args, input)
	}
	return args
}

// resolveModel — toolKey için kullanılacak modeli döndürür. Sıra:
// 1) Proje config'i (.cem.yaml > models > <key>)
// 2) Global config (~/.cem/config.yaml > tools > <key> > model)
// 3) Boş → CLI default kullanılır.
//
// Tool'un ModelFlag'i yoksa (örn. agy) hangi config'de ne yazarsa yazsın
// "" döner — CLI default kullanılır, header de doğru "(default)" gösterir.
func resolveModel(toolKey string, rc *ResolvedConfig) string {
	if meta, ok := KnownTools[toolKey]; ok && meta.ModelFlag == "" {
		return "" // tool model seçimini desteklemiyor
	}
	if rc.Project != nil && rc.Project.Models != nil {
		if m, ok := rc.Project.Models[toolKey]; ok && m != "" {
			return m
		}
	}
	if t, ok := rc.Global.Tools[toolKey]; ok && t.Model != "" {
		return t.Model
	}
	return ""
}

// printAIHeader — her AI çıktısının üstüne kim olduğunu + hangi modeli
// kullandığını gösteren başlık. Örnek:
//
//	─── 🧠 thinker · claude (opus) ───
//	─── ✍️  writer · agy (gemini-3-flash) ───
//	─── 🧠 thinker · claude (default) ───   // model seçilmemiş, CLI default
func printAIHeader(role, toolKey string, rc *ResolvedConfig) {
	bar := strings.Repeat("─", 3)
	model := resolveModel(toolKey, rc)
	if model == "" {
		model = "default"
	}
	fmt.Println()
	fmt.Println(styleBold.Render(fmt.Sprintf("  %s %s · %s (%s) %s",
		bar, role, toolKey, model, bar)))
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

// authFailRe — stderr'de yetkilendirme hatası imzaları (401, eksik token,
// interaktif OAuth prompt'ları). Rate-limit'ten farklıdır: rotasyonla
// çözülmez, kullanıcı login/key müdahalesi gerekir.
var authFailRe = regexp.MustCompile(`(?i)(401|unauthorized|missing bearer|invalid api key|not.?logged.?in|please run /login|please log in|authentication failed|authentication required|please visit the url|paste the authorization code|authentication interrupted|waiting for authentication)`)

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
// toolKey paramı ile 'cem auth <toolKey>' önerebiliyoruz.
func hintAuth(bin, toolKey string, meta ToolMeta, cfg *GlobalConfig) {
	fmt.Println()
	fmt.Println(styleWarn.Render("  ⚠ " + bin + " yetkilendirilmemiş — auth eksik veya interaktif login akışı kesildi"))
	fmt.Println(styleDim.Render(fmt.Sprintf("    Önerilen: cem auth %s         (pano-yapıştır yardımcısı dahil)", toolKey)))
	if meta.Provider != "" {
		if len(cfg.APIKeys[meta.Provider]) > 0 {
			fmt.Println(styleDim.Render(fmt.Sprintf(
				"    Veya kayıtlı %d %s key var ama biri/hepsi geçersiz olabilir:",
				len(cfg.APIKeys[meta.Provider]), meta.Provider)))
			fmt.Println(styleDim.Render("      cem keys list"))
			fmt.Println(styleDim.Render(fmt.Sprintf("      cem keys remove %s <index>", meta.Provider)))
		} else {
			fmt.Println(styleDim.Render(fmt.Sprintf("    Veya yeni API key: cem keys add %s", meta.Provider)))
		}
	}
}

// stopWriter — ilk yazımda spinner'ı durdurur, sonrasında verileri inner'a iletir.
// Interactive prompt'ların (OAuth URL'leri, kod yapıştırma çağrıları) spinner
// tarafından üzerine yazılmasını engeller.
type stopWriter struct {
	sp      *Spinner
	inner   io.Writer
	stopped bool
}

func (w *stopWriter) Write(p []byte) (int, error) {
	if !w.stopped && w.sp != nil {
		w.sp.Stop()
		w.stopped = true
	}
	return w.inner.Write(p)
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
			lad := os.Getenv("LOCALAPPDATA")
			appd := os.Getenv("APPDATA")
			// Cursor native installer Windows'ta .cmd + .ps1 launcher koyar
			// (.exe değil — JS-tabanlı agent), root: %LOCALAPPDATA%\cursor-agent\
			if lad != "" {
				for _, base := range []string{
					filepath.Join(lad, "cursor-agent"),
					filepath.Join(lad, "Programs", "cursor-agent"),
					filepath.Join(lad, "Programs", "cursor"),
				} {
					for _, name := range []string{
						"cursor-agent.cmd", "cursor-agent.exe", "cursor-agent.ps1",
						"agent.cmd", "agent.exe",
					} {
						candidates = append(candidates, filepath.Join(base, name))
					}
				}
			}
			if appd != "" {
				// Legacy npm global bin (eski cemi npm install ile gelmişse)
				candidates = append(candidates,
					filepath.Join(appd, "npm", "cursor-agent.cmd"),
					filepath.Join(appd, "npm", "cursor-agent.ps1"),
					filepath.Join(appd, "npm", "cursor-agent"))
			}
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "cursor-agent.exe"),
				filepath.Join(home, ".local", "bin", "cursor-agent.cmd"))
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
	args := buildArgs(meta, toolKey, rc, input)

	return withKeyRotation(meta, rc.Global, func(env []string) error {
		cmd := exec.Command(bin, args...)
		if !meta.PromptAsArg {
			cmd.Stdin = strings.NewReader(input)
		}
		cmd.Stdout = os.Stdout
		// stderr'i hem konsola yansıt hem buffer'a yaz (rate-limit / auth imzasını yakalamak için).
		// runTool zaten spinner çalıştırmıyor, stopWriter pass-through olur.
		var errBuf bytes.Buffer
		cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)
		cmd.Env = env
		err := cmd.Run()
		if err != nil && looksLikeRateLimit(errBuf.String()) {
			return errRateLimit
		}
		if err != nil {
			if looksLikeAuthFailure(errBuf.String()) {
				hintAuth(bin, toolKey, meta, rc.Global)
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

	return captureToolWithSpinner(toolKey, rc, input, nil)
}

// captureToolWithSpinner — captureTool'un spinner-aware versiyonu. Subprocess
// ilk byte'ı stderr'e yazdığında verilen spinner durur (OAuth prompt'ları görünsün).
// sp nil ise düz capture.
func captureToolWithSpinner(toolKey string, rc *ResolvedConfig, input string, sp *Spinner) (string, error) {
	bin := resolveCommand(toolKey, rc)
	if _, err := exec.LookPath(bin); err != nil {
		fmt.Println(styleError.Render(
			fmt.Sprintf("✗ %s bulunamadı — kurmak için: cemi %s", bin, toolKey)))
		return "", err
	}
	meta := KnownTools[toolKey]
	args := buildArgs(meta, toolKey, rc, input)
	var captured bytes.Buffer
	err := withKeyRotation(meta, rc.Global, func(env []string) error {
		cmd := exec.Command(bin, args...)
		if !meta.PromptAsArg {
			cmd.Stdin = strings.NewReader(input)
		}
		captured.Reset()
		// Stream + capture: thinker çıktısı plugin/terminal'e ANINDA akar
		// ve buffer'a kopyalanır (writer fazı için).
		cmd.Stdout = io.MultiWriter(os.Stdout, &captured)
		var errBuf bytes.Buffer
		cmd.Stderr = &stopWriter{
			sp:    sp,
			inner: io.MultiWriter(os.Stderr, &errBuf),
		}
		cmd.Env = env
		runErr := cmd.Run()
		if runErr != nil && looksLikeRateLimit(errBuf.String()) {
			return errRateLimit
		}
		if runErr != nil && looksLikeAuthFailure(errBuf.String()) {
			hintAuth(bin, toolKey, meta, rc.Global)
		}
		return runErr
	})
	if err != nil {
		return "", err
	}
	return strings.TrimRight(captured.String(), "\n"), nil
}
