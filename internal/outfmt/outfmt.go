package outfmt

import (
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
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteJSONFiltered writes JSON with optional filtering
func WriteJSONFiltered(w io.Writer, v interface{}, query string) error {
	if query == "" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}

	// Convert typed struct to generic interface{} for gojq compatibility.
	// gojq cannot traverse Go structs directly - it needs map[string]interface{}.
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return err
	}

	result, err := filter.Apply(data, query)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
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
