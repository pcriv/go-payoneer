package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pablocrivella/go-payoneer/pkg/payoneer"
)

// ValidateResponse checks the HTTP response for errors.
// It handles both non-2xx status codes and business errors in 2xx responses.
func ValidateResponse(resp *http.Response) error {
	if resp.Body == nil {
		if resp.StatusCode >= 400 {
			return &payoneer.APIError{HTTPStatusCode: resp.StatusCode}
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
			return &payoneer.APIError{HTTPStatusCode: resp.StatusCode}
		}
		return nil
	}

	var apiErr payoneer.APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// If it's not JSON, only return error if StatusCode >= 400
		if resp.StatusCode >= 400 {
			return &payoneer.APIError{HTTPStatusCode: resp.StatusCode, Message: string(body)}
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
	// We look for status != "Success" or presence of error_code.
	if apiErr.Status != "" && !strings.EqualFold(apiErr.Status, "Success") {
		return &apiErr
	}
	if apiErr.Code != "" {
		return &apiErr
	}

	return nil
}
