package schemas

import "time"

type Workspace struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Workspace) TableName() string { return "workspaces" }

type WorkspaceMember struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	WorkspaceID int64     `gorm:"column:workspace_id;index;uniqueIndex:idx_ws_user"`
	UserID      int64     `gorm:"column:user_id;index;uniqueIndex:idx_ws_user"`
	Role        string    `gorm:"column:role"`
	User        User      `gorm:"foreignKey:UserID"`
	Workspace   Workspace `gorm:"foreignKey:WorkspaceID"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (WorkspaceMember) TableName() string { return "workspace_members" }
