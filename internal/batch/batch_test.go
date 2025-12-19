package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadItems_JSONArray(t *testing.T) {
	// Create temp file with JSON array
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `[{"name": "test1"}, {"name": "test2"}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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
