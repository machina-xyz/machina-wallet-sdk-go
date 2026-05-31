package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// RawSignOptions controls a raw-signing request against a custodial key.
// Exactly one of HashHex or MessageBytes should be set; the SDK does not
// hash on behalf of the caller for raw signing.
type RawSignOptions struct {
	// HashHex is the hex-encoded payload to sign as-is. Use this for chains
	// where the SDK should sign a pre-computed digest (Secp256k1, etc).
	HashHex string `json:"hash_hex,omitempty"`
	// MessageBytes is the raw byte payload, base64-encoded over the wire,
	// for chains where the server selects the digest (Ed25519, etc).
	MessageBytes []byte `json:"message,omitempty"`
	// Curve is the elliptic curve identifier expected by the server, e.g.
	// "secp256k1", "ed25519", "p256". Optional — defaults are chosen per
	// wallet curve at the server.
	Curve string `json:"curve,omitempty"`
	// SubKeyPath is the optional BIP-32 derivation path appended to the
	// wallet master, when the wallet supports derived keys.
	SubKeyPath string `json:"sub_key_path,omitempty"`
}

// RawSignature is the response from RawSign.
type RawSignature struct {
	SignatureHex string    `json:"signature_hex"`
	Curve        string    `json:"curve"`
	SignedAt     time.Time `json:"signed_at"`
}

// JsonWebKey is a minimal RFC 7517 JWK representation used to wrap export
// recipients. Fields are passed through to the server verbatim.
type JsonWebKey struct {
	Kty    string `json:"kty"`
	Use    string `json:"use,omitempty"`
	Alg    string `json:"alg,omitempty"`
	Kid    string `json:"kid,omitempty"`
	Crv    string `json:"crv,omitempty"`
	X      string `json:"x,omitempty"`
	Y      string `json:"y,omitempty"`
	N      string `json:"n,omitempty"`
	E      string `json:"e,omitempty"`
	KeyOps []string `json:"key_ops,omitempty"`
}

// ExportedWallet is the wrapped key material returned by ExportWallet.
type ExportedWallet struct {
	WalletID            string     `json:"wallet_id"`
	EncryptedKeyB64     string     `json:"encrypted_key_b64"`
	WrappingAlgorithm   string     `json:"wrapping_algorithm"`
	RecipientKid        string     `json:"recipient_kid,omitempty"`
	ServerEphemeralJWK  JsonWebKey `json:"server_ephemeral_jwk"`
	ExportedAt          time.Time  `json:"exported_at"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
}

// ImportWalletRequest is the input shape for ImportWallet.
type ImportWalletRequest struct {
	OwnerUserID       string     `json:"owner_user_id"`
	Chain             Chain      `json:"chain"`
	CustodyKind       CustodyKind `json:"custody_kind"`
	Label             string     `json:"label,omitempty"`
	EncryptedKeyB64   string     `json:"encrypted_key_b64"`
	WrappingAlgorithm string     `json:"wrapping_algorithm"`
	WrappingJWK       JsonWebKey `json:"wrapping_jwk"`
}

// UpdateWalletRequest is a partial update to a wallet record. Only non-nil
// fields are sent over the wire.
type UpdateWalletRequest struct {
	Label        *string       `json:"label,omitempty"`
	Status       *WalletStatus `json:"status,omitempty"`
	FrozenReason *string       `json:"frozen_reason,omitempty"`
}

// RawSign requests a raw signature from a custodial wallet without
// constructing or broadcasting a transaction. It returns a hex-encoded
// signature on the requested curve.
func (c *Client) RawSign(ctx context.Context, walletID string, opts RawSignOptions) (*RawSignature, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	if opts.HashHex == "" && len(opts.MessageBytes) == 0 {
		return nil, &Error{Code: ErrCodeValidation, Message: "RawSignOptions: one of HashHex or MessageBytes is required"}
	}
	if opts.HashHex != "" && len(opts.MessageBytes) > 0 {
		return nil, &Error{Code: ErrCodeValidation, Message: "RawSignOptions: HashHex and MessageBytes are mutually exclusive"}
	}
	var out RawSignature
	path := fmt.Sprintf("/v1/wallets/%s/raw_sign", walletID)
	if err := c.transport.request(ctx, http.MethodPost, path, nil, opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportWallet returns the wallet's key material encrypted to the supplied
// recipient public key. The server selects the wrapping algorithm based on
// the recipient JWK; the SDK does not unwrap the result.
func (c *Client) ExportWallet(ctx context.Context, walletID string, recipientPublicKey JsonWebKey) (*ExportedWallet, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	if recipientPublicKey.Kty == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "recipientPublicKey.Kty is required"}
	}
	body := map[string]any{"recipient_jwk": recipientPublicKey}
	var out ExportedWallet
	path := fmt.Sprintf("/v1/wallets/%s/export", walletID)
	if err := c.transport.request(ctx, http.MethodPost, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ImportWallet provisions a new wallet from caller-supplied, wrapped key
// material. The server unwraps the key using its private half of the
// previously-published wrapping JWK.
func (c *Client) ImportWallet(ctx context.Context, req ImportWalletRequest) (*Wallet, error) {
	if req.OwnerUserID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "ImportWalletRequest.OwnerUserID is required"}
	}
	if req.Chain == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "ImportWalletRequest.Chain is required"}
	}
	if req.EncryptedKeyB64 == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "ImportWalletRequest.EncryptedKeyB64 is required"}
	}
	if req.WrappingAlgorithm == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "ImportWalletRequest.WrappingAlgorithm is required"}
	}
	var out Wallet
	if err := c.transport.request(ctx, http.MethodPost, "/v1/wallets/import", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateWallet applies a partial update to a wallet.
func (c *Client) UpdateWallet(ctx context.Context, walletID string, patch UpdateWalletRequest) (*Wallet, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var out Wallet
	path := fmt.Sprintf("/v1/wallets/%s", walletID)
	if err := c.transport.request(ctx, http.MethodPatch, path, nil, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
