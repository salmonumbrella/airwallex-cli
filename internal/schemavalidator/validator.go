package schemavalidator

import (
	"fmt"
	"regexp"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

// MissingField represents a required field that was not provided
type MissingField struct {
	Key         string
	Path        string
	Description string
}

// Validate checks that all required schema fields are present
func Validate(schema *api.Schema, provided map[string]string) ([]MissingField, error) {
	var missing []MissingField

	for _, field := range schema.Fields {
		if !field.Required {
			continue
		}

		path := field.Path
		if path == "" {
			path = field.Key
		}

		value, ok := provided[path]
		if !ok || value == "" {
			missing = append(missing, MissingField{
				Key:         field.Key,
				Path:        path,
				Description: field.Description,
			})
		}
	}

	return missing, nil
}

// ValidatePattern checks if a value matches the schema's pattern
func ValidatePattern(value, pattern string) error {
	if pattern == "" {
		return nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	if !re.MatchString(value) {
		return fmt.Errorf("value %q does not match pattern %q", value, pattern)
	}

	return nil
}

// FormatMissingFields returns a human-readable error message
func FormatMissingFields(missing []MissingField) string {
	if len(missing) == 0 {
		return ""
	}

	msg := "missing required fields:\n"
	for _, m := range missing {
		if m.Description != "" {
			msg += fmt.Sprintf("  - %s: %s\n", m.Key, m.Description)
		} else {
			msg += fmt.Sprintf("  - %s\n", m.Key)
		}
	}
	return msg
}
