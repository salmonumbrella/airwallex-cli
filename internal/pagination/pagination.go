// Package pagination provides cursor-based pagination for API list operations.
// This pattern enables efficient iteration through large result sets without
// offset-based jumping, which is more reliable for concurrent data changes.
package pagination

import (
	"net/url"
	"strconv"
)

// Identifiable is implemented by items that have an ID for cursor-based pagination.
type Identifiable interface {
	GetID() string
}

// Options configures pagination for list requests.
type Options struct {
	Limit  int    // Max items per page (default: 20, max: 100)
	Cursor string // Cursor from previous page (empty for first page)
}

// QueryParams converts options to URL query parameters.
func (o Options) QueryParams() url.Values {
	params := url.Values{}

	limit := o.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	params.Set("page_size", strconv.Itoa(limit))

	if o.Cursor != "" {
		params.Set("after_id", o.Cursor)
	}
	return params
}

// Result holds paginated results.
type Result[T Identifiable] struct {
	Items   []T  `json:"items"`
	HasMore bool `json:"has_more"`
}

// NextCursor returns the cursor for the next page.
// Returns empty string if there are no more pages.
func (r Result[T]) NextCursor() string {
	if !r.HasMore || len(r.Items) == 0 {
		return ""
	}
	return r.Items[len(r.Items)-1].GetID()
}

// NextCommand returns the CLI command to fetch the next page.
// cmdBase is the base command (e.g., "airwallex transfers list").
func (r Result[T]) NextCommand(cmdBase string) string {
	cursor := r.NextCursor()
	if cursor == "" {
		return ""
	}
	return cmdBase + " --after " + cursor
}
