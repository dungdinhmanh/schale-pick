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

var version = "0.1.0" // overridden at build time via -ldflags "-X main.version=vX.Y.Z"

// debugLog discards all output by default.
// Reassigned to a real logger when --debug is passed.
var debugLog = log.New(io.Discard, "", log.LstdFlags)

func main() {
	debug      := flag.Bool("debug",       false, "write debug log to /tmp/student-picker-debug.log")
	showVer    := flag.Bool("version",     false, "print version and exit")
	clearCache := flag.Bool("clear-cache", false, "delete all cached portraits and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "student-picker %s\n", version)
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "Browse Blue Archive student portraits and set one as your fastfetch logo.")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "USAGE:")
		fmt.Fprintln(os.Stdout, "  student-picker [flags]")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "FLAGS:")
		fmt.Fprintln(os.Stdout, "  --version       print version and exit")
		fmt.Fprintln(os.Stdout, "  --clear-cache   delete all cached portraits and exit")
		fmt.Fprintln(os.Stdout, "  --debug         write debug log to /tmp/student-picker-debug.log")
		fmt.Fprintln(os.Stdout, "  --help          show this help")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "KEYS (inside the TUI):")
		fmt.Fprintln(os.Stdout, "  arrows / hjkl   navigate grid")
		fmt.Fprintln(os.Stdout, "  Tab             switch Browse / Installed tabs")
		fmt.Fprintln(os.Stdout, "  /               search")
		fmt.Fprintln(os.Stdout, "  Enter           select student → patch fastfetch config")
		fmt.Fprintln(os.Stdout, "  i               settings")
		fmt.Fprintln(os.Stdout, "  ?               help")
		fmt.Fprintln(os.Stdout, "  q / Ctrl+C      quit")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintf(os.Stdout,  "CACHE DIR:\n  %s\n", CacheDir())
	}

	flag.Parse()

	if *showVer {
		fmt.Printf("student-picker %s\n", version)
		os.Exit(0)
	}

	if *clearCache {
		cacheDir := CacheDir()
		if err := os.RemoveAll(cacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "clear-cache: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Cache cleared: %s\n", cacheDir)
		os.Exit(0)
	}

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
