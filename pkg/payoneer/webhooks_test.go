package payoneer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
