package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// GasSpendPoint is a single point in a gas-spend time series.
type GasSpendPoint struct {
	Window         string `json:"window"`
	SpentAtoms     string `json:"spent_atoms"`
	SponsoredAtoms string `json:"sponsored_atoms,omitempty"`
	TxCount        int    `json:"tx_count"`
}

// GasSpendReport aggregates a series of GasSpendPoint records.
type GasSpendReport struct {
	Chain        Chain           `json:"chain"`
	WindowKind   string          `json:"window_kind"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	TotalAtoms   string          `json:"total_atoms"`
	TotalTxCount int             `json:"total_tx_count"`
	Series       []GasSpendPoint `json:"series"`
}

// GasSpendOptions controls window selection for GasSpend queries.
type GasSpendOptions struct {
	Chain      Chain
	WindowKind string // "hour" | "day" | "week" | "month"
	StartTime  *time.Time
	EndTime    *time.Time
}

// GasSpendService is the typed wrapper for /v1/gas_spend.
type GasSpendService struct{ client *Client }

// GasSpend returns the typed GasSpendService for this client.
func (c *Client) GasSpend() *GasSpendService { return &GasSpendService{client: c} }

func (s *GasSpendService) queryFromOpts(opts *GasSpendOptions) url.Values {
	q := url.Values{}
	if opts == nil {
		return q
	}
	setIfNonEmpty(q, "chain", string(opts.Chain))
	setIfNonEmpty(q, "window_kind", opts.WindowKind)
	if opts.StartTime != nil {
		q.Set("start_time", opts.StartTime.UTC().Format(time.RFC3339))
	}
	if opts.EndTime != nil {
		q.Set("end_time", opts.EndTime.UTC().Format(time.RFC3339))
	}
	return q
}

// GetByWallet returns a gas-spend report scoped to a single wallet.
func (s *GasSpendService) GetByWallet(ctx context.Context, walletID string, opts *GasSpendOptions) (*GasSpendReport, error) {
	if walletID == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}
	q := s.queryFromOpts(opts)
	var out GasSpendReport
	path := fmt.Sprintf("/v1/gas_spend/wallets/%s", walletID)
	if err := s.client.transport.request(ctx, http.MethodGet, path, q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetByTenant returns a gas-spend report scoped to a tenant.
func (s *GasSpendService) GetByTenant(ctx context.Context, tenantID uuid.UUID, opts *GasSpendOptions) (*GasSpendReport, error) {
	if tenantID == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "tenantID is required"}
	}
	q := s.queryFromOpts(opts)
	var out GasSpendReport
	path := fmt.Sprintf("/v1/gas_spend/tenants/%s", tenantID.String())
	if err := s.client.transport.request(ctx, http.MethodGet, path, q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
