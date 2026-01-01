// Package reqbuilder provides utilities for building nested request maps from flat CLI flags.
//
// FUTURE INFRASTRUCTURE: This package is not yet integrated into the beneficiaries
// command but is ready for use. It was created as part of the schema-driven
// architecture to convert flat CLI flag paths (e.g., "beneficiary.bank_details.account_name")
// into the nested JSON structures required by the Airwallex API.
//
// Integration points:
//   - beneficiaries create: build request body from --field flags
//   - beneficiaries update: merge partial updates with existing data
//   - transfers create: build complex nested payment instructions
//
// Example usage:
//
//	fields := map[string]string{
//	    "beneficiary.bank_details.account_name": "John Doe",
//	    "beneficiary.bank_details.account_number": "123456789",
//	    "beneficiary.entity_type": "PERSONAL",
//	}
//	request := reqbuilder.BuildNestedMap(fields)
//	// Result: {"beneficiary": {"bank_details": {"account_name": "John Doe", ...}, "entity_type": "PERSONAL"}}
package reqbuilder

import (
	"sort"
	"strings"
)

// BuildNestedMap converts flat "path.to.field" keys into nested maps
func BuildNestedMap(fields map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	// Sort keys to ensure deterministic processing order.
	// Shorter paths are processed first, then alphabetically.
	// This ensures "parent" is processed before "parent.child",
	// allowing nested paths to overwrite scalar values with maps.
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		// First by length (shorter first)
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) < len(keys[j])
		}
		// Then alphabetically for determinism
		return keys[i] < keys[j]
	})

	for _, path := range keys {
		value := fields[path]
		if value == "" {
			continue
		}
		setNestedValue(result, path, value)
	}

	return result
}

// setNestedValue sets a value at a dot-separated path
func setNestedValue(m map[string]interface{}, path, value string) {
	parts := strings.Split(path, ".")
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			current[part] = value
		} else {
			// Intermediate part - ensure map exists
			if nested, ok := current[part].(map[string]interface{}); ok {
				current = nested
			} else {
				current[part] = make(map[string]interface{})
				current = current[part].(map[string]interface{})
			}
		}
	}
}

// AddRoutingType adds the routing type field for a routing value
func AddRoutingType(m map[string]interface{}, index int, routingType string) {
	if routingType == "" {
		return
	}

	beneficiary, ok := m["beneficiary"].(map[string]interface{})
	if !ok {
		return
	}

	bankDetails, ok := beneficiary["bank_details"].(map[string]interface{})
	if !ok {
		return
	}

	typeKey := "account_routing_type1"
	if index == 2 {
		typeKey = "account_routing_type2"
	}
	bankDetails[typeKey] = routingType
}

// MergeRequest merges additional fields into a request map
func MergeRequest(base, additional map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Merge additional (recursive for maps)
	for k, v := range additional {
		if existing, ok := result[k]; ok {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if additionalMap, ok := v.(map[string]interface{}); ok {
					result[k] = MergeRequest(existingMap, additionalMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}
