package transport

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// LoggingTransport is an http.RoundTripper that logs requests and responses.
type LoggingTransport struct {
	Next   http.RoundTripper
	Logger *slog.Logger
}

// RoundTrip logs the request and response.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log Request
	t.logRequest(req)

	resp, err := t.Next.RoundTrip(req)
	if err != nil {
		t.Logger.Error("Request failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err),
		)
		return nil, err
	}

	// Log Response
	t.logResponse(resp, time.Since(start))

	return resp, nil
}

func (t *LoggingTransport) logRequest(req *http.Request) {
	attrs := []any{
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
	}

	// Log headers (will be redacted by the RedactionHandler if configured)
	for name, values := range req.Header {
		attrs = append(attrs, slog.String(name, values[0]))
	}

	// Log body if it's small and not a stream (simplified for now)
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 { // Only log small bodies
				attrs = append(attrs, slog.String("body", string(body)))
			}
		}
	}

	t.Logger.Info("Sending Request", attrs...)
}

func (t *LoggingTransport) logResponse(resp *http.Response, duration time.Duration) {
	attrs := []any{
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
	}

	// Log headers
	for name, values := range resp.Header {
		attrs = append(attrs, slog.String(name, values[0]))
	}

	// Log body
	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 {
				attrs = append(attrs, slog.String("body", string(body)))
			}
		}
	}

	t.Logger.Info("Received Response", attrs...)
}
