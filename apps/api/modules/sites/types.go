package sites

import "time"

type CreateRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	WorkspaceID int64  `json:"workspace_id"`
}

type UpdateRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

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
