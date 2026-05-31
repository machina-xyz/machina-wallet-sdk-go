package machinawallet

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AuthorizationContext layers a second authentication artifact on top of the
// app-level signing key. This is the pattern used for end-user-scoped calls
// where the SDK signs on behalf of a specific user session.
//
// The base app credentials still authenticate the request; the
// AuthorizationContext adds either a P-256 user signature, a bearer JWT, or
// delegates signing to an external signer.
type AuthorizationContext interface {
	// apply mutates req to attach the authorization header(s). The body is
	// passed pre-marshaled to allow the implementation to compute a digest
	// over the canonical bytes.
	apply(req *http.Request, body []byte) error
}

// HeaderAuthorization is the request header for the authorization artifact.
const (
	HeaderAuthorization      = "X-Machina-Authorization"
	HeaderAuthorizationMode  = "X-Machina-Authorization-Mode"
	HeaderAuthorizationKeyID = "X-Machina-Authorization-Key-Id"
)

// Authorization mode discriminators.
const (
	AuthorizationModeP256       = "p256"
	AuthorizationModeJWT        = "jwt"
	AuthorizationModeExtEd25519 = "ext_ed25519"
	AuthorizationModeExtHMAC    = "ext_hmac_sha256"
)

// P256PrivateKeyAuth signs the canonical request bytes with a PEM-encoded
// P-256 private key. The signature is attached as a base64 string.
type P256PrivateKeyAuth struct {
	// PEM is the PEM-encoded EC PRIVATE KEY or PKCS8 PRIVATE KEY for a
	// secp256r1 (P-256) curve.
	PEM string
	// KeyID is the optional kid header value matching a registered user key.
	KeyID string
}

// apply implements AuthorizationContext for P256PrivateKeyAuth.
func (a P256PrivateKeyAuth) apply(req *http.Request, body []byte) error {
	if strings.TrimSpace(a.PEM) == "" {
		return errors.New("machinawallet: P256PrivateKeyAuth.PEM is required")
	}
	sig, err := signP256(a.PEM, canonicalAuthString(req, body))
	if err != nil {
		return err
	}
	req.Header.Set(HeaderAuthorization, sig)
	req.Header.Set(HeaderAuthorizationMode, AuthorizationModeP256)
	if a.KeyID != "" {
		req.Header.Set(HeaderAuthorizationKeyID, a.KeyID)
	}
	return nil
}

// JWTAuth attaches a pre-issued JWT as the authorization artifact. The token
// is sent verbatim and the server verifies it against its trusted issuers.
type JWTAuth struct {
	Token string
}

// apply implements AuthorizationContext for JWTAuth.
func (a JWTAuth) apply(req *http.Request, _ []byte) error {
	if strings.TrimSpace(a.Token) == "" {
		return errors.New("machinawallet: JWTAuth.Token is required")
	}
	req.Header.Set(HeaderAuthorization, "Bearer "+a.Token)
	req.Header.Set(HeaderAuthorizationMode, AuthorizationModeJWT)
	return nil
}

// ExternalSigner is the contract for offloading authorization signing to an
// HSM, KMS, hardware wallet, or remote signer service. Implementations must
// be safe for concurrent use.
type ExternalSigner interface {
	// Sign returns a signature over canonical. Implementations should sign
	// the bytes as-is; the SDK does not perform additional hashing.
	Sign(canonical []byte) ([]byte, error)
	// Mode is the wire-level mode discriminator. Common values are
	// AuthorizationModeExtEd25519 and AuthorizationModeExtHMAC.
	Mode() string
}

// ExternalSignerAuth delegates authorization signing to an ExternalSigner.
type ExternalSignerAuth struct {
	Signer ExternalSigner
	KeyID  string
}

// apply implements AuthorizationContext for ExternalSignerAuth.
func (a ExternalSignerAuth) apply(req *http.Request, body []byte) error {
	if a.Signer == nil {
		return errors.New("machinawallet: ExternalSignerAuth.Signer is required")
	}
	sig, err := a.Signer.Sign([]byte(canonicalAuthString(req, body)))
	if err != nil {
		return fmt.Errorf("machinawallet: external signer: %w", err)
	}
	req.Header.Set(HeaderAuthorization, base64.StdEncoding.EncodeToString(sig))
	mode := a.Signer.Mode()
	if mode == "" {
		mode = AuthorizationModeExtEd25519
	}
	req.Header.Set(HeaderAuthorizationMode, mode)
	if a.KeyID != "" {
		req.Header.Set(HeaderAuthorizationKeyID, a.KeyID)
	}
	return nil
}

// canonicalAuthString builds the canonical string the user-key signs.
// The format mirrors the app-level signer but keys on the Authorization
// header semantics rather than the X-Machina-Service header.
func canonicalAuthString(req *http.Request, body []byte) string {
	ts := req.Header.Get(HeaderTimestamp)
	if ts == "" {
		ts = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return canonicalString(req.Method, req.URL.Path, body, ts)
}

// WithAuthorization returns a shallow copy of the client that attaches the
// given authorization context to every outgoing request.
//
// The returned client shares the same HTTP transport and signer as the
// original; only the auth-context layer is added.
func (c *Client) WithAuthorization(ctx AuthorizationContext) *Client {
	clone := *c
	cloneT := *c.transport
	cloneT.auth = ctx
	clone.transport = &cloneT
	return &clone
}

// signP256 PEM-decodes the supplied private key, signs the canonical bytes,
// and returns a base64-encoded signature. We accept either an SEC1 EC
// PRIVATE KEY block or a PKCS8 PRIVATE KEY block.
func signP256(pemStr, canonical string) (string, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return "", errors.New("machinawallet: P256 PEM block missing")
	}

	// We sign with Ed25519 fall-through when the key is actually an
	// Ed25519 private key — supporting the most common HSM exports.
	// Otherwise we attempt ECDSA over P-256.
	switch block.Type {
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("machinawallet: parsing EC private key: %w", err)
		}
		// The standard ecdsa package returns r,s. We use a deterministic
		// format: hash first, then sign. To stay stdlib-only we delegate
		// to crypto/ecdsa via the higher level Sign helper.
		return signECDSA(key, canonical)
	case "PRIVATE KEY":
		raw, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("machinawallet: parsing PKCS8 private key: %w", err)
		}
		switch k := raw.(type) {
		case ed25519.PrivateKey:
			sig := ed25519.Sign(k, []byte(canonical))
			return base64.StdEncoding.EncodeToString(sig), nil
		default:
			// ECDSA path: cast via the local helper.
			return signECDSAAny(k, canonical)
		}
	default:
		return "", fmt.Errorf("machinawallet: unsupported PEM block type %q", block.Type)
	}
}
