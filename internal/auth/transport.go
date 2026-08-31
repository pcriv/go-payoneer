package auth

import (
	"context"
	"net/http"
	"sync"
)

// LazyAuthTransport lazily initializes the OAuth2 transport on the first request.
type LazyAuthTransport struct {
	Base         http.RoundTripper
	Provider     AuthProvider
	ClientParams *http.Client

	mu       sync.Mutex
	authNext http.RoundTripper
}

// RoundTrip executes a single HTTP transaction, initializing auth on the first call.
func (t *LazyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Provider == nil {
		return t.Base.RoundTrip(req)
	}

	t.mu.Lock()
	if t.authNext == nil {
		authNext, err := t.Provider(req.Context(), t.ClientParams)
		if err != nil {
			t.mu.Unlock()

			return nil, err
		}
		t.authNext = authNext
	}
	t.mu.Unlock()

	return t.authNext.RoundTrip(req)
}

// Initialize eagerly initializes the OAuth2 transport if it hasn't been already.
func (t *LazyAuthTransport) Initialize(ctx context.Context) error {
	if t.Provider == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.authNext == nil {
		authNext, err := t.Provider(ctx, t.ClientParams)
		if err != nil {
			return err
		}
		t.authNext = authNext
	}

	return nil
}
