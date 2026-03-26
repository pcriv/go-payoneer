package payoneer

import (
	"errors"
	"fmt"
)

// Sentinel errors for common validation failures.
var (
	// ErrProgramIDRequired is returned when a service method requires ProgramID but it is not set.
	ErrProgramIDRequired = errors.New("program_id is required")
	// ErrAccountIDRequired is returned when accountID is empty.
	ErrAccountIDRequired = errors.New("account_id is required")
	// ErrTransactionIDRequired is returned when transactionID is empty.
	ErrTransactionIDRequired = errors.New("transaction_id is required")
	// ErrPayeeIDRequired is returned when payeeID is empty.
	ErrPayeeIDRequired = errors.New("payee_id is required")
	// ErrClientReferenceIDRequired is returned when clientReferenceID is empty.
	ErrClientReferenceIDRequired = errors.New("client_reference_id is required")
	// ErrAuthenticationFailed is returned when the OAuth2 authentication flow fails.
	ErrAuthenticationFailed = errors.New("authentication failed")
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
