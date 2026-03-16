package payoneer

import (
	stderrors "errors"

	"github.com/pcriv/go-payoneer/pkg/payoneer/errors"
)

// Sentinel errors for common validation failures.
var (
	// ErrProgramIDRequired is returned when a service method requires ProgramID but it is not set.
	ErrProgramIDRequired = stderrors.New("program_id is required")
	// ErrAccountIDRequired is returned when accountID is empty.
	ErrAccountIDRequired = stderrors.New("account_id is required")
	// ErrTransactionIDRequired is returned when transactionID is empty.
	ErrTransactionIDRequired = stderrors.New("transaction_id is required")
	// ErrPayeeIDRequired is returned when payeeID is empty.
	ErrPayeeIDRequired = stderrors.New("payee_id is required")
	// ErrClientReferenceIDRequired is returned when clientReferenceID is empty.
	ErrClientReferenceIDRequired = stderrors.New("client_reference_id is required")
)

// APIError represents an error returned by the Payoneer API.
// It is re-exported from the errors package to maintain the public API.
type APIError = errors.APIError
