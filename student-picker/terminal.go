package main

import (
	"os"
	"strings"
)

func detectTerminalType() string {
	termProgram := os.Getenv("TERM_PROGRAM")

	switch strings.ToLower(termProgram) {
	case "kitty":
		return "kitty"
	case "wezterm":
		return "sixel"
	case "iterm.app", "iterm2":
		return "iterm"
	case "mintty":
		return "sixel"
	}

	term := os.Getenv("TERM")
	if strings.Contains(term, "kitty") {
		return "kitty"
	}
	if strings.Contains(term, "xterm") {
		return "sixel"
	}

	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "kitty"
	}

	return "kitty-direct"
}

func getFastfetchImageType(termType string) string {
	switch termType {
	case "kitty":
		return "kitty"
	case "sixel":
		return "sixel"
	case "iterm":
		return "iterm"
	default:
		return "kitty-direct"
	}
}
