package payoneer

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/pcriv/go-payoneer/internal/auth"
	"github.com/pcriv/go-payoneer/internal/transport"
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

// WithLogger sets the logger for the Client and wraps it with a RedactionHandler.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		redactor := transport.NewRedactionHandler(
			logger.Handler(),
			[]string{"Authorization"},
			[]string{"client_secret", "access_token", "refresh_token", "client_id"},
		)
		c.Logger = slog.New(redactor)
	}
}

// WithAuthBaseURL sets the base URL for the Payoneer OAuth2 endpoints.
func WithAuthBaseURL(url string) Option {
	return func(c *Client) {
		c.AuthBaseURL = url
	}
}

// WithSandbox configures the client to use the Payoneer sandbox environment.
func WithSandbox() Option {
	return func(c *Client) {
		c.BaseURL = SandboxBaseURL
		c.AuthBaseURL = SandboxAuthBaseURL
	}
}

// WithRetries sets the maximum number of retries for failed requests.
func WithRetries(max int) Option {
	return func(c *Client) {
		c.retryMax = max
	}
}

// WithRetryWait sets the minimum and maximum wait time between retries.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Client) {
		c.retryWaitMin = min
		c.retryWaitMax = max
	}
}

// WithTracerProvider sets the OpenTelemetry tracer provider for the Client.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Client) {
		c.tracerProvider = tp
	}
}

// WithMeterProvider sets the OpenTelemetry meter provider for the Client.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(c *Client) {
		c.meterProvider = mp
	}
}

// WithScopes sets the OAuth 2.0 scopes to request during authentication.
func WithScopes(scopes ...string) Option {
	return func(c *Client) {
		c.scopes = scopes
	}
}

// WithClientCredentials configures the client to use OAuth 2.0 Client Credentials flow.
func WithClientCredentials(clientID, clientSecret string) Option {
	return func(c *Client) {
		c.authFn = func(ctx context.Context, c *Client) (*http.Client, error) {
			return auth.NewClientCredentialsClient(ctx, c.AuthBaseURL, clientID, clientSecret, c.scopes, c.tokenStore)
		}
	}
}

// WithAuthCode configures the client to use OAuth 2.0 Authorization Code flow.
func WithAuthCode(clientID, clientSecret, code, redirectURL string) Option {
	return func(c *Client) {
		c.authFn = func(ctx context.Context, c *Client) (*http.Client, error) {
			return auth.NewAuthCodeClient(ctx, c.AuthBaseURL, clientID, clientSecret, code, redirectURL, c.scopes, c.tokenStore)
		}
	}
}

// WithTokenStore configures the client to use a custom TokenStore.
func WithTokenStore(store auth.TokenStore) Option {
	return func(c *Client) {
		c.tokenStore = store
	}
}

// WithProgramID sets the Program ID for the Client.
func WithProgramID(id string) Option {
	return func(c *Client) {
		c.ProgramID = id
	}
}
