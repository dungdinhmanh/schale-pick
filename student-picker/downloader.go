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

// metaEntry is the persistent format for offline meta storage.
type metaEntry struct {
	Name         string `json:"n"`
	FamilyName   string `json:"f"`
	PersonalName string `json:"p"`
}

// SaveMeta saves student metadata for offline use, preserving the full Name field.
func SaveMeta(students []Student) error {
	meta := make(map[int]metaEntry, len(students))
	for _, s := range students {
		meta[s.Id] = metaEntry{
			Name:         s.Name,
			FamilyName:   s.FamilyName,
			PersonalName: s.PersonalName,
		}
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(CacheDir(), "meta.json")
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, data, 0644)
}

// LoadMeta loads the lightweight mapping from cache.
// Supports both new struct format (map[int]metaEntry) and old string format (map[int]string).
func LoadMeta() ([]Student, error) {
	path := filepath.Join(CacheDir(), "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try new struct format first
	var metaNew map[int]metaEntry
	if err := json.Unmarshal(data, &metaNew); err == nil && len(metaNew) > 0 {
		// Validate it's actually the new format (has at least one non-empty Name or FamilyName)
		isNew := false
		for _, e := range metaNew {
			if e.Name != "" || e.FamilyName != "" {
				isNew = true
				break
			}
		}
		if isNew {
			students := make([]Student, 0, len(metaNew))
			for id, e := range metaNew {
				s := Student{
					Id:           id,
					Name:         e.Name,
					FamilyName:   e.FamilyName,
					PersonalName: e.PersonalName,
				}
				computeVariant(&s)
				students = append(students, s)
			}
			return students, nil
		}
	}

	// Backwards compat: old format map[int]string "FamilyName PersonalName"
	var metaOld map[int]string
	if err := json.Unmarshal(data, &metaOld); err != nil {
		return nil, err
	}
	students := make([]Student, 0, len(metaOld))
	for id, name := range metaOld {
		parts := strings.SplitN(name, " ", 2)
		fam, pers := "", name
		if len(parts) == 2 {
			fam = parts[0]
			pers = parts[1]
		}
		s := Student{Id: id, FamilyName: fam, PersonalName: pers}
		computeVariant(&s)
		students = append(students, s)
	}
	return students, nil
}
