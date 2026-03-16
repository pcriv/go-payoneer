package payoneer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// WebhookEvent represents a Payoneer webhook notification (IPCN).
type WebhookEvent struct {
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	Timestamp string          `json:"timestamp"`
	Content   json.RawMessage `json:"content"`
}

const (
	// HeaderPayoneerSignature is the HTTP header containing the HMAC signature.
	HeaderPayoneerSignature = "X-Payoneer-Signature"
)

// ValidateSignature verifies the authenticity of a webhook payload using HMAC SHA-256.
func ValidateSignature(payload []byte, signature string, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ParseWebhook extracts the signature from the request, validates it, and unmarshals the body.
func ParseWebhook(r *http.Request, secret string) (*WebhookEvent, error) {
	signature := r.Header.Get(HeaderPayoneerSignature)
	if signature == "" {
		return nil, errors.New("missing " + HeaderPayoneerSignature + " header")
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB max
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()

	if !ValidateSignature(body, signature, secret) {
		return nil, errors.New("invalid signature")
	}

	var event WebhookEvent
	if uerr := json.Unmarshal(body, &event); uerr != nil {
		return nil, uerr
	}

	return &event, nil
}

// WebhookValidator returns an HTTP middleware that validates the Payoneer signature.
// If validation fails, it returns 401 Unauthorized.
// If validation succeeds, it replaces r.Body with a new reader so it can be read again.
func WebhookValidator(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signature := r.Header.Get(HeaderPayoneerSignature)
			if signature == "" {
				http.Error(w, "missing signature", http.StatusUnauthorized)

				return
			}

			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB max
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)

				return
			}
			_ = r.Body.Close()

			if !ValidateSignature(body, signature, secret) {
				http.Error(w, "invalid signature", http.StatusUnauthorized)

				return
			}

			// Restore body for next handler
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
