package machinawallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
)

// ecdsaSig is the ASN.1 DER shape for an ECDSA signature.
type ecdsaSig struct{ R, S *big.Int }

// signECDSA signs canonical with key and returns a base64-encoded DER
// signature. We always hash with SHA-256 to match the P-256 verifier.
func signECDSA(key *ecdsa.PrivateKey, canonical string) (string, error) {
	hash := sha256.Sum256([]byte(canonical))
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return "", fmt.Errorf("machinawallet: ecdsa sign: %w", err)
	}
	der, err := asn1.Marshal(ecdsaSig{R: r, S: s})
	if err != nil {
		return "", fmt.Errorf("machinawallet: ecdsa der-encode: %w", err)
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

// signECDSAAny narrows an interface{} to *ecdsa.PrivateKey.
func signECDSAAny(raw any, canonical string) (string, error) {
	key, ok := raw.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("machinawallet: PKCS8 key is not an ECDSA private key")
	}
	return signECDSA(key, canonical)
}
