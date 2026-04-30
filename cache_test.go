package main

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestCache(t *testing.T, maxSize int) *Cache {
	t.Helper()
	dir := t.TempDir()
	return NewCache(filepath.Join(dir, "portrait"), maxSize)
}

func putDummy(t *testing.T, c *Cache, id int) {
	t.Helper()
	if err := c.Put(id, []byte("fake-image-data")); err != nil {
		t.Fatalf("Put(%d) error: %v", id, err)
	}
}

// --- NewCache ---

func TestNewCache_InitiallyEmpty(t *testing.T) {
	c := newTestCache(t, 5)
	if got := len(c.List()); got != 0 {
		t.Errorf("new cache should be empty, got %d items", got)
	}
}

// --- Put & Get ---

func TestCache_PutAndGet(t *testing.T) {
	c := newTestCache(t, 5)
	putDummy(t, c, 1001)

	img, ok := c.Get(1001)
	if !ok {
		t.Fatal("Get(1001) returned false, expected true")
	}
	if img.StudentId != 1001 {
		t.Errorf("StudentId = %d, want 1001", img.StudentId)
	}
	if _, err := os.Stat(img.PortraitPath); err != nil {
		t.Errorf("portrait file should exist: %v", err)
	}
}

func TestCache_GetMissing(t *testing.T) {
	c := newTestCache(t, 5)
	_, ok := c.Get(9999)
	if ok {
		t.Error("Get on missing key should return false")
	}
}

// --- LRU eviction ---

func TestCache_EvictsOldestWhenFull(t *testing.T) {
	c := newTestCache(t, 3)
	putDummy(t, c, 1)
	putDummy(t, c, 2)
	putDummy(t, c, 3)
	putDummy(t, c, 4) // should evict 1

	if _, ok := c.Get(1); ok {
		t.Error("student 1 should have been evicted")
	}
	for _, id := range []int{2, 3, 4} {
		if _, ok := c.Get(id); !ok {
			t.Errorf("student %d should still be cached", id)
		}
	}
}

func TestCache_GetRefreshesLRUOrder(t *testing.T) {
	c := newTestCache(t, 3)
	putDummy(t, c, 1)
	putDummy(t, c, 2)
	putDummy(t, c, 3)

	// Access 1 to make it recently used
	c.Get(1)

	// Adding 4 should evict 2 (oldest not recently used)
	putDummy(t, c, 4)

	if _, ok := c.Get(2); ok {
		t.Error("student 2 should have been evicted (LRU), not student 1")
	}
	if _, ok := c.Get(1); !ok {
		t.Error("student 1 should still be cached (was recently accessed)")
	}
}

// --- List ---

func TestCache_ListNewestFirst(t *testing.T) {
	c := newTestCache(t, 5)
	putDummy(t, c, 1)
	putDummy(t, c, 2)
	putDummy(t, c, 3)

	list := c.List()
	if len(list) != 3 {
		t.Fatalf("List() returned %d items, want 3", len(list))
	}
	// Newest first
	if list[0].StudentId != 3 {
		t.Errorf("list[0].StudentId = %d, want 3 (newest first)", list[0].StudentId)
	}
	if list[2].StudentId != 1 {
		t.Errorf("list[2].StudentId = %d, want 1 (oldest last)", list[2].StudentId)
	}
}

// --- Clear ---

func TestCache_Clear(t *testing.T) {
	c := newTestCache(t, 5)
	putDummy(t, c, 1)
	putDummy(t, c, 2)

	if err := c.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}
	if got := len(c.List()); got != 0 {
		t.Errorf("after Clear, List() returned %d items, want 0", got)
	}
	if _, ok := c.Get(1); ok {
		t.Error("Get after Clear should return false")
	}
}

// --- SetMaxSize ---

func TestCache_SetMaxSizeEvictsExcess(t *testing.T) {
	c := newTestCache(t, 5)
	putDummy(t, c, 1)
	putDummy(t, c, 2)
	putDummy(t, c, 3)
	putDummy(t, c, 4)
	putDummy(t, c, 5)

	c.SetMaxSize(2)

	list := c.List()
	if len(list) != 2 {
		t.Errorf("after SetMaxSize(2), List() returned %d items, want 2", len(list))
	}
	// Only the 2 newest should remain
	for _, img := range list {
		if img.StudentId != 4 && img.StudentId != 5 {
			t.Errorf("unexpected student %d in cache after shrink", img.StudentId)
		}
	}
}

func TestCache_SetMaxSizeClampedToOne(t *testing.T) {
	c := newTestCache(t, 5)
	c.SetMaxSize(0) // should clamp to 1
	if c.maxSize != 1 {
		t.Errorf("maxSize = %d after SetMaxSize(0), want 1", c.maxSize)
	}
}
