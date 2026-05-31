package goals

import "time"

type CreateRequest struct {
	SiteID    int64  `json:"site_id"`
	Name      string `json:"name"`
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name"`
	PagePath  string `json:"page_path"`
	MatchType string `json:"match_type"`
}

type UpdateRequest struct {
	Name      string `json:"name"`
	GoalType  string `json:"goal_type"`
	EventName string `json:"event_name"`
	PagePath  string `json:"page_path"`
	MatchType string `json:"match_type"`
}

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

type GoalConversion struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	GoalType       string  `json:"goal_type"`
	Conversions    int64   `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
}

type ConversionsResponse struct {
	Goals         []GoalConversion `json:"goals"`
	TotalVisitors int64           `json:"total_visitors"`
}
