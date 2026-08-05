package schemas

import "time"

type Pageview struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	SiteID       int64     `gorm:"column:site_id;index:idx_site_created,priority:1"`
	Path         string    `gorm:"column:path;index"`
	Referrer     string    `gorm:"column:referrer"`
	UserAgent    string    `gorm:"column:user_agent"`
	Browser      string    `gorm:"column:browser"`
	OS           string    `gorm:"column:os"`
	Device       string    `gorm:"column:device"`
	Language     string    `gorm:"column:language"`
	Country      string    `gorm:"column:country"`
	VisitorID    string    `gorm:"column:visitor_id;index"`
	ScreenWidth  int       `gorm:"column:screen_width"`
	UTMSource    string    `gorm:"column:utm_source"`
	UTMMedium    string    `gorm:"column:utm_medium"`
	UTMCampaign  string    `gorm:"column:utm_campaign"`
	UTMTerm      string    `gorm:"column:utm_term"`
	UTMContent   string    `gorm:"column:utm_content"`
	PerfDNS      *int      `gorm:"column:perf_dns"`
	PerfTCP      *int      `gorm:"column:perf_tcp"`
	PerfTTFB     *int      `gorm:"column:perf_ttfb"`
	PerfDOMLoad  *int      `gorm:"column:perf_dom_load"`
	PerfPageLoad *int      `gorm:"column:perf_page_load"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index;index:idx_site_created,priority:2"`
}

func (Pageview) TableName() string { return "pageviews" }
