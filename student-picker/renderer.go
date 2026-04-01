package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func getIconPath(studentId int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("sp-icon-%d.webp", studentId))
}

func getPngIconPath(studentId int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("sp-icon-%d.png", studentId))
}

func ensureIconCached(studentId int, downloader *Downloader) string {
	iconPath := getIconPath(studentId)
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		if downloader != nil {
			data, err := downloader.DownloadIcon(studentId)
			if err == nil {
				os.WriteFile(iconPath, data, 0644)
			}
		}
	}
	return iconPath
}

func convertWebpToPng(webpPath, pngPath string) error {
	cmd := exec.Command("magick", webpPath, pngPath)
	return cmd.Run()
}

func ensurePngIcon(studentId int, downloader *Downloader, hasMagick bool) string {
	pngPath := getPngIconPath(studentId)
	if _, err := os.Stat(pngPath); err == nil {
		return pngPath
	}

	webpPath := ensureIconCached(studentId, downloader)
	if hasMagick {
		if err := convertWebpToPng(webpPath, pngPath); err == nil {
			return pngPath
		}
	}
	return webpPath
}

func getPortraitPath(studentId int) string {
	return filepath.Join(CacheDir(), "portrait", fmt.Sprintf("%d.webp", studentId))
}

func ensurePortraitCached(studentId int, downloader *Downloader) string {
	portraitPath := getPortraitPath(studentId)
	if _, err := os.Stat(portraitPath); os.IsNotExist(err) {
		if downloader != nil {
			data, err := downloader.DownloadPortrait(studentId)
			if err == nil {
				os.MkdirAll(filepath.Dir(portraitPath), 0755)
				os.WriteFile(portraitPath, data, 0644)
				return portraitPath
			}
			return ""
		}
	}
	return portraitPath
}
