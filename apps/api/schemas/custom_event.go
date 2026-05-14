package schemas

import "time"

type CustomEvent struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SiteID    int64     `gorm:"column:site_id;index:idx_event_site_created,priority:1"`
	VisitorID string    `gorm:"column:visitor_id"`
	Path      string    `gorm:"column:path"`
	Name      string    `gorm:"column:name;index"`
	Props     string    `gorm:"column:props;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_event_site_created,priority:2"`
}

func (CustomEvent) TableName() string { return "custom_events" }
