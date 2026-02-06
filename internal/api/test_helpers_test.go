package api

import "encoding/json"

// jn creates a json.Number from a raw number string. The string must match
// the exact JSON text in the test mock response, since json.Unmarshal
// preserves the verbatim number text when decoding into json.Number.
// For example, "1000.50" in JSON becomes json.Number("1000.50"), not "1000.5".
func jn(s string) json.Number {
	return json.Number(s)
}
