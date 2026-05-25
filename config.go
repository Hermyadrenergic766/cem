package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ─── Tipler ──────────────────────────────────────────────────────────────────

type Roles struct {
	Thinker string `yaml:"thinker"`
	Writer  string `yaml:"writer"`
}

type InstalledTool struct {
	Command string `yaml:"command"`
	Version string `yaml:"version,omitempty"`
}

type GlobalConfig struct {
	Version string                   `yaml:"version"`
	Tools   map[string]InstalledTool `yaml:"tools"`
	Roles   Roles                    `yaml:"roles"`
	Setup   bool                     `yaml:"setup_done"`
}

type ProjectConfig struct {
	Roles *Roles `yaml:"roles,omitempty"`
}

type ResolvedConfig struct {
	Global  *GlobalConfig
	Project *ProjectConfig
}

func (rc *ResolvedConfig) HasProjectConfig() bool {
	return rc.Project != nil && rc.Project.Roles != nil
}

func (rc *ResolvedConfig) ActiveRoles() Roles {
	if rc.HasProjectConfig() {
		r := *rc.Project.Roles
		if r.Thinker == "" {
			r.Thinker = rc.Global.Roles.Thinker
		}
		if r.Writer == "" {
			r.Writer = rc.Global.Roles.Writer
		}
		return r
	}
	return rc.Global.Roles
}

// ─── KnownTools ──────────────────────────────────────────────────────────────

type ToolMeta struct {
	Name        string
	Description string
	// Deprecated boş değilse setup/cemi listelerinde uyarı satırı olarak basılır.
	Deprecated string
	// InstallCmd — doğrudan exec (kabuk yok). nil → manuel kurulum.
	InstallCmd []string
	// InstallShellUnix — sh -c ile çalıştırılır (curl|bash gibi pipe'lar için).
	// Linux + macOS'ta InstallCmd'den önce gelir.
	InstallShellUnix string
	// InstallShellWin — cmd /c ile çalıştırılır. Windows'ta InstallCmd'den önce gelir.
	InstallShellWin string
	// VersionFlag — cemi listesinde sürümü göstermek için (örn. "--version").
	VersionFlag string
	// RunFlags — cem aracı çalıştırırken eklenen flag'ler. Boşsa stdin'le çağrılır;
	// "-p" gibi tek-atış flag'i, AI CLI'larının interaktif REPL'e geçmesini önler.
	RunFlags []string
	// PromptAsArg true ise input, RunFlags'ten sonra son pozisyonel arg olarak verilir;
	// false ise (varsayılan) stdin üzerinden pipe edilir. Codex 'exec "prompt"'
	// gibi pozisyonel pattern bekleyen araçlar için gerekli.
	PromptAsArg bool
}

// KnownTools — desteklenen AI CLI araçları. Description kullanıcıya gösterilir.
// Deprecated alanı doluysa setup/cemi listelerinde uyarı görüntülenir.
var KnownTools = map[string]ToolMeta{
	"claude": {
		Name:             "Claude",
		Description:      "Anthropic Claude Code (code.claude.com) — native installer, auto-update",
		InstallShellUnix: "curl -fsSL https://claude.ai/install.sh | bash",
		InstallShellWin:  "irm https://claude.ai/install.ps1 | iex",
		VersionFlag:      "--version",
		RunFlags:         []string{"-p"}, // print mode (non-interactive)
	},
	"agy": {
		Name:             "Antigravity",
		Description:      "Google Antigravity CLI — Gemini CLI'ın halefi (antigravity.google)",
		InstallShellUnix: "curl -fsSL https://antigravity.google/cli/install.sh | bash",
		InstallShellWin:  "irm https://antigravity.google/cli/install.ps1 | iex",
		VersionFlag:      "--version",
		RunFlags:         []string{"-p"},
	},
	"gpt": {
		Name:        "Codex",
		Description: "OpenAI Codex CLI (developers.openai.com/codex)",
		InstallCmd:  []string{"npm", "install", "-g", "@openai/codex"},
		VersionFlag: "--version",
		RunFlags:    []string{"exec"}, // non-interactive subcommand
		PromptAsArg: true,              // codex exec "prompt"
	},
	"cursor": {
		Name:        "Cursor",
		Description: "Cursor terminal agent (cursor-agent)",
		InstallCmd:  []string{"npm", "install", "-g", "cursor-agent"},
		VersionFlag: "--version",
	},
}

// orderedToolKeys — wizard/installer listelerinin sabit sırası.
// KnownTools map iterasyonu rastgele; UI tutarlılığı için bu liste kullanılır.
var orderedToolKeys = []string{
	"claude", "agy", "gpt", "cursor",
}

// ─── Yollar ──────────────────────────────────────────────────────────────────

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cem", "config.yaml"), nil
}

const projectConfigName = ".cem.yaml"

// ─── Global config ───────────────────────────────────────────────────────────

func loadGlobalConfig() (*GlobalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &GlobalConfig{
		Version: "1.0",
		Tools:   map[string]InstalledTool{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("global config parse: %w", err)
	}
	if cfg.Tools == nil {
		cfg.Tools = map[string]InstalledTool{}
	}
	return cfg, nil
}

func saveGlobalConfig(cfg *GlobalConfig) error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ─── Project config ──────────────────────────────────────────────────────────

func loadProjectConfig() (*ProjectConfig, error) {
	data, err := os.ReadFile(projectConfigName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	pc := &ProjectConfig{}
	if err := yaml.Unmarshal(data, pc); err != nil {
		return nil, fmt.Errorf("project config parse: %w", err)
	}
	if pc.Roles == nil {
		return nil, nil
	}
	return pc, nil
}

func SaveProjectConfig(pc *ProjectConfig) error {
	data, err := yaml.Marshal(pc)
	if err != nil {
		return err
	}
	return os.WriteFile(projectConfigName, data, 0o644)
}

// ─── Resolved ────────────────────────────────────────────────────────────────

func LoadConfig() (*ResolvedConfig, error) {
	g, err := loadGlobalConfig()
	if err != nil {
		return nil, err
	}
	p, err := loadProjectConfig()
	if err != nil {
		return nil, err
	}
	return &ResolvedConfig{Global: g, Project: p}, nil
}
