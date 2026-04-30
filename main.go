package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

// debugLog discards all output by default.
// Reassigned to a real logger when --debug is passed.
var debugLog = log.New(io.Discard, "", log.LstdFlags)

func main() {
	debug := flag.Bool("debug", false, "write debug log to /tmp/student-picker-debug.log")
	flag.Parse()

	if *debug {
		f, err := os.OpenFile("/tmp/student-picker-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			debugLog = log.New(f, "", log.LstdFlags)
			if lf, lerr := tea.LogToFile("/tmp/student-picker-debug.log", "bubbletea"); lerr == nil {
				defer lf.Close()
			}
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
