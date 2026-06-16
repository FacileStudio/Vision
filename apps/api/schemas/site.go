package schemas

import "time"

type Site struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Domain      string    `gorm:"column:domain;uniqueIndex"`
	OwnerID     int64     `gorm:"column:owner_id;index"`
	WorkspaceID int64     `gorm:"column:workspace_id;index"`
	ShareToken  *string   `gorm:"column:share_token;uniqueIndex"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Site) TableName() string { return "sites" }
