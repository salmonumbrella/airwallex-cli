package api

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

func (e *APIError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("%s: %s (source: %s)", e.Code, e.Message, e.Source)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func ParseAPIError(body []byte) *APIError {
	var e APIError
	if err := json.Unmarshal(body, &e); err != nil {
		return &APIError{
			Code:    "unknown",
			Message: string(body),
		}
	}
	return &e
}
