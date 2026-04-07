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

func renderImageTermimg(imagePath string, x, y, w, h int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("termimg panicked: %v", r)
		}
	}()
	return DrawImageFile(imagePath, x, y, w, h)
}

func clearImagesTermimg() {
	defer func() { recover() }()
	ClearImages()
}

func cleanupTermimg() {
	defer func() { recover() }()
	CloseTerminal()
}
