package main

import (
	"testing"
)

func TestActiveRoles_GlobalOnly(t *testing.T) {
	rc := &ResolvedConfig{
		Global: &GlobalConfig{
			Roles: Roles{Thinker: "claude", Writer: "agy"},
		},
	}
	got := rc.ActiveRoles()
	if got.Thinker != "claude" || got.Writer != "agy" {
		t.Fatalf("global roller bekleniyordu, alındı: %+v", got)
	}
	if rc.HasProjectConfig() {
		t.Fatal("HasProjectConfig false olmalı")
	}
}

func TestActiveRoles_ProjectOverridesGlobal(t *testing.T) {
	rc := &ResolvedConfig{
		Global: &GlobalConfig{
			Roles: Roles{Thinker: "claude", Writer: "agy"},
		},
		Project: &ProjectConfig{
			Roles: &Roles{Thinker: "gemini", Writer: "aider"},
		},
	}
	got := rc.ActiveRoles()
	if got.Thinker != "gemini" {
		t.Errorf("thinker proje değeri: gemini, alındı: %s", got.Thinker)
	}
	if got.Writer != "aider" {
		t.Errorf("writer proje değeri: aider, alındı: %s", got.Writer)
	}
	if !rc.HasProjectConfig() {
		t.Fatal("HasProjectConfig true olmalı")
	}
}

func TestActiveRoles_ProjectPartialFallsBackToGlobal(t *testing.T) {
	rc := &ResolvedConfig{
		Global: &GlobalConfig{
			Roles: Roles{Thinker: "claude", Writer: "agy"},
		},
		Project: &ProjectConfig{
			Roles: &Roles{Thinker: "gemini"},
		},
	}
	got := rc.ActiveRoles()
	if got.Thinker != "gemini" {
		t.Errorf("thinker proje değeri: gemini, alındı: %s", got.Thinker)
	}
	if got.Writer != "agy" {
		t.Errorf("writer global'a düşmeli: agy, alındı: %s", got.Writer)
	}
}

func TestKnownTools_HasExpectedKeys(t *testing.T) {
	expected := []string{
		"claude", "agy", "aider", "gemini", "gpt",
		"goose", "cody", "continue", "openhands", "cursor",
	}
	for _, key := range expected {
		meta, ok := KnownTools[key]
		if !ok {
			t.Errorf("KnownTools içinde eksik: %s", key)
			continue
		}
		if meta.Name == "" {
			t.Errorf("%s: Name boş", key)
		}
		if len(meta.InstallCmd) == 0 {
			t.Errorf("%s: InstallCmd boş", key)
		}
	}
}

func TestOrderedToolKeys_MatchesKnownTools(t *testing.T) {
	if len(orderedToolKeys) != len(KnownTools) {
		t.Errorf("orderedToolKeys (%d) ≠ KnownTools (%d) — UI sırası map'in tüm anahtarlarını içermeli",
			len(orderedToolKeys), len(KnownTools))
	}
	seen := map[string]bool{}
	for _, k := range orderedToolKeys {
		if seen[k] {
			t.Errorf("orderedToolKeys'de tekrar: %s", k)
		}
		seen[k] = true
		if _, ok := KnownTools[k]; !ok {
			t.Errorf("orderedToolKeys'de bulunan %q KnownTools'da yok", k)
		}
	}
}

func TestGemini_DeprecationAnnounced(t *testing.T) {
	meta := KnownTools["gemini"]
	if meta.Deprecated == "" {
		t.Error("gemini: 2026-06-16 personal-use deprecation Deprecated alanında belirtilmeli")
	}
}

func TestAgy_IsAntigravity(t *testing.T) {
	meta := KnownTools["agy"]
	if meta.Name != "Antigravity" {
		t.Errorf("agy.Name 'Antigravity' olmalı, alındı: %q", meta.Name)
	}
}

func TestKnownTools_InstallCmdManagers(t *testing.T) {
	allowed := map[string]bool{"npm": true, "pip": true}
	for key, meta := range KnownTools {
		if len(meta.InstallCmd) == 0 {
			continue
		}
		mgr := meta.InstallCmd[0]
		if !allowed[mgr] {
			t.Errorf("%s: beklenmeyen install yöneticisi %q (sadece npm/pip)", key, mgr)
		}
	}
}
