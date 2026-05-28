package middleware

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitStoreCleanupRemovesStaleEntries(t *testing.T) {
	store := newRateLimitStore(10*time.Millisecond, time.Nanosecond)
	base := time.Now()

	store.limiterFor("auth_login|ip:203.0.113.1", rate.Every(time.Second), 1, base)
	store.limiterFor("auth_login|ip:203.0.113.2", rate.Every(time.Second), 1, base.Add(20*time.Millisecond))

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.entries["auth_login|ip:203.0.113.1"]; ok {
		t.Fatalf("expected stale entry to be removed from the store")
	}

	if _, ok := store.entries["auth_login|ip:203.0.113.2"]; !ok {
		t.Fatalf("expected recent entry to be kept in the store")
	}
}
