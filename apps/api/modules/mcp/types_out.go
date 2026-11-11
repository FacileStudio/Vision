package mcp

import "github.com/FacileStudio/Vision/apps/api/modules/analytics"

// spaceOut is one space the person belongs to.
type spaceOut struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Sites   int64  `json:"sites"`
	Members int64  `json:"members"`
}

// spacesOutput answers vision_list_spaces.
type spacesOutput struct {
	Spaces []spaceOut `json:"spaces"`
}

// siteOut is one site.
type siteOut struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	SpaceID  int64  `json:"space_id"`
	Space    string `json:"space,omitempty"`
	URL      string `json:"url"`
	ShareURL string `json:"share_url,omitempty"`
}

// sitesOutput answers vision_list_sites.
type sitesOutput struct {
	Sites     []siteOut `json:"sites"`
	Truncated bool      `json:"truncated,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// siteOutput answers the tools that write one site.
type siteOutput struct {
	Status          string   `json:"status"`
	Site            *siteOut `json:"site,omitempty"`
	SiteID          int64    `json:"site_id,omitempty"`
	TrackingSnippet string   `json:"tracking_snippet,omitempty"`
	ConfirmURL      string   `json:"confirm_url,omitempty"`
	Message         string   `json:"message,omitempty"`
}

// rangeOut is the days a report covers, in UTC.
type rangeOut struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// countOut is one row of a top list.
type countOut struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// statsOutput answers vision_get_stats. Detailed is set only for response_format detailed.
type statsOutput struct {
	Site               siteOut                     `json:"site"`
	Range              rangeOut                    `json:"range"`
	LiveVisitors       int64                       `json:"live_visitors"`
	Pageviews          int64                       `json:"pageviews"`
	Visitors           int64                       `json:"visitors"`
	PrevPageviews      int64                       `json:"previous_pageviews"`
	PrevVisitors       int64                       `json:"previous_visitors"`
	BounceRate         float64                     `json:"bounce_rate"`
	AvgSessionSeconds  float64                     `json:"avg_session_seconds"`
	PagesPerSession    float64                     `json:"pages_per_session"`
	TopPages           []countOut                  `json:"top_pages"`
	TopReferrers       []countOut                  `json:"top_referrers"`
	TopCountries       []countOut                  `json:"top_countries"`
	TopDevices         []countOut                  `json:"top_devices"`
	TopBrowsers        []countOut                  `json:"top_browsers"`
	TopEvents          []countOut                  `json:"top_events"`
	Detailed           *analytics.OverviewResponse `json:"detailed,omitempty"`
	DetailedListsLimit int                         `json:"detailed_lists_limit,omitempty"`
}

// goalOut is one goal with its conversions in the range.
type goalOut struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	GoalType       string  `json:"goal_type"`
	PagePath       string  `json:"page_path,omitempty"`
	MatchType      string  `json:"match_type,omitempty"`
	EventName      string  `json:"event_name,omitempty"`
	Conversions    int64   `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
}

// goalsOutput answers vision_list_goals.
type goalsOutput struct {
	Site          siteOut   `json:"site"`
	Range         rangeOut  `json:"range"`
	TotalVisitors int64     `json:"total_visitors"`
	Goals         []goalOut `json:"goals"`
}

// goalOutput answers the tools that write one goal.
type goalOutput struct {
	Status     string   `json:"status"`
	Goal       *goalOut `json:"goal,omitempty"`
	GoalID     int64    `json:"goal_id,omitempty"`
	ConfirmURL string   `json:"confirm_url,omitempty"`
	Message    string   `json:"message,omitempty"`
}
