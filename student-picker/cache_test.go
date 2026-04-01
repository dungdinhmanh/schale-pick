package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 3)

	if cache.dir != dir {
		t.Errorf("expected dir %s, got %s", dir, cache.dir)
	}
	if cache.maxSize != 3 {
		t.Errorf("expected maxSize 3, got %d", cache.maxSize)
	}
	if cache.items == nil {
		t.Error("items map should not be nil")
	}
	if len(cache.order) != 0 {
		t.Error("order should be empty initially")
	}
}

func TestCachePutAndGet(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	iconData := []byte("fake-icon-data")
	portraitData := []byte("fake-portrait-data")

	err := cache.Put(1, iconData, portraitData)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	img, ok := cache.Get(1)
	if !ok {
		t.Fatal("expected to find student 1 in cache")
	}

	if img.StudentId != 1 {
		t.Errorf("expected StudentId 1, got %d", img.StudentId)
	}
	if img.IconPath == "" {
		t.Error("IconPath should not be empty")
	}
	if img.PortraitPath == "" {
		t.Error("PortraitPath should not be empty")
	}
}

func TestCacheGetNotFound(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	_, ok := cache.Get(999)
	if ok {
		t.Error("expected Get to return false for non-existent key")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 3)

	for i := 1; i <= 3; i++ {
		err := cache.Put(i, []byte("icon"), []byte("portrait"))
		if err != nil {
			t.Fatalf("Put(%d) failed: %v", i, err)
		}
	}

	err := cache.Put(4, []byte("icon"), []byte("portrait"))
	if err != nil {
		t.Fatalf("Put(4) failed: %v", err)
	}

	_, ok := cache.Get(1)
	if ok {
		t.Error("student 1 should have been evicted")
	}

	for _, id := range []int{2, 3, 4} {
		_, ok := cache.Get(id)
		if !ok {
			t.Errorf("student %d should exist in cache", id)
		}
	}
}

func TestCacheLRUAccessOrder(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 3)

	for i := 1; i <= 3; i++ {
		cache.Put(i, []byte("icon"), []byte("portrait"))
	}

	cache.Get(1)

	cache.Put(4, []byte("icon"), []byte("portrait"))

	_, ok := cache.Get(2)
	if ok {
		t.Error("student 2 should have been evicted")
	}

	_, ok = cache.Get(1)
	if !ok {
		t.Error("student 1 should still exist after accessing it")
	}
}

func TestCacheList(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	for i := 1; i <= 3; i++ {
		cache.Put(i, []byte("icon"), []byte("portrait"))
	}

	list := cache.List()

	if len(list) != 3 {
		t.Errorf("expected 3 items, got %d", len(list))
	}

	if list[0].StudentId != 3 {
		t.Errorf("expected first item to be 3 (newest), got %d", list[0].StudentId)
	}
	if list[2].StudentId != 1 {
		t.Errorf("expected last item to be 1 (oldest), got %d", list[2].StudentId)
	}
}

func TestCacheClear(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	for i := 1; i <= 3; i++ {
		cache.Put(i, []byte("icon"), []byte("portrait"))
	}

	err := cache.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if len(cache.items) != 0 {
		t.Error("items should be empty after Clear")
	}
	if len(cache.order) != 0 {
		t.Error("order should be empty after Clear")
	}

	list := cache.List()
	if len(list) != 0 {
		t.Error("List should return empty after Clear")
	}
}

func TestCacheFileCreation(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	iconData := []byte("test-icon")
	portraitData := []byte("test-portrait")

	err := cache.Put(100, iconData, portraitData)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	img, _ := cache.Get(100)

	if _, err := os.Stat(img.IconPath); os.IsNotExist(err) {
		t.Error("icon file should exist")
	}

	if _, err := os.Stat(img.PortraitPath); os.IsNotExist(err) {
		t.Error("portrait file should exist")
	}

	iconContent, err := os.ReadFile(img.IconPath)
	if err != nil {
		t.Fatalf("failed to read icon file: %v", err)
	}
	if string(iconContent) != "test-icon" {
		t.Errorf("icon content mismatch: got %s", iconContent)
	}

	portraitContent, err := os.ReadFile(img.PortraitPath)
	if err != nil {
		t.Fatalf("failed to read portrait file: %v", err)
	}
	if string(portraitContent) != "test-portrait" {
		t.Errorf("portrait content mismatch: got %s", portraitContent)
	}
}

func TestCacheFilePathFormat(t *testing.T) {
	dir := t.TempDir()
	cache := NewCache(dir, 10)

	cache.Put(42, []byte("icon"), []byte("portrait"))

	img, _ := cache.Get(42)

	expectedIcon := filepath.Join(dir, "icon", "42.webp")
	if img.IconPath != expectedIcon {
		t.Errorf("icon path: expected %s, got %s", expectedIcon, img.IconPath)
	}

	expectedPortrait := filepath.Join(dir, "portrait", "42.webp")
	if img.PortraitPath != expectedPortrait {
		t.Errorf("portrait path: expected %s, got %s", expectedPortrait, img.PortraitPath)
	}
}
