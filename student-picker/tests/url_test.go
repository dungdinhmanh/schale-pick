package picker_test

import (
	"fmt"
	"strings"
	"testing"

	"student-picker/internal/picker"
)

func TestGetStudentIconURL(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{1, fmt.Sprintf("%s/1.webp", picker.IconBaseURL)},
		{100, fmt.Sprintf("%s/100.webp", picker.IconBaseURL)},
		{0, fmt.Sprintf("%s/0.webp", picker.IconBaseURL)},
		{999999, fmt.Sprintf("%s/999999.webp", picker.IconBaseURL)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id_%d", tt.id), func(t *testing.T) {
			result := picker.GetStudentIconURL(tt.id)
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
		{1, fmt.Sprintf("%s/1.webp", picker.PortraitBaseURL)},
		{100, fmt.Sprintf("%s/100.webp", picker.PortraitBaseURL)},
		{0, fmt.Sprintf("%s/0.webp", picker.PortraitBaseURL)},
		{999999, fmt.Sprintf("%s/999999.webp", picker.PortraitBaseURL)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id_%d", tt.id), func(t *testing.T) {
			result := picker.GetStudentPortraitURL(tt.id)
			if result != tt.expected {
				t.Errorf("GetStudentPortraitURL(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestIconURLContainsBaseURL(t *testing.T) {
	url := picker.GetStudentIconURL(42)
	if !strings.HasPrefix(url, picker.IconBaseURL) {
		t.Errorf("Icon URL should start with IconBaseURL, got %s", url)
	}
	if !strings.Contains(url, "42.webp") {
		t.Errorf("Icon URL should contain '42.webp', got %s", url)
	}
}

func TestPortraitURLContainsBaseURL(t *testing.T) {
	url := picker.GetStudentPortraitURL(42)
	if !strings.HasPrefix(url, picker.PortraitBaseURL) {
		t.Errorf("Portrait URL should start with PortraitBaseURL, got %s", url)
	}
	if !strings.Contains(url, "42.webp") {
		t.Errorf("Portrait URL should contain '42.webp', got %s", url)
	}
}

func TestURLFormatConsistency(t *testing.T) {
	iconURL := picker.GetStudentIconURL(1)
	portraitURL := picker.GetStudentPortraitURL(1)
	if !strings.HasSuffix(iconURL, ".webp") {
		t.Errorf("Icon URL should end with .webp, got %s", iconURL)
	}
	if !strings.HasSuffix(portraitURL, ".webp") {
		t.Errorf("Portrait URL should end with .webp, got %s", portraitURL)
	}
}
