package webhooks

import "fmt"

type CreateWebhookRequest struct {
	URL           string `json:"url"`
	Secret        string `json:"secret"`
	IntervalHours int    `json:"interval_hours"`
}

type UpdateWebhookRequest struct {
	URL           string `json:"url"`
	Secret        string `json:"secret"`
	IntervalHours int    `json:"interval_hours"`
	Enabled       bool   `json:"enabled"`
}

type WebhookResponse struct {
	ID            int64   `json:"id"`
	URL           string  `json:"url"`
	Period        string  `json:"period"`
	IntervalHours int     `json:"interval_hours"`
	Enabled       bool    `json:"enabled"`
	LastSentAt    *string `json:"last_sent_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type ReportPayload struct {
	EventType    string        `json:"event_type"`
	Site         ReportSite    `json:"site"`
	Period       ReportPeriod  `json:"period"`
	Metrics      ReportMetrics `json:"metrics"`
	TopPages     []TopItem     `json:"top_pages"`
	TopReferrers []TopItem     `json:"top_referrers"`
	TopCountries []TopItem     `json:"top_countries"`
	TopBrowsers  []TopItem     `json:"top_browsers"`
	TopDevices   []TopItem     `json:"top_devices"`
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

func periodLabel(hours int) string {
	switch hours {
	case 1:
		return "hourly"
	case 24:
		return "daily"
	case 168:
		return "weekly"
	case 720:
		return "monthly"
	}
	if hours < 24 {
		return fmt.Sprintf("every %dh", hours)
	}
	days := hours / 24
	if hours%24 == 0 {
		if days == 1 {
			return "daily"
		}
		return fmt.Sprintf("every %dd", days)
	}
	return fmt.Sprintf("every %dh", hours)
}
