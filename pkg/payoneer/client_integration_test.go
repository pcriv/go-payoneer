package payoneer_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
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
			"error":             "Bad Request",
			"error_description": "Custom error message",
			"error_details": map[string]any{
				"code":     400,
				"sub_code": 1000,
			},
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

	if apiErr.ErrorType != "Bad Request" {
		t.Errorf("expected error 'Bad Request', got %q", apiErr.ErrorType)
	}

	if apiErr.SubCode() != 1000 {
		t.Errorf("expected sub_code 1000, got %d", apiErr.SubCode())
	}
}

// newCountingAuthServer returns a test server that serves the OAuth token and
// resource endpoints, counting successful token fetches. If failFirstAuth is
// true, the first full Authenticate attempt fails (both auth-style probes
// return 401); subsequent fetches return a valid token.
func newCountingAuthServer(t *testing.T, failFirstAuth bool) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var tokenFetches atomic.Int64 // counts successful token responses
	var rawHits atomic.Int64      // counts all token-endpoint hits (incl. probe duplicates)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/oauth2/token":
			hit := rawHits.Add(1)
			// Oauth2's clientcredentials probes two auth styles per fetch,
			// so a single "failed attempt" is two hits. Fail both.
			if failFirstAuth && hit <= 2 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid_client"}`))

				return
			}
			tokenFetches.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "mock_token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case "/v4/accounts":
			if r.Header.Get("Authorization") != "Bearer mock_token" {
				w.WriteHeader(http.StatusUnauthorized)

				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "Success"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return server, &tokenFetches
}

// TestClient_LazyAuthentication verifies that Do triggers authentication
// automatically when Authenticate was never called explicitly.
func TestClient_LazyAuthentication(t *testing.T) {
	server, tokenHits := newCountingAuthServer(t, false)

	client := payoneer.NewClient(
		payoneer.WithBaseURL(server.URL),
		payoneer.WithAuthBaseURL(server.URL),
		payoneer.WithClientCredentials("id", "secret"),
	)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/v4/accounts", nil)
	if err != nil {
		t.Fatalf("NewRequest() failed: %v", err)
	}

	var result map[string]any
	if err := client.Do(req, &result); err != nil {
		t.Fatalf("Do() failed: %v", err)
	}

	if got := tokenHits.Load(); got != 1 {
		t.Errorf("expected exactly 1 token fetch, got %d", got)
	}
	if result["status"] != "Success" {
		t.Errorf("expected Success, got %v", result["status"])
	}
}

// TestClient_ConcurrentAuthenticationOnce verifies that N concurrent Do calls
// result in exactly one authFn invocation.
func TestClient_ConcurrentAuthenticationOnce(t *testing.T) {
	server, tokenHits := newCountingAuthServer(t, false)

	client := payoneer.NewClient(
		payoneer.WithBaseURL(server.URL),
		payoneer.WithAuthBaseURL(server.URL),
		payoneer.WithClientCredentials("id", "secret"),
	)

	const n = 16
	var wg sync.WaitGroup
	errs := make(chan error, n)
	start := make(chan struct{})
	for range n {
		wg.Go(func() {
			req, err := client.NewRequest(context.Background(), http.MethodGet, "/v4/accounts", nil)
			if err != nil {
				errs <- err
				return
			}
			<-start
			errs <- client.Do(req, nil)
		})
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Do() failed: %v", err)
		}
	}

	if got := tokenHits.Load(); got != 1 {
		t.Errorf("expected exactly 1 token fetch across %d concurrent calls, got %d", n, got)
	}
}

// TestClient_AuthRetryAfterFailure verifies that a failed Authenticate call
// does not cache the failure — a subsequent call can still succeed.
func TestClient_AuthRetryAfterFailure(t *testing.T) {
	server, tokenHits := newCountingAuthServer(t, true)

	client := payoneer.NewClient(
		payoneer.WithBaseURL(server.URL),
		payoneer.WithAuthBaseURL(server.URL),
		payoneer.WithClientCredentials("id", "secret"),
	)

	ctx := context.Background()
	if err := client.Authenticate(ctx); err == nil {
		t.Fatal("expected first Authenticate() to fail")
	}

	req, err := client.NewRequest(ctx, http.MethodGet, "/v4/accounts", nil)
	if err != nil {
		t.Fatalf("NewRequest() failed: %v", err)
	}
	if err := client.Do(req, nil); err != nil {
		t.Fatalf("Do() after failed auth should succeed: %v", err)
	}

	if got := tokenHits.Load(); got != 1 {
		t.Errorf("expected 1 successful token fetch after retry, got %d", got)
	}
}
