package oway

import (
	"fmt"
	"net/http"
)

// Error represents an error from the Oway API
type Error struct {
	// Message is the human-readable error message
	Message string

	// Code is the error code from the API (e.g., "INVALID_ADDRESS", "AUTH_FAILED")
	Code string

	// StatusCode is the HTTP status code
	StatusCode int

	// RequestID is the request ID for debugging (from x-request-id header)
	RequestID string
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s (code: %s, status: %d, request_id: %s)", e.Message, e.Code, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("%s (code: %s, status: %d)", e.Message, e.Code, e.StatusCode)
}

// IsRetryable determines if this error represents a transient failure that should be retried
func (e *Error) IsRetryable() bool {
	switch e.StatusCode {
	case http.StatusTooManyRequests: // 429 - rate limit
		return true
	case http.StatusInternalServerError: // 500 - might be transient
		return true
	case http.StatusNotImplemented: // 501 - permanent error
		return false
	case http.StatusBadGateway: // 502 - might be transient
		return true
	case http.StatusServiceUnavailable: // 503 - temporary outage
		return true
	case http.StatusGatewayTimeout: // 504 - might be transient
		return true
	default:
		// Other 5xx errors - retry by default
		if e.StatusCode >= 500 {
			return true
		}
		// 4xx errors (except 429) are client errors - don't retry
		return false
	}
}

// IsClientError returns true if this is a 4xx client error
func (e *Error) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if this is a 5xx server error
func (e *Error) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// NewError creates a new Oway error
func NewError(message, code string, statusCode int, requestID string) *Error {
	return &Error{
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
		RequestID:  requestID,
	}
}
