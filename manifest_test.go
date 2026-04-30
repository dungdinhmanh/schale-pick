package main

import (
	"encoding/json"
	"testing"
)

// --- parseManifest ---

func TestParseManifest_Valid(t *testing.T) {
	raw := []Student{
		{Id: 10000, FamilyName: "Rikuhachima", PersonalName: "Aru", Name: "Aru"},
		{Id: 10001, FamilyName: "Sorasaki", PersonalName: "Hina", Name: "Hina (Swimsuit)"},
	}
	data, _ := json.Marshal(raw)

	students, err := parseManifest(data)
	if err != nil {
		t.Fatalf("parseManifest returned error: %v", err)
	}
	if len(students) != 2 {
		t.Fatalf("got %d students, want 2", len(students))
	}

	// computeVariant should have run
	hina := students[1]
	if hina.Base != "Hina" {
		t.Errorf("Base = %q, want %q", hina.Base, "Hina")
	}
	if hina.Variant != "Swimsuit" {
		t.Errorf("Variant = %q, want %q", hina.Variant, "Swimsuit")
	}
}

func TestParseManifest_InvalidJSON(t *testing.T) {
	_, err := parseManifest([]byte(`not valid json`))
	if err == nil {
		t.Error("parseManifest should return an error for invalid JSON")
	}
}

func TestParseManifest_EmptyArray(t *testing.T) {
	students, err := parseManifest([]byte(`[]`))
	if err != nil {
		t.Fatalf("parseManifest([]) returned error: %v", err)
	}
	if len(students) != 0 {
		t.Errorf("got %d students, want 0", len(students))
	}
}

// --- filterStudents ---

func makeStudents() []Student {
	ss := []Student{
		{Id: 10000, FamilyName: "Rikuhachima", PersonalName: "Aru", Name: "Aru"},
		{Id: 10001, FamilyName: "Sorasaki", PersonalName: "Hina", Name: "Hina (Swimsuit)"},
		{Id: 10002, FamilyName: "Tendou", PersonalName: "Alice", Name: "Alice"},
		{Id: 10003, FamilyName: "Sorasaki", PersonalName: "Hina", Name: "Hina (New Year)"},
	}
	for i := range ss {
		computeVariant(&ss[i])
	}
	return ss
}

func TestFilterStudents_EmptyQueryReturnsAll(t *testing.T) {
	ss := makeStudents()
	got := filterStudents(ss, "")
	if len(got) != len(ss) {
		t.Errorf("empty query: got %d, want %d", len(got), len(ss))
	}
}

func TestFilterStudents_ByPersonalName(t *testing.T) {
	got := filterStudents(makeStudents(), "Aru")
	if len(got) != 1 || got[0].PersonalName != "Aru" {
		t.Errorf("filter 'Aru': got %+v", got)
	}
}

func TestFilterStudents_CaseInsensitive(t *testing.T) {
	got := filterStudents(makeStudents(), "aru")
	if len(got) != 1 {
		t.Errorf("case-insensitive 'aru': got %d results, want 1", len(got))
	}
}

func TestFilterStudents_ByFamilyName(t *testing.T) {
	// Both Hina variants share FamilyName "Sorasaki"
	got := filterStudents(makeStudents(), "Sorasaki")
	if len(got) != 2 {
		t.Errorf("filter 'Sorasaki': got %d results, want 2", len(got))
	}
}

func TestFilterStudents_ByVariant(t *testing.T) {
	got := filterStudents(makeStudents(), "Swimsuit")
	if len(got) != 1 || got[0].Variant != "Swimsuit" {
		t.Errorf("filter 'Swimsuit': got %+v", got)
	}
}

func TestFilterStudents_ByID(t *testing.T) {
	got := filterStudents(makeStudents(), "10002")
	if len(got) != 1 || got[0].Id != 10002 {
		t.Errorf("filter '10002': got %+v", got)
	}
}

func TestFilterStudents_NoMatch(t *testing.T) {
	got := filterStudents(makeStudents(), "zzznomatch")
	if len(got) != 0 {
		t.Errorf("no-match query: got %d results, want 0", len(got))
	}
}
