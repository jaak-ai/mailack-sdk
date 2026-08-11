package mailack

import "fmt"

// APIError is the decoded form of mailack's error envelope:
//
//	{"error":{"code":"snake_case","message":"human readable"}}
type APIError struct {
	// Status is the HTTP status code.
	Status int
	// Code is the machine-readable error code (e.g. "quota_exceeded").
	Code string
	// Message is a human-readable description.
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mailack: HTTP %d %s: %s", e.Status, e.Code, e.Message)
}

// Is reports whether err is an *APIError with the given code.
func Is(err error, code string) bool {
	if e, ok := err.(*APIError); ok {
		return e.Code == code
	}
	return false
}
