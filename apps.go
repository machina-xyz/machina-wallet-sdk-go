package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// AppEnvironment is the deployment environment of an app.
type AppEnvironment string

// Supported environments.
const (
	AppEnvProduction AppEnvironment = "production"
	AppEnvStaging    AppEnvironment = "staging"
	AppEnvDev        AppEnvironment = "development"
)

// App is the wire shape returned by the apps API.
type App struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description,omitempty"`
	Environment AppEnvironment `json:"environment"`
	PublicKey   string         `json:"public_key,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AppPage is a paginated list of apps.
type AppPage struct {
	Items      []App   `json:"items"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListAppsOptions is the query parameter set for List.
type ListAppsOptions struct {
	Environment AppEnvironment
	Limit       int
	Cursor      string
}

// CreateAppRequest is the input shape for Create.
type CreateAppRequest struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug,omitempty"`
	Description string         `json:"description,omitempty"`
	Environment AppEnvironment `json:"environment,omitempty"`
}

// UpdateAppRequest is the input shape for Update.
type UpdateAppRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Environment *AppEnvironment `json:"environment,omitempty"`
}

// RotatedAppSecret carries the freshly rotated secret material. The plaintext
// secret is only returned once and must be captured by the caller.
type RotatedAppSecret struct {
	AppID         uuid.UUID `json:"app_id"`
	NewSecret     string    `json:"new_secret"`
	SecretPreview string    `json:"secret_preview,omitempty"`
	RotatedAt     time.Time `json:"rotated_at"`
}

// AppsService is the typed wrapper for /v1/apps.
type AppsService struct{ client *Client }

// Apps returns the typed AppsService for this client.
func (c *Client) Apps() *AppsService { return &AppsService{client: c} }

// List returns apps for the calling tenant.
func (s *AppsService) List(ctx context.Context, opts *ListAppsOptions) (*AppPage, error) {
	q := url.Values{}
	if opts != nil {
		setIfNonEmpty(q, "environment", string(opts.Environment))
		setIfNonZero(q, "limit", opts.Limit)
		setIfNonEmpty(q, "cursor", opts.Cursor)
	}
	var page AppPage
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/apps", q, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get fetches a single app by id.
func (s *AppsService) Get(ctx context.Context, id uuid.UUID) (*App, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "app id is required"}
	}
	var a App
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/apps/"+id.String(), nil, nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Create provisions a new app for the calling tenant.
func (s *AppsService) Create(ctx context.Context, req *CreateAppRequest) (*App, error) {
	if req == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "CreateAppRequest is required"}
	}
	if req.Name == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "Name is required"}
	}
	var a App
	if err := s.client.transport.request(ctx, http.MethodPost, "/v1/apps", nil, req, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Update applies a partial update to an app.
func (s *AppsService) Update(ctx context.Context, id uuid.UUID, patch *UpdateAppRequest) (*App, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "app id is required"}
	}
	if patch == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "UpdateAppRequest is required"}
	}
	var a App
	if err := s.client.transport.request(ctx, http.MethodPatch, "/v1/apps/"+id.String(), nil, patch, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Delete removes an app and revokes all secrets associated with it.
func (s *AppsService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return &Error{Code: ErrCodeValidation, Message: "app id is required"}
	}
	return s.client.transport.request(ctx, http.MethodDelete, "/v1/apps/"+id.String(), nil, nil, nil)
}

// RotateSecret issues a new signing secret for the app and revokes the
// previous one after a brief overlap window. The plaintext secret is only
// returned in the response and never logged on the server.
func (s *AppsService) RotateSecret(ctx context.Context, id uuid.UUID) (*RotatedAppSecret, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "app id is required"}
	}
	var out RotatedAppSecret
	path := fmt.Sprintf("/v1/apps/%s/rotate_secret", id.String())
	if err := s.client.transport.request(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
