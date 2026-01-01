package pagination

import (
	"testing"
)

func TestOptions_QueryParams_WithCursor(t *testing.T) {
	opts := Options{
		Limit:  10,
		Cursor: "abc123",
	}
	params := opts.QueryParams()

	if params.Get("page_size") != "10" {
		t.Errorf("page_size = %q, want %q", params.Get("page_size"), "10")
	}
	if params.Get("after_id") != "abc123" {
		t.Errorf("after_id = %q, want %q", params.Get("after_id"), "abc123")
	}
}

func TestOptions_QueryParams_WithoutCursor(t *testing.T) {
	opts := Options{Limit: 20}
	params := opts.QueryParams()

	if params.Get("page_size") != "20" {
		t.Errorf("page_size = %q, want %q", params.Get("page_size"), "20")
	}
	if params.Get("after_id") != "" {
		t.Errorf("after_id should be empty, got %q", params.Get("after_id"))
	}
}

func TestOptions_QueryParams_DefaultLimit(t *testing.T) {
	opts := Options{} // No limit set
	params := opts.QueryParams()

	if params.Get("page_size") != "20" {
		t.Errorf("page_size = %q, want %q (default)", params.Get("page_size"), "20")
	}
}

func TestOptions_QueryParams_MaxLimit(t *testing.T) {
	opts := Options{Limit: 500} // Over max
	params := opts.QueryParams()

	if params.Get("page_size") != "100" {
		t.Errorf("page_size = %q, want %q (max)", params.Get("page_size"), "100")
	}
}

type testItem struct {
	ID string `json:"id"`
}

func (t testItem) GetID() string { return t.ID }

func TestResult_NextCursor_FromItems(t *testing.T) {
	result := Result[testItem]{
		Items:   []testItem{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		HasMore: true,
	}
	if cursor := result.NextCursor(); cursor != "c" {
		t.Errorf("NextCursor = %q, want %q", cursor, "c")
	}
}

func TestResult_NextCursor_NoMore(t *testing.T) {
	result := Result[testItem]{
		Items:   []testItem{{ID: "a"}, {ID: "b"}},
		HasMore: false,
	}
	if cursor := result.NextCursor(); cursor != "" {
		t.Errorf("NextCursor = %q, want empty when HasMore=false", cursor)
	}
}

func TestResult_NextCursor_EmptyItems(t *testing.T) {
	result := Result[testItem]{
		Items:   []testItem{},
		HasMore: true,
	}
	if cursor := result.NextCursor(); cursor != "" {
		t.Errorf("NextCursor = %q, want empty when no items", cursor)
	}
}

func TestResult_NextCommand(t *testing.T) {
	result := Result[testItem]{
		Items:   []testItem{{ID: "xyz"}},
		HasMore: true,
	}
	cmd := result.NextCommand("airwallex transfers list")
	expected := "airwallex transfers list --after xyz"
	if cmd != expected {
		t.Errorf("NextCommand = %q, want %q", cmd, expected)
	}
}

func TestResult_NextCommand_NoMore(t *testing.T) {
	result := Result[testItem]{
		Items:   []testItem{{ID: "xyz"}},
		HasMore: false,
	}
	cmd := result.NextCommand("airwallex transfers list")
	if cmd != "" {
		t.Errorf("NextCommand = %q, want empty when no more pages", cmd)
	}
}
