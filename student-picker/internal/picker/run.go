package picker

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

// Run starts the student-picker TUI.
func Run() {
	if f, err := tea.LogToFile("/tmp/debug.log", "debug"); err == nil {
		defer f.Close()
	}

	m := newModel()
	p := tea.NewProgram(m)

	defer func() {
		cleanupTermimg()
		cleanupTempFiles()
	}()

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.status != "" && !fm.statusIsErr {
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
