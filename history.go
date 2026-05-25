package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// historyMaxInputLen — log satırında inputtan tutulacak maksimum karakter
const historyMaxInputLen = 80

// historyPath — ~/.cem/history.log
func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cem", "history.log"), nil
}

func modeLabel(m Mode) string {
	switch m {
	case ModeThink:
		return "think"
	case ModeWrite:
		return "write"
	case ModePair:
		return "pair"
	}
	return "?"
}

// AppendHistory — bir komutu history.log'a yazar. Hata sessizce yutulur:
// log yazılamaması komut çalışmasını engellememeli.
func AppendHistory(mode Mode, role, input string, exitCode int) {
	path, err := historyPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	// Input'u tek satıra çevir + kısalt
	flat := strings.ReplaceAll(input, "\n", " ")
	flat = strings.ReplaceAll(flat, "\t", " ")
	if len(flat) > historyMaxInputLen {
		flat = flat[:historyMaxInputLen] + "…"
	}

	line := fmt.Sprintf("%s\t%s\t%s\t%d\t%s\n",
		time.Now().UTC().Format(time.RFC3339),
		modeLabel(mode),
		role,
		exitCode,
		flat,
	)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
