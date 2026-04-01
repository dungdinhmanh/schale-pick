package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Downloader handles concurrent image downloads with semaphore
type Downloader struct {
	client    *http.Client
	semaphore chan struct{} // max concurrent downloads
}

// NewDownloader creates a new Downloader with maxConcurrent limit
func NewDownloader(maxConcurrent int) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// DownloadImage downloads an image from url and returns bytes
func (d *Downloader) DownloadImage(url string) ([]byte, error) {
	// Acquire semaphore
	d.semaphore <- struct{}{}
	defer func() { <-d.semaphore }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// DownloadPortrait downloads a portrait image for a student
func (d *Downloader) DownloadPortrait(id int) ([]byte, error) {
	return d.DownloadImage(GetStudentPortraitURL(id))
}

// DownloadIcon downloads an icon image for a student
func (d *Downloader) DownloadIcon(id int) ([]byte, error) {
	return d.DownloadImage(GetStudentIconURL(id))
}
