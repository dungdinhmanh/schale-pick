package picker

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
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
