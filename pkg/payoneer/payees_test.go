package payoneer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayeesService(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.ProgramID = "123"

	t.Run("CreateRegistrationLink", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/123/payees/registration-link", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("got method %s, want POST", r.Method)
			}

			var req RegistrationLinkRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}

			if req.PayeeID != "payee1" {
				t.Errorf("got payee_id %s, want payee1", req.PayeeID)
			}

			if req.LanguageID != "en" {
				t.Errorf("got language_id %s, want en", req.LanguageID)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(apiResult[RegistrationLinkResult]{
				Result: RegistrationLinkResult{
					RegistrationLink: "https://payoneer.com/reg/123",
					Token:            "abc-token",
				},
			})
		})

		result, err := client.Payees.CreateRegistrationLink(context.Background(), "payee1", WithLanguage("en"))
		if err != nil {
			t.Fatalf("CreateRegistrationLink failed: %v", err)
		}

		if result.RegistrationLink != "https://payoneer.com/reg/123" {
			t.Errorf("got link %s, want https://payoneer.com/reg/123", result.RegistrationLink)
		}

		if result.Token != "abc-token" {
			t.Errorf("got token %s, want abc-token", result.Token)
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/123/payees/payee1/status", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("got method %s, want GET", r.Method)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(apiResult[PayeeStatus]{
				Result: PayeeStatus{
					AccountID:        "3676945",
					Status:           PayeeStatusDetail{Type: 1, Description: "Active"},
					RegistrationDate: "2021-05-05",
					PayoutMethod:     &PayeePayoutMethod{Type: "BANK", Currency: "USD"},
				},
			})
		})

		status, err := client.Payees.GetStatus(context.Background(), "payee1")
		if err != nil {
			t.Fatalf("GetStatus failed: %v", err)
		}

		if status.Status.Type != 1 {
			t.Errorf("got status type %d, want 1", status.Status.Type)
		}

		if status.Status.Description != "Active" {
			t.Errorf("got status description %s, want Active", status.Status.Description)
		}

		if status.AccountID != "3676945" {
			t.Errorf("got account_id %s, want 3676945", status.AccountID)
		}

		if status.RegistrationDate != "2021-05-05" {
			t.Errorf("got registration_date %s, want 2021-05-05", status.RegistrationDate)
		}

		if status.PayoutMethod == nil || status.PayoutMethod.Type != "BANK" {
			t.Errorf("got payout_method %+v, want BANK/USD", status.PayoutMethod)
		}
	})
}
