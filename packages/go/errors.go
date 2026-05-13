package oway

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Oway-Inc/oway-sdk/packages/go/client"
)

// Error is the typed error returned by every Oway SDK method when the API
// rejects a request or returns an unexpected response. It carries the parsed
// RFC 9457 ProblemDetail so callers can branch on Code (machine-readable
// reason), inspect per-field Violations on validation errors, and quote
// RequestID when reporting an issue.
type Error struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Violations []Violation
	RawBody    []byte
}

// Violation describes a single field-level validation failure returned in
// ProblemDetail.violations[].
type Violation struct {
	Field  string
	Reason string
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != "" && e.Message != "" {
		return fmt.Sprintf("oway: %s (%s) [status=%d req=%s]", e.Message, e.Code, e.StatusCode, e.RequestID)
	}
	if e.Message != "" {
		return fmt.Sprintf("oway: %s [status=%d req=%s]", e.Message, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("oway: request failed [status=%d req=%s]", e.StatusCode, e.RequestID)
}

// IsRetryable reports whether the error represents a transient failure that
// is worth retrying. Only well-known transient HTTP status codes return
// true; all other 4xx and unknown errors return false.
func (e *Error) IsRetryable() bool {
	if e == nil {
		return false
	}
	switch e.StatusCode {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// IsClientError returns true if the underlying response was a 4xx.
func (e *Error) IsClientError() bool {
	return e != nil && e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the underlying response was a 5xx.
func (e *Error) IsServerError() bool {
	return e != nil && e.StatusCode >= 500 && e.StatusCode < 600
}

// AsError extracts an *Error from a wrapped error chain. Returns (nil, false)
// when err is unrelated to the SDK.
func AsError(err error) (*Error, bool) {
	var oe *Error
	if errors.As(err, &oe) {
		return oe, true
	}
	return nil, false
}

// parseHTTPError converts a non-2xx response into an *Error. It is tolerant
// of malformed bodies: any field that does not parse cleanly is left empty
// so the caller still gets StatusCode + RawBody to work with.
func parseHTTPError(status int, requestID string, body []byte) *Error {
	out := &Error{StatusCode: status, RequestID: requestID, RawBody: body}

	if len(body) == 0 {
		return out
	}

	var pd client.ProblemDetail
	if err := json.Unmarshal(body, &pd); err != nil {
		return out
	}

	if pd.Reason != nil {
		out.Code = *pd.Reason
	}
	switch {
	case pd.Detail != nil && *pd.Detail != "":
		out.Message = *pd.Detail
	case pd.Title != nil && *pd.Title != "":
		out.Message = *pd.Title
	}

	if pd.Violations != nil {
		out.Violations = make([]Violation, 0, len(*pd.Violations))
		for _, v := range *pd.Violations {
			vv := Violation{}
			if v.Field != nil {
				vv.Field = *v.Field
			}
			if v.Reason != nil {
				vv.Reason = *v.Reason
			}
			out.Violations = append(out.Violations, vv)
		}
	}

	return out
}
