package main

import (
    "image"
    "log"
    "time"

    "github.com/srlehn/termimg"
    _ "github.com/srlehn/termimg/drawers/all"
    _ "github.com/srlehn/termimg/terminals"
		_ "golang.org/x/image/webp"
)

func main() {
    tm, err := termimg.Terminal()
    if err != nil {
        log.Fatal(err)
    }
    defer tm.Close()

    timg := termimg.NewImageFileName(`10081.webp`)
    err = tm.Draw(timg, image.Rect(10, 10, 40, 25))
    if err != nil {
        log.Fatal(err)
    }

    time.Sleep(2 * time.Second)
}
