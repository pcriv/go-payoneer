package payoneer

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/pablocrivella/go-payoneer/internal/auth"
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

// WithClientCredentials configures the client to use OAuth 2.0 Client Credentials flow.
func WithClientCredentials(clientID, clientSecret string) Option {
	return func(c *Client) {
		c.authFn = func(ctx context.Context, c *Client) (*http.Client, error) {
			return auth.NewClientCredentialsClient(ctx, c.BaseURL, clientID, clientSecret, c.tokenStore), nil
		}
	}
}

// WithAuthCode configures the client to use OAuth 2.0 Authorization Code flow.
func WithAuthCode(clientID, clientSecret, code, redirectURL string) Option {
	return func(c *Client) {
		c.authFn = func(ctx context.Context, c *Client) (*http.Client, error) {
			return auth.NewAuthCodeClient(ctx, c.BaseURL, clientID, clientSecret, code, redirectURL, c.tokenStore)
		}
	}
}

// WithTokenStore configures the client to use a custom TokenStore.
func WithTokenStore(store auth.TokenStore) Option {
	return func(c *Client) {
		c.tokenStore = store
	}
}
