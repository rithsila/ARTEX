package client

import "fmt"

// APIError represents a sanitized API error from ARTEX client operations.
// It deliberately does not leak upstream tokens, authorization headers, or upstream internal details.
type APIError struct {
	Code       string `json:"code"`
	Status     int    `json:"status"`
	Dispatched bool   `json:"dispatched"`
	Message    string `json:"message,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("api error (%s, status %d, dispatched=%t): %s", e.Code, e.Status, e.Dispatched, e.Message)
	}
	return fmt.Sprintf("api error (%s, status %d, dispatched=%t)", e.Code, e.Status, e.Dispatched)
}
