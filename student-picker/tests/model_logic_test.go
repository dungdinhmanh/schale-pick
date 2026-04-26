package picker_test

import (
	"testing"

	"student-picker/internal/picker"
)

// --- IsNavKey ---

func TestIsNavKeyTrue(t *testing.T) {
	navKeys := []string{"up", "down", "left", "right", "k", "j", "h", "l", "pgup", "pgdown", "home", "end"}
	for _, k := range navKeys {
		if !picker.IsNavKey(k) {
			t.Errorf("IsNavKey(%q) = false, want true", k)
		}
	}
}

func TestIsNavKeyFalse(t *testing.T) {
	nonNav := []string{"enter", "esc", "tab", "q", "?", "/", "i", "ctrl+c"}
	for _, k := range nonNav {
		if picker.IsNavKey(k) {
			t.Errorf("IsNavKey(%q) = true, want false", k)
		}
	}
}

// --- FilterStudents ---

func TestFilterStudentsEmptyQueryReturnsAll(t *testing.T) {
	students := []picker.Student{
		{Id: 1, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
	}
	got := picker.FilterStudents(students, "")
	if len(got) != 1 {
		t.Errorf("empty query should return all, got %d", len(got))
	}
}

func TestFilterStudentsByPersonalNameVariant(t *testing.T) {
	students := []picker.Student{
		{Id: 1, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
		{Id: 2, Name: "Hina", PersonalName: "Hina", Base: "Hina"},
	}
	got := picker.FilterStudents(students, "hana")
	if len(got) != 1 || got[0].Id != 1 {
		t.Errorf("filter by PersonalName failed, got %v", got)
	}
}

func TestFilterStudentsByVariant(t *testing.T) {
	students := []picker.Student{
		{Id: 1, Name: "Hanako (Swimsuit)", PersonalName: "Hanako", Base: "Hanako", Variant: "Swimsuit"},
		{Id: 2, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
	}
	got := picker.FilterStudents(students, "swimsuit")
	if len(got) != 1 || got[0].Id != 1 {
		t.Errorf("filter by Variant failed, got %v", got)
	}
}

func TestFilterStudentsByBase(t *testing.T) {
	students := []picker.Student{
		{Id: 10, Name: "Hina (Swimsuit)", PersonalName: "Hina", Base: "Hina", Variant: "Swimsuit"},
		{Id: 11, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
	}
	got := picker.FilterStudents(students, "hina")
	if len(got) != 1 || got[0].Id != 10 {
		t.Errorf("filter by Base failed, got %v", got)
	}
}

func TestFilterStudentsByFullName(t *testing.T) {
	students := []picker.Student{
		{Id: 10, Name: "Hina (Swimsuit)", PersonalName: "Hina", Base: "Hina", Variant: "Swimsuit"},
		{Id: 11, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
	}
	got := picker.FilterStudents(students, "hina (swim")
	if len(got) != 1 || got[0].Id != 10 {
		t.Errorf("filter by full Name failed, got %v", got)
	}
}

func TestFilterStudentsByNumericId(t *testing.T) {
	students := []picker.Student{
		{Id: 23007, Name: "Hanako", PersonalName: "Hanako", Base: "Hanako"},
		{Id: 10074, Name: "Hanako (Swimsuit)", PersonalName: "Hanako", Base: "Hanako", Variant: "Swimsuit"},
	}
	got := picker.FilterStudents(students, "23007")
	if len(got) != 1 || got[0].Id != 23007 {
		t.Errorf("filter by ID failed, got %v", got)
	}
}

// --- ParseManifest with ComputeVariant ---

func TestParseManifestComputesVariants(t *testing.T) {
	raw := `[{"Id":10074,"FamilyName":"Ichinose","PersonalName":"Hanako","Name":"Hanako (Swimsuit)"}]`
	students, err := picker.ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("ParseManifest error: %v", err)
	}
	if len(students) != 1 {
		t.Fatalf("expected 1 student, got %d", len(students))
	}
	s := students[0]
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

func TestParseManifestNoVariant(t *testing.T) {
	raw := `[{"Id":23007,"FamilyName":"Ichinose","PersonalName":"Hanako","Name":"Hanako"}]`
	students, err := picker.ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("ParseManifest error: %v", err)
	}
	s := students[0]
	if s.Variant != "" {
		t.Errorf("Variant should be empty, got %q", s.Variant)
	}
	if s.Abbr != "" {
		t.Errorf("Abbr should be empty, got %q", s.Abbr)
	}
}
