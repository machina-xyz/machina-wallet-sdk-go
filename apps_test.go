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

func TestAppsCRUDAndRotate(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/apps":
			if r.URL.Query().Get("environment") != "production" {
				t.Errorf("env = %q", r.URL.Query().Get("environment"))
			}
			_ = json.NewEncoder(w).Encode(AppPage{Items: []App{{ID: id, Name: "demo"}}})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(App{ID: id})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/rotate_secret"):
			_ = json.NewEncoder(w).Encode(RotatedAppSecret{AppID: id, NewSecret: "shh", RotatedAt: time.Now().UTC()})
		case r.Method == http.MethodPost:
			var req CreateAppRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			_ = json.NewEncoder(w).Encode(App{ID: id, Name: req.Name})
		case r.Method == http.MethodPatch:
			var patch UpdateAppRequest
			_ = json.NewDecoder(r.Body).Decode(&patch)
			a := App{ID: id}
			if patch.Name != nil {
				a.Name = *patch.Name
			}
			_ = json.NewEncoder(w).Encode(a)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	page, err := c.Apps().List(context.Background(), &ListAppsOptions{Environment: AppEnvProduction, Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("len = %d", len(page.Items))
	}

	got, err := c.Apps().Get(context.Background(), id)
	if err != nil || got.ID != id {
		t.Errorf("Get failed: %v", err)
	}

	created, err := c.Apps().Create(context.Background(), &CreateAppRequest{Name: "demo"})
	if err != nil || created.Name != "demo" {
		t.Errorf("Create failed: %v", err)
	}

	name := "renamed"
	updated, err := c.Apps().Update(context.Background(), id, &UpdateAppRequest{Name: &name})
	if err != nil || updated.Name != "renamed" {
		t.Errorf("Update failed: %v", err)
	}

	if err := c.Apps().Delete(context.Background(), id); err != nil {
		t.Errorf("Delete: %v", err)
	}

	rot, err := c.Apps().RotateSecret(context.Background(), id)
	if err != nil || rot.NewSecret != "shh" {
		t.Errorf("RotateSecret failed: %v", err)
	}
}

func TestAppsValidation(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	})

	ctx := context.Background()
	if _, err := c.Apps().Get(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Apps().Create(ctx, nil); err == nil {
		t.Error("expected validation for nil request")
	}
	if _, err := c.Apps().Create(ctx, &CreateAppRequest{}); err == nil {
		t.Error("expected validation for empty name")
	}
	if _, err := c.Apps().Update(ctx, uuid.Nil, &UpdateAppRequest{}); err == nil {
		t.Error("expected validation for nil id")
	}
	if _, err := c.Apps().Update(ctx, uuid.New(), nil); err == nil {
		t.Error("expected validation for nil patch")
	}
	if err := c.Apps().Delete(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
	if _, err := c.Apps().RotateSecret(ctx, uuid.Nil); err == nil {
		t.Error("expected validation")
	}
}
