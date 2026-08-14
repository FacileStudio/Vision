package sites

import "time"

// CreateRequest is the payload for registering a site.
type CreateRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	WorkspaceID int64  `json:"workspace_id"`
}

// UpdateRequest is the payload for editing a site.
type UpdateRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// SiteResponse is a site safe to return to a client.
type SiteResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	OwnerID     int64     `json:"owner_id"`
	WorkspaceID int64     `json:"workspace_id"`
	ShareToken  *string   `json:"share_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
