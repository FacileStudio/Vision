package analytics

type OverviewResponse struct {
	TotalPageviews    int64          `json:"total_pageviews"`
	UniqueVisitors    int64          `json:"unique_visitors"`
	TopPages          []PageStat     `json:"top_pages"`
	TopReferrers      []ReferrerStat `json:"top_referrers"`
	TopCountries      []CountryStat  `json:"top_countries"`
	PageviewsPerDay   []DayStat      `json:"pageviews_per_day"`
}

type PageStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

type DayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
