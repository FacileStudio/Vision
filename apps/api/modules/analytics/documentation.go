package analytics

import (
	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "analytics",
	Description: "Analytics and aggregated statistics for tracked sites.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/analytics/{siteId}/overview",
			Summary:      "Site analytics overview",
			Description:  "Returns aggregated stats: total pageviews, unique visitors, top pages, referrers, countries, and daily breakdown.",
			Auth:         "bearer",
			QueryParams:  []documentation.Field{{Name: "from", Type: "string", Description: "Start date (YYYY-MM-DD)"}, {Name: "to", Type: "string", Description: "End date (YYYY-MM-DD)"}, {Name: "granularity", Type: "string", Description: "day/hour"}},
			ResponseBody: OverviewResponse{},
			PathParams:   []documentation.Field{{Name: "siteId", Type: "string", Description: "Site ID"}},
		},
		{
			Method:       "GET",
			Path:         "/analytics/{siteId}/realtime",
			Summary:      "Site realtime active visitors",
			Auth:         "bearer",
			ResponseBody: map[string]int64{"visitors": 0},
			PathParams:   []documentation.Field{{Name: "siteId", Type: "string", Description: "Site ID"}},
		},
		{
			Method:      "GET",
			Path:        "/analytics/{siteId}/export",
			Summary:     "Export analytics data as CSV",
			Auth:        "bearer",
			QueryParams: []documentation.Field{{Name: "from", Type: "string", Description: "Start date (YYYY-MM-DD)"}, {Name: "to", Type: "string", Description: "End date (YYYY-MM-DD)"}},
			PathParams:  []documentation.Field{{Name: "siteId", Type: "string", Description: "Site ID"}},
		},
		{
			Method:       "GET",
			Path:         "/share/{token}",
			Summary:      "Public shared site analytics overview",
			QueryParams:  []documentation.Field{{Name: "from", Type: "string", Description: "Start date (YYYY-MM-DD)"}, {Name: "to", Type: "string", Description: "End date (YYYY-MM-DD)"}, {Name: "granularity", Type: "string", Description: "day/hour"}},
			ResponseBody: map[string]any{},
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Share Token"}},
		},
		{
			Method:       "GET",
			Path:         "/share/{token}/realtime",
			Summary:      "Public shared site active visitors",
			ResponseBody: map[string]int64{"visitors": 0},
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Share Token"}},
		},
	},
}
