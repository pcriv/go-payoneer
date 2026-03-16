package payoneer

import (
	stderrors "errors"

	"github.com/pcriv/go-payoneer/pkg/payoneer/errors"
)

// Sentinel errors for common validation failures.
var (
	// ErrProgramIDRequired is returned when a service method requires ProgramID but it is not set.
	ErrProgramIDRequired = stderrors.New("program_id is required")
)

// APIError represents an error returned by the Payoneer API.
// It is re-exported from the errors package to maintain the public API.
type APIError = errors.APIError

// AsAPIError returns true if err is a *APIError.
func AsAPIError(err error, target **APIError) bool {
	return errors.AsAPIError(err, target)
}
