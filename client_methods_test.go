package machinawallet

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestClientAllMethods exercises every public client method against a single
// dispatching mock server. It is a coverage-broad smoke test, not a
// behavioral test — those live in their own files.
func TestClientAllMethods(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/wallets" && r.Method == http.MethodPost:
			_, _ = io.Copy(io.Discard, r.Body)
			_ = json.NewEncoder(w).Encode(Wallet{ID: "wal_new"})
		case r.URL.Path == "/v1/wallets/wal_x/balances":
			_ = json.NewEncoder(w).Encode([]Balance{{Chain: ChainEthereum, AmountAtoms: "1000", Decimals: 18}})
		case r.URL.Path == "/v1/wallets/wal_x/transactions" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(TransactionPage{Items: []Transaction{}})
		case r.URL.Path == "/v1/wallets/wal_x/transactions/prepare":
			_ = json.NewEncoder(w).Encode(UnsignedTx{UnsignedTxBytes: "0xdeadbeef", Chain: ChainEthereum})
		case r.URL.Path == "/v1/wallets/wal_x/transactions/broadcast":
			_ = json.NewEncoder(w).Encode(SubmittedTransaction{TxID: "tx_1", Status: TxStatusSubmitted})
		case r.URL.Path == "/v1/wallets/wal_x/policies" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode([]SpendPolicy{{ID: "pol_1"}})
		case r.URL.Path == "/v1/wallets/wal_x/policies/pol_1" && r.Method == http.MethodPatch:
			_ = json.NewEncoder(w).Encode(SpendPolicy{ID: "pol_1", Name: "renamed"})
		case r.URL.Path == "/v1/wallets/wal_x/policies/check":
			_ = json.NewEncoder(w).Encode(PolicyCheckResult{Allowed: true})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	ctx := context.Background()

	if _, err := c.CreateWallet(ctx, &CreateWalletRequest{OwnerUserID: "u", Chain: ChainEthereum, CustodyKind: CustodyMachinaCustody}); err != nil {
		t.Errorf("CreateWallet: %v", err)
	}
	if _, err := c.CreateWallet(ctx, nil); err == nil {
		t.Errorf("expected nil-request error")
	}

	balances, err := c.GetBalances(ctx, "wal_x")
	if err != nil {
		t.Errorf("GetBalances: %v", err)
	}
	if len(balances) != 1 {
		t.Errorf("balances len = %d", len(balances))
	}
	if _, err := c.GetBalances(ctx, ""); err == nil {
		t.Error("expected empty-id error")
	}
	if i, ok := balances[0].AmountBigInt(); !ok || i.Int64() != 1000 {
		t.Errorf("AmountBigInt = %v %v", i, ok)
	}

	if _, err := c.GetTransactions(ctx, "wal_x", &ListTransactionsOptions{Status: TxStatusConfirmed, Limit: 10, Cursor: "abc"}); err != nil {
		t.Errorf("GetTransactions: %v", err)
	}
	if _, err := c.GetTransactions(ctx, "", nil); err == nil {
		t.Error("expected empty-id error")
	}

	if _, err := c.PrepareTransaction(ctx, "wal_x", TxIntent{To: "0x"}); err != nil {
		t.Errorf("PrepareTransaction: %v", err)
	}
	if _, err := c.PrepareTransaction(ctx, "", TxIntent{}); err == nil {
		t.Error("expected empty-id error")
	}

	if _, err := c.BroadcastTransaction(ctx, "wal_x", SignedTx{SignedTxBytes: "0xabc"}); err != nil {
		t.Errorf("BroadcastTransaction: %v", err)
	}
	if _, err := c.BroadcastTransaction(ctx, "", SignedTx{}); err == nil {
		t.Error("expected empty-id error")
	}

	if _, err := c.ListPolicies(ctx, "wal_x"); err != nil {
		t.Errorf("ListPolicies: %v", err)
	}
	if _, err := c.ListPolicies(ctx, ""); err == nil {
		t.Error("expected empty-id error")
	}

	name := "renamed"
	if _, err := c.UpdatePolicy(ctx, "wal_x", "pol_1", SpendPolicyPatch{Name: &name}); err != nil {
		t.Errorf("UpdatePolicy: %v", err)
	}
	if _, err := c.UpdatePolicy(ctx, "", "", SpendPolicyPatch{}); err == nil {
		t.Error("expected empty-id error")
	}

	if _, err := c.CheckPolicy(ctx, "wal_x", TxIntent{To: "0x"}); err != nil {
		t.Errorf("CheckPolicy: %v", err)
	}
	if _, err := c.CheckPolicy(ctx, "", TxIntent{}); err == nil {
		t.Error("expected empty-id error")
	}

	cfg := c.Config()
	if cfg.AppID != "app_test" {
		t.Errorf("Config().AppID = %q", cfg.AppID)
	}
}

func TestSubmitTransactionEmptyID(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	if _, err := c.SubmitTransaction(context.Background(), "", TxIntent{}); err == nil {
		t.Error("expected empty-id error")
	}
}

func TestDeletePolicyEmptyID(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	if err := c.DeletePolicy(context.Background(), "", ""); err == nil {
		t.Error("expected empty-id error")
	}
}

func TestStreamTransactionsEmptyID(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	if _, _, err := c.StreamTransactions(context.Background(), ""); err == nil {
		t.Error("expected empty-id error")
	}
}

func TestStreamTransactionsErrorStatus(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"auth_error","message":"nope"}}`))
	})
	if _, _, err := c.StreamTransactions(context.Background(), "wal_x"); err == nil {
		t.Error("expected error for 403")
	}
}
