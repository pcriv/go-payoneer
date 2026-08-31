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
	Next            http.RoundTripper
	Logger          *slog.Logger
	RedactedHeaders []string
	RedactedFields  []string
}

// RoundTrip logs the request and response.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log Request
	t.logRequest(req)

	resp, err := t.Next.RoundTrip(req)
	if err != nil {
		t.Logger.ErrorContext(req.Context(), "request failed",
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

func (t *LoggingTransport) isRedactedHeader(name string) bool {
	lower := strings.ToLower(name)
	for _, h := range t.RedactedHeaders {
		if strings.ToLower(h) == lower {
			return true
		}
	}

	return false
}

func (t *LoggingTransport) isRedactedField(name string) bool {
	lower := strings.ToLower(name)
	for _, f := range t.RedactedFields {
		if strings.ToLower(f) == lower {
			return true
		}
	}

	return false
}

func (t *LoggingTransport) redactMap(m map[string]any) map[string]any {
	if len(t.RedactedFields) == 0 {
		return m
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if t.isRedactedField(k) {
			out[k] = "[REDACTED]"

			continue
		}

		switch typedValue := v.(type) {
		case map[string]any:
			out[k] = t.redactMap(typedValue)
		case []any:
			var sliceOut []any
			for _, item := range typedValue {
				if itemMap, ok := item.(map[string]any); ok {
					sliceOut = append(sliceOut, t.redactMap(itemMap))
				} else {
					sliceOut = append(sliceOut, item)
				}
			}
			out[k] = sliceOut
		default:
			out[k] = v
		}
	}

	return out
}

func (t *LoggingTransport) logRequest(req *http.Request) {
	attrs := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
	}

	for name, values := range req.Header {
		if t.isRedactedHeader(name) {
			attrs = append(attrs, slog.String(name, "[REDACTED]"))
		} else {
			attrs = append(attrs, slog.String(name, strings.Join(values, ", ")))
		}
	}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 {
				var bodyMap map[string]any
				if uerr := json.Unmarshal(body, &bodyMap); uerr == nil {
					attrs = append(attrs, slog.Any("body", t.redactMap(bodyMap)))
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

	for name, values := range resp.Header {
		if t.isRedactedHeader(name) {
			attrs = append(attrs, slog.String(name, "[REDACTED]"))
		} else {
			attrs = append(attrs, slog.String(name, strings.Join(values, ", ")))
		}
	}

	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) < 1024 {
				var bodyMap map[string]any
				if uerr := json.Unmarshal(body, &bodyMap); uerr == nil {
					attrs = append(attrs, slog.Any("body", t.redactMap(bodyMap)))
				} else {
					attrs = append(attrs, slog.String("body", string(body)))
				}
			}
		}
	}

	t.Logger.LogAttrs(resp.Request.Context(), slog.LevelInfo, "received response", attrs...)
}
