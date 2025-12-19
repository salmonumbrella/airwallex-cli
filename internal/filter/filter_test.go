package filter

import (
	"testing"
)

func TestApply(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"id": "1", "status": "active"},
			map[string]interface{}{"id": "2", "status": "inactive"},
		},
	}

	// Test simple property access
	result, err := Apply(data, ".items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	// Test filter
	result, err = Apply(data, `.items[] | select(.status == "active")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if item["id"] != "1" {
		t.Errorf("expected id=1, got %v", item["id"])
	}
}

func TestApplyToJSON(t *testing.T) {
	input := `{"name": "test", "value": 42}`

	output, err := ApplyToJSON([]byte(input), ".name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != `"test"` {
		t.Errorf("expected '\"test\"', got %s", string(output))
	}
}
