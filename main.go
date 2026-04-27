package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if os.Getenv("STUDENT_PICKER_DEBUG") == "1" {
		if f, err := tea.LogToFile("/tmp/student-picker-debug.log", "debug"); err == nil {
			defer f.Close()
		}
	}

	p := tea.NewProgram(newModel())

	defer func() {
		cleanupTermimg()
		cleanupTempFiles()
	}()

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(model); ok && fm.status != "" && !fm.statusIsErr {
		fmt.Println(fm.status)
	}
}

func cleanupTempFiles() {
	for _, pat := range []string{"sp-icon-*.webp", "sp-icon-*.png"} {
		if matches, _ := filepath.Glob(filepath.Join(os.TempDir(), pat)); matches != nil {
			for _, f := range matches {
				os.Remove(f)
			}
		}
	}
}
