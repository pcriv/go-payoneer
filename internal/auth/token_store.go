package auth

import (
	"sync"

	"golang.org/x/oauth2"
)

// TokenStore is a thread-safe interface for storing and retrieving tokens.
type TokenStore interface {
	Get() *oauth2.Token
	Set(token *oauth2.Token)
}

// InMemoryStore is an in-memory implementation of TokenStore.
type InMemoryStore struct {
	mu    sync.RWMutex
	token *oauth2.Token
}

// NewInMemoryStore returns a new InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

// Get returns the current token.
func (s *InMemoryStore) Get() *oauth2.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

// Set stores the token.
func (s *InMemoryStore) Set(token *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = token
}
