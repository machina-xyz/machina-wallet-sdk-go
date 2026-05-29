package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStreamTransactions(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter is not a Flusher")
		}
		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, "data: {\"id\":\"tx_%d\",\"wallet_id\":\"wal_x\",\"status\":\"submitted\",\"intent_kind\":\"transfer\",\"amount_atoms\":\"100\",\"tenant_id\":\"tnt_1\",\"created_at\":\"2026-01-01T00:00:00Z\"}\n\n", i)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, AppID: "a", AppSecret: "s", Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, errs, err := c.StreamTransactions(ctx, "wal_x")
	if err != nil {
		t.Fatalf("StreamTransactions: %v", err)
	}

	got := 0
	for {
		select {
		case tx, ok := <-events:
			if !ok {
				if got != 3 {
					t.Errorf("received %d events, want 3", got)
				}
				return
			}
			if tx.WalletID != "wal_x" {
				t.Errorf("wallet id = %q", tx.WalletID)
			}
			got++
		case err := <-errs:
			if err != nil {
				t.Logf("stream err: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("timed out waiting for events")
		}
	}
}

func TestStreamTransactionsCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		for {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(20 * time.Millisecond):
				fmt.Fprint(w, "data: {}\n\n")
				if flusher != nil {
					flusher.Flush()
				}
			}
		}
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, AppID: "a", AppSecret: "s", Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	events, _, err := c.StreamTransactions(ctx, "wal_x")
	if err != nil {
		t.Fatalf("StreamTransactions: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Drain until the events channel closes
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-events:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("events channel did not close after cancellation")
		}
	}
}
