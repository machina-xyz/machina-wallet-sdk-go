package machinawallet

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestTestCredentialsMint(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test_credentials" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var wire mintWire
		_ = json.NewDecoder(r.Body).Decode(&wire)
		if wire.Mode != "ed25519" {
			t.Errorf("mode = %q", wire.Mode)
		}
		if wire.TTLSeconds != 60 {
			t.Errorf("ttl_seconds = %d", wire.TTLSeconds)
		}
		if wire.Label != "ci-run-42" {
			t.Errorf("label = %q", wire.Label)
		}
		_ = json.NewEncoder(w).Encode(TestCredential{
			AppID:       "app_x",
			AppSecret:   "secret",
			SigningMode: "ed25519",
			BaseURL:     "https://sandbox.machina.money",
			ExpiresAt:   time.Now().Add(time.Hour).UTC(),
			IssuedAt:    time.Now().UTC(),
		})
	})

	got, err := c.TestCredentials().Mint(context.Background(), &MintTestCredentialsRequest{
		Mode:  "ed25519",
		TTL:   60 * time.Second,
		Label: "ci-run-42",
	})
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if got.AppID != "app_x" {
		t.Errorf("AppID = %q", got.AppID)
	}
}

func TestTestCredentialsMintNilRequest(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var wire mintWire
		_ = json.NewDecoder(r.Body).Decode(&wire)
		if wire.Mode != "" || wire.TTLSeconds != 0 || wire.Label != "" {
			t.Errorf("expected zero wire, got %+v", wire)
		}
		_ = json.NewEncoder(w).Encode(TestCredential{AppID: "app_default"})
	})

	got, err := c.TestCredentials().Mint(context.Background(), nil)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if got.AppID != "app_default" {
		t.Errorf("AppID = %q", got.AppID)
	}
}
