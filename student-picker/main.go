package main

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/bubbletea/v2"
)

func main() {
	if f, err := tea.LogToFile("/tmp/debug.log", "debug"); err == nil {
		defer f.Close()
	}
	m := newModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.status != "" && !fm.statusIsErr {
		fmt.Println(fm.status)
	}

	cleanupTempFiles()
}

func cleanupTempFiles() {
	webpPattern := filepath.Join(os.TempDir(), "sp-icon-*.webp")
	if matches, _ := filepath.Glob(webpPattern); matches != nil {
		for _, f := range matches {
			os.Remove(f)
		}
	}

	pngPattern := filepath.Join(os.TempDir(), "sp-icon-*.png")
	if matches, _ := filepath.Glob(pngPattern); matches != nil {
		for _, f := range matches {
			os.Remove(f)
		}
	}

	clearImagesTermimg()
	cleanupTermimg()
}
