package transport_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/pablocrivella/go-payoneer/internal/transport"
)

func TestRedactionHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	redactor := transport.NewRedactionHandler(handler, []string{"Authorization"}, []string{"client_secret", "access_token", "refresh_token"})
	logger := slog.New(redactor)

	t.Run("Redact Sensitive Header", func(t *testing.T) {
		buf.Reset()
		logger.Info("Request with Auth",
			slog.String("Authorization", "Bearer sensitive-token"),
			slog.String("User-Agent", "Go-SDK"),
		)

		output := buf.String()
		if bytes.Contains([]byte(output), []byte("sensitive-token")) {
			t.Errorf("Expected token to be redacted, got %s", output)
		}
		if !bytes.Contains([]byte(output), []byte("[REDACTED]")) {
			t.Errorf("Expected [REDACTED] in output, got %s", output)
		}
		if !bytes.Contains([]byte(output), []byte("Go-SDK")) {
			t.Errorf("Expected non-sensitive field to be present, got %s", output)
		}
	})

	t.Run("Redact Sensitive Body Fields", func(t *testing.T) {
		buf.Reset()
		logger.Info("Token Response",
			slog.String("access_token", "token-123"),
			slog.String("refresh_token", "refresh-456"),
			slog.String("client_secret", "secret-789"),
			slog.String("expires_in", "3600"),
		)

		output := buf.String()
		if bytes.Contains([]byte(output), []byte("token-123")) || bytes.Contains([]byte(output), []byte("refresh-456")) || bytes.Contains([]byte(output), []byte("secret-789")) {
			t.Errorf("Expected sensitive fields to be redacted, got %s", output)
		}
		if !bytes.Contains([]byte(output), []byte("3600")) {
			t.Errorf("Expected non-sensitive field to be present, got %s", output)
		}
	})

	t.Run("Case-insensitive Header Redaction", func(t *testing.T) {
		buf.Reset()
		logger.Info("Request with lower-case auth",
			slog.String("authorization", "bearer-token"),
		)

		output := buf.String()
		if bytes.Contains([]byte(output), []byte("bearer-token")) {
			t.Errorf("Expected token to be redacted, got %s", output)
		}
	})
}
