package machinawallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// staticP256Key generates a fresh P-256 EC key and returns the SEC1 PEM.
func staticP256Key(t *testing.T) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa generate: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
}

// pkcs8Ed25519Key returns a PKCS8-wrapped Ed25519 private key as PEM.
func pkcs8Ed25519Key(t *testing.T) string {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519 generate: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// pkcs8P256Key returns a PKCS8-wrapped P-256 private key as PEM.
func pkcs8P256Key(t *testing.T) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestP256PrivateKeyAuthSECPem(t *testing.T) {
	t.Parallel()

	pem := staticP256Key(t)

	var sawAuth, sawMode, sawKID string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get(HeaderAuthorization)
		sawMode = r.Header.Get(HeaderAuthorizationMode)
		sawKID = r.Header.Get(HeaderAuthorizationKeyID)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	withAuth := c.WithAuthorization(P256PrivateKeyAuth{PEM: pem, KeyID: "kid-1"})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if sawAuth == "" {
		t.Error("missing authorization header")
	}
	if sawMode != AuthorizationModeP256 {
		t.Errorf("mode = %q", sawMode)
	}
	if sawKID != "kid-1" {
		t.Errorf("kid = %q", sawKID)
	}

	// Original client should be unaffected.
	c2, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeaderAuthorization) != "" {
			t.Error("original client should not carry authorization header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	if _, err := c2.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
}

func TestP256PrivateKeyAuthPKCS8Ed25519(t *testing.T) {
	t.Parallel()

	pem := pkcs8Ed25519Key(t)

	var sawAuth string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get(HeaderAuthorization)
		_, _ = w.Write([]byte(`{}`))
	})

	withAuth := c.WithAuthorization(P256PrivateKeyAuth{PEM: pem})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if sawAuth == "" {
		t.Error("missing authorization header")
	}
	if _, err := base64.StdEncoding.DecodeString(sawAuth); err != nil {
		t.Errorf("auth header should be base64: %v", err)
	}
}

func TestP256PrivateKeyAuthPKCS8P256(t *testing.T) {
	t.Parallel()

	pem := pkcs8P256Key(t)

	var sawMode string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		sawMode = r.Header.Get(HeaderAuthorizationMode)
		_, _ = w.Write([]byte(`{}`))
	})

	withAuth := c.WithAuthorization(P256PrivateKeyAuth{PEM: pem})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if sawMode != AuthorizationModeP256 {
		t.Errorf("mode = %q", sawMode)
	}
}

func TestP256PrivateKeyAuthBadPEM(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})

	withAuth := c.WithAuthorization(P256PrivateKeyAuth{PEM: "garbage"})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Fatal("expected PEM decode error")
	}

	withAuth2 := c.WithAuthorization(P256PrivateKeyAuth{PEM: ""})
	if _, err := withAuth2.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Fatal("expected empty PEM error")
	}

	wrong := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte{0x01, 0x02}}))
	withAuth3 := c.WithAuthorization(P256PrivateKeyAuth{PEM: wrong})
	if _, err := withAuth3.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Fatal("expected unsupported PEM error")
	}
}

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	var sawAuth, sawMode string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get(HeaderAuthorization)
		sawMode = r.Header.Get(HeaderAuthorizationMode)
		_, _ = w.Write([]byte(`{}`))
	})

	withAuth := c.WithAuthorization(JWTAuth{Token: "eyJraWQi.tok.sig"})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if !strings.HasPrefix(sawAuth, "Bearer ") {
		t.Errorf("auth = %q", sawAuth)
	}
	if sawMode != AuthorizationModeJWT {
		t.Errorf("mode = %q", sawMode)
	}

	withAuth = c.WithAuthorization(JWTAuth{Token: ""})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Error("expected error for empty token")
	}
}

// fakeSigner implements ExternalSigner with HMAC.
type fakeSigner struct {
	key  []byte
	mode string
}

func (f *fakeSigner) Sign(canonical []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, f.key)
	mac.Write(canonical)
	return mac.Sum(nil), nil
}

func (f *fakeSigner) Mode() string { return f.mode }

type erroringSigner struct{}

func (erroringSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("hsm down") }
func (erroringSigner) Mode() string                { return AuthorizationModeExtHMAC }

func TestExternalSignerAuth(t *testing.T) {
	t.Parallel()

	var sawAuth, sawMode, sawKID string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get(HeaderAuthorization)
		sawMode = r.Header.Get(HeaderAuthorizationMode)
		sawKID = r.Header.Get(HeaderAuthorizationKeyID)
		_, _ = w.Write([]byte(`{}`))
	})

	signer := &fakeSigner{key: []byte("ext-secret"), mode: AuthorizationModeExtHMAC}
	withAuth := c.WithAuthorization(ExternalSignerAuth{Signer: signer, KeyID: "hsm-1"})
	if _, err := withAuth.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if sawAuth == "" {
		t.Error("missing auth header")
	}
	if _, err := base64.StdEncoding.DecodeString(sawAuth); err != nil {
		t.Errorf("auth should be base64: %v", err)
	}
	if sawMode != AuthorizationModeExtHMAC {
		t.Errorf("mode = %q", sawMode)
	}
	if sawKID != "hsm-1" {
		t.Errorf("kid = %q", sawKID)
	}

	// Empty mode falls back to ed25519 default.
	signer2 := &fakeSigner{key: []byte("x"), mode: ""}
	var seenMode string
	c2, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenMode = r.Header.Get(HeaderAuthorizationMode)
		_, _ = w.Write([]byte(`{}`))
	})
	withAuth2 := c2.WithAuthorization(ExternalSignerAuth{Signer: signer2})
	if _, err := withAuth2.GetWallet(context.Background(), "wal_x"); err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if seenMode != AuthorizationModeExtEd25519 {
		t.Errorf("expected default mode, got %q", seenMode)
	}

	// Nil signer is rejected.
	c3, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) })
	withAuth3 := c3.WithAuthorization(ExternalSignerAuth{})
	if _, err := withAuth3.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Error("expected nil-signer error")
	}

	// Errors from signer are propagated as auth errors.
	c4, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) })
	withAuth4 := c4.WithAuthorization(ExternalSignerAuth{Signer: erroringSigner{}})
	if _, err := withAuth4.GetWallet(context.Background(), "wal_x"); err == nil {
		t.Error("expected signer error")
	}
}
