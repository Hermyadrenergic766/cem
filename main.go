package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	bin := filepath.Base(os.Args[0])
	// Windows uzantısını at: cem.exe → cem
	bin = strings.TrimSuffix(bin, ".exe")
	// Geliştirme: `go run .` çıktısında "cem" / "main" olabilir
	bin = strings.ToLower(bin)

	switch bin {
	case "cemi":
		initCemiCmd()
		if err := cemiRootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	case "cemir":
		initCemirCmd()
		if err := cemirRootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	default:
		// cem ve diğer adlar
		init_uninstall()
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
