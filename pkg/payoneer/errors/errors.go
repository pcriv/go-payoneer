package errors

import (
	"errors"
	"fmt"
)

// APIError represents an error returned by the Payoneer API.
type APIError struct {
	HTTPStatusCode int    `json:"-"`
	Code           string `json:"error_code,omitempty"`
	Message        string `json:"description,omitempty"`
	ErrorType      string `json:"error,omitempty"`             // For OAuth2 errors
	ErrorDesc      string `json:"error_description,omitempty"` // For OAuth2 errors
	Status         string `json:"status,omitempty"`
}

const (
	// ErrCodePayoutNotFound is the Payoneer error code for payout not found.
	ErrCodePayoutNotFound = "2306"
)

func (e *APIError) Error() string {
	if e.ErrorType != "" {
		return fmt.Sprintf("payoneer api error: %s - %s (HTTP %d)", e.ErrorType, e.ErrorDesc, e.HTTPStatusCode)
	}

	return fmt.Sprintf("payoneer api error: %s - %s (Status: %s, HTTP %d)", e.Code, e.Message, e.Status, e.HTTPStatusCode)
}

// AsAPIError returns true if err is a *APIError.
func AsAPIError(err error, target **APIError) bool {
	return errors.As(err, target)
}
