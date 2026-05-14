package analytics

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "analytics",
	Description: "Analytics and aggregated statistics for tracked sites.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/analytics/{siteId}/overview",
			Summary:      "Site analytics overview",
			Description:  "Returns aggregated stats: total pageviews, unique visitors, top pages, referrers, countries, and daily breakdown. Accepts optional ?from=YYYY-MM-DD&to=YYYY-MM-DD query params (defaults to last 30 days).",
			Auth:         "bearer",
			ResponseBody: "OverviewResponse",
			PathParams:   []documentation.Field{{Name: "siteId", Type: "int", Description: "Site ID"}},
		},
	},
}
