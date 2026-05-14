package schemas

import "time"

type Pageview struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	SiteID    int64     `gorm:"column:site_id;index"`
	Path      string    `gorm:"column:path;index"`
	Referrer  string    `gorm:"column:referrer"`
	UserAgent string    `gorm:"column:user_agent"`
	Browser   string    `gorm:"column:browser"`
	OS        string    `gorm:"column:os"`
	Device    string    `gorm:"column:device"`
	Language  string    `gorm:"column:language"`
	Country   string    `gorm:"column:country"`
	VisitorID   string    `gorm:"column:visitor_id;index"`
	ScreenWidth int       `gorm:"column:screen_width"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (Pageview) TableName() string { return "pageviews" }
