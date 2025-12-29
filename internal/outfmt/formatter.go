package outfmt

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

// Formatter handles output formatting for commands.
type Formatter struct {
	ctx       context.Context
	out       io.Writer
	errOut    io.Writer
	tabWriter *tabwriter.Writer
}

// OutputOption configures a Formatter.
type OutputOption func(*Formatter)

// WithWriter sets the output writer (default: os.Stdout).
func WithWriter(w io.Writer) OutputOption {
	return func(f *Formatter) {
		f.out = w
		f.tabWriter = tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	}
}

// WithErrWriter sets the error writer (default: os.Stderr).
func WithErrWriter(w io.Writer) OutputOption {
	return func(f *Formatter) {
		f.errOut = w
	}
}

// FromContext creates a Formatter from context with optional configuration.
// Priority for writers:
//  1. If WithWriter/WithErrWriter options provided → use them
//  2. Else if GetIO(ctx) returns non-nil → use it
//  3. Else → fall back to os.Stdout/os.Stderr
func FromContext(ctx context.Context, opts ...OutputOption) *Formatter {
	// Start with context IO if available, otherwise use defaults
	io := iocontext.GetIO(ctx)
	f := &Formatter{
		ctx:    ctx,
		out:    io.Out,
		errOut: io.ErrOut,
	}
	f.tabWriter = tabwriter.NewWriter(f.out, 0, 4, 2, ' ', 0)

	// Options override context IO
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Output writes data as JSON or text based on context format.
// For JSON mode, applies JQ filtering if a query is present.
func (f *Formatter) Output(data any) error {
	if IsJSON(f.ctx) {
		return WriteJSONFiltered(f.out, data, GetQuery(f.ctx))
	}
	return nil
}

// StartTable writes table headers and returns true if in text mode.
// Returns false if in JSON mode (caller should skip row writing).
func (f *Formatter) StartTable(headers []string) bool {
	if IsJSON(f.ctx) {
		return false
	}

	u := ui.FromContext(f.ctx)
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(f.tabWriter, "\t")
		}
		_, _ = fmt.Fprint(f.tabWriter, u.FormatHeader(h))
	}
	_, _ = fmt.Fprintln(f.tabWriter)
	return true
}

// Row writes a single row to the table.
func (f *Formatter) Row(columns ...string) {
	for i, col := range columns {
		if i > 0 {
			_, _ = fmt.Fprint(f.tabWriter, "\t")
		}
		_, _ = fmt.Fprint(f.tabWriter, col)
	}
	_, _ = fmt.Fprintln(f.tabWriter)
}

// ColumnType indicates how a column value should be colorized.
type ColumnType int

const (
	// ColumnPlain indicates no special colorization.
	ColumnPlain ColumnType = iota
	// ColumnStatus indicates a status value (COMPLETED, PENDING, FAILED, etc.).
	ColumnStatus
	// ColumnAmount indicates a currency amount.
	ColumnAmount
	// ColumnCurrency indicates a currency code.
	ColumnCurrency
)

// ColorRow writes a row with colorization based on column types.
// columnTypes specifies how each column should be colorized.
// If columnTypes is shorter than columns, remaining columns are treated as plain.
func (f *Formatter) ColorRow(columnTypes []ColumnType, columns ...string) {
	u := ui.FromContext(f.ctx)
	for i, col := range columns {
		if i > 0 {
			_, _ = fmt.Fprint(f.tabWriter, "\t")
		}

		// Determine column type
		var colType ColumnType
		if i < len(columnTypes) {
			colType = columnTypes[i]
		}

		// Apply colorization based on type
		var formatted string
		switch colType {
		case ColumnStatus:
			formatted = u.FormatStatus(col)
		case ColumnAmount:
			formatted = u.FormatAmount(col)
		case ColumnCurrency:
			formatted = u.FormatCurrency(col)
		default:
			formatted = col
		}

		_, _ = fmt.Fprint(f.tabWriter, formatted)
	}
	_, _ = fmt.Fprintln(f.tabWriter)
}

// EndTable flushes the table output.
func (f *Formatter) EndTable() error {
	return f.tabWriter.Flush()
}

// Empty writes a message to stderr indicating no results were found.
func (f *Formatter) Empty(message string) {
	_, _ = fmt.Fprintln(f.errOut, message)
}

// OutputList outputs a slice of items as either JSON or a text table.
// In JSON mode, items are output directly. In text mode, headers define
// columns and rowFn extracts column values from each item.
// Sort and limit transformations are applied based on context flags.
func (f *Formatter) OutputList(items any, headers []string, rowFn func(item any) []string) error {
	return f.OutputListWithColors(items, headers, nil, rowFn)
}

// OutputListWithColors outputs a slice of items with optional colorization.
// columnTypes specifies how each column should be colorized. If nil, no colorization is applied.
func (f *Formatter) OutputListWithColors(items any, headers []string, columnTypes []ColumnType, rowFn func(item any) []string) error {
	val := reflect.ValueOf(items)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return fmt.Errorf("items must be a slice or array, got %T", items)
	}

	// Apply sort and limit transformations
	processed, err := applySortAndLimit(f.ctx, val)
	if err != nil {
		return err
	}

	// If JSON mode, output items directly
	if IsJSON(f.ctx) {
		return f.Output(processed.Interface())
	}

	// Text mode: use table methods
	if !f.StartTable(headers) {
		return nil
	}

	for i := 0; i < processed.Len(); i++ {
		item := processed.Index(i).Interface()
		cols := rowFn(item)
		if columnTypes != nil {
			f.ColorRow(columnTypes, cols...)
		} else {
			f.Row(cols...)
		}
	}

	return f.EndTable()
}

// applySortAndLimit applies sorting and limiting to a slice based on context flags.
// Returns a new reflect.Value containing the processed slice.
func applySortAndLimit(ctx context.Context, val reflect.Value) (reflect.Value, error) {
	// Make a copy to avoid modifying the original slice
	result := reflect.MakeSlice(val.Type(), val.Len(), val.Len())
	reflect.Copy(result, val)

	// Apply sorting if requested
	sortBy := GetSortBy(ctx)
	if sortBy != "" {
		if err := sortSlice(result, sortBy, GetDesc(ctx)); err != nil {
			return reflect.Value{}, err
		}
	}

	// Apply limit if requested
	limit := GetLimit(ctx)
	if limit > 0 && limit < result.Len() {
		result = result.Slice(0, limit)
	}

	return result, nil
}

// sortSlice sorts a reflect.Value slice by the given field name.
// Supports string, int, float, and time.Time fields.
func sortSlice(slice reflect.Value, fieldName string, descending bool) error {
	if slice.Len() == 0 {
		return nil
	}

	// Get the element type and find the field
	elemType := slice.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("cannot sort non-struct elements")
	}

	// Find the field by name (case-insensitive) or by json tag
	fieldIndex, err := findField(elemType, fieldName)
	if err != nil {
		return err
	}

	// Use sort.Sort with a custom sorter that can swap reflect.Values
	sorter := &reflectSorter{
		slice:      slice,
		fieldIndex: fieldIndex,
		descending: descending,
	}
	sort.Sort(sorter)

	return nil
}

// reflectSorter implements sort.Interface for reflect.Value slices.
type reflectSorter struct {
	slice      reflect.Value
	fieldIndex int
	descending bool
}

func (s *reflectSorter) Len() int {
	return s.slice.Len()
}

func (s *reflectSorter) Less(i, j int) bool {
	vi := getFieldValue(s.slice.Index(i), s.fieldIndex)
	vj := getFieldValue(s.slice.Index(j), s.fieldIndex)

	cmp := compareValues(vi, vj)
	if s.descending {
		return cmp > 0
	}
	return cmp < 0
}

func (s *reflectSorter) Swap(i, j int) {
	// Get values
	vi := s.slice.Index(i)
	vj := s.slice.Index(j)
	// Create temp and swap
	tmp := reflect.New(vi.Type()).Elem()
	tmp.Set(vi)
	vi.Set(vj)
	vj.Set(tmp)
}

// findField finds a struct field by name (case-insensitive) or json tag.
// Returns the field index or an error with available field names.
func findField(t reflect.Type, name string) (int, error) {
	nameLower := strings.ToLower(name)
	// Convert snake_case to compare with struct field names
	nameNoUnderscore := strings.ReplaceAll(nameLower, "_", "")

	var availableFields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldNameLower := strings.ToLower(field.Name)
		availableFields = append(availableFields, field.Name)

		// Check exact match (case-insensitive)
		if fieldNameLower == nameLower {
			return i, nil
		}

		// Check without underscores (e.g., created_at matches CreatedAt)
		if strings.ReplaceAll(fieldNameLower, "_", "") == nameNoUnderscore {
			return i, nil
		}

		// Check json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if strings.ToLower(tagName) == nameLower {
				return i, nil
			}
		}
	}

	return -1, fmt.Errorf("field %q not found; available fields: %s", name, strings.Join(availableFields, ", "))
}

// getFieldValue extracts the field value from a struct (handling pointers).
func getFieldValue(v reflect.Value, fieldIndex int) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v.Field(fieldIndex)
}

// compareValues compares two reflect.Values and returns -1, 0, or 1.
// Supports string, int, float, and time.Time types.
func compareValues(a, b reflect.Value) int {
	// Handle invalid/nil values
	if !a.IsValid() && !b.IsValid() {
		return 0
	}
	if !a.IsValid() {
		return -1
	}
	if !b.IsValid() {
		return 1
	}

	// Handle time.Time specially (it's a struct)
	if a.Type() == reflect.TypeOf(time.Time{}) {
		ta := a.Interface().(time.Time)
		tb := b.Interface().(time.Time)
		if ta.Before(tb) {
			return -1
		}
		if ta.After(tb) {
			return 1
		}
		return 0
	}

	switch a.Kind() {
	case reflect.String:
		sa := a.String()
		sb := b.String()
		// Try parsing as time if it looks like a timestamp
		if ta, err := parseTime(sa); err == nil {
			if tb, err := parseTime(sb); err == nil {
				if ta.Before(tb) {
					return -1
				}
				if ta.After(tb) {
					return 1
				}
				return 0
			}
		}
		return strings.Compare(sa, sb)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ia := a.Int()
		ib := b.Int()
		if ia < ib {
			return -1
		}
		if ia > ib {
			return 1
		}
		return 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ua := a.Uint()
		ub := b.Uint()
		if ua < ub {
			return -1
		}
		if ua > ub {
			return 1
		}
		return 0

	case reflect.Float32, reflect.Float64:
		fa := a.Float()
		fb := b.Float()
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0

	default:
		// Fallback: compare string representations
		return strings.Compare(fmt.Sprint(a.Interface()), fmt.Sprint(b.Interface()))
	}
}

// parseTime attempts to parse a string as a time value.
func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}
