package payoneer

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for common validation failures.
var (
	// ErrProgramIDRequired is returned when a service method requires ProgramID but it is not set.
	ErrProgramIDRequired = errors.New("program_id is required")
	// ErrPayeeIDRequired is returned when payeeID is empty.
	ErrPayeeIDRequired = errors.New("payee_id is required")
	// ErrClientReferenceIDRequired is returned when clientReferenceID is empty.
	ErrClientReferenceIDRequired = errors.New("client_reference_id is required")
	// ErrAuthenticationFailed is returned when the OAuth2 authentication flow fails.
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// APIError represents an error returned by the Payoneer API.
//
// For HTTP 4xx/5xx errors, the API returns:
//
//	{
//	  "error": "Not Found",
//	  "error_description": "The requested resource could not be found",
//	  "error_details": { "code": 404, "sub_code": null }
//	}
//
// For OAuth2 errors, the error and error_description fields are used directly.
type APIError struct {
	HTTPStatusCode int           `json:"-"`
	ErrorType      string        `json:"error,omitempty"`
	ErrorDesc      string        `json:"error_description,omitempty"`
	ErrorDetails   *ErrorDetails `json:"error_details,omitempty"`
	Status         string        `json:"status,omitempty"`
}

// ErrorDetails contains the structured error information from the Payoneer API.
type ErrorDetails struct {
	Code    int           `json:"code,omitempty"`
	SubCode *int          `json:"sub_code,omitempty"`
	Target  string        `json:"target,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

// ErrorDetail represents a single validation error from the Payoneer API.
type ErrorDetail struct {
	Code    int    `json:"code,omitempty"`
	Target  string `json:"target,omitempty"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
	SubCode *int   `json:"sub_code,omitempty"`
}

const (
	// ErrSubCodePayoutNotFound is the Payoneer sub_code for payout not found.
	ErrSubCodePayoutNotFound = 2306
)

func (e *APIError) Error() string {
	var parts []string

	if e.ErrorType != "" {
		parts = append(parts, e.ErrorType)
	}

	if e.ErrorDesc != "" {
		parts = append(parts, e.ErrorDesc)
	}

	msg := strings.Join(parts, " - ")
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", e.HTTPStatusCode)
	}

	if e.ErrorDetails != nil && e.ErrorDetails.SubCode != nil {
		return fmt.Sprintf("payoneer api error: %s (HTTP %d, sub_code %d)", msg, e.HTTPStatusCode, *e.ErrorDetails.SubCode)
	}

	return fmt.Sprintf("payoneer api error: %s (HTTP %d)", msg, e.HTTPStatusCode)
}

// SubCode returns the sub_code from error_details, or 0 if not present.
func (e *APIError) SubCode() int {
	if e.ErrorDetails != nil && e.ErrorDetails.SubCode != nil {
		return *e.ErrorDetails.SubCode
	}

	return 0
}
