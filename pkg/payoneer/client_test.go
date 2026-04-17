package payoneer_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestNewClient(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		client := payoneer.NewClient()
		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
		if client.BaseURL != "https://api.payoneer.com" {
			t.Errorf("expected default BaseURL to be https://api.payoneer.com, got %s", client.BaseURL)
		}
		if client.AuthBaseURL != "https://login.payoneer.com" {
			t.Errorf("expected default AuthBaseURL to be https://login.payoneer.com, got %s", client.AuthBaseURL)
		}
	})

	t.Run("Custom BaseURL", func(t *testing.T) {
		url := "https://api.sandbox.payoneer.com"
		client := payoneer.NewClient(payoneer.WithBaseURL(url))
		if client.BaseURL != url {
			t.Errorf("expected BaseURL to be %s, got %s", url, client.BaseURL)
		}
	})

	t.Run("Custom Timeout", func(t *testing.T) {
		timeout := 10 * time.Second
		client := payoneer.NewClient(payoneer.WithTimeout(timeout))
		if client.HTTPClient().Timeout != timeout {
			t.Errorf("expected timeout to be %s, got %s", timeout, client.HTTPClient().Timeout)
		}
	})

	t.Run("Custom Logger", func(t *testing.T) {
		logger := slog.Default()
		client := payoneer.NewClient(payoneer.WithLogger(logger))
		if client.Logger == nil {
			t.Fatal("expected logger to be non-nil")
		}
		// The logger should be wrapped in a RedactionHandler
		if client.Logger == logger {
			t.Error("expected logger to be wrapped with RedactionHandler, not the original")
		}
	})

	t.Run("Service initialization", func(t *testing.T) {
		client := payoneer.NewClient()
		if client.Payouts == nil {
			t.Error("expected Payouts service to be initialized")
		}
	})

	t.Run("Custom ProgramID", func(t *testing.T) {
		programID := "123456"
		client := payoneer.NewClient(payoneer.WithProgramID(programID))
		if client.ProgramID != programID {
			t.Errorf("expected ProgramID to be %s, got %s", programID, client.ProgramID)
		}
	})
}
