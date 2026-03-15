package payoneer

import (
	"log/slog"
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets the base URL for the Payoneer API.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.BaseURL = url
	}
}

// WithHTTPClient sets the HTTP client to use for requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

// WithTimeout sets the timeout for the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{}
		}
		c.HTTPClient.Timeout = timeout
	}
}

// WithLogger sets the logger for the Client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.Logger = logger
	}
}
