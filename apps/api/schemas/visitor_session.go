package schemas

import "time"

// VisitorSession aggregates a series of pageviews from one visitor.
type VisitorSession struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	SiteID        int64     `gorm:"column:site_id;index"`
	VisitorID     string    `gorm:"column:visitor_id"`
	StartedAt     time.Time `gorm:"column:started_at"`
	EndedAt       time.Time `gorm:"column:ended_at"`
	EntryPath     string    `gorm:"column:entry_path"`
	ExitPath      string    `gorm:"column:exit_path"`
	PageviewCount int       `gorm:"column:pageview_count;default:1"`
	Duration      int       `gorm:"column:duration;default:0"`
	IsBounce      bool      `gorm:"column:is_bounce;default:true"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (VisitorSession) TableName() string { return "visitor_sessions" }
