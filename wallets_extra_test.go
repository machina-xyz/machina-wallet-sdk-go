package machinawallet

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRawSign(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/raw_sign") {
			t.Errorf("path = %q", r.URL.Path)
		}
		var opts RawSignOptions
		_ = json.NewDecoder(r.Body).Decode(&opts)
		if opts.HashHex != "deadbeef" {
			t.Errorf("hash = %q", opts.HashHex)
		}
		_ = json.NewEncoder(w).Encode(RawSignature{SignatureHex: "abc", Curve: "secp256k1", SignedAt: time.Now().UTC()})
	})

	got, err := c.RawSign(context.Background(), "wal_x", RawSignOptions{HashHex: "deadbeef", Curve: "secp256k1"})
	if err != nil {
		t.Fatalf("RawSign: %v", err)
	}
	if got.SignatureHex != "abc" {
		t.Errorf("sig = %q", got.SignatureHex)
	}

	if _, err := c.RawSign(context.Background(), "", RawSignOptions{HashHex: "x"}); err == nil {
		t.Error("expected validation for empty walletID")
	}
	if _, err := c.RawSign(context.Background(), "w", RawSignOptions{}); err == nil {
		t.Error("expected validation for empty payload")
	}
	if _, err := c.RawSign(context.Background(), "w", RawSignOptions{HashHex: "x", MessageBytes: []byte{1}}); err == nil {
		t.Error("expected validation for both fields set")
	}
}

func TestExportWallet(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/export") {
			t.Errorf("path = %q", r.URL.Path)
		}
		out := ExportedWallet{
			WalletID:          "wal_x",
			EncryptedKeyB64:   "base64key",
			WrappingAlgorithm: "ECDH-ES+A256KW",
			ExportedAt:        time.Now().UTC(),
		}
		_ = json.NewEncoder(w).Encode(out)
	})

	jwk := JsonWebKey{Kty: "EC", Crv: "P-256", X: "x", Y: "y"}
	got, err := c.ExportWallet(context.Background(), "wal_x", jwk)
	if err != nil {
		t.Fatalf("ExportWallet: %v", err)
	}
	if got.WalletID != "wal_x" {
		t.Errorf("wallet_id = %q", got.WalletID)
	}

	if _, err := c.ExportWallet(context.Background(), "", jwk); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.ExportWallet(context.Background(), "w", JsonWebKey{}); err == nil {
		t.Error("expected validation for empty kty")
	}
}

func TestImportWallet(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/wallets/import" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Wallet{ID: "wal_new", Chain: ChainEthereum})
	})

	req := ImportWalletRequest{
		OwnerUserID:       "usr_1",
		Chain:             ChainEthereum,
		CustodyKind:       CustodySelfCustody,
		EncryptedKeyB64:   "data",
		WrappingAlgorithm: "ECDH-ES",
	}
	got, err := c.ImportWallet(context.Background(), req)
	if err != nil {
		t.Fatalf("ImportWallet: %v", err)
	}
	if got.ID != "wal_new" {
		t.Errorf("id = %q", got.ID)
	}

	if _, err := c.ImportWallet(context.Background(), ImportWalletRequest{}); err == nil {
		t.Error("expected validation for empty owner")
	}
	if _, err := c.ImportWallet(context.Background(), ImportWalletRequest{OwnerUserID: "u"}); err == nil {
		t.Error("expected validation for empty chain")
	}
	if _, err := c.ImportWallet(context.Background(), ImportWalletRequest{OwnerUserID: "u", Chain: ChainEthereum}); err == nil {
		t.Error("expected validation for empty key")
	}
	if _, err := c.ImportWallet(context.Background(), ImportWalletRequest{OwnerUserID: "u", Chain: ChainEthereum, EncryptedKeyB64: "k"}); err == nil {
		t.Error("expected validation for empty algorithm")
	}
}

func TestUpdateWallet(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		var patch UpdateWalletRequest
		_ = json.NewDecoder(r.Body).Decode(&patch)
		_ = json.NewEncoder(w).Encode(Wallet{ID: "wal_x", Status: *patch.Status})
	})

	frozen := WalletStatusFrozen
	got, err := c.UpdateWallet(context.Background(), "wal_x", UpdateWalletRequest{Status: &frozen})
	if err != nil {
		t.Fatalf("UpdateWallet: %v", err)
	}
	if got.Status != WalletStatusFrozen {
		t.Errorf("status = %v", got.Status)
	}

	if _, err := c.UpdateWallet(context.Background(), "", UpdateWalletRequest{}); err == nil {
		t.Error("expected validation")
	}
}
