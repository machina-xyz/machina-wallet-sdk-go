package machinawallet

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// UserStatus is the lifecycle state of a user record.
type UserStatus string

// User status values.
const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// IdentifierKind discriminates a linked identifier on a user.
type IdentifierKind string

// Supported linked identifier kinds. These mirror the Privy parity set.
const (
	IdentifierKindEmail           IdentifierKind = "email"
	IdentifierKindPhone           IdentifierKind = "phone"
	IdentifierKindWallet          IdentifierKind = "wallet"
	IdentifierKindFarcaster       IdentifierKind = "farcaster"
	IdentifierKindDiscord         IdentifierKind = "discord"
	IdentifierKindTwitter         IdentifierKind = "twitter"
	IdentifierKindTelegram        IdentifierKind = "telegram"
	IdentifierKindInstagram       IdentifierKind = "instagram"
	IdentifierKindTiktok          IdentifierKind = "tiktok"
	IdentifierKindGithub          IdentifierKind = "github"
	IdentifierKindLinkedin        IdentifierKind = "linkedin"
	IdentifierKindAppleOAuth      IdentifierKind = "apple_oauth"
	IdentifierKindGoogleOAuth     IdentifierKind = "google_oauth"
	IdentifierKindSpotify         IdentifierKind = "spotify"
	IdentifierKindCrossAppAccount IdentifierKind = "cross_app_account"
)

// LinkedIdentifier is a single identifier linked to a user.
type LinkedIdentifier struct {
	ID         uuid.UUID      `json:"id"`
	Kind       IdentifierKind `json:"kind"`
	Value      string         `json:"value"`
	Chain      *Chain         `json:"chain,omitempty"`
	VerifiedAt *time.Time     `json:"verified_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// User is the wire shape returned by the user management API.
type User struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	Status            UserStatus             `json:"status"`
	PrimaryEmail      *string                `json:"primary_email,omitempty"`
	PrimaryPhone      *string                `json:"primary_phone,omitempty"`
	LinkedIdentifiers []LinkedIdentifier     `json:"linked_identifiers,omitempty"`
	CustomMetadata    map[string]interface{} `json:"custom_metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	DeletedAt         *time.Time             `json:"deleted_at,omitempty"`
}

// UserPage is a paginated list of users.
type UserPage struct {
	Items      []User  `json:"items"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListUsersOptions is the query parameter set for ListUsers.
type ListUsersOptions struct {
	Status UserStatus
	Limit  int
	Cursor string
}

// CreateUserRequest is the input shape for Create.
type CreateUserRequest struct {
	PrimaryEmail      *string                `json:"primary_email,omitempty"`
	PrimaryPhone      *string                `json:"primary_phone,omitempty"`
	LinkedIdentifiers []LinkIdentifierInput  `json:"linked_identifiers,omitempty"`
	CustomMetadata    map[string]interface{} `json:"custom_metadata,omitempty"`
}

// LinkIdentifierInput is a single identifier supplied at user creation time
// or via the LinkIdentifier endpoint.
type LinkIdentifierInput struct {
	Kind  IdentifierKind `json:"kind"`
	Value string         `json:"value"`
	Chain *Chain         `json:"chain,omitempty"`
}

// LinkIdentifierRequest mirrors LinkIdentifierInput for the link endpoint.
type LinkIdentifierRequest = LinkIdentifierInput

// UpdateUserRequest is the input shape for Update. Only non-nil fields are
// transmitted.
type UpdateUserRequest struct {
	PrimaryEmail   *string                `json:"primary_email,omitempty"`
	PrimaryPhone   *string                `json:"primary_phone,omitempty"`
	Status         *UserStatus            `json:"status,omitempty"`
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"`
}

// DeleteUserOptions controls deletion semantics. When Hard is true the record
// and all linked artifacts are purged immediately; otherwise a soft delete is
// performed and the user can be restored.
type DeleteUserOptions struct {
	Hard bool
}

// UsersService is the typed wrapper for the /v1/users API surface.
//
// Obtain an instance via Client.Users(). A UsersService is safe for concurrent
// use by multiple goroutines.
type UsersService struct{ client *Client }

// Users returns the typed UsersService for this client.
func (c *Client) Users() *UsersService { return &UsersService{client: c} }

// List returns users for the calling tenant.
func (s *UsersService) List(ctx context.Context, opts *ListUsersOptions) (*UserPage, error) {
	q := url.Values{}
	if opts != nil {
		setIfNonEmpty(q, "status", string(opts.Status))
		setIfNonZero(q, "limit", opts.Limit)
		setIfNonEmpty(q, "cursor", opts.Cursor)
	}
	var page UserPage
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/users", q, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Get fetches a single user by id.
func (s *UsersService) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	var u User
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/users/"+id.String(), nil, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Create provisions a new user.
func (s *UsersService) Create(ctx context.Context, req *CreateUserRequest) (*User, error) {
	if req == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "CreateUserRequest is required"}
	}
	var u User
	if err := s.client.transport.request(ctx, http.MethodPost, "/v1/users", nil, req, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Update applies a partial update to a user.
func (s *UsersService) Update(ctx context.Context, id uuid.UUID, patch *UpdateUserRequest) (*User, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	if patch == nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "UpdateUserRequest is required"}
	}
	var u User
	if err := s.client.transport.request(ctx, http.MethodPatch, "/v1/users/"+id.String(), nil, patch, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// Delete removes a user. Soft by default, hard when opts.Hard is true.
func (s *UsersService) Delete(ctx context.Context, id uuid.UUID, opts *DeleteUserOptions) error {
	if id == uuid.Nil {
		return &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	q := url.Values{}
	if opts != nil && opts.Hard {
		q.Set("hard", "true")
	}
	return s.client.transport.request(ctx, http.MethodDelete, "/v1/users/"+id.String(), q, nil, nil)
}

// getByLookup is a shared helper for the typed getBy* endpoints.
func (s *UsersService) getByLookup(ctx context.Context, kind, value string) (*User, error) {
	if value == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: kind + " is required"}
	}
	q := url.Values{"kind": []string{kind}, "value": []string{value}}
	var u User
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/users/lookup", q, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByEmail looks up a user by primary or linked email address.
func (s *UsersService) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.getByLookup(ctx, "email", email)
}

// GetByPhone looks up a user by primary or linked phone number (E.164).
func (s *UsersService) GetByPhone(ctx context.Context, phone string) (*User, error) {
	return s.getByLookup(ctx, "phone", phone)
}

// GetByWalletAddress looks up a user by a linked wallet address on chain.
func (s *UsersService) GetByWalletAddress(ctx context.Context, chain Chain, address string) (*User, error) {
	if chain == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "chain is required"}
	}
	if address == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "address is required"}
	}
	q := url.Values{
		"kind":  []string{"wallet"},
		"chain": []string{string(chain)},
		"value": []string{address},
	}
	var u User
	if err := s.client.transport.request(ctx, http.MethodGet, "/v1/users/lookup", q, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByFarcasterFid looks up a user by Farcaster FID.
func (s *UsersService) GetByFarcasterFid(ctx context.Context, fid uint64) (*User, error) {
	return s.getByLookup(ctx, "farcaster", strconv.FormatUint(fid, 10))
}

// GetByDiscordID looks up a user by Discord user id.
func (s *UsersService) GetByDiscordID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "discord", id)
}

// GetByTwitterID looks up a user by Twitter (X) user id.
func (s *UsersService) GetByTwitterID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "twitter", id)
}

// GetByTelegramID looks up a user by Telegram user id.
func (s *UsersService) GetByTelegramID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "telegram", id)
}

// GetByInstagramID looks up a user by Instagram user id.
func (s *UsersService) GetByInstagramID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "instagram", id)
}

// GetByTiktokID looks up a user by TikTok user id.
func (s *UsersService) GetByTiktokID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "tiktok", id)
}

// GetByGithubID looks up a user by GitHub user id.
func (s *UsersService) GetByGithubID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "github", id)
}

// GetByLinkedinID looks up a user by LinkedIn user id.
func (s *UsersService) GetByLinkedinID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "linkedin", id)
}

// GetByAppleSubject looks up a user by Sign in with Apple subject claim.
func (s *UsersService) GetByAppleSubject(ctx context.Context, sub string) (*User, error) {
	return s.getByLookup(ctx, "apple_oauth", sub)
}

// GetByGoogleSubject looks up a user by Google OAuth subject claim.
func (s *UsersService) GetByGoogleSubject(ctx context.Context, sub string) (*User, error) {
	return s.getByLookup(ctx, "google_oauth", sub)
}

// GetBySpotifyID looks up a user by Spotify user id.
func (s *UsersService) GetBySpotifyID(ctx context.Context, id string) (*User, error) {
	return s.getByLookup(ctx, "spotify", id)
}

// GetByCrossAppSubject looks up a user by cross-app subject identifier.
func (s *UsersService) GetByCrossAppSubject(ctx context.Context, sub string) (*User, error) {
	return s.getByLookup(ctx, "cross_app_account", sub)
}

// LinkIdentifier attaches a new identifier to a user.
func (s *UsersService) LinkIdentifier(ctx context.Context, id uuid.UUID, identifier LinkIdentifierRequest) (*User, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	if identifier.Kind == "" || identifier.Value == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "identifier kind and value are required"}
	}
	var u User
	path := fmt.Sprintf("/v1/users/%s/identifiers", id.String())
	if err := s.client.transport.request(ctx, http.MethodPost, path, nil, identifier, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// UnlinkIdentifier removes a linked identifier from a user.
func (s *UsersService) UnlinkIdentifier(ctx context.Context, id uuid.UUID, identifierID uuid.UUID) (*User, error) {
	if id == uuid.Nil || identifierID == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id and identifier id are required"}
	}
	var u User
	path := fmt.Sprintf("/v1/users/%s/identifiers/%s", id.String(), identifierID.String())
	if err := s.client.transport.request(ctx, http.MethodDelete, path, nil, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// customMetadataEnvelope wraps map sends/receives so the API can extend the
// shape without breaking the SDK.
type customMetadataEnvelope struct {
	CustomMetadata map[string]interface{} `json:"custom_metadata"`
}

// GetCustomMetadata returns the custom metadata map for a user.
func (s *UsersService) GetCustomMetadata(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	var env customMetadataEnvelope
	path := fmt.Sprintf("/v1/users/%s/custom_metadata", id.String())
	if err := s.client.transport.request(ctx, http.MethodGet, path, nil, nil, &env); err != nil {
		return nil, err
	}
	if env.CustomMetadata == nil {
		env.CustomMetadata = map[string]interface{}{}
	}
	return env.CustomMetadata, nil
}

// ReplaceCustomMetadata replaces the entire custom metadata map for a user.
func (s *UsersService) ReplaceCustomMetadata(ctx context.Context, id uuid.UUID, metadata map[string]interface{}) (map[string]interface{}, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	var env customMetadataEnvelope
	path := fmt.Sprintf("/v1/users/%s/custom_metadata", id.String())
	if err := s.client.transport.request(ctx, http.MethodPut, path, nil, customMetadataEnvelope{CustomMetadata: metadata}, &env); err != nil {
		return nil, err
	}
	if env.CustomMetadata == nil {
		env.CustomMetadata = map[string]interface{}{}
	}
	return env.CustomMetadata, nil
}

// MergeCustomMetadata merges patch into the existing custom metadata using
// shallow JSON merge semantics: keys with null values are deleted, keys with
// non-null values overwrite.
func (s *UsersService) MergeCustomMetadata(ctx context.Context, id uuid.UUID, patch map[string]interface{}) (map[string]interface{}, error) {
	if id == uuid.Nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "user id is required"}
	}
	if patch == nil {
		patch = map[string]interface{}{}
	}
	var env customMetadataEnvelope
	path := fmt.Sprintf("/v1/users/%s/custom_metadata", id.String())
	if err := s.client.transport.request(ctx, http.MethodPatch, path, nil, customMetadataEnvelope{CustomMetadata: patch}, &env); err != nil {
		return nil, err
	}
	if env.CustomMetadata == nil {
		env.CustomMetadata = map[string]interface{}{}
	}
	return env.CustomMetadata, nil
}
