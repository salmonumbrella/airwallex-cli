package api

import (
	"fmt"
	"regexp"
)

// Common ID patterns for Airwallex resources
var (
	// Generic ID pattern: alphanumeric with underscores/dashes, reasonable length
	resourceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)
)

// ValidateResourceID validates that an ID follows expected patterns
func ValidateResourceID(id, resourceType string) error {
	if id == "" {
		return fmt.Errorf("%s ID cannot be empty", resourceType)
	}
	if len(id) > 128 {
		return fmt.Errorf("%s ID too long (max 128 characters)", resourceType)
	}
	if !resourceIDPattern.MatchString(id) {
		return fmt.Errorf("%s ID contains invalid characters", resourceType)
	}
	return nil
}
