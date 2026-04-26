package picker_test

import (
	"testing"

	"student-picker/internal/picker"
)

// --- MakeAbbr ---

func TestMakeAbbrEmpty(t *testing.T) {
	if got := picker.MakeAbbr(""); got != "" {
		t.Errorf("MakeAbbr('') = %q, want ''", got)
	}
}

func TestMakeAbbrSingleWord(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Swimsuit", "SW"},
		{"Bunny", "BU"},
		{"X", "X"},
		{"ab", "AB"},
	}
	for _, c := range cases {
		got := picker.MakeAbbr(c.in)
		if got != c.want {
			t.Errorf("MakeAbbr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMakeAbbrMultiWord(t *testing.T) {
	cases := []struct{ in, want string }{
		{"New Year", "NY"},
		{"Hot Spring", "HS"},
		{"New Year Party", "NYP"},
	}
	for _, c := range cases {
		got := picker.MakeAbbr(c.in)
		if got != c.want {
			t.Errorf("MakeAbbr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- ComputeVariant ---

func TestComputeVariantNoParens(t *testing.T) {
	s := picker.Student{Name: "Hanako", PersonalName: "Hanako"}
	picker.ComputeVariant(&s)
	if s.Base != "Hanako" {
		t.Errorf("Base = %q, want 'Hanako'", s.Base)
	}
	if s.Variant != "" {
		t.Errorf("Variant = %q, want ''", s.Variant)
	}
	if s.Abbr != "" {
		t.Errorf("Abbr = %q, want ''", s.Abbr)
	}
}

func TestComputeVariantSingleWord(t *testing.T) {
	s := picker.Student{Name: "Hanako (Swimsuit)", PersonalName: "Hanako"}
	picker.ComputeVariant(&s)
	if s.Base != "Hanako" {
		t.Errorf("Base = %q, want 'Hanako'", s.Base)
	}
	if s.Variant != "Swimsuit" {
		t.Errorf("Variant = %q, want 'Swimsuit'", s.Variant)
	}
	if s.Abbr != "SW" {
		t.Errorf("Abbr = %q, want 'SW'", s.Abbr)
	}
}

func TestComputeVariantMultiWord(t *testing.T) {
	s := picker.Student{Name: "Hina (New Year)", PersonalName: "Hina"}
	picker.ComputeVariant(&s)
	if s.Base != "Hina" {
		t.Errorf("Base = %q, want 'Hina'", s.Base)
	}
	if s.Variant != "New Year" {
		t.Errorf("Variant = %q, want 'New Year'", s.Variant)
	}
	if s.Abbr != "NY" {
		t.Errorf("Abbr = %q, want 'NY'", s.Abbr)
	}
}

func TestComputeVariantEmptyNameFallsBackToPersonal(t *testing.T) {
	s := picker.Student{Name: "", PersonalName: "Hanako", FamilyName: "Ichinose"}
	picker.ComputeVariant(&s)
	if s.Name != "Hanako" {
		t.Errorf("Name should fall back to PersonalName, got %q", s.Name)
	}
	if s.Variant != "" {
		t.Errorf("Variant should be empty, got %q", s.Variant)
	}
}

// --- DisplayName ---

func TestDisplayNameFitsFullName(t *testing.T) {
	s := picker.Student{Name: "Hina (Swimsuit)", Base: "Hina", Variant: "Swimsuit", Abbr: "SW"}
	name, abbrUsed := picker.DisplayName(s, 20)
	if name != "Hina (Swimsuit)" {
		t.Errorf("DisplayName = %q, want 'Hina (Swimsuit)'", name)
	}
	if abbrUsed {
		t.Error("abbrUsed should be false when full name fits")
	}
}

func TestDisplayNameUsesAbbrWhenTight(t *testing.T) {
	s := picker.Student{Name: "Hina (Swimsuit)", Base: "Hina", Variant: "Swimsuit", Abbr: "SW"}
	name, abbrUsed := picker.DisplayName(s, 10)
	if name != "Hina (SW)" {
		t.Errorf("DisplayName = %q, want 'Hina (SW)'", name)
	}
	if !abbrUsed {
		t.Error("abbrUsed should be true")
	}
}

func TestDisplayNameTruncatesWhenNoAbbr(t *testing.T) {
	s := picker.Student{Name: "VeryLongNameHere", Base: "VeryLongNameHere"}
	name, _ := picker.DisplayName(s, 8)
	if picker.RuneLen(name) > 8 {
		t.Errorf("truncated name %q is wider than 8", name)
	}
}

func TestDisplayNameNoVariantNoTruncNeeded(t *testing.T) {
	s := picker.Student{Name: "Hanako", Base: "Hanako"}
	name, abbrUsed := picker.DisplayName(s, 10)
	if name != "Hanako" {
		t.Errorf("DisplayName = %q, want 'Hanako'", name)
	}
	if abbrUsed {
		t.Error("abbrUsed should be false")
	}
}

// --- RuneLen ---

func TestRuneLen(t *testing.T) {
	if picker.RuneLen("abc") != 3 {
		t.Error("ascii RuneLen wrong")
	}
	if picker.RuneLen("こんにちは") != 5 {
		t.Error("cjk RuneLen wrong")
	}
}
