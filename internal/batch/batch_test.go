package batch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadItems_JSONArray(t *testing.T) {
	// Create temp file with JSON array
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `[{"name": "test1"}, {"name": "test2"}]`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	items, err := ReadItems(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestReadItems_NDJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.ndjson")
	content := `{"name": "test1"}
{"name": "test2"}
{"name": "test3"}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	items, err := ReadItems(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestParseJSON_MaxInputSize(t *testing.T) {
	// Create input larger than 10MB
	largeInput := strings.Repeat(`{"id": "test"}`+"\n", 1000000) // ~14MB
	reader := strings.NewReader(largeInput)

	_, err := parseJSON(reader)
	if err == nil {
		t.Error("expected error for input > 10MB, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' error, got: %v", err)
	}
}

func TestParseJSON_MaxItemCount(t *testing.T) {
	// Create input with more than 10000 items
	var items []string
	for i := 0; i < 10001; i++ {
		items = append(items, fmt.Sprintf(`{"id": "%d"}`, i))
	}
	input := "[" + strings.Join(items, ",") + "]"
	reader := strings.NewReader(input)

	_, err := parseJSON(reader)
	if err == nil {
		t.Error("expected error for > 10000 items, got nil")
	}
	if !strings.Contains(err.Error(), "too many") {
		t.Errorf("expected 'too many' error, got: %v", err)
	}
}

func TestParseJSON_NDJSONLargeLine(t *testing.T) {
	largeValue := strings.Repeat("a", 70*1024) // >64KB default scanner token
	input := fmt.Sprintf("{\"data\":%q}\n", largeValue)
	reader := strings.NewReader(input)

	items, err := parseJSON(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0]["data"] != largeValue {
		t.Errorf("expected data length %d, got %d", len(largeValue), len(items[0]["data"].(string)))
	}
}
