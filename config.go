package machinawallet

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Default values for client configuration.
const (
	DefaultTimeout    = 10 * time.Second
	DefaultMaxRetries = 3
	DefaultBaseURL    = "https://api.machina.money"
)

// Config configures a Client. AppSecret is server-side only and must never be
// exposed to browsers or mobile clients.
type Config struct {
	// BaseURL is the MACHINA API root, for example "https://api.machina.money".
	// Trailing slashes are stripped. Defaults to DefaultBaseURL when empty.
	BaseURL string

	// AppID is the public application identifier issued by the developer
	// portal. Sent as the X-Machina-Service header on every request.
	AppID string

	// AppSecret is either a 64-char hex Ed25519 seed (recommended) or any other
	// string interpreted as an HMAC-SHA256 secret. The signing mode is
	// auto-detected from the secret format.
	AppSecret string

	// Timeout is the per-request HTTP timeout. Defaults to DefaultTimeout.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts for retryable errors
	// (5xx and 429). Defaults to DefaultMaxRetries.
	MaxRetries int

	// Logger is used for SDK-internal diagnostic logging. When nil, a no-op
	// logger is used.
	Logger *slog.Logger

	// HTTPClient is the underlying http.Client. When nil, a new client with
	// Timeout is constructed. Provide your own to customize transport, cookies,
	// or proxies. The SDK will set Timeout on this client if it is zero.
	HTTPClient *http.Client

	// UserAgent overrides the default SDK User-Agent header.
	UserAgent string
}

// Validate returns an error if the config is incomplete or invalid.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.AppID) == "" {
		return errors.New("machinawallet: Config.AppID is required")
	}
	if strings.TrimSpace(c.AppSecret) == "" {
		return errors.New("machinawallet: Config.AppSecret is required")
	}
	return nil
}

// withDefaults returns a copy of the config with zero-valued fields replaced
// by sensible defaults.
func (c Config) withDefaults() Config {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = DefaultMaxRetries
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.NewTextHandler(discardWriter{}, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	if c.UserAgent == "" {
		c.UserAgent = "machina-wallet-sdk-go/0.1.0"
	}
	return c
}

// discardWriter implements io.Writer for the default no-op logger.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
