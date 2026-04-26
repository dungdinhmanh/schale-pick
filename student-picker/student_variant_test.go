package main

import "testing"

// --- makeAbbr ---

func TestMakeAbbrEmpty(t *testing.T) {
	if got := makeAbbr(""); got != "" {
		t.Errorf("makeAbbr('') = %q, want ''", got)
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
		got := makeAbbr(c.in)
		if got != c.want {
			t.Errorf("makeAbbr(%q) = %q, want %q", c.in, got, c.want)
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
		got := makeAbbr(c.in)
		if got != c.want {
			t.Errorf("makeAbbr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- computeVariant ---

func TestComputeVariantNoParens(t *testing.T) {
	s := Student{Name: "Hanako", PersonalName: "Hanako"}
	computeVariant(&s)
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
	s := Student{Name: "Hanako (Swimsuit)", PersonalName: "Hanako"}
	computeVariant(&s)
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
	s := Student{Name: "Hina (New Year)", PersonalName: "Hina"}
	computeVariant(&s)
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
	s := Student{Name: "", PersonalName: "Hanako", FamilyName: "Ichinose"}
	computeVariant(&s)
	if s.Name != "Hanako" {
		t.Errorf("Name should fall back to PersonalName, got %q", s.Name)
	}
	if s.Variant != "" {
		t.Errorf("Variant should be empty, got %q", s.Variant)
	}
}

// --- displayName ---

func TestDisplayNameFitsFullName(t *testing.T) {
	s := Student{Name: "Hina (Swimsuit)", Base: "Hina", Variant: "Swimsuit", Abbr: "SW"}
	name, abbrUsed := displayName(s, 20)
	if name != "Hina (Swimsuit)" {
		t.Errorf("displayName = %q, want 'Hina (Swimsuit)'", name)
	}
	if abbrUsed {
		t.Error("abbrUsed should be false when full name fits")
	}
}

func TestDisplayNameUsesAbbrWhenTight(t *testing.T) {
	s := Student{Name: "Hina (Swimsuit)", Base: "Hina", Variant: "Swimsuit", Abbr: "SW"}
	// "Hina (Swimsuit)" = 15 chars, "Hina (SW)" = 9 chars
	name, abbrUsed := displayName(s, 10)
	if name != "Hina (SW)" {
		t.Errorf("displayName = %q, want 'Hina (SW)'", name)
	}
	if !abbrUsed {
		t.Error("abbrUsed should be true")
	}
}

func TestDisplayNameTruncatesWhenNoAbbr(t *testing.T) {
	s := Student{Name: "VeryLongNameHere", Base: "VeryLongNameHere", Variant: "", Abbr: ""}
	name, _ := displayName(s, 8)
	if runeLen(name) > 8 {
		t.Errorf("truncated name %q is wider than 8", name)
	}
}

func TestDisplayNameNoVariantNoTruncNeeded(t *testing.T) {
	s := Student{Name: "Hanako", Base: "Hanako", Variant: "", Abbr: ""}
	name, abbrUsed := displayName(s, 10)
	if name != "Hanako" {
		t.Errorf("displayName = %q, want 'Hanako'", name)
	}
	if abbrUsed {
		t.Error("abbrUsed should be false")
	}
}

// --- runeLen ---

func TestRuneLen(t *testing.T) {
	if runeLen("abc") != 3 {
		t.Error("ascii runeLen wrong")
	}
	if runeLen("こんにちは") != 5 {
		t.Error("cjk runeLen wrong")
	}
}
