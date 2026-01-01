package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
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
	_ = f.EndTable()

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
		_ = f.EndTable()
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
	_ = f.EndTable()

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

// sortTestItem is a test struct with various field types for sort testing
type sortTestItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Amount    float64   `json:"amount"`
	CreatedAt string    `json:"created_at"`
	Timestamp time.Time `json:"timestamp"`
}

func TestFormatter_OutputList_SortByStringAscending(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted: Alice, Bob, Charlie
	if len(lines) < 4 {
		t.Fatalf("Expected 4 lines (header + 3 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("First row should be Alice, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second row should be Bob, got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Charlie") {
		t.Errorf("Third row should be Charlie, got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortByStringDescending(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	ctx = WithDesc(ctx, true)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted descending: Charlie, Bob, Alice
	if !strings.Contains(lines[1], "Charlie") {
		t.Errorf("First row should be Charlie, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second row should be Bob, got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Alice") {
		t.Errorf("Third row should be Alice, got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortByIntAscending(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "ID")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 30, Name: "Third"},
		{ID: 10, Name: "First"},
		{ID: 20, Name: "Second"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted: 10, 20, 30
	if !strings.Contains(lines[1], "First") {
		t.Errorf("First row should be First (ID=10), got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Second") {
		t.Errorf("Second row should be Second (ID=20), got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Third") {
		t.Errorf("Third row should be Third (ID=30), got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortByFloat(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "amount")
	ctx = WithDesc(ctx, true)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Small", Amount: 10.50},
		{ID: 2, Name: "Large", Amount: 1000.00},
		{ID: 3, Name: "Medium", Amount: 500.25},
	}

	headers := []string{"ID", "NAME", "AMOUNT"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name, fmt.Sprintf("%.2f", i.Amount)}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted descending by amount: Large, Medium, Small
	if !strings.Contains(lines[1], "Large") {
		t.Errorf("First row should be Large (1000.00), got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Medium") {
		t.Errorf("Second row should be Medium (500.25), got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Small") {
		t.Errorf("Third row should be Small (10.50), got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortByTimeString(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "created_at")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Third", CreatedAt: "2024-03-15T10:00:00Z"},
		{ID: 2, Name: "First", CreatedAt: "2024-01-01T10:00:00Z"},
		{ID: 3, Name: "Second", CreatedAt: "2024-02-14T10:00:00Z"},
	}

	headers := []string{"ID", "NAME", "CREATED_AT"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name, i.CreatedAt}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted by date ascending
	if !strings.Contains(lines[1], "First") {
		t.Errorf("First row should be First (2024-01), got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Second") {
		t.Errorf("Second row should be Second (2024-02), got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Third") {
		t.Errorf("Third row should be Third (2024-03), got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortByTimeTime(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Timestamp")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Third", Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "First", Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Second", Timestamp: time.Date(2024, 2, 14, 10, 0, 0, 0, time.UTC)},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted by time ascending
	if !strings.Contains(lines[1], "First") {
		t.Errorf("First row should be First, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Second") {
		t.Errorf("Second row should be Second, got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Third") {
		t.Errorf("Third row should be Third, got: %s", lines[3])
	}
}

func TestFormatter_OutputList_SortBySnakeCaseField(t *testing.T) {
	// Test that snake_case field names work (via json tag)
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "created_at")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "B", CreatedAt: "2024-02-01"},
		{ID: 2, Name: "A", CreatedAt: "2024-01-01"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if !strings.Contains(lines[1], "A") {
		t.Errorf("First row should be A (earlier date), got: %s", lines[1])
	}
}

func TestFormatter_OutputList_SortByInvalidField(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "invalid_field")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err == nil {
		t.Error("OutputList() should return error for invalid field")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
	if !strings.Contains(err.Error(), "available fields") {
		t.Errorf("Error should list available fields, got: %v", err)
	}
	// Should mention actual field names
	if !strings.Contains(err.Error(), "ID") || !strings.Contains(err.Error(), "Name") {
		t.Errorf("Error should list ID and Name as available fields, got: %v", err)
	}
}

func TestFormatter_OutputList_Limit(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithLimit(ctx, 2)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
		{ID: 3, Name: "Third"},
		{ID: 4, Name: "Fourth"},
		{ID: 5, Name: "Fifth"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should have header + 2 rows = 3 lines
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "First") {
		t.Errorf("First row should be First, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Second") {
		t.Errorf("Second row should be Second, got: %s", lines[2])
	}
}

func TestFormatter_OutputList_LimitZeroMeansNoLimit(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithLimit(ctx, 0)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
		{ID: 3, Name: "Third"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should have all items (header + 3 rows = 4 lines)
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (header + 3 rows), got %d", len(lines))
	}
}

func TestFormatter_OutputList_LimitLargerThanSlice(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithLimit(ctx, 100) // Limit larger than slice size
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should have all items (limit is larger than slice)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 rows), got %d", len(lines))
	}
}

func TestFormatter_OutputList_SortAndLimit(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	ctx = WithLimit(ctx, 2)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
		{ID: 4, Name: "Dave"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted then limited: Alice, Bob (first 2 after sorting)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("First row should be Alice, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second row should be Bob, got: %s", lines[2])
	}
}

func TestFormatter_OutputList_SortDescAndLimit(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "ID")
	ctx = WithDesc(ctx, true)
	ctx = WithLimit(ctx, 3)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "First"},
		{ID: 5, Name: "Fifth"},
		{ID: 3, Name: "Third"},
		{ID: 2, Name: "Second"},
		{ID: 4, Name: "Fourth"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted descending then limited: 5, 4, 3
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (header + 3 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[1], "Fifth") {
		t.Errorf("First row should be Fifth (ID=5), got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Fourth") {
		t.Errorf("Second row should be Fourth (ID=4), got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Third") {
		t.Errorf("Third row should be Third (ID=3), got: %s", lines[3])
	}
}

func TestFormatter_OutputList_JSONModeWithSort(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")
	ctx = WithSortBy(ctx, "Name")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	// Parse JSON to verify order
	var result []sortTestItem
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result))
	}
	if result[0].Name != "Alice" {
		t.Errorf("First item should be Alice, got: %s", result[0].Name)
	}
	if result[1].Name != "Bob" {
		t.Errorf("Second item should be Bob, got: %s", result[1].Name)
	}
	if result[2].Name != "Charlie" {
		t.Errorf("Third item should be Charlie, got: %s", result[2].Name)
	}
}

func TestFormatter_OutputList_JSONModeWithLimit(t *testing.T) {
	ctx := WithFormat(context.Background(), "json")
	ctx = WithLimit(ctx, 2)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
		{ID: 3, Name: "Third"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	// Parse JSON to verify limit
	var result []sortTestItem
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items (limited), got %d", len(result))
	}
}

func TestFormatter_OutputList_EmptySliceWithSort(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	var items []sortTestItem

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	// Should work fine with empty slice
	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected only header line for empty slice, got %d lines", len(lines))
	}
}

func TestFormatter_OutputList_PointerSlice(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []*sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(*sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should be sorted: Alice, Bob, Charlie
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("First row should be Alice, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "Bob") {
		t.Errorf("Second row should be Bob, got: %s", lines[2])
	}
	if !strings.Contains(lines[3], "Charlie") {
		t.Errorf("Third row should be Charlie, got: %s", lines[3])
	}
}

func TestFormatter_OutputList_DoesNotModifyOriginal(t *testing.T) {
	ctx := WithFormat(context.Background(), "text")
	ctx = WithSortBy(ctx, "Name")
	ctx = WithLimit(ctx, 2)
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	items := []sortTestItem{
		{ID: 1, Name: "Charlie"},
		{ID: 2, Name: "Alice"},
		{ID: 3, Name: "Bob"},
	}

	headers := []string{"ID", "NAME"}
	rowFn := func(item any) []string {
		i := item.(sortTestItem)
		return []string{fmt.Sprintf("%d", i.ID), i.Name}
	}

	err := f.OutputList(items, headers, rowFn)
	if err != nil {
		t.Fatalf("OutputList() error = %v", err)
	}

	// Original slice should not be modified
	if len(items) != 3 {
		t.Errorf("Original slice length modified: expected 3, got %d", len(items))
	}
	if items[0].Name != "Charlie" {
		t.Errorf("Original slice order modified: first item should still be Charlie, got %s", items[0].Name)
	}
}

// Template output tests

func TestFormatter_Output_Template(t *testing.T) {
	ctx := WithTemplate(context.Background(), "{{.TransferID}}: {{currency .Amount .Currency}}")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	data := struct {
		TransferID string
		Amount     float64
		Currency   string
	}{
		TransferID: "tfr_123",
		Amount:     1000.50,
		Currency:   "USD",
	}

	err := f.Output(data)
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}

	expected := "tfr_123: 1000.50 USD"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestFormatter_Output_TemplateWithHelpers(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     any
		expected string
	}{
		{
			name:     "upper function",
			template: "{{upper .Name}}",
			data:     struct{ Name string }{"alice"},
			expected: "ALICE",
		},
		{
			name:     "lower function",
			template: "{{lower .Name}}",
			data:     struct{ Name string }{"ALICE"},
			expected: "alice",
		},
		{
			name:     "truncate function",
			template: "{{truncate .Name 5}}",
			data:     struct{ Name string }{"Hello World"},
			expected: "Hello...",
		},
		{
			name:     "truncate short string unchanged",
			template: "{{truncate .Name 20}}",
			data:     struct{ Name string }{"Hello"},
			expected: "Hello",
		},
		{
			name:     "currency function",
			template: "{{currency .Amount .Code}}",
			data: struct {
				Amount float64
				Code   string
			}{1234.56, "EUR"},
			expected: "1234.56 EUR",
		},
		{
			name:     "join function",
			template: "{{join .Tags \", \"}}",
			data:     struct{ Tags []string }{[]string{"a", "b", "c"}},
			expected: "a, b, c",
		},
		{
			name:     "default with empty value",
			template: "{{default \"N/A\" .Value}}",
			data:     struct{ Value string }{""},
			expected: "N/A",
		},
		{
			name:     "default with non-empty value",
			template: "{{default \"N/A\" .Value}}",
			data:     struct{ Value string }{"present"},
			expected: "present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithTemplate(context.Background(), tt.template)
			var buf bytes.Buffer
			f := FromContext(ctx, WithWriter(&buf))

			err := f.Output(tt.data)
			if err != nil {
				t.Fatalf("Output() error = %v", err)
			}

			if buf.String() != tt.expected {
				t.Errorf("got %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestFormatter_Output_TemplateInvalidSyntax(t *testing.T) {
	ctx := WithTemplate(context.Background(), "{{.InvalidBracket")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	err := f.Output(struct{ Name string }{"test"})
	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
	if !strings.Contains(err.Error(), "invalid template") {
		t.Errorf("error should mention 'invalid template', got: %v", err)
	}
}

func TestFormatter_Output_TemplateTakesPrecedence(t *testing.T) {
	// When both template and JSON are set, template should take precedence
	ctx := WithFormat(context.Background(), "json")
	ctx = WithTemplate(ctx, "{{.Name}}")
	var buf bytes.Buffer
	f := FromContext(ctx, WithWriter(&buf))

	err := f.Output(struct{ Name string }{"test"})
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}

	// Should use template, not JSON
	if buf.String() != "test" {
		t.Errorf("template should take precedence over JSON, got: %q", buf.String())
	}
}
