package schemas

import "time"

type APIKey struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	UserID      int64      `gorm:"column:user_id;index"`
	SiteID      *int64     `gorm:"column:site_id;index"`
	WorkspaceID *int64     `gorm:"column:workspace_id;index"`
	Name        string     `gorm:"column:name"`
	Prefix      string     `gorm:"column:prefix"`
	KeyHint     string     `gorm:"column:key_hint"`
	KeyHash     string     `gorm:"column:key_hash;uniqueIndex"`
	Scopes      string     `gorm:"column:scopes;default:read"`
	IsActive    bool       `gorm:"column:is_active;default:true"`
	LastUsedAt  *time.Time `gorm:"column:last_used_at"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (APIKey) TableName() string { return "api_keys" }
