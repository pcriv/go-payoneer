package payoneer

import (
	"log/slog"
	"net/http"
	"time"
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
}

// NewClient returns a new Payoneer Client with the provided options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
