package schemas

import "time"

// Goal is a conversion target defined on a site.
type Goal struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SiteID    int64     `gorm:"column:site_id;index"`
	OwnerID   int64     `gorm:"column:owner_id;index"`
	Name      string    `gorm:"column:name"`
	GoalType  string    `gorm:"column:goal_type"`
	EventName *string   `gorm:"column:event_name"`
	PagePath  *string   `gorm:"column:page_path"`
	MatchType string    `gorm:"column:match_type;default:exact"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Goal) TableName() string { return "goals" }
