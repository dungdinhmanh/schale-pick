package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/srlehn/termimg"
	_ "github.com/srlehn/termimg/drawers/all"
	_ "github.com/srlehn/termimg/terminals"
	"github.com/srlehn/termimg/term"
	_ "golang.org/x/image/webp"
)

type imgJob struct {
	url  string
	rect image.Rectangle
}

var mu sync.Mutex

func fetchAndDraw(tm *term.Terminal, job imgJob, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(job.url)
	if err != nil {
		log.Println("fetch error:", err)
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("decode error:", err)
		return
	}

	timg := termimg.NewImage(img)
	mu.Lock()
	if err := tm.Draw(timg, job.rect); err != nil {
		log.Println("draw error:", err)
	}
	mu.Unlock()
}

func main() {
	tm, err := termimg.Terminal()
	if err != nil {
		log.Fatal(err)
	}
	defer tm.Close()

	jobs := []imgJob{
		{"https://raw.githubusercontent.com/SchaleDB/SchaleDB/refs/heads/main/images/student/icon/10000.webp", image.Rect(0,  0,  20, 12)},
		{"https://raw.githubusercontent.com/SchaleDB/SchaleDB/refs/heads/main/images/student/icon/10002.webp", image.Rect(20, 0,  40, 12)},
		{"https://raw.githubusercontent.com/SchaleDB/SchaleDB/refs/heads/main/images/student/icon/10003.webp", image.Rect(40, 0,  60, 12)},
	}

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go fetchAndDraw(tm, job, &wg)
	}
	wg.Wait()

	time.Sleep(3 * time.Second)
}
