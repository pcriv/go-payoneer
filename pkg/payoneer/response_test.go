package payoneer

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestValidateResponse(t *testing.T) {
	t.Run("Valid HTTP 200 Success", func(t *testing.T) {
		body := `{"status": "Success", "data": {}}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := validateResponse(resp)
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
		body := `{"error": "Bad Request", "error_description": "The request is missing a required parameter", "error_details": {"code": 400, "sub_code": 1000}}`
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := validateResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := errors.AsType[*APIError](err)
		if !ok {
			t.Fatalf("expected APIError, got %T", err)
		}

		if apiErr.HTTPStatusCode != 400 {
			t.Errorf("expected HTTPStatusCode 400, got %d", apiErr.HTTPStatusCode)
		}

		if apiErr.ErrorType != "Bad Request" {
			t.Errorf("expected ErrorType 'Bad Request', got %q", apiErr.ErrorType)
		}

		if apiErr.ErrorDetails == nil {
			t.Fatal("expected ErrorDetails to be present")
		}

		if apiErr.ErrorDetails.Code != 400 {
			t.Errorf("expected error_details.code 400, got %d", apiErr.ErrorDetails.Code)
		}

		if apiErr.SubCode() != 1000 {
			t.Errorf("expected sub_code 1000, got %d", apiErr.SubCode())
		}
	})

	t.Run("HTTP 200 with Business Error", func(t *testing.T) {
		// Payoneer sometimes returns 200 OK with an error status in the body
		body := `{"status": "Failure", "error": "Business validation failed"}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := validateResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := errors.AsType[*APIError](err)
		if !ok {
			t.Fatalf("expected APIError, got %T", err)
		}

		if apiErr.HTTPStatusCode != 200 {
			t.Errorf("expected HTTPStatusCode 200, got %d", apiErr.HTTPStatusCode)
		}

		if apiErr.Status != "Failure" {
			t.Errorf("expected Status 'Failure', got %q", apiErr.Status)
		}
	})

	t.Run("HTTP 200 with Resource Status (Not an error)", func(t *testing.T) {
		// A transaction can have status "Pending" which is not an API error
		body := `{"id": "123", "status": "Pending", "amount": 100}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := validateResponse(resp)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("HTTP 404 with error_details", func(t *testing.T) {
		body := `{"error": "Not Found", "error_description": "The requested resource could not be found", "error_details": {"code": 404, "sub_code": null}}`
		resp := &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}

		err := validateResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		apiErr, ok := errors.AsType[*APIError](err)
		if !ok {
			t.Fatalf("expected APIError, got %T", err)
		}

		if apiErr.HTTPStatusCode != 404 {
			t.Errorf("expected HTTPStatusCode 404, got %d", apiErr.HTTPStatusCode)
		}

		if apiErr.ErrorType != "Not Found" {
			t.Errorf("expected ErrorType 'Not Found', got %q", apiErr.ErrorType)
		}

		if apiErr.ErrorDetails == nil {
			t.Fatal("expected ErrorDetails to be present")
		}

		if apiErr.ErrorDetails.Code != 404 {
			t.Errorf("expected error_details.code 404, got %d", apiErr.ErrorDetails.Code)
		}

		if apiErr.SubCode() != 0 {
			t.Errorf("expected sub_code 0 (null), got %d", apiErr.SubCode())
		}
	})
}
