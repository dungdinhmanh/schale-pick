package picker_test

import (
	"encoding/json"
	"testing"

	"student-picker/internal/picker"
)

func TestParseManifestSuccess(t *testing.T) {
	testData := []picker.Student{
		{Id: 1, FamilyName: "Test", PersonalName: "Student"},
		{Id: 2, FamilyName: "Another", PersonalName: "Person"},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	students, err := picker.ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}

	if len(students) != 2 {
		t.Errorf("expected 2 students, got %d", len(students))
	}
}

func TestParseManifestEmpty(t *testing.T) {
	students, err := picker.ParseManifest([]byte("[]"))
	if err != nil {
		t.Fatalf("ParseManifest with empty array failed: %v", err)
	}
	if len(students) != 0 {
		t.Errorf("expected 0 students, got %d", len(students))
	}
}

func TestParseManifestInvalidJSON(t *testing.T) {
	_, err := picker.ParseManifest([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseManifestPartialFields(t *testing.T) {
	raw := `[{"Id":1,"FamilyName":"Test"}]`
	students, err := picker.ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}
	if len(students) != 1 {
		t.Fatalf("expected 1 student, got %d", len(students))
	}
	if students[0].Id != 1 {
		t.Errorf("expected Id 1, got %d", students[0].Id)
	}
}

func TestParseManifestExtraFields(t *testing.T) {
	raw := `[{"Id":1,"FamilyName":"Test","PersonalName":"Student","UnknownField":"value"}]`
	students, err := picker.ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("ParseManifest with extra fields failed: %v", err)
	}
	if len(students) != 1 {
		t.Errorf("expected 1 student, got %d", len(students))
	}
}

func TestFilterStudentsEmpty(t *testing.T) {
	students := []picker.Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "Test"},
	}
	result := picker.FilterStudents(students, "")
	if len(result) != 1 {
		t.Errorf("expected 1 result for empty query, got %d", len(result))
	}
}

func TestFilterStudentsByFamilyName(t *testing.T) {
	students := []picker.Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "A"},
		{Id: 2, FamilyName: "Bob", PersonalName: "B"},
		{Id: 3, FamilyName: "Charlie", PersonalName: "C"},
	}
	result := picker.FilterStudents(students, "alice")
	if len(result) != 1 {
		t.Errorf("expected 1 result for 'alice', got %d", len(result))
	}
	if len(result) > 0 && result[0].Id != 1 {
		t.Errorf("expected student with Id 1, got %d", result[0].Id)
	}
}

func TestFilterStudentsByPersonalName(t *testing.T) {
	students := []picker.Student{
		{Id: 1, FamilyName: "A", PersonalName: "Alice"},
		{Id: 2, FamilyName: "B", PersonalName: "Bob"},
	}
	result := picker.FilterStudents(students, "bob")
	if len(result) != 1 {
		t.Errorf("expected 1 result for 'bob', got %d", len(result))
	}
	if len(result) > 0 && result[0].Id != 2 {
		t.Errorf("expected student with Id 2, got %d", result[0].Id)
	}
}

func TestFilterStudentsById(t *testing.T) {
	students := []picker.Student{
		{Id: 100, FamilyName: "A", PersonalName: "B"},
		{Id: 200, FamilyName: "C", PersonalName: "D"},
	}
	result := picker.FilterStudents(students, "100")
	if len(result) != 1 {
		t.Errorf("expected 1 result for '100', got %d", len(result))
	}
}

func TestFilterStudentsCaseInsensitive(t *testing.T) {
	students := []picker.Student{
		{Id: 1, FamilyName: "ALICE", PersonalName: "Test"},
	}
	result := picker.FilterStudents(students, "alice")
	if len(result) != 1 {
		t.Errorf("expected 1 result (case insensitive), got %d", len(result))
	}
}

func TestFilterStudentsNoMatch(t *testing.T) {
	students := []picker.Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "Bob"},
	}
	result := picker.FilterStudents(students, "xyz")
	if len(result) != 0 {
		t.Errorf("expected 0 results for no match, got %d", len(result))
	}
}
