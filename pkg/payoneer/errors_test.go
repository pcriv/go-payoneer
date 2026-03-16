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

	t.Run("Business error format", func(t *testing.T) {
		err := &APIError{
			HTTPStatusCode: 400,
			Code:           "1234",
			Message:        "Validation failed",
			Status:         "Failure",
		}

		got := err.Error()
		want := "payoneer api error: 1234 - Validation failed (Status: Failure, HTTP 400)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
