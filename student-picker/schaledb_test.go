package main

import (
	"encoding/json"
	"testing"
)

func TestParseManifestSuccess(t *testing.T) {
	testData := []Student{
		{Id: 1, FamilyName: "Test", PersonalName: "Student"},
		{Id: 2, FamilyName: "Another", PersonalName: "Person"},
	}
	jsonData, _ := json.Marshal(testData)

	students, err := parseManifest(jsonData)
	if err != nil {
		t.Fatalf("parseManifest failed: %v", err)
	}

	if len(students) != 2 {
		t.Errorf("expected 2 students, got %d", len(students))
	}

	if students[0].Id != 1 {
		t.Errorf("expected first student Id 1, got %d", students[0].Id)
	}
	if students[0].FamilyName != "Test" {
		t.Errorf("expected FamilyName 'Test', got %s", students[0].FamilyName)
	}
	if students[0].PersonalName != "Student" {
		t.Errorf("expected PersonalName 'Student', got %s", students[0].PersonalName)
	}
}

func TestParseManifestEmpty(t *testing.T) {
	jsonData := []byte("[]")

	students, err := parseManifest(jsonData)
	if err != nil {
		t.Fatalf("parseManifest failed: %v", err)
	}

	if len(students) != 0 {
		t.Errorf("expected 0 students for empty array, got %d", len(students))
	}
}

func TestParseManifestInvalidJSON(t *testing.T) {
	_, err := parseManifest([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseManifestPartialJSON(t *testing.T) {
	_, err := parseManifest([]byte("{\"incomplete\": "))
	if err == nil {
		t.Error("expected error for partial JSON")
	}
}

func TestParseManifestExtraFields(t *testing.T) {
	testData := []byte(`[
		{"Id": 100, "FamilyName": "Test", "PersonalName": "User", "Extra": "ignored"}
	]`)

	students, err := parseManifest(testData)
	if err != nil {
		t.Fatalf("parseManifest failed: %v", err)
	}

	if len(students) != 1 {
		t.Errorf("expected 1 student, got %d", len(students))
	}
	if students[0].Id != 100 {
		t.Errorf("expected Id 100, got %d", students[0].Id)
	}
}

func TestManifestLoadedMsg(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "A", PersonalName: "B"},
	}
	msg := manifestLoadedMsg{students: students}

	if len(msg.students) != 1 {
		t.Errorf("expected 1 student, got %d", len(msg.students))
	}
}

func TestCacheErrorMsg(t *testing.T) {
	msg := cacheErrorMsg{}

	if msg.err != nil {
		t.Error("error should be nil initially")
	}
}

func TestFilterStudentsEmpty(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "Test"},
	}
	result := filterStudents(students, "")
	if len(result) != 1 {
		t.Errorf("expected 1 result for empty query, got %d", len(result))
	}
}

func TestFilterStudentsByFamilyName(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "A"},
		{Id: 2, FamilyName: "Bob", PersonalName: "B"},
		{Id: 3, FamilyName: "Charlie", PersonalName: "C"},
	}
	result := filterStudents(students, "alice")
	if len(result) != 1 {
		t.Errorf("expected 1 result for 'alice', got %d", len(result))
	}
	if len(result) > 0 && result[0].Id != 1 {
		t.Errorf("expected student with Id 1, got %d", result[0].Id)
	}
}

func TestFilterStudentsByPersonalName(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "A", PersonalName: "Alice"},
		{Id: 2, FamilyName: "B", PersonalName: "Bob"},
	}
	result := filterStudents(students, "bob")
	if len(result) != 1 {
		t.Errorf("expected 1 result for 'bob', got %d", len(result))
	}
	if len(result) > 0 && result[0].Id != 2 {
		t.Errorf("expected student with Id 2, got %d", result[0].Id)
	}
}

func TestFilterStudentsById(t *testing.T) {
	students := []Student{
		{Id: 100, FamilyName: "A", PersonalName: "B"},
		{Id: 200, FamilyName: "C", PersonalName: "D"},
	}
	result := filterStudents(students, "100")
	if len(result) != 1 {
		t.Errorf("expected 1 result for '100', got %d", len(result))
	}
}

func TestFilterStudentsCaseInsensitive(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "ALICE", PersonalName: "Test"},
	}
	result := filterStudents(students, "alice")
	if len(result) != 1 {
		t.Errorf("expected 1 result (case insensitive), got %d", len(result))
	}
}

func TestFilterStudentsNoMatch(t *testing.T) {
	students := []Student{
		{Id: 1, FamilyName: "Alice", PersonalName: "Bob"},
	}
	result := filterStudents(students, "xyz")
	if len(result) != 0 {
		t.Errorf("expected 0 results for no match, got %d", len(result))
	}
}
