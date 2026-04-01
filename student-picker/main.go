package main

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/bubbletea/v2"
)

func main() {
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
	pattern := filepath.Join(os.TempDir(), "sp-icon-*.webp")
	matches, _ := filepath.Glob(pattern)
	for _, f := range matches {
		os.Remove(f)
	}

	clearImagesTermimg()
	cleanupTermimg()
}
