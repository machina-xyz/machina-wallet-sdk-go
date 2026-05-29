package machinawallet

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Signing mode discriminators sent in the X-Machina-Mode header.
const (
	SigningModeEd25519 = "ed25519"
	SigningModeHMAC    = "hmac"
)

// HeaderService is the request header containing the app id.
const (
	HeaderService   = "X-Machina-Service"
	HeaderTimestamp = "X-Machina-Timestamp"
	HeaderSignature = "X-Machina-Signature"
	HeaderMode      = "X-Machina-Mode"
)

var hex64Re = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

// Signer produces signed-request headers for an HTTP request. Implementations
// must be safe for concurrent use.
type Signer interface {
	// Sign returns the signature string and the signing mode for a request.
	// The timestamp argument lets callers inject a deterministic time in
	// tests; production callers should pass time.Now().
	Sign(method, path string, body []byte, timestamp time.Time) (signature string, mode string, err error)
}

// Ed25519Signer signs the canonical string with an Ed25519 private key derived
// from a 32-byte seed.
type Ed25519Signer struct {
	priv ed25519.PrivateKey
}

// NewEd25519Signer returns a signer initialized from a 64-char hex seed.
func NewEd25519Signer(seedHex string) (*Ed25519Signer, error) {
	if !hex64Re.MatchString(seedHex) {
		return nil, errors.New("machinawallet: ed25519 seed must be 64 hex characters")
	}
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("machinawallet: decoding ed25519 seed: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("machinawallet: ed25519 seed must be %d bytes", ed25519.SeedSize)
	}
	return &Ed25519Signer{priv: ed25519.NewKeyFromSeed(seed)}, nil
}

// Sign implements the Signer interface for Ed25519.
func (s *Ed25519Signer) Sign(method, path string, body []byte, timestamp time.Time) (string, string, error) {
	canonical := canonicalString(method, path, body, formatTimestamp(timestamp))
	sig := ed25519.Sign(s.priv, []byte(canonical))
	return base64.StdEncoding.EncodeToString(sig), SigningModeEd25519, nil
}

// HMACSigner signs the canonical string with HMAC-SHA256.
type HMACSigner struct {
	secret []byte
}

// NewHMACSigner returns a signer initialized from the raw secret string.
func NewHMACSigner(secret string) (*HMACSigner, error) {
	if secret == "" {
		return nil, errors.New("machinawallet: HMAC secret must not be empty")
	}
	return &HMACSigner{secret: []byte(secret)}, nil
}

// Sign implements the Signer interface for HMAC-SHA256.
func (s *HMACSigner) Sign(method, path string, body []byte, timestamp time.Time) (string, string, error) {
	canonical := canonicalString(method, path, body, formatTimestamp(timestamp))
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(canonical))
	sum := mac.Sum(nil)
	return "sha256=" + hex.EncodeToString(sum), SigningModeHMAC, nil
}

// NewSigner selects an Ed25519 or HMAC signer based on the format of the app
// secret. A 64-character hex string is treated as an Ed25519 seed; any other
// non-empty string is treated as an HMAC-SHA256 secret.
func NewSigner(appSecret string) (Signer, error) {
	if hex64Re.MatchString(appSecret) {
		return NewEd25519Signer(appSecret)
	}
	return NewHMACSigner(appSecret)
}

// canonicalString builds the canonical string that is signed. The format is
// fixed and must match the server's expectation exactly.
func canonicalString(method, path string, body []byte, timestamp string) string {
	if body == nil {
		body = []byte{}
	}
	bodyHash := sha256.Sum256(body)
	return fmt.Sprintf("%s|%s|%s|%s", method, path, hex.EncodeToString(bodyHash[:]), timestamp)
}

// formatTimestamp renders a time.Time as the unix epoch seconds string used in
// canonical signing. This matches the wire format used by the Python and
// TypeScript SDKs.
func formatTimestamp(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
