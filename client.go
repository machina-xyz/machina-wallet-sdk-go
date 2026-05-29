package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Client is the primary SDK entrypoint. A Client is safe for concurrent use by
// multiple goroutines.
type Client struct {
	cfg       Config
	signer    Signer
	transport *transport
}

// NewClient validates cfg and constructs a Client. It returns an error if cfg
// is missing required fields or the app secret is malformed.
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg = cfg.withDefaults()
	signer, err := NewSigner(cfg.AppSecret)
	if err != nil {
		return nil, err
	}
	c := &Client{cfg: cfg, signer: signer}
	c.transport = newTransport(cfg, signer)
	return c, nil
}

// Config returns the resolved configuration in use by the client. The returned
// value is a copy and may be modified without affecting the client.
func (c *Client) Config() Config { return c.cfg }

// ListWallets returns wallets owned by the application. Pass nil opts for the
// default paging.
func (c *Client) ListWallets(ctx context.Context, opts *ListWalletsOptions) (*WalletPage, error) {
	q := url.Values{}
	if opts != nil {
		setIfNonEmpty(q, "owner_user_id", opts.OwnerUserID)
		setIfNonEmpty(q, "chain", string(opts.Chain))
		setIfNonEmpty(q, "status", string(opts.Status))
		setIfNonZero(q, "limit", opts.Limit)
		setIfNonEmpty(q, "cursor", opts.Cursor)
	}
	var page WalletPage
	if err := c.transport.request(ctx, http.MethodGet, "/v1/wallets", q, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// GetWallet fetches a single wallet by id.
func (c *Client) GetWallet(ctx context.Context, walletID string) (*Wallet, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var w Wallet
	if err := c.transport.request(ctx, http.MethodGet, fmt.Sprintf("/v1/wallets/%s", walletID), nil, nil, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// CreateWallet provisions a new wallet for the given owner and chain.
func (c *Client) CreateWallet(ctx context.Context, req *CreateWalletRequest) (*Wallet, error) {
	if req == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "CreateWalletRequest is required"}
	}
	var w Wallet
	if err := c.transport.request(ctx, http.MethodPost, "/v1/wallets", nil, req, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// GetBalances returns the on-chain balances for a wallet.
func (c *Client) GetBalances(ctx context.Context, walletID string) ([]Balance, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var balances []Balance
	if err := c.transport.request(ctx, http.MethodGet, fmt.Sprintf("/v1/wallets/%s/balances", walletID), nil, nil, &balances); err != nil {
		return nil, err
	}
	return balances, nil
}

// GetTransactions returns a page of transactions for a wallet.
func (c *Client) GetTransactions(ctx context.Context, walletID string, opts *ListTransactionsOptions) (*TransactionPage, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	q := url.Values{}
	if opts != nil {
		setIfNonEmpty(q, "status", string(opts.Status))
		setIfNonZero(q, "limit", opts.Limit)
		setIfNonEmpty(q, "cursor", opts.Cursor)
	}
	var page TransactionPage
	if err := c.transport.request(ctx, http.MethodGet, fmt.Sprintf("/v1/wallets/%s/transactions", walletID), q, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// SubmitTransaction submits an intent to a custodial wallet for one-shot
// signing and broadcast. Returns an ApprovalRequiredError if a policy requires
// human approval and a PolicyDeniedError if the intent is denied.
func (c *Client) SubmitTransaction(ctx context.Context, walletID string, intent TxIntent) (*SubmittedTransaction, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var out SubmittedTransaction
	if err := c.transport.request(ctx, http.MethodPost, fmt.Sprintf("/v1/wallets/%s/transactions", walletID), nil, intent, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PrepareTransaction is step 1 of the self-custody two-step flow. Returns the
// unsigned transaction payload for the caller to sign locally.
func (c *Client) PrepareTransaction(ctx context.Context, walletID string, intent TxIntent) (*UnsignedTx, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var out UnsignedTx
	if err := c.transport.request(ctx, http.MethodPost, fmt.Sprintf("/v1/wallets/%s/transactions/prepare", walletID), nil, intent, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BroadcastTransaction is step 2 of the self-custody two-step flow. Submits
// the client-signed payload to the network via the wallet service.
func (c *Client) BroadcastTransaction(ctx context.Context, walletID string, signed SignedTx) (*SubmittedTransaction, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var out SubmittedTransaction
	if err := c.transport.request(ctx, http.MethodPost, fmt.Sprintf("/v1/wallets/%s/transactions/broadcast", walletID), nil, signed, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPolicies returns all spend policies attached to a wallet.
func (c *Client) ListPolicies(ctx context.Context, walletID string) ([]SpendPolicy, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var policies []SpendPolicy
	if err := c.transport.request(ctx, http.MethodGet, fmt.Sprintf("/v1/wallets/%s/policies", walletID), nil, nil, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// CreatePolicy attaches a new spend policy to a wallet.
func (c *Client) CreatePolicy(ctx context.Context, walletID string, p NewSpendPolicy) (*SpendPolicy, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	if p.SchemaVersion == "" {
		p.SchemaVersion = "v1"
	}
	if p.Rules == nil {
		p.Rules = []map[string]any{}
	}
	var out SpendPolicy
	if err := c.transport.request(ctx, http.MethodPost, fmt.Sprintf("/v1/wallets/%s/policies", walletID), nil, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePolicy applies a partial update to a policy.
func (c *Client) UpdatePolicy(ctx context.Context, walletID, policyID string, patch SpendPolicyPatch) (*SpendPolicy, error) {
	if walletID == "" || policyID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID and policyID are required"}
	}
	var out SpendPolicy
	if err := c.transport.request(ctx, http.MethodPatch, fmt.Sprintf("/v1/wallets/%s/policies/%s", walletID, policyID), nil, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePolicy removes a spend policy from a wallet.
func (c *Client) DeletePolicy(ctx context.Context, walletID, policyID string) error {
	if walletID == "" || policyID == "" {
		return &Error{Code: ErrCodeValidation, Message: "walletID and policyID are required"}
	}
	return c.transport.request(ctx, http.MethodDelete, fmt.Sprintf("/v1/wallets/%s/policies/%s", walletID, policyID), nil, nil, nil)
}

// CheckPolicy evaluates an intent against the wallet's policies without
// submitting it for execution.
func (c *Client) CheckPolicy(ctx context.Context, walletID string, intent TxIntent) (*PolicyCheckResult, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	var out PolicyCheckResult
	if err := c.transport.request(ctx, http.MethodPost, fmt.Sprintf("/v1/wallets/%s/policies/check", walletID), nil, intent, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
