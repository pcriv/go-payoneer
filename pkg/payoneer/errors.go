package payoneer

import (
	"github.com/pablocrivella/go-payoneer/pkg/payoneer/errors"
)

// APIError represents an error returned by the Payoneer API.
// It is re-exported from the errors package to maintain the public API.
type APIError = errors.APIError

// AsAPIError returns true if err is a *APIError.
func AsAPIError(err error, target **APIError) bool {
	return errors.AsAPIError(err, target)
}
