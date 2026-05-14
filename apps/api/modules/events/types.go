package events

type PageviewRequest struct {
	Hostname  string `json:"hostname"`
	Path      string `json:"path"`
	Referrer  string `json:"referrer"`
	Language  string `json:"language"`
	VisitorID   string `json:"visitor_id"`
	ScreenWidth int    `json:"screen_width"`
	UTMSource   string `json:"utm_source"`
	UTMMedium   string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	UTMTerm     string `json:"utm_term"`
	UTMContent  string `json:"utm_content"`
}
