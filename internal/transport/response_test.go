package transport_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/pcriv/go-payoneer/internal/transport"
	payoneererrors "github.com/pcriv/go-payoneer/pkg/payoneer/errors"
)

func TestValidateResponse(t *testing.T) {
	t.Run("Valid HTTP 200 Success", func(t *testing.T) {
		body := `{"status": "Success", "data": {}}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := transport.ValidateResponse(resp)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Ensure body can still be read
		newBody, _ := io.ReadAll(resp.Body)
		if string(newBody) != body {
			t.Errorf("expected body to be preserved, got %s", string(newBody))
		}
	})

	t.Run("HTTP 400 Bad Request", func(t *testing.T) {
		body := `{"error": "invalid_request", "error_description": "The request is missing a required parameter"}`
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := transport.ValidateResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var apiErr *payoneererrors.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T", err)
		}

		if apiErr.HTTPStatusCode != 400 {
			t.Errorf("expected HTTPStatusCode 400, got %d", apiErr.HTTPStatusCode)
		}
	})

	t.Run("HTTP 200 with Business Error", func(t *testing.T) {
		// Payoneer sometimes returns 200 OK with an error status in the body
		body := `{"status": "Failure", "error_code": "1234", "description": "Business validation failed"}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := transport.ValidateResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var apiErr *payoneererrors.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T", err)
		}

		if apiErr.HTTPStatusCode != 200 {
			t.Errorf("expected HTTPStatusCode 200, got %d", apiErr.HTTPStatusCode)
		}
		if apiErr.Code != "1234" {
			t.Errorf("expected Code 1234, got %s", apiErr.Code)
		}
	})

	t.Run("HTTP 200 with Resource Status (Not an error)", func(t *testing.T) {
		// A transaction can have status "Pending" which is not an API error
		body := `{"id": "123", "status": "Pending", "amount": 100}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := transport.ValidateResponse(resp)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
