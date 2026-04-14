package payoneer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WebhookEvent represents a Payoneer webhook notification (IPCN).
type WebhookEvent struct {
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	Timestamp string          `json:"timestamp"`
	Content   json.RawMessage `json:"content"`
}

// Application-Name values Payoneer uses in the Authorization header.
const (
	AppNameSandbox    = "Webhook-IPCNSender-sandbox"
	AppNameProduction = "Webhook-IPCNSender-production"

	authScheme       = "hmacauth"
	defaultMaxSkew   = 5 * time.Minute
	defaultMaxBody   = 1 << 20 // 1 MB
	authHeaderFields = 4       // AppName, Signature, Nonce, Timestamp
)

// NonceStore records seen webhook nonces for replay protection. Implementations
// must be safe for concurrent use. Remember must return true only the first
// time a nonce is observed; subsequent calls with the same nonce return false.
// Entries older than the caller's accepted clock-skew window may be evicted.
type NonceStore interface {
	Remember(nonce string, ts time.Time) bool
}

// WebhookConfig configures Payoneer webhook signature verification.
type WebhookConfig struct {
	// DisableSignatureVerification turns off HMAC/timestamp/nonce checks.
	// Use only when relying on IP whitelisting and TLS for authenticity, or
	// when Payoneer HMAC authentication has not been provisioned for the
	// integration. When true, Secret/ExpectedAppName/MaxClockSkew/NonceStore
	// are ignored and the Authorization header is not required.
	DisableSignatureVerification bool
	// Secret is the shared HMAC key provisioned by Payoneer.
	Secret string
	// ExpectedAppName, if set, requires the Authorization header's
	// Application-Name component to match exactly (e.g. AppNameProduction).
	ExpectedAppName string
	// MaxClockSkew bounds how far the webhook timestamp may drift from Now.
	// Zero selects the default (5 minutes); a negative value disables the check.
	MaxClockSkew time.Duration
	// NonceStore, if set, provides replay protection; requests whose nonce has
	// already been observed are rejected.
	NonceStore NonceStore
	// MaxBodyBytes caps the accepted body size. Zero selects 1 MB.
	MaxBodyBytes int64
	// Now overrides the clock for tests. Nil uses time.Now.
	Now func() time.Time
}

func (c WebhookConfig) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}

	return time.Now()
}

func (c WebhookConfig) maxBody() int64 {
	if c.MaxBodyBytes > 0 {
		return c.MaxBodyBytes
	}

	return defaultMaxBody
}

// AuthorizationParts holds the fields extracted from a Payoneer
// `Authorization: hmacauth <AppName>:<Signature>:<Nonce>:<Timestamp>` header.
type AuthorizationParts struct {
	AppName   string
	Signature string
	Nonce     string
	Timestamp string
}

// ParseAuthorizationHeader parses a Payoneer hmacauth Authorization header.
func ParseAuthorizationHeader(header string) (AuthorizationParts, error) {
	var parts AuthorizationParts

	scheme, rest, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok {
		return parts, errors.New("authorization header missing scheme")
	}

	if !strings.EqualFold(scheme, authScheme) {
		return parts, fmt.Errorf("unexpected authorization scheme %q", scheme)
	}

	fields := strings.Split(rest, ":")
	if len(fields) != authHeaderFields {
		return parts, fmt.Errorf("authorization header expected %d fields, got %d", authHeaderFields, len(fields))
	}

	parts = AuthorizationParts{
		AppName:   fields[0],
		Signature: fields[1],
		Nonce:     fields[2],
		Timestamp: fields[3],
	}

	if parts.AppName == "" || parts.Signature == "" || parts.Nonce == "" || parts.Timestamp == "" {
		return AuthorizationParts{}, errors.New("authorization header has empty field")
	}

	return parts, nil
}

// ComputeSignature returns the base64 HMAC-SHA256 of payload||nonce||timestamp
// keyed by secret, matching the Payoneer HMAC webhook specification.
func ComputeSignature(payload []byte, nonce, timestamp, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	h.Write([]byte(nonce))
	h.Write([]byte(timestamp))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifySignature constant-time compares signature against the expected HMAC
// for the given payload, nonce and timestamp.
func VerifySignature(payload []byte, nonce, timestamp, signature, secret string) bool {
	expected := ComputeSignature(payload, nonce, timestamp, secret)

	return hmac.Equal([]byte(signature), []byte(expected))
}

// verifyRequest performs full Authorization-header verification against cfg
// and returns the parsed parts on success.
func verifyRequest(body []byte, header string, cfg WebhookConfig) (AuthorizationParts, error) {
	parts, err := ParseAuthorizationHeader(header)
	if err != nil {
		return parts, err
	}

	if cfg.ExpectedAppName != "" && parts.AppName != cfg.ExpectedAppName {
		return parts, fmt.Errorf("unexpected application name %q", parts.AppName)
	}

	ts, err := strconv.ParseInt(parts.Timestamp, 10, 64)
	if err != nil {
		return parts, fmt.Errorf("invalid timestamp: %w", err)
	}

	skew := cfg.MaxClockSkew
	if skew == 0 {
		skew = defaultMaxSkew
	}

	if skew > 0 {
		delta := cfg.now().Sub(time.Unix(ts, 0))
		if delta < 0 {
			delta = -delta
		}

		if delta > skew {
			return parts, fmt.Errorf("timestamp outside allowed skew (%s)", delta)
		}
	}

	if !VerifySignature(body, parts.Nonce, parts.Timestamp, parts.Signature, cfg.Secret) {
		return parts, errors.New("invalid signature")
	}

	if cfg.NonceStore != nil && !cfg.NonceStore.Remember(parts.Nonce, time.Unix(ts, 0)) {
		return parts, errors.New("nonce already used")
	}

	return parts, nil
}

// ParseWebhook verifies the Authorization header, reads the body, and
// unmarshals the IPCN payload. When cfg.DisableSignatureVerification is true,
// the Authorization header is not required and no cryptographic checks run.
func ParseWebhook(r *http.Request, cfg WebhookConfig) (*WebhookEvent, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, cfg.maxBody()))
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()

	if !cfg.DisableSignatureVerification {
		header := r.Header.Get("Authorization")
		if header == "" {
			return nil, errors.New("missing Authorization header")
		}

		if _, verr := verifyRequest(body, header, cfg); verr != nil {
			return nil, verr
		}
	}

	var event WebhookEvent
	if uerr := json.Unmarshal(body, &event); uerr != nil {
		return nil, uerr
	}

	return &event, nil
}

// WebhookValidator returns middleware that verifies the Payoneer Authorization
// header, responds 401 on failure, and restores r.Body for downstream handlers.
// When cfg.DisableSignatureVerification is true the middleware still caps the
// body size and restores r.Body, but performs no cryptographic checks.
func WebhookValidator(cfg WebhookConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(io.LimitReader(r.Body, cfg.maxBody()))
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)

				return
			}
			_ = r.Body.Close()

			if !cfg.DisableSignatureVerification {
				header := r.Header.Get("Authorization")
				if header == "" {
					http.Error(w, "missing Authorization header", http.StatusUnauthorized)

					return
				}

				if _, verr := verifyRequest(body, header, cfg); verr != nil {
					http.Error(w, verr.Error(), http.StatusUnauthorized)

					return
				}
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
