package picker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FetchManifest downloads the student manifest JSON from SchaleDB
// and caches it to the cache directory.
// Returns the raw JSON bytes on success.
func FetchManifest() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, SchemaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Ensure cache directory exists
	cacheDir := CacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Cache to file
	cachePath := filepath.Join(cacheDir, "students.json")
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write cache: %w", err)
	}

	return data, nil
}