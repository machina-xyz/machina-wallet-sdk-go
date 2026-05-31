package machinawallet

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/machina-xyz/machina-wallet-sdk-go/webhooks"
)

func TestWebhookHelperVerifyAndParse(t *testing.T) {
	t.Parallel()

	walletID := uuid.New()
	ownerID := uuid.New()
	tenantID := uuid.New()
	payload, err := json.Marshal(webhooks.WalletCreatedEvent{
		Envelope: webhooks.Envelope{
			Type:     webhooks.TypeWalletCreated,
			ID:       "evt_1",
			Time:     time.Now().UTC(),
			TenantID: &tenantID,
		},
		Data: webhooks.WalletCreatedData{
			WalletID:    walletID,
			OwnerUserID: ownerID,
			Chain:       "ethereum",
			Address:     "0xabc",
			CustodyKind: "machina_custody",
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	secret := "whsec_helper"
	header := signWebhook(t, secret, payload, time.Now().Unix())

	helper := NewWebhookHelper(secret, WithWebhookTolerance(time.Minute))
	ev, err := helper.VerifyAndParse(payload, header)
	if err != nil {
		t.Fatalf("VerifyAndParse: %v", err)
	}
	if ev.EventType() != webhooks.TypeWalletCreated {
		t.Errorf("type = %q", ev.EventType())
	}

	wc, ok := ev.(*webhooks.WalletCreatedEvent)
	if !ok {
		t.Fatalf("expected *WalletCreatedEvent, got %T", ev)
	}
	if wc.Data.WalletID != walletID {
		t.Errorf("wallet_id = %v", wc.Data.WalletID)
	}
}

func TestWebhookHelperRejectsBadSignature(t *testing.T) {
	t.Parallel()

	helper := NewWebhookHelper("whsec_xyz")
	_, err := helper.VerifyAndParse([]byte(`{}`), "t=1,v1=deadbeef")
	if err == nil {
		t.Fatal("expected verification error")
	}
}

func TestWebhookHelperUnsafeUnwrap(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"type": "machina.user.created.v1",
		"id":   "evt_x",
		"time": "2026-01-01T00:00:00Z",
		"data": { "user_id": "00000000-0000-0000-0000-000000000001" }
	}`)
	helper := NewWebhookHelper("whsec")
	ev, err := helper.UnsafeUnwrap(body)
	if err != nil {
		t.Fatalf("UnsafeUnwrap: %v", err)
	}
	if ev.EventType() != webhooks.TypeUserCreated {
		t.Errorf("type = %q", ev.EventType())
	}

	if _, err := helper.UnsafeUnwrap(nil); err == nil {
		t.Error("expected validation for empty body")
	}
	if _, err := helper.UnsafeUnwrap([]byte(`not json`)); err == nil {
		t.Error("expected validation for invalid body")
	}
}

func TestWebhookHelperUnknownEvent(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"type": "machina.not.a.real.event.v9",
		"id":   "evt_x",
		"time": "2026-01-01T00:00:00Z",
		"data": { "x": 1 }
	}`)
	helper := NewWebhookHelper("whsec")
	ev, err := helper.UnsafeUnwrap(body)
	if err != nil {
		t.Fatalf("UnsafeUnwrap: %v", err)
	}
	if _, ok := ev.(webhooks.UnknownEvent); !ok {
		t.Errorf("expected UnknownEvent, got %T", ev)
	}
}

func TestWebhookHelperVerifyOnly(t *testing.T) {
	t.Parallel()

	secret := "whsec_xx"
	payload := []byte(`{"hello":"world"}`)
	header := signWebhook(t, secret, payload, time.Now().Unix())

	helper := NewWebhookHelper(secret)
	if err := helper.Verify(payload, header); err != nil {
		t.Errorf("Verify: %v", err)
	}

	// Tamper.
	if err := helper.Verify([]byte(`evil`), header); err == nil {
		t.Error("expected mismatch")
	}
}

func TestWebhookHelperPropagatesAuthSentinel(t *testing.T) {
	t.Parallel()

	// Use a fresh timestamp so the tolerance check passes; only the HMAC
	// should mismatch, surfacing as ErrAuth.
	secret := "whsec_xx"
	helper := NewWebhookHelper(secret)
	header := signWebhook(t, "other-secret", []byte(`{}`), time.Now().Unix())
	err := helper.Verify([]byte(`{}`), header)
	if !errors.Is(err, ErrAuth) {
		t.Errorf("expected ErrAuth sentinel, got %v", err)
	}
}

func TestWebhooksParseAllTypes(t *testing.T) {
	t.Parallel()

	// Spot-check a handful of types across categories to ensure the registry
	// switch covers them.
	types := []string{
		webhooks.TypeWalletCreated,
		webhooks.TypeUserSuspended,
		webhooks.TypeGasTankBalanceLow,
		webhooks.TypeBillingPaymentSucceeded,
		webhooks.TypeRiskScoreUpdated,
		webhooks.TypeSimulationFailed,
		webhooks.TypeSmartAccountDeployed,
		webhooks.TypeUserOperationIncluded,
		webhooks.TypeHwWalletSigningSessionStarted,
		webhooks.TypeNftTransferConfirmed,
		webhooks.TypeTreasuryMovementConfirmed,
		webhooks.TypeWebhooksDelivered,
		webhooks.TypeStakingPositionOpened,
		webhooks.TypeTaxFormGenerated,
		webhooks.TypeNameResolved,
		webhooks.TypeTokenVerified,
		webhooks.TypeYieldAllocationCreated,
		webhooks.TypeSubsChargeSucceeded,
	}

	for _, tp := range types {
		body := []byte(`{"type":"` + tp + `","id":"evt","time":"2026-01-01T00:00:00Z","data":{}}`)
		ev, err := webhooks.Parse(body)
		if err != nil {
			t.Errorf("%s: %v", tp, err)
			continue
		}
		if ev.EventType() != tp {
			t.Errorf("%s parsed as %s", tp, ev.EventType())
		}
		if ev.EventID() != "evt" {
			t.Errorf("%s: event id mismatch %q", tp, ev.EventID())
		}
		if ev.EventTime().IsZero() {
			t.Errorf("%s: zero time", tp)
		}
	}
}

func TestWebhooksParseInvalidJSON(t *testing.T) {
	t.Parallel()

	if _, err := webhooks.Parse([]byte(`not json`)); err == nil {
		t.Error("expected error")
	}
}

func TestWebhookHelperWithToleranceDefault(t *testing.T) {
	t.Parallel()

	h := NewWebhookHelper("x")
	if h.tolerance != DefaultWebhookTolerance {
		t.Errorf("tolerance = %v", h.tolerance)
	}

	h2 := NewWebhookHelper("x", WithWebhookTolerance(2*time.Minute))
	if h2.tolerance != 2*time.Minute {
		t.Errorf("tolerance = %v", h2.tolerance)
	}

	// Negative tolerance is ignored.
	h3 := NewWebhookHelper("x", WithWebhookTolerance(-time.Second))
	if h3.tolerance != DefaultWebhookTolerance {
		t.Errorf("tolerance = %v", h3.tolerance)
	}

	// Coverage placeholder for unused import.
	_ = strings.TrimSpace("")
}
