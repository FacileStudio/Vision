package workspaces

import "time"

// CreateRequest is the payload for creating a workspace.
type CreateRequest struct {
	Name string `json:"name"`
}

// UpdateRequest is the payload for renaming a workspace.
type UpdateRequest struct {
	Name string `json:"name"`
}

// AddMemberRequest adds a member by email with a role.
type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UpdateMemberRequest changes a member's role.
type UpdateMemberRequest struct {
	Role string `json:"role"`
}

// WorkspaceResponse is a workspace safe to return to a client.
type WorkspaceResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	MemberCount int64     `json:"member_count"`
	SiteCount   int64     `json:"site_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MemberResponse is a workspace member safe to return to a client.
type MemberResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
