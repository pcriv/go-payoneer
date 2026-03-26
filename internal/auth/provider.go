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

// NewClientCredentialsClient returns an http.Client authenticated via Client Credentials flow.
// It eagerly fetches an initial token to validate the credentials.
func NewClientCredentialsClient(ctx context.Context, baseURL, clientID, clientSecret string, scopes []string, store TokenStore) (*http.Client, error) {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     Endpoints(baseURL).TokenURL,
		Scopes:       scopes,
	}

	ts := config.TokenSource(ctx)

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

	return oauth2.NewClient(ctx, ts), nil
}

// NewAuthCodeClient returns an http.Client authenticated via Authorization Code flow.
func NewAuthCodeClient(ctx context.Context, baseURL, clientID, clientSecret, code, redirectURL string, scopes []string, store TokenStore) (*http.Client, error) {
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

	return oauth2.NewClient(ctx, ts), nil
}

type storedTokenSource struct {
	inner oauth2.TokenSource
	store TokenStore
}

func (s *storedTokenSource) Token() (*oauth2.Token, error) {
	// If store has a valid token, we could return it, but oauth2.ReuseTokenSource
	// is usually better. However, our goal is to keep the store updated.

	t, err := s.inner.Token()
	if err != nil {
		return nil, err
	}

	// We store the token every time it's retrieved from the inner source.
	// x/oauth2's TokenSource returned by config.TokenSource will refresh it.
	s.store.Set(t)

	return t, nil
}
