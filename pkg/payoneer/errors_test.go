package payoneer

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	t.Run("OAuth2 error format", func(t *testing.T) {
		err := &APIError{
			HTTPStatusCode: 401,
			ErrorType:      "invalid_client",
			ErrorDesc:      "Client authentication failed",
		}

		got := err.Error()
		want := "payoneer api error: invalid_client - Client authentication failed (HTTP 401)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("API error with error_details", func(t *testing.T) {
		subCode := 4012
		err := &APIError{
			HTTPStatusCode: 401,
			ErrorType:      "Unauthorized",
			ErrorDesc:      "The access token is not valid for this account",
			ErrorDetails: &ErrorDetails{
				Code:    401,
				SubCode: &subCode,
			},
		}

		got := err.Error()
		want := "payoneer api error: Unauthorized - The access token is not valid for this account (HTTP 401, sub_code 4012)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Minimal error with only HTTP status", func(t *testing.T) {
		err := &APIError{
			HTTPStatusCode: 500,
		}

		got := err.Error()
		want := "payoneer api error: HTTP 500 (HTTP 500)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("SubCode helper", func(t *testing.T) {
		subCode := 2306
		err := &APIError{
			HTTPStatusCode: 404,
			ErrorDetails: &ErrorDetails{
				SubCode: &subCode,
			},
		}

		if err.SubCode() != 2306 {
			t.Errorf("expected SubCode() = 2306, got %d", err.SubCode())
		}
	})

	t.Run("SubCode helper returns 0 when absent", func(t *testing.T) {
		err := &APIError{HTTPStatusCode: 400}

		if err.SubCode() != 0 {
			t.Errorf("expected SubCode() = 0, got %d", err.SubCode())
		}
	})
}
