package oway

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"testing"
)

// createKeys runs CreateShipment against a server that fails the first
// failures attempts with 503, and returns the Idempotency-Key of each attempt.
func createKeys(t *testing.T, ctx context.Context, failures int) []string {
	t.Helper()
	return createKeysFailing(t, ctx, failures, http.StatusServiceUnavailable, `{"reason":"unavailable"}`)
}

// createKeysFailing is createKeys with the failing attempts' status and body.
func createKeysFailing(t *testing.T, ctx context.Context, failures, status int, body string) []string {
	t.Helper()
	var mu sync.Mutex
	var keys []string
	c, srv := newTestServer(t, okToken, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		attempt := len(keys)
		mu.Unlock()
		if attempt <= failures {
			writeJSON(w, status, body)
			return
		}
		writeJSON(w, http.StatusOK, `{"orderNumber":"ABC12"}`)
	})
	defer srv.Close()
	c.cfg.MaxRetries = failures

	if _, err := c.CreateShipment(ctx, &ShipmentRequest{Description: "Example freight"}); err != nil {
		t.Fatalf("CreateShipment: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), keys...)
}

func TestCreateShipment_RetriesSendOneKey(t *testing.T) {
	keys := createKeys(t, context.Background(), 1)
	if len(keys) != 2 {
		t.Fatalf("want 2 attempts, got %d", len(keys))
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(keys[0]) {
		t.Errorf("want a generated key, got %q", keys[0])
	}
	if keys[1] != keys[0] {
		t.Errorf("a retry must resend the same key: %q then %q", keys[0], keys[1])
	}
	if again := createKeys(t, context.Background(), 0); again[0] == keys[0] {
		t.Error("each call gets its own key")
	}
}

func TestCreateShipment_SendsTheCallersKey(t *testing.T) {
	keys := createKeys(t, WithIdempotencyKey(context.Background(), "order-4711"), 1)
	if len(keys) != 2 || keys[0] != "order-4711" || keys[1] != "order-4711" {
		t.Errorf("want the caller's key on every attempt, got %q", keys)
	}
}

// A create still in progress under the same key is answered 409 with a reason;
// the SDK decodes it, retries, and resends the same key.
func TestCreateShipment_RetriesACreateInProgress(t *testing.T) {
	keys := createKeysFailing(t, context.Background(), 1, http.StatusConflict,
		`{"status":409,"title":"Idempotency Key In Progress","reason":"idempotency_key_in_progress"}`)
	if len(keys) != 2 || keys[0] == "" || keys[1] != keys[0] {
		t.Errorf("want two attempts under one key, got %q", keys)
	}
}
