package payoneer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// validateResponse checks the HTTP response for errors.
// It handles both non-2xx status codes and business errors in 2xx responses.
func validateResponse(resp *http.Response) error {
	if resp.Body == nil {
		if resp.StatusCode >= 400 {
			return &APIError{HTTPStatusCode: resp.StatusCode}
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
			return &APIError{HTTPStatusCode: resp.StatusCode}
		}

		return nil
	}

	var apiErr APIError
	if uerr := json.Unmarshal(body, &apiErr); uerr != nil {
		// If it's not JSON, only return error if StatusCode >= 400
		if resp.StatusCode >= 400 {
			return &APIError{HTTPStatusCode: resp.StatusCode, ErrorType: string(body)}
		}

		return nil
	}

	apiErr.HTTPStatusCode = resp.StatusCode

	// Check for HTTP error
	if resp.StatusCode >= 400 {
		return &apiErr
	}

	// Check for business error in 200 OK.
	// Some Payoneer APIs return 200 OK even if the operation failed.
	if apiErr.ErrorType != "" {
		return &apiErr
	}

	if strings.EqualFold(apiErr.Status, "Failure") {
		return &apiErr
	}

	return nil
}
