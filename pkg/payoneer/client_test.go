package payoneer_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/pablocrivella/go-payoneer/pkg/payoneer"
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
		if client.HTTPClient.Timeout != timeout {
			t.Errorf("expected timeout to be %s, got %s", timeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("Custom Logger", func(t *testing.T) {
		logger := slog.Default()
		client := payoneer.NewClient(payoneer.WithLogger(logger))
		if client.Logger != logger {
			t.Error("expected logger to be custom logger")
		}
	})

	t.Run("Service initialization", func(t *testing.T) {
		client := payoneer.NewClient()
		if client.Accounts == nil {
			t.Error("expected Accounts service to be initialized")
		}
		if client.Payouts == nil {
			t.Error("expected Payouts service to be initialized")
		}
	})
}
