package main

import "testing"

// --- computeVariant ---

func TestComputeVariant_NoVariant(t *testing.T) {
	s := Student{Name: "Hina"}
	computeVariant(&s)
	if s.Base != "Hina" {
		t.Errorf("Base = %q, want %q", s.Base, "Hina")
	}
	if s.Variant != "" {
		t.Errorf("Variant = %q, want empty", s.Variant)
	}
	if s.Abbr != "" {
		t.Errorf("Abbr = %q, want empty", s.Abbr)
	}
}

func TestComputeVariant_SingleWordVariant(t *testing.T) {
	s := Student{Name: "Hina (Swimsuit)"}
	computeVariant(&s)
	if s.Base != "Hina" {
		t.Errorf("Base = %q, want %q", s.Base, "Hina")
	}
	if s.Variant != "Swimsuit" {
		t.Errorf("Variant = %q, want %q", s.Variant, "Swimsuit")
	}
	if s.Abbr != "SW" {
		t.Errorf("Abbr = %q, want %q", s.Abbr, "SW")
	}
}

func TestComputeVariant_MultiWordVariant(t *testing.T) {
	s := Student{Name: "Hina (New Year)"}
	computeVariant(&s)
	if s.Base != "Hina" {
		t.Errorf("Base = %q, want %q", s.Base, "Hina")
	}
	if s.Variant != "New Year" {
		t.Errorf("Variant = %q, want %q", s.Variant, "New Year")
	}
	if s.Abbr != "NY" {
		t.Errorf("Abbr = %q, want %q", s.Abbr, "NY")
	}
}

func TestComputeVariant_FallsBackToPersonalName(t *testing.T) {
	s := Student{PersonalName: "Aru"}
	computeVariant(&s)
	if s.Base != "Aru" {
		t.Errorf("Base = %q, want %q", s.Base, "Aru")
	}
}

// --- makeAbbr ---

func TestMakeAbbr_SingleWord(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Swimsuit", "SW"},
		{"Bunny", "BU"},
		{"X", "X"},
	}
	for _, c := range cases {
		got := makeAbbr(c.input)
		if got != c.want {
			t.Errorf("makeAbbr(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestMakeAbbr_MultiWord(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"New Year", "NY"},
		{"Hot Spring", "HS"},
		{"Maid Cafe Special", "MCS"},
	}
	for _, c := range cases {
		got := makeAbbr(c.input)
		if got != c.want {
			t.Errorf("makeAbbr(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestMakeAbbr_Empty(t *testing.T) {
	if got := makeAbbr(""); got != "" {
		t.Errorf("makeAbbr(\"\") = %q, want empty", got)
	}
}

// --- displayName ---

func TestDisplayName_FitsExactly(t *testing.T) {
	s := Student{Name: "Aru"}
	computeVariant(&s)
	name, abbrUsed := displayName(s, 10)
	if name != "Aru" {
		t.Errorf("name = %q, want %q", name, "Aru")
	}
	if abbrUsed {
		t.Error("abbrUsed should be false when name fits")
	}
}

func TestDisplayName_UsesAbbr(t *testing.T) {
	s := Student{Name: "Hina (Swimsuit)"}
	computeVariant(&s)
	// "Hina (Swimsuit)" = 16 runes, "Hina (SW)" = 9 runes — fits in width 10
	name, abbrUsed := displayName(s, 10)
	if name != "Hina (SW)" {
		t.Errorf("name = %q, want %q", name, "Hina (SW)")
	}
	if !abbrUsed {
		t.Error("abbrUsed should be true when abbreviation was applied")
	}
}

func TestDisplayName_Truncates(t *testing.T) {
	s := Student{Name: "Abcdefghij"}
	computeVariant(&s)
	name, _ := displayName(s, 6)
	// Should be first 5 runes + "…"
	if name != "Abcde…" {
		t.Errorf("name = %q, want %q", name, "Abcde…")
	}
}

// --- runeLen ---

func TestRuneLen(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"Aru", 3},
		{"Hina (SW)", 9},
		{"", 0},
		{"日本語", 3},
	}
	for _, c := range cases {
		if got := runeLen(c.input); got != c.want {
			t.Errorf("runeLen(%q) = %d, want %d", c.input, got, c.want)
		}
	}
}
