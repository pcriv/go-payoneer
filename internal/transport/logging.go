package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
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
		t.Logger.Error("request failed",
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
	attrs := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
	}

	// Log headers (will be redacted by the RedactionHandler if configured)
	for name, values := range req.Header {
		attrs = append(attrs, slog.String(name, strings.Join(values, ", ")))
	}

	// Log body if it's small and not a stream
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 { // Only log small bodies
				var bodyMap map[string]any
				if uerr := json.Unmarshal(body, &bodyMap); uerr == nil {
					// Add body fields as a group for the RedactionHandler to process
					attrs = append(attrs, slog.Any("body", bodyMap))
				} else {
					attrs = append(attrs, slog.String("body", string(body)))
				}
			}
		}
	}

	t.Logger.LogAttrs(req.Context(), slog.LevelInfo, "sending request", attrs...)
}

func (t *LoggingTransport) logResponse(resp *http.Response, duration time.Duration) {
	attrs := []slog.Attr{
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
	}

	// Log headers
	for name, values := range resp.Header {
		attrs = append(attrs, slog.String(name, strings.Join(values, ", ")))
	}

	// Log body
	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 {
				var bodyMap map[string]any
				if uerr := json.Unmarshal(body, &bodyMap); uerr == nil {
					attrs = append(attrs, slog.Any("body", bodyMap))
				} else {
					attrs = append(attrs, slog.String("body", string(body)))
				}
			}
		}
	}

	t.Logger.LogAttrs(resp.Request.Context(), slog.LevelInfo, "received response", attrs...)
}
