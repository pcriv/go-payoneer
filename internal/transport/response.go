package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pcriv/go-payoneer/pkg/payoneer/errors"
)

// ValidateResponse checks the HTTP response for errors.
// It handles both non-2xx status codes and business errors in 2xx responses.
func ValidateResponse(resp *http.Response) error {
	if resp.Body == nil {
		if resp.StatusCode >= 400 {
			return &errors.APIError{HTTPStatusCode: resp.StatusCode}
		}

		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Restore body for subsequent reading
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	if len(body) == 0 {
		if resp.StatusCode >= 400 {
			return &errors.APIError{HTTPStatusCode: resp.StatusCode}
		}

		return nil
	}

	var apiErr errors.APIError
	if uerr := json.Unmarshal(body, &apiErr); uerr != nil {
		// If it's not JSON, only return error if StatusCode >= 400
		if resp.StatusCode >= 400 {
			return &errors.APIError{HTTPStatusCode: resp.StatusCode, Message: string(body)}
		}

		return nil
	}

	apiErr.HTTPStatusCode = resp.StatusCode

	// Check for HTTP error
	if resp.StatusCode >= 400 {
		return &apiErr
	}

	// Check for business error in 200 OK
	// Some Payoneer APIs return 200 OK even if the operation failed.
	// We look for status == "Failure" or presence of error_code.
	if apiErr.Code != "" {
		return &apiErr
	}

	if strings.EqualFold(apiErr.Status, "Failure") {
		return &apiErr
	}

	return nil
}
