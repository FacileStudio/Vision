package schemas

import "time"

// BotHit counts the requests one bot made to one path of a site on one UTC day.
type BotHit struct {
	ID       int64     `gorm:"column:id;primaryKey"`
	SiteID   int64     `gorm:"column:site_id"`
	Day      time.Time `gorm:"column:day"`
	Category string    `gorm:"column:category"`
	Name     string    `gorm:"column:name"`
	Path     string    `gorm:"column:path"`
	Hits     int64     `gorm:"column:hits"`
}

// TableName pins the table name.
func (BotHit) TableName() string { return "bot_hits" }
