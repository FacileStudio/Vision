package analytics

// Filters narrows an analytics query by dimension.
type Filters struct {
	Country  string
	Browser  string
	OS       string
	Device   string
	Path     string
	Referrer string
}

// PerformanceStats holds average Web performance timings.
type PerformanceStats struct {
	AvgDNS      float64 `json:"avg_dns"`
	AvgTCP      float64 `json:"avg_tcp"`
	AvgTTFB     float64 `json:"avg_ttfb"`
	AvgDOMLoad  float64 `json:"avg_dom_load"`
	AvgPageLoad float64 `json:"avg_page_load"`
	SampleCount int64   `json:"sample_count"`
}

// EventStat counts a single custom event name.
type EventStat struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// OverviewResponse is the full dashboard payload for a site and period.
type OverviewResponse struct {
	TotalPageviews           int64             `json:"total_pageviews"`
	UniqueVisitors           int64             `json:"unique_visitors"`
	TopPages                 []PageStat        `json:"top_pages"`
	TopReferrers             []ReferrerStat    `json:"top_referrers"`
	TopCountries             []CountryStat     `json:"top_countries"`
	TopBrowsers              []BrowserStat     `json:"top_browsers"`
	TopOS                    []OSStat          `json:"top_os"`
	TopDevices               []DeviceStat      `json:"top_devices"`
	PageviewsPerDay          []DayStat         `json:"pageviews_per_day"`
	TopScreens               []ScreenStat      `json:"top_screens"`
	UniqueVisitorsPerDay     []DayStat         `json:"unique_visitors_per_day"`
	HourlyDistribution       []HourStat        `json:"hourly_distribution"`
	PrevTotalPageviews       int64             `json:"prev_total_pageviews"`
	PrevUniqueVisitors       int64             `json:"prev_unique_visitors"`
	BounceRate               float64           `json:"bounce_rate"`
	AvgSessionDuration       float64           `json:"avg_session_duration"`
	PagesPerSession          float64           `json:"pages_per_session"`
	PrevBounceRate           float64           `json:"prev_bounce_rate"`
	PrevAvgSessionDuration   float64           `json:"prev_avg_session_duration"`
	PrevPagesPerSession      float64           `json:"prev_pages_per_session"`
	TopEntryPages            []PageStat        `json:"top_entry_pages"`
	TopExitPages             []PageStat        `json:"top_exit_pages"`
	TopUTMSources            []UTMStat         `json:"top_utm_sources"`
	TopUTMMediums            []UTMStat         `json:"top_utm_mediums"`
	TopUTMCampaigns          []UTMStat         `json:"top_utm_campaigns"`
	PrevPageviewsPerDay      []DayStat         `json:"prev_pageviews_per_day"`
	PrevUniqueVisitorsPerDay []DayStat         `json:"prev_unique_visitors_per_day"`
	Performance              *PerformanceStats `json:"performance"`
	TopEvents                []EventStat       `json:"top_events"`
}

// PageStat counts pageviews for one path.
type PageStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

// ReferrerStat counts pageviews from one referrer.
type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

// CountryStat counts pageviews from one country.
type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// BrowserStat counts pageviews from one browser.
type BrowserStat struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

// OSStat counts pageviews from one operating system.
type OSStat struct {
	OS    string `json:"os"`
	Count int64  `json:"count"`
}

// DeviceStat counts pageviews by device type.
type DeviceStat struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}

// DayStat counts an event per day.
type DayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ScreenStat counts pageviews by screen-width class.
type ScreenStat struct {
	Screen string `json:"screen"`
	Count  int64  `json:"count"`
}

// HourStat counts an event per hour of day.
type HourStat struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// UTMStat counts pageviews by one UTM value.
type UTMStat struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}
