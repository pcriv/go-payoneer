package payoneer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"event_type":"payout_created","event_id":"123"}`)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		secret    string
		want      bool
	}{
		{
			name:      "valid signature",
			payload:   payload,
			signature: validSignature,
			secret:    secret,
			want:      true,
		},
		{
			name:      "invalid signature",
			payload:   payload,
			signature: "invalid-signature",
			secret:    secret,
			want:      false,
		},
		{
			name:      "invalid secret",
			payload:   payload,
			signature: validSignature,
			secret:    "wrong-secret",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateSignature(tt.payload, tt.signature, tt.secret); got != tt.want {
				t.Errorf("ValidateSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseWebhook(t *testing.T) {
	secret := "test-secret"
	payload := `{"event_type":"payout_created","event_id":"123","timestamp":"2023-01-01T00:00:00Z","content":{"status":"COMPLETED"}}`
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	t.Run("valid webhook", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))
		req.Header.Set("X-Payoneer-Signature", signature)

		event, err := ParseWebhook(req, secret)
		if err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}

		if event.EventType != "payout_created" {
			t.Errorf("got event type %s, want payout_created", event.EventType)
		}
	})

	t.Run("missing signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))

		_, err := ParseWebhook(req, secret)
		if err == nil {
			t.Error("expected error for missing signature")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))
		req.Header.Set("X-Payoneer-Signature", "wrong")

		_, err := ParseWebhook(req, secret)
		if err == nil {
			t.Error("expected error for invalid signature")
		}
	})
}

func TestWebhookMiddleware(t *testing.T) {
	secret := "test-secret"
	payload := `{"event_type":"payout_created"}`

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != payload {
			t.Errorf("got body %s, want %s", string(body), payload)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := WebhookValidator(secret)(nextHandler)

	t.Run("valid signature proceeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))
		req.Header.Set("X-Payoneer-Signature", signature)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))
		req.Header.Set("X-Payoneer-Signature", "wrong")
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("missing signature returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(payload))
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})
}
