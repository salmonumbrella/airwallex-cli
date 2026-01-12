package cmd

import "fmt"

// validateDateRangeFlags validates --from/--to style flags with custom label names.
func validateDateRangeFlags(from, to, fromLabel, toLabel string, validateRange bool) error {
	if err := validateDate(from); err != nil {
		return fmt.Errorf("invalid %s date: %w", fromLabel, err)
	}
	if err := validateDate(to); err != nil {
		return fmt.Errorf("invalid %s date: %w", toLabel, err)
	}
	if validateRange {
		if err := validateDateRange(from, to); err != nil {
			return err
		}
	}
	return nil
}

// parseDateRangeRFC3339 validates date flags and converts to RFC3339 strings.
func parseDateRangeRFC3339(from, to, fromLabel, toLabel string, validateRange bool) (string, string, error) {
	if err := validateDateRangeFlags(from, to, fromLabel, toLabel, validateRange); err != nil {
		return "", "", err
	}

	fromRFC3339 := ""
	if from != "" {
		var err error
		fromRFC3339, err = convertDateToRFC3339(from)
		if err != nil {
			return "", "", fmt.Errorf("invalid %s date: %w", fromLabel, err)
		}
	}

	toRFC3339 := ""
	if to != "" {
		var err error
		toRFC3339, err = convertDateToRFC3339End(to)
		if err != nil {
			return "", "", fmt.Errorf("invalid %s date: %w", toLabel, err)
		}
	}

	return fromRFC3339, toRFC3339, nil
}
