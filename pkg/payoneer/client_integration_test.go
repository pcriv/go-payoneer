package payoneer_test

import (
	"context"
	"encoding/json"
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
		payoneer.WithClientCredentials("client_id", "client_secret"),
	)

	// 2. Test Authenticate
	err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	// 3. Test a Resource Call (Manual call to v4/accounts)
	// In the real SDK, this will be client.Accounts.List(...)
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

	var apiErr *payoneer.APIError
	if !payoneer.AsAPIError(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Code != "ERR_123" {
		t.Errorf("expected code ERR_123, got %s", apiErr.Code)
	}
}

func TestAccounts_GetBalances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/accounts/123/balances" {
			t.Errorf("expected path /v2/accounts/123/balances, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items": [{"currency": "USD", "balance": 123.45, "type": "AVAILABLE"}]}`)
	}))
	defer server.Close()

	client := payoneer.NewClient(payoneer.WithBaseURL(server.URL))
	balances, err := client.Accounts.GetBalances(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetBalances() failed: %v", err)
	}

	if len(balances) != 1 || balances[0].Currency != "USD" || balances[0].Amount != 12345 {
		t.Errorf("unexpected balances: %+v", balances)
	}
}

func TestAccounts_Transactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/accounts/123/transactions" {
			// Check query params
			q := r.URL.Query()
			if q.Get("status") != "COMPLETED" {
				t.Errorf("expected status COMPLETED, got %s", q.Get("status"))
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(
				w,
				`{"items": [{"id": "tx1", "amount": 10.50, "currency": "USD", "status": "COMPLETED", "created_at": "2023-01-01T10:00:00Z"}]}`,
			)

			return
		}

		if r.Method == http.MethodGet && r.URL.Path == "/v2/accounts/123/transactions/tx1" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id": "tx1", "amount": 10.50, "currency": "USD", "status": "COMPLETED", "created_at": "2023-01-01T10:00:00Z"}`)

			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := payoneer.NewClient(payoneer.WithBaseURL(server.URL))
	ctx := context.Background()

	t.Run("ListTransactions", func(t *testing.T) {
		txs, err := client.Accounts.ListTransactions(ctx, "123", payoneer.WithStatus("COMPLETED"))
		if err != nil {
			t.Fatalf("ListTransactions() failed: %v", err)
		}
		if len(txs) != 1 || txs[0].ID != "tx1" || txs[0].Amount != 1050 {
			t.Errorf("unexpected transactions: %+v", txs)
		}
	})

	t.Run("GetTransaction", func(t *testing.T) {
		tx, err := client.Accounts.GetTransaction(ctx, "123", "tx1")
		if err != nil {
			t.Fatalf("GetTransaction() failed: %v", err)
		}
		if tx.ID != "tx1" || tx.Amount != 1050 {
			t.Errorf("unexpected transaction: %+v", tx)
		}
	})
}
