package workspaces

import "time"

type CreateRequest struct {
	Name string `json:"name"`
}

type UpdateRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRequest struct {
	Role string `json:"role"`
}

type WorkspaceResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	MemberCount int64     `json:"member_count"`
	SiteCount   int64     `json:"site_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MemberResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
