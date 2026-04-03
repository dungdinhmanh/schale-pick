package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// SaveMeta saves a lightweight mapping of ID to Name for offline use
func SaveMeta(students []Student) error {
	meta := make(map[int]string)
	for _, s := range students {
		meta[s.Id] = s.FamilyName + " " + s.PersonalName
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(CacheDir(), "meta.json")
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, data, 0644)
}

// LoadMeta loads the lightweight mapping from cache
func LoadMeta() ([]Student, error) {
	path := filepath.Join(CacheDir(), "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta map[int]string
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	students := make([]Student, 0, len(meta))
	for id, name := range meta {
		names := strings.SplitN(name, " ", 2)
		fam, pers := "", ""
		if len(names) > 0 { fam = names[0] }
		if len(names) > 1 { pers = names[1] }
		students = append(students, Student{Id: id, FamilyName: fam, PersonalName: pers})
	}
	return students, nil
}
