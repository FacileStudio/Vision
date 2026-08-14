package apikeys

import "time"

// CreateRequest is the payload for creating an API key.
type CreateRequest struct {
	Name        string `json:"name"`
	Scopes      string `json:"scopes"`
	SiteID      *int64 `json:"site_id"`
	WorkspaceID *int64 `json:"workspace_id"`
}

// CreateResponse returns the newly minted key alongside its record.
type CreateResponse struct {
	Key string         `json:"key"`
	API APIKeyResponse `json:"api_key"`
}

// APIKeyResponse is a key stripped of its secret, safe to return to a client.
type APIKeyResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"`
	KeyHint     string     `json:"key_hint"`
	Scopes      string     `json:"scopes"`
	SiteID      *int64     `json:"site_id"`
	WorkspaceID *int64     `json:"workspace_id"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
