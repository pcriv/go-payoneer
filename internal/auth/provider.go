package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Endpoints for Payoneer OAuth2.
func Endpoints(baseURL string) oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  baseURL + "/api/v2/oauth2/authorize",
		TokenURL: baseURL + "/api/v2/oauth2/token",
	}
}

// AuthProvider initializes and returns an authenticated RoundTripper.
type AuthProvider func(ctx context.Context, baseClient *http.Client) (http.RoundTripper, error)

// NewClientCredentialsProvider returns an AuthProvider for the Client Credentials flow.
func NewClientCredentialsProvider(
	baseURL, clientID, clientSecret string,
	scopes []string,
	store TokenStore,
) AuthProvider {
	return func(ctx context.Context, baseClient *http.Client) (http.RoundTripper, error) {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, baseClient)

		config := &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     Endpoints(baseURL).TokenURL,
			Scopes:       scopes,
		}

		ts := config.TokenSource(ctx)

		// Eagerly fetch initial token to validate credentials
		token, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to obtain token from %s: %w", config.TokenURL, err)
		}

		if store != nil {
			store.Set(token)
			ts = &storedTokenSource{
				inner: ts,
				store: store,
			}
		}

		return &oauth2.Transport{
			Source: ts,
			Base:   baseClient.Transport,
		}, nil
	}
}

// NewAuthCodeProvider returns an AuthProvider for the Authorization Code flow.
func NewAuthCodeProvider(
	baseURL, clientID, clientSecret, code, redirectURL string,
	scopes []string,
	store TokenStore,
) AuthProvider {
	return func(ctx context.Context, baseClient *http.Client) (http.RoundTripper, error) {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, baseClient)

		config := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     Endpoints(baseURL),
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		}

		var token *oauth2.Token
		var err error

		if store != nil {
			token = store.Get()
		}

		if token == nil {
			token, err = config.Exchange(ctx, code)
			if err != nil {
				return nil, err
			}
			if store != nil {
				store.Set(token)
			}
		}

		ts := config.TokenSource(ctx, token)
		if store != nil {
			ts = &storedTokenSource{
				inner: ts,
				store: store,
			}
		}

		return &oauth2.Transport{
			Source: ts,
			Base:   baseClient.Transport,
		}, nil
	}
}

type storedTokenSource struct {
	inner oauth2.TokenSource
	store TokenStore
}

func (s *storedTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.inner.Token()
	if err != nil {
		return nil, err
	}
	s.store.Set(t)

	return t, nil
}
