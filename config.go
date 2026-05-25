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
	Deprecated  string
	InstallCmd  []string
	VersionFlag string
}

// KnownTools — desteklenen AI CLI araçları. Description kullanıcıya gösterilir.
// Deprecated alanı doluysa setup/cemi listelerinde uyarı görüntülenir.
var KnownTools = map[string]ToolMeta{
	"claude": {
		Name:        "Claude",
		Description: "Anthropic Claude Code (anthropic.com)",
		InstallCmd:  []string{"npm", "install", "-g", "@anthropic-ai/claude-code"},
		VersionFlag: "--version",
	},
	"agy": {
		Name:        "Antigravity",
		Description: "Google Antigravity — autonomous coding agent (formerly Gemini CLI)",
		InstallCmd:  []string{"npm", "install", "-g", "@google/antigravity-cli"},
		VersionFlag: "--version",
	},
	"aider": {
		Name:        "Aider",
		Description: "Open-source pair-programming AI (aider.chat)",
		InstallCmd:  []string{"pip", "install", "--upgrade", "aider-chat"},
		VersionFlag: "--version",
	},
	"gemini": {
		Name:        "Gemini",
		Description: "Google Gemini CLI",
		Deprecated:  "personal use ends 2026-06-16 — prefer 'agy' (Antigravity)",
		InstallCmd:  []string{"npm", "install", "-g", "@google/gemini-cli"},
		VersionFlag: "--version",
	},
	"gpt": {
		Name:        "Codex",
		Description: "OpenAI Codex CLI (formerly gpt CLI)",
		InstallCmd:  []string{"npm", "install", "-g", "@openai/codex"},
		VersionFlag: "--version",
	},
	"goose": {
		Name:        "Goose",
		Description: "Block's open-source AI agent (block.github.io/goose)",
		InstallCmd:  []string{"pip", "install", "--upgrade", "goose-ai"},
		VersionFlag: "--version",
	},
	"cody": {
		Name:        "Cody",
		Description: "Sourcegraph Cody CLI",
		InstallCmd:  []string{"npm", "install", "-g", "@sourcegraph/cody"},
		VersionFlag: "--version",
	},
	"continue": {
		Name:        "Continue",
		Description: "Continue.dev CLI — autopilot for VSCode/JetBrains",
		InstallCmd:  []string{"npm", "install", "-g", "@continuedev/cli"},
		VersionFlag: "--version",
	},
	"openhands": {
		Name:        "OpenHands",
		Description: "OpenHands (formerly OpenDevin) — autonomous SWE agent",
		InstallCmd:  []string{"pip", "install", "--upgrade", "openhands-ai"},
		VersionFlag: "--version",
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
	"claude", "agy", "aider", "gemini", "gpt",
	"goose", "cody", "continue", "openhands", "cursor",
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
