package transport

import (
	"context"
	"log/slog"
	"strings"
)

// RedactionHandler is a slog.Handler that redacts sensitive information.
type RedactionHandler struct {
	handler slog.Handler
	headers []string
	fields  []string
}

// NewRedactionHandler creates a new RedactionHandler.
func NewRedactionHandler(handler slog.Handler, headers []string, fields []string) *RedactionHandler {
	// Normalize headers and fields to lower case for case-insensitive matching
	normalizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		normalizedHeaders[i] = strings.ToLower(h)
	}

	normalizedFields := make([]string, len(fields))
	for i, f := range fields {
		normalizedFields[i] = strings.ToLower(f)
	}

	return &RedactionHandler{
		handler: handler,
		headers: normalizedHeaders,
		fields:  normalizedFields,
	}
}

// Enabled returns true if the handler is enabled for the given level.
func (h *RedactionHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle handles the record.
func (h *RedactionHandler) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, h.redactAttr(a))
		return true
	})

	// Create a new record with the redacted attributes
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(attrs...)

	return h.handler.Handle(ctx, newRecord)
}

// WithAttrs returns a new RedactionHandler with the given attributes.
func (h *RedactionHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redactedAttrs := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		redactedAttrs[i] = h.redactAttr(a)
	}
	return &RedactionHandler{
		handler: h.handler.WithAttrs(redactedAttrs),
		headers: h.headers,
		fields:  h.fields,
	}
}

// WithGroup returns a new RedactionHandler with the given group.
func (h *RedactionHandler) WithGroup(name string) slog.Handler {
	return &RedactionHandler{
		handler: h.handler.WithGroup(name),
		headers: h.headers,
		fields:  h.fields,
	}
}

func (h *RedactionHandler) redactAttr(a slog.Attr) slog.Attr {
	keyLower := strings.ToLower(a.Key)

	for _, header := range h.headers {
		if keyLower == header {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	for _, field := range h.fields {
		if keyLower == field {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	// Handle recursive redaction for Groups
	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		args := make([]any, len(groupAttrs))
		for i, ga := range groupAttrs {
			args[i] = h.redactAttr(ga)
		}
		return slog.Group(a.Key, args...)
	}

	return a
}
