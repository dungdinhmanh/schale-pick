package main

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// CachedImage represents a cached student image
type CachedImage struct {
	StudentId    int
	PortraitPath string
	CachedAt     time.Time
}

// Cache is an LRU cache for student images
type Cache struct {
	dir     string
	maxSize int
	items   map[int]CachedImage // key = studentId
	order   []int               // ordered list (newest last)
	mu      sync.Mutex
}

// NewCache creates a new Cache with the specified directory and max size
func NewCache(dir string, maxSize int) *Cache {
	return &Cache{
		dir:     dir,
		maxSize: maxSize,
		items:   make(map[int]CachedImage),
		order:   []int{},
	}
}

// Get returns the cached image for a student id, or false if not found
func (c *Cache) Get(studentId int) (CachedImage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	img, ok := c.items[studentId]
	if !ok {
		return CachedImage{}, false
	}

	// Move to end (most recently used)
	c.moveToEnd(studentId)
	return img, true
}

// Put adds an image to the cache, evicting oldest if necessary
func (c *Cache) Put(studentId int, portraitData []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}

	// Write portrait
	portraitPath := filepath.Join(c.dir, "portrait", strconv.Itoa(studentId)+".webp")
	if err := os.MkdirAll(filepath.Dir(portraitPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(portraitPath, portraitData, 0644); err != nil {
		return err
	}

	// Add to cache
	c.items[studentId] = CachedImage{
		StudentId:    studentId,
		PortraitPath: portraitPath,
		CachedAt:     time.Now(),
	}
	c.order = append(c.order, studentId)

	// Evict if over max size
	if len(c.order) > c.maxSize {
		c.evictOldest()
	}

	return nil
}

// List returns all cached images in LRU order (newest first)
func (c *Cache) List() []CachedImage {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]CachedImage, 0, len(c.items))
	// Iterate in reverse order (newest first)
	for i := len(c.order) - 1; i >= 0; i-- {
		if img, ok := c.items[c.order[i]]; ok {
			result = append(result, img)
		}
	}
	return result
}

// Clear removes all cached files
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.items {
		os.Remove(img.PortraitPath)
	}
	c.items = make(map[int]CachedImage)
	c.order = []int{}
	return nil
}

// SetMaxSize updates the max cache size and evicts oldest entries if needed
func (c *Cache) SetMaxSize(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	c.maxSize = n
	for len(c.order) > c.maxSize {
		c.evictOldest()
	}
}

func (c *Cache) moveToEnd(studentId int) {
	for i, id := range c.order {
		if id == studentId {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, studentId)
			return
		}
	}
}

func (c *Cache) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	oldest := c.order[0]
	c.order = c.order[1:]
	if img, ok := c.items[oldest]; ok {
		os.Remove(img.PortraitPath)
	}
	delete(c.items, oldest)
}
