package mcp

// spacesInput is vision_list_spaces's input: it takes none.
type spacesInput struct{}

// sitesInput is vision_list_sites's input.
type sitesInput struct {
	SpaceID int64  `json:"space_id,omitempty" jsonschema:"only the sites of this space, from vision_list_spaces. Default every space you belong to"`
	Query   string `json:"query,omitempty" jsonschema:"keep sites whose name or domain contains this, case ignored"`
}

// statsInput is vision_get_stats's input.
type statsInput struct {
	SiteID         int64  `json:"site_id" jsonschema:"the site, from vision_list_sites"`
	From           string `json:"from,omitempty" jsonschema:"first day, such as 2026-11-01, in UTC. Default 30 days ago"`
	To             string `json:"to,omitempty" jsonschema:"last day, included. Default today"`
	Granularity    string `json:"granularity,omitempty" jsonschema:"bucket of the detailed time series: hour, day (default), week or month"`
	Country        string `json:"country,omitempty" jsonschema:"only visits from this country code, such as FR"`
	Browser        string `json:"browser,omitempty" jsonschema:"only visits with this browser, such as Firefox"`
	OS             string `json:"os,omitempty" jsonschema:"only visits with this operating system, such as macOS"`
	Device         string `json:"device,omitempty" jsonschema:"only visits from this device type: Desktop, Mobile or Tablet"`
	Path           string `json:"path,omitempty" jsonschema:"only pageviews whose path contains this, such as /pricing"`
	Referrer       string `json:"referrer,omitempty" jsonschema:"only visits whose referrer contains this, such as news.ycombinator.com"`
	ResponseFormat string `json:"response_format,omitempty" jsonschema:"concise (default): totals and the top 10 of each list; detailed: every list, daily series, UTM, entry and exit pages, performance"`
}

// goalsInput is vision_list_goals's input.
type goalsInput struct {
	SiteID int64  `json:"site_id" jsonschema:"the site, from vision_list_sites"`
	From   string `json:"from,omitempty" jsonschema:"first day of the conversion count, such as 2026-11-01. Default 30 days ago"`
	To     string `json:"to,omitempty" jsonschema:"last day, included. Default today"`
}

// createSiteInput is vision_create_site's input.
type createSiteInput struct {
	SpaceID int64  `json:"space_id" jsonschema:"the space the site belongs to, from vision_list_spaces; you must be its owner, admin or editor"`
	Name    string `json:"name" jsonschema:"a display name, such as Marketing site"`
	Domain  string `json:"domain" jsonschema:"the host the script runs on, such as example.com, without scheme or path"`
}

// updateSiteInput is vision_update_site's input: only the fields given change.
type updateSiteInput struct {
	SiteID int64  `json:"site_id" jsonschema:"the site, from vision_list_sites"`
	Name   string `json:"name,omitempty" jsonschema:"the new display name"`
	Domain string `json:"domain,omitempty" jsonschema:"the new host; visits from the old host stop counting"`
}

// siteInput names one site.
type siteInput struct {
	SiteID int64 `json:"site_id" jsonschema:"the site, from vision_list_sites"`
}

// createGoalInput is vision_create_goal's input.
type createGoalInput struct {
	SiteID    int64  `json:"site_id" jsonschema:"the site, from vision_list_sites"`
	Name      string `json:"name" jsonschema:"a display name, such as Signed up"`
	GoalType  string `json:"goal_type" jsonschema:"pageview (a visitor reaches a path) or event (a visitor sends a custom event)"`
	PagePath  string `json:"page_path,omitempty" jsonschema:"for pageview goals: the path, such as /thanks"`
	MatchType string `json:"match_type,omitempty" jsonschema:"for pageview goals: exact (default), starts_with or contains"`
	EventName string `json:"event_name,omitempty" jsonschema:"for event goals: the name the site's script sends"`
}

// updateGoalInput is vision_update_goal's input: only the fields given change.
type updateGoalInput struct {
	GoalID    int64  `json:"goal_id" jsonschema:"one of your own goals, from vision_list_goals"`
	Name      string `json:"name,omitempty" jsonschema:"the new display name"`
	GoalType  string `json:"goal_type,omitempty" jsonschema:"pageview or event; pass page_path or event_name to match"`
	PagePath  string `json:"page_path,omitempty" jsonschema:"for pageview goals: the path"`
	MatchType string `json:"match_type,omitempty" jsonschema:"for pageview goals: exact, starts_with or contains"`
	EventName string `json:"event_name,omitempty" jsonschema:"for event goals: the event name"`
}

// goalInput names one goal.
type goalInput struct {
	GoalID int64 `json:"goal_id" jsonschema:"one of your own goals, from vision_list_goals"`
}
