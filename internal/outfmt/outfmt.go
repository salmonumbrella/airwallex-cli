package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/salmonumbrella/airwallex-cli/internal/filter"
)

type contextKey string

const (
	formatKey   contextKey = "output_format"
	queryKey    contextKey = "query_filter"
	templateKey contextKey = "template_format"
	yesKey      contextKey = "yes_flag"
	limitKey    contextKey = "limit_flag"
	sortByKey   contextKey = "sort_by_flag"
	descKey     contextKey = "desc_flag"
)

func WithFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, formatKey, format)
}

func GetFormat(ctx context.Context) string {
	if v, ok := ctx.Value(formatKey).(string); ok {
		return v
	}
	return "text"
}

func IsJSON(ctx context.Context) bool {
	return GetFormat(ctx) == "json"
}

func WithQuery(ctx context.Context, query string) context.Context {
	return context.WithValue(ctx, queryKey, query)
}

func GetQuery(ctx context.Context) string {
	if v, ok := ctx.Value(queryKey).(string); ok {
		return v
	}
	return ""
}

func WithTemplate(ctx context.Context, tmpl string) context.Context {
	return context.WithValue(ctx, templateKey, tmpl)
}

func GetTemplate(ctx context.Context) string {
	if v, ok := ctx.Value(templateKey).(string); ok {
		return v
	}
	return ""
}

func WriteJSON(w io.Writer, v interface{}) error {
	// Round-trip through JSON to normalize nil slices to [] instead of null.
	// Go's json.Marshal produces null for nil slices, which breaks jq filters
	// like .items[] with "cannot iterate over: null".
	data, err := normalizeJSON(v)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// WriteJSONFiltered writes JSON with optional filtering
func WriteJSONFiltered(w io.Writer, v interface{}, query string) error {
	// Convert typed struct to generic interface{} for gojq compatibility.
	// gojq cannot traverse Go structs directly - it needs map[string]interface{}.
	// Also normalizes nil slices to [] to prevent jq "cannot iterate over: null".
	data, err := normalizeJSON(v)
	if err != nil {
		return err
	}

	if query == "" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	result, err := filter.Apply(data, query)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// normalizeJSON marshals v to JSON and re-decodes it, converting null values
// for known collection keys into empty arrays []. This prevents jq filters
// like .items[] from failing with "cannot iterate over: null".
func normalizeJSON(v interface{}) (interface{}, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var data interface{}
	dec := json.NewDecoder(bytes.NewReader(jsonBytes))
	dec.UseNumber()
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	NullsToEmpty(data)
	return data, nil
}

// NullsToEmpty recursively walks a decoded JSON value and replaces null values
// inside objects with empty arrays [] when the key name matches a known
// collection field. This prevents jq filters from failing with
// "cannot iterate over: null" when Go nil slices serialize as null.
func NullsToEmpty(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			if child == nil && looksLikeSliceKey(k) {
				val[k] = []interface{}{}
			} else {
				NullsToEmpty(child)
			}
		}
	case []interface{}:
		for _, child := range val {
			NullsToEmpty(child)
		}
	}
}

// looksLikeSliceKey returns true if a JSON key name likely represents an
// array/slice field. Uses known collection field names from the Airwallex API.
func looksLikeSliceKey(key string) bool {
	switch key {
	case "items", "rates", "events", "errors", "fields", "limits",
		"balances", "currencies", "transaction_types", "transfer_methods",
		"payment_methods", "enum":
		return true
	}
	return false
}

// Yes flag context functions

func WithYes(ctx context.Context, yes bool) context.Context {
	return context.WithValue(ctx, yesKey, yes)
}

func GetYes(ctx context.Context) bool {
	if v, ok := ctx.Value(yesKey).(bool); ok {
		return v
	}
	return false
}

// Limit flag context functions

func WithLimit(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, limitKey, limit)
}

func GetLimit(ctx context.Context) int {
	if v, ok := ctx.Value(limitKey).(int); ok {
		return v
	}
	return 0
}

// SortBy flag context functions

func WithSortBy(ctx context.Context, sortBy string) context.Context {
	return context.WithValue(ctx, sortByKey, sortBy)
}

func GetSortBy(ctx context.Context) string {
	if v, ok := ctx.Value(sortByKey).(string); ok {
		return v
	}
	return ""
}

// Desc flag context functions

func WithDesc(ctx context.Context, desc bool) context.Context {
	return context.WithValue(ctx, descKey, desc)
}

func GetDesc(ctx context.Context) bool {
	if v, ok := ctx.Value(descKey).(bool); ok {
		return v
	}
	return false
}
