package oway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestServer(t *testing.T, tokenHandler, apiHandler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", tokenHandler)
	mux.HandleFunc("/", apiHandler)
	srv := httptest.NewServer(mux)

	c, err := New(Config{
		ClientID:     "client_test",
		ClientSecret: "secret_test",
		APIKey:       "oway_sk_test",
		BaseURL:      srv.URL,
		MaxRetries:   -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, srv
}

func okToken(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok", "expiresIn": 3600})
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func TestRefreshToken_DecodesAccessToken(t *testing.T) {
	var calls int32
	c, srv := newTestServer(t,
		func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&calls, 1)
			okToken(w, nil)
		},
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer tok" {
				t.Errorf("missing bearer token: %q", r.Header.Get("Authorization"))
			}
			writeJSON(w, 200, `{}`)
		},
	)
	defer srv.Close()

	if _, err := c.GetShipment(context.Background(), "ABC12"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 token call, got %d", got)
	}
}

func TestRefreshToken_CachesAcrossCalls(t *testing.T) {
	var tokenCalls int32
	c, srv := newTestServer(t,
		func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&tokenCalls, 1)
			okToken(w, nil)
		},
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, 200, `{}`)
		},
	)
	defer srv.Close()

	for i := 0; i < 5; i++ {
		if _, err := c.GetShipment(context.Background(), "ABC12"); err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&tokenCalls); got != 1 {
		t.Errorf("expected token cached, got %d calls", got)
	}
}

func TestRefreshToken_FailsOnMalformedBody(t *testing.T) {
	c, srv := newTestServer(t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{not json`))
		},
		func(http.ResponseWriter, *http.Request) {},
	)
	defer srv.Close()

	_, err := c.GetShipment(context.Background(), "ABC12")
	if err == nil {
		t.Fatal("expected error on malformed token body")
	}
	if !strings.Contains(err.Error(), "decode token response") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRefreshToken_FailsOnEmptyAccessToken(t *testing.T) {
	c, srv := newTestServer(t,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accessToken":"","expiresIn":0}`))
		},
		func(http.ResponseWriter, *http.Request) {},
	)
	defer srv.Close()

	_, err := c.GetShipment(context.Background(), "ABC12")
	if err == nil {
		t.Fatal("expected error on empty token")
	}
}

func TestParseHTTPError_PopulatesAllFields(t *testing.T) {
	body := []byte(`{
		"status": 422,
		"title": "Unprocessable Entity",
		"detail": "lane not covered",
		"reason": "no_coverage"
	}`)
	e := parseHTTPError(422, "req-123", body)
	if e.StatusCode != 422 || e.Code != "no_coverage" || e.Message != "lane not covered" || e.RequestID != "req-123" {
		t.Errorf("unexpected parsed error: %+v", e)
	}
}

func TestParseHTTPError_FallsBackToTitle(t *testing.T) {
	e := parseHTTPError(400, "", []byte(`{"title":"Bad Request"}`))
	if e.Message != "Bad Request" {
		t.Errorf("expected title fallback, got %q", e.Message)
	}
}

func TestParseHTTPError_ParsesViolations(t *testing.T) {
	e := parseHTTPError(400, "", []byte(`{
		"reason":"validation_failed",
		"violations":[
			{"field":"pickupAddress.zipCode","reason":"must be 5 digits"},
			{"field":"orderComponents[0].weight","reason":"required"}
		]
	}`))
	if len(e.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(e.Violations))
	}
	if e.Violations[0].Field != "pickupAddress.zipCode" || e.Violations[1].Reason != "required" {
		t.Errorf("violations not parsed: %+v", e.Violations)
	}
}

func TestParseHTTPError_TolerantOfMalformedBody(t *testing.T) {
	e := parseHTTPError(500, "req", []byte(`{not json`))
	if e.StatusCode != 500 || e.Message != "" || len(e.RawBody) == 0 {
		t.Errorf("expected status+raw retained, got %+v", e)
	}
}

func TestError_IsRetryable(t *testing.T) {
	cases := []struct {
		status int
		want   bool
	}{
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{408, true},
		{422, false},
		{429, true},
		{500, true},
		{501, false},
		{502, true},
		{503, true},
		{504, true},
	}
	for _, tc := range cases {
		got := (&Error{StatusCode: tc.status}).IsRetryable()
		if got != tc.want {
			t.Errorf("status %d: want %v got %v", tc.status, tc.want, got)
		}
	}
}

func TestRetry_OnTransientUntilSuccess(t *testing.T) {
	var attempts int32
	c, srv := newTestServer(t,
		okToken,
		func(w http.ResponseWriter, _ *http.Request) {
			if atomic.AddInt32(&attempts, 1) < 3 {
				writeJSON(w, 503, `{"reason":"unavailable"}`)
				return
			}
			writeJSON(w, 200, `{}`)
		},
	)
	defer srv.Close()
	c.cfg.MaxRetries = 3

	if _, err := c.GetShipment(context.Background(), "ABC12"); err != nil {
		t.Fatalf("expected eventual success, got %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetry_StopsOnNonRetryable(t *testing.T) {
	var attempts int32
	c, srv := newTestServer(t,
		okToken,
		func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&attempts, 1)
			writeJSON(w, 422, `{"reason":"no_coverage"}`)
		},
	)
	defer srv.Close()
	c.cfg.MaxRetries = 3

	_, err := c.GetShipment(context.Background(), "ABC12")
	oe, ok := AsError(err)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if oe.Code != "no_coverage" {
		t.Errorf("expected code no_coverage, got %q", oe.Code)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("expected 1 attempt on 422, got %d", got)
	}
}

func TestRetry_RespectsContextCancellation(t *testing.T) {
	var attempts int32
	c, srv := newTestServer(t,
		okToken,
		func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&attempts, 1)
			writeJSON(w, 503, `{}`)
		},
	)
	defer srv.Close()
	c.cfg.MaxRetries = 10

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := c.GetShipment(ctx, "ABC12")
	if err == nil {
		t.Fatal("expected context error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}

func TestWithCompanyAPIKey_OverridesDefault(t *testing.T) {
	var seenKey atomic.Value
	c, srv := newTestServer(t,
		okToken,
		func(w http.ResponseWriter, r *http.Request) {
			seenKey.Store(r.Header.Get("x-oway-api-key"))
			writeJSON(w, 200, `{}`)
		},
	)
	defer srv.Close()

	ctx := WithCompanyAPIKey(context.Background(), "oway_sk_tenant_xyz")
	if _, err := c.GetShipment(ctx, "ABC12"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, _ := seenKey.Load().(string); got != "oway_sk_tenant_xyz" {
		t.Errorf("expected per-request key, got %q", got)
	}
}

func TestRequestID_IsUniquePerRequest(t *testing.T) {
	seen := sync.Map{}
	var collisions int32
	c, srv := newTestServer(t,
		okToken,
		func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("x-request-id")
			if _, loaded := seen.LoadOrStore(id, true); loaded {
				atomic.AddInt32(&collisions, 1)
			}
			writeJSON(w, 200, `{}`)
		},
	)
	defer srv.Close()

	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.GetShipment(context.Background(), "ABC12"); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&collisions); got > 0 {
		t.Errorf("request-id collisions: %d", got)
	}
}

func TestNilRequestReturnsError(t *testing.T) {
	c, srv := newTestServer(t, okToken, func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when request is nil")
	})
	defer srv.Close()

	if _, err := c.RequestQuote(context.Background(), nil); err == nil {
		t.Error("expected error for nil QuoteRequest")
	}
	if _, err := c.CreateShipment(context.Background(), nil); err == nil {
		t.Error("expected error for nil ShipmentRequest")
	}
}

func TestAsError(t *testing.T) {
	oe := &Error{StatusCode: 422, Code: "no_coverage"}
	if got, ok := AsError(oe); !ok || got.Code != "no_coverage" {
		t.Errorf("AsError direct: %+v %v", got, ok)
	}
	if _, ok := AsError(nil); ok {
		t.Error("AsError(nil) should be false")
	}
}
