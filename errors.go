package vonage

import (
	"fmt"
	"net/http"
)

// Error represents a Vonage API error
type Error struct {
	StatusCode int
	Type       string
	Title      string
	Detail     string
	Instance   string
	Raw        string
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("vonage: %s - %s (status: %d)", e.Title, e.Detail, e.StatusCode)
	}
	if e.Raw != "" {
		return fmt.Sprintf("vonage: status %d - %s", e.StatusCode, e.Raw)
	}
	return fmt.Sprintf("vonage: status %d", e.StatusCode)
}

// IsNotFound returns true if the error is a 404 Not Found
func (e *Error) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func (e *Error) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden
func (e *Error) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsRateLimited returns true if the error is a 429 Too Many Requests
func (e *Error) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// NewError creates a new Vonage error
func NewError(statusCode int, body string) *Error {
	return &Error{
		StatusCode: statusCode,
		Raw:        body,
	}
}

// Common errors
var (
	ErrNotConfigured     = fmt.Errorf("vonage: credentials not configured")
	ErrPrivateKeyMissing = fmt.Errorf("vonage: private key not configured")
	ErrSessionNotFound   = fmt.Errorf("vonage: session not found")
	ErrSessionExpired    = fmt.Errorf("vonage: session expired")
)
