package transport_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pcriv/go-payoneer/internal/transport"
)

func TestLoggingTransport_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "abc123")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	lt := &transport.LoggingTransport{
		Next:   http.DefaultTransport,
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want 200", resp.StatusCode)
	}

	// Verify body is still readable after logging
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"status":"ok"}` {
		t.Errorf("got body %q, want {\"status\":\"ok\"}", string(body))
	}

	logs := buf.String()
	if !strings.Contains(logs, "sending request") {
		t.Error("expected 'sending request' in logs")
	}
	if !strings.Contains(logs, "received response") {
		t.Error("expected 'received response' in logs")
	}
}

func TestLoggingTransport_RequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body to verify it was preserved after logging
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"key":"value"}` {
			t.Errorf("server got body %q, want {\"key\":\"value\"}", string(body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	lt := &transport.LoggingTransport{
		Next:   http.DefaultTransport,
		Logger: logger,
	}

	reqBody := bytes.NewBufferString(`{"key":"value"}`)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/test", reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestLoggingTransport_MultiValueHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	lt := &transport.LoggingTransport{
		Next:   http.DefaultTransport,
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	logs := buf.String()
	// Both cookie values should be logged (joined with ", ")
	if !strings.Contains(logs, "a=1") || !strings.Contains(logs, "b=2") {
		t.Errorf("expected multi-value headers in logs, got: %s", logs)
	}
}

func TestLoggingTransport_ErrorResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	lt := &transport.LoggingTransport{
		Next:   http.DefaultTransport,
		Logger: logger,
	}

	// Request to an invalid address to trigger a transport error
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:1", nil)
	_, err := lt.RoundTrip(req) //nolint:bodyclose // error path returns nil response
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	logs := buf.String()
	if !strings.Contains(logs, "request failed") {
		t.Error("expected 'request failed' in logs")
	}
}
