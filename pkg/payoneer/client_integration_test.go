package payoneer_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestClient_Integration(t *testing.T) {
	// 1. Setup a Mock Payoneer API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/oauth2/token":
			// Mock Token Response
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "mock_token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case "/v4/accounts":
			// Check for Authorization Header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer mock_token" {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":             "invalid_token",
					"error_description": "Token is invalid",
				})

				return
			}

			// Mock Successful Resource Response
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "Success",
				"data":   []any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := payoneer.NewClient(
		payoneer.WithBaseURL(server.URL),
		payoneer.WithAuthBaseURL(server.URL),
		payoneer.WithClientCredentials("client_id", "client_secret"),
	)

	// 2. Test Authenticate
	err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	// 3. Test a Resource Call
	req, err := client.NewRequest(ctx, http.MethodGet, "/v4/accounts", nil)
	if err != nil {
		t.Fatalf("NewRequest() failed: %v", err)
	}

	var result map[string]any
	err = client.Do(req, &result)
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}

	if result["status"] != "Success" {
		t.Errorf("expected Success status, got %v", result["status"])
	}
}

func TestClient_AuthenticateFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<html>Error page</html>`)
	}))
	defer server.Close()

	client := payoneer.NewClient(
		payoneer.WithAuthBaseURL(server.URL),
		payoneer.WithClientCredentials("bad_id", "bad_secret"),
	)

	err := client.Authenticate(context.Background())
	if err == nil {
		t.Fatal("expected Authenticate() to fail with invalid credentials")
	}

	if !errors.Is(err, payoneer.ErrAuthenticationFailed) {
		t.Errorf("expected ErrAuthenticationFailed, got: %v", err)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code":  "ERR_123",
			"description": "Custom error message",
			"status":      "Failure",
		})
	}))
	defer server.Close()

	client := payoneer.NewClient(payoneer.WithBaseURL(server.URL))
	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/test", nil)

	err := client.Do(req, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := errors.AsType[*payoneer.APIError](err)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Code != "ERR_123" {
		t.Errorf("expected code ERR_123, got %s", apiErr.Code)
	}
}

