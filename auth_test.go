package machinawallet

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestCanonicalString(t *testing.T) {
	t.Parallel()

	got := canonicalString("POST", "/v1/wallets", []byte(`{"chain":"ethereum"}`), "1700000000")
	parts := strings.Split(got, "|")
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts, got %d: %q", len(parts), got)
	}
	if parts[0] != "POST" {
		t.Errorf("method = %q, want POST", parts[0])
	}
	if parts[1] != "/v1/wallets" {
		t.Errorf("path = %q", parts[1])
	}
	bodyHash := sha256.Sum256([]byte(`{"chain":"ethereum"}`))
	if parts[2] != hex.EncodeToString(bodyHash[:]) {
		t.Errorf("body hash mismatch")
	}
	if parts[3] != "1700000000" {
		t.Errorf("timestamp = %q", parts[3])
	}
}

func TestCanonicalStringEmptyBody(t *testing.T) {
	t.Parallel()

	got := canonicalString("GET", "/v1/wallets", nil, "1700000000")
	emptyHash := sha256.Sum256([]byte{})
	if !strings.Contains(got, hex.EncodeToString(emptyHash[:])) {
		t.Errorf("empty body hash missing from %q", got)
	}
}

func TestNewSignerEd25519(t *testing.T) {
	t.Parallel()

	seed := strings.Repeat("ab", 32)
	signer, err := NewSigner(seed)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	if _, ok := signer.(*Ed25519Signer); !ok {
		t.Fatalf("expected Ed25519Signer, got %T", signer)
	}

	sig, mode, err := signer.Sign("GET", "/v1/wallets", nil, time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if mode != SigningModeEd25519 {
		t.Errorf("mode = %q, want %q", mode, SigningModeEd25519)
	}

	rawSig, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if len(rawSig) != ed25519.SignatureSize {
		t.Errorf("sig len = %d, want %d", len(rawSig), ed25519.SignatureSize)
	}

	seedBytes, _ := hex.DecodeString(seed)
	priv := ed25519.NewKeyFromSeed(seedBytes)
	pub := priv.Public().(ed25519.PublicKey)
	canonical := canonicalString("GET", "/v1/wallets", nil, "1700000000")
	if !ed25519.Verify(pub, []byte(canonical), rawSig) {
		t.Errorf("ed25519 signature did not verify")
	}
}

func TestNewSignerHMAC(t *testing.T) {
	t.Parallel()

	signer, err := NewSigner("not-a-hex-seed-but-an-hmac-secret")
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	if _, ok := signer.(*HMACSigner); !ok {
		t.Fatalf("expected HMACSigner, got %T", signer)
	}

	sig, mode, err := signer.Sign("POST", "/v1/wallets", []byte(`{"x":1}`), time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if mode != SigningModeHMAC {
		t.Errorf("mode = %q, want %q", mode, SigningModeHMAC)
	}
	if !strings.HasPrefix(sig, "sha256=") {
		t.Errorf("hmac signature should start with sha256=, got %q", sig)
	}

	mac := hmac.New(sha256.New, []byte("not-a-hex-seed-but-an-hmac-secret"))
	canonical := canonicalString("POST", "/v1/wallets", []byte(`{"x":1}`), "1700000000")
	mac.Write([]byte(canonical))
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if sig != expected {
		t.Errorf("sig = %q, want %q", sig, expected)
	}
}

func TestNewEd25519SignerInvalidSeed(t *testing.T) {
	t.Parallel()

	if _, err := NewEd25519Signer("not-hex"); err == nil {
		t.Errorf("expected error for non-hex seed")
	}
	if _, err := NewEd25519Signer(strings.Repeat("z", 64)); err == nil {
		t.Errorf("expected error for non-hex chars")
	}
}

func TestNewHMACSignerEmpty(t *testing.T) {
	t.Parallel()

	if _, err := NewHMACSigner(""); err == nil {
		t.Errorf("expected error for empty HMAC secret")
	}
}
