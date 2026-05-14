package analytics

type OverviewResponse struct {
	TotalPageviews  int64          `json:"total_pageviews"`
	UniqueVisitors  int64          `json:"unique_visitors"`
	TopPages        []PageStat     `json:"top_pages"`
	TopReferrers    []ReferrerStat `json:"top_referrers"`
	TopCountries    []CountryStat  `json:"top_countries"`
	TopBrowsers     []BrowserStat  `json:"top_browsers"`
	TopOS           []OSStat       `json:"top_os"`
	TopDevices      []DeviceStat   `json:"top_devices"`
	PageviewsPerDay []DayStat      `json:"pageviews_per_day"`
	TopScreens      []ScreenStat   `json:"top_screens"`
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

type BrowserStat struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

type OSStat struct {
	OS    string `json:"os"`
	Count int64  `json:"count"`
}

type DeviceStat struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}

type DayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ScreenStat struct {
	Screen string `json:"screen"`
	Count  int64  `json:"count"`
}
