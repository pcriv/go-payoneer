package payoneer

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEventStructsMatchDocExamples(t *testing.T) {
	t.Run("PaymentRequestAccepted", func(t *testing.T) {
		body := `{"Payee Id": "testpayee123","IntPaymentId": "test payment ID 123"}`
		var ev PaymentRequestAcceptedEvent
		if err := json.Unmarshal([]byte(body), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.PayeeID != "testpayee123" || ev.IntPaymentID != "test payment ID 123" {
			t.Errorf("got %+v", ev)
		}
	})

	t.Run("CancelPayout", func(t *testing.T) {
		body := `{"Payee Id":"test345","IntPaymentId":"v2_51d1fece","Reason Code":"10009","Reason Description":"Action+cannot+be+performed+because+payee+is+inactive","Payment Amount":"357.65","Canceled Payment Date":"2022-01-17T20%3a03%3a05Z"}`
		var ev CancelPayoutEvent
		if err := json.Unmarshal([]byte(body), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.PayeeID != "test345" || ev.IntPaymentID != "v2_51d1fece" ||
			ev.ReasonCode != "10009" || ev.PaymentAmount != "357.65" ||
			ev.CanceledPaymentDate != "2022-01-17T20%3a03%3a05Z" {
			t.Errorf("got %+v", ev)
		}
	})

	t.Run("PayeeApproved", func(t *testing.T) {
		body := `{"Payee Id":"150002404758209","Payoneer Id":"1965321","Session Id":"976-150000409001425"}`
		var ev PayeeApprovedEvent
		if err := json.Unmarshal([]byte(body), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.PayeeID != "150002404758209" || ev.PayoneerID != "1965321" ||
			ev.SessionID != "976-150000409001425" {
			t.Errorf("got %+v", ev)
		}
	})

	t.Run("PayeeDeclined", func(t *testing.T) {
		body := `{"Payee Id":"7d26313074d0","Payoneer Id":"4791210","Session Id":"","Reason Description":"Incorrect information"}`
		var ev PayeeDeclinedEvent
		if err := json.Unmarshal([]byte(body), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.PayeeID != "7d26313074d0" || ev.PayoneerID != "4791210" ||
			ev.ReasonDescription != "Incorrect information" {
			t.Errorf("got %+v", ev)
		}
	})

	t.Run("PayoutSentToBank", func(t *testing.T) {
		body := `{"Payee Id":"testpayee123","Amount":"18.00","IntPaymentId":"test payment ID 123","Currency":"USD","Transaction Date":"2022-01-18T07:01:12Z"}`
		var ev PayoutSentToBankEvent
		if err := json.Unmarshal([]byte(body), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.PayeeID != "testpayee123" || ev.Amount != "18.00" ||
			ev.IntPaymentID != "test payment ID 123" || ev.Currency != "USD" ||
			ev.TransactionDate != "2022-01-18T07:01:12Z" {
			t.Errorf("got %+v", ev)
		}
	})
}

const (
	testSecret = "test-secret"
	testURL    = "https://example.com/webhooks"
)

func fixedClock(ts int64) func() time.Time {
	return func() time.Time { return time.Unix(ts, 0) }
}

// signedRequest builds a POST webhook request signed per Payoneer's
// specification. The request URL matches testURL so that default URL
// reconstruction (via r.Host + r.URL.RequestURI()) yields the same value used
// when signing.
func signedRequest(t *testing.T, body, nonce string, ts int64, appName string) *http.Request {
	t.Helper()

	tsStr := strconv.FormatInt(ts, 10)
	sig := ComputeSignature(http.MethodPost, testURL, []byte(body), appName, nonce, tsStr, testSecret)
	req := httptest.NewRequest(http.MethodPost, testURL, strings.NewReader(body))
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
	method := http.MethodPost
	u := testURL
	payload := []byte(`{"Payee Id":"p"}`)
	app := AppNameSandbox
	nonce := "nonce-1"
	ts := "1700000000"
	sig := ComputeSignature(method, u, payload, app, nonce, ts, testSecret)

	if !VerifySignature(method, u, payload, app, nonce, ts, sig, testSecret) {
		t.Error("expected signature to verify")
	}
	if VerifySignature(method, u, payload, app, nonce, ts, sig, "wrong-secret") {
		t.Error("wrong secret should fail")
	}
	if VerifySignature(method, u, payload, app, "other-nonce", ts, sig, testSecret) {
		t.Error("different nonce should fail")
	}
	if VerifySignature(method, u, payload, app, nonce, "1700000001", sig, testSecret) {
		t.Error("different timestamp should fail")
	}
	if VerifySignature(method, u, []byte("tampered"), app, nonce, ts, sig, testSecret) {
		t.Error("tampered payload should fail")
	}
	if VerifySignature(method, "https://example.com/other", payload, app, nonce, ts, sig, testSecret) {
		t.Error("different URL should fail")
	}
	if VerifySignature(http.MethodGet, u, payload, app, nonce, ts, sig, testSecret) {
		t.Error("different method should fail")
	}
	if VerifySignature(method, u, payload, "other-app", nonce, ts, sig, testSecret) {
		t.Error("different app name should fail")
	}
}

// TestPayoneerURLEncode pins percent-encoding of the reserved characters.
func TestPayoneerURLEncode(t *testing.T) {
	got := strings.ToUpper(payoneerURLEncode(";/?:@&=+$,#[]!'()*"))
	want := "%3B%2F%3F%3A%40%26%3D%2B%24%2C%23%5B%5D%21%27%28%29%2A"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseWebhook(t *testing.T) {
	body := `{"Payee Id":"test345","IntPaymentId":"v2_51d1fece","Reason Code":"10009","Reason Description":"inactive","Payment Amount":"357.65","Canceled Payment Date":"2026-04-14T06:41:51Z"}`
	var ts int64 = 1700000000
	cfg := WebhookConfig{
		Secret:          testSecret,
		ExpectedAppName: AppNameSandbox,
		Now:             fixedClock(ts),
	}

	t.Run("valid webhook", func(t *testing.T) {
		req := signedRequest(t, body, "n1", ts, AppNameSandbox)

		wh, err := ParseWebhook(req, cfg)
		if err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}

		var ev CancelPayoutEvent
		if derr := wh.Decode(&ev); derr != nil {
			t.Fatalf("decode failed: %v", derr)
		}
		if ev.PayeeID != "test345" || ev.IntPaymentID != "v2_51d1fece" || ev.ReasonCode != "10009" {
			t.Errorf("unexpected event: %+v", ev)
		}
		if wh.Auth.AppName != AppNameSandbox {
			t.Errorf("got app name %q, want %q", wh.Auth.AppName, AppNameSandbox)
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
	body := `{"Payee Id":"p1","IntPaymentId":"ip1"}`
	cfg := WebhookConfig{DisableSignatureVerification: true}

	t.Run("no Authorization header still accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
		wh, err := ParseWebhook(req, cfg)
		if err != nil {
			t.Fatalf("ParseWebhook failed: %v", err)
		}

		var ev PaymentRequestAcceptedEvent
		if derr := wh.Decode(&ev); derr != nil {
			t.Fatalf("decode failed: %v", derr)
		}
		if ev.PayeeID != "p1" || ev.IntPaymentID != "ip1" {
			t.Errorf("unexpected event: %+v", ev)
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
