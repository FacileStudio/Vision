package schemas

import "time"

// Webhook delivers periodic analytics reports to a URL.
type Webhook struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	OwnerID       int64      `gorm:"column:owner_id;index"`
	WorkspaceID   *int64     `gorm:"column:workspace_id;index"`
	URL           string     `gorm:"column:url"`
	Secret        string     `gorm:"column:secret"`
	Period        string     `gorm:"column:period"`
	IntervalHours int        `gorm:"column:interval_hours;default:24"`
	Enabled       bool       `gorm:"column:enabled;default:true"`
	LastSentAt    *time.Time `gorm:"column:last_sent_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Webhook) TableName() string { return "webhooks" }
