package webhooks

import "fmt"

// CreateWebhookRequest is the payload for creating a report webhook.
type CreateWebhookRequest struct {
	URL           string `json:"url"`
	Secret        string `json:"secret"`
	IntervalHours int    `json:"interval_hours"`
	WorkspaceID   *int64 `json:"workspace_id"`
}

// UpdateWebhookRequest is the payload for editing a webhook.
type UpdateWebhookRequest struct {
	URL           string `json:"url"`
	Secret        string `json:"secret"`
	IntervalHours int    `json:"interval_hours"`
	Enabled       bool   `json:"enabled"`
}

// WebhookResponse is a webhook safe to return to a client.
type WebhookResponse struct {
	ID            int64   `json:"id"`
	URL           string  `json:"url"`
	Period        string  `json:"period"`
	IntervalHours int     `json:"interval_hours"`
	Enabled       bool    `json:"enabled"`
	WorkspaceID   *int64  `json:"workspace_id"`
	LastSentAt    *string `json:"last_sent_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ReportPayload is the JSON body sent to a webhook endpoint.
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

// ReportSite identifies the site a report describes.
type ReportSite struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// ReportPeriod describes the interval a report covers.
type ReportPeriod struct {
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

// ReportMetrics carries the headline numbers for a report.
type ReportMetrics struct {
	TotalPageviews     int64   `json:"total_pageviews"`
	UniqueVisitors     int64   `json:"unique_visitors"`
	ViewsPerVisitor    float64 `json:"views_per_visitor"`
	PrevTotalPageviews int64   `json:"prev_total_pageviews"`
	PrevUniqueVisitors int64   `json:"prev_unique_visitors"`
	PageviewsChangePct float64 `json:"pageviews_change_pct"`
	VisitorsChangePct  float64 `json:"visitors_change_pct"`
}

// TopItem is a ranked name/count pair in a report.
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
