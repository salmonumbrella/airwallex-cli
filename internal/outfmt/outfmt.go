package outfmt

import (
	"context"
	"encoding/json"
	"io"

	"github.com/salmonumbrella/airwallex-cli/internal/filter"
)

type contextKey string

const formatKey contextKey = "output_format"
const queryKey contextKey = "query_filter"

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

	result, err := filter.Apply(v, query)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
