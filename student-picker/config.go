package main

import (
	"os"
	"path"
)

// Schema URLs
const (
	SchemaURL       = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/refs/heads/main/data/en/students.json"
	IconBaseURL     = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/icon"
	PortraitBaseURL = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/portrait"
)

// Cache settings
const (
	DefaultCacheSize = 5
)

// Grid layout
const (
	GridCols = 5
	GridRows = 4
)

// Preview dimensions
const (
	PreviewW = 40
	PreviewH = 30
)

// CacheDir returns the cache directory path.
// Expands ~ to the user's home directory.
func CacheDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = os.Getenv("USERPROFILE")
	}
	if homeDir == "" {
		homeDir = "/tmp"
	}
	return path.Join(homeDir, ".cache", "student-picker")
}
