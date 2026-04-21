package payoneer

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // required by Payoneer signature spec (body digest)
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Webhook is the result of a verified Payoneer webhook request. Payoneer
// webhooks do not carry a common envelope: the event type is implied by the
// endpoint URL the notification was sent to, and the body contains only
// event-specific fields. Callers should unmarshal Body into the typed event
// struct that matches the endpoint (for example CancelPayoutEvent).
type Webhook struct {
	// Body is the raw JSON request body.
	Body []byte
	// Auth holds the parsed Authorization header; zero value when signature
	// verification is disabled.
	Auth AuthorizationParts
}

// Decode unmarshals the webhook body into v.
func (w *Webhook) Decode(v any) error {
	return json.Unmarshal(w.Body, v)
}

// PaymentRequestAcceptedEvent is the payload of the "Payment Request Accepted"
// webhook — triggered when Payoneer registers a payment request.
type PaymentRequestAcceptedEvent struct {
	PayeeID      string `json:"Payee Id"`
	IntPaymentID string `json:"IntPaymentId"`
}

// CancelPayoutEvent is the payload of the "Cancel Payout" webhook — triggered
// after a payment is cancelled on Payoneer's platform.
//
// Note: Payoneer's published examples show PaymentAmount and
// CanceledPaymentDate as strings, occasionally URL-percent-encoded
// (e.g. "2022-01-17T20%3a03%3a05Z"); production payloads may arrive
// unencoded. Callers that need typed values should URL-decode and parse after
// unmarshalling.
type CancelPayoutEvent struct {
	PayeeID             string `json:"Payee Id"`
	IntPaymentID        string `json:"IntPaymentId"`
	ReasonCode          string `json:"Reason Code"`
	ReasonDescription   string `json:"Reason Description"`
	PaymentAmount       string `json:"Payment Amount"`
	CanceledPaymentDate string `json:"Canceled Payment Date"`
}

// PayoutSentToBankEvent is the payload of the "Sent to bank" webhook
// (aka "Load Money to Bank / iACH") — triggered after Payoneer submits
// instructions to transfer funds to a payee's local bank account.
// See https://developer.payoneer.com/docs/mass-payouts-and-services.html#/db962d857712f-load-bank.
type PayoutSentToBankEvent struct {
	PayeeID         string `json:"Payee Id"`
	Amount          string `json:"Amount"`
	IntPaymentID    string `json:"IntPaymentId"`
	Currency        string `json:"Currency"`
	TransactionDate string `json:"Transaction Date"`
}

// AccountCardLoadedEvent is the payload of the "Account/Card Loaded
// Confirmation" webhook — triggered when funds are loaded onto a payee's card.
// See https://developer.payoneer.com/docs/mass-payouts-v4.html#/60655d18fabdd-account-card-loaded.
type AccountCardLoadedEvent struct {
	PayeeID         string `json:"Payee Id"`
	Amount          string `json:"Amount"`
	IntPaymentID    string `json:"IntPaymentId"`
	Currency        string `json:"Currency"`
	TransactionDate string `json:"Transaction Date"`
}

// PayeeApprovedEvent is the payload of the "Approved" webhook — triggered
// when Payoneer approves a payee application.
type PayeeApprovedEvent struct {
	PayeeID    string `json:"Payee Id"`
	PayoneerID string `json:"Payoneer Id"`
	SessionID  string `json:"Session Id"`
}

// PayeeDeclinedEvent is the payload of the "Declined" webhook — triggered
// when Payoneer declines a payee application.
type PayeeDeclinedEvent struct {
	PayeeID           string `json:"Payee Id"`
	PayoneerID        string `json:"Payoneer Id"`
	SessionID         string `json:"Session Id"`
	ReasonDescription string `json:"Reason Description"`
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
	// URLBuilder, if set, reconstructs the full URL that Payoneer signed
	// against. When nil the default uses X-Forwarded-Proto (or r.TLS) for the
	// scheme, X-Forwarded-Host (or r.Host) for the host, and r.URL.RequestURI()
	// for path+query. Override when your deployment rewrites paths or uses a
	// different proxy header scheme.
	URLBuilder func(*http.Request) string
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

// ComputeSignature returns the base64 HMAC-SHA256 of the Payoneer signing
// string for the given request, validated byte-for-byte against a real
// Payoneer sandbox webhook. The signing string is:
//
//	AppName + Method + EncodedURL + Timestamp + Nonce + BodyDigest
//
// where BodyDigest is "" for non-POST requests and base64(MD5(body)) otherwise,
// and EncodedURL is the request URL normalized (lowercased, query parameters
// sorted alphabetically), URL-encoded, then lowercased again.
func ComputeSignature(method, requestURL string, body []byte, appName, nonce, timestamp, secret string) string {
	method = strings.ToUpper(method)

	bodyDigest := ""
	if method == http.MethodPost && body != nil {
		sum := md5.Sum(body) //nolint:gosec // spec-mandated
		bodyDigest = base64.StdEncoding.EncodeToString(sum[:])
	}

	encodedURL := strings.ToLower(payoneerURLEncode(normalizeURI(requestURL)))
	toSign := appName + method + encodedURL + timestamp + nonce + bodyDigest

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(toSign))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifySignature constant-time compares signature against the expected HMAC
// for the given request components.
func VerifySignature(method, requestURL string, body []byte, appName, nonce, timestamp, signature, secret string) bool {
	expected := ComputeSignature(method, requestURL, body, appName, nonce, timestamp, secret)

	return hmac.Equal([]byte(signature), []byte(expected))
}

// normalizeURI lowercases the URI and, when a query string is present, splits
// its parameters on "&", sorts them alphabetically, and rejoins.
func normalizeURI(raw string) string {
	lower := strings.ToLower(raw)

	path, query, hasQuery := strings.Cut(lower, "?")
	if !hasQuery {
		return lower
	}

	params := strings.Split(query, "&")
	sort.Strings(params)

	return path + "?" + strings.Join(params, "&")
}

// payoneerURLEncode applies the encoding Payoneer uses inside the signing
// string: unreserved set A-Z, a-z, 0-9 and "-_.*", space → "+". Go's
// url.QueryEscape follows RFC 3986 (unreserved includes "~", encodes "*"),
// so we emit "%7E" for "~"; url.QueryEscape already emits "%2A" for "*".
func payoneerURLEncode(s string) string {
	e := url.QueryEscape(s)

	return strings.ReplaceAll(e, "~", "%7E")
}

// verifyRequest performs full Authorization-header verification against cfg
// and returns the parsed parts on success.
func verifyRequest(method, requestURL string, body []byte, header string, cfg WebhookConfig) (AuthorizationParts, error) {
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

	if !VerifySignature(method, requestURL, body, parts.AppName, parts.Nonce, parts.Timestamp, parts.Signature, cfg.Secret) {
		return parts, errors.New("invalid signature")
	}

	if cfg.NonceStore != nil && !cfg.NonceStore.Remember(parts.Nonce, time.Unix(ts, 0)) {
		return parts, errors.New("nonce already used")
	}

	return parts, nil
}

// defaultURLBuilder reconstructs the URL Payoneer signed against from a
// received request, preferring proxy headers when present.
func defaultURLBuilder(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	host := r.Host
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}

	return scheme + "://" + host + r.URL.RequestURI()
}

func (c WebhookConfig) buildURL(r *http.Request) string {
	if c.URLBuilder != nil {
		return c.URLBuilder(r)
	}

	return defaultURLBuilder(r)
}

// ParseWebhook verifies the Authorization header and reads the body. The
// returned Webhook exposes the raw JSON (to be decoded into an event struct
// matching the endpoint) and the parsed Authorization parts. When
// cfg.DisableSignatureVerification is true the Authorization header is not
// required and no cryptographic checks run; Webhook.Auth is then zero.
func ParseWebhook(r *http.Request, cfg WebhookConfig) (*Webhook, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, cfg.maxBody()))
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()

	wh := &Webhook{Body: body}

	if cfg.DisableSignatureVerification {
		return wh, nil
	}

	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errors.New("missing Authorization header")
	}

	parts, verr := verifyRequest(r.Method, cfg.buildURL(r), body, header, cfg)
	if verr != nil {
		return nil, verr
	}

	wh.Auth = parts

	return wh, nil
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

				if _, verr := verifyRequest(r.Method, cfg.buildURL(r), body, header, cfg); verr != nil {
					http.Error(w, verr.Error(), http.StatusUnauthorized)

					return
				}
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
