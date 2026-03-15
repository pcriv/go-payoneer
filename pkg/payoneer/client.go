package payoneer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/pablocrivella/go-payoneer/internal/auth"
	"github.com/pablocrivella/go-payoneer/internal/transport"
)

const (
	// DefaultBaseURL is the production Payoneer API URL.
	DefaultBaseURL = "https://api.payoneer.com"
	// SandboxBaseURL is the sandbox Payoneer API URL.
	SandboxBaseURL = "https://api.sandbox.payoneer.com"
)

// Client is the main SDK client for communicating with the Payoneer API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger

	tokenStore auth.TokenStore
	authFn     func(ctx context.Context, c *Client) (*http.Client, error)

	// Retry configuration
	retryMax     int
	retryWaitMin time.Duration
	retryWaitMax time.Duration

	// OpenTelemetry
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	tracer         trace.Tracer

	// Program configuration
	ProgramID string

	// Services
	common   service
	Accounts *AccountsService
	Payouts  *PayoutsService
}

// retryableLogger is a wrapper around slog.Logger that implements retryablehttp.LeveledLogger.
type retryableLogger struct {
	l *slog.Logger
}

func (r *retryableLogger) Error(msg string, keysAndValues ...interface{}) {
	r.l.Error(msg, keysAndValues...)
}
func (r *retryableLogger) Info(msg string, keysAndValues ...interface{}) {
	r.l.Info(msg, keysAndValues...)
}
func (r *retryableLogger) Debug(msg string, keysAndValues ...interface{}) {
	r.l.Debug(msg, keysAndValues...)
}
func (r *retryableLogger) Warn(msg string, keysAndValues ...interface{}) {
	r.l.Warn(msg, keysAndValues...)
}

func (c *Client) wrapTransport(httpClient *http.Client) *http.Client {
	// 1. Base Transport (already in httpClient.Transport or DefaultTransport)
	next := httpClient.Transport
	if next == nil {
		next = http.DefaultTransport
	}

	// 2. OpenTelemetry Instrumentation
	if c.tracerProvider != nil || c.meterProvider != nil {
		opts := []otelhttp.Option{}
		if c.tracerProvider != nil {
			opts = append(opts, otelhttp.WithTracerProvider(c.tracerProvider))
		}
		if c.meterProvider != nil {
			opts = append(opts, otelhttp.WithMeterProvider(c.meterProvider))
		}
		next = otelhttp.NewTransport(next, opts...)
	}

	// 3. Retry Logic
	if c.retryMax > 0 {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = c.retryMax
		if c.retryWaitMin > 0 {
			retryClient.RetryWaitMin = c.retryWaitMin
		}
		if c.retryWaitMax > 0 {
			retryClient.RetryWaitMax = c.retryWaitMax
		}
		if c.Logger != nil {
			retryClient.Logger = &retryableLogger{l: c.Logger}
		}

		retryClient.HTTPClient.Transport = next
		retryClient.HTTPClient.Timeout = httpClient.Timeout

		httpClient = retryClient.StandardClient()
		next = httpClient.Transport
	}

	// 4. Logging Transport
	if c.Logger != nil {
		httpClient.Transport = &transport.LoggingTransport{
			Next:   next,
			Logger: c.Logger,
		}
	} else {
		httpClient.Transport = next
	}

	return httpClient
}

// NewClient returns a new Payoneer Client with the provided options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger:     slog.Default(),
		tokenStore: auth.NewInMemoryStore(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.tracerProvider != nil {
		c.tracer = c.tracerProvider.Tracer("go-payoneer")
	}

	c.HTTPClient = c.wrapTransport(c.HTTPClient)

	c.common.client = c
	c.Accounts = (*AccountsService)(&c.common)
	c.Payouts = (*PayoutsService)(&c.common)

	return c
}

// Authenticate triggers the OAuth 2.0 authentication flow if configured.
// It replaces the internal HTTP client with one that automatically handles token injection and refreshing.
func (c *Client) Authenticate(ctx context.Context) error {
	if c.authFn == nil {
		return nil
	}

	httpClient, err := c.authFn(ctx, c)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Preserve timeout and wrap the new client's transport.
	httpClient.Timeout = c.HTTPClient.Timeout
	c.HTTPClient = c.wrapTransport(httpClient)

	return nil
}

// NewRequest creates an authenticated API request.
func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-payoneer-sdk/0.1.0")

	return req, nil
}

// Do executes an API request and parses the response into v.
// It automatically validates the response using transport-level validation.
func (c *Client) Do(req *http.Request, v any) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := transport.ValidateResponse(resp); err != nil {
		return resp, err
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return resp, err
		}
	}

	return resp, nil
}
