package goals

import "time"

// CreateRequest is the payload for defining a goal.
type CreateRequest struct {
	SiteID    int64  `json:"site_id"`
	Name      string `json:"name"`
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name"`
	PagePath  string `json:"page_path"`
	MatchType string `json:"match_type"`
}

// UpdateRequest is the payload for editing a goal.
type UpdateRequest struct {
	Name      string `json:"name"`
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name"`
	PagePath  string `json:"page_path"`
	MatchType string `json:"match_type"`
}

// GoalResponse is a goal safe to return to a client.
type GoalResponse struct {
	ID        int64     `json:"id"`
	SiteID    int64     `json:"site_id"`
	Name      string    `json:"name"`
	GoalType  string    `json:"goal_type"`
	EventName *string   `json:"event_name"`
	PagePath  *string   `json:"page_path"`
	MatchType string    `json:"match_type"`
	CreatedAt time.Time `json:"created_at"`
}

// GoalConversion reports how many distinct visitors converted for a goal.
type GoalConversion struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	GoalType       string  `json:"goal_type"`
	Conversions    int64   `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
}

// ConversionsResponse aggregates conversion counts for a site's goals.
type ConversionsResponse struct {
	Goals         []GoalConversion `json:"goals"`
	TotalVisitors int64            `json:"total_visitors"`
}
