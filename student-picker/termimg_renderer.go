package main

import (
	"image"

	"github.com/srlehn/termimg"
	_ "github.com/srlehn/termimg/drawers/all"
)

func renderImageTermimg(imagePath string, x, y, w, h int) error {
	bounds := image.Rect(x, y, x+w, y+h)
	return termimg.DrawFile(imagePath, bounds)
}

func clearImagesTermimg() {
	termimg.CleanUp()
}

func cleanupTermimg() {
	termimg.CleanUp()
}
