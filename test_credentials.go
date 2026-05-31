package machinawallet

import (
	"context"
	"net/http"
	"time"
)

// TestCredential is a short-lived credential issued by the development sandbox
// for automated tests. It is never accepted on production endpoints.
type TestCredential struct {
	AppID       string    `json:"app_id"`
	AppSecret   string    `json:"app_secret"`
	SigningMode string    `json:"signing_mode"` // "ed25519" | "hmac"
	BaseURL     string    `json:"base_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	IssuedAt    time.Time `json:"issued_at"`
}

// MintTestCredentialsRequest controls the credential mint shape.
type MintTestCredentialsRequest struct {
	// Mode is the signing mode the issued credential should use. When empty
	// the server picks a default suitable for the SDK.
	Mode string `json:"mode,omitempty"`
	// TTL is the desired credential lifetime. Server enforces an upper
	// bound; zero requests the default (typically 1 hour).
	TTL time.Duration `json:"-"`
	// Label is an optional human-readable tag attached to the credential
	// for diagnostics.
	Label string `json:"label,omitempty"`
}

// TestCredentialsService is the typed wrapper for /v1/test_credentials.
type TestCredentialsService struct{ client *Client }

// TestCredentials returns the typed TestCredentialsService for this client.
func (c *Client) TestCredentials() *TestCredentialsService {
	return &TestCredentialsService{client: c}
}

// mintWire is the JSON shape transmitted to the server. We translate the
// human-friendly time.Duration field into integer seconds because the API
// speaks seconds-since-epoch and second-valued durations.
type mintWire struct {
	Mode       string `json:"mode,omitempty"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
	Label      string `json:"label,omitempty"`
}

// Mint requests a fresh sandbox credential.
func (s *TestCredentialsService) Mint(ctx context.Context, req *MintTestCredentialsRequest) (*TestCredential, error) {
	wire := mintWire{}
	if req != nil {
		wire.Mode = req.Mode
		wire.Label = req.Label
		if req.TTL > 0 {
			wire.TTLSeconds = int(req.TTL / time.Second)
		}
	}
	var out TestCredential
	if err := s.client.transport.request(ctx, http.MethodPost, "/v1/test_credentials", nil, wire, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
