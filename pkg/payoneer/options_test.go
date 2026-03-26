package payoneer

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestWithSandbox(t *testing.T) {
	c := NewClient(WithSandbox())
	if c.BaseURL != SandboxBaseURL {
		t.Errorf("BaseURL: got %s, want %s", c.BaseURL, SandboxBaseURL)
	}
	if c.AuthBaseURL != SandboxAuthBaseURL {
		t.Errorf("AuthBaseURL: got %s, want %s", c.AuthBaseURL, SandboxAuthBaseURL)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient(WithHTTPClient(custom))
	if c.HTTPClient.Timeout != 5*time.Second {
		t.Errorf("got timeout %v, want 5s", c.HTTPClient.Timeout)
	}
}

func TestWithTokenStore(t *testing.T) {
	c := NewClient()
	if c.tokenStore == nil {
		t.Error("expected default token store, got nil")
	}
}

func TestTransactionListOptions(t *testing.T) {
	t.Run("WithPage", func(t *testing.T) {
		v := url.Values{}
		WithPage(3)(&v)
		if v.Get("page") != "3" {
			t.Errorf("got %s, want 3", v.Get("page"))
		}
	})

	t.Run("WithPageSize", func(t *testing.T) {
		v := url.Values{}
		WithPageSize(25)(&v)
		if v.Get("page_size") != "25" {
			t.Errorf("got %s, want 25", v.Get("page_size"))
		}
	})

	t.Run("WithFrom", func(t *testing.T) {
		v := url.Values{}
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		WithFrom(ts)(&v)
		if v.Get("from") != "2025-01-15T00:00:00Z" {
			t.Errorf("got %s, want 2025-01-15T00:00:00Z", v.Get("from"))
		}
	})

	t.Run("WithTo", func(t *testing.T) {
		v := url.Values{}
		ts := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)
		WithTo(ts)(&v)
		if v.Get("to") != "2025-06-30T23:59:59Z" {
			t.Errorf("got %s, want 2025-06-30T23:59:59Z", v.Get("to"))
		}
	})
}

func TestRegistrationOptions(t *testing.T) {
	t.Run("WithRedirectURL", func(t *testing.T) {
		req := &RegistrationLinkRequest{}
		WithRedirectURL("https://example.com/callback")(req)
		if req.RedirectURL != "https://example.com/callback" {
			t.Errorf("got %s, want https://example.com/callback", req.RedirectURL)
		}
	})

	t.Run("WithAlreadyHaveAnAccount", func(t *testing.T) {
		req := &RegistrationLinkRequest{}
		WithAlreadyHaveAnAccount(true)(req)
		if !req.AlreadyHaveAnAccount {
			t.Error("expected AlreadyHaveAnAccount to be true")
		}
	})

	t.Run("WithPayeeDetails", func(t *testing.T) {
		req := &RegistrationLinkRequest{}
		WithPayeeDetails("John", "Doe", "john@example.com")(req)

		fn, ok := req.FirstName.Get()
		if !ok || fn != "John" {
			t.Errorf("got FirstName %v, want John", fn)
		}

		ln, ok := req.LastName.Get()
		if !ok || ln != "Doe" {
			t.Errorf("got LastName %v, want Doe", ln)
		}

		em, ok := req.Email.Get()
		if !ok || em != "john@example.com" {
			t.Errorf("got Email %v, want john@example.com", em)
		}
	})
}
