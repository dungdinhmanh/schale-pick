package main

import (
	"fmt"
	"image"

	"github.com/srlehn/termimg"
	_ "github.com/srlehn/termimg/drawers/all"
	_ "github.com/srlehn/termimg/terminals"
)

func renderImageTermimg(imagePath string, x, y, w, h int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("termimg paniced: %v", r)
		}
	}()
	bounds := image.Rect(x, y, x+w, y+h)
	err = termimg.DrawFile(imagePath, bounds)
	return err
}

func clearImagesTermimg() {
	defer func() {
		recover()
	}()
	termimg.CleanUp()
}

func cleanupTermimg() {
	defer func() {
		recover()
	}()
	termimg.CleanUp()
}
