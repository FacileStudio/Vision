package events

type PageviewRequest struct {
	Path      string `json:"path"`
	Referrer  string `json:"referrer"`
	Language  string `json:"language"`
	VisitorID string `json:"visitor_id"`
}
