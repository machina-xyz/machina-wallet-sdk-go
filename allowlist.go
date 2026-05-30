package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// AllowlistEntryKind discriminates an allowlist entry.
type AllowlistEntryKind string

// Supported allowlist entry kinds.
const (
	AllowlistKindAddress AllowlistEntryKind = "address"
	AllowlistKindDomain  AllowlistEntryKind = "domain"
	AllowlistKindIP      AllowlistEntryKind = "ip"
)

// AllowlistEntry is the wire shape of a single allowlist entry.
type AllowlistEntry struct {
	ID        uuid.UUID          `json:"id"`
	TenantID  uuid.UUID          `json:"tenant_id"`
	Kind      AllowlistEntryKind `json:"kind"`
	Value     string             `json:"value"`
	Chain     *Chain             `json:"chain,omitempty"`
	Label     string             `json:"label,omitempty"`
	Enabled   bool               `json:"enabled"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// AllowlistPage is a paginated list of allowlist entries.
type AllowlistPage struct {
	Items      []AllowlistEntry `json:"items"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

// ListAllowlistOptions is the query parameter set for List.
type ListAllowlistOptions struct {
	Kind    AllowlistEntryKind
	Chain   Chain
	Enabled *bool
	Limit   int
	Cursor  string
}

// CreateAllowlistEntryRequest is the input shape for Create.
type CreateAllowlistEntryRequest struct {
	Kind  AllowlistEntryKind `json:"kind"`
	Value string             `json:"value"`
	Chain *Chain             `json:"chain,omitempty"`
	Label string             `json:"label,omitempty"`
}

// UpdateAllowlistEntryRequest is the input shape for Update.
type UpdateAllowlistEntryRequest struct {
	Label   *string `json:"label,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// AllowlistService is the typed wrapper for /v1/allowlist.
type AllowlistService struct{ client *Client }

// Allowlist returns the typed AllowlistService for this client.
func (c *Client) Allowlist() *AllowlistService { return &AllowlistService{client: c} }

// List returns allowlist entries for the calling tenant.
func (s *AllowlistService) List(ctx context.Context, opts *ListAllowlistOptions) (*AllowlistPage, error) {
	q := url.Values{}
	if opts != nil {
		setIfNonEmpty(q, "kind", string(opts.Kind))
		setIfNonEmpty(q, "chain", string(opts.Chain))
		if opts.Enabled != nil {
			if *opts.Enabled {
				q.Set("enabled", "true")
			} else {
				q.Set("enabled", "false")
			}
		}
		setIfNonZero(q, "limit", opts.Limit)
		setIfNonEmpty(q, "cursor", opts.Cursor)
	}
	var page AllowlistPage
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/allowlist", q, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get fetches a single allowlist entry.
func (s *AllowlistService) Get(ctx context.Context, id uuid.UUID) (*AllowlistEntry, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "allowlist id is required"}
	}
	var e AllowlistEntry
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/allowlist/"+id.String(), nil, nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Create adds an entry to the allowlist.
func (s *AllowlistService) Create(ctx context.Context, req *CreateAllowlistEntryRequest) (*AllowlistEntry, error) {
	if req == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "CreateAllowlistEntryRequest is required"}
	}
	if req.Kind == "" || req.Value == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "Kind and Value are required"}
	}
	var out AllowlistEntry
	if err := s.client.transport.request(ctx, http.MethodPost, "/v1/allowlist", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to an allowlist entry.
func (s *AllowlistService) Update(ctx context.Context, id uuid.UUID, patch *UpdateAllowlistEntryRequest) (*AllowlistEntry, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "allowlist id is required"}
	}
	if patch == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "UpdateAllowlistEntryRequest is required"}
	}
	var out AllowlistEntry
	if err := s.client.transport.request(ctx, http.MethodPatch, "/v1/allowlist/"+id.String(), nil, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an allowlist entry.
func (s *AllowlistService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return &Error{Code: ErrCodeValidation, Message: "allowlist id is required"}
	}
	return s.client.transport.request(ctx, http.MethodDelete, "/v1/allowlist/"+id.String(), nil, nil, nil)
}

// Enable toggles an entry on.
func (s *AllowlistService) Enable(ctx context.Context, id uuid.UUID) (*AllowlistEntry, error) {
	return s.toggle(ctx, id, true)
}

// Disable toggles an entry off without removing it.
func (s *AllowlistService) Disable(ctx context.Context, id uuid.UUID) (*AllowlistEntry, error) {
	return s.toggle(ctx, id, false)
}

func (s *AllowlistService) toggle(ctx context.Context, id uuid.UUID, enabled bool) (*AllowlistEntry, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "allowlist id is required"}
	}
	action := "enable"
	if !enabled {
		action = "disable"
	}
	var out AllowlistEntry
	path := fmt.Sprintf("/v1/allowlist/%s/%s", id.String(), action)
	if err := s.client.transport.request(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
