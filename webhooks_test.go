package machinawallet

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func signWebhook(t *testing.T, secret string, payload []byte, ts int64) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", ts)))
	mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifySignatureValid(t *testing.T) {
	t.Parallel()

	now := time.Now()
	secret := "whsec_abc"
	payload := []byte(`{"hello":"world"}`)
	header := signWebhook(t, secret, payload, now.Unix())

	if err := verifySignatureAt(payload, header, secret, time.Minute, now); err != nil {
		t.Errorf("verify failed: %v", err)
	}
}

func TestVerifySignatureTampered(t *testing.T) {
	t.Parallel()

	now := time.Now()
	secret := "whsec_abc"
	payload := []byte(`{"hello":"world"}`)
	header := signWebhook(t, secret, payload, now.Unix())

	tampered := []byte(`{"hello":"evil"}`)
	if err := verifySignatureAt(tampered, header, secret, time.Minute, now); err == nil {
		t.Error("expected mismatch")
	}
}

func TestVerifySignatureExpired(t *testing.T) {
	t.Parallel()

	now := time.Now()
	signed := now.Add(-10 * time.Minute)
	secret := "whsec_abc"
	payload := []byte(`{}`)
	header := signWebhook(t, secret, payload, signed.Unix())

	err := verifySignatureAt(payload, header, secret, time.Minute, now)
	if err == nil {
		t.Error("expected expiry error")
	}
	if !strings.Contains(err.Error(), "tolerance") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifySignatureMalformed(t *testing.T) {
	t.Parallel()

	if err := VerifySignature([]byte(`{}`), "garbage", "secret", time.Minute); err == nil {
		t.Error("expected malformed error")
	}
	if err := VerifySignature([]byte(`{}`), "", "secret", time.Minute); err == nil {
		t.Error("expected empty header error")
	}
	if err := VerifySignature([]byte(`{}`), "t=1,v1=abc", "", time.Minute); err == nil {
		t.Error("expected empty secret error")
	}
}

func TestParseEvent(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"specversion": "1.0",
		"id": "evt_1",
		"source": "/machina/wallet",
		"type": "com.machina.wallet.transaction.confirmed",
		"data": {"tx_id": "tx_42"}
	}`)
	ev, err := ParseEvent(payload)
	if err != nil {
		t.Fatalf("ParseEvent: %v", err)
	}
	if ev.ID != "evt_1" {
		t.Errorf("ID = %q", ev.ID)
	}
	if ev.Type != "com.machina.wallet.transaction.confirmed" {
		t.Errorf("Type = %q", ev.Type)
	}
}

func TestParseEventMissingFields(t *testing.T) {
	t.Parallel()

	if _, err := ParseEvent([]byte(`{}`)); err == nil {
		t.Error("expected error for missing fields")
	}
	if _, err := ParseEvent([]byte(`not json`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
	if _, err := ParseEvent(nil); err == nil {
		t.Error("expected error for empty payload")
	}
}

func TestSentinelUnwrap(t *testing.T) {
	t.Parallel()

	// Sanity: verify the ErrAuth sentinel matches via errors.Is on
	// webhook auth errors emitted by verifySignatureAt.
	err := verifySignatureAt([]byte(`{}`), "t=1,v1=deadbeef", "secret", time.Minute, time.Unix(2, 0))
	if !errors.Is(err, ErrAuth) {
		t.Errorf("expected auth error sentinel, got %v", err)
	}
}
