package batch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadItems reads JSON items from a file or stdin.
// If filename is "-" or empty, reads from stdin.
// Supports both JSON array format and newline-delimited JSON (NDJSON).
func ReadItems(filename string) ([]map[string]interface{}, error) {
	var reader io.Reader

	if filename == "" || filename == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
		reader = f
	}

	return parseJSON(reader)
}

func parseJSON(r io.Reader) ([]map[string]interface{}, error) {
	// Read all content
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Try parsing as JSON array first
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err == nil {
		return items, nil
	}

	// Try parsing as NDJSON (newline-delimited JSON)
	items = nil
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
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
