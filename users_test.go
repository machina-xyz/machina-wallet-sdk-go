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

func TestUsersList(t *testing.T) {
	t.Parallel()

	want := UserPage{Items: []User{{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Status:   UserStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}}}

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("status = %q", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(want)
	})

	got, err := c.Users().List(context.Background(), &ListUsersOptions{Status: UserStatusActive, Limit: 10})
	if err != nil {
		t.Fatalf("Users.List: %v", err)
	}
	if len(got.Items) != 1 {
		t.Errorf("len = %d", len(got.Items))
	}
}

func TestUsersGet(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/"+id.String() {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(User{ID: id, Status: UserStatusActive})
	})

	got, err := c.Users().Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Users.Get: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID = %v", got.ID)
	}

	if _, err := c.Users().Get(context.Background(), uuid.Nil); err == nil {
		t.Error("expected validation error for nil id")
	}
}

func TestUsersCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	var sawDeleteHard bool
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode: %v", err)
			}
			if req.PrimaryEmail == nil || *req.PrimaryEmail != "a@b.com" {
				t.Errorf("primary_email = %v", req.PrimaryEmail)
			}
			_ = json.NewEncoder(w).Encode(User{ID: id, PrimaryEmail: req.PrimaryEmail})
		case http.MethodPatch:
			var patch UpdateUserRequest
			_ = json.NewDecoder(r.Body).Decode(&patch)
			_ = json.NewEncoder(w).Encode(User{ID: id, Status: *patch.Status})
		case http.MethodDelete:
			if r.URL.Query().Get("hard") == "true" {
				sawDeleteHard = true
			}
			w.WriteHeader(http.StatusNoContent)
		}
	})

	email := "a@b.com"
	created, err := c.Users().Create(context.Background(), &CreateUserRequest{PrimaryEmail: &email})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != id {
		t.Errorf("ID = %v", created.ID)
	}

	susp := UserStatusSuspended
	updated, err := c.Users().Update(context.Background(), id, &UpdateUserRequest{Status: &susp})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Status != UserStatusSuspended {
		t.Errorf("status = %v", updated.Status)
	}

	if err := c.Users().Delete(context.Background(), id, &DeleteUserOptions{Hard: true}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !sawDeleteHard {
		t.Error("expected hard=true query")
	}

	if err := c.Users().Delete(context.Background(), uuid.Nil, nil); err == nil {
		t.Error("expected validation error")
	}
	if _, err := c.Users().Create(context.Background(), nil); err == nil {
		t.Error("expected validation error for nil request")
	}
	if _, err := c.Users().Update(context.Background(), uuid.Nil, &UpdateUserRequest{}); err == nil {
		t.Error("expected validation error for nil id")
	}
	if _, err := c.Users().Update(context.Background(), id, nil); err == nil {
		t.Error("expected validation error for nil patch")
	}
}

func TestUsersGetByLookups(t *testing.T) {
	t.Parallel()

	calls := map[string]string{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/lookup" {
			t.Errorf("path = %q", r.URL.Path)
		}
		calls[r.URL.Query().Get("kind")] = r.URL.Query().Get("value")
		_ = json.NewEncoder(w).Encode(User{ID: uuid.New()})
	})

	ctx := context.Background()
	cases := []struct {
		kind string
		call func() error
	}{
		{"email", func() error { _, err := c.Users().GetByEmail(ctx, "x@y.z"); return err }},
		{"phone", func() error { _, err := c.Users().GetByPhone(ctx, "+1555"); return err }},
		{"farcaster", func() error { _, err := c.Users().GetByFarcasterFid(ctx, 12345); return err }},
		{"discord", func() error { _, err := c.Users().GetByDiscordID(ctx, "d1"); return err }},
		{"twitter", func() error { _, err := c.Users().GetByTwitterID(ctx, "t1"); return err }},
		{"telegram", func() error { _, err := c.Users().GetByTelegramID(ctx, "tg1"); return err }},
		{"instagram", func() error { _, err := c.Users().GetByInstagramID(ctx, "ig1"); return err }},
		{"tiktok", func() error { _, err := c.Users().GetByTiktokID(ctx, "tt1"); return err }},
		{"github", func() error { _, err := c.Users().GetByGithubID(ctx, "gh1"); return err }},
		{"linkedin", func() error { _, err := c.Users().GetByLinkedinID(ctx, "li1"); return err }},
		{"apple_oauth", func() error { _, err := c.Users().GetByAppleSubject(ctx, "app-sub"); return err }},
		{"google_oauth", func() error { _, err := c.Users().GetByGoogleSubject(ctx, "g-sub"); return err }},
		{"spotify", func() error { _, err := c.Users().GetBySpotifyID(ctx, "sp1"); return err }},
		{"cross_app_account", func() error { _, err := c.Users().GetByCrossAppSubject(ctx, "xa1"); return err }},
	}
	for _, cc := range cases {
		if err := cc.call(); err != nil {
			t.Errorf("%s: %v", cc.kind, err)
		}
		if _, ok := calls[cc.kind]; !ok {
			t.Errorf("expected call for %s", cc.kind)
		}
	}

	if _, err := c.Users().GetByEmail(ctx, ""); err == nil {
		t.Error("expected validation error for empty email")
	}
}

func TestUsersGetByWalletAddress(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("kind") != "wallet" {
			t.Errorf("kind = %q", r.URL.Query().Get("kind"))
		}
		if r.URL.Query().Get("chain") != "ethereum" {
			t.Errorf("chain = %q", r.URL.Query().Get("chain"))
		}
		if r.URL.Query().Get("value") != "0xabc" {
			t.Errorf("value = %q", r.URL.Query().Get("value"))
		}
		_ = json.NewEncoder(w).Encode(User{ID: uuid.New()})
	})

	if _, err := c.Users().GetByWalletAddress(context.Background(), ChainEthereum, "0xabc"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := c.Users().GetByWalletAddress(context.Background(), "", "0xabc"); err == nil {
		t.Error("expected validation error for empty chain")
	}
	if _, err := c.Users().GetByWalletAddress(context.Background(), ChainEthereum, ""); err == nil {
		t.Error("expected validation error for empty address")
	}
}

func TestUsersLinkUnlinkIdentifier(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	iid := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if !strings.HasSuffix(r.URL.Path, "/identifiers") {
				t.Errorf("path = %q", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(User{ID: uid})
		case http.MethodDelete:
			if !strings.Contains(r.URL.Path, iid.String()) {
				t.Errorf("path missing identifier id: %q", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(User{ID: uid})
		}
	})

	if _, err := c.Users().LinkIdentifier(context.Background(), uid, LinkIdentifierRequest{Kind: IdentifierKindGithub, Value: "u"}); err != nil {
		t.Fatalf("Link: %v", err)
	}
	if _, err := c.Users().UnlinkIdentifier(context.Background(), uid, iid); err != nil {
		t.Fatalf("Unlink: %v", err)
	}

	if _, err := c.Users().LinkIdentifier(context.Background(), uuid.Nil, LinkIdentifierRequest{Kind: IdentifierKindGithub, Value: "u"}); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Users().LinkIdentifier(context.Background(), uid, LinkIdentifierRequest{}); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Users().UnlinkIdentifier(context.Background(), uuid.Nil, iid); err == nil {
		t.Error("expected validation")
	}
}

func TestUsersCustomMetadata(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(customMetadataEnvelope{CustomMetadata: map[string]any{"k": "v"}})
		case http.MethodPut:
			var env customMetadataEnvelope
			_ = json.NewDecoder(r.Body).Decode(&env)
			_ = json.NewEncoder(w).Encode(env)
		case http.MethodPatch:
			var env customMetadataEnvelope
			_ = json.NewDecoder(r.Body).Decode(&env)
			env.CustomMetadata["merged"] = true
			_ = json.NewEncoder(w).Encode(env)
		}
	})

	got, err := c.Users().GetCustomMetadata(context.Background(), uid)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got["k"] != "v" {
		t.Errorf("k = %v", got["k"])
	}

	replaced, err := c.Users().ReplaceCustomMetadata(context.Background(), uid, map[string]any{"new": "value"})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if replaced["new"] != "value" {
		t.Errorf("new = %v", replaced["new"])
	}

	merged, err := c.Users().MergeCustomMetadata(context.Background(), uid, map[string]any{"patch": "yes"})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if merged["merged"] != true {
		t.Errorf("merged = %v", merged["merged"])
	}

	// Validation paths.
	if _, err := c.Users().GetCustomMetadata(context.Background(), uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Users().ReplaceCustomMetadata(context.Background(), uuid.Nil, nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Users().MergeCustomMetadata(context.Background(), uuid.Nil, nil); err == nil {
		t.Error("expected validation")
	}
}
