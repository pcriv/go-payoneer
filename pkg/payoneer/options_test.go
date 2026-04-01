package payoneer

import (
	"net/http"
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
		if req.AlreadyHaveAnAccount != "true" {
			t.Errorf("got AlreadyHaveAnAccount %s, want true", req.AlreadyHaveAnAccount)
		}
	})

	t.Run("WithPayeeContact", func(t *testing.T) {
		req := &RegistrationLinkRequest{}
		WithPayeeContact("John", "Doe", "john@example.com")(req)

		if req.Payee == nil || req.Payee.Contact == nil {
			t.Fatal("expected Payee.Contact to be set")
		}

		if req.Payee.Contact.FirstName != "John" {
			t.Errorf("got FirstName %s, want John", req.Payee.Contact.FirstName)
		}

		if req.Payee.Contact.LastName != "Doe" {
			t.Errorf("got LastName %s, want Doe", req.Payee.Contact.LastName)
		}

		if req.Payee.Contact.Email != "john@example.com" {
			t.Errorf("got Email %s, want john@example.com", req.Payee.Contact.Email)
		}
	})
}
