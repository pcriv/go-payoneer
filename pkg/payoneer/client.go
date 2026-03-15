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

	// Services
	common   service
	Accounts *AccountsService
	Payouts  *PayoutsService
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

	// Wrap the HTTP client's transport with logging if a logger is provided.
	if c.Logger != nil {
		next := c.HTTPClient.Transport
		if next == nil {
			next = http.DefaultTransport
		}
		c.HTTPClient.Transport = &transport.LoggingTransport{
			Next:   next,
			Logger: c.Logger,
		}
	}

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

	// Preserve timeout and wrap the new client's transport with logging.
	httpClient.Timeout = c.HTTPClient.Timeout
	if c.Logger != nil {
		next := httpClient.Transport
		if next == nil {
			next = http.DefaultTransport
		}
		httpClient.Transport = &transport.LoggingTransport{
			Next:   next,
			Logger: c.Logger,
		}
	}

	c.HTTPClient = httpClient
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
