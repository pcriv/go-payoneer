package payoneer_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestClient_LogRedaction(t *testing.T) {
	// 1. Setup a buffer to capture logs
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	// 2. Mock Server that expects sensitive data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "secret_token_in_body", "status": "Success"}`))
	}))
	defer server.Close()

	ctx := context.Background()
	client := payoneer.NewClient(
		payoneer.WithBaseURL(server.URL),
		payoneer.WithLogger(logger),
	)

	// 3. Perform a request with a sensitive header
	req, _ := client.NewRequest(ctx, http.MethodGet, "/v4/accounts", nil)
	req.Header.Set("Authorization", "Bearer sensitive_token_in_header")

	err := client.Do(req, nil)
	if err != nil {
		t.Fatalf("Do() failed: %v", err)
	}

	// 4. Verify logs
	logOutput := logBuf.String()

	// Check Header Redaction
	if strings.Contains(logOutput, "sensitive_token_in_header") {
		t.Errorf("Logs contain sensitive token from header: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Authorization=[REDACTED]") && !strings.Contains(logOutput, "authorization=[REDACTED]") {
		t.Errorf("Logs do not contain redacted Authorization header: %s", logOutput)
	}

	// Check Body Redaction (access_token field)
	if strings.Contains(logOutput, "secret_token_in_body") {
		t.Errorf("Logs contain sensitive token from body: %s", logOutput)
	}
	if !strings.Contains(logOutput, "access_token=[REDACTED]") {
		t.Errorf("Logs do not contain redacted access_token body field: %s", logOutput)
	}
}
