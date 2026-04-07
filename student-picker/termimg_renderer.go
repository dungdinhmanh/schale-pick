package main

import (
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/srlehn/termimg"
	_ "github.com/srlehn/termimg/drawers/all"
	"github.com/srlehn/termimg/term"
	_ "github.com/srlehn/termimg/terminals"
	_ "golang.org/x/image/webp"
	_ "image/jpeg"
	_ "image/png"
)

var (
	terminal     *term.Terminal
	terminalMu   sync.Mutex
	terminalOnce sync.Once
	terminalErr  error
	drawMu       sync.Mutex
)

func InitTerminal() error {
	terminalOnce.Do(func() {
		terminal, terminalErr = termimg.Terminal()
	})
	return terminalErr
}

func GetTerminal() (*term.Terminal, error) {
	if terminal == nil {
		return nil, fmt.Errorf("terminal not initialized")
	}
	return terminal, nil
}

func CloseTerminal() {
	if terminal != nil {
		terminal.Close()
		terminal = nil
	}
}

func DrawImage(img image.Image, x, y, w, h int) error {
	tm, err := GetTerminal()
	if err != nil {
		return err
	}

	timg := termimg.NewImage(img)
	rect := image.Rect(x, y, x+w, y+h)

	drawMu.Lock()
	defer drawMu.Unlock()
	return tm.Draw(timg, rect)
}

func DrawImageFile(imagePath string, x, y, w, h int) error {
	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	return DrawImage(img, x, y, w, h)
}

func ClearImages() {
	drawMu.Lock()
	defer drawMu.Unlock()

	if terminal != nil {
		terminalMu.Lock()
		termimg.CleanUp()
		terminalMu.Unlock()
	}
}

func renderImageTermimg(imagePath string, x, y, w, h int, preserveAspectRatio bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("termimg panicked: %v", r)
		}
	}()

	if !preserveAspectRatio {
		return DrawImageFile(imagePath, x, y, w, h)
	}

	// For aspect ratio preservation, we need to consider cell dimensions
	// (cells are rectangular: ~9px wide × ~18px tall, not square!)
	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return DrawImage(img, x, y, w, h)
	}

	// Get cell dimensions to properly calculate aspect ratios
	tm, err := GetTerminal()
	if err != nil {
		// Fallback: let termimg handle it directly
		return DrawImage(img, x, y, w, h)
	}

	cellW, cellH, err := tm.CellSize()
	if err != nil || cellW <= 0 || cellH <= 0 {
		// Fallback: let termimg handle it directly
		return DrawImage(img, x, y, w, h)
	}

	// Convert cell dimensions to pixel dimensions
	wPx := float64(w) * cellW
	hPx := float64(h) * cellH

	// Calculate aspect ratios in pixels
	imgAspect := float64(imgW) / float64(imgH)
	targetAspect := wPx / hPx

	// Calculate fitted dimensions in pixels
	var newWpx, newHpx float64
	if imgAspect > targetAspect {
		// Image is wider than target - fit by width
		newWpx = wPx
		newHpx = wPx / imgAspect
	} else {
		// Image is taller than target - fit by height
		newHpx = hPx
		newWpx = hPx * imgAspect
	}

	// Convert fitted pixel dimensions back to cells
	// Use ceiling to ensure we don't lose precision
	newW := int((newWpx + cellW/2) / cellW) // Round to nearest cell
	newH := int((newHpx + cellH/2) / cellH)

	// Ensure at least 1 cell
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	// Calculate centering offsets in cells
	xOff := (w - newW) / 2
	yOff := (h - newH) / 2

	return DrawImage(img, x+xOff, y+yOff, newW, newH)
}

func clearImagesTermimg() {
	defer func() { recover() }()
	ClearImages()
}

func cleanupTermimg() {
	defer func() { recover() }()
	CloseTerminal()
}
