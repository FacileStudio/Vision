package events

// PerformanceData carries cumulative Web performance timings.
type PerformanceData struct {
	DNS      int `json:"dns"`
	TCP      int `json:"tcp"`
	TTFB     int `json:"ttfb"`
	DOMLoad  int `json:"dom_load"`
	PageLoad int `json:"page_load"`
}

// PageviewRequest is the tracker payload for a single pageview.
type PageviewRequest struct {
	Hostname    string           `json:"hostname"`
	Path        string           `json:"path"`
	Referrer    string           `json:"referrer"`
	Language    string           `json:"language"`
	VisitorID   string           `json:"visitor_id"`
	ScreenWidth int              `json:"screen_width"`
	UTMSource   string           `json:"utm_source"`
	UTMMedium   string           `json:"utm_medium"`
	UTMCampaign string           `json:"utm_campaign"`
	UTMTerm     string           `json:"utm_term"`
	UTMContent  string           `json:"utm_content"`
	Performance *PerformanceData `json:"performance"`
}

// CustomEventRequest is the tracker payload for one custom event.
type CustomEventRequest struct {
	Hostname   string                 `json:"hostname"`
	Path       string                 `json:"path"`
	VisitorID  string                 `json:"visitor_id"`
	EventName  string                 `json:"event_name"`
	EventProps map[string]interface{} `json:"event_props"`
}
