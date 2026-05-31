// Package webhooks provides typed event payload structs for the MACHINA
// webhook delivery system. Every event_type emitted by the OVERWATCH engine
// has a corresponding struct here; receivers can switch on the Event
// interface to obtain the strongly-typed payload.
//
// The shape of an event on the wire is:
//
//	{
//	  "type": "machina.wallet.created.v1",
//	  "id":   "evt_...",
//	  "time": "2026-05-30T12:00:00Z",
//	  "data": { ... }
//	}
//
// Use the WebhookHelper in the parent package to verify, unwrap, and parse a
// raw request body into an Event in a single call.
package webhooks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is the sealed interface implemented by every typed event. Receivers
// type-switch on a returned Event to access the concrete payload.
type Event interface {
	// EventType returns the wire-level event_type discriminator.
	EventType() string
	// EventID is the per-delivery event id.
	EventID() string
	// EventTime is the publish time on the source service.
	EventTime() time.Time
}

// Envelope is the shared envelope every typed event embeds.
type Envelope struct {
	Type        string          `json:"type"`
	ID          string          `json:"id"`
	Source      string          `json:"source,omitempty"`
	Time        time.Time       `json:"time"`
	SpecVersion string          `json:"specversion,omitempty"`
	Subject     string          `json:"subject,omitempty"`
	TenantID    *uuid.UUID      `json:"tenant_id,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

// EventType returns the wire-level event_type discriminator.
func (e Envelope) EventType() string { return e.Type }

// EventID returns the per-delivery event id.
func (e Envelope) EventID() string { return e.ID }

// EventTime returns the publish time on the source service.
func (e Envelope) EventTime() time.Time { return e.Time }

// ===========================================================================
// Wallet events (10)
// ===========================================================================

// WalletCreatedData is the data payload for machina.wallet.created.v1.
type WalletCreatedData struct {
	WalletID    uuid.UUID `json:"wallet_id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	Chain       string    `json:"chain"`
	Address     string    `json:"address"`
	CustodyKind string    `json:"custody_kind"`
}

// WalletCreatedEvent is the typed event for machina.wallet.created.v1.
type WalletCreatedEvent struct {
	Envelope
	Data WalletCreatedData `json:"data"`
}

// WalletFrozenData is the data payload for machina.wallet.frozen.v1.
type WalletFrozenData struct {
	WalletID uuid.UUID `json:"wallet_id"`
	Reason   string    `json:"reason,omitempty"`
}

// WalletFrozenEvent is the typed event for machina.wallet.frozen.v1.
type WalletFrozenEvent struct {
	Envelope
	Data WalletFrozenData `json:"data"`
}

// WalletUnfrozenData is the data payload for machina.wallet.unfrozen.v1.
type WalletUnfrozenData struct {
	WalletID uuid.UUID `json:"wallet_id"`
}

// WalletUnfrozenEvent is the typed event for machina.wallet.unfrozen.v1.
type WalletUnfrozenEvent struct {
	Envelope
	Data WalletUnfrozenData `json:"data"`
}

// WalletTxSubmittedData is the data payload for machina.wallet.tx_submitted.v1.
type WalletTxSubmittedData struct {
	WalletID    uuid.UUID `json:"wallet_id"`
	TxID        uuid.UUID `json:"tx_id"`
	ChainTxHash string    `json:"chain_tx_hash,omitempty"`
	AmountAtoms string    `json:"amount_atoms"`
}

// WalletTxSubmittedEvent is the typed event for machina.wallet.tx_submitted.v1.
type WalletTxSubmittedEvent struct {
	Envelope
	Data WalletTxSubmittedData `json:"data"`
}

// WalletPolicyAddedData is the data payload for machina.wallet.policy_added.v1.
type WalletPolicyAddedData struct {
	WalletID uuid.UUID `json:"wallet_id"`
	PolicyID uuid.UUID `json:"policy_id"`
	Name     string    `json:"name"`
}

// WalletPolicyAddedEvent is the typed event for machina.wallet.policy_added.v1.
type WalletPolicyAddedEvent struct {
	Envelope
	Data WalletPolicyAddedData `json:"data"`
}

// WalletPolicyUpdatedData is the data payload for machina.wallet.policy_updated.v1.
type WalletPolicyUpdatedData struct {
	WalletID uuid.UUID `json:"wallet_id"`
	PolicyID uuid.UUID `json:"policy_id"`
}

// WalletPolicyUpdatedEvent is the typed event for machina.wallet.policy_updated.v1.
type WalletPolicyUpdatedEvent struct {
	Envelope
	Data WalletPolicyUpdatedData `json:"data"`
}

// WalletPolicyRemovedData is the data payload for machina.wallet.policy_removed.v1.
type WalletPolicyRemovedData struct {
	WalletID uuid.UUID `json:"wallet_id"`
	PolicyID uuid.UUID `json:"policy_id"`
}

// WalletPolicyRemovedEvent is the typed event for machina.wallet.policy_removed.v1.
type WalletPolicyRemovedEvent struct {
	Envelope
	Data WalletPolicyRemovedData `json:"data"`
}

// WalletPolicyCheckedData is the data payload for machina.wallet.policy_checked.v1.
type WalletPolicyCheckedData struct {
	WalletID uuid.UUID `json:"wallet_id"`
	Allowed  bool      `json:"allowed"`
	Reason   string    `json:"reason,omitempty"`
}

// WalletPolicyCheckedEvent is the typed event for machina.wallet.policy_checked.v1.
type WalletPolicyCheckedEvent struct {
	Envelope
	Data WalletPolicyCheckedData `json:"data"`
}

// WalletPolicyApprovalRequiredData is the data payload for
// machina.wallet.policy_approval_required.v1.
type WalletPolicyApprovalRequiredData struct {
	WalletID   uuid.UUID `json:"wallet_id"`
	ApprovalID uuid.UUID `json:"approval_id"`
	PolicyID   uuid.UUID `json:"policy_id"`
}

// WalletPolicyApprovalRequiredEvent is the typed event for
// machina.wallet.policy_approval_required.v1.
type WalletPolicyApprovalRequiredEvent struct {
	Envelope
	Data WalletPolicyApprovalRequiredData `json:"data"`
}

// ===========================================================================
// User events (8)
// ===========================================================================

// UserCreatedData is the data payload for machina.user.created.v1.
type UserCreatedData struct {
	UserID       uuid.UUID `json:"user_id"`
	PrimaryEmail string    `json:"primary_email,omitempty"`
	PrimaryPhone string    `json:"primary_phone,omitempty"`
}

// UserCreatedEvent is the typed event for machina.user.created.v1.
type UserCreatedEvent struct {
	Envelope
	Data UserCreatedData `json:"data"`
}

// UserDeletedData is the data payload for machina.user.deleted.v1.
type UserDeletedData struct {
	UserID uuid.UUID `json:"user_id"`
}

// UserDeletedEvent is the typed event for machina.user.deleted.v1.
type UserDeletedEvent struct {
	Envelope
	Data UserDeletedData `json:"data"`
}

// UserHardDeletedData is the data payload for machina.user.hard_deleted.v1.
type UserHardDeletedData struct {
	UserID uuid.UUID `json:"user_id"`
}

// UserHardDeletedEvent is the typed event for machina.user.hard_deleted.v1.
type UserHardDeletedEvent struct {
	Envelope
	Data UserHardDeletedData `json:"data"`
}

// UserRestoredData is the data payload for machina.user.restored.v1.
type UserRestoredData struct {
	UserID uuid.UUID `json:"user_id"`
}

// UserRestoredEvent is the typed event for machina.user.restored.v1.
type UserRestoredEvent struct {
	Envelope
	Data UserRestoredData `json:"data"`
}

// UserSuspendedData is the data payload for machina.user.suspended.v1.
type UserSuspendedData struct {
	UserID uuid.UUID `json:"user_id"`
	Reason string    `json:"reason,omitempty"`
}

// UserSuspendedEvent is the typed event for machina.user.suspended.v1.
type UserSuspendedEvent struct {
	Envelope
	Data UserSuspendedData `json:"data"`
}

// UserIdentifierLinkedData is the data payload for machina.user.identifier_linked.v1.
type UserIdentifierLinkedData struct {
	UserID       uuid.UUID `json:"user_id"`
	IdentifierID uuid.UUID `json:"identifier_id"`
	Kind         string    `json:"kind"`
}

// UserIdentifierLinkedEvent is the typed event for machina.user.identifier_linked.v1.
type UserIdentifierLinkedEvent struct {
	Envelope
	Data UserIdentifierLinkedData `json:"data"`
}

// UserIdentifierUnlinkedData is the data payload for
// machina.user.identifier_unlinked.v1.
type UserIdentifierUnlinkedData struct {
	UserID       uuid.UUID `json:"user_id"`
	IdentifierID uuid.UUID `json:"identifier_id"`
	Kind         string    `json:"kind"`
}

// UserIdentifierUnlinkedEvent is the typed event for
// machina.user.identifier_unlinked.v1.
type UserIdentifierUnlinkedEvent struct {
	Envelope
	Data UserIdentifierUnlinkedData `json:"data"`
}

// UserCustomMetadataUpdatedData is the data payload for
// machina.user.custom_metadata_updated.v1.
type UserCustomMetadataUpdatedData struct {
	UserID uuid.UUID              `json:"user_id"`
	Diff   map[string]interface{} `json:"diff,omitempty"`
}

// UserCustomMetadataUpdatedEvent is the typed event for
// machina.user.custom_metadata_updated.v1.
type UserCustomMetadataUpdatedEvent struct {
	Envelope
	Data UserCustomMetadataUpdatedData `json:"data"`
}

// UserKycStatusChangedData is the data payload for machina.user.kyc_status_changed.v1.
type UserKycStatusChangedData struct {
	UserID    uuid.UUID `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
}

// UserKycStatusChangedEvent is the typed event for machina.user.kyc_status_changed.v1.
type UserKycStatusChangedEvent struct {
	Envelope
	Data UserKycStatusChangedData `json:"data"`
}

// ===========================================================================
// Gas tank events (10)
// ===========================================================================

// GasTankCreatedData is the data payload for machina.gas_tank.created.v1.
type GasTankCreatedData struct {
	GasTankID uuid.UUID `json:"gas_tank_id"`
	Chain     string    `json:"chain"`
}

// GasTankCreatedEvent is the typed event for machina.gas_tank.created.v1.
type GasTankCreatedEvent struct {
	Envelope
	Data GasTankCreatedData `json:"data"`
}

// GasTankUpdatedData is the data payload for machina.gas_tank.updated.v1.
type GasTankUpdatedData struct {
	GasTankID uuid.UUID `json:"gas_tank_id"`
}

// GasTankUpdatedEvent is the typed event for machina.gas_tank.updated.v1.
type GasTankUpdatedEvent struct {
	Envelope
	Data GasTankUpdatedData `json:"data"`
}

// GasTankBalanceLowData is the data payload for machina.gas_tank.balance_low.v1.
type GasTankBalanceLowData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankBalanceLowEvent is the typed event for machina.gas_tank.balance_low.v1.
type GasTankBalanceLowEvent struct {
	Envelope
	Data GasTankBalanceLowData `json:"data"`
}

// GasTankBalanceCriticalData is the data payload for machina.gas_tank.balance_critical.v1.
type GasTankBalanceCriticalData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankBalanceCriticalEvent is the typed event for
// machina.gas_tank.balance_critical.v1.
type GasTankBalanceCriticalEvent struct {
	Envelope
	Data GasTankBalanceCriticalData `json:"data"`
}

// GasTankDepositDetectedData is the data payload for
// machina.gas_tank.deposit_detected.v1.
type GasTankDepositDetectedData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	TxHash      string    `json:"tx_hash"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankDepositDetectedEvent is the typed event for
// machina.gas_tank.deposit_detected.v1.
type GasTankDepositDetectedEvent struct {
	Envelope
	Data GasTankDepositDetectedData `json:"data"`
}

// GasTankDepositConfirmedData is the data payload for
// machina.gas_tank.deposit_confirmed.v1.
type GasTankDepositConfirmedData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	TxHash      string    `json:"tx_hash"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankDepositConfirmedEvent is the typed event for
// machina.gas_tank.deposit_confirmed.v1.
type GasTankDepositConfirmedEvent struct {
	Envelope
	Data GasTankDepositConfirmedData `json:"data"`
}

// GasTankWithdrawalInitiatedData is the data payload for
// machina.gas_tank.withdrawal_initiated.v1.
type GasTankWithdrawalInitiatedData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankWithdrawalInitiatedEvent is the typed event for
// machina.gas_tank.withdrawal_initiated.v1.
type GasTankWithdrawalInitiatedEvent struct {
	Envelope
	Data GasTankWithdrawalInitiatedData `json:"data"`
}

// GasTankSponsorshipApprovedData is the data payload for
// machina.gas_tank.sponsorship_approved.v1.
type GasTankSponsorshipApprovedData struct {
	GasTankID uuid.UUID `json:"gas_tank_id"`
	UserOpID  uuid.UUID `json:"user_op_id"`
}

// GasTankSponsorshipApprovedEvent is the typed event for
// machina.gas_tank.sponsorship_approved.v1.
type GasTankSponsorshipApprovedEvent struct {
	Envelope
	Data GasTankSponsorshipApprovedData `json:"data"`
}

// GasTankSponsorshipDeniedData is the data payload for
// machina.gas_tank.sponsorship_denied.v1.
type GasTankSponsorshipDeniedData struct {
	GasTankID uuid.UUID `json:"gas_tank_id"`
	UserOpID  uuid.UUID `json:"user_op_id"`
	Reason    string    `json:"reason,omitempty"`
}

// GasTankSponsorshipDeniedEvent is the typed event for
// machina.gas_tank.sponsorship_denied.v1.
type GasTankSponsorshipDeniedEvent struct {
	Envelope
	Data GasTankSponsorshipDeniedData `json:"data"`
}

// GasTankSponsorshipSettledData is the data payload for
// machina.gas_tank.sponsorship_settled.v1.
type GasTankSponsorshipSettledData struct {
	GasTankID   uuid.UUID `json:"gas_tank_id"`
	UserOpID    uuid.UUID `json:"user_op_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// GasTankSponsorshipSettledEvent is the typed event for
// machina.gas_tank.sponsorship_settled.v1.
type GasTankSponsorshipSettledEvent struct {
	Envelope
	Data GasTankSponsorshipSettledData `json:"data"`
}

// ===========================================================================
// Billing events (8)
// ===========================================================================

// BillingInvoiceGeneratedData is the data payload for machina.billing.invoice_generated.v1.
type BillingInvoiceGeneratedData struct {
	InvoiceID   uuid.UUID `json:"invoice_id"`
	AmountAtoms string    `json:"amount_atoms"`
	Currency    string    `json:"currency"`
}

// BillingInvoiceGeneratedEvent is the typed event for machina.billing.invoice_generated.v1.
type BillingInvoiceGeneratedEvent struct {
	Envelope
	Data BillingInvoiceGeneratedData `json:"data"`
}

// BillingInvoicePaidData is the data payload for machina.billing.invoice_paid.v1.
type BillingInvoicePaidData struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	PaidAt    time.Time `json:"paid_at"`
}

// BillingInvoicePaidEvent is the typed event for machina.billing.invoice_paid.v1.
type BillingInvoicePaidEvent struct {
	Envelope
	Data BillingInvoicePaidData `json:"data"`
}

// BillingInvoicePastDueData is the data payload for machina.billing.invoice_past_due.v1.
type BillingInvoicePastDueData struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
}

// BillingInvoicePastDueEvent is the typed event for machina.billing.invoice_past_due.v1.
type BillingInvoicePastDueEvent struct {
	Envelope
	Data BillingInvoicePastDueData `json:"data"`
}

// BillingPaymentSucceededData is the data payload for machina.billing.payment_succeeded.v1.
type BillingPaymentSucceededData struct {
	PaymentID   uuid.UUID `json:"payment_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// BillingPaymentSucceededEvent is the typed event for machina.billing.payment_succeeded.v1.
type BillingPaymentSucceededEvent struct {
	Envelope
	Data BillingPaymentSucceededData `json:"data"`
}

// BillingPaymentFailedData is the data payload for machina.billing.payment_failed.v1.
type BillingPaymentFailedData struct {
	PaymentID uuid.UUID `json:"payment_id"`
	Reason    string    `json:"reason,omitempty"`
}

// BillingPaymentFailedEvent is the typed event for machina.billing.payment_failed.v1.
type BillingPaymentFailedEvent struct {
	Envelope
	Data BillingPaymentFailedData `json:"data"`
}

// BillingSubscriptionCreatedData is the data payload for machina.billing.subscription_created.v1.
type BillingSubscriptionCreatedData struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	PlanID         uuid.UUID `json:"plan_id"`
}

// BillingSubscriptionCreatedEvent is the typed event for
// machina.billing.subscription_created.v1.
type BillingSubscriptionCreatedEvent struct {
	Envelope
	Data BillingSubscriptionCreatedData `json:"data"`
}

// BillingSubscriptionCanceledData is the data payload for
// machina.billing.subscription_canceled.v1.
type BillingSubscriptionCanceledData struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
}

// BillingSubscriptionCanceledEvent is the typed event for
// machina.billing.subscription_canceled.v1.
type BillingSubscriptionCanceledEvent struct {
	Envelope
	Data BillingSubscriptionCanceledData `json:"data"`
}

// BillingQuotaExceededData is the data payload for machina.billing.quota_exceeded.v1.
type BillingQuotaExceededData struct {
	Metric string `json:"metric"`
	Limit  string `json:"limit"`
	Actual string `json:"actual"`
}

// BillingQuotaExceededEvent is the typed event for machina.billing.quota_exceeded.v1.
type BillingQuotaExceededEvent struct {
	Envelope
	Data BillingQuotaExceededData `json:"data"`
}

// ===========================================================================
// Risk events (4)
// ===========================================================================

// RiskScoreUpdatedData is the data payload for machina.risk.score_updated.v1.
type RiskScoreUpdatedData struct {
	Subject string  `json:"subject"`
	Score   float64 `json:"score"`
}

// RiskScoreUpdatedEvent is the typed event for machina.risk.score_updated.v1.
type RiskScoreUpdatedEvent struct {
	Envelope
	Data RiskScoreUpdatedData `json:"data"`
}

// RiskLimitExceededData is the data payload for machina.risk.limit_exceeded.v1.
type RiskLimitExceededData struct {
	Subject string `json:"subject"`
	Limit   string `json:"limit"`
}

// RiskLimitExceededEvent is the typed event for machina.risk.limit_exceeded.v1.
type RiskLimitExceededEvent struct {
	Envelope
	Data RiskLimitExceededData `json:"data"`
}

// RiskPolicyViolatedData is the data payload for machina.risk.policy_violated.v1.
type RiskPolicyViolatedData struct {
	Subject  string `json:"subject"`
	PolicyID string `json:"policy_id"`
}

// RiskPolicyViolatedEvent is the typed event for machina.risk.policy_violated.v1.
type RiskPolicyViolatedEvent struct {
	Envelope
	Data RiskPolicyViolatedData `json:"data"`
}

// RiskThresholdBreachedData is the data payload for machina.risk.threshold_breached.v1.
type RiskThresholdBreachedData struct {
	Subject     string  `json:"subject"`
	Threshold   float64 `json:"threshold"`
	Observation float64 `json:"observation"`
}

// RiskThresholdBreachedEvent is the typed event for machina.risk.threshold_breached.v1.
type RiskThresholdBreachedEvent struct {
	Envelope
	Data RiskThresholdBreachedData `json:"data"`
}

// ===========================================================================
// Simulation events (4)
// ===========================================================================

// SimulationCompletedData is the data payload for machina.simulation.completed.v1.
type SimulationCompletedData struct {
	SimulationID uuid.UUID `json:"simulation_id"`
	Success      bool      `json:"success"`
}

// SimulationCompletedEvent is the typed event for machina.simulation.completed.v1.
type SimulationCompletedEvent struct {
	Envelope
	Data SimulationCompletedData `json:"data"`
}

// SimulationFailedData is the data payload for machina.simulation.failed.v1.
type SimulationFailedData struct {
	SimulationID uuid.UUID `json:"simulation_id"`
	Error        string    `json:"error,omitempty"`
}

// SimulationFailedEvent is the typed event for machina.simulation.failed.v1.
type SimulationFailedEvent struct {
	Envelope
	Data SimulationFailedData `json:"data"`
}

// SimulationRevertDetectedData is the data payload for machina.simulation.revert_detected.v1.
type SimulationRevertDetectedData struct {
	SimulationID uuid.UUID `json:"simulation_id"`
	Reason       string    `json:"reason,omitempty"`
}

// SimulationRevertDetectedEvent is the typed event for machina.simulation.revert_detected.v1.
type SimulationRevertDetectedEvent struct {
	Envelope
	Data SimulationRevertDetectedData `json:"data"`
}

// SimulationWarningRaisedData is the data payload for machina.simulation.warning_raised.v1.
type SimulationWarningRaisedData struct {
	SimulationID uuid.UUID `json:"simulation_id"`
	Warning      string    `json:"warning"`
}

// SimulationWarningRaisedEvent is the typed event for machina.simulation.warning_raised.v1.
type SimulationWarningRaisedEvent struct {
	Envelope
	Data SimulationWarningRaisedData `json:"data"`
}

// ===========================================================================
// Smart account events (4)
// ===========================================================================

// SmartAccountCreatedData is the data payload for machina.smart_account.created.v1.
type SmartAccountCreatedData struct {
	AccountID uuid.UUID `json:"account_id"`
	Chain     string    `json:"chain"`
}

// SmartAccountCreatedEvent is the typed event for machina.smart_account.created.v1.
type SmartAccountCreatedEvent struct {
	Envelope
	Data SmartAccountCreatedData `json:"data"`
}

// SmartAccountDeployedData is the data payload for machina.smart_account.deployed.v1.
type SmartAccountDeployedData struct {
	AccountID uuid.UUID `json:"account_id"`
	Address   string    `json:"address"`
}

// SmartAccountDeployedEvent is the typed event for machina.smart_account.deployed.v1.
type SmartAccountDeployedEvent struct {
	Envelope
	Data SmartAccountDeployedData `json:"data"`
}

// SmartAccountSessionKeyAddedData is the data payload for
// machina.smart_account.session_key_added.v1.
type SmartAccountSessionKeyAddedData struct {
	AccountID uuid.UUID `json:"account_id"`
	KeyID     uuid.UUID `json:"key_id"`
}

// SmartAccountSessionKeyAddedEvent is the typed event for
// machina.smart_account.session_key_added.v1.
type SmartAccountSessionKeyAddedEvent struct {
	Envelope
	Data SmartAccountSessionKeyAddedData `json:"data"`
}

// SmartAccountSessionKeyRevokedData is the data payload for
// machina.smart_account.session_key_revoked.v1.
type SmartAccountSessionKeyRevokedData struct {
	AccountID uuid.UUID `json:"account_id"`
	KeyID     uuid.UUID `json:"key_id"`
}

// SmartAccountSessionKeyRevokedEvent is the typed event for
// machina.smart_account.session_key_revoked.v1.
type SmartAccountSessionKeyRevokedEvent struct {
	Envelope
	Data SmartAccountSessionKeyRevokedData `json:"data"`
}

// ===========================================================================
// UserOperation events (5)
// ===========================================================================

// UserOperationBuiltData is the data payload for machina.user_operation.built.v1.
type UserOperationBuiltData struct {
	UserOpID uuid.UUID `json:"user_op_id"`
}

// UserOperationBuiltEvent is the typed event for machina.user_operation.built.v1.
type UserOperationBuiltEvent struct {
	Envelope
	Data UserOperationBuiltData `json:"data"`
}

// UserOperationSubmittedData is the data payload for machina.user_operation.submitted.v1.
type UserOperationSubmittedData struct {
	UserOpID uuid.UUID `json:"user_op_id"`
	TxHash   string    `json:"tx_hash,omitempty"`
}

// UserOperationSubmittedEvent is the typed event for machina.user_operation.submitted.v1.
type UserOperationSubmittedEvent struct {
	Envelope
	Data UserOperationSubmittedData `json:"data"`
}

// UserOperationIncludedData is the data payload for machina.user_operation.included.v1.
type UserOperationIncludedData struct {
	UserOpID    uuid.UUID `json:"user_op_id"`
	BlockNumber uint64    `json:"block_number"`
}

// UserOperationIncludedEvent is the typed event for machina.user_operation.included.v1.
type UserOperationIncludedEvent struct {
	Envelope
	Data UserOperationIncludedData `json:"data"`
}

// UserOperationFailedData is the data payload for machina.user_operation.failed.v1.
type UserOperationFailedData struct {
	UserOpID uuid.UUID `json:"user_op_id"`
	Reason   string    `json:"reason,omitempty"`
}

// UserOperationFailedEvent is the typed event for machina.user_operation.failed.v1.
type UserOperationFailedEvent struct {
	Envelope
	Data UserOperationFailedData `json:"data"`
}

// UserOperationSponsoredData is the data payload for machina.user_operation.sponsored.v1.
type UserOperationSponsoredData struct {
	UserOpID  uuid.UUID `json:"user_op_id"`
	GasTankID uuid.UUID `json:"gas_tank_id"`
}

// UserOperationSponsoredEvent is the typed event for machina.user_operation.sponsored.v1.
type UserOperationSponsoredEvent struct {
	Envelope
	Data UserOperationSponsoredData `json:"data"`
}

// ===========================================================================
// Hardware wallet events (4)
// ===========================================================================

// HwWalletDeviceRegisteredData is the data payload for machina.hw_wallet.device_registered.v1.
type HwWalletDeviceRegisteredData struct {
	DeviceID uuid.UUID `json:"device_id"`
	Vendor   string    `json:"vendor"`
}

// HwWalletDeviceRegisteredEvent is the typed event for machina.hw_wallet.device_registered.v1.
type HwWalletDeviceRegisteredEvent struct {
	Envelope
	Data HwWalletDeviceRegisteredData `json:"data"`
}

// HwWalletDeviceDeregisteredData is the data payload for machina.hw_wallet.device_deregistered.v1.
type HwWalletDeviceDeregisteredData struct {
	DeviceID uuid.UUID `json:"device_id"`
}

// HwWalletDeviceDeregisteredEvent is the typed event for
// machina.hw_wallet.device_deregistered.v1.
type HwWalletDeviceDeregisteredEvent struct {
	Envelope
	Data HwWalletDeviceDeregisteredData `json:"data"`
}

// HwWalletSigningSessionStartedData is the data payload for
// machina.hw_wallet.signing_session_started.v1.
type HwWalletSigningSessionStartedData struct {
	SessionID uuid.UUID `json:"session_id"`
	DeviceID  uuid.UUID `json:"device_id"`
}

// HwWalletSigningSessionStartedEvent is the typed event for
// machina.hw_wallet.signing_session_started.v1.
type HwWalletSigningSessionStartedEvent struct {
	Envelope
	Data HwWalletSigningSessionStartedData `json:"data"`
}

// HwWalletSigningSessionSignedData is the data payload for
// machina.hw_wallet.signing_session_signed.v1.
type HwWalletSigningSessionSignedData struct {
	SessionID    uuid.UUID `json:"session_id"`
	SignatureHex string    `json:"signature_hex"`
}

// HwWalletSigningSessionSignedEvent is the typed event for
// machina.hw_wallet.signing_session_signed.v1.
type HwWalletSigningSessionSignedEvent struct {
	Envelope
	Data HwWalletSigningSessionSignedData `json:"data"`
}

// ===========================================================================
// NFT events (4)
// ===========================================================================

// NftTransferInitiatedData is the data payload for machina.nft.transfer_initiated.v1.
type NftTransferInitiatedData struct {
	NftID    uuid.UUID `json:"nft_id"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	TokenID  string    `json:"token_id"`
}

// NftTransferInitiatedEvent is the typed event for machina.nft.transfer_initiated.v1.
type NftTransferInitiatedEvent struct {
	Envelope
	Data NftTransferInitiatedData `json:"data"`
}

// NftTransferConfirmedData is the data payload for machina.nft.transfer_confirmed.v1.
type NftTransferConfirmedData struct {
	NftID  uuid.UUID `json:"nft_id"`
	TxHash string    `json:"tx_hash"`
}

// NftTransferConfirmedEvent is the typed event for machina.nft.transfer_confirmed.v1.
type NftTransferConfirmedEvent struct {
	Envelope
	Data NftTransferConfirmedData `json:"data"`
}

// NftTransferFailedData is the data payload for machina.nft.transfer_failed.v1.
type NftTransferFailedData struct {
	NftID  uuid.UUID `json:"nft_id"`
	Reason string    `json:"reason,omitempty"`
}

// NftTransferFailedEvent is the typed event for machina.nft.transfer_failed.v1.
type NftTransferFailedEvent struct {
	Envelope
	Data NftTransferFailedData `json:"data"`
}

// NftCollectionIndexedData is the data payload for machina.nft.collection_indexed.v1.
type NftCollectionIndexedData struct {
	CollectionID uuid.UUID `json:"collection_id"`
	ItemCount    int       `json:"item_count"`
}

// NftCollectionIndexedEvent is the typed event for machina.nft.collection_indexed.v1.
type NftCollectionIndexedEvent struct {
	Envelope
	Data NftCollectionIndexedData `json:"data"`
}

// ===========================================================================
// Treasury events (4)
// ===========================================================================

// TreasuryCreatedData is the data payload for machina.treasury.created.v1.
type TreasuryCreatedData struct {
	TreasuryID uuid.UUID `json:"treasury_id"`
}

// TreasuryCreatedEvent is the typed event for machina.treasury.created.v1.
type TreasuryCreatedEvent struct {
	Envelope
	Data TreasuryCreatedData `json:"data"`
}

// TreasuryMovementInitiatedData is the data payload for
// machina.treasury.movement_initiated.v1.
type TreasuryMovementInitiatedData struct {
	MovementID  uuid.UUID `json:"movement_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// TreasuryMovementInitiatedEvent is the typed event for
// machina.treasury.movement_initiated.v1.
type TreasuryMovementInitiatedEvent struct {
	Envelope
	Data TreasuryMovementInitiatedData `json:"data"`
}

// TreasuryMovementConfirmedData is the data payload for
// machina.treasury.movement_confirmed.v1.
type TreasuryMovementConfirmedData struct {
	MovementID uuid.UUID `json:"movement_id"`
	TxHash     string    `json:"tx_hash,omitempty"`
}

// TreasuryMovementConfirmedEvent is the typed event for
// machina.treasury.movement_confirmed.v1.
type TreasuryMovementConfirmedEvent struct {
	Envelope
	Data TreasuryMovementConfirmedData `json:"data"`
}

// TreasuryRebalanceExecutedData is the data payload for
// machina.treasury.rebalance_executed.v1.
type TreasuryRebalanceExecutedData struct {
	TreasuryID uuid.UUID `json:"treasury_id"`
}

// TreasuryRebalanceExecutedEvent is the typed event for
// machina.treasury.rebalance_executed.v1.
type TreasuryRebalanceExecutedEvent struct {
	Envelope
	Data TreasuryRebalanceExecutedData `json:"data"`
}

// ===========================================================================
// Webhook delivery events (3)
// ===========================================================================

// WebhooksDeliveredData is the data payload for machina.webhooks.delivered.v1.
type WebhooksDeliveredData struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
}

// WebhooksDeliveredEvent is the typed event for machina.webhooks.delivered.v1.
type WebhooksDeliveredEvent struct {
	Envelope
	Data WebhooksDeliveredData `json:"data"`
}

// WebhooksDeliveryFailedData is the data payload for
// machina.webhooks.delivery_failed.v1.
type WebhooksDeliveryFailedData struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	Endpoint   string    `json:"endpoint"`
	Error      string    `json:"error,omitempty"`
}

// WebhooksDeliveryFailedEvent is the typed event for
// machina.webhooks.delivery_failed.v1.
type WebhooksDeliveryFailedEvent struct {
	Envelope
	Data WebhooksDeliveryFailedData `json:"data"`
}

// WebhooksDeliveryDeadLetteredData is the data payload for
// machina.webhooks.delivery_dead_lettered.v1.
type WebhooksDeliveryDeadLetteredData struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	Endpoint   string    `json:"endpoint"`
}

// WebhooksDeliveryDeadLetteredEvent is the typed event for
// machina.webhooks.delivery_dead_lettered.v1.
type WebhooksDeliveryDeadLetteredEvent struct {
	Envelope
	Data WebhooksDeliveryDeadLetteredData `json:"data"`
}

// ===========================================================================
// Other categories — staking, tax, name, token, yield, subscriptions
// ===========================================================================

// StakingPositionOpenedData is the data payload for machina.staking.position_opened.v1.
type StakingPositionOpenedData struct {
	PositionID  uuid.UUID `json:"position_id"`
	AmountAtoms string    `json:"amount_atoms"`
	Validator   string    `json:"validator,omitempty"`
}

// StakingPositionOpenedEvent is the typed event for machina.staking.position_opened.v1.
type StakingPositionOpenedEvent struct {
	Envelope
	Data StakingPositionOpenedData `json:"data"`
}

// StakingPositionClosedData is the data payload for machina.staking.position_closed.v1.
type StakingPositionClosedData struct {
	PositionID uuid.UUID `json:"position_id"`
}

// StakingPositionClosedEvent is the typed event for machina.staking.position_closed.v1.
type StakingPositionClosedEvent struct {
	Envelope
	Data StakingPositionClosedData `json:"data"`
}

// StakingRewardsMarkedData is the data payload for machina.staking.rewards_marked.v1.
type StakingRewardsMarkedData struct {
	PositionID  uuid.UUID `json:"position_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// StakingRewardsMarkedEvent is the typed event for machina.staking.rewards_marked.v1.
type StakingRewardsMarkedEvent struct {
	Envelope
	Data StakingRewardsMarkedData `json:"data"`
}

// StakingSlashingEventData is the data payload for machina.staking.slashing_event.v1.
type StakingSlashingEventData struct {
	PositionID  uuid.UUID `json:"position_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// StakingSlashingEventEvent is the typed event for machina.staking.slashing_event.v1.
type StakingSlashingEventEvent struct {
	Envelope
	Data StakingSlashingEventData `json:"data"`
}

// TaxLotAcquiredData is the data payload for machina.tax.lot_acquired.v1.
type TaxLotAcquiredData struct {
	LotID       uuid.UUID `json:"lot_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// TaxLotAcquiredEvent is the typed event for machina.tax.lot_acquired.v1.
type TaxLotAcquiredEvent struct {
	Envelope
	Data TaxLotAcquiredData `json:"data"`
}

// TaxLotDisposedData is the data payload for machina.tax.lot_disposed.v1.
type TaxLotDisposedData struct {
	LotID       uuid.UUID `json:"lot_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// TaxLotDisposedEvent is the typed event for machina.tax.lot_disposed.v1.
type TaxLotDisposedEvent struct {
	Envelope
	Data TaxLotDisposedData `json:"data"`
}

// TaxFormGeneratedData is the data payload for machina.tax.form_generated.v1.
type TaxFormGeneratedData struct {
	FormID uuid.UUID `json:"form_id"`
	Kind   string    `json:"kind"`
}

// TaxFormGeneratedEvent is the typed event for machina.tax.form_generated.v1.
type TaxFormGeneratedEvent struct {
	Envelope
	Data TaxFormGeneratedData `json:"data"`
}

// NameResolvedData is the data payload for machina.name.resolved.v1.
type NameResolvedData struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// NameResolvedEvent is the typed event for machina.name.resolved.v1.
type NameResolvedEvent struct {
	Envelope
	Data NameResolvedData `json:"data"`
}

// NameReverseResolvedData is the data payload for machina.name.reverse_resolved.v1.
type NameReverseResolvedData struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

// NameReverseResolvedEvent is the typed event for machina.name.reverse_resolved.v1.
type NameReverseResolvedEvent struct {
	Envelope
	Data NameReverseResolvedData `json:"data"`
}

// TokenAddedData is the data payload for machina.token.added.v1.
type TokenAddedData struct {
	TokenID uuid.UUID `json:"token_id"`
	Symbol  string    `json:"symbol"`
}

// TokenAddedEvent is the typed event for machina.token.added.v1.
type TokenAddedEvent struct {
	Envelope
	Data TokenAddedData `json:"data"`
}

// TokenVerifiedData is the data payload for machina.token.verified.v1.
type TokenVerifiedData struct {
	TokenID uuid.UUID `json:"token_id"`
}

// TokenVerifiedEvent is the typed event for machina.token.verified.v1.
type TokenVerifiedEvent struct {
	Envelope
	Data TokenVerifiedData `json:"data"`
}

// TokenBlockedData is the data payload for machina.token.blocked.v1.
type TokenBlockedData struct {
	TokenID uuid.UUID `json:"token_id"`
	Reason  string    `json:"reason,omitempty"`
}

// TokenBlockedEvent is the typed event for machina.token.blocked.v1.
type TokenBlockedEvent struct {
	Envelope
	Data TokenBlockedData `json:"data"`
}

// TokenPriceRefreshedData is the data payload for machina.token.price_refreshed.v1.
type TokenPriceRefreshedData struct {
	TokenID  uuid.UUID `json:"token_id"`
	PriceUsd string    `json:"price_usd"`
}

// TokenPriceRefreshedEvent is the typed event for machina.token.price_refreshed.v1.
type TokenPriceRefreshedEvent struct {
	Envelope
	Data TokenPriceRefreshedData `json:"data"`
}

// YieldAllocationCreatedData is the data payload for machina.yield.allocation_created.v1.
type YieldAllocationCreatedData struct {
	AllocationID uuid.UUID `json:"allocation_id"`
	AmountAtoms  string    `json:"amount_atoms"`
}

// YieldAllocationCreatedEvent is the typed event for machina.yield.allocation_created.v1.
type YieldAllocationCreatedEvent struct {
	Envelope
	Data YieldAllocationCreatedData `json:"data"`
}

// YieldAllocationWithdrawnData is the data payload for machina.yield.allocation_withdrawn.v1.
type YieldAllocationWithdrawnData struct {
	AllocationID uuid.UUID `json:"allocation_id"`
}

// YieldAllocationWithdrawnEvent is the typed event for machina.yield.allocation_withdrawn.v1.
type YieldAllocationWithdrawnEvent struct {
	Envelope
	Data YieldAllocationWithdrawnData `json:"data"`
}

// YieldRebalanceTriggeredData is the data payload for machina.yield.rebalance_triggered.v1.
type YieldRebalanceTriggeredData struct {
	AllocationID uuid.UUID `json:"allocation_id"`
}

// YieldRebalanceTriggeredEvent is the typed event for machina.yield.rebalance_triggered.v1.
type YieldRebalanceTriggeredEvent struct {
	Envelope
	Data YieldRebalanceTriggeredData `json:"data"`
}

// YieldRebalanceCompletedData is the data payload for machina.yield.rebalance_completed.v1.
type YieldRebalanceCompletedData struct {
	AllocationID uuid.UUID `json:"allocation_id"`
}

// YieldRebalanceCompletedEvent is the typed event for machina.yield.rebalance_completed.v1.
type YieldRebalanceCompletedEvent struct {
	Envelope
	Data YieldRebalanceCompletedData `json:"data"`
}

// YieldRebalanceFailedData is the data payload for machina.yield.rebalance_failed.v1.
type YieldRebalanceFailedData struct {
	AllocationID uuid.UUID `json:"allocation_id"`
	Reason       string    `json:"reason,omitempty"`
}

// YieldRebalanceFailedEvent is the typed event for machina.yield.rebalance_failed.v1.
type YieldRebalanceFailedEvent struct {
	Envelope
	Data YieldRebalanceFailedData `json:"data"`
}

// SubsChargeSucceededData is the data payload for machina.subs.charge_succeeded.v1.
type SubsChargeSucceededData struct {
	ChargeID    uuid.UUID `json:"charge_id"`
	AmountAtoms string    `json:"amount_atoms"`
}

// SubsChargeSucceededEvent is the typed event for machina.subs.charge_succeeded.v1.
type SubsChargeSucceededEvent struct {
	Envelope
	Data SubsChargeSucceededData `json:"data"`
}

// SubsChargeFailedData is the data payload for machina.subs.charge_failed.v1.
type SubsChargeFailedData struct {
	ChargeID uuid.UUID `json:"charge_id"`
	Reason   string    `json:"reason,omitempty"`
}

// SubsChargeFailedEvent is the typed event for machina.subs.charge_failed.v1.
type SubsChargeFailedEvent struct {
	Envelope
	Data SubsChargeFailedData `json:"data"`
}

// SubsSubscriptionCreatedData is the data payload for machina.subs.subscription_created.v1.
type SubsSubscriptionCreatedData struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	PlanID         uuid.UUID `json:"plan_id"`
}

// SubsSubscriptionCreatedEvent is the typed event for machina.subs.subscription_created.v1.
type SubsSubscriptionCreatedEvent struct {
	Envelope
	Data SubsSubscriptionCreatedData `json:"data"`
}

// SubsSubscriptionCancelledData is the data payload for machina.subs.subscription_cancelled.v1.
type SubsSubscriptionCancelledData struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
}

// SubsSubscriptionCancelledEvent is the typed event for machina.subs.subscription_cancelled.v1.
type SubsSubscriptionCancelledEvent struct {
	Envelope
	Data SubsSubscriptionCancelledData `json:"data"`
}

// ===========================================================================
// Generic / Unknown
// ===========================================================================

// UnknownEvent is the fallback for events whose type is not recognized by
// this version of the SDK. Receivers can still inspect the raw Data payload.
type UnknownEvent struct {
	Envelope
}

// ===========================================================================
// Type registry
// ===========================================================================

// All registered event type constants. Receivers can compare against these
// instead of typed string literals.
const (
	TypeWalletCreated                = "machina.wallet.created.v1"
	TypeWalletFrozen                 = "machina.wallet.frozen.v1"
	TypeWalletUnfrozen               = "machina.wallet.unfrozen.v1"
	TypeWalletTxSubmitted            = "machina.wallet.tx_submitted.v1"
	TypeWalletPolicyAdded            = "machina.wallet.policy_added.v1"
	TypeWalletPolicyUpdated          = "machina.wallet.policy_updated.v1"
	TypeWalletPolicyRemoved          = "machina.wallet.policy_removed.v1"
	TypeWalletPolicyChecked          = "machina.wallet.policy_checked.v1"
	TypeWalletPolicyApprovalRequired = "machina.wallet.policy_approval_required.v1"

	TypeUserCreated                = "machina.user.created.v1"
	TypeUserDeleted                = "machina.user.deleted.v1"
	TypeUserHardDeleted            = "machina.user.hard_deleted.v1"
	TypeUserRestored               = "machina.user.restored.v1"
	TypeUserSuspended              = "machina.user.suspended.v1"
	TypeUserIdentifierLinked       = "machina.user.identifier_linked.v1"
	TypeUserIdentifierUnlinked     = "machina.user.identifier_unlinked.v1"
	TypeUserCustomMetadataUpdated  = "machina.user.custom_metadata_updated.v1"
	TypeUserKycStatusChanged       = "machina.user.kyc_status_changed.v1"

	TypeGasTankCreated              = "machina.gas_tank.created.v1"
	TypeGasTankUpdated              = "machina.gas_tank.updated.v1"
	TypeGasTankBalanceLow           = "machina.gas_tank.balance_low.v1"
	TypeGasTankBalanceCritical      = "machina.gas_tank.balance_critical.v1"
	TypeGasTankDepositDetected      = "machina.gas_tank.deposit_detected.v1"
	TypeGasTankDepositConfirmed     = "machina.gas_tank.deposit_confirmed.v1"
	TypeGasTankWithdrawalInitiated  = "machina.gas_tank.withdrawal_initiated.v1"
	TypeGasTankSponsorshipApproved  = "machina.gas_tank.sponsorship_approved.v1"
	TypeGasTankSponsorshipDenied    = "machina.gas_tank.sponsorship_denied.v1"
	TypeGasTankSponsorshipSettled   = "machina.gas_tank.sponsorship_settled.v1"

	TypeBillingInvoiceGenerated     = "machina.billing.invoice_generated.v1"
	TypeBillingInvoicePaid          = "machina.billing.invoice_paid.v1"
	TypeBillingInvoicePastDue       = "machina.billing.invoice_past_due.v1"
	TypeBillingPaymentSucceeded     = "machina.billing.payment_succeeded.v1"
	TypeBillingPaymentFailed        = "machina.billing.payment_failed.v1"
	TypeBillingSubscriptionCreated  = "machina.billing.subscription_created.v1"
	TypeBillingSubscriptionCanceled = "machina.billing.subscription_canceled.v1"
	TypeBillingQuotaExceeded        = "machina.billing.quota_exceeded.v1"

	TypeRiskScoreUpdated      = "machina.risk.score_updated.v1"
	TypeRiskLimitExceeded     = "machina.risk.limit_exceeded.v1"
	TypeRiskPolicyViolated    = "machina.risk.policy_violated.v1"
	TypeRiskThresholdBreached = "machina.risk.threshold_breached.v1"

	TypeSimulationCompleted      = "machina.simulation.completed.v1"
	TypeSimulationFailed         = "machina.simulation.failed.v1"
	TypeSimulationRevertDetected = "machina.simulation.revert_detected.v1"
	TypeSimulationWarningRaised  = "machina.simulation.warning_raised.v1"

	TypeSmartAccountCreated           = "machina.smart_account.created.v1"
	TypeSmartAccountDeployed          = "machina.smart_account.deployed.v1"
	TypeSmartAccountSessionKeyAdded   = "machina.smart_account.session_key_added.v1"
	TypeSmartAccountSessionKeyRevoked = "machina.smart_account.session_key_revoked.v1"

	TypeUserOperationBuilt     = "machina.user_operation.built.v1"
	TypeUserOperationSubmitted = "machina.user_operation.submitted.v1"
	TypeUserOperationIncluded  = "machina.user_operation.included.v1"
	TypeUserOperationFailed    = "machina.user_operation.failed.v1"
	TypeUserOperationSponsored = "machina.user_operation.sponsored.v1"

	TypeHwWalletDeviceRegistered      = "machina.hw_wallet.device_registered.v1"
	TypeHwWalletDeviceDeregistered    = "machina.hw_wallet.device_deregistered.v1"
	TypeHwWalletSigningSessionStarted = "machina.hw_wallet.signing_session_started.v1"
	TypeHwWalletSigningSessionSigned  = "machina.hw_wallet.signing_session_signed.v1"

	TypeNftTransferInitiated = "machina.nft.transfer_initiated.v1"
	TypeNftTransferConfirmed = "machina.nft.transfer_confirmed.v1"
	TypeNftTransferFailed    = "machina.nft.transfer_failed.v1"
	TypeNftCollectionIndexed = "machina.nft.collection_indexed.v1"

	TypeTreasuryCreated            = "machina.treasury.created.v1"
	TypeTreasuryMovementInitiated  = "machina.treasury.movement_initiated.v1"
	TypeTreasuryMovementConfirmed  = "machina.treasury.movement_confirmed.v1"
	TypeTreasuryRebalanceExecuted  = "machina.treasury.rebalance_executed.v1"

	TypeWebhooksDelivered            = "machina.webhooks.delivered.v1"
	TypeWebhooksDeliveryFailed       = "machina.webhooks.delivery_failed.v1"
	TypeWebhooksDeliveryDeadLettered = "machina.webhooks.delivery_dead_lettered.v1"

	TypeStakingPositionOpened = "machina.staking.position_opened.v1"
	TypeStakingPositionClosed = "machina.staking.position_closed.v1"
	TypeStakingRewardsMarked  = "machina.staking.rewards_marked.v1"
	TypeStakingSlashingEvent  = "machina.staking.slashing_event.v1"

	TypeTaxLotAcquired   = "machina.tax.lot_acquired.v1"
	TypeTaxLotDisposed   = "machina.tax.lot_disposed.v1"
	TypeTaxFormGenerated = "machina.tax.form_generated.v1"

	TypeNameResolved        = "machina.name.resolved.v1"
	TypeNameReverseResolved = "machina.name.reverse_resolved.v1"

	TypeTokenAdded          = "machina.token.added.v1"
	TypeTokenVerified       = "machina.token.verified.v1"
	TypeTokenBlocked        = "machina.token.blocked.v1"
	TypeTokenPriceRefreshed = "machina.token.price_refreshed.v1"

	TypeYieldAllocationCreated   = "machina.yield.allocation_created.v1"
	TypeYieldAllocationWithdrawn = "machina.yield.allocation_withdrawn.v1"
	TypeYieldRebalanceTriggered  = "machina.yield.rebalance_triggered.v1"
	TypeYieldRebalanceCompleted  = "machina.yield.rebalance_completed.v1"
	TypeYieldRebalanceFailed     = "machina.yield.rebalance_failed.v1"

	TypeSubsChargeSucceeded       = "machina.subs.charge_succeeded.v1"
	TypeSubsChargeFailed          = "machina.subs.charge_failed.v1"
	TypeSubsSubscriptionCreated   = "machina.subs.subscription_created.v1"
	TypeSubsSubscriptionCancelled = "machina.subs.subscription_cancelled.v1"
)

// Parse decodes the raw JSON body of a webhook delivery into the strongest
// typed event the SDK knows. Unknown types return an UnknownEvent with the
// raw payload preserved.
//
// Use this from the parent package's WebhookHelper or directly when the
// caller has already verified the signature.
func Parse(rawBody []byte) (Event, error) {
	var env Envelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return nil, err
	}

	switch env.Type {
	case TypeWalletCreated:
		var e WalletCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletFrozen:
		var e WalletFrozenEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletUnfrozen:
		var e WalletUnfrozenEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletTxSubmitted:
		var e WalletTxSubmittedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletPolicyAdded:
		var e WalletPolicyAddedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletPolicyUpdated:
		var e WalletPolicyUpdatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletPolicyRemoved:
		var e WalletPolicyRemovedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletPolicyChecked:
		var e WalletPolicyCheckedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWalletPolicyApprovalRequired:
		var e WalletPolicyApprovalRequiredEvent
		return decodeInto(rawBody, &e, env)

	case TypeUserCreated:
		var e UserCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserDeleted:
		var e UserDeletedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserHardDeleted:
		var e UserHardDeletedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserRestored:
		var e UserRestoredEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserSuspended:
		var e UserSuspendedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserIdentifierLinked:
		var e UserIdentifierLinkedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserIdentifierUnlinked:
		var e UserIdentifierUnlinkedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserCustomMetadataUpdated:
		var e UserCustomMetadataUpdatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserKycStatusChanged:
		var e UserKycStatusChangedEvent
		return decodeInto(rawBody, &e, env)

	case TypeGasTankCreated:
		var e GasTankCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankUpdated:
		var e GasTankUpdatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankBalanceLow:
		var e GasTankBalanceLowEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankBalanceCritical:
		var e GasTankBalanceCriticalEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankDepositDetected:
		var e GasTankDepositDetectedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankDepositConfirmed:
		var e GasTankDepositConfirmedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankWithdrawalInitiated:
		var e GasTankWithdrawalInitiatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankSponsorshipApproved:
		var e GasTankSponsorshipApprovedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankSponsorshipDenied:
		var e GasTankSponsorshipDeniedEvent
		return decodeInto(rawBody, &e, env)
	case TypeGasTankSponsorshipSettled:
		var e GasTankSponsorshipSettledEvent
		return decodeInto(rawBody, &e, env)

	case TypeBillingInvoiceGenerated:
		var e BillingInvoiceGeneratedEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingInvoicePaid:
		var e BillingInvoicePaidEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingInvoicePastDue:
		var e BillingInvoicePastDueEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingPaymentSucceeded:
		var e BillingPaymentSucceededEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingPaymentFailed:
		var e BillingPaymentFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingSubscriptionCreated:
		var e BillingSubscriptionCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingSubscriptionCanceled:
		var e BillingSubscriptionCanceledEvent
		return decodeInto(rawBody, &e, env)
	case TypeBillingQuotaExceeded:
		var e BillingQuotaExceededEvent
		return decodeInto(rawBody, &e, env)

	case TypeRiskScoreUpdated:
		var e RiskScoreUpdatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeRiskLimitExceeded:
		var e RiskLimitExceededEvent
		return decodeInto(rawBody, &e, env)
	case TypeRiskPolicyViolated:
		var e RiskPolicyViolatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeRiskThresholdBreached:
		var e RiskThresholdBreachedEvent
		return decodeInto(rawBody, &e, env)

	case TypeSimulationCompleted:
		var e SimulationCompletedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSimulationFailed:
		var e SimulationFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSimulationRevertDetected:
		var e SimulationRevertDetectedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSimulationWarningRaised:
		var e SimulationWarningRaisedEvent
		return decodeInto(rawBody, &e, env)

	case TypeSmartAccountCreated:
		var e SmartAccountCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSmartAccountDeployed:
		var e SmartAccountDeployedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSmartAccountSessionKeyAdded:
		var e SmartAccountSessionKeyAddedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSmartAccountSessionKeyRevoked:
		var e SmartAccountSessionKeyRevokedEvent
		return decodeInto(rawBody, &e, env)

	case TypeUserOperationBuilt:
		var e UserOperationBuiltEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserOperationSubmitted:
		var e UserOperationSubmittedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserOperationIncluded:
		var e UserOperationIncludedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserOperationFailed:
		var e UserOperationFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeUserOperationSponsored:
		var e UserOperationSponsoredEvent
		return decodeInto(rawBody, &e, env)

	case TypeHwWalletDeviceRegistered:
		var e HwWalletDeviceRegisteredEvent
		return decodeInto(rawBody, &e, env)
	case TypeHwWalletDeviceDeregistered:
		var e HwWalletDeviceDeregisteredEvent
		return decodeInto(rawBody, &e, env)
	case TypeHwWalletSigningSessionStarted:
		var e HwWalletSigningSessionStartedEvent
		return decodeInto(rawBody, &e, env)
	case TypeHwWalletSigningSessionSigned:
		var e HwWalletSigningSessionSignedEvent
		return decodeInto(rawBody, &e, env)

	case TypeNftTransferInitiated:
		var e NftTransferInitiatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeNftTransferConfirmed:
		var e NftTransferConfirmedEvent
		return decodeInto(rawBody, &e, env)
	case TypeNftTransferFailed:
		var e NftTransferFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeNftCollectionIndexed:
		var e NftCollectionIndexedEvent
		return decodeInto(rawBody, &e, env)

	case TypeTreasuryCreated:
		var e TreasuryCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTreasuryMovementInitiated:
		var e TreasuryMovementInitiatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTreasuryMovementConfirmed:
		var e TreasuryMovementConfirmedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTreasuryRebalanceExecuted:
		var e TreasuryRebalanceExecutedEvent
		return decodeInto(rawBody, &e, env)

	case TypeWebhooksDelivered:
		var e WebhooksDeliveredEvent
		return decodeInto(rawBody, &e, env)
	case TypeWebhooksDeliveryFailed:
		var e WebhooksDeliveryFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeWebhooksDeliveryDeadLettered:
		var e WebhooksDeliveryDeadLetteredEvent
		return decodeInto(rawBody, &e, env)

	case TypeStakingPositionOpened:
		var e StakingPositionOpenedEvent
		return decodeInto(rawBody, &e, env)
	case TypeStakingPositionClosed:
		var e StakingPositionClosedEvent
		return decodeInto(rawBody, &e, env)
	case TypeStakingRewardsMarked:
		var e StakingRewardsMarkedEvent
		return decodeInto(rawBody, &e, env)
	case TypeStakingSlashingEvent:
		var e StakingSlashingEventEvent
		return decodeInto(rawBody, &e, env)

	case TypeTaxLotAcquired:
		var e TaxLotAcquiredEvent
		return decodeInto(rawBody, &e, env)
	case TypeTaxLotDisposed:
		var e TaxLotDisposedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTaxFormGenerated:
		var e TaxFormGeneratedEvent
		return decodeInto(rawBody, &e, env)

	case TypeNameResolved:
		var e NameResolvedEvent
		return decodeInto(rawBody, &e, env)
	case TypeNameReverseResolved:
		var e NameReverseResolvedEvent
		return decodeInto(rawBody, &e, env)

	case TypeTokenAdded:
		var e TokenAddedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTokenVerified:
		var e TokenVerifiedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTokenBlocked:
		var e TokenBlockedEvent
		return decodeInto(rawBody, &e, env)
	case TypeTokenPriceRefreshed:
		var e TokenPriceRefreshedEvent
		return decodeInto(rawBody, &e, env)

	case TypeYieldAllocationCreated:
		var e YieldAllocationCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeYieldAllocationWithdrawn:
		var e YieldAllocationWithdrawnEvent
		return decodeInto(rawBody, &e, env)
	case TypeYieldRebalanceTriggered:
		var e YieldRebalanceTriggeredEvent
		return decodeInto(rawBody, &e, env)
	case TypeYieldRebalanceCompleted:
		var e YieldRebalanceCompletedEvent
		return decodeInto(rawBody, &e, env)
	case TypeYieldRebalanceFailed:
		var e YieldRebalanceFailedEvent
		return decodeInto(rawBody, &e, env)

	case TypeSubsChargeSucceeded:
		var e SubsChargeSucceededEvent
		return decodeInto(rawBody, &e, env)
	case TypeSubsChargeFailed:
		var e SubsChargeFailedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSubsSubscriptionCreated:
		var e SubsSubscriptionCreatedEvent
		return decodeInto(rawBody, &e, env)
	case TypeSubsSubscriptionCancelled:
		var e SubsSubscriptionCancelledEvent
		return decodeInto(rawBody, &e, env)

	default:
		return UnknownEvent{Envelope: env}, nil
	}
}

// decodeInto is a small helper that decodes the raw body into the typed
// event and returns it as an Event interface. It is a no-op for the parsed
// envelope — Go's encoding/json populates the embedded struct in-place.
func decodeInto(raw []byte, out Event, _ Envelope) (Event, error) {
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, err
	}
	return out, nil
}
