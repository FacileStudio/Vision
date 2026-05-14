package schemas

import "time"

type Pageview struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SiteID    int64     `gorm:"column:site_id;index"`
	Path      string    `gorm:"column:path;index"`
	Referrer  string    `gorm:"column:referrer"`
	UserAgent string    `gorm:"column:user_agent"`
	Language  string    `gorm:"column:language"`
	Country   string    `gorm:"column:country"`
	VisitorID string    `gorm:"column:visitor_id;index"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (Pageview) TableName() string { return "pageviews" }
