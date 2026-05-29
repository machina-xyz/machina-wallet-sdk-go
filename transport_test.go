package machinawallet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryAfterHonored(t *testing.T) {
	t.Parallel()

	var attempts int32
	var firstHit time.Time
	var secondHit time.Time

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			firstHit = time.Now()
			w.Header().Set("Retry-After", "1")
			http.Error(w, `{"error":{"code":"rate_limit_error","message":"slow down"}}`, http.StatusTooManyRequests)
			return
		}
		secondHit = time.Now()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wal_x"}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		BaseURL:    srv.URL,
		AppID:      "app",
		AppSecret:  "secret",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}

	gap := secondHit.Sub(firstHit)
	if gap < 900*time.Millisecond {
		t.Errorf("retry gap %v was less than Retry-After hint", gap)
	}
}

func TestShouldRetryClassification(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status int
		want   bool
	}{
		{500, true},
		{502, true},
		{429, true},
		{400, false},
		{404, false},
		{401, false},
		{200, false},
	}
	for _, tc := range cases {
		if got := shouldRetry(tc.status, nil); got != tc.want {
			t.Errorf("shouldRetry(%d) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestRetryAfterParsing(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", "3")
	if got := retryAfter(resp); got != 3*time.Second {
		t.Errorf("retryAfter seconds = %v", got)
	}

	resp = &http.Response{Header: http.Header{}}
	if got := retryAfter(resp); got != 0 {
		t.Errorf("retryAfter empty = %v", got)
	}
}
