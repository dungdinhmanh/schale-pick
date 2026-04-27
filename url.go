package main

import "fmt"

// GetStudentPortraitURL returns the portrait URL for a student
func GetStudentPortraitURL(id int) string {
	return fmt.Sprintf("%s/%d.webp", PortraitBaseURL, id)
}