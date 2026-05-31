package machinawallet

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGasSpendByWallet(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/gas_spend/wallets/") {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("window_kind") != "day" {
			t.Errorf("window_kind = %q", r.URL.Query().Get("window_kind"))
		}
		if r.URL.Query().Get("chain") != "ethereum" {
			t.Errorf("chain = %q", r.URL.Query().Get("chain"))
		}
		_ = json.NewEncoder(w).Encode(GasSpendReport{
			Chain:        ChainEthereum,
			WindowKind:   "day",
			StartTime:    time.Now().Add(-24 * time.Hour).UTC(),
			EndTime:      time.Now().UTC(),
			TotalAtoms:   "1000",
			TotalTxCount: 5,
			Series:       []GasSpendPoint{{Window: "2026-05-30", SpentAtoms: "1000", TxCount: 5}},
		})
	})

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	got, err := c.GasSpend().GetByWallet(context.Background(), "wal_x", &GasSpendOptions{
		Chain: ChainEthereum, WindowKind: "day", StartTime: &start, EndTime: &end,
	})
	if err != nil {
		t.Fatalf("GetByWallet: %v", err)
	}
	if got.TotalAtoms != "1000" {
		t.Errorf("total = %q", got.TotalAtoms)
	}

	if _, err := c.GasSpend().GetByWallet(context.Background(), "", nil); err == nil {
		t.Error("expected validation")
	}
}

func TestGasSpendByTenant(t *testing.T) {
	t.Parallel()

	tid := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, tid.String()) {
			t.Errorf("missing tenant id in path: %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(GasSpendReport{TotalAtoms: "42"})
	})

	got, err := c.GasSpend().GetByTenant(context.Background(), tid, nil)
	if err != nil {
		t.Fatalf("GetByTenant: %v", err)
	}
	if got.TotalAtoms != "42" {
		t.Errorf("total = %q", got.TotalAtoms)
	}

	if _, err := c.GasSpend().GetByTenant(context.Background(), uuid.Nil, nil); err == nil {
		t.Error("expected validation")
	}
}
