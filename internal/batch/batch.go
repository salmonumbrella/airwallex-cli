package batch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// MaxInputSize is the maximum allowed input size in bytes (10MB)
	MaxInputSize = 10 * 1024 * 1024
	// MaxItemCount is the maximum number of items allowed in a batch
	MaxItemCount = 10000
)

// ReadItems reads JSON items from a file or stdin.
// If filename is "-" or empty, reads from stdin.
// Supports both JSON array format and newline-delimited JSON (NDJSON).
func ReadItems(filename string) ([]map[string]interface{}, error) {
	var reader io.Reader

	if filename == "" || filename == "-" {
		reader = os.Stdin
	} else {
		//nolint:gosec // G304: filename comes from user input, intentional
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = f.Close() }()
		reader = f
	}

	return parseJSON(reader)
}

func parseJSON(r io.Reader) ([]map[string]interface{}, error) {
	// Enforce input size limit using LimitReader
	limitedReader := io.LimitReader(r, MaxInputSize+1)

	// Read all content
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Check if input exceeded the size limit
	if len(data) > MaxInputSize {
		return nil, fmt.Errorf("input too large: exceeds maximum size of %d bytes", MaxInputSize)
	}

	// Try parsing as JSON array first
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err == nil {
		// Check item count limit
		if len(items) > MaxItemCount {
			return nil, fmt.Errorf("too many items: batch contains %d items, maximum is %d", len(items), MaxItemCount)
		}
		return items, nil
	}

	// Try parsing as NDJSON (newline-delimited JSON)
	items = nil
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), MaxInputSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item map[string]interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("failed to parse JSON line: %w", err)
		}
		items = append(items, item)

		// Check item count limit while parsing NDJSON
		if len(items) > MaxItemCount {
			return nil, fmt.Errorf("too many items: batch contains more than %d items", MaxItemCount)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no valid JSON items found in input")
	}

	return items, nil
}

// Result represents the result of a batch operation
type Result struct {
	Index   int                    `json:"index"`
	Success bool                   `json:"success"`
	ID      string                 `json:"id,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
}

// Summary summarizes batch results
type Summary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}
