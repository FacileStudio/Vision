package analytics

import (
	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "analytics",
	Description: "Analytics and aggregated statistics for tracked sites.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/analytics/summary",
			Summary:     "Every readable site's headline numbers",
			Description: "Lists every site in the workspaces the caller belongs to, sorted by unique visitors: visitors, pageviews, the same counts for the window of equal length right before, live visitors, and a zero-filled pageview series bucketed by hour (windows of two days or less) or by day. A space the caller is not a member of answers an empty list.",
			Auth:        "bearer",
			QueryParams: []documentation.Field{
				{Name: "space", Type: "string", Description: "all (default) or a workspace ID"},
				{Name: "from", Type: "string", Description: "Start, RFC 3339 or YYYY-MM-DD (default 30 days ago)"},
				{Name: "to", Type: "string", Description: "End, RFC 3339 or YYYY-MM-DD (default now)"},
			},
			ResponseBody: SummaryResponse{},
		},
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
			Path:        "/analytics/{siteId}/anomalies",
			Summary:     "Detected anomalies of a site",
			Description: "Lists the days a metric left its usual range, newest day first. Each day is compared with the same weekday of the previous four to eight weeks (median and MAD), with a minimum volume. metric is visitors, pageviews, goal_conversions or top_referrer_share. breakdown names the paths, referrers, countries and devices that moved most, as pageview counts against the baseline average; those values were written by visitors.",
			Auth:        "bearer",
			QueryParams: []documentation.Field{
				{Name: "from", Type: "string", Description: "First day to include (YYYY-MM-DD)"},
				{Name: "to", Type: "string", Description: "Last day to include (YYYY-MM-DD)"},
				{Name: "metric", Type: "string", Description: "visitors, pageviews, goal_conversions or top_referrer_share"},
				{Name: "limit", Type: "number", Description: "At most this many anomalies (default 50, max 200)"},
			},
			ResponseBody: AnomaliesResponse{},
			PathParams:   []documentation.Field{{Name: "siteId", Type: "string", Description: "Site ID"}},
		},
		{
			Method:      "GET",
			Path:        "/analytics/{siteId}/compare",
			Summary:     "Compare two periods",
			Description: "Sets a period against another: pageviews, visitors, each goal's conversions, and per dimension (path, referrer, country, device) the values whose pageviews moved most. The comparison period defaults to the one of equal length right before. Periods are whole UTC days, at most 366 each.",
			Auth:        "bearer",
			QueryParams: []documentation.Field{
				{Name: "from", Type: "string", Description: "First day of the period (YYYY-MM-DD, default 6 days ago)"},
				{Name: "to", Type: "string", Description: "Last day of the period, included (YYYY-MM-DD, default today)"},
				{Name: "compare_from", Type: "string", Description: "First day of the comparison period (YYYY-MM-DD)"},
				{Name: "compare_to", Type: "string", Description: "Last day of the comparison period, included (YYYY-MM-DD)"},
				{Name: "limit", Type: "number", Description: "Values per dimension (default 10, max 50)"},
			},
			ResponseBody: CompareResponse{},
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
