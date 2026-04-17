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
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/pcriv/go-payoneer/internal/auth"
	"github.com/pcriv/go-payoneer/internal/transport"
)

const (
	// DefaultBaseURL is the production Payoneer API URL.
	DefaultBaseURL = "https://api.payoneer.com"
	// SandboxBaseURL is the sandbox Payoneer API URL.
	SandboxBaseURL = "https://api.sandbox.payoneer.com"
	// DefaultAuthBaseURL is the production Payoneer OAuth2 URL.
	DefaultAuthBaseURL = "https://login.payoneer.com"
	// SandboxAuthBaseURL is the sandbox Payoneer OAuth2 URL.
	SandboxAuthBaseURL = "https://login.sandbox.payoneer.com"
)

// Client is the main SDK client for communicating with the Payoneer API.
type Client struct {
	BaseURL     string
	AuthBaseURL string
	Logger      *slog.Logger

	// httpClient is swapped atomically when Authenticate runs, so any number
	// of goroutines can call HTTPClient() or Do concurrently without a race.
	httpClient atomic.Pointer[http.Client]

	tokenStore auth.TokenStore
	authFn     func(ctx context.Context, c *Client) (*http.Client, error)
	scopes     []string

	authMu   sync.Mutex
	authDone bool

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
	common  service
	Payouts *PayoutsService
	Payees  *PayeesService
}

// retryableLogger is a wrapper around slog.Logger that implements retryablehttp.LeveledLogger.
type retryableLogger struct {
	l *slog.Logger
}

func (r *retryableLogger) Error(msg string, keysAndValues ...any) {
	r.l.Error(msg, keysAndValues...)
}

func (r *retryableLogger) Info(msg string, keysAndValues ...any) {
	r.l.Info(msg, keysAndValues...)
}

func (r *retryableLogger) Debug(msg string, keysAndValues ...any) {
	r.l.Debug(msg, keysAndValues...)
}

func (r *retryableLogger) Warn(msg string, keysAndValues ...any) {
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
		BaseURL:     DefaultBaseURL,
		AuthBaseURL: DefaultAuthBaseURL,
		Logger:      slog.Default(),
		tokenStore:  auth.NewInMemoryStore(),
	}
	c.httpClient.Store(&http.Client{Timeout: 30 * time.Second})

	for _, opt := range opts {
		opt(c)
	}

	if c.tracerProvider != nil {
		c.tracer = c.tracerProvider.Tracer("go-payoneer")
	}

	c.httpClient.Store(c.wrapTransport(c.httpClient.Load()))

	c.common.client = c
	c.Payouts = (*PayoutsService)(&c.common)
	c.Payees = (*PayeesService)(&c.common)

	return c
}

// HTTPClient returns the underlying *http.Client. The returned pointer may be
// swapped by Authenticate, so callers should invoke this method each time
// they need the current client rather than caching the pointer.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient.Load()
}

// Authenticate eagerly performs the configured OAuth 2.0 flow. Calling it is
// optional: Do performs the same work lazily on the first request. Use this
// when you want credential errors to surface at startup rather than on first
// API call. Subsequent calls after a successful authentication are no-ops;
// a failed attempt is not cached and the next call retries.
func (c *Client) Authenticate(ctx context.Context) error {
	return c.ensureAuthenticated(ctx)
}

// ensureAuthenticated runs authFn at most once successfully. Concurrent
// callers block until the first attempt completes. If authFn fails, authDone
// stays false so the next call retries.
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	if c.authFn == nil {
		return nil
	}

	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.authDone {
		return nil
	}

	httpClient, err := c.authFn(ctx, c)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
	}

	httpClient.Timeout = c.httpClient.Load().Timeout
	c.httpClient.Store(c.wrapTransport(httpClient))
	c.authDone = true

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
		if e := json.NewEncoder(buf).Encode(body); e != nil {
			return nil, e
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
	req.Header.Set("User-Agent", "go-payoneer-sdk/1.0.0")

	return req, nil
}

// apiResult is a generic envelope for Payoneer API responses that wrap
// their payload under a "result" key.
type apiResult[T any] struct {
	Result T `json:"result"`
}

// Do executes an API request and parses the response into v.
// It automatically validates the response using transport-level validation.
// If the client was configured with an OAuth flow and Authenticate has not
// run successfully yet, Do authenticates lazily on first use.
func (c *Client) Do(req *http.Request, v any) error {
	if err := c.ensureAuthenticated(req.Context()); err != nil {
		return err
	}

	// #nosec G704
	resp, err := c.httpClient.Load().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if verr := validateResponse(resp); verr != nil {
		return verr
	}

	if v != nil {
		if derr := json.NewDecoder(resp.Body).Decode(v); derr != nil {
			return derr
		}
	}

	return nil
}
