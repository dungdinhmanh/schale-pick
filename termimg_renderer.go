package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"

	_ "image/jpeg"

	"github.com/srlehn/termimg"
	_ "github.com/srlehn/termimg/drawers/all"
	"github.com/srlehn/termimg/term"
	_ "github.com/srlehn/termimg/terminals"
	_ "golang.org/x/image/webp"
)

// Kitty replaces placements with same image ID, eliminating flicker on re-render.
// Use distinct IDs per region to ensure replacement, not layering.
const (
	previewImageID = 100
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

func drawImageWithID(img image.Image, x, y, w, h, imageID int) error {
	tm, err := GetTerminal()
	if err != nil {
		return err
	}

	cellW, cellH, err := tm.CellSize()
	if err != nil || cellW <= 0 || cellH <= 0 {
		return DrawImage(img, x, y, w, h)
	}

	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return DrawImage(img, x, y, w, h)
	}

	wPx := float64(w) * cellW
	hPx := float64(h) * cellH
	imgAspect := float64(imgW) / float64(imgH)
	targetAspect := wPx / hPx

	var newWpx, newHpx float64
	if imgAspect > targetAspect {
		newWpx = wPx
		newHpx = wPx / imgAspect
	} else {
		newHpx = hPx
		newWpx = hPx * imgAspect
	}

	newW := int((newWpx + cellW/2) / cellW)
	newH := int((newHpx + cellH/2) / cellH)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	// Clamp so image never exceeds requested cell box (prevents bleed into next row)
	if newW > w {
		newW = w
	}
	if newH > h {
		newH = h
	}

	xOff := (w - newW) / 2
	yOff := (h - newH) / 2

	bytBuf := new(bytes.Buffer)
	if err := png.Encode(bytBuf, img); err != nil {
		return err
	}
	imgBase64 := base64.StdEncoding.EncodeToString(bytBuf.Bytes())

	posX := x + xOff + 1
	posY := y + yOff + 1

	const kittyLimit = 4096
	const zIndex = 2

	var kittyStr string
	settings := fmt.Sprintf("a=T,t=d,f=100,i=%d,X=%d,Y=%d,c=%d,r=%d,z=%d,", imageID, posX, posY, newW, newH, zIndex)

	i := 0
	for ; i < (len(imgBase64)-1)/kittyLimit; i++ {
		kittyStr += fmt.Sprintf("\x1b_G%sC=1,m=1;%s\x1b\\", settings, imgBase64[i*kittyLimit:(i+1)*kittyLimit])
		settings = ""
	}
	kittyStr += fmt.Sprintf("\x1b_G%sC=1,m=0;%s\x1b\\", settings, imgBase64[i*kittyLimit:])
	kittyStr = fmt.Sprintf("\x1b[%d;%dH%s", posY, posX, kittyStr)

	drawMu.Lock()
	defer drawMu.Unlock()
	// Delete previous placement of this image ID before drawing the new one
	_, _ = tm.Printf("\x1b_Ga=d,d=i,i=%d;\x1b\\", imageID)
	_, err = tm.Printf("%s", kittyStr)
	return err
}

func drawPreviewImage(imagePath string, x, y, w, h int) error {
	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	return drawImageWithID(img, x, y, w, h, previewImageID)
}

func clearImagesTermimg() {
	defer func() { recover() }()
	if terminal == nil {
		return
	}
	drawMu.Lock()
	defer drawMu.Unlock()
	// Delete the preview image placement using the same syntax that works in drawImageWithID.
	// Sending generic "a=d" without an explicit d=A is interpreted inconsistently across
	// terminals; targeting our specific image ID is reliable.
	_, _ = terminal.Printf("\x1b_Ga=d,d=i,i=%d;\x1b\\", previewImageID)
	// Also issue a delete-all (uppercase A frees data too) as a belt-and-suspenders.
	_, _ = terminal.Printf("\x1b_Ga=d,d=A;\x1b\\")
}

func cleanupTermimg() {
	defer func() { recover() }()
	CloseTerminal()
}
