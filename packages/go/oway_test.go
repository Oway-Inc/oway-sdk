package oway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestTokenManagement(t *testing.T) {
	tokenCallCount := 0
	var mu sync.Mutex

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		tokenCallCount++
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "test_token", "expires_in": 3600}`))
	}))
	defer tokenServer.Close()

	client, err := New(Config{
		ClientID:     "client_test",
		ClientSecret: "secret_test",
		APIKey:       "oway_sk_test_123",
		BaseURL:      "https://api.oway.io",
		TokenURL:     tokenServer.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should fetch token on first request", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.getAccessToken(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if tokenCallCount != 1 {
			t.Errorf("Expected 1 token call, got %d", tokenCallCount)
		}
	})

	t.Run("should reuse cached token", func(t *testing.T) {
		ctx := context.Background()
		initialCount := tokenCallCount

		for i := 0; i < 5; i++ {
			_, err := client.getAccessToken(ctx)
			if err != nil {
				t.Fatal(err)
			}
		}

		if tokenCallCount != initialCount {
			t.Errorf("Token should be cached, but was called %d more times", tokenCallCount-initialCount)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		retryable  bool
	}{
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"404 Not Found", 404, false},
		{"429 Rate Limit", 429, true},
		{"500 Server Error", 500, true},
		{"501 Not Implemented", 501, false},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{
				Message:    "Test error",
				Code:       "TEST_ERROR",
				StatusCode: tt.statusCode,
			}

			if err.IsRetryable() != tt.retryable {
				t.Errorf("Status %d: expected retryable=%v, got %v",
					tt.statusCode, tt.retryable, err.IsRetryable())
			}
		})
	}
}
