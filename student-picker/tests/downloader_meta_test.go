package picker_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"student-picker/internal/picker"
)

func withTempCacheDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	old := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	return dir, func() {
		os.Setenv("HOME", old)
	}
}

func TestSaveLoadMetaRoundtrip(t *testing.T) {
	_, cleanup := withTempCacheDir(t)
	defer cleanup()

	students := []picker.Student{
		{Id: 10074, Name: "Hanako (Swimsuit)", FamilyName: "Ichinose", PersonalName: "Hanako",
			Base: "Hanako", Variant: "Swimsuit", Abbr: "SW"},
		{Id: 23007, Name: "Hanako", FamilyName: "Ichinose", PersonalName: "Hanako"},
	}

	if err := picker.SaveMeta(students); err != nil {
		t.Fatalf("SaveMeta: %v", err)
	}

	loaded, err := picker.LoadMeta()
	if err != nil {
		t.Fatalf("LoadMeta: %v", err)
	}

	byID := map[int]picker.Student{}
	for _, s := range loaded {
		byID[s.Id] = s
	}

	h := byID[10074]
	if h.Name != "Hanako (Swimsuit)" {
		t.Errorf("Name = %q, want 'Hanako (Swimsuit)'", h.Name)
	}
	if h.Variant != "Swimsuit" {
		t.Errorf("Variant = %q, want 'Swimsuit'", h.Variant)
	}
	if h.Abbr != "SW" {
		t.Errorf("Abbr = %q, want 'SW'", h.Abbr)
	}

	n := byID[23007]
	if n.Name != "Hanako" {
		t.Errorf("Name = %q, want 'Hanako'", n.Name)
	}
}

func TestLoadMetaBackwardCompatOldFormat(t *testing.T) {
	_, cleanup := withTempCacheDir(t)
	defer cleanup()

	rawOld := `{"10074":"Ichinose Hanako","23007":"Ichinose Hanako"}`
	cacheP := filepath.Join(picker.CacheDir(), "meta.json")
	os.MkdirAll(filepath.Dir(cacheP), 0755)
	if err := os.WriteFile(cacheP, []byte(rawOld), 0644); err != nil {
		t.Fatalf("write old meta: %v", err)
	}

	loaded, err := picker.LoadMeta()
	if err != nil {
		t.Fatalf("LoadMeta (old format): %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 students from old format, got %d", len(loaded))
	}
	for _, s := range loaded {
		if s.PersonalName == "" && s.FamilyName == "" {
			t.Errorf("student id=%d has empty names", s.Id)
		}
	}
}

func TestLoadMetaNewStructFormat(t *testing.T) {
	_, cleanup := withTempCacheDir(t)
	defer cleanup()

	type entry struct {
		N string `json:"n"`
		F string `json:"f"`
		P string `json:"p"`
	}
	raw := map[string]entry{
		"10074": {N: "Hanako (Swimsuit)", F: "Ichinose", P: "Hanako"},
	}
	data, _ := json.Marshal(raw)
	cacheP := filepath.Join(picker.CacheDir(), "meta.json")
	os.MkdirAll(filepath.Dir(cacheP), 0755)
	os.WriteFile(cacheP, data, 0644)

	loaded, err := picker.LoadMeta()
	if err != nil {
		t.Fatalf("LoadMeta (new format): %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 student, got %d", len(loaded))
	}
	s := loaded[0]
	if s.Name != "Hanako (Swimsuit)" {
		t.Errorf("Name = %q, want 'Hanako (Swimsuit)'", s.Name)
	}
	if s.Abbr != "SW" {
		t.Errorf("Abbr after LoadMeta = %q, want 'SW'", s.Abbr)
	}
}
