package auth

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	token := &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(time.Hour),
	}

	// Test Set and Get
	store.Set(token)
	got := store.Get()
	if got.AccessToken != token.AccessToken {
		t.Errorf("expected access token %s, got %s", token.AccessToken, got.AccessToken)
	}

	// Test Concurrent Access
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.Set(&oauth2.Token{AccessToken: string(rune(i))})
			_ = store.Get()
		}(i)
	}
	wg.Wait()
}
