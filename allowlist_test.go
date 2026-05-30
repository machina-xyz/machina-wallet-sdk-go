package machinawallet

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAllowlistCRUD(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/allowlist":
			// Two list calls hit this branch: the first asserts the full
			// happy-path query, the second only exercises the disabled
			// filter. Distinguish by the kind parameter.
			if r.URL.Query().Get("kind") == "address" {
				if r.URL.Query().Get("enabled") != "true" {
					t.Errorf("enabled = %q", r.URL.Query().Get("enabled"))
				}
			} else {
				if r.URL.Query().Get("enabled") != "false" {
					t.Errorf("expected enabled=false, got %q", r.URL.Query().Get("enabled"))
				}
			}
			_ = json.NewEncoder(w).Encode(AllowlistPage{Items: []AllowlistEntry{{ID: id, Kind: AllowlistKindAddress}}})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(AllowlistEntry{ID: id, Kind: AllowlistKindAddress})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/allowlist":
			_ = json.NewEncoder(w).Encode(AllowlistEntry{ID: id, Kind: AllowlistKindAddress, Value: "0xabc"})
		case r.Method == http.MethodPatch:
			var patch UpdateAllowlistEntryRequest
			_ = json.NewDecoder(r.Body).Decode(&patch)
			_ = json.NewEncoder(w).Encode(AllowlistEntry{ID: id, Label: *patch.Label})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/enable"):
			_ = json.NewEncoder(w).Encode(AllowlistEntry{ID: id, Enabled: true})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/disable"):
			_ = json.NewEncoder(w).Encode(AllowlistEntry{ID: id, Enabled: false})
		}
	})

	enabled := true
	page, err := c.Allowlist().List(context.Background(), &ListAllowlistOptions{Kind: AllowlistKindAddress, Enabled: &enabled, Limit: 25})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("len = %d", len(page.Items))
	}

	got, err := c.Allowlist().Get(context.Background(), id)
	if err != nil || got.ID != id {
		t.Errorf("Get failed: %v", err)
	}

	created, err := c.Allowlist().Create(context.Background(), &CreateAllowlistEntryRequest{Kind: AllowlistKindAddress, Value: "0xabc"})
	if err != nil || created.Value != "0xabc" {
		t.Errorf("Create failed: %v", err)
	}

	label := "tagged"
	updated, err := c.Allowlist().Update(context.Background(), id, &UpdateAllowlistEntryRequest{Label: &label})
	if err != nil || updated.Label != "tagged" {
		t.Errorf("Update failed: %v", err)
	}

	if err := c.Allowlist().Delete(context.Background(), id); err != nil {
		t.Errorf("Delete: %v", err)
	}

	en, err := c.Allowlist().Enable(context.Background(), id)
	if err != nil || !en.Enabled {
		t.Errorf("Enable failed: %v", err)
	}
	dis, err := c.Allowlist().Disable(context.Background(), id)
	if err != nil || dis.Enabled {
		t.Errorf("Disable failed: %v", err)
	}

	// disabled list filter
	disabled := false
	if _, err := c.Allowlist().List(context.Background(), &ListAllowlistOptions{Enabled: &disabled}); err != nil {
		t.Errorf("List disabled: %v", err)
	}
}

func TestAllowlistValidation(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	ctx := context.Background()
	if _, err := c.Allowlist().Get(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Allowlist().Create(ctx, nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Allowlist().Create(ctx, &CreateAllowlistEntryRequest{}); err == nil {
		t.Error("expected validation for missing fields")
	}
	if _, err := c.Allowlist().Update(ctx, uuid.Nil, nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Allowlist().Update(ctx, uuid.New(), nil); err == nil {
		t.Error("expected validation for nil patch")
	}
	if err := c.Allowlist().Delete(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Allowlist().Enable(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Allowlist().Disable(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
}
