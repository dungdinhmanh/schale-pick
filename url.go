package main

import (
	"fmt"
	"net/url"
)

// GetStudentPortraitURL returns the portrait URL for a student
func GetStudentPortraitURL(id int) string {
	u, err := url.JoinPath(PortraitBaseURL, fmt.Sprintf("%d.webp", id))
	if err != nil {
		return fmt.Sprintf("%s/%d.webp", PortraitBaseURL, id)
	}
	return u
}