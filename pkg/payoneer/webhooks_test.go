package payoneer

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const testSecret = "test-secret"

func fixedClock(ts int64) func() time.Time {
	return func() time.Time { return time.Unix(ts, 0) }
}

func signedRequest(t *testing.T, body, nonce string, ts int64, appName string) *http.Request {
	t.Helper()

	tsStr := strconv.FormatInt(ts, 10)
	sig := ComputeSignature([]byte(body), nonce, tsStr, testSecret)
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set("Authorization", "hmacauth "+appName+":"+sig+":"+nonce+":"+tsStr)

	return req
}

func TestParseAuthorizationHeader(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   AuthorizationParts
		err    bool
	}{
		{
			name:   "valid",
			header: "hmacauth App:Sig:Nonce:123",
			want:   AuthorizationParts{AppName: "App", Signature: "Sig", Nonce: "Nonce", Timestamp: "123"},
		},
		{name: "wrong scheme", header: "Bearer App:Sig:Nonce:123", err: true},
		{name: "missing fields", header: "hmacauth App:Sig:Nonce", err: true},
		{name: "empty field", header: "hmacauth App::Nonce:123", err: true},
		{name: "no scheme", header: "hmacauthonly", err: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAuthorizationHeader(tc.header)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"event_type":"payout_created"}`)
	nonce := "nonce-1"
	ts := "1700000000"
	sig := ComputeSignature(payload, nonce, ts, testSecret)

	if !VerifySignature(payload, nonce, ts, sig, testSecret) {
		t.Error("expected signature to verify")
	}
	if VerifySignature(payload, nonce, ts, sig, "wrong-secret") {
		t.Error("wrong secret should fail")
	}
	if VerifySignature(payload, "other-nonce", ts, sig, testSecret) {
		t.Error("different nonce should fail")
	}
	if VerifySignature(payload, nonce, "1700000001", sig, testSecret) {
		t.Error("different timestamp should fail")
	}
	if VerifySignature([]byte("tampered"), nonce, ts, sig, testSecret) {
		t.Error("tampered payload should fail")
	}
}

func TestParseWebhook(t *testing.T) {
	body := `{"event_type":"payout_created","event_id":"123","timestamp":"2023-01-01T00:00:00Z","content":{"status":"COMPLETED"}}`
	var ts int64 = 1700000000
	cfg := WebhookConfig{
		Secret:          testSecret,
		ExpectedAppName: AppNameSandbox,
		Now:             fixedClock(ts),
	}

	t.Run("valid webhook", func(t *testing.T) {
		req := signedRequest(t, body, "n1", ts, AppNameSandbox)

		event, err := ParseWebhook(req, cfg)
		if err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}
		if event.EventType != "payout_created" {
			t.Errorf("got event type %s, want payout_created", event.EventType)
		}
	})

	t.Run("missing Authorization", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		if _, err := ParseWebhook(req, cfg); err == nil {
			t.Error("expected error for missing header")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		req.Header.Set("Authorization", "hmacauth "+AppNameSandbox+":bad:nonce:"+strconv.FormatInt(ts, 10))
		if _, err := ParseWebhook(req, cfg); err == nil {
			t.Error("expected error for bad signature")
		}
	})

	t.Run("wrong app name", func(t *testing.T) {
		req := signedRequest(t, body, "n1", ts, AppNameProduction)
		if _, err := ParseWebhook(req, cfg); err == nil {
			t.Error("expected error for wrong app name")
		}
	})

	t.Run("timestamp skew exceeded", func(t *testing.T) {
		skewed := ts + int64((10 * time.Minute).Seconds())
		req := signedRequest(t, body, "n2", skewed, AppNameSandbox)
		if _, err := ParseWebhook(req, cfg); err == nil {
			t.Error("expected error for timestamp skew")
		}
	})

	t.Run("skew disabled accepts old timestamp", func(t *testing.T) {
		loose := cfg
		loose.MaxClockSkew = -1
		old := ts - int64((24 * time.Hour).Seconds())
		req := signedRequest(t, body, "n3", old, AppNameSandbox)
		if _, err := ParseWebhook(req, loose); err != nil {
			t.Errorf("expected success with skew disabled: %v", err)
		}
	})
}

func TestParseWebhookSignatureDisabled(t *testing.T) {
	body := `{"event_type":"payout_created"}`
	cfg := WebhookConfig{DisableSignatureVerification: true}

	t.Run("no Authorization header still accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		event, err := ParseWebhook(req, cfg)
		if err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}
		if event.EventType != "payout_created" {
			t.Errorf("got event type %s, want payout_created", event.EventType)
		}
	})

	t.Run("garbage Authorization header still accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		req.Header.Set("Authorization", "bogus")
		if _, err := ParseWebhook(req, cfg); err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}
	})

	t.Run("middleware passes through without header", func(t *testing.T) {
		handlerCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		rr := httptest.NewRecorder()
		WebhookValidator(cfg)(next).ServeHTTP(rr, req)
		if !handlerCalled || rr.Code != http.StatusOK {
			t.Errorf("handler not called or wrong status: called=%v code=%d", handlerCalled, rr.Code)
		}
	})
}

type memoryNonceStore struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func (s *memoryNonceStore) Remember(nonce string, _ time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen == nil {
		s.seen = map[string]struct{}{}
	}
	if _, ok := s.seen[nonce]; ok {
		return false
	}
	s.seen[nonce] = struct{}{}

	return true
}

func TestParseWebhookReplayProtection(t *testing.T) {
	body := `{"event_type":"payout_created"}`
	var ts int64 = 1700000000
	cfg := WebhookConfig{
		Secret:     testSecret,
		NonceStore: &memoryNonceStore{},
		Now:        fixedClock(ts),
	}

	first := signedRequest(t, body, "replay-nonce", ts, AppNameSandbox)
	if _, err := ParseWebhook(first, cfg); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	second := signedRequest(t, body, "replay-nonce", ts, AppNameSandbox)
	if _, err := ParseWebhook(second, cfg); err == nil {
		t.Error("expected replay rejection on second call")
	}
}

func TestWebhookMiddleware(t *testing.T) {
	body := `{"event_type":"payout_created"}`
	var ts int64 = 1700000000
	cfg := WebhookConfig{Secret: testSecret, Now: fixedClock(ts)}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != body {
			t.Errorf("got body %s, want %s", string(got), body)
		}
		w.WriteHeader(http.StatusOK)
	})
	mw := WebhookValidator(cfg)(nextHandler)

	t.Run("valid signature proceeds", func(t *testing.T) {
		req := signedRequest(t, body, "n1", ts, AppNameSandbox)
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want 200", rr.Code)
		}
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		req.Header.Set("Authorization", "hmacauth App:bad:nonce:"+strconv.FormatInt(ts, 10))
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want 401", rr.Code)
		}
	})

	t.Run("missing Authorization returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want 401", rr.Code)
		}
	})
}
