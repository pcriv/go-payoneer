package payoneer_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/pcriv/go-payoneer/pkg/payoneer"
)

func TestClientRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "Success"}`)
	}))
	defer ts.Close()

	client := payoneer.NewClient(
		payoneer.WithBaseURL(ts.URL),
		payoneer.WithRetries(3),
		payoneer.WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)

	ctx := context.Background()
	req, err := client.NewRequest(ctx, "GET", "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	err = client.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientOTel(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exp))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := payoneer.NewClient(
		payoneer.WithBaseURL(ts.URL),
		payoneer.WithTracerProvider(tp),
	)

	ctx := context.Background()
	req, err := client.NewRequest(ctx, "GET", "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	err = client.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) == 0 {
		t.Error("expected at least one span, got none")
	}

	found := false
	for _, span := range spans {
		if span.Name == "HTTP GET" {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected HTTP GET span, got %v", spans)
	}
}
