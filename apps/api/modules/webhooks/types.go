package webhooks

type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
	Period string `json:"period"`
}

type UpdateWebhookRequest struct {
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Period  string `json:"period"`
	Enabled bool   `json:"enabled"`
}

type WebhookResponse struct {
	ID         int64   `json:"id"`
	SiteID     int64   `json:"site_id"`
	URL        string  `json:"url"`
	Period     string  `json:"period"`
	Enabled    bool    `json:"enabled"`
	LastSentAt *string `json:"last_sent_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type ReportPayload struct {
	EventType    string       `json:"event_type"`
	Site         ReportSite   `json:"site"`
	Period       ReportPeriod `json:"period"`
	Metrics      ReportMetrics `json:"metrics"`
	TopPages     []TopItem    `json:"top_pages"`
	TopReferrers []TopItem    `json:"top_referrers"`
	TopCountries []TopItem    `json:"top_countries"`
	TopBrowsers  []TopItem    `json:"top_browsers"`
	TopDevices   []TopItem    `json:"top_devices"`
}

type ReportSite struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type ReportPeriod struct {
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

type ReportMetrics struct {
	TotalPageviews     int64   `json:"total_pageviews"`
	UniqueVisitors     int64   `json:"unique_visitors"`
	ViewsPerVisitor    float64 `json:"views_per_visitor"`
	PrevTotalPageviews int64   `json:"prev_total_pageviews"`
	PrevUniqueVisitors int64   `json:"prev_unique_visitors"`
	PageviewsChangePct float64 `json:"pageviews_change_pct"`
	VisitorsChangePct  float64 `json:"visitors_change_pct"`
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
