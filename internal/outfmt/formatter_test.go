package outfmt

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestFormatter_Output_JSON(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	data := map[string]string{"key": "value"}
	err := f.Output(data)
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `"key": "value"`) {
		t.Errorf("Output() JSON missing expected content, got:\n%s", got)
	}
}

func TestFormatter_Output_JSONWithQuery(t *testing.T) {
	// Note: gojq filter requires data to be in JSON-unmarshaled form
	ctx := WithFormat(context.Background(), "json")
	ctx = WithQuery(ctx, ".[0]")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	// Use []interface{} and map[string]interface{} like JSON unmarshal would produce
	data := []interface{}{
		map[string]interface{}{"name": "test", "id": 1},
		map[string]interface{}{"name": "other", "id": 2},
	}

	err := f.Output(data)
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if !strings.Contains(got, `"name": "test"`) {
		t.Errorf("Output() with query missing expected content, got:\n%s", got)
	}
}

func TestFormatter_StartTable_TextMode(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	headers := []string{"COL1", "COL2", "COL3"}
	if !f.StartTable(headers) {
		t.Error("StartTable() should return true in text mode")
	}

	f.Row("a", "b", "c")
	f.Row("d", "e", "f")
	f.EndTable()

	got := buf.String()
	if !strings.Contains(got, "COL1") {
		t.Errorf("StartTable() missing headers, got:\n%s", got)
	}
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") {
		t.Errorf("Row() missing data, got:\n%s", got)
	}
}

func TestFormatter_StartTable_JSONMode(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	if f.StartTable([]string{"COL1", "COL2"}) {
		t.Error("StartTable() should return false in JSON mode")
	}

	// Nothing should be written in JSON mode
	if buf.Len() != 0 {
		t.Errorf("StartTable() wrote data in JSON mode: %s", buf.String())
	}
}

func TestFormatter_Empty(t *testing.T) {
	ctx := context.Background()
	var errBuf bytes.Buffer
	f := FromContext(ctx, WithErrWriter(&errBuf))

	f.Empty("No items found")

	got := strings.TrimSpace(errBuf.String())
	want := "No items found"
	if got != want {
		t.Errorf("Empty() = %q, want %q", got, want)
	}
}

func TestFormatter_WithCustomWriters(t *testing.T) {
	ctx := context.Background()
	var outBuf, errBuf bytes.Buffer
	f := FromContext(ctx, WithWriter(&outBuf), WithErrWriter(&errBuf))

	// Test that custom writers are used
	f.Empty("error message")
	if errBuf.Len() == 0 {
		t.Error("WithErrWriter() not used")
	}

	if f.StartTable([]string{"H1"}) {
		f.Row("val")
		f.EndTable()
	}
	if outBuf.Len() == 0 {
		t.Error("WithWriter() not used")
	}
}

func TestFormatter_TableFormatting(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	f.StartTable([]string{"ID", "NAME", "STATUS"})
	f.Row("1", "Alice", "active")
	f.Row("2", "Bob", "inactive")
	f.EndTable()

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header + 2 rows), got %d", len(lines))
	}

	// Check header
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "NAME") {
		t.Errorf("Header line incorrect: %s", lines[0])
	}

	// Check rows
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("First data row incorrect: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second data row incorrect: %s", lines[2])
	}
}

type testItem struct {
	ID     int
	Name   string
	Status string
}

func TestFormatter_OutputList_TextMode(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []testItem{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "inactive"},
	}

	headers := []string{"ID", "NAME", "STATUS"}
	rowFn := func(item any) []string {
		i := item.(testItem)
		return []string{
			fmt.Sprintf("%d", i.ID),
			i.Name,
			i.Status,
		}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header + 2 rows), got %d", len(lines))
	}

	// Check header
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "NAME") {
		t.Errorf("Header line incorrect: %s", lines[0])
	}

	// Check rows
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("First data row incorrect: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second data row incorrect: %s", lines[2])
	}
}

func TestFormatter_OutputList_JSONMode(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []testItem{
		{ID: 1, Name: "Alice", Status: "active"},
		{ID: 2, Name: "Bob", Status: "inactive"},
	}

	headers := []string{"ID", "NAME", "STATUS"}
	rowFn := func(item any) []string {
		i := item.(testItem)
		return []string{
			fmt.Sprintf("%d", i.ID),
			i.Name,
			i.Status,
		}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	// Should output JSON, not table
	if !strings.Contains(got, `"ID": 1`) {
		t.Errorf("OutputList() JSON missing expected content, got:\n%s", got)
	}
	if !strings.Contains(got, `"Name": "Alice"`) {
		t.Errorf("OutputList() JSON missing expected content, got:\n%s", got)
	}
}

func TestFormatter_OutputList_EmptySlice(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	var items []testItem

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(testItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should only have header, no rows
	if len(lines) != 1 {
		t.Errorf("Expected only header line for empty slice, got %d lines", len(lines))
	}
}

func TestFormatter_OutputList_InvalidInput(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	// Pass a non-slice
	notASlice := "not a slice"

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		return []string{"", ""}
	}

	err := f.OutputList(notASlice, headers, rowFn)
	if err == nil {
		t.Error("OutputList() should return error for non-slice input")
	}
	if !strings.Contains(err.Error(), "must be a slice or array") {
		t.Errorf("OutputList() error message incorrect: %v", err)
	}
}
