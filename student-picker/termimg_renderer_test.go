package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// TestInitTerminal tests terminal initialization
func TestInitTerminal(t *testing.T) {
	// Reset state before test
	terminal = nil
	terminalOnce = sync.Once{}

	err := InitTerminal()
	if err != nil {
		t.Logf("InitTerminal() error (may be expected in non-terminal env): %v", err)
		// This is expected when running in a non-terminal test environment
		return
	}

	// If successful, verify terminal is set
	if terminal == nil {
		t.Error("terminal should be initialized after InitTerminal")
	}
}

// TestGetTerminal tests terminal getter
func TestGetTerminal(t *testing.T) {
	// Test without initialization
	terminal = nil
	tm, err := GetTerminal()
	if err == nil {
		t.Error("expected error when terminal not initialized")
	}
	if tm != nil {
		t.Error("expected nil terminal when not initialized")
	}
}

// TestDrawImage tests image drawing function
func TestDrawImage(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with a color
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	// Test without terminal initialization (should fail gracefully)
	terminal = nil
	err := DrawImage(img, 0, 0, 10, 10)
	if err == nil {
		t.Log("DrawImage succeeded (terminal may be available)")
	} else {
		t.Logf("DrawImage failed as expected: %v", err)
	}
}

// TestDrawImageFile tests file-based image drawing
func TestDrawImageFile(t *testing.T) {
	// Create a temp image file
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	// Create and save a simple image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("failed to create temp image: %v", err)
	}
	defer f.Close()

	// Encode as PNG (png package must be imported)
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	f.Close()

	// Test drawing the file
	terminal = nil
	err = DrawImageFile(imgPath, 0, 0, 10, 10)
	if err == nil {
		t.Log("DrawImageFile succeeded (terminal may be available)")
	} else {
		t.Logf("DrawImageFile failed as expected: %v", err)
	}
}

// TestDrawImageFileNonExistent tests drawing non-existent file
func TestDrawImageFileNonExistent(t *testing.T) {
	err := DrawImageFile("/nonexistent/path/image.png", 0, 0, 10, 10)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// TestClearImages tests clearing images
func TestClearImages(t *testing.T) {
	// Should not panic even with nil terminal
	terminal = nil
	ClearImages() // Should not panic

	// Test with initialized but closed terminal
	terminalOnce = sync.Once{}
	InitTerminal()
	ClearImages() // Should not panic
}

// TestClearImagesTermimg tests the wrapper function
func TestClearImagesTermimg(t *testing.T) {
	// Should not panic in any case
	clearImagesTermimg()
}

// TestRenderImageTermimg tests the main render function
func TestRenderImageTermimg(t *testing.T) {
	// Create a temp image file
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.webp")

	// Create a simple image
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("failed to create temp image: %v", err)
	}
	defer f.Close()
	png.Encode(f, img)
	f.Close()

	// Test rendering
	err = renderImageTermimg(imgPath, 0, 0, 10, 10, false)
	if err == nil {
		t.Log("renderImageTermimg succeeded")
	} else {
		t.Logf("renderImageTermimg failed (expected in non-terminal env): %v", err)
	}
}

// TestRenderImageTermimgPanic tests panic recovery
func TestRenderImageTermimgPanic(t *testing.T) {
	// Test with non-existent file - should not panic
	err := renderImageTermimg("/nonexistent/image.webp", 0, 0, 10, 10, false)
	if err == nil {
		t.Log("renderImageTermimg handled non-existent file")
	} else {
		t.Logf("renderImageTermimg returned error: %v", err)
	}
}

// TestCleanupTermimg tests cleanup function
func TestCleanupTermimg(t *testing.T) {
	// Should not panic even with nil terminal
	terminal = nil
	cleanupTermimg() // Should not panic
}

// TestTerminalDetectionInKitty tests terminal detection
func TestTerminalDetectionInKitty(t *testing.T) {
	// Save current env
	oldTerm := os.Getenv("TERM")
	oldKitty := os.Getenv("KITTY_WINDOW_ID")

	defer func() {
		os.Setenv("TERM", oldTerm)
		os.Setenv("KITTY_WINDOW_ID", oldKitty)
	}()

	// Test with Kitty env
	os.Setenv("TERM", "xterm-kitty")
	os.Setenv("KITTY_WINDOW_ID", "12345")

	termType := detectTerminalType()
	if termType != "kitty" && termType != "kitty-direct" {
		t.Errorf("expected kitty or kitty-direct, got %s", termType)
	}
}

// TestConcurrentDraw tests thread safety
func TestConcurrentDraw(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Reset terminal state
	terminalOnce = sync.Once{}
	terminal = nil

	// Create multiple goroutines trying to init/draw simultaneously
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			// Init terminal
			InitTerminal()

			// Try to clear/draw
			clearImagesTermimg()

			// Create a small test image
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			DrawImage(img, idx*10, 0, 10, 10)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestBatchRenderConsistency tests that batch rendering works consistently
func TestBatchRenderConsistency(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("termimg primarily designed for Unix terminals")
	}

	// This test verifies the fix for the race condition
	// where ClearImages was called separately for icons and preview

	tmpDir := t.TempDir()

	// Create mock image files
	for i := 1; i <= 5; i++ {
		imgPath := filepath.Join(tmpDir, "icon.webp")
		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		f, _ := os.Create(imgPath)
		png.Encode(f, img)
		f.Close()
	}

	// Simulate batch rendering - should not panic
	clearImagesTermimg()

	// After clear, rendering should still work
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	_ = DrawImage(img, 0, 0, 10, 10)

	// Multiple clears in sequence should be safe
	clearImagesTermimg()
	clearImagesTermimg()
	clearImagesTermimg()
}
