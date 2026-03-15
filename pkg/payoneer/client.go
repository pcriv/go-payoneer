package payoneer

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/pablocrivella/go-payoneer/internal/auth"
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
	common   service // Reuse a single struct instead of allocating one for each service on the heap.
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

	c.common.client = c
	c.Accounts = (*AccountsService)(&c.common)
	c.Payouts = (*PayoutsService)(&c.common)

	return c
}
