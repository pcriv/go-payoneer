package payoneer

import (
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if !ValidateSignature(body, signature, secret) {
		return nil, errors.New("invalid signature")
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}

	return &event, nil
}
