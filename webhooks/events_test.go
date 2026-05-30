package webhooks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// allRegisteredTypes is the canonical list of typed event_type constants
// exposed by this package. The Parse switch must cover every entry.
var allRegisteredTypes = []string{
	TypeWalletCreated, TypeWalletFrozen, TypeWalletUnfrozen, TypeWalletTxSubmitted,
	TypeWalletPolicyAdded, TypeWalletPolicyUpdated, TypeWalletPolicyRemoved,
	TypeWalletPolicyChecked, TypeWalletPolicyApprovalRequired,

	TypeUserCreated, TypeUserDeleted, TypeUserHardDeleted, TypeUserRestored,
	TypeUserSuspended, TypeUserIdentifierLinked, TypeUserIdentifierUnlinked,
	TypeUserCustomMetadataUpdated, TypeUserKycStatusChanged,

	TypeGasTankCreated, TypeGasTankUpdated, TypeGasTankBalanceLow,
	TypeGasTankBalanceCritical, TypeGasTankDepositDetected,
	TypeGasTankDepositConfirmed, TypeGasTankWithdrawalInitiated,
	TypeGasTankSponsorshipApproved, TypeGasTankSponsorshipDenied,
	TypeGasTankSponsorshipSettled,

	TypeBillingInvoiceGenerated, TypeBillingInvoicePaid, TypeBillingInvoicePastDue,
	TypeBillingPaymentSucceeded, TypeBillingPaymentFailed,
	TypeBillingSubscriptionCreated, TypeBillingSubscriptionCanceled,
	TypeBillingQuotaExceeded,

	TypeRiskScoreUpdated, TypeRiskLimitExceeded, TypeRiskPolicyViolated,
	TypeRiskThresholdBreached,

	TypeSimulationCompleted, TypeSimulationFailed, TypeSimulationRevertDetected,
	TypeSimulationWarningRaised,

	TypeSmartAccountCreated, TypeSmartAccountDeployed,
	TypeSmartAccountSessionKeyAdded, TypeSmartAccountSessionKeyRevoked,

	TypeUserOperationBuilt, TypeUserOperationSubmitted, TypeUserOperationIncluded,
	TypeUserOperationFailed, TypeUserOperationSponsored,

	TypeHwWalletDeviceRegistered, TypeHwWalletDeviceDeregistered,
	TypeHwWalletSigningSessionStarted, TypeHwWalletSigningSessionSigned,

	TypeNftTransferInitiated, TypeNftTransferConfirmed, TypeNftTransferFailed,
	TypeNftCollectionIndexed,

	TypeTreasuryCreated, TypeTreasuryMovementInitiated,
	TypeTreasuryMovementConfirmed, TypeTreasuryRebalanceExecuted,

	TypeWebhooksDelivered, TypeWebhooksDeliveryFailed, TypeWebhooksDeliveryDeadLettered,

	TypeStakingPositionOpened, TypeStakingPositionClosed, TypeStakingRewardsMarked,
	TypeStakingSlashingEvent,

	TypeTaxLotAcquired, TypeTaxLotDisposed, TypeTaxFormGenerated,

	TypeNameResolved, TypeNameReverseResolved,

	TypeTokenAdded, TypeTokenVerified, TypeTokenBlocked, TypeTokenPriceRefreshed,

	TypeYieldAllocationCreated, TypeYieldAllocationWithdrawn,
	TypeYieldRebalanceTriggered, TypeYieldRebalanceCompleted, TypeYieldRebalanceFailed,

	TypeSubsChargeSucceeded, TypeSubsChargeFailed,
	TypeSubsSubscriptionCreated, TypeSubsSubscriptionCancelled,
}

func TestParseEveryRegisteredType(t *testing.T) {
	t.Parallel()

	if len(allRegisteredTypes) < 64 {
		t.Fatalf("registered type count = %d, want at least 64", len(allRegisteredTypes))
	}

	for _, tp := range allRegisteredTypes {
		body := []byte(`{"type":"` + tp + `","id":"evt","time":"2026-01-01T00:00:00Z","data":{}}`)
		ev, err := Parse(body)
		if err != nil {
			t.Errorf("%s parse: %v", tp, err)
			continue
		}
		if ev.EventType() != tp {
			t.Errorf("%s decoded as %s", tp, ev.EventType())
		}
		if _, ok := ev.(UnknownEvent); ok {
			t.Errorf("%s decoded as UnknownEvent", tp)
		}
	}
}

func TestParseUnknown(t *testing.T) {
	t.Parallel()

	ev, err := Parse([]byte(`{"type":"machina.unknown.v9","id":"e","time":"2026-01-01T00:00:00Z","data":{}}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, ok := ev.(UnknownEvent); !ok {
		t.Errorf("expected UnknownEvent, got %T", ev)
	}
	if ev.EventID() != "e" {
		t.Errorf("event id = %q", ev.EventID())
	}
}

func TestParseInvalidJSON(t *testing.T) {
	t.Parallel()

	if _, err := Parse([]byte(`not-json`)); err == nil {
		t.Error("expected error")
	}
}

func TestEnvelopeMethods(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.New()
	body, err := json.Marshal(WalletCreatedEvent{
		Envelope: Envelope{
			Type:     TypeWalletCreated,
			ID:       "evt_1",
			Time:     now,
			TenantID: &id,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ev, err := Parse(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ev.EventID() != "evt_1" {
		t.Errorf("id = %q", ev.EventID())
	}
	if !ev.EventTime().Equal(now) {
		t.Errorf("time = %v, want %v", ev.EventTime(), now)
	}
}
