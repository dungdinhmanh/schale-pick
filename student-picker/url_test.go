package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetStudentIconURL(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{1, fmt.Sprintf("%s/1.webp", IconBaseURL)},
		{100, fmt.Sprintf("%s/100.webp", IconBaseURL)},
		{0, fmt.Sprintf("%s/0.webp", IconBaseURL)},
		{999999, fmt.Sprintf("%s/999999.webp", IconBaseURL)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id_%d", tt.id), func(t *testing.T) {
			result := GetStudentIconURL(tt.id)
			if result != tt.expected {
				t.Errorf("GetStudentIconURL(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestGetStudentPortraitURL(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{1, fmt.Sprintf("%s/1.webp", PortraitBaseURL)},
		{100, fmt.Sprintf("%s/100.webp", PortraitBaseURL)},
		{0, fmt.Sprintf("%s/0.webp", PortraitBaseURL)},
		{999999, fmt.Sprintf("%s/999999.webp", PortraitBaseURL)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id_%d", tt.id), func(t *testing.T) {
			result := GetStudentPortraitURL(tt.id)
			if result != tt.expected {
				t.Errorf("GetStudentPortraitURL(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestIconURLContainsBaseURL(t *testing.T) {
	id := 42
	url := GetStudentIconURL(id)

	if !strings.HasPrefix(url, IconBaseURL) {
		t.Errorf("Icon URL should start with IconBaseURL, got %s", url)
	}

	if !strings.Contains(url, "42.webp") {
		t.Errorf("Icon URL should contain '42.webp', got %s", url)
	}
}

func TestPortraitURLContainsBaseURL(t *testing.T) {
	id := 42
	url := GetStudentPortraitURL(id)

	if !strings.HasPrefix(url, PortraitBaseURL) {
		t.Errorf("Portrait URL should start with PortraitBaseURL, got %s", url)
	}

	if !strings.Contains(url, "42.webp") {
		t.Errorf("Portrait URL should contain '42.webp', got %s", url)
	}
}

func TestURLFormatConsistency(t *testing.T) {
	id := 1
	iconURL := GetStudentIconURL(id)
	portraitURL := GetStudentPortraitURL(id)

	if !strings.HasSuffix(iconURL, ".webp") {
		t.Errorf("Icon URL should end with .webp, got %s", iconURL)
	}
	if !strings.HasSuffix(portraitURL, ".webp") {
		t.Errorf("Portrait URL should end with .webp, got %s", portraitURL)
	}
}
