// Package utils provides utility functions.
package utils

import "strings"

// SanitizeInput sanitizes user input to prevent injection attacks
func SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	// Limit length to prevent abuse
	const maxInputLength = 100
	if len(input) > maxInputLength {
		input = input[:maxInputLength]
	}

	return input
}

// SanitizeUsername sanitizes and normalizes usernames
func SanitizeUsername(username string) string {
	// Sanitize input first
	username = SanitizeInput(username)

	// Convert to lowercase for consistency
	username = strings.ToLower(username)

	// Remove @ prefix if present
	username = strings.TrimPrefix(username, "@")

	return username
}

// SanitizeRoleName sanitizes and normalizes role names
func SanitizeRoleName(roleName string) string {
	// Sanitize input first
	roleName = SanitizeInput(roleName)

	// Convert to lowercase for consistency
	roleName = strings.ToLower(roleName)

	return roleName
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Unique removes duplicate strings from a slice
func Unique(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
