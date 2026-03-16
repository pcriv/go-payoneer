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

	t.Run("RegistrationURL", func(t *testing.T) {
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

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RegistrationLinkResponse{
				RegistrationLink: "https://payoneer.com/reg/123",
			})
		})

		link, err := client.Payees.RegistrationURL(context.Background(), "payee1", WithLanguage("en"))
		if err != nil {
			t.Fatalf("RegistrationURL failed: %v", err)
		}

		if link != "https://payoneer.com/reg/123" {
			t.Errorf("got link %s, want https://payoneer.com/reg/123", link)
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/123/payees/payee1/status", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("got method %s, want GET", r.Method)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PayeeStatus{
				PayeeID:           "payee1",
				StatusCode:        1,
				StatusDescription: "ACTIVE",
			})
		})

		status, err := client.Payees.GetStatus(context.Background(), "payee1")
		if err != nil {
			t.Fatalf("GetStatus failed: %v", err)
		}

		if status.StatusCode != 1 {
			t.Errorf("got status code %d, want 1", status.StatusCode)
		}
	})

	t.Run("Get", func(t *testing.T) {
		mux.HandleFunc("/v4/programs/123/payees/payee1", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Payee{
				PayeeID:   "payee1",
				FirstName: Some("John"),
				LastName:  Some("Doe"),
			})
		})

		payee, err := client.Payees.Get(context.Background(), "payee1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		firstName, _ := payee.FirstName.Get()
		if firstName != "John" {
			t.Errorf("got first name %s, want John", firstName)
		}
	})
}
