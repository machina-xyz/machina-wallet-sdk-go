package machinawallet

import (
	"encoding/json"
	"math/big"
	"time"
)

// Chain is a supported blockchain. New chains are added without breaking older
// SDK versions; treat unknown values as opaque strings.
type Chain string

// Supported chains.
const (
	ChainEthereum Chain = "ethereum"
	ChainSolana   Chain = "solana"
	ChainBase     Chain = "base"
	ChainArbitrum Chain = "arbitrum"
	ChainOptimism Chain = "optimism"
	ChainPolygon  Chain = "polygon"
	ChainBitcoin  Chain = "bitcoin"
)

// WalletStatus is the lifecycle state of a wallet record.
type WalletStatus string

// Wallet status values.
const (
	WalletStatusActive     WalletStatus = "active"
	WalletStatusFrozen     WalletStatus = "frozen"
	WalletStatusRecovering WalletStatus = "recovering"
)

// CustodyKind describes the custody arrangement for a wallet.
type CustodyKind string

// Custody values.
const (
	CustodyMachinaCustody   CustodyKind = "machina_custody"
	CustodySelfCustody      CustodyKind = "self_custody"
	CustodySmartAccount     CustodyKind = "smart_account"
	CustodyMpcInstitutional CustodyKind = "mpc_institutional"
)

// TxStatus is the lifecycle state of a wallet transaction.
type TxStatus string

// Transaction status values.
const (
	TxStatusPending   TxStatus = "pending"
	TxStatusSubmitted TxStatus = "submitted"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
)

// Wallet is the wire shape returned by the wallet management API.
type Wallet struct {
	ID           string       `json:"id"`
	OwnerUserID  string       `json:"owner_user_id"`
	Chain        Chain        `json:"chain"`
	Address      string       `json:"address"`
	Status       WalletStatus `json:"status"`
	CustodyKind  CustodyKind  `json:"custody_kind"`
	FrozenReason *string      `json:"frozen_reason,omitempty"`
	TenantID     string       `json:"tenant_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Balance is a single asset balance for a wallet. Amounts are u256 atoms
// serialized as decimal strings to preserve precision.
type Balance struct {
	Chain       Chain   `json:"chain"`
	TokenMint   *string `json:"token_mint,omitempty"`
	AmountAtoms string  `json:"amount_atoms"`
	Decimals    int     `json:"decimals"`
}

// AmountBigInt parses AmountAtoms as a *big.Int. Returns nil and false if the
// value is not a valid integer literal.
func (b Balance) AmountBigInt() (*big.Int, bool) {
	i := new(big.Int)
	_, ok := i.SetString(b.AmountAtoms, 10)
	if !ok {
		return nil, false
	}
	return i, true
}

// TxIntent is a user intent to move value. Used as input to submit, prepare,
// and check policy calls.
type TxIntent struct {
	To          string         `json:"to"`
	AmountAtoms string         `json:"amount_atoms"`
	Chain       Chain          `json:"chain"`
	TokenMint   *string        `json:"token_mint,omitempty"`
	IntentKind  string         `json:"intent_kind,omitempty"`
	Memo        *string        `json:"memo,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UnsignedTx is a server-prepared unsigned transaction returned by
// PrepareTransaction in the self-custody two-step flow.
type UnsignedTx struct {
	UnsignedTxBytes  string         `json:"unsigned_tx_bytes"`
	PayloadToSign    string         `json:"payload_to_sign"`
	Chain            Chain          `json:"chain"`
	PreparedAt       time.Time      `json:"prepared_at"`
	ExpiresAt        time.Time      `json:"expires_at"`
	DelegationPolicy map[string]any `json:"delegation_policy,omitempty"`
}

// SignedTx is a client-signed transaction blob ready for broadcast.
type SignedTx struct {
	SignedTxBytes string `json:"signed_tx_bytes"`
	Chain         Chain  `json:"chain"`
}

// SubmittedTransaction is the response from SubmitTransaction and
// BroadcastTransaction.
type SubmittedTransaction struct {
	TxID        string   `json:"tx_id"`
	ChainTxHash *string  `json:"chain_tx_hash,omitempty"`
	Status      TxStatus `json:"status"`
}

// Transaction is the wire shape of a wallet transaction record.
type Transaction struct {
	ID          string     `json:"id"`
	WalletID    string     `json:"wallet_id"`
	ChainTxHash *string    `json:"chain_tx_hash,omitempty"`
	Status      TxStatus   `json:"status"`
	IntentKind  string     `json:"intent_kind"`
	AmountAtoms string     `json:"amount_atoms"`
	TokenMint   *string    `json:"token_mint,omitempty"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	TenantID    string     `json:"tenant_id"`
	CreatedAt   time.Time  `json:"created_at"`
}

// WalletPage is a paginated list of wallets.
type WalletPage struct {
	Items      []Wallet `json:"items"`
	NextCursor *string  `json:"next_cursor,omitempty"`
}

// TransactionPage is a paginated list of transactions.
type TransactionPage struct {
	Items      []Transaction `json:"items"`
	NextCursor *string       `json:"next_cursor,omitempty"`
}

// SpendPolicy is a policy attached to a wallet that controls outgoing value.
type SpendPolicy struct {
	ID            string           `json:"id"`
	WalletID      string           `json:"wallet_id"`
	Name          string           `json:"name"`
	SchemaVersion string           `json:"schema_version"`
	Rules         []map[string]any `json:"rules"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// NewSpendPolicy is the input shape for CreatePolicy.
type NewSpendPolicy struct {
	Name          string           `json:"name"`
	SchemaVersion string           `json:"schema_version,omitempty"`
	Rules         []map[string]any `json:"rules"`
}

// SpendPolicyPatch is the input shape for UpdatePolicy. Only non-nil fields
// are sent to the server.
type SpendPolicyPatch struct {
	Name  *string           `json:"name,omitempty"`
	Rules *[]map[string]any `json:"rules,omitempty"`
}

// PolicyCheckResult is the result of CheckPolicy. Allowed is false if any
// matching policy denies the intent.
type PolicyCheckResult struct {
	Allowed           bool    `json:"allowed"`
	MatchedPolicyID   *string `json:"matched_policy_id,omitempty"`
	MatchedPolicyName *string `json:"matched_policy_name,omitempty"`
	Reason            *string `json:"reason,omitempty"`
	RequiresApproval  bool    `json:"requires_approval"`
	ApprovalID        *string `json:"approval_id,omitempty"`
}

// CreateWalletRequest is the input shape for CreateWallet.
type CreateWalletRequest struct {
	OwnerUserID string      `json:"owner_user_id"`
	Chain       Chain       `json:"chain"`
	CustodyKind CustodyKind `json:"custody_kind"`
	Label       string      `json:"label,omitempty"`
}

// ListWalletsOptions is the query parameter set for ListWallets.
type ListWalletsOptions struct {
	OwnerUserID string
	Chain       Chain
	Status      WalletStatus
	Limit       int
	Cursor      string
}

// ListTransactionsOptions is the query parameter set for GetTransactions.
type ListTransactionsOptions struct {
	Status TxStatus
	Limit  int
	Cursor string
}

// CloudEvent is a CloudEvents v1.0 envelope as emitted by the OVERWATCH
// engine. Used by webhook receivers.
type CloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Subject         *string         `json:"subject,omitempty"`
	Time            *time.Time      `json:"time,omitempty"`
	DataContentType *string         `json:"datacontenttype,omitempty"`
	Data            json.RawMessage `json:"data,omitempty"`
}
