package main

import "fmt"

// GetStudentIconURL returns the icon URL for a student
func GetStudentIconURL(id int) string {
	return fmt.Sprintf("%s/%d.webp", IconBaseURL, id)
}

// GetStudentPortraitURL returns the portrait URL for a student
func GetStudentPortraitURL(id int) string {
	return fmt.Sprintf("%s/%d.webp", PortraitBaseURL, id)
}