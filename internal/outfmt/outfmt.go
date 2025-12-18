package outfmt

import (
	"context"
	"encoding/json"
	"io"
)

type contextKey string

const formatKey contextKey = "output_format"

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

func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
